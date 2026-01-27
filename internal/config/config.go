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
		ProductionMode string `json:",default=false"` // "true" to enable
		AdminEmail     string `json:",optional"`      // For Let's Encrypt notifications
	}
	Admin struct {
		Username string `json:",optional"` // Backoffice admin username
		Password string `json:",optional"` // Backoffice admin password
	}
	Auth struct {
		AccessSecret       string
		AccessExpire       int64
		RefreshTokenExpire int64 `json:",default=604800"` // 7 days in seconds
	}
	Database struct {
		// SQLite (used when Levee is disabled)
		SQLitePath string `json:",default=./data/gobot.db"`

		// PostgreSQL (optional, for scaling later)
		Host            string `json:",default=localhost"`
		Port            int    `json:",default=5432"`
		User            string `json:",default=postgres"`
		Password        string `json:",optional"`
		DBName          string `json:",default=gobot"`
		SSLMode         string `json:",default=disable"`
		MaxOpenConns    int    `json:",default=25"`
		MaxIdleConns    int    `json:",default=5"`
		ConnMaxLifetime int    `json:",default=300"` // seconds
	}
	Stripe struct {
		SecretKey      string `json:",optional"` // sk_test_xxx or sk_live_xxx
		PublishableKey string `json:",optional"` // pk_test_xxx or pk_live_xxx
		WebhookSecret  string `json:",optional"` // whsec_xxx
		SuccessURL     string `json:",default=/app/account"`
		CancelURL      string `json:",default=/pricing"`
	}
	// Products defines subscription products for standalone mode (without Levee)
	// These are synced to Stripe on startup. See docs/stripe.md for configuration details.
	Products []Product `json:",optional"`
	Security struct {
		// CSRF protection settings
		CSRFEnabled      string `json:",default=true"` // "true" to enable
		CSRFSecret       string `json:",optional"`     // If empty, uses Auth.AccessSecret
		CSRFTokenExpiry  int64  `json:",default=43200"` // 12 hours in seconds
		CSRFSecureCookie string `json:",default=true"` // "true" for secure cookies

		// Rate limiting settings
		RateLimitEnabled      string `json:",default=true"` // "true" to enable
		RateLimitRequests     int    `json:",default=100"`  // requests per interval
		RateLimitInterval     int    `json:",default=60"`   // interval in seconds
		RateLimitBurst        int    `json:",default=20"`   // burst size
		AuthRateLimitRequests int    `json:",default=5"`    // auth endpoints rate limit
		AuthRateLimitInterval int    `json:",default=60"`   // auth rate limit interval

		// Security headers settings
		EnableSecurityHeaders string `json:",default=true"` // "true" to enable
		ContentSecurityPolicy string `json:",optional"`     // Override default CSP
		AllowedOrigins        string `json:",optional"`     // CORS allowed origins (comma-separated)

		// HTTPS enforcement
		ForceHTTPS string `json:",default=false"` // "true" in production

		// Input validation
		MaxRequestBodySize int64 `json:",default=10485760"` // 10MB default
		MaxURLLength       int   `json:",default=2048"`
	}
	Email struct {
		SMTPHost    string `json:",optional"`
		SMTPPort    int    `json:",optional,default=587"`
		SMTPUser    string `json:",optional"`
		SMTPPass    string `json:",optional"`
		FromAddress string `json:",optional"`
		FromName    string `json:",default=gobot"`
		ReplyTo     string `json:",optional"`
		BaseURL     string `json:",default=http://localhost:27458"` // For links in emails
	}
	Analytics struct {
		Enabled       string `json:",default=true"`    // "true" to enable
		Provider      string `json:",default=console"` // Provider: segment, mixpanel, posthog, console
		APIKey        string `json:",optional"`        // Provider API key
		Endpoint      string `json:",optional"`        // Custom endpoint URL
		BatchSize     int    `json:",default=50"`      // Events per batch
		FlushInterval int    `json:",default=30"`      // Flush interval in seconds
		Debug         string `json:",default=false"`   // "true" for verbose logging
	}
	Subscription struct {
		Enabled             string `json:",default=true"` // "true" to enable
		EnforceQuotas       string `json:",default=true"` // "true" to enforce
		FreeTierAnalyses    int    `json:",default=5"`    // Free tier monthly analysis limit
		FreeTierHistoryDays int    `json:",default=7"`    // Free tier history retention days
		ProTierHistoryDays  int    `json:",default=30"`   // Pro tier history retention days
		TeamTierHistoryDays int    `json:",default=-1"`   // Team tier history retention (-1 = unlimited)
	}
	AI struct {
		Enabled     string `json:",default=true"`                       // "true" to enable
		APIKey      string `json:",optional"`                           // Anthropic API key
		Model       string `json:",default=claude-sonnet-4-5-20250929"` // Claude model to use
		MaxTokens   int    `json:",default=4096"`                       // Max tokens per response
		TimeoutSecs int    `json:",default=60"`                         // Request timeout in seconds
	}
	CostAlerts struct {
		Enabled            string  `json:",default=true"` // "true" to enable
		AdminEmail         string  `json:",optional"`     // Email to receive cost alerts
		DailyCostThreshold float64 `json:",default=10.0"` // Alert if daily costs exceed this USD amount
		UserCostThreshold  float64 `json:",default=1.0"`  // Alert if single user's daily costs exceed this
	}
	Outlet struct {
		BaseURL string `json:",optional"` // Outlet.sh instance URL (e.g., https://email.yourdomain.com)
		APIKey  string `json:",optional"` // Outlet.sh API key

		// List slugs for subscriber management
		WaitlistListSlug   string `json:",default=waitlist"`
		NewsletterListSlug string `json:",default=newsletter"`
		UsersListSlug      string `json:",default=users"`

		// Sequence slugs for email automation
		OnboardingSequence      string `json:",default=onboarding"`
		TrialConversionSequence string `json:",default=trial-conversion"`

		// Template slugs for transactional emails
		EmailVerificationTemplate string `json:",default=email-verification"`
		PasswordResetTemplate     string `json:",default=password-reset"`
		WelcomeTemplate           string `json:",default=welcome"`
	}
	Levee struct {
		APIKey      string `json:",optional"`      // Levee API key
		BaseURL     string `json:",optional"`      // Custom API endpoint (optional)
		GRPCAddress string `json:",optional"`      // gRPC endpoint for LLM streaming (e.g., levee.localrivet.com:9889)
		Enabled     string `json:",default=true"`  // "true" to enable

		// Checkout redirect URLs
		CheckoutSuccessURL string `json:",default=/app/billing/success"` // Redirect after successful checkout
		CheckoutCancelURL  string `json:",default=/pricing"`             // Redirect after cancelled checkout

		// List slugs
		WaitlistListSlug    string `json:",default=waitlist"`
		NewsletterListSlug  string `json:",default=newsletter-subscribers"`
		UsersListSlug       string `json:",default=users"`
		ProductHuntListSlug string `json:",default=product-hunt-launchers"`

		// Sequence slugs
		WaitlistNurtureSequence string `json:",default=waitlist-nurture"`
		OnboardingSequence      string `json:",default=onboarding-sequence"`
		TrialConversionSequence string `json:",default=free-to-pro-upgrade"`

		// Product slugs for checkout (these map to Levee price nicknames)
		FreeProductSlug        string `json:",default=free"`
		ProMonthlyProductSlug  string `json:",default=pro"`
		ProYearlyProductSlug   string `json:",default=pro-yearly"`
		TeamMonthlyProductSlug string `json:",default=team"`
		TeamYearlyProductSlug  string `json:",default=team-yearly"`

		// Template slugs for transactional emails
		EmailVerificationTemplate     string `json:",default=email-verification"`
		PasswordResetTemplate         string `json:",default=password-reset"`
		PasswordChangedTemplate       string `json:",default=password-changed"`
		AccountDeletedTemplate        string `json:",default=account-deleted"`
		WaitlistConfirmTemplate       string `json:",default=waitlist-confirm"`
		SubscriptionWelcomeTemplate   string `json:",default=subscription-welcome"`
		SubscriptionRenewalTemplate   string `json:",default=subscription-renewal"`
		SubscriptionCancelledTemplate string `json:",default=subscription-cancelled"`
		PaymentFailedTemplate         string `json:",default=payment-failed"`
		TrialEndingTemplate           string `json:",default=trial-ending"`
		CostAlertTemplate             string `json:",default=cost-alert"`
	}
	OAuth struct {
		// Google OAuth
		GoogleEnabled      string `json:",default=false"` // "true" to enable
		GoogleClientID     string `json:",optional"`
		GoogleClientSecret string `json:",optional"`

		// GitHub OAuth
		GitHubEnabled      string `json:",default=false"` // "true" to enable
		GitHubClientID     string `json:",optional"`
		GitHubClientSecret string `json:",optional"`

		// Callback URL base (defaults to App.BaseURL)
		CallbackBaseURL string `json:",optional"`
	}
	Features struct {
		// Enable/disable features
		OrganizationsEnabled string `json:",default=true"`  // "true" to enable
		NotificationsEnabled string `json:",default=true"`  // "true" to enable
		OAuthEnabled         string `json:",default=false"` // "true" to enable
	}
}

