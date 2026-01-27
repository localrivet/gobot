package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
)

// AdminAuthMiddleware checks if the authenticated user is an admin
// This runs AFTER go-zero's JWT middleware has validated the token
type AdminAuthMiddleware struct {
	adminUsername string
}

func NewAdminAuthMiddleware(adminUsername string) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		adminUsername: adminUsername,
	}
}

func (m *AdminAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if admin username is configured
		if m.adminUsername == "" {
			logx.Error("[AdminAuth] Admin username not configured")
			adminForbidden(w, "Admin access not configured")
			return
		}

		// Get email from JWT claims (set by go-zero JWT middleware)
		email := r.Context().Value("email")
		if email == nil {
			logx.Error("[AdminAuth] No email claim in JWT")
			adminForbidden(w, "Admin access required")
			return
		}

		emailStr, ok := email.(string)
		if !ok {
			logx.Error("[AdminAuth] Invalid email claim type in JWT")
			adminForbidden(w, "Admin access required")
			return
		}

		// Check if the authenticated user is the admin
		if emailStr != m.adminUsername {
			logx.Infof("[AdminAuth] Non-admin user attempted admin access: %s", emailStr)
			adminForbidden(w, "Admin access required")
			return
		}

		logx.Infof("[AdminAuth] Admin access granted: %s", emailStr)
		next(w, r)
	}
}

// adminForbidden sends a 403 response for non-admin users
func adminForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
