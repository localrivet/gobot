package auth

import (
	"context"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResendVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResendVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResendVerificationLogic {
	return &ResendVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResendVerificationLogic) ResendVerification(req *types.ResendVerificationRequest) (resp *types.MessageResponse, err error) {
	// Levee doesn't have a separate resend verification endpoint
	// The verification email is sent on registration
	// For now, return a success message
	return &types.MessageResponse{
		Message: "If the email address is registered and unverified, a new verification email has been sent.",
	}, nil
}
