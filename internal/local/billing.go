package local

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gobot/internal/config"
	"gobot/internal/db"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

var (
	ErrNoSubscription      = errors.New("no subscription found")
	ErrStripeNotConfigured = errors.New("stripe not configured")
)

// BillingService handles local billing with direct Stripe integration
type BillingService struct {
	store  *db.Store
	config config.Config
	stripe *client.API
}

// NewBillingService creates a new local billing service
func NewBillingService(store *db.Store, cfg config.Config) *BillingService {
	var stripeClient *client.API
	if cfg.Stripe.SecretKey != "" {
		stripeClient = client.New(cfg.Stripe.SecretKey, nil)
	}
	return &BillingService{
		store:  store,
		config: cfg,
		stripe: stripeClient,
	}
}

// Plan represents a subscription plan
type Plan struct {
	ID          string
	Name        string
	DisplayName string
	Description string
	Price       int64 // in cents
	Currency    string
	Interval    string
	Features    []string
}

// GetPlans returns available subscription plans from config
func (s *BillingService) GetPlans() []Plan {
	if len(s.config.Products) > 0 {
		var plans []Plan
		for _, product := range s.config.Products {
			for _, price := range product.Prices {
				planID := product.Slug
				if price.Slug != "" && price.Slug != "default" {
					planID = product.Slug + "-" + price.Slug
				}
				plans = append(plans, Plan{
					ID:          planID,
					Name:        planID,
					DisplayName: product.Name,
					Description: product.Description,
					Price:       price.Amount,
					Currency:    price.Currency,
					Interval:    price.Interval,
					Features:    product.Features,
				})
			}
		}
		return plans
	}

	// Fallback defaults
	return []Plan{
		{
			ID:          "free",
			Name:        "free",
			DisplayName: "Free",
			Description: "Get started for free",
			Price:       0,
			Currency:    "usd",
			Interval:    "month",
			Features:    []string{"Basic features", "Community support"},
		},
		{
			ID:          "pro",
			Name:        "pro",
			DisplayName: "Pro",
			Description: "For professionals",
			Price:       2900,
			Currency:    "usd",
			Interval:    "month",
			Features:    []string{"All Free features", "Priority support", "Advanced features"},
		},
	}
}

// GetSubscription returns a user's current subscription
func (s *BillingService) GetSubscription(ctx context.Context, userID string) (*db.Subscription, error) {
	sub, err := s.store.GetSubscriptionByUserID(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, ErrNoSubscription
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &sub, nil
}

// CreateCheckoutSession creates a Stripe checkout session
func (s *BillingService) CreateCheckoutSession(ctx context.Context, userID, email, planName string) (string, error) {
	if s.stripe == nil {
		return "", ErrStripeNotConfigured
	}

	priceID := s.getPriceID(planName)
	if priceID == "" {
		return "", fmt.Errorf("unknown plan: %s", planName)
	}

	customerID, err := s.getOrCreateCustomer(ctx, userID, email)
	if err != nil {
		return "", fmt.Errorf("failed to get/create customer: %w", err)
	}

	baseURL := s.config.App.BaseURL
	successURL := baseURL + s.config.Stripe.SuccessURL + "?session_id={CHECKOUT_SESSION_ID}"
	cancelURL := baseURL + s.config.Stripe.CancelURL

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}
	params.AddMetadata("user_id", userID)

	sess, err := s.stripe.CheckoutSessions.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// CreateBillingPortalSession creates a Stripe billing portal session
func (s *BillingService) CreateBillingPortalSession(ctx context.Context, userID string) (string, error) {
	if s.stripe == nil {
		return "", ErrStripeNotConfigured
	}

	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return "", err
	}

	if !sub.StripeCustomerID.Valid || sub.StripeCustomerID.String == "" {
		return "", fmt.Errorf("no stripe customer found")
	}

	returnURL := s.config.App.BaseURL + "/app/account"

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sub.StripeCustomerID.String),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := s.stripe.BillingPortalSessions.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create portal session: %w", err)
	}

	return sess.URL, nil
}

// CancelSubscription cancels a user's subscription
func (s *BillingService) CancelSubscription(ctx context.Context, userID string, atPeriodEnd bool) error {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return err
	}

	if !sub.StripeSubscriptionID.Valid || sub.StripeSubscriptionID.String == "" {
		// Just update local status for free plan
		cancelFlag := int64(0)
		if atPeriodEnd {
			cancelFlag = 1
		}
		return s.store.CancelSubscription(ctx, db.CancelSubscriptionParams{
			Status:            "cancelled",
			CancelAtPeriodEnd: cancelFlag,
			UserID:            userID,
		})
	}

	if s.stripe == nil {
		return ErrStripeNotConfigured
	}

	// Cancel in Stripe
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(atPeriodEnd),
	}
	_, err = s.stripe.Subscriptions.Update(sub.StripeSubscriptionID.String, params)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Update local status
	status := "cancelled"
	if atPeriodEnd {
		status = "active"
	}
	cancelFlag := int64(0)
	if atPeriodEnd {
		cancelFlag = 1
	}
	return s.store.CancelSubscription(ctx, db.CancelSubscriptionParams{
		Status:            status,
		CancelAtPeriodEnd: cancelFlag,
		UserID:            userID,
	})
}

