package auth

import (
	"context"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuthConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuthConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthConfigLogic {
	return &GetAuthConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuthConfigLogic) GetAuthConfig() (resp *types.AuthConfigResponse, err error) {
	// Return OAuth provider configuration
	// Only return enabled status if OAuth feature is enabled and in local mode
	googleEnabled := false
	githubEnabled := false

	if l.svcCtx.UseLocal() && l.svcCtx.Config.IsOAuthEnabled() {
		googleEnabled = l.svcCtx.Config.IsGoogleOAuthEnabled()
		githubEnabled = l.svcCtx.Config.IsGitHubOAuthEnabled()
	}

	return &types.AuthConfigResponse{
		GoogleEnabled: googleEnabled,
		GitHubEnabled: githubEnabled,
	}, nil
}
