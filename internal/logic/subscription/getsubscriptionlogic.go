package subscription

import (
	"context"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubscriptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubscriptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscriptionLogic {
	return &GetSubscriptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubscriptionLogic) GetSubscription() (resp *types.GetSubscriptionResponse, err error) {
	// Get user ID from JWT context
	userID, err := auth.GetCustomerIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get user ID from context: %v", err)
		return nil, err
	}

	// Use local billing when Levee is disabled
	if l.svcCtx.UseLocal() {
		return l.getSubscriptionLocal(userID)
	}

	// Use Levee when enabled
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("billing service not configured")
	}

	// Get email from JWT context for Levee lookup
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	// Get subscriptions from Levee SDK
	subsResp, err := l.svcCtx.Levee.Customers.ListCustomerSubscriptions(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get subscriptions for %s: %v", email, err)
		return nil, err
	}

	subs := subsResp.Subscriptions

	// Default to free tier if no subscription
	if len(subs) == 0 {
		return l.freeSubscriptionResponse(), nil
	}

	// Return first active subscription
	sub := subs[0]
	return &types.GetSubscriptionResponse{
		Subscription: types.UserSubscription{
			Id:                 sub.ID,
			PlanId:             sub.ProductName,
			PlanName:           sub.ProductName,
			Status:             sub.Status,
			BillingCycle:       sub.Interval,
			CurrentPeriodStart: sub.CurrentPeriodStart,
			CurrentPeriodEnd:   sub.CurrentPeriodEnd,
			CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		},
		Plan: types.SubscriptionPlan{
			Id:          sub.ProductName,
			Name:        sub.ProductName,
			DisplayName: sub.PriceName,
			Price:       int(sub.AmountCents),
			Currency:    sub.Currency,
			Interval:    sub.Interval,
		},
	}, nil
}

// getSubscriptionLocal gets subscription from local SQLite
func (l *GetSubscriptionLogic) getSubscriptionLocal(userID string) (*types.GetSubscriptionResponse, error) {
	if l.svcCtx.Billing == nil {
		return nil, fmt.Errorf("local billing service not configured")
	}

	sub, err := l.svcCtx.Billing.GetSubscription(l.ctx, userID)
	if err != nil {
		// Return free tier if no subscription found
		return l.freeSubscriptionResponse(), nil
	}

	// Find the matching plan
	plans := l.svcCtx.Billing.GetPlans()
	var plan types.SubscriptionPlan
	for _, p := range plans {
		if p.Name == sub.PlanID {
			plan = types.SubscriptionPlan{
				Id:          p.ID,
				Name:        p.Name,
				DisplayName: p.DisplayName,
				Description: p.Description,
				Price:       int(p.Price),
				Currency:    p.Currency,
				Interval:    p.Interval,
				Features:    p.Features,
			}
			break
		}
	}

	// Convert nullable timestamps
	var periodStart, periodEnd string
	if sub.CurrentPeriodStart.Valid {
		periodStart = time.Unix(sub.CurrentPeriodStart.Int64, 0).Format("2006-01-02T15:04:05Z")
	}
	if sub.CurrentPeriodEnd.Valid {
		periodEnd = time.Unix(sub.CurrentPeriodEnd.Int64, 0).Format("2006-01-02T15:04:05Z")
	}

	return &types.GetSubscriptionResponse{
		Subscription: types.UserSubscription{
			Id:                 sub.ID,
			PlanId:             sub.PlanID,
			PlanName:           sub.PlanID,
			Status:             sub.Status,
			BillingCycle:       "monthly",
			CurrentPeriodStart: periodStart,
			CurrentPeriodEnd:   periodEnd,
			CancelAtPeriodEnd:  sub.CancelAtPeriodEnd == 1,
		},
		Plan: plan,
	}, nil
}

// freeSubscriptionResponse returns the default free tier response
func (l *GetSubscriptionLogic) freeSubscriptionResponse() *types.GetSubscriptionResponse {
	return &types.GetSubscriptionResponse{
		Subscription: types.UserSubscription{
			Id:                "free",
			PlanId:            "free",
			PlanName:          "free",
			Status:            "active",
			BillingCycle:      "monthly",
			CancelAtPeriodEnd: false,
		},
		Plan: types.SubscriptionPlan{
			Id:          "free",
			Name:        "free",
			DisplayName: "Free",
			Description: "Get started with basic features",
			Price:       0,
			Currency:    "usd",
			Interval:    "month",
			Features:    []string{"Basic features", "Community support"},
		},
	}
}
