package subscription

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFeatureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckFeatureLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFeatureLogic {
	return &CheckFeatureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Feature access rules - customize based on your plans
var planFeatures = map[string][]string{
	"free": {"basic", "community_support"},
	"pro":  {"basic", "community_support", "priority_support", "analytics", "unlimited_projects"},
	"team": {"basic", "community_support", "priority_support", "analytics", "unlimited_projects", "team", "api", "admin"},
}

func (l *CheckFeatureLogic) CheckFeature(req *types.CheckFeatureRequest) (resp *types.CheckFeatureResponse, err error) {
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("levee service not configured")
	}

	// Get email from JWT context
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	// Get current subscription from Levee SDK
	subsResp, err := l.svcCtx.Levee.Customers.ListCustomerSubscriptions(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get subscriptions for %s: %v", email, err)
		return nil, err
	}

	subs := subsResp.Subscriptions

	// Default to free plan
	planName := "free"
	if len(subs) > 0 {
		planName = subs[0].ProductName
	}

	// Check if feature is available for this plan
	features, ok := planFeatures[planName]
	if !ok {
		features = planFeatures["free"]
	}

	hasAccess := false
	for _, f := range features {
		if f == req.Feature {
			hasAccess = true
			break
		}
	}

	message := ""
	if !hasAccess {
		message = "Upgrade your plan to access this feature."
	}

	return &types.CheckFeatureResponse{
		HasAccess: hasAccess,
		Feature:   req.Feature,
		PlanName:  planName,
		Message:   message,
	}, nil
}
