package subscription

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	levee "github.com/almatuck/levee-go"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCheckoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCheckoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCheckoutLogic {
	return &CreateCheckoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// getProductSlug maps plan name and cycle to Levee product slug
func (l *CreateCheckoutLogic) getProductSlug(planName, billingCycle string) string {
	isYearly := billingCycle == "yearly" || billingCycle == "annual"
	switch planName {
	case "free":
		return l.svcCtx.Config.Levee.FreeProductSlug
	case "pro":
		if isYearly {
			return l.svcCtx.Config.Levee.ProYearlyProductSlug
		}
		return l.svcCtx.Config.Levee.ProMonthlyProductSlug
	case "team":
		if isYearly {
			return l.svcCtx.Config.Levee.TeamYearlyProductSlug
		}
		return l.svcCtx.Config.Levee.TeamMonthlyProductSlug
	default:
		return ""
	}
}

func (l *CreateCheckoutLogic) CreateCheckout(req *types.CreateCheckoutRequest) (resp *types.CreateCheckoutResponse, err error) {
	// Get email and user ID from JWT context
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	userID, _ := auth.GetCustomerIDFromContext(l.ctx)

	// Use local billing when Levee is disabled
	if l.svcCtx.UseLocal() {
		return l.createCheckoutLocal(userID, email, req)
	}

	// Use Levee when enabled
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("billing service not configured")
	}

	// Get product slug
	productSlug := l.getProductSlug(req.PlanName, req.BillingCycle)
	if productSlug == "" {
		return nil, fmt.Errorf("invalid plan: %s %s", req.PlanName, req.BillingCycle)
	}

	// Build success/cancel URLs from config
	baseURL := l.svcCtx.Config.App.BaseURL
	successURL := baseURL + l.svcCtx.Config.Levee.CheckoutSuccessURL
	cancelURL := baseURL + l.svcCtx.Config.Levee.CheckoutCancelURL

	// Create order via Levee SDK (returns checkout URL)
	orderResp, err := l.svcCtx.Levee.Orders.CreateOrder(l.ctx, &levee.OrderRequest{
		Email:       email,
		ProductSlug: productSlug,
		SuccessUrl:  successURL,
		CancelUrl:   cancelURL,
	})
	if err != nil {
		l.Errorf("Failed to create checkout for %s: %v", email, err)
		return nil, err
	}

	return &types.CreateCheckoutResponse{
		CheckoutUrl: orderResp.CheckoutUrl,
	}, nil
}

// createCheckoutLocal creates checkout using direct Stripe integration
func (l *CreateCheckoutLogic) createCheckoutLocal(userID, email string, req *types.CreateCheckoutRequest) (*types.CreateCheckoutResponse, error) {
	if l.svcCtx.Billing == nil {
		return nil, fmt.Errorf("local billing service not configured")
	}

	// Map plan name to Stripe price
	planName := req.PlanName
	if req.BillingCycle == "yearly" || req.BillingCycle == "annual" {
		planName = planName + "-yearly"
	}

	checkoutURL, err := l.svcCtx.Billing.CreateCheckoutSession(l.ctx, userID, email, planName)
	if err != nil {
		l.Errorf("Failed to create checkout for %s: %v", email, err)
		return nil, err
	}

	return &types.CreateCheckoutResponse{
		CheckoutUrl: checkoutURL,
	}, nil
}
