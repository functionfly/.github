package trustapi

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/sirupsen/logrus"
)

// Context keys for storing authenticated partner and API key
type contextKey string

const (
	partnerContextKey contextKey = "partner"
	apiKeyContextKey  contextKey = "api_key"
	usageContextKey   contextKey = "usage"
)

// APIKeyAuthMiddleware provides API key authentication for partner requests
type APIKeyAuthMiddleware struct {
	apikeyRepo *apikey.Repository // unified platform API key repository
	trustRepo  *trustapi.Repository // Trust API-specific repository (partners, rate limits, usage)
	logger     *logrus.Logger
}

// NewAPIKeyAuthMiddleware creates a new API key auth middleware
// apikeyRepo is the unified platform API key repository, trustRepo is for Trust-specific data
func NewAPIKeyAuthMiddleware(apikeyRepo *apikey.Repository, trustRepo *trustapi.Repository) *APIKeyAuthMiddleware {
	return &APIKeyAuthMiddleware{
		apikeyRepo: apikeyRepo,
		trustRepo:  trustRepo,
		logger:     logrus.New(),
	}
}

// Authenticate returns a middleware function that authenticates requests using API keys
func (m *APIKeyAuthMiddleware) Authenticate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				m.writeAuthError(w, http.StatusUnauthorized, "Missing Authorization header", "missing_auth")
				return
			}

			// Support both "Bearer <key>" and "X-API-Key: <key>" formats
			var rawKey string
			if strings.HasPrefix(authHeader, "Bearer ") {
				rawKey = strings.TrimPrefix(authHeader, "Bearer ")
			} else if strings.HasPrefix(authHeader, apikey.PrefixTrust) {
				rawKey = authHeader
			} else {
				m.writeAuthError(w, http.StatusUnauthorized, "Invalid Authorization header format", "invalid_auth")
				return
			}

			if rawKey == "" {
				m.writeAuthError(w, http.StatusUnauthorized, "Missing API key", "missing_api_key")
				return
			}

			// Validate the API key using the unified apikey repository
			apiKey, err := m.apikeyRepo.ValidateAPIKey(rawKey)
			if err != nil {
				m.logger.WithError(err).Warn("API key validation failed")
				m.writeAuthError(w, http.StatusUnauthorized, "Invalid API key", "invalid_api_key")
				return
			}

			// Get partner from Trust repository
			partner, err := m.trustRepo.GetPartnerByID(apiKey.PartnerID)
			if err != nil {
				m.logger.WithError(err).Warn("Partner not found for API key")
				m.writeAuthError(w, http.StatusForbidden, "Partner not found", "partner_not_found")
				return
			}

			// Check if partner is active
			if partner.Status != string(trustapi.PartnerStatusActive) {
				m.writeAuthError(w, http.StatusForbidden, "Partner account is not active", "partner_inactive")
				return
			}

			// Check IP allowlist if configured
			clientIP := getClientIP(r)
			if !m.apikeyRepo.CheckIPAllowed(apiKey, clientIP) {
				m.logger.WithFields(logrus.Fields{
					"partner_id": partner.ID,
					"key_id":     apiKey.KeyID,
					"client_ip":  clientIP,
				}).Warn("IP address not in allowlist")
				m.writeAuthError(w, http.StatusForbidden, "IP address not allowed", "ip_not_allowed")
				return
			}

			// Store partner and API key in context
			ctx := context.WithValue(r.Context(), partnerContextKey, partner)
			ctx = context.WithValue(ctx, apiKeyContextKey, apiKey)

			// Increment key usage
			if err := m.apikeyRepo.IncrementKeyUsage(apiKey.ID); err != nil {
				m.logger.WithError(err).Warn("Failed to increment key usage")
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireScope returns a middleware that checks if the API key has a required scope
func (m *APIKeyAuthMiddleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := getAPIKeyFromContext(r)
			if apiKey == nil {
				m.writeAuthError(w, http.StatusUnauthorized, "Not authenticated", "unauthenticated")
				return
			}

			if !apiKey.HasScope(scope) && !apiKey.HasScope("*") {
				m.writeAuthError(w, http.StatusForbidden, fmt.Sprintf("Missing required scope: %s", scope), "insufficient_scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware provides per-partner rate limiting
type RateLimitMiddleware struct {
	repo  *trustapi.Repository
	logger *logrus.Logger
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(repo *trustapi.Repository) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		repo:  repo,
		logger: logrus.New(),
	}
}

// RateLimit returns a middleware that enforces per-partner rate limits
func (m *RateLimitMiddleware) RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			partner := getPartnerFromContext(r)
			if partner == nil {
				// No partner in context means auth middleware wasn't applied
				// or this is a public endpoint
				next.ServeHTTP(w, r)
				return
			}

			// Get rate limit config for partner tier
			tierConfig := trustapi.GetRateLimitConfig(partner.Tier)

			// Check per-minute limit
			allowed, remaining, err := m.repo.CheckRateLimit(partner.ID, "minute", tierConfig.PerMinute)
			if err != nil {
				m.logger.WithError(err).Error("Failed to check rate limit")
				// Fail open - allow request but log the error
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				m.logger.WithFields(logrus.Fields{
					"partner_id": partner.ID,
					"tier":       partner.Tier,
					"limit":      tierConfig.PerMinute,
				}).Warn("Rate limit exceeded (per-minute)")
				writeRateLimitError(w, tierConfig.PerMinute, remaining)
				return
			}

		// Increment rate limit counter synchronously so failures are visible
		if err := m.repo.IncrementRateLimit(partner.ID, "minute"); err != nil {
			m.logger.WithError(err).Warn("Failed to increment rate limit counter")
		}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", tierConfig.PerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining-1))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))

			// Check monthly quota
			if partner.CurrentMonthUsage >= partner.MonthlyRequestLimit {
				m.logger.WithFields(logrus.Fields{
					"partner_id": partner.ID,
					"usage":      partner.CurrentMonthUsage,
					"limit":      partner.MonthlyRequestLimit,
				}).Warn("Monthly quota exceeded")
				writeError(w, http.StatusTooManyRequests, "Monthly quota exceeded", "quota_exceeded")
				return
			}

		// Increment usage counter synchronously so failures are visible
		if err := m.repo.IncrementUsage(partner.ID, 1); err != nil {
			m.logger.WithError(err).Warn("Failed to increment usage counter")
		}

			next.ServeHTTP(w, r)
		})
	}
}