// HandleCheckoutCompleted handles a successful checkout
func (s *BillingService) HandleCheckoutCompleted(ctx context.Context, customerID, subscriptionID, userID string) error {
	return s.store.UpdateSubscriptionStripeIDs(ctx, db.UpdateSubscriptionStripeIDsParams{
		StripeCustomerID:     sql.NullString{String: customerID, Valid: customerID != ""},
		StripeSubscriptionID: sql.NullString{String: subscriptionID, Valid: subscriptionID != ""},
		PlanID:               "pro",
		Status:               "active",
		UserID:               userID,
	})
}

// HandleSubscriptionUpdated handles subscription updates from webhooks
func (s *BillingService) HandleSubscriptionUpdated(ctx context.Context, stripeSubID, status string, periodStart, periodEnd int64, cancelAtPeriodEnd bool) error {
	cancelFlag := int64(0)
	if cancelAtPeriodEnd {
		cancelFlag = 1
	}

	return s.store.UpdateSubscriptionStatus(ctx, db.UpdateSubscriptionStatusParams{
		Status:               status,
		PeriodStart:          sql.NullInt64{Int64: periodStart, Valid: periodStart > 0},
		PeriodEnd:            sql.NullInt64{Int64: periodEnd, Valid: periodEnd > 0},
		CancelAtPeriodEnd:    cancelFlag,
		StripeSubscriptionID: sql.NullString{String: stripeSubID, Valid: true},
	})
}

// HandleSubscriptionDeleted handles subscription deletion (downgrade to free)
func (s *BillingService) HandleSubscriptionDeleted(ctx context.Context, stripeSubID string) error {
	return s.store.DowngradeToFree(ctx, sql.NullString{String: stripeSubID, Valid: true})
}

