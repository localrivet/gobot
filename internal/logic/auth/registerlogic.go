package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.LoginResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	authResp, err := l.svcCtx.Auth.Register(l.ctx, req.Email, req.Password, req.Name)
	if err != nil {
		l.Errorf("Registration failed for %s: %v", req.Email, err)
		return nil, err
	}

	l.Infof("User registered: %s", req.Email)

	return &types.LoginResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt.UnixMilli(),
	}, nil
}
