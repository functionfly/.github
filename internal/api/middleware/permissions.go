package middleware

import (
	"net/http"
	"os"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/sirupsen/logrus"
)

// requirePermission middleware checks if user has required permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r)
			if claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Allow super_admin and admin users to bypass permission checks
			if claims.Role == "super_admin" || claims.Role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// TEMPORARY: Bypass permission checks in development mode
			// This matches the development bypass in authStore.ts and MFA middleware
			if os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development" {
				logrus.WithFields(logrus.Fields{
					"user_id": claims.UserID,
					"email":   claims.Email,
				}).Info("Permission checks bypassed for development environment")
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has the required permission
			if !m.hasPermission(claims, permission) {
				logrus.WithFields(logrus.Fields{
					"user_id":    claims.UserID,
					"email":      claims.Email,
					"permission": permission,
				}).Warn("Permission denied")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasPermission checks if the user claims contain the required permission
func (m *AuthMiddleware) hasPermission(claims *auth.Claims, permission string) bool {
	if claims.Permissions == nil {
		return false
	}

	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
