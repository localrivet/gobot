package admin

import (
	"context"
	"fmt"
	"time"

	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListSubscriptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all subscriptions (paginated)
func NewAdminListSubscriptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListSubscriptionsLogic {
	return &AdminListSubscriptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListSubscriptionsLogic) AdminListSubscriptions(req *types.AdminListSubscriptionsRequest) (resp *types.AdminListSubscriptionsResponse, err error) {
	// Only support local/standalone mode
	if !l.svcCtx.UseLocal() {
		return nil, fmt.Errorf("admin subscriptions list only available in standalone mode")
	}

	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	statusFilter := req.Status

	// Get total count (with filter)
	totalCount, err := l.svcCtx.DB.CountSubscriptionsFiltered(l.ctx, statusFilter)
	if err != nil {
		l.Errorf("Failed to count subscriptions: %v", err)
		return nil, fmt.Errorf("failed to count subscriptions")
	}

	// Get subscriptions
	offset := int64((page - 1) * pageSize)
	subscriptions, err := l.svcCtx.DB.ListSubscriptionsPaginated(l.ctx, db.ListSubscriptionsPaginatedParams{
		StatusFilter: statusFilter,
		PageOffset:   offset,
		PageSize:     int64(pageSize),
	})
	if err != nil {
		l.Errorf("Failed to list subscriptions: %v", err)
		return nil, fmt.Errorf("failed to list subscriptions")
	}

	// Convert to response type
	adminSubs := make([]types.AdminSubscription, 0, len(subscriptions))
	for _, s := range subscriptions {
		adminSubs = append(adminSubs, types.AdminSubscription{
			Id:        s.ID,
			UserId:    s.UserID,
			UserEmail: s.UserEmail,
			PlanName:  s.PlanID,
			Status:    s.Status,
			Amount:    0,   // Would need to look up from config/Stripe
			Currency:  "usd",
			Interval:  "month", // Default, would need to look up
			CreatedAt: time.Unix(s.CreatedAt, 0).Format(time.RFC3339),
		})
	}

	return &types.AdminListSubscriptionsResponse{
		Subscriptions: adminSubs,
		TotalCount:    int(totalCount),
		Page:          page,
		PageSize:      pageSize,
	}, nil
}
