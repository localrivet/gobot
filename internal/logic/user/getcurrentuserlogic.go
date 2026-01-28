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

type GetCurrentUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentUserLogic {
	return &GetCurrentUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentUserLogic) GetCurrentUser() (resp *types.GetUserResponse, err error) {
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	user, err := l.svcCtx.Auth.GetUserByEmail(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get user %s: %v", email, err)
		return nil, err
	}

	return &types.GetUserResponse{
		User: types.User{
			Id:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: user.EmailVerified == 1,
			CreatedAt:     time.Unix(user.CreatedAt, 0).Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     time.Unix(user.UpdatedAt, 0).Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
