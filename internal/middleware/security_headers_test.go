package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	config := DefaultSecurityHeaders()

	handler := SecurityHeadersMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check all security headers are set
	tests := []struct {
		header   string
		expected string
	}{
		{"Content-Security-Policy", config.ContentSecurityPolicy},
		{"Permissions-Policy", config.PermissionsPolicy},
		{"Referrer-Policy", config.ReferrerPolicy},
		{"Strict-Transport-Security", config.StrictTransportSecurity},
		{"X-Content-Type-Options", config.XContentTypeOptions},
		{"X-Frame-Options", config.XFrameOptions},
		{"X-XSS-Protection", config.XXSSProtection},
		{"Cache-Control", config.CacheControl},
		{"Pragma", config.Pragma},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := rec.Header().Get(tt.header)
			if got != tt.expected {
				t.Errorf("Header %s = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}

func TestAPISecurityHeaders(t *testing.T) {
	config := APISecurityHeaders()

	// Verify API-specific settings
	if config.ContentSecurityPolicy != "default-src 'none'; frame-ancestors 'none'" {
		t.Errorf("API CSP should be restrictive, got %s", config.ContentSecurityPolicy)
	}

	if config.XFrameOptions != "DENY" {
		t.Errorf("X-Frame-Options should be DENY, got %s", config.XFrameOptions)
	}
}

func TestSecurityHeadersMiddlewareWithNilConfig(t *testing.T) {
	// Should use default config when nil is passed
	handler := SecurityHeadersMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should have default headers set
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected default X-Content-Type-Options header")
	}
}

func TestSecureHandler(t *testing.T) {
	config := DefaultSecurityHeaders()

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := SecureHandler(baseHandler, config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Expected X-Frame-Options header")
	}
}

func TestSecureHandlerFunc(t *testing.T) {
	config := DefaultSecurityHeaders()

	baseHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}

	handler := SecureHandlerFunc(baseHandler, config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected X-Content-Type-Options header")
	}
}

func TestSecurityHeadersEmptyValues(t *testing.T) {
	// Create config with empty values
	config := &SecurityHeaders{
		ContentSecurityPolicy: "",
		XContentTypeOptions:   "",
		XFrameOptions:         "",
	}

	handler := SecurityHeadersMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Empty config values should not set headers
	if rec.Header().Get("Content-Security-Policy") != "" {
		t.Error("Empty CSP should not be set")
	}
}
