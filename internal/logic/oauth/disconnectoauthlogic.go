package oauth

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisconnectOAuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDisconnectOAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisconnectOAuthLogic {
	return &DisconnectOAuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DisconnectOAuthLogic) DisconnectOAuth(req *types.DisconnectOAuthRequest) (resp *types.MessageResponse, err error) {
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

	// Verify user has this OAuth connection
	_, err = l.svcCtx.DB.Queries.GetOAuthConnectionByUserAndProvider(l.ctx, db.GetOAuthConnectionByUserAndProviderParams{
		UserID:   userID.String(),
		Provider: req.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("OAuth provider %s is not connected", req.Provider)
	}

	// Get user to check if they have a password set
	user, err := l.svcCtx.Auth.GetUserByEmail(l.ctx, "")
	if err == nil && user != nil {
		// Check if user has other login methods
		connections, _ := l.svcCtx.DB.Queries.ListUserOAuthConnections(l.ctx, userID.String())
		hasPassword := user.PasswordHash != ""

		if !hasPassword && len(connections) <= 1 {
			return nil, fmt.Errorf("cannot disconnect your only login method; please set a password first")
		}
	}

	// Delete the OAuth connection
	err = l.svcCtx.DB.Queries.DeleteOAuthConnectionByProvider(l.ctx, db.DeleteOAuthConnectionByProviderParams{
		UserID:   userID.String(),
		Provider: req.Provider,
	})
	if err != nil {
		l.Errorf("Failed to disconnect OAuth: %v", err)
		return nil, err
	}

	l.Infof("User %s disconnected OAuth provider: %s", userID.String(), req.Provider)

	return &types.MessageResponse{
		Message: fmt.Sprintf("Successfully disconnected %s", req.Provider),
	}, nil
}
