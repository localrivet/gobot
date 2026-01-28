package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordRequest) (resp *types.MessageResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	err = l.svcCtx.Auth.ResetPassword(l.ctx, req.Token, req.NewPassword)
	if err != nil {
		l.Errorf("Reset password failed: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Password has been reset successfully.",
	}, nil
}