// getOrCreateCustomer gets or creates a Stripe customer for a user
func (s *BillingService) getOrCreateCustomer(ctx context.Context, userID, email string) (string, error) {
	sub, err := s.store.GetSubscriptionByUserID(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	if sub.StripeCustomerID.Valid && sub.StripeCustomerID.String != "" {
		return sub.StripeCustomerID.String, nil
	}

	// Create new Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	params.AddMetadata("user_id", userID)

	cust, err := s.stripe.Customers.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	// Save customer ID
	err = s.store.UpdateSubscriptionStripeIDs(ctx, db.UpdateSubscriptionStripeIDsParams{
		StripeCustomerID:     sql.NullString{String: cust.ID, Valid: true},
		StripeSubscriptionID: sub.StripeSubscriptionID,
		PlanID:               sub.PlanID,
		Status:               sub.Status,
		UserID:               userID,
	})
	if err != nil {
		return "", err
	}

	return cust.ID, nil
}

// getPriceID returns the Stripe price ID for a plan name
func (s *BillingService) getPriceID(planName string) string {
	for _, product := range s.config.Products {
		for _, price := range product.Prices {
			priceKey := product.Slug
			if price.Slug != "" && price.Slug != "default" {
				priceKey = product.Slug + "-" + price.Slug
			}
			if planName == priceKey || planName == product.Slug {
				return price.StripePriceID
			}
		}
	}
	return ""
}

// GetProduct returns a product by slug
func (s *BillingService) GetProduct(slug string) *config.Product {
	for _, product := range s.config.Products {
		if product.Slug == slug {
			return &product
		}
	}
	return nil
}

// GetPrice returns a price by product slug and price slug
func (s *BillingService) GetPrice(productSlug, priceSlug string) *config.Price {
	product := s.GetProduct(productSlug)
	if product == nil {
		return nil
	}
	for _, price := range product.Prices {
		if price.Slug == priceSlug || (priceSlug == "" && price.Slug == "default") {
			return &price
		}
	}
	if len(product.Prices) > 0 {
		return &product.Prices[0]
	}
	return nil
}

// SyncProductsToStripe syncs products from config to Stripe
func (s *BillingService) SyncProductsToStripe(ctx context.Context) ([]config.Product, error) {
	if s.stripe == nil {
		return nil, ErrStripeNotConfigured
	}

	var syncedProducts []config.Product

	for _, product := range s.config.Products {
		syncedProduct, err := s.syncProduct(ctx, product)
		if err != nil {
			return nil, fmt.Errorf("failed to sync product %s: %w", product.Slug, err)
		}
		syncedProducts = append(syncedProducts, *syncedProduct)
	}

	return syncedProducts, nil
}

// syncProduct syncs a single product to Stripe
func (s *BillingService) syncProduct(_ context.Context, product config.Product) (*config.Product, error) {
	existingProduct, err := s.findStripeProduct(product.Slug)
	if err != nil {
		return nil, err
	}

	var stripeProductID string

	if existingProduct != nil {
		stripeProductID = existingProduct.ID
		params := &stripe.ProductParams{
			Name:        stripe.String(product.Name),
			Description: stripe.String(product.Description),
		}
		params.AddMetadata("gobot_slug", product.Slug)

		_, err = s.stripe.Products.Update(stripeProductID, params)
		if err != nil {
			return nil, fmt.Errorf("failed to update product: %w", err)
		}
	} else {
		params := &stripe.ProductParams{
			Name:        stripe.String(product.Name),
			Description: stripe.String(product.Description),
		}
		params.AddMetadata("gobot_slug", product.Slug)

		newProduct, err := s.stripe.Products.New(params)
		if err != nil {
			return nil, fmt.Errorf("failed to create product: %w", err)
		}
		stripeProductID = newProduct.ID
	}

	var syncedPrices []config.Price
	for _, price := range product.Prices {
		syncedPrice, err := s.syncPrice(stripeProductID, product.Slug, price)
		if err != nil {
			return nil, fmt.Errorf("failed to sync price %s: %w", price.Slug, err)
		}
		syncedPrices = append(syncedPrices, *syncedPrice)
	}

	return &config.Product{
		Slug:            product.Slug,
		Name:            product.Name,
		Description:     product.Description,
		Features:        product.Features,
		Default:         product.Default,
		StripeProductID: stripeProductID,
		Prices:          syncedPrices,
	}, nil
}

// syncPrice syncs a single price to Stripe
func (s *BillingService) syncPrice(stripeProductID, productSlug string, price config.Price) (*config.Price, error) {
	priceKey := productSlug + "-" + price.Slug

	existingPrice, err := s.findStripePrice(priceKey)
	if err != nil {
		return nil, err
	}

	if existingPrice != nil {
		return &config.Price{
			Slug:          price.Slug,
			Amount:        price.Amount,
			Currency:      price.Currency,
			Interval:      price.Interval,
			IntervalCount: price.IntervalCount,
			TrialDays:     price.TrialDays,
			StripePriceID: existingPrice.ID,
		}, nil
	}

	params := &stripe.PriceParams{
		Product:    stripe.String(stripeProductID),
		Currency:   stripe.String(price.Currency),
		UnitAmount: stripe.Int64(price.Amount),
	}
	params.AddMetadata("gobot_key", priceKey)

	if price.Interval != "one_time" && price.Interval != "" {
		params.Recurring = &stripe.PriceRecurringParams{
			Interval:      stripe.String(price.Interval),
			IntervalCount: stripe.Int64(price.IntervalCount),
		}
		if price.TrialDays > 0 {
			params.Recurring.TrialPeriodDays = stripe.Int64(price.TrialDays)
		}
	}

	newPrice, err := s.stripe.Prices.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	return &config.Price{
		Slug:          price.Slug,
		Amount:        price.Amount,
		Currency:      price.Currency,
		Interval:      price.Interval,
		IntervalCount: price.IntervalCount,
		TrialDays:     price.TrialDays,
		StripePriceID: newPrice.ID,
	}, nil
}

// findStripeProduct looks for an existing product by slug metadata
func (s *BillingService) findStripeProduct(slug string) (*stripe.Product, error) {
	params := &stripe.ProductListParams{}
	params.Filters.AddFilter("limit", "", "100")

	i := s.stripe.Products.List(params)
	for i.Next() {
		p := i.Product()
		if p.Metadata["gobot_slug"] == slug {
			return p, nil
		}
	}
	if err := i.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

// findStripePrice looks for an existing price by key metadata
func (s *BillingService) findStripePrice(priceKey string) (*stripe.Price, error) {
	params := &stripe.PriceListParams{}
	params.Filters.AddFilter("limit", "", "100")

	i := s.stripe.Prices.List(params)
	for i.Next() {
		p := i.Price()
		if p.Metadata["gobot_key"] == priceKey {
			return p, nil
		}
	}
	if err := i.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

// Subscription type alias for backward compatibility
type Subscription = db.Subscription

// Helper to convert subscription to time values
func (s *BillingService) GetSubscriptionTimes(sub *db.Subscription) (start, end time.Time, cancelAtEnd bool) {
	if sub.CurrentPeriodStart.Valid {
		start = time.Unix(sub.CurrentPeriodStart.Int64, 0)
	}
	if sub.CurrentPeriodEnd.Valid {
		end = time.Unix(sub.CurrentPeriodEnd.Int64, 0)
	}
	cancelAtEnd = sub.CancelAtPeriodEnd == 1
	return
}
