package user

import (
	"context"
	"errors"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAccountLogic {
	return &DeleteAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAccountLogic) DeleteAccount(req *types.DeleteAccountRequest) (resp *types.MessageResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	_, err = l.svcCtx.Auth.Login(l.ctx, email, req.Password)
	if err != nil {
		l.Errorf("Password verification failed for delete account: %v", err)
		return nil, errors.New("invalid password")
	}

	user, err := l.svcCtx.Auth.GetUserByEmail(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get user %s: %v", email, err)
		return nil, err
	}

	err = l.svcCtx.Auth.DeleteUser(l.ctx, user.ID)
	if err != nil {
		l.Errorf("Failed to delete user %s: %v", email, err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Account deleted successfully.",
	}, nil
}
