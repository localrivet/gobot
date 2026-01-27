package oauth

import (
	"context"
	"fmt"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OAuthCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOAuthCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OAuthCallbackLogic {
	return &OAuthCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OAuthCallback is deprecated - OAuth callbacks are handled directly at /oauth/{provider}/callback
// This endpoint exists for API compatibility but should not be called directly.
func (l *OAuthCallbackLogic) OAuthCallback(req *types.OAuthLoginRequest) (resp *types.OAuthLoginResponse, err error) {
	return nil, fmt.Errorf("OAuth callbacks should use /oauth/%s/callback (browser redirect), not the API endpoint", req.Provider)
}
