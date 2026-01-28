package svc

import (
	"gobot/internal/agenthub"
	"gobot/internal/config"
	"gobot/internal/db"
	"gobot/internal/local"
	"gobot/internal/middleware"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config             config.Config
	SecurityMiddleware *middleware.SecurityMiddleware

	DB    *db.Store
	Auth  *local.AuthService
	Email *local.EmailService

	AgentHub *agenthub.Hub
}

func NewServiceContext(c config.Config) *ServiceContext {
	securityMw := middleware.NewSecurityMiddleware(c)
	logx.Info("Security middleware initialized")

	svc := &ServiceContext{
		Config:             c,
		SecurityMiddleware: securityMw,
		AgentHub:           agenthub.NewHub(),
	}

	emailService := local.NewEmailService(c)
	if emailService.IsConfigured() {
		svc.Email = emailService
		logx.Info("Email service initialized")
	} else {
		logx.Info("Email not configured - transactional emails disabled")
	}

	database, err := db.NewSQLite(c.Database.SQLitePath)
	if err != nil {
		logx.Errorf("Failed to initialize SQLite database: %v", err)
	} else {
		svc.DB = database
		logx.Infof("SQLite database initialized at %s", c.Database.SQLitePath)

		svc.Auth = local.NewAuthService(database, c)
		logx.Info("Auth service initialized")
	}

	return svc
}

func (svc *ServiceContext) Close() {
	if svc.DB != nil {
		svc.DB.Close()
		logx.Info("SQLite database connection closed")
	}
	logx.Info("Service context closed")
}

func (svc *ServiceContext) UseLocal() bool {
	return svc.DB != nil
}
