package svc

import (
	"gobot/internal/config"
	"gobot/internal/db"
	"gobot/internal/local"
	"gobot/internal/middleware"

	leveeSDK "github.com/almatuck/levee-go"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config             config.Config
	SecurityMiddleware *middleware.SecurityMiddleware
	AdminAuth          rest.Middleware // Admin backoffice basic auth

	// Levee SDK (used when Levee.Enabled=true)
	Levee *leveeSDK.Client

	// Local services (used when Levee.Enabled=false)
	DB      *db.Store
	Auth    *local.AuthService
	Billing *local.BillingService
	Email   *local.EmailService // SMTP or Outlet.sh (available in both modes)
}

// NewServiceContext creates a new service context
func NewServiceContext(c config.Config) *ServiceContext {
	// Initialize security middleware
	securityMw := middleware.NewSecurityMiddleware(c)
	logx.Info("Security middleware initialized")

	// Initialize admin auth middleware (only needs username - password validated at login)
	adminAuth := middleware.NewAdminAuthMiddleware(c.Admin.Username)

	svc := &ServiceContext{
		Config:             c,
		SecurityMiddleware: securityMw,
		AdminAuth:          adminAuth.Handle,
	}

	// Initialize email service (SMTP or Outlet.sh - available in both modes)
	emailService := local.NewEmailService(c)
	if emailService.IsConfigured() {
		svc.Email = emailService
		logx.Info("Email service initialized")
	} else {
		logx.Info("Email not configured - transactional emails disabled")
	}

	if c.IsLeveeEnabled() && c.Levee.APIKey != "" {
		// Mode 1: Use Levee for auth, billing, email
		baseURL := c.Levee.BaseURL
		if baseURL == "" {
			baseURL = "https://api.levee.sh"
		}
		client, err := leveeSDK.NewClient(c.Levee.APIKey, baseURL)
		if err != nil {
			logx.Errorf("Failed to create Levee client: %v", err)
		} else {
			svc.Levee = client
			logx.Info("Levee SDK client initialized (using Levee for auth/billing)")
		}
	} else {
		// Mode 2: Use local SQLite + direct Stripe
		logx.Info("Levee disabled - using local SQLite + direct Stripe")

		// Initialize SQLite database
		database, err := db.NewSQLite(c.Database.SQLitePath)
		if err != nil {
			logx.Errorf("Failed to initialize SQLite database: %v", err)
		} else {
			svc.DB = database
			logx.Infof("SQLite database initialized at %s", c.Database.SQLitePath)

			// Initialize local services
			svc.Auth = local.NewAuthService(database, c)
			svc.Billing = local.NewBillingService(database, c)
			logx.Info("Local auth and billing services initialized")
		}
	}

	return svc
}

// Close closes any open connections
func (svc *ServiceContext) Close() {
	if svc.DB != nil {
		svc.DB.Close()
		logx.Info("SQLite database connection closed")
	}
	logx.Info("Service context closed")
}

// UseLevee returns true if Levee is enabled and configured
func (svc *ServiceContext) UseLevee() bool {
	return svc.Levee != nil
}

// UseLocal returns true if using local SQLite + Stripe
func (svc *ServiceContext) UseLocal() bool {
	return svc.DB != nil
}
