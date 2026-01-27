package auth

import (
	"context"
	"fmt"
	"time"

	"gobot/internal/svc"
	"gobot/internal/types"

	levee "github.com/almatuck/levee-go"
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
	// Use local auth when Levee is disabled
	if l.svcCtx.UseLocal() {
		return l.refreshTokenLocal(req)
	}

	// Use Levee when enabled
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	// Refresh token via Levee SDK
	authResp, err := l.svcCtx.Levee.Auth.RefreshToken(l.ctx, &levee.SDKRefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		l.Errorf("Token refresh failed: %v", err)
		return nil, err
	}

	// Parse expiry time
	expiresAt, _ := time.Parse(time.RFC3339, authResp.ExpiresAt)

	return &types.RefreshTokenResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    expiresAt.UnixMilli(),
	}, nil
}

// refreshTokenLocal handles token refresh with local SQLite auth
func (l *RefreshTokenLogic) refreshTokenLocal(req *types.RefreshTokenRequest) (*types.RefreshTokenResponse, error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("local auth service not configured")
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
