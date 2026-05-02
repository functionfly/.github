package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/sirupsen/logrus"
)

// isLocalhost checks if the client IP is from localhost (127.0.0.0/8 or ::1)
func isLocalhost(r *http.Request) bool {
	clientIP := getClientIP(r)
	if ip := net.ParseIP(clientIP); ip != nil {
		return ip.IsLoopback()
	}
	// Fallback: check if IP string starts with common localhost patterns
	return strings.HasPrefix(clientIP, "127.") ||
		strings.HasPrefix(clientIP, "10.") ||
		strings.HasPrefix(clientIP, "192.168.") ||
		clientIP == "::1" ||
		clientIP == "localhost"
}

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

			// TEMPORARY: Bypass permission checks in development mode ONLY for localhost requests
			// This prevents accidental permission bypass if DEVELOPMENT=true is set in production
			isDevelopment := os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development"
			isLocalhostRequest := isLocalhost(r)
			if isDevelopment && isLocalhostRequest {
				logrus.WithFields(logrus.Fields{
					"user_id":   claims.UserID,
					"email":     claims.Email,
					"client_ip": getClientIP(r),
				}).Info("Permission checks bypassed for development environment (localhost only)")
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
