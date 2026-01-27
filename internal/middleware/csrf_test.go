package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCSRFProtection_GenerateToken(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	csrf := NewCSRFProtection(config)

	token, err := csrf.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	if token.Token == "" {
		t.Error("Generated token should not be empty")
	}

	if token.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired")
	}
}

func TestCSRFProtection_ValidateToken(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	csrf := NewCSRFProtection(config)

	// Generate a valid token
	token, err := csrf.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	// Validate the token
	err = csrf.ValidateToken(token.Token)
	if err != nil {
		t.Errorf("ValidateToken() should succeed for valid token: %v", err)
	}

	// Test with invalid token
	err = csrf.ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken() should fail for invalid token")
	}

	// Test with empty token
	err = csrf.ValidateToken("")
	if err != ErrInvalidCSRFToken {
		t.Error("ValidateToken() should return ErrInvalidCSRFToken for empty token")
	}
}

func TestCSRFProtection_Middleware(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	csrf := NewCSRFProtection(config)

	handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// GET request should pass (safe method)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET request should pass, got status %d", rec.Code)
	}

	// POST request without token should fail
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("POST without token should fail, got status %d", rec.Code)
	}

	// POST request with valid token should pass
	token, _ := csrf.GenerateToken()
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-CSRF-Token", token.Token)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST with valid token should pass, got status %d", rec.Code)
	}
}

func TestCSRFProtection_SkipPaths(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	config.SkipPaths = []string{"/api/"}
	csrf := NewCSRFProtection(config)

	handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// POST to /api/ should skip CSRF
	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST to skipped path should pass, got status %d", rec.Code)
	}

	// POST to other path should require CSRF
	req = httptest.NewRequest(http.MethodPost, "/other", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("POST to non-skipped path should fail, got status %d", rec.Code)
	}
}

func TestCSRFProtection_GetTokenHandler(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	csrf := NewCSRFProtection(config)

	handler := csrf.GetTokenHandler()

	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetTokenHandler should return 200, got %d", rec.Code)
	}

	// Check that cookie is set
	cookies := rec.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == config.CookieName {
			found = true
			break
		}
	}
	if !found {
		t.Error("CSRF cookie should be set")
	}
}

func TestDoubleSubmitCookie(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	dsc := NewDoubleSubmitCookie(config)

	handler := dsc.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// GET request should pass and set cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET should pass, got %d", rec.Code)
	}

	// Get the cookie value
	cookies := rec.Result().Cookies()
	var cookieValue string
	for _, cookie := range cookies {
		if cookie.Name == config.CookieName {
			cookieValue = cookie.Value
			break
		}
	}

	if cookieValue == "" {
		t.Fatal("CSRF cookie should be set on GET")
	}

	// POST with matching cookie and header should pass
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: config.CookieName, Value: cookieValue})
	req.Header.Set(config.HeaderName, cookieValue)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST with matching tokens should pass, got %d", rec.Code)
	}

	// POST with mismatched tokens should fail
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: config.CookieName, Value: cookieValue})
	req.Header.Set(config.HeaderName, "different-value")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("POST with mismatched tokens should fail, got %d", rec.Code)
	}
}

func TestCSRFProtection_SafeMethods(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	csrf := NewCSRFProtection(config)

	handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	safeMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace}

	for _, method := range safeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("%s should pass without CSRF token, got %d", method, rec.Code)
			}
		})
	}
}

func TestCSRFProtection_UnsafeMethods(t *testing.T) {
	secret := []byte("test-secret-key-must-be-32-bytes-")
	config := DefaultCSRFConfig(secret)
	csrf := NewCSRFProtection(config)

	handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	unsafeMethods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range unsafeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("%s without CSRF token should fail, got %d", method, rec.Code)
			}
		})
	}
}
