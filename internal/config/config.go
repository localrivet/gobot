package config

import (
	"strings"

	"github.com/zeromicro/go-zero/rest"
)

// parseBool parses a string as boolean with a default value.
// Accepts: "true", "1", "yes" as true; empty or other values return default.
func parseBool(s string, defaultVal bool) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return defaultVal
	}
	return s == "true" || s == "1" || s == "yes"
}

type Config struct {
	rest.RestConf
	App struct {
		BaseURL        string `json:",optional"`
		Domain         string `json:",optional"`
		ProductionMode string `json:",default=false"`
		AdminEmail     string `json:",optional"`
	}
	Auth struct {
		AccessSecret       string
		AccessExpire       int64
		RefreshTokenExpire int64 `json:",default=604800"`
	}
	Database struct {
		SQLitePath      string `json:",default=./data/gobot.db"`
		Host            string `json:",default=localhost"`
		Port            int    `json:",default=5432"`
		User            string `json:",default=postgres"`
		Password        string `json:",optional"`
		DBName          string `json:",default=gobot"`
		SSLMode         string `json:",default=disable"`
		MaxOpenConns    int    `json:",default=25"`
		MaxIdleConns    int    `json:",default=5"`
		ConnMaxLifetime int    `json:",default=300"`
	}
	Security struct {
		CSRFEnabled           string `json:",default=true"`
		CSRFSecret            string `json:",optional"`
		CSRFTokenExpiry       int64  `json:",default=43200"`
		CSRFSecureCookie      string `json:",default=true"`
		RateLimitEnabled      string `json:",default=true"`
		RateLimitRequests     int    `json:",default=100"`
		RateLimitInterval     int    `json:",default=60"`
		RateLimitBurst        int    `json:",default=20"`
		AuthRateLimitRequests int    `json:",default=5"`
		AuthRateLimitInterval int    `json:",default=60"`
		EnableSecurityHeaders string `json:",default=true"`
		ContentSecurityPolicy string `json:",optional"`
		AllowedOrigins        string `json:",optional"`
		ForceHTTPS            string `json:",default=false"`
		MaxRequestBodySize    int64  `json:",default=10485760"`
		MaxURLLength          int    `json:",default=2048"`
	}
	Email struct {
		SMTPHost    string `json:",optional"`
		SMTPPort    int    `json:",optional,default=587"`
		SMTPUser    string `json:",optional"`
		SMTPPass    string `json:",optional"`
		FromAddress string `json:",optional"`
		FromName    string `json:",default=gobot"`
		ReplyTo     string `json:",optional"`
		BaseURL     string `json:",default=http://localhost:27458"`
	}
	Analytics struct {
		Enabled       string `json:",default=true"`
		Provider      string `json:",default=console"`
		APIKey        string `json:",optional"`
		Endpoint      string `json:",optional"`
		BatchSize     int    `json:",default=50"`
		FlushInterval int    `json:",default=30"`
		Debug         string `json:",default=false"`
	}
	AI struct {
		Enabled     string `json:",default=true"`
		APIKey      string `json:",optional"`
		Model       string `json:",default=claude-sonnet-4-5-20250929"`
		MaxTokens   int    `json:",default=4096"`
		TimeoutSecs int    `json:",default=60"`
	}
	OAuth struct {
		GoogleEnabled      string `json:",default=false"`
		GoogleClientID     string `json:",optional"`
		GoogleClientSecret string `json:",optional"`
		GitHubEnabled      string `json:",default=false"`
		GitHubClientID     string `json:",optional"`
		GitHubClientSecret string `json:",optional"`
		CallbackBaseURL    string `json:",optional"`
	}
	Features struct {
		OrganizationsEnabled string `json:",default=true"`
		NotificationsEnabled string `json:",default=true"`
		OAuthEnabled         string `json:",default=false"`
	}
}

func (c Config) IsProductionMode() bool {
	return parseBool(c.App.ProductionMode, false)
}

func (c Config) IsCSRFEnabled() bool {
	return parseBool(c.Security.CSRFEnabled, true)
}

func (c Config) IsCSRFSecureCookie() bool {
	return parseBool(c.Security.CSRFSecureCookie, true)
}

func (c Config) IsRateLimitEnabled() bool {
	return parseBool(c.Security.RateLimitEnabled, true)
}

func (c Config) IsSecurityHeadersEnabled() bool {
	return parseBool(c.Security.EnableSecurityHeaders, true)
}

func (c Config) IsForceHTTPS() bool {
	return parseBool(c.Security.ForceHTTPS, false)
}

func (c Config) IsAnalyticsEnabled() bool {
	return parseBool(c.Analytics.Enabled, true)
}

func (c Config) IsAnalyticsDebug() bool {
	return parseBool(c.Analytics.Debug, false)
}

func (c Config) IsAIEnabled() bool {
	return parseBool(c.AI.Enabled, true)
}

func (c Config) IsGoogleOAuthEnabled() bool {
	return parseBool(c.OAuth.GoogleEnabled, false)
}

func (c Config) IsGitHubOAuthEnabled() bool {
	return parseBool(c.OAuth.GitHubEnabled, false)
}

func (c Config) IsOrganizationsEnabled() bool {
	return parseBool(c.Features.OrganizationsEnabled, true)
}

func (c Config) IsNotificationsEnabled() bool {
	return parseBool(c.Features.NotificationsEnabled, true)
}

func (c Config) IsOAuthEnabled() bool {
	return parseBool(c.Features.OAuthEnabled, false)
}
