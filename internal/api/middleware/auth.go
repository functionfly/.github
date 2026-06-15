package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// contextKey is an unexported type used for context keys in this package.
// Using a typed key prevents collisions with keys from other packages.
type contextKey int

const (
	contextKeyUser contextKey = iota
	contextKeyActingTenantID
)

// AuthMiddleware contains authentication middleware functions
type AuthMiddleware struct {
	authSvc *auth.AuthService
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(authSvc *auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authSvc: authSvc,
	}
}

// extractUserFromToken extracts user information from JWT token in Authorization header or httpOnly cookie
func (m *AuthMiddleware) extractUserFromToken(r *http.Request) (*auth.Claims, error) {
	// Try Authorization header first (API clients)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return m.authSvc.ValidateToken(r.Context(), parts[1])
		}
	}

	// Fall back to httpOnly cookie (browser clients)
	if cookie, err := r.Cookie(auth.CookieNameAccessToken); err == nil && cookie.Value != "" {
		return m.authSvc.ValidateToken(r.Context(), cookie.Value)
	}

	return nil, fmt.Errorf("missing authorization header and cookie")
}

// OptionalAuth parses a valid Bearer JWT when present and sets user context.
// Missing or invalid tokens are ignored (no 401) so public routes keep working.
// Use on /v1 (and /v2) so dashboard requests with Authorization update activity / online status.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			next.ServeHTTP(w, r)
			return
		}
		claims, err := m.extractUserFromToken(r)
		if err != nil || claims == nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUser, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth middleware validates JWT token and adds user context.
// It skips authentication for OPTIONS preflight requests so CORS works.
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for CORS preflight requests
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.extractUserFromToken(r)
		if err != nil {
			logrus.WithError(err).Warn("Authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user claims to request context using typed key to prevent collisions
		ctx := context.WithValue(r.Context(), contextKeyUser, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequirePermission returns a middleware that checks if the authenticated user has the required permission.
// Must be used after RequireAuth since it reads claims from context.
func (m *AuthMiddleware) RequirePermission(requiredPermission string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r)
			if claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !claims.HasPermission(requiredPermission) {
				logrus.WithFields(logrus.Fields{
					"user_id":              claims.UserID,
					"required_permission":  requiredPermission,
					"user_permissions":     claims.Permissions,
				}).Warn("Permission denied")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(r *http.Request) *auth.Claims {
	return GetClaimsFromContext(r.Context())
}

// GetClaimsFromContext extracts auth claims from context (for use when only context is available, e.g. versioning middleware).
func GetClaimsFromContext(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(contextKeyUser).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// GetTenantID extracts the tenant ID from the request context
// Returns the TenantID from the user's claims if authenticated
func GetTenantID(r *http.Request) (uuid.UUID, bool) {
	claims := GetUserFromContext(r)
	if claims != nil {
		return claims.TenantID, true
	}
	return uuid.Nil, false
}

// getActingTenantID gets the tenant ID the admin is currently acting as (for tenant-scoped operations)
func GetActingTenantID(r *http.Request) *uuid.UUID {
	if tenantID, ok := r.Context().Value(contextKeyActingTenantID).(*uuid.UUID); ok {
		return tenantID
	}
	return nil
}

// SetActingTenantID sets the tenant context for tenant-scoped operations
func SetActingTenantID(r *http.Request, tenantID *uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), contextKeyActingTenantID, tenantID)
	return r.WithContext(ctx)
}

// SetUserInContext injects auth claims into the request context.
// Intended for use in tests that need to simulate an authenticated request
// without going through the full JWT middleware.
func SetUserInContext(r *http.Request, claims *auth.Claims) *http.Request {
	ctx := context.WithValue(r.Context(), contextKeyUser, claims)
	return r.WithContext(ctx)
}

// requireTenantContext middleware ensures the request has a valid tenant context for tenant-scoped operations
func (m *AuthMiddleware) RequireTenantContext(repo storage.Repository) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			actingTenantID := GetActingTenantID(r)
			if actingTenantID == nil {
				http.Error(w, "Tenant context required for this operation", http.StatusBadRequest)
				return
			}

			// Verify the tenant exists and is not suspended
			tenant, err := repo.GetTenantByID(ctx, *actingTenantID)
			if err != nil {
				logrus.WithError(err).Error("Failed to verify tenant context")
				http.Error(w, "Failed to verify tenant context", http.StatusInternalServerError)
				return
			}
			if tenant == nil {
				http.Error(w, "Tenant not found", http.StatusNotFound)
				return
			}
			if tenant.Status == "suspended" {
				http.Error(w, "Cannot operate on suspended tenant", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