// ========== Helper Methods ==========

// IsProductionMode returns true if production mode is enabled.
func (c Config) IsProductionMode() bool {
	return parseBool(c.App.ProductionMode, false)
}

// IsCSRFEnabled returns true if CSRF protection is enabled.
func (c Config) IsCSRFEnabled() bool {
	return parseBool(c.Security.CSRFEnabled, true)
}

// IsCSRFSecureCookie returns true if CSRF cookies should be secure.
func (c Config) IsCSRFSecureCookie() bool {
	return parseBool(c.Security.CSRFSecureCookie, true)
}

// IsRateLimitEnabled returns true if rate limiting is enabled.
func (c Config) IsRateLimitEnabled() bool {
	return parseBool(c.Security.RateLimitEnabled, true)
}

// IsSecurityHeadersEnabled returns true if security headers are enabled.
func (c Config) IsSecurityHeadersEnabled() bool {
	return parseBool(c.Security.EnableSecurityHeaders, true)
}

// IsForceHTTPS returns true if HTTPS should be enforced.
func (c Config) IsForceHTTPS() bool {
	return parseBool(c.Security.ForceHTTPS, false)
}

// IsAnalyticsEnabled returns true if analytics is enabled.
func (c Config) IsAnalyticsEnabled() bool {
	return parseBool(c.Analytics.Enabled, true)
}

