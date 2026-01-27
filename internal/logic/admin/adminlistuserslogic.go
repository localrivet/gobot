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

type AdminListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all users (paginated)
func NewAdminListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListUsersLogic {
	return &AdminListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListUsersLogic) AdminListUsers(req *types.AdminListUsersRequest) (resp *types.AdminListUsersResponse, err error) {
	// Only support local/standalone mode
	if !l.svcCtx.UseLocal() {
		return nil, fmt.Errorf("admin users list only available in standalone mode")
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

	search := req.Search

	// Get total count
	totalCount, err := l.svcCtx.DB.CountUsers(l.ctx)
	if err != nil {
		l.Errorf("Failed to count users: %v", err)
		return nil, fmt.Errorf("failed to count users")
	}

	// Get users
	offset := int64((page - 1) * pageSize)
	users, err := l.svcCtx.DB.ListUsersPaginated(l.ctx, db.ListUsersPaginatedParams{
		Search:     search,
		PageOffset: offset,
		PageSize:   int64(pageSize),
	})
	if err != nil {
		l.Errorf("Failed to list users: %v", err)
		return nil, fmt.Errorf("failed to list users")
	}

	// Convert to response type
	adminUsers := make([]types.AdminUser, 0, len(users))
	for _, u := range users {
		// Get user's subscription to determine plan
		plan := "free"
		status := "active"
		sub, err := l.svcCtx.DB.GetSubscriptionByUserID(l.ctx, u.ID)
		if err == nil {
			plan = sub.PlanID
			status = sub.Status
		}

		adminUsers = append(adminUsers, types.AdminUser{
			Id:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Plan:      plan,
			Status:    status,
			CreatedAt: time.Unix(u.CreatedAt, 0).Format(time.RFC3339),
		})
	}

	return &types.AdminListUsersResponse{
		Users:      adminUsers,
		TotalCount: int(totalCount),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}
