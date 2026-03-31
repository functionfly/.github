package gba

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	apimiddleware "github.com/functionfly/functionfly/internal/api/middleware"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyTenantID is the context key for tenant ID
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeySession is the context key for session
	ContextKeySession ContextKey = "session"
)

// Middleware provides authentication middleware
type Middleware struct {
	auth   *Auth
	logger *logrus.Logger
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(auth *Auth) *Middleware {
	return &Middleware{
		auth:   auth,
		logger: auth.Logger(),
	}
}

// RequirePermission middleware checks if user has required permission based on their role
func (m *Middleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.auth.IsEnabled("session") {
				// Fallback to legacy auth - continue
				m.nextWithLegacyAuth(w, r, permission, next)
				return
			}

			token := m.auth.sessions.GetSessionTokenFromRequest(r)
			if token == "" {
				http.Error(w, `{"message": "Authentication required"}`, http.StatusUnauthorized)
				return
			}

			session, err := m.auth.sessions.ValidateSession(m.auth.GetDB(), token)
			if err != nil {
				m.logger.WithError(err).Debug("Invalid session")
				http.Error(w, `{"message": "Invalid or expired session"}`, http.StatusUnauthorized)
				return
			}

			// Get user to check role and permissions
			var user User
			if err := m.auth.GetDB().First(&user, session.UserID).Error; err != nil {
				http.Error(w, `{"message": "User not found"}`, http.StatusInternalServerError)
				return
			}

			// Check if user has the required permission based on their role
			if !m.hasPermission(user.Role, permission) {
				m.logger.WithFields(logrus.Fields{
					"user_id":    session.UserID,
					"email":      user.Email,
					"role":       user.Role,
					"permission": permission,
				}).Warn("Permission denied")
				http.Error(w, `{"message": "Forbidden"}`, http.StatusForbidden)
				return
			}

			// Add session info to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyUserID, session.UserID)
			ctx = context.WithValue(ctx, ContextKeyTenantID, session.TenantID)
			ctx = context.WithValue(ctx, ContextKeySession, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware adds auth info to context if available
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.auth.IsEnabled("session") {
			next.ServeHTTP(w, r)
			return
		}

		token := m.auth.sessions.GetSessionTokenFromRequest(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		session, err := m.auth.sessions.ValidateSession(m.auth.GetDB(), token)
		if err != nil {
			// Invalid session, but optional - continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Add session info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyUserID, session.UserID)
		ctx = context.WithValue(ctx, ContextKeyTenantID, session.TenantID)
		ctx = context.WithValue(ctx, ContextKeySession, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// hasPermission checks if a role has a specific permission
func (m *Middleware) hasPermission(role, permission string) bool {
	permissions := m.auth.getPermissionsForRole(role)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// nextWithLegacyAuth handles fallback to legacy auth system (JWT claims from upstream middleware)
func (m *Middleware) nextWithLegacyAuth(w http.ResponseWriter, r *http.Request, permission string, next http.Handler) {
	w.Header().Set("X-Auth-Mode", "legacy")
	claims := apimiddleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"message": "Authentication required"}`, http.StatusUnauthorized)
		return
	}
	if claims.Role == "super_admin" || claims.Role == "admin" {
		next.ServeHTTP(w, r)
		return
	}
	if (os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development") && os.Getenv("PRODUCTION_ENV") != "true" {
		m.logger.WithFields(logrus.Fields{"user_id": claims.UserID, "email": claims.Email}).
			Debug("Legacy permission check bypassed for development")
		next.ServeHTTP(w, r)
		return
	}
	for _, p := range claims.Permissions {
		if p == permission {
			next.ServeHTTP(w, r)
			return
		}
	}
	m.logger.WithFields(logrus.Fields{
		"user_id": claims.UserID, "email": claims.Email, "permission": permission,
	}).Warn("Legacy permission denied")
	http.Error(w, `{"message": "Forbidden"}`, http.StatusForbidden)
}

// ExtractTenant middleware extracts and validates tenant context
func (m *Middleware) ExtractTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := m.extractTenantID(r)
		if tenantID != uuid.Nil {
			ctx := context.WithValue(r.Context(), ContextKeyTenantID, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserID extracts the user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return userID, ok
}

// GetTenantID extracts the tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(ContextKeyTenantID).(uuid.UUID)
	return tenantID, ok
}

// GetSession extracts the session from context
func GetSession(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(ContextKeySession).(*Session)
	return session, ok
}

// extractTenantID extracts tenant ID from request
func (m *Middleware) extractTenantID(r *http.Request) uuid.UUID {
	// Check header first
	tenantHeader := r.Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		if id, err := uuid.Parse(tenantHeader); err == nil {
			return id
		}
	}

	// Check subdomain
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		subdomain := parts[0]
		// Skip reserved subdomains
		reserved := []string{"www", "api", "auth", "admin", "app", "staging", "dev"}
		for _, r := range reserved {
			if strings.EqualFold(subdomain, r) {
				return uuid.Nil
			}
		}
		// Look up active tenant by subdomain
		var tenant Tenant
		if err := m.auth.GetDB().Where("subdomain = ? AND status = ?", subdomain, "active").First(&tenant).Error; err == nil {
			return tenant.ID
		}
	}

	return uuid.Nil
}