// IsAnalyticsDebug returns true if analytics debug mode is enabled.
func (c Config) IsAnalyticsDebug() bool {
	return parseBool(c.Analytics.Debug, false)
}

// IsSubscriptionEnabled returns true if subscription features are enabled.
func (c Config) IsSubscriptionEnabled() bool {
	return parseBool(c.Subscription.Enabled, true)
}

// IsEnforceQuotas returns true if usage quotas should be enforced.
func (c Config) IsEnforceQuotas() bool {
	return parseBool(c.Subscription.EnforceQuotas, true)
}

// IsAIEnabled returns true if AI features are enabled.
func (c Config) IsAIEnabled() bool {
	return parseBool(c.AI.Enabled, true)
}

// IsCostAlertsEnabled returns true if cost alerts are enabled.
func (c Config) IsCostAlertsEnabled() bool {
	return parseBool(c.CostAlerts.Enabled, true)
}

// IsLeveeEnabled returns true if Levee integration is enabled.
func (c Config) IsLeveeEnabled() bool {
	return parseBool(c.Levee.Enabled, true)
}

// IsGoogleOAuthEnabled returns true if Google OAuth is enabled.
func (c Config) IsGoogleOAuthEnabled() bool {
	return parseBool(c.OAuth.GoogleEnabled, false)
}

// IsGitHubOAuthEnabled returns true if GitHub OAuth is enabled.
func (c Config) IsGitHubOAuthEnabled() bool {
	return parseBool(c.OAuth.GitHubEnabled, false)
}

// IsOrganizationsEnabled returns true if organizations feature is enabled.
func (c Config) IsOrganizationsEnabled() bool {
	return parseBool(c.Features.OrganizationsEnabled, true)
}

// IsNotificationsEnabled returns true if notifications feature is enabled.
func (c Config) IsNotificationsEnabled() bool {
	return parseBool(c.Features.NotificationsEnabled, true)
}

// IsOAuthEnabled returns true if OAuth feature is enabled.
func (c Config) IsOAuthEnabled() bool {
	return parseBool(c.Features.OAuthEnabled, false)
}

// Product defines a subscription product with its prices
// Products are synced to Stripe on startup (standalone mode only)
type Product struct {
	// Slug is the unique identifier for this product (e.g., "free", "pro", "team")
	// Used in checkout URLs and API calls
	Slug string `json:"slug"`

	// Name is the display name shown to customers (e.g., "Pro Plan")
	Name string `json:"name"`

	// Description is shown on pricing page and Stripe checkout
	Description string `json:"description,optional"`

	// Features is a list of features included in this plan
	// Displayed on pricing pages and used for feature gating
	Features []string `json:"features,optional"`

	// Prices defines the pricing options for this product
	// A product can have multiple prices (e.g., monthly and yearly)
	Prices []Price `json:"prices,optional"`

	// Default marks this as the default plan for new users (only one should be true)
	Default bool `json:"default,optional"`

	// StripeProductID is auto-populated after syncing to Stripe
	StripeProductID string `json:"stripeProductId,optional"`
}

// Price defines a single pricing option for a product
type Price struct {
	// Slug identifies this price (e.g., "monthly", "yearly")
	Slug string `json:"slug"`

	// Amount in cents (e.g., 2900 for $29.00, 0 for free)
	Amount int64 `json:"amount"`

	// Currency code (default: "usd")
	Currency string `json:"currency,default=usd"`

	// Interval: "month", "year", or "one_time"
	Interval string `json:"interval,default=month"`

	// IntervalCount: billing frequency (default: 1)
	// e.g., interval=month + intervalCount=3 = quarterly billing
	IntervalCount int64 `json:"intervalCount,default=1"`

	// TrialDays: number of trial days (0 = no trial)
	TrialDays int64 `json:"trialDays,optional"`

	// StripePriceID is auto-populated after syncing to Stripe
	StripePriceID string `json:"stripePriceId,optional"`
}
