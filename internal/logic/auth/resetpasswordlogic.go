package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	levee "github.com/almatuck/levee-go"
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
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("levee service not configured")
	}

	// Reset password via Levee SDK
	_, err = l.svcCtx.Levee.Auth.ResetPassword(l.ctx, &levee.SDKResetPasswordRequest{
		Token:           req.Token,
		Password:        req.NewPassword,
		ConfirmPassword: req.NewPassword,
	})
	if err != nil {
		l.Errorf("Reset password failed: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Password has been reset successfully.",
	}, nil
}
