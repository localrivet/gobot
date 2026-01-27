package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gobot/internal/svc"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"github.com/zeromicro/go-zero/core/logx"
)

// StripeHandler creates a raw HTTP handler for Stripe webhooks
// This must be registered directly with server.AddRoute() to access raw body
// for signature verification (go-zero handlers parse the body too early)
func StripeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read raw body for signature verification
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logx.Errorf("Failed to read webhook body: %v", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		// Get Stripe signature header
		sig := r.Header.Get("Stripe-Signature")
		if sig == "" {
			logx.Error("Missing Stripe-Signature header")
			http.Error(w, "Missing signature", http.StatusBadRequest)
			return
		}

		// Verify webhook signature
		webhookSecret := svcCtx.Config.Stripe.WebhookSecret
		if webhookSecret == "" {
			logx.Error("Stripe webhook secret not configured")
			http.Error(w, "Webhook not configured", http.StatusInternalServerError)
			return
		}

		event, err := webhook.ConstructEventWithOptions(body, sig, webhookSecret, webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		})
		if err != nil {
			logx.Errorf("Failed to verify webhook signature: %v", err)
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		// Handle the event
		if err := handleStripeEvent(svcCtx, &event); err != nil {
			logx.Errorf("Failed to handle webhook event %s: %v", event.Type, err)
			http.Error(w, "Failed to handle event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received": true}`))
	}
}

// handleStripeEvent processes different Stripe event types
func handleStripeEvent(svcCtx *svc.ServiceContext, event *stripe.Event) error {
	ctx := context.Background()

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return err
		}
		return handleCheckoutCompleted(svcCtx, ctx, &session)

	case "customer.subscription.created", "customer.subscription.updated":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return err
		}
		return handleSubscriptionUpdated(svcCtx, ctx, &subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return err
		}
		return handleSubscriptionDeleted(svcCtx, ctx, &subscription)

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return err
		}
		logx.Infof("Invoice payment succeeded: %s", invoice.ID)

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return err
		}
		logx.Infof("[WARN] Invoice payment failed: %s for customer %s", invoice.ID, invoice.Customer.ID)

	default:
		logx.Infof("Unhandled Stripe event type: %s", event.Type)
	}

	return nil
}

// handleCheckoutCompleted processes successful checkout sessions
func handleCheckoutCompleted(svcCtx *svc.ServiceContext, ctx context.Context, session *stripe.CheckoutSession) error {
	customerID := ""
	if session.Customer != nil {
		customerID = session.Customer.ID
	}
	subscriptionID := ""
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID
	}

	logx.Infof("Checkout completed: session=%s customer=%s subscription=%s",
		session.ID, customerID, subscriptionID)

	// Get user ID from metadata
	userID, ok := session.Metadata["user_id"]
	if !ok || userID == "" {
		logx.Info("No user_id in checkout session metadata")
		return nil
	}

	// Update subscription in database
	if svcCtx.Billing != nil {
		return svcCtx.Billing.HandleCheckoutCompleted(ctx, customerID, subscriptionID, userID)
	}

	return nil
}

// handleSubscriptionUpdated processes subscription updates
func handleSubscriptionUpdated(svcCtx *svc.ServiceContext, ctx context.Context, subscription *stripe.Subscription) error {
	logx.Infof("Subscription updated: id=%s status=%s", subscription.ID, subscription.Status)

	if svcCtx.Billing != nil {
		// Get period from subscription items (Stripe API 2025-03-31 moved these fields)
		var periodStart, periodEnd int64
		if len(subscription.Items.Data) > 0 {
			item := subscription.Items.Data[0]
			periodStart = item.CurrentPeriodStart
			periodEnd = item.CurrentPeriodEnd
		}

		return svcCtx.Billing.HandleSubscriptionUpdated(
			ctx,
			subscription.ID,
			string(subscription.Status),
			periodStart,
			periodEnd,
			subscription.CancelAtPeriodEnd,
		)
	}

	return nil
}

// handleSubscriptionDeleted processes subscription cancellations
func handleSubscriptionDeleted(svcCtx *svc.ServiceContext, ctx context.Context, subscription *stripe.Subscription) error {
	logx.Infof("Subscription deleted: id=%s", subscription.ID)

	if svcCtx.Billing != nil {
		return svcCtx.Billing.HandleSubscriptionDeleted(ctx, subscription.ID)
	}

	return nil
}
