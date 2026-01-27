package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	// ErrInvalidCSRFToken is returned when the CSRF token is invalid
	ErrInvalidCSRFToken = errors.New("invalid or missing CSRF token")
	// ErrCSRFTokenExpired is returned when the CSRF token has expired
	ErrCSRFTokenExpired = errors.New("CSRF token expired")
)

// CSRFConfig holds the configuration for CSRF protection
type CSRFConfig struct {
	// Secret is the key used to sign tokens (must be at least 32 bytes)
	Secret []byte
	// TokenLength is the length of the random token in bytes (default: 32)
	TokenLength int
	// TokenExpiry is how long tokens are valid (default: 12 hours)
	TokenExpiry time.Duration
	// CookieName is the name of the CSRF cookie (default: "csrf_token")
	CookieName string
	// HeaderName is the name of the CSRF header (default: "X-CSRF-Token")
	HeaderName string
	// FormFieldName is the name of the CSRF form field (default: "csrf_token")
	FormFieldName string
	// SecureCookie sets the Secure flag on the cookie (default: true in production)
	SecureCookie bool
	// HTTPOnly sets the HttpOnly flag on the cookie (default: false for JS access)
	HTTPOnly bool
	// SameSite sets the SameSite attribute (default: Strict)
	SameSite http.SameSite
	// CookiePath sets the cookie path (default: "/")
	CookiePath string
	// CookieDomain sets the cookie domain (optional)
	CookieDomain string
	// SkipPaths are paths that skip CSRF validation (e.g., public API endpoints)
	SkipPaths []string
	// SkipMethods are HTTP methods that skip CSRF validation (default: GET, HEAD, OPTIONS, TRACE)
	SkipMethods []string
	// ErrorHandler is called when CSRF validation fails
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig(secret []byte) *CSRFConfig {
	return &CSRFConfig{
		Secret:        secret,
		TokenLength:   32,
		TokenExpiry:   12 * time.Hour,
		CookieName:    "csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "csrf_token",
		SecureCookie:  true,
		HTTPOnly:      false, // Allow JS to read for SPA
		SameSite:      http.SameSiteStrictMode,
		CookiePath:    "/",
		SkipMethods:   []string{"GET", "HEAD", "OPTIONS", "TRACE"},
		ErrorHandler:  defaultCSRFErrorHandler,
	}
}

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFProtection provides CSRF protection for HTTP handlers
type CSRFProtection struct {
	config *CSRFConfig
	tokens sync.Map // In-memory token store for validation
}

// NewCSRFProtection creates a new CSRF protection instance
func NewCSRFProtection(config *CSRFConfig) *CSRFProtection {
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.TokenExpiry == 0 {
		config.TokenExpiry = 12 * time.Hour
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "csrf_token"
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if len(config.SkipMethods) == 0 {
		config.SkipMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCSRFErrorHandler
	}

	csrf := &CSRFProtection{
		config: config,
	}

	// Start cleanup goroutine
	go csrf.cleanupExpiredTokens()

	return csrf
}

// GenerateToken creates a new CSRF token
func (c *CSRFProtection) GenerateToken() (*CSRFToken, error) {
	// Generate random bytes
	randomBytes := make([]byte, c.config.TokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	// Create token string
	token := base64.URLEncoding.EncodeToString(randomBytes)
	expiresAt := time.Now().Add(c.config.TokenExpiry)

	// Create signature
	signedToken := c.signToken(token, expiresAt)

	// Store token hash for validation
	tokenHash := c.hashToken(signedToken)
	c.tokens.Store(tokenHash, expiresAt)

	return &CSRFToken{
		Token:     signedToken,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateToken validates a CSRF token
func (c *CSRFProtection) ValidateToken(token string) error {
	if token == "" {
		return ErrInvalidCSRFToken
	}

	tokenHash := c.hashToken(token)

	// Check if token exists in store
	if expiresAt, ok := c.tokens.Load(tokenHash); ok {
		if time.Now().After(expiresAt.(time.Time)) {
			c.tokens.Delete(tokenHash)
			return ErrCSRFTokenExpired
		}
		return nil
	}

	// Verify token signature
	if !c.verifyTokenSignature(token) {
		return ErrInvalidCSRFToken
	}

	return nil
}

// signToken creates a signed token
func (c *CSRFProtection) signToken(token string, expiresAt time.Time) string {
	// Create message: token|expiry_unix
	message := token + "|" + expiresAt.Format(time.RFC3339)

	// Create HMAC signature
	h := sha256.New()
	h.Write(c.config.Secret)
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	// Return signed token: message|signature
	return message + "|" + signature
}

// verifyTokenSignature verifies the token signature
func (c *CSRFProtection) verifyTokenSignature(signedToken string) bool {
	parts := strings.Split(signedToken, "|")
	if len(parts) != 3 {
		return false
	}

	token := parts[0]
	expiryStr := parts[1]
	providedSig := parts[2]

	// Parse and check expiry
	expiresAt, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return false
	}
	if time.Now().After(expiresAt) {
		return false
	}

	// Recreate signature
	message := token + "|" + expiryStr
	h := sha256.New()
	h.Write(c.config.Secret)
	h.Write([]byte(message))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	// Constant-time comparison
	return subtle.ConstantTimeCompare([]byte(providedSig), []byte(expectedSig)) == 1
}

// hashToken creates a hash of the token for storage
func (c *CSRFProtection) hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// cleanupExpiredTokens periodically removes expired tokens
func (c *CSRFProtection) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.tokens.Range(func(key, value interface{}) bool {
			if expiresAt, ok := value.(time.Time); ok {
				if now.After(expiresAt) {
					c.tokens.Delete(key)
				}
			}
			return true
		})
	}
}

// Middleware returns an HTTP middleware that enforces CSRF protection
func (c *CSRFProtection) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if method should skip CSRF
			if c.shouldSkipMethod(r.Method) {
				// For safe methods, still set a token cookie if not present
				if _, err := r.Cookie(c.config.CookieName); err == http.ErrNoCookie {
					c.setTokenCookie(w)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Check if path should skip CSRF
			if c.shouldSkipPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from request (header or form)
			token := c.getTokenFromRequest(r)
			if token == "" {
				c.config.ErrorHandler(w, r, ErrInvalidCSRFToken)
				return
			}

			// Validate token
			if err := c.ValidateToken(token); err != nil {
				c.config.ErrorHandler(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// setTokenCookie sets a new CSRF token cookie
func (c *CSRFProtection) setTokenCookie(w http.ResponseWriter) {
	token, err := c.GenerateToken()
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     c.config.CookieName,
		Value:    token.Token,
		Path:     c.config.CookiePath,
		Domain:   c.config.CookieDomain,
		Expires:  token.ExpiresAt,
		Secure:   c.config.SecureCookie,
		HttpOnly: c.config.HTTPOnly,
		SameSite: c.config.SameSite,
	})
}

// getTokenFromRequest extracts the CSRF token from the request
func (c *CSRFProtection) getTokenFromRequest(r *http.Request) string {
	// Check header first
	if token := r.Header.Get(c.config.HeaderName); token != "" {
		return token
	}

	// Check form field
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			if token := r.FormValue(c.config.FormFieldName); token != "" {
				return token
			}
		}
	}

	return ""
}

// shouldSkipMethod checks if the HTTP method should skip CSRF validation
func (c *CSRFProtection) shouldSkipMethod(method string) bool {
	for _, m := range c.config.SkipMethods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// shouldSkipPath checks if the path should skip CSRF validation
func (c *CSRFProtection) shouldSkipPath(path string) bool {
	for _, p := range c.config.SkipPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// defaultCSRFErrorHandler is the default error handler for CSRF failures
func defaultCSRFErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"CSRF validation failed","message":"forbidden"}`))
}

// GetTokenHandler returns an HTTP handler that provides a new CSRF token
func (c *CSRFProtection) GetTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := c.GenerateToken()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"failed to generate token"}`))
			return
		}

		// Set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     c.config.CookieName,
			Value:    token.Token,
			Path:     c.config.CookiePath,
			Domain:   c.config.CookieDomain,
			Expires:  token.ExpiresAt,
			Secure:   c.config.SecureCookie,
			HttpOnly: c.config.HTTPOnly,
			SameSite: c.config.SameSite,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"` + token.Token + `"}`))
	}
}

// DoubleSubmitCookie implements the double submit cookie pattern
// This is an alternative CSRF protection method that doesn't require server-side state
type DoubleSubmitCookie struct {
	config *CSRFConfig
}

// NewDoubleSubmitCookie creates a new double submit cookie CSRF protection
func NewDoubleSubmitCookie(config *CSRFConfig) *DoubleSubmitCookie {
	return &DoubleSubmitCookie{config: config}
}

// Middleware returns middleware that validates CSRF using double submit cookie pattern
func (d *DoubleSubmitCookie) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip safe methods
			if d.isSafeMethod(r.Method) {
				// Ensure cookie exists
				if _, err := r.Cookie(d.config.CookieName); err == http.ErrNoCookie {
					d.setCookie(w)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Skip configured paths
			for _, path := range d.config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Get cookie token
			cookie, err := r.Cookie(d.config.CookieName)
			if err != nil {
				d.config.ErrorHandler(w, r, ErrInvalidCSRFToken)
				return
			}

			// Get header/form token
			headerToken := r.Header.Get(d.config.HeaderName)
			if headerToken == "" {
				if r.Method == http.MethodPost {
					r.ParseForm()
					headerToken = r.FormValue(d.config.FormFieldName)
				}
			}

			// Compare tokens (constant-time)
			if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
				d.config.ErrorHandler(w, r, ErrInvalidCSRFToken)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (d *DoubleSubmitCookie) isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

func (d *DoubleSubmitCookie) setCookie(w http.ResponseWriter) {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	token := base64.URLEncoding.EncodeToString(randomBytes)

	http.SetCookie(w, &http.Cookie{
		Name:     d.config.CookieName,
		Value:    token,
		Path:     d.config.CookiePath,
		Domain:   d.config.CookieDomain,
		Secure:   d.config.SecureCookie,
		HttpOnly: false, // Must be readable by JS
		SameSite: d.config.SameSite,
		MaxAge:   int(d.config.TokenExpiry.Seconds()),
	})
}
