package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	levee "github.com/almatuck/levee-go"
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
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("levee service not configured")
	}

	// Verify email via Levee SDK
	_, err = l.svcCtx.Levee.Auth.VerifyEmail(l.ctx, &levee.SDKVerifyEmailRequest{
		Token: req.Token,
	})
	if err != nil {
		l.Errorf("Email verification failed: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Email verified successfully.",
	}, nil
}
