package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	authResp, err := l.svcCtx.Auth.Login(l.ctx, req.Email, req.Password)
	if err != nil {
		l.Errorf("Login failed for %s: %v", req.Email, err)
		return nil, err
	}

	l.Infof("User logged in: %s", req.Email)

	return &types.LoginResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt.UnixMilli(),
	}, nil
}
