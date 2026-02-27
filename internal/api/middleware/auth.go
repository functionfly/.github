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

// extractUserFromToken extracts user information from JWT token in Authorization header
func (m *AuthMiddleware) extractUserFromToken(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	return m.authSvc.ValidateToken(parts[1])
}

// requireAuth middleware validates JWT token and adds user context
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.extractUserFromToken(r)
		if err != nil {
			logrus.WithError(err).Warn("Authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user claims to request context
		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// getUserFromContext extracts user claims from request context
func GetUserFromContext(r *http.Request) *auth.Claims {
	if claims, ok := r.Context().Value("user").(*auth.Claims); ok {
		return claims
	}
	return nil
}

// getActingTenantID gets the tenant ID the admin is currently acting as (for tenant-scoped operations)
func GetActingTenantID(r *http.Request) *uuid.UUID {
	if tenantID, ok := r.Context().Value("acting_tenant_id").(*uuid.UUID); ok {
		return tenantID
	}
	return nil
}

// setActingTenantID sets the tenant context for tenant-scoped operations
func SetActingTenantID(r *http.Request, tenantID *uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), "acting_tenant_id", tenantID)
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
			tenant, err := repo.GetTenantByID(*actingTenantID)
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