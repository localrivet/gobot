package notification

import (
	"context"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadCountLogic {
	return &GetUnreadCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUnreadCountLogic) GetUnreadCount() (resp *types.GetUnreadCountResponse, err error) {
	// Check if notifications are enabled
	if !l.svcCtx.Config.IsNotificationsEnabled() {
		return &types.GetUnreadCountResponse{Count: 0}, nil
	}

	if !l.svcCtx.UseLocal() {
		return &types.GetUnreadCountResponse{Count: 0}, nil
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get user ID: %v", err)
		return nil, err
	}

	// Get unread count
	count, err := l.svcCtx.DB.Queries.CountUnreadNotifications(l.ctx, userID.String())
	if err != nil {
		l.Errorf("Failed to count unread notifications: %v", err)
		return nil, err
	}

	return &types.GetUnreadCountResponse{Count: int(count)}, nil
}
