package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ctx is the default context for Redis operations
var ctx = context.Background()

const (
	// CSRFTokenHeader is the header name for CSRF token
	CSRFTokenHeader = "X-CSRF-Token"
	// CSRFTokenCookie is the cookie name for CSRF token (double-submit pattern)
	CSRFTokenCookie = "csrf_token"
	// CSRFTokenTTL is the TTL for CSRF tokens in Redis
	CSRFTokenTTL = 1 * time.Hour
	// CSRFTokenLength is the number of bytes for CSRF token generation
	CSRFTokenLength = 32
)

const (
	// CSRFKeyPrefix is the Redis key prefix for CSRF tokens
	CSRFKeyPrefix = "csrf:admin:"
)

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Token     string    `json:"token"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CSRFMiddleware handles CSRF token generation and validation for admin routes
type CSRFMiddleware struct {
	redisClient *cache.UpstashRedisClient
	authSvc     *auth.AuthService
	logger      *logrus.Entry
}

// NewCSRFMiddleware creates a new CSRF middleware instance
func NewCSRFMiddleware(redisClient *cache.UpstashRedisClient, authSvc *auth.AuthService) *CSRFMiddleware {
	return &CSRFMiddleware{
		redisClient: redisClient,
		authSvc:     authSvc,
		logger:      logrus.WithField("middleware", "csrf"),
	}
}

// GenerateToken generates a new CSRF token for the given session
func (m *CSRFMiddleware) GenerateToken(sessionID string) (*CSRFToken, error) {
	// Generate cryptographically secure random bytes
	tokenBytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		m.logger.WithError(err).Error("Failed to generate CSRF token random bytes")
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	// Encode as base64url (URL-safe base64 without padding)
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	csrfToken := &CSRFToken{
		Token:     token,
		SessionID: sessionID,
		CreatedAt: now,
		ExpiresAt: now.Add(CSRFTokenTTL),
	}

	// Store token in Redis with TTL
	if m.redisClient != nil {
		key := m.getRedisKey(sessionID)
		tokenData, err := json.Marshal(csrfToken)
		if err != nil {
			m.logger.WithError(err).Error("Failed to marshal CSRF token for Redis")
			return nil, fmt.Errorf("failed to store CSRF token: %w", err)
		}

		if err := m.redisClient.Set(ctx, key, tokenData, CSRFTokenTTL); err != nil {
			m.logger.WithError(err).Error("Failed to store CSRF token in Redis")
			return nil, fmt.Errorf("failed to store CSRF token: %w", err)
		}

		m.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"expires_at": csrfToken.ExpiresAt,
		}).Debug("Generated CSRF token and stored in Redis")
	}

	return csrfToken, nil
}

// ValidateToken validates a CSRF token for the given session
func (m *CSRFMiddleware) ValidateToken(sessionID, token string) error {
	if token == "" {
		return fmt.Errorf("CSRF token is required")
	}

	// SECURITY FIX: Fail-closed on Redis unavailability
	// Previously, CSRF validation was skipped when Redis was unavailable, creating a security gap
	// This rejects the request instead of allowing it through
	if m.redisClient == nil {
		m.logger.Error("Redis not available, rejecting CSRF validation for security (fail-closed)")
		return fmt.Errorf("service temporarily unavailable: Redis is required for CSRF validation")
	}

	key := m.getRedisKey(sessionID)
	tokenData, err := m.redisClient.GetBytes(ctx, key)
	if err != nil {
		if err.Error() == "key not found" || err.Error() == "redis: nil" {
			m.logger.WithFields(logrus.Fields{
				"session_id": sessionID,
			}).Warn("CSRF token not found or expired")
			return fmt.Errorf("CSRF token is missing, expired, or invalid")
		}
		m.logger.WithError(err).Error("Failed to get CSRF token from Redis")
		return fmt.Errorf("failed to validate CSRF token")
	}

	var storedToken CSRFToken
	if err := json.Unmarshal(tokenData, &storedToken); err != nil {
		m.logger.WithError(err).Error("Failed to unmarshal stored CSRF token")
		return fmt.Errorf("failed to validate CSRF token")
	}

	// Compare tokens using constant-time comparison to prevent timing attacks
	if !secureCompare(storedToken.Token, token) {
		m.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
		}).Warn("CSRF token mismatch")
		return fmt.Errorf("CSRF token is missing, expired, or invalid")
	}

	// Check if token has expired
	if time.Now().UTC().After(storedToken.ExpiresAt) {
		m.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
		}).Warn("CSRF token has expired")
		return fmt.Errorf("CSRF token is missing, expired, or invalid")
	}

	m.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
	}).Debug("CSRF token validated successfully")

	return nil
}

// InvalidateToken removes a CSRF token from Redis (e.g., on logout)
func (m *CSRFMiddleware) InvalidateToken(sessionID string) error {
	if m.redisClient == nil {
		return nil
	}

	key := m.getRedisKey(sessionID)
	if _, err := m.redisClient.Del(ctx, key); err != nil {
		m.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to invalidate CSRF token")
		return fmt.Errorf("failed to invalidate CSRF token: %w", err)
	}

	m.logger.WithField("session_id", sessionID).Debug("CSRF token invalidated")
	return nil
}

// HandleGetCSRFToken handles GET /v1/admin/csrf - returns a new CSRF token for the current session
func (m *CSRFMiddleware) HandleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := claims.UserID.String()
	token, err := m.GenerateToken(sessionID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to generate CSRF token")
		http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token.Token,
		"expires_at": token.ExpiresAt.Format(time.RFC3339),
	})
}

// csrfExemptPaths are paths where CSRF is not required.
// Auth endpoints are JWT-protected and used to bootstrap/renew sessions before a CSRF token exists.
// Security events is fire-and-forget telemetry — not a mutating action on data.
// Signup invite revoke is HMAC-protected and intended for programmatic/API access.
var csrfExemptPaths = map[string]bool{
	"/v1/admin/auth/session":    true,
	"/v1/admin/auth/last-login":   true,
	"/v1/admin/security/events":   true, // client telemetry, fire-and-forget
}

// isCSRFExempt checks if a path is exempt from CSRF validation
func isCSRFExempt(path string) bool {
	// Check exact matches
	if csrfExemptPaths[path] {
		return true
	}

	// Check pattern matches
	if strings.HasPrefix(path, "/v1/admin/signup-invites/") && strings.HasSuffix(path, "/revoke") {
		return true
	}

	return false
}

// RequireCSRF is middleware that validates CSRF tokens on mutating requests
func (m *CSRFMiddleware) RequireCSRF(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only validate on mutating methods
		method := r.Method
		if method != http.MethodPost && method != http.MethodPut &&
			method != http.MethodPatch && method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		// Exempt auth session endpoints — they're JWT-protected and used to
		// bootstrap/renew sessions before a CSRF token exists.
		if isCSRFExempt(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get session from context
		claims := GetUserFromContext(r)
		if claims == nil {
			// Unauthenticated — CSRF doesn't apply (no session to hijack)
			next.ServeHTTP(w, r)
			return
		}

		sessionID := claims.UserID.String()

		// Extract CSRF token from header
		token := r.Header.Get(CSRFTokenHeader)

		// Also check for double-submit cookie as fallback
		if token == "" {
			if cookie, err := r.Cookie(CSRFTokenCookie); err == nil {
				token = cookie.Value
			}
		}

		if token == "" {
			m.logger.WithFields(logrus.Fields{
				"method":     method,
				"path":       r.URL.Path,
				"session_id": sessionID,
			}).Warn("CSRF token missing from request")
			writeCSRFError(w, "CSRF token is missing, expired, or invalid", "X-CSRF-Token")
			return
		}

		if err := m.ValidateToken(sessionID, token); err != nil {
			m.logger.WithFields(logrus.Fields{
				"method":     method,
				"path":       r.URL.Path,
				"session_id": sessionID,
				"error":      err.Error(),
			}).Warn("CSRF token validation failed")
			// Check if it's a service unavailable error (Redis down)
			if strings.Contains(err.Error(), "service temporarily unavailable") {
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
			writeCSRFError(w, "CSRF token is missing, expired, or invalid", "X-CSRF-Token")
			return
		}

		// Update last activity on the token (refresh TTL)
		if err := m.refreshTokenTTL(sessionID); err != nil {
			m.logger.WithError(err).Warn("Failed to refresh CSRF token TTL")
		}

		next.ServeHTTP(w, r)
	}
}

// refreshTokenTTL refreshes the TTL of a CSRF token
func (m *CSRFMiddleware) refreshTokenTTL(sessionID string) error {
	if m.redisClient == nil {
		return nil
	}

	key := m.getRedisKey(sessionID)
	_, err := m.redisClient.Expire(ctx, key, CSRFTokenTTL)
	return err
}

// getRedisKey returns the Redis key for a CSRF token
func (m *CSRFMiddleware) getRedisKey(sessionID string) string {
	return CSRFKeyPrefix + sessionID
}

// writeCSRFError writes a CSRF error response
func writeCSRFError(w http.ResponseWriter, message, requiredHeader string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":           "csrf_token_invalid",
		"message":         message,
		"required_header": requiredHeader,
	})
}

// secureCompare performs a constant-time comparison of two strings
// to prevent timing attacks
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// extractClientIP extracts the client IP address from the request,
// considering X-Forwarded-For and X-Real-IP headers
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// CSRFHandler returns an HTTP handler for CSRF token operations
func (m *CSRFMiddleware) CSRFHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodGet:
			// GET /v1/admin/csrf - get a new CSRF token
			m.HandleGetCSRFToken(w, r)
		case http.MethodDelete:
			// DELETE /v1/admin/csrf - invalidate current CSRF token
			claims := GetUserFromContext(r)
			if claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			sessionID := vars["sessionId"]
			if sessionID == "" {
				sessionID = claims.UserID.String()
			}
			if err := m.InvalidateToken(sessionID); err != nil {
				m.logger.WithError(err).Error("Failed to invalidate CSRF token")
				http.Error(w, "Failed to invalidate CSRF token", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
