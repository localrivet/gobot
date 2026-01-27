package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOAuthUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOAuthUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOAuthUrlLogic {
	return &GetOAuthUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOAuthUrlLogic) GetOAuthUrl(req *types.GetOAuthUrlRequest) (resp *types.GetOAuthUrlResponse, err error) {
	if !l.svcCtx.Config.IsOAuthEnabled() {
		return nil, fmt.Errorf("OAuth feature is not enabled")
	}

	if !l.svcCtx.UseLocal() {
		return nil, fmt.Errorf("OAuth not available in this mode")
	}

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		l.Errorf("Failed to generate state: %v", err)
		return nil, err
	}

	// Determine callback URL base
	callbackBase := l.svcCtx.Config.OAuth.CallbackBaseURL
	if callbackBase == "" {
		callbackBase = l.svcCtx.Config.App.BaseURL
	}

	var authURL string

	switch req.Provider {
	case "google":
		if !l.svcCtx.Config.IsGoogleOAuthEnabled() {
			return nil, fmt.Errorf("Google OAuth is not enabled")
		}
		authURL = buildGoogleAuthURL(
			l.svcCtx.Config.OAuth.GoogleClientID,
			callbackBase+"/oauth/google/callback",
			state,
		)
	case "github":
		if !l.svcCtx.Config.IsGitHubOAuthEnabled() {
			return nil, fmt.Errorf("GitHub OAuth is not enabled")
		}
		authURL = buildGitHubAuthURL(
			l.svcCtx.Config.OAuth.GitHubClientID,
			callbackBase+"/oauth/github/callback",
			state,
		)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", req.Provider)
	}

	return &types.GetOAuthUrlResponse{
		Url:   authURL,
		State: state,
	}, nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildGoogleAuthURL(clientID, redirectURI, state string) string {
	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func buildGitHubAuthURL(clientID, redirectURI, state string) string {
	params := url.Values{
		"client_id":    {clientID},
		"redirect_uri": {redirectURI},
		"scope":        {"user:email"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}
