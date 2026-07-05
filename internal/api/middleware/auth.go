package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
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
	authSvc   *auth.AuthService
	apiKeyRepo *apikey.Repository
	repo       storage.Repository
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(authSvc *auth.AuthService, apiKeyRepo *apikey.Repository, repo storage.Repository) *AuthMiddleware {
	return &AuthMiddleware{
		authSvc:   authSvc,
		apiKeyRepo: apiKeyRepo,
		repo:       repo,
	}
}

// looksLikeJWT returns true for tokens of the form xxx.yyy.zzz.
func looksLikeJWT(tok string) bool {
	if tok == "" {
		return false
	}
	parts := strings.Split(tok, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

// authenticateWithAPIKey validates an API key and returns full Claims with permissions.
func (m *AuthMiddleware) authenticateWithAPIKey(rawKey string) (*auth.Claims, error) {
	if m.apiKeyRepo == nil {
		return nil, fmt.Errorf("API key repository not configured")
	}

	hasher := apikey.NewHasher()
	keyHash := hasher.HashDeterministic(rawKey)

	apiKey, err := m.apiKeyRepo.GetByHash(context.Background(), keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is not active")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	user, err := m.repo.GetUserByID(context.Background(), apiKey.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	claims := &auth.Claims{
		UserID:   user.ID,
		Email:    user.Email,
		TenantID: user.TenantID,
		Role:     user.Role,
	}

	claims.Permissions = m.authSvc.GetPermissionsForRole(user.Role)

	return claims, nil
}

// extractUserFromToken extracts user information from JWT token or API key in Authorization header, query param, or httpOnly cookie
func (m *AuthMiddleware) extractUserFromToken(r *http.Request) (*auth.Claims, error) {
	// Try Authorization header first (API clients)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			raw := parts[1]

			// Try JWT first if it looks like a JWT
			if looksLikeJWT(raw) {
				if claims, err := m.authSvc.ValidateToken(r.Context(), raw); err == nil && claims != nil {
					return claims, nil
				}
				// JWT validation failed - don't fall through to API key for malformed JWTs
				return nil, fmt.Errorf("invalid JWT token")
			}

			// Not a JWT - try API key authentication
			return m.authenticateWithAPIKey(raw)
		}
	}

	// Try query parameter (WebSocket connections can't set headers) - JWT only
	if token := r.URL.Query().Get("token"); token != "" {
		return m.authSvc.ValidateToken(r.Context(), token)
	}

	// Fall back to httpOnly cookie (browser clients) - JWT only
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
			apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
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
				apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
				return
			}

			if !claims.HasPermission(requiredPermission) {
				logrus.WithFields(logrus.Fields{
					"user_id":              claims.UserID,
					"required_permission":  requiredPermission,
					"user_permissions":     claims.Permissions,
				}).Warn("Permission denied")
				apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
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
				apierror.WriteError(w, apierror.NewBadRequest("Tenant context required for this operation"))
				return
			}

			// Verify the tenant exists and is not suspended
			tenant, err := repo.GetTenantByID(r.Context(), *actingTenantID)
			if err != nil {
				logrus.WithError(err).Error("Failed to verify tenant context")
				apierror.WriteError(w, apierror.NewInternal("Failed to verify tenant context"))
				return
			}
			if tenant == nil {
				apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
				return
			}
			if tenant.Status == "suspended" {
				apierror.WriteError(w, apierror.NewForbidden("Cannot operate on suspended tenant"))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
