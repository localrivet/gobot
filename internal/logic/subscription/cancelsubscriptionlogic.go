package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelSubscriptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelSubscriptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelSubscriptionLogic {
	return &CancelSubscriptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelSubscriptionLogic) CancelSubscription(req *types.CancelSubscriptionRequest) (resp *types.CancelSubscriptionResponse, err error) {
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("levee service not configured")
	}

	// Get email from JWT context
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	// Get current subscriptions from Levee SDK
	subsResp, err := l.svcCtx.Levee.Customers.ListCustomerSubscriptions(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get subscriptions for %s: %v", email, err)
		return nil, err
	}

	if len(subsResp.Subscriptions) == 0 {
		return nil, errors.New("no active subscription found")
	}

	// Cancel the first active subscription via Levee SDK
	sub := subsResp.Subscriptions[0]
	_, err = l.svcCtx.Levee.Billing.CancelSubscription(l.ctx, sub.ID)
	if err != nil {
		l.Errorf("Failed to cancel subscription for %s: %v", email, err)
		return nil, err
	}

	now := time.Now().Format(time.RFC3339)
	return &types.CancelSubscriptionResponse{
		Message:     "Subscription cancelled successfully.",
		CancelledAt: now,
		EffectiveAt: sub.CurrentPeriodEnd,
	}, nil
}
