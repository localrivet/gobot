package user

import (
	"context"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCurrentUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCurrentUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCurrentUserLogic {
	return &UpdateCurrentUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCurrentUserLogic) UpdateCurrentUser(req *types.UpdateUserRequest) (resp *types.GetUserResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	user, err := l.svcCtx.Auth.GetUserByEmail(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get user %s: %v", email, err)
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	err = l.svcCtx.Auth.UpdateUser(l.ctx, user)
	if err != nil {
		l.Errorf("Failed to update user %s: %v", email, err)
		return nil, err
	}

	return &types.GetUserResponse{
		User: types.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: time.Unix(user.CreatedAt, 0).Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}
