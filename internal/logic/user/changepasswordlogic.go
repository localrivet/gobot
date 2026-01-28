package user

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (resp *types.MessageResponse, err error) {
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
		l.Errorf("Failed to get user: %v", err)
		return nil, err
	}

	err = l.svcCtx.Auth.ChangePassword(l.ctx, user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		l.Errorf("Failed to change password for %s: %v", email, err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Password changed successfully.",
	}, nil
}
