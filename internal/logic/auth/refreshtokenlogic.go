package auth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	authResp, err := l.svcCtx.Auth.RefreshToken(l.ctx, req.RefreshToken)
	if err != nil {
		l.Errorf("Token refresh failed: %v", err)
		return nil, err
	}

	return &types.RefreshTokenResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt.UnixMilli(),
	}, nil
}
