package notification

import (
	"context"
	"time"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNotificationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNotificationsLogic {
	return &ListNotificationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNotificationsLogic) ListNotifications(req *types.ListNotificationsRequest) (resp *types.ListNotificationsResponse, err error) {
	// Check if notifications are enabled
	if !l.svcCtx.Config.IsNotificationsEnabled() {
		return &types.ListNotificationsResponse{
			Notifications: []types.Notification{},
			UnreadCount:   0,
			TotalCount:    0,
		}, nil
	}

	if !l.svcCtx.UseLocal() {
		return &types.ListNotificationsResponse{
			Notifications: []types.Notification{},
			UnreadCount:   0,
			TotalCount:    0,
		}, nil
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get user ID: %v", err)
		return nil, err
	}

	// Set defaults
	pageSize := int64(20)
	if req.PageSize > 0 && req.PageSize <= 100 {
		pageSize = int64(req.PageSize)
	}
	page := 1
	if req.Page > 0 {
		page = req.Page
	}
	offset := int64((page - 1)) * pageSize

	var notifications []db.Notification
	if req.Unread {
		notifications, err = l.svcCtx.DB.Queries.ListUnreadNotifications(l.ctx, db.ListUnreadNotificationsParams{
			UserID:   userID.String(),
			PageSize: pageSize,
		})
	} else {
		notifications, err = l.svcCtx.DB.Queries.ListUserNotifications(l.ctx, db.ListUserNotificationsParams{
			UserID:     userID.String(),
			PageOffset: offset,
			PageSize:   pageSize,
		})
	}
	if err != nil {
		l.Errorf("Failed to list notifications: %v", err)
		return nil, err
	}

	// Get unread count
	unreadCount, err := l.svcCtx.DB.Queries.CountUnreadNotifications(l.ctx, userID.String())
	if err != nil {
		l.Errorf("Failed to count unread notifications: %v", err)
		return nil, err
	}

	// Convert to response type
	result := make([]types.Notification, len(notifications))
	for i, n := range notifications {
		result[i] = types.Notification{
			Id:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			Body:      n.Body.String,
			ActionUrl: n.ActionUrl.String,
			Icon:      n.Icon.String,
			CreatedAt: time.Unix(n.CreatedAt, 0).Format(time.RFC3339),
		}
		if n.ReadAt.Valid {
			result[i].ReadAt = time.Unix(n.ReadAt.Int64, 0).Format(time.RFC3339)
		}
	}

	return &types.ListNotificationsResponse{
		Notifications: result,
		UnreadCount:   int(unreadCount),
		TotalCount:    len(result),
	}, nil
}
