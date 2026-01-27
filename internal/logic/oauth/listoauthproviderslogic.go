package oauth

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOAuthProvidersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOAuthProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOAuthProvidersLogic {
	return &ListOAuthProvidersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOAuthProvidersLogic) ListOAuthProviders() (resp *types.ListOAuthProvidersResponse, err error) {
	if !l.svcCtx.Config.IsOAuthEnabled() {
		return nil, fmt.Errorf("OAuth feature is not enabled")
	}

	if !l.svcCtx.UseLocal() {
		return nil, fmt.Errorf("OAuth not available in this mode")
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get user ID: %v", err)
		return nil, err
	}

	// Get user's OAuth connections
	connections, err := l.svcCtx.DB.Queries.ListUserOAuthConnections(l.ctx, userID.String())
	if err != nil {
		l.Errorf("Failed to list OAuth connections: %v", err)
		return nil, err
	}

	// Build map of connected providers
	connectedProviders := make(map[string]string)
	for _, conn := range connections {
		connectedProviders[conn.Provider] = conn.Email.String
	}

	// Build provider list
	var providers []types.OAuthProvider

	// Google
	if l.svcCtx.Config.IsGoogleOAuthEnabled() {
		email := connectedProviders["google"]
		providers = append(providers, types.OAuthProvider{
			Name:      "google",
			Connected: email != "",
			Email:     email,
		})
	}

	// GitHub
	if l.svcCtx.Config.IsGitHubOAuthEnabled() {
		email := connectedProviders["github"]
		providers = append(providers, types.OAuthProvider{
			Name:      "github",
			Connected: email != "",
			Email:     email,
		})
	}

	return &types.ListOAuthProvidersResponse{
		Providers: providers,
	}, nil
}
