package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyEmailLogic {
	return &VerifyEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyEmailLogic) VerifyEmail(req *types.EmailVerificationRequest) (resp *types.MessageResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	err = l.svcCtx.Auth.VerifyEmail(l.ctx, req.Token)
	if err != nil {
		l.Errorf("Email verification failed: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Email verified successfully.",
	}, nil
}
