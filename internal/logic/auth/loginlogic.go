package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"gobot/internal/svc"
	"gobot/internal/types"

	levee "github.com/almatuck/levee-go"
	"github.com/golang-jwt/jwt/v4"
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
	// Check for admin login first (works in both modes)
	if req.Email == l.svcCtx.Config.Admin.Username {
		return l.loginAdmin(req)
	}

	// Use local auth when Levee is disabled
	if l.svcCtx.UseLocal() {
		return l.loginLocal(req)
	}

	// Use Levee when enabled
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	// Authenticate via Levee SDK
	authResp, err := l.svcCtx.Levee.Auth.Login(l.ctx, &levee.SDKLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		l.Errorf("Login failed for %s: %v", req.Email, err)
		return nil, err
	}

	// Parse expiry time
	expiresAt, _ := time.Parse(time.RFC3339, authResp.ExpiresAt)

	l.Infof("User logged in: %s", req.Email)

	return &types.LoginResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    expiresAt.UnixMilli(),
	}, nil
}

// loginLocal handles login with local SQLite auth
func (l *LoginLogic) loginLocal(req *types.LoginRequest) (*types.LoginResponse, error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("local auth service not configured")
	}

	authResp, err := l.svcCtx.Auth.Login(l.ctx, req.Email, req.Password)
	if err != nil {
		l.Errorf("Login failed for %s: %v", req.Email, err)
		return nil, err
	}

	l.Infof("User logged in (local): %s", req.Email)

	return &types.LoginResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt.UnixMilli(),
	}, nil
}

// loginAdmin handles admin login using credentials from env
// On first login, automatically creates the admin as a real user in the database
func (l *LoginLogic) loginAdmin(req *types.LoginRequest) (*types.LoginResponse, error) {
	// Constant-time comparison to prevent timing attacks
	expectedPass := []byte(l.svcCtx.Config.Admin.Password)
	providedPass := []byte(req.Password)

	if subtle.ConstantTimeCompare(expectedPass, providedPass) != 1 {
		l.Errorf("Admin login failed: invalid password")
		return nil, fmt.Errorf("invalid credentials")
	}

	// In standalone mode, create admin as a real user on first login
	if l.svcCtx.UseLocal() {
		return l.ensureAdminUserAndLogin(req)
	}

	// Levee mode: generate admin-only JWT (no database user)
	now := time.Now()
	accessExpiry := now.Add(time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second)

	claims := jwt.MapClaims{
		"userId": "admin",
		"email":  l.svcCtx.Config.Admin.Username,
		"iat":    now.Unix(),
		"exp":    accessExpiry.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		l.Errorf("Failed to sign admin token: %v", err)
		return nil, fmt.Errorf("failed to generate token")
	}

	l.Infof("Admin logged in: %s", req.Email)

	return &types.LoginResponse{
		Token:       accessToken,
		ExpiresAt:   accessExpiry.UnixMilli(),
		CheckoutUrl: "/backoffice",
	}, nil
}

// ensureAdminUserAndLogin creates the admin user if they don't exist, then logs them in
// The config password is the master admin password - if it validates, we trust the login
func (l *LoginLogic) ensureAdminUserAndLogin(req *types.LoginRequest) (*types.LoginResponse, error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("local auth service not configured")
	}

	// Check if admin user already exists
	user, err := l.svcCtx.Auth.GetUserByEmail(l.ctx, req.Email)
	if err != nil {
		// User doesn't exist - create them
		l.Infof("Creating admin user on first login: %s", req.Email)
		authResp, err := l.svcCtx.Auth.Register(l.ctx, req.Email, req.Password, "Admin")
		if err != nil {
			l.Errorf("Failed to create admin user: %v", err)
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}

		l.Infof("Admin user created and logged in: %s", req.Email)
		return &types.LoginResponse{
			Token:        authResp.Token,
			RefreshToken: authResp.RefreshToken,
			ExpiresAt:    authResp.ExpiresAt.UnixMilli(),
		}, nil
	}

	// User exists - password was already validated against config, generate tokens directly
	// This allows the config password to be the master admin password
	authResp, err := l.svcCtx.Auth.GenerateTokensForUser(l.ctx, user.ID, user.Email)
	if err != nil {
		l.Errorf("Failed to generate admin tokens: %v", err)
		return nil, err
	}

	l.Infof("Admin logged in: %s", req.Email)
	return &types.LoginResponse{
		Token:        authResp.Token,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt.UnixMilli(),
	}, nil
}