// UsageTrackingMiddleware tracks API usage for billing
type UsageTrackingMiddleware struct {
	repo  *trustapi.Repository
	logger *logrus.Logger
}

// NewUsageTrackingMiddleware creates a new usage tracking middleware
func NewUsageTrackingMiddleware(repo *trustapi.Repository) *UsageTrackingMiddleware {
	return &UsageTrackingMiddleware{
		repo:  repo,
		logger: logrus.New(),
	}
}

// Track returns a middleware that tracks API usage
func (m *UsageTrackingMiddleware) Track() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Only track if partner is authenticated
			partner := getPartnerFromContext(r)
			if partner == nil {
				return
			}

			apiKey := getAPIKeyFromContext(r)
			responseTime := time.Since(startTime).Milliseconds()

			// Create usage record
			usage := &trustapi.TrustAPIUsage{
				PartnerID:      partner.ID,
				Endpoint:       r.URL.Path,
				Method:         r.Method,
				RequestID:      generateRequestID(),
				StatusCode:     wrapped.statusCode,
				ResponseTimeMs: int(responseTime),
				IPAddress:      getClientIP(r),
				UserAgent:      r.UserAgent(),
			}

			if apiKey != nil {
				usage.APIKeyID = &apiKey.ID
			}

		// Record usage synchronously so failures are visible
		if err := m.repo.RecordUsage(usage); err != nil {
			m.logger.WithError(err).Warn("Failed to record usage")
		}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper functions

// getPartnerFromContext retrieves the authenticated partner from the request context
func getPartnerFromContext(r *http.Request) *trustapi.TrustAPIPartner {
	if partner, ok := r.Context().Value(partnerContextKey).(*trustapi.TrustAPIPartner); ok {
		return partner
	}
	return nil
}

// getAPIKeyFromContext retrieves the authenticated API key from the request context
func getAPIKeyFromContext(r *http.Request) *apikey.APIKey {
	if apiKey, ok := r.Context().Value(apiKeyContextKey).(*apikey.APIKey); ok {
		return apiKey
	}
	return nil
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	id, err := randomString(8)
	if err != nil {
		// Fallback to non-cryptographic random if crypto/rand fails
		return fmt.Sprintf("req_%d_%x", time.Now().UnixNano(), time.Now().UnixNano()%0xFFFFFFFF)
	}
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), id)
}

// randomString generates a cryptographically random string of given length
func randomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	randBytes := make([]byte, n)
	if _, err := crand.Read(randBytes); err != nil {
		return "", fmt.Errorf("crypto/rand unavailable: %w", err)
	}
	for i := range b {
		b[i] = letters[int(randBytes[i])%len(letters)]
	}
	return string(b), nil
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, err string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":"%s","code":"%s"}`, err, code)
}

// writeRateLimitError writes a rate limit exceeded error response
func writeRateLimitError(w http.ResponseWriter, limit int, remaining int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, `{"error":"Rate limit exceeded","code":"rate_limit_exceeded","retry_after":60}`)
}

// writeAuthError writes a JSON error response for auth middleware
func (m *APIKeyAuthMiddleware) writeAuthError(w http.ResponseWriter, status int, err string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":"%s","code":"%s"}`, err, code)
}
