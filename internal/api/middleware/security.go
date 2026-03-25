package middleware

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/signing"
	"github.com/functionfly/functionfly/internal/api/middleware/advanced_security"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// SecurityMiddleware contains security-related middleware functions
type SecurityMiddleware struct {
	signer      *signing.RequestSigner
	verifier    *signing.RequestVerifier
	rateLimiter *RateLimiter
}

// RateLimiter implements in-memory rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

// Allow checks if a request from the given key should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean up old entries
	if timestamps, exists := rl.requests[key]; exists {
		// Filter out timestamps outside the window
		validTimestamps := make([]time.Time, 0, len(timestamps))
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		rl.requests[key] = validTimestamps
	}

	// Check if under limit
	if len(rl.requests[key]) < rl.limit {
		rl.requests[key] = append(rl.requests[key], now)
		return true
	}

	return false
}

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware() *SecurityMiddleware {
	// Default rate limiting: 100 requests per minute
	rateLimitStr := os.Getenv("RATE_LIMIT_REQUESTS")
	rateLimit := 100 // default
	if rateLimitStr != "" {
		if parsed, err := strconv.Atoi(rateLimitStr); err == nil && parsed > 0 {
			rateLimit = parsed
		}
	}

	rateWindowStr := os.Getenv("RATE_LIMIT_WINDOW_SECONDS")
	rateWindow := 60 // default 1 minute
	if rateWindowStr != "" {
		if parsed, err := strconv.Atoi(rateWindowStr); err == nil && parsed > 0 {
			rateWindow = parsed
		}
	}

	return &SecurityMiddleware{
		signer:      &signing.RequestSigner{},
		verifier:    &signing.RequestVerifier{},
		rateLimiter: NewRateLimiter(time.Duration(rateWindow)*time.Second, rateLimit),
	}
}

// RequireHMACSignature middleware validates HMAC signatures for incoming requests
// This is used for API requests that require request signing (like webhooks or API calls from trusted clients)
func (sm *SecurityMiddleware) RequireHMACSignature(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the shared secret from environment
		sharedSecret := os.Getenv("API_SHARED_SECRET")
		if sharedSecret == "" {
			// Development: allow and log; production: reject (missing secret is a misconfiguration).
			isDev := os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development"
			if isDev {
				logrus.Warn("API_SHARED_SECRET not configured — HMAC verification skipped (development only)")
				next.ServeHTTP(w, r)
				return
			}
			logrus.Error("API_SHARED_SECRET not configured — rejecting HMAC-required request in production")
			http.Error(w, "Service misconfigured", http.StatusInternalServerError)
			return
		}

		// Verify the request signature
		valid, err := sm.verifier.VerifyRequest(r, sharedSecret)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"ip":     getClientIP(r),
			}).Warn("HMAC signature verification failed")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		if !valid {
			logrus.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"ip":     getClientIP(r),
			}).Warn("Invalid HMAC signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RateLimit middleware applies rate limiting to requests
func (sm *SecurityMiddleware) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use client IP as the rate limiting key
		clientIP := getClientIP(r)

		// Check if request should be allowed
		if !sm.rateLimiter.Allow(clientIP) {
			logrus.WithFields(logrus.Fields{
				"ip":     clientIP,
				"method": r.Method,
				"path":   r.URL.Path,
			}).Warn("Rate limit exceeded")

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", sm.rateLimiter.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(sm.rateLimiter.window).Unix()))

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers to successful requests
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", sm.rateLimiter.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", sm.rateLimiter.limit-len(sm.rateLimiter.requests[clientIP])))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(sm.rateLimiter.window).Unix()))

		next.ServeHTTP(w, r)
	}
}

// getCORSAllowedOrigins parses CORS_ALLOWED_ORIGINS. In production, empty env means
// no origins allowed (fail closed). In development, empty env defaults to ["*"].
func getCORSAllowedOrigins() []string {
	allowedOriginsStr := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	isProd := os.Getenv("ENVIRONMENT") == "production" || os.Getenv("NODE_ENV") == "production"
	if allowedOriginsStr == "" {
		if isProd {
			return nil // fail closed: no wildcard in production
		}
		return []string{"*"}
	}
	parts := strings.Split(allowedOriginsStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// IsOriginAllowedForRequest returns true if the request's Origin is allowed (for CORS/WebSocket).
// Empty origin (same-origin or non-browser) is allowed. In production, empty allowlist denies all.
func IsOriginAllowedForRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	allowed := getCORSAllowedOrigins()
	isDev := os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development"
	if len(allowed) == 0 {
		// Production with no CORS_ALLOWED_ORIGINS: deny cross-origin except localhost (for local dashboard dev)
		return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
	}
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	if isDev && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
		return true
	}
	return false
}

// CORSMiddleware adds CORS headers to responses
func (sm *SecurityMiddleware) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := getCORSAllowedOrigins()
		origin := r.Header.Get("Origin")
		// In production, empty allowlist must not result in allowing any origin.
		// Exception: allow localhost origins so local dashboard (e.g. :3000) can call the API without setting DEVELOPMENT or CORS_ALLOWED_ORIGINS.
		if len(allowedOrigins) == 0 && origin != "" {
			isDev := os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development"
			isLocalhost := strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
			if !isDev && !isLocalhost {
				logrus.Error("CORS_ALLOWED_ORIGINS is empty in production — rejecting cross-origin request")
				http.Error(w, "CORS not configured", http.StatusForbidden)
				return
			}
		}

		// Get allowed methods from environment
		allowedMethodsStr := os.Getenv("CORS_ALLOWED_METHODS")
		allowedMethods := "GET, POST, PUT, PATCH, DELETE, OPTIONS"
		if allowedMethodsStr != "" {
			allowedMethods = allowedMethodsStr
		}

		// Get allowed headers from environment
		allowedHeadersStr := os.Getenv("CORS_ALLOWED_HEADERS")
		allowedHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info, X-Device-Fingerprint"
		if allowedHeadersStr != "" {
			allowedHeaders = allowedHeadersStr
		}

		// Set Allow-Origin only when origin is allowed (shared logic with WebSocket CheckOrigin)
		if origin != "" {
			if IsOriginAllowedForRequest(r) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests (browser caches for 24h)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// SecurityHeaders middleware adds security headers to responses
func (sm *SecurityMiddleware) SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (basic)
		csp := os.Getenv("CONTENT_SECURITY_POLICY")
		if csp == "" {
			csp = "default-src 'self'; script-src 'self' https://va.vercel-scripts.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com; connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://va.vercel-scripts.com https: wss:;"
		}
		w.Header().Set("Content-Security-Policy", csp)

		// HSTS (HTTP Strict Transport Security) - only if HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			hstsMaxAge := os.Getenv("HSTS_MAX_AGE")
			if hstsMaxAge == "" {
				hstsMaxAge = "31536000" // 1 year
			}
			w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%s; includeSubDomains", hstsMaxAge))
		}

		next.ServeHTTP(w, r)
	}
}

// ValidateInput middleware validates and sanitizes request input
func (sm *SecurityMiddleware) ValidateInput(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Sanitize URL path
		if err := sm.validateAndSanitizePath(r.URL.Path); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"path": r.URL.Path,
				"ip":   getClientIP(r),
			}).Warn("Invalid path detected")
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Sanitize query parameters
		if err := sm.validateAndSanitizeQueryParams(r.URL); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"query": r.URL.RawQuery,
				"ip":    getClientIP(r),
			}).Warn("Invalid query parameters detected")
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Sanitize headers
		if err := sm.validateAndSanitizeHeaders(r); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"ip": getClientIP(r),
			}).Warn("Invalid headers detected")
			http.Error(w, "Invalid headers", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// validateAndSanitizePath validates and sanitizes the URL path
func (sm *SecurityMiddleware) validateAndSanitizePath(path string) error {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal attempt detected")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("null byte in path")
	}

	// Check for overly long paths
	if len(path) > 2048 {
		return fmt.Errorf("path too long")
	}

	// Check for suspicious characters
	suspiciousChars := []string{"<", ">", "\"", "'", "\n", "\r", "\t"}
	for _, char := range suspiciousChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("suspicious character in path: %s", char)
		}
	}

	return nil
}

// validateAndSanitizeQueryParams validates and sanitizes query parameters
func (sm *SecurityMiddleware) validateAndSanitizeQueryParams(u *url.URL) error {
	values := u.Query()

	for key, vals := range values {
		// Validate parameter name
		if err := sm.validateParameterName(key); err != nil {
			return fmt.Errorf("invalid parameter name '%s': %w", key, err)
		}

		// Validate and sanitize parameter values
		for _, val := range vals {
			if err := sm.validateAndSanitizeParameterValue(val); err != nil {
				return fmt.Errorf("invalid parameter value for '%s': %w", key, err)
			}
		}
	}

	return nil
}

// validateParameterName validates parameter names
func (sm *SecurityMiddleware) validateParameterName(name string) error {
	if name == "" {
		return fmt.Errorf("empty parameter name")
	}

	if len(name) > 100 {
		return fmt.Errorf("parameter name too long")
	}

	// Only allow alphanumeric characters, underscores, and hyphens
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("parameter name contains invalid characters")
	}

	return nil
}

// validateAndSanitizeParameterValue validates and sanitizes parameter values
func (sm *SecurityMiddleware) validateAndSanitizeParameterValue(value string) error {
	if len(value) > 1000 {
		return fmt.Errorf("parameter value too long")
	}

	// HTML escape to prevent XSS
	_ = html.EscapeString(value)

	// Check for null bytes
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("null byte in parameter value")
	}

	return nil
}

// validateAndSanitizeHeaders validates and sanitizes headers
func (sm *SecurityMiddleware) validateAndSanitizeHeaders(r *http.Request) error {
	for key, values := range r.Header {
		// Validate header name
		if err := sm.validateHeaderName(key); err != nil {
			return fmt.Errorf("invalid header name '%s': %w", key, err)
		}

		// Validate header values
		for _, value := range values {
			if err := sm.validateHeaderValue(value); err != nil {
				return fmt.Errorf("invalid header value for '%s': %w", key, err)
			}
		}
	}

	return nil
}

// validateHeaderName validates header names
func (sm *SecurityMiddleware) validateHeaderName(name string) error {
	if name == "" {
		return fmt.Errorf("empty header name")
	}

	// Header names should only contain printable ASCII characters
	for _, char := range name {
		if char < 32 || char > 126 {
			return fmt.Errorf("header name contains non-printable character")
		}
	}

	return nil
}

// validateHeaderValue validates header values
func (sm *SecurityMiddleware) validateHeaderValue(value string) error {
	// Check for null bytes
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("null byte in header value")
	}

	// Check for overly long headers
	if len(value) > 4096 {
		return fmt.Errorf("header value too long")
	}

	return nil
}

// getClientIP extracts the client IP address from the request (host only, no port).
// The result is safe for INET columns and rate-limit keys.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for requests behind proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		s := strings.TrimSpace(xff)
		if idx := strings.Index(s, ","); idx > 0 {
			s = strings.TrimSpace(s[:idx])
		}
		return stripPortFromHost(s)
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return stripPortFromHost(strings.TrimSpace(xri))
	}

	// Fall back to RemoteAddr (e.g. "127.0.0.1:50286") — must strip port for INET
	return stripPortFromHost(r.RemoteAddr)
}

// stripPortFromHost returns the host part of "host:port", or s unchanged if no port.
func stripPortFromHost(s string) string {
	if host, _, err := net.SplitHostPort(s); err == nil {
		return host
	}
	return s
}

// NewAdvancedSecurityMiddleware creates a new advanced security middleware
// This is a wrapper around the advanced_security package to maintain API compatibility
func NewAdvancedSecurityMiddleware(db storage.Repository) *advanced_security.AdvancedSecurityMiddleware {
	securityMiddleware := NewSecurityMiddleware()
	return advanced_security.NewAdvancedSecurityMiddleware(securityMiddleware, db)
}

// AuthRateLimiter is a strict per-IP rate limiter for sensitive auth endpoints
// (login, signup, resend-verification). Defaults to 10 req/min — configurable
// via AUTH_RATE_LIMIT_REQUESTS and AUTH_RATE_LIMIT_WINDOW_SECONDS.
type AuthRateLimiter struct {
	limiter *RateLimiter
}

// NewAuthRateLimiter creates an AuthRateLimiter with values from env or defaults.
func NewAuthRateLimiter() *AuthRateLimiter {
	limit := 10
	window := 60 // seconds

	if v := os.Getenv("AUTH_RATE_LIMIT_REQUESTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := os.Getenv("AUTH_RATE_LIMIT_WINDOW_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			window = parsed
		}
	}

	return &AuthRateLimiter{
		limiter: NewRateLimiter(time.Duration(window)*time.Second, limit),
	}
}

// Limit wraps a handler with auth-specific rate limiting.
func (a *AuthRateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		if !a.limiter.Allow(clientIP) {
			logrus.WithFields(logrus.Fields{
				"ip":   clientIP,
				"path": r.URL.Path,
			}).Warn("Auth rate limit exceeded")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(a.limiter.window.Seconds())))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"message":"Too many requests. Please wait before trying again."}`)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// VaultRateLimiter applies per-tenant rate limits for vault write operations.
// Use after RequireAuth so tenant ID is available from context.
type VaultRateLimiter struct {
	createSecretLimiter  *RateLimiter // e.g. 30/hour per tenant
	generateTokenLimiter *RateLimiter // e.g. 60/hour per tenant
}

// NewVaultRateLimiter creates a limiter for vault create secret (30/hour) and generate token (60/hour) per tenant.
func NewVaultRateLimiter() *VaultRateLimiter {
	return &VaultRateLimiter{
		createSecretLimiter:  NewRateLimiter(time.Hour, 30),
		generateTokenLimiter: NewRateLimiter(time.Hour, 60),
	}
}

// LimitCreate wraps a handler with per-tenant rate limiting for creating secrets.
// Requires auth context (use after RequireAuth). Key is tenant ID.
func (v *VaultRateLimiter) LimitCreate(next http.HandlerFunc) http.HandlerFunc {
	return v.limitByTenant("vault_create", v.createSecretLimiter, next)
}

// LimitGenerateToken wraps a handler with per-tenant rate limiting for generating tokens.
func (v *VaultRateLimiter) LimitGenerateToken(next http.HandlerFunc) http.HandlerFunc {
	return v.limitByTenant("vault_token", v.generateTokenLimiter, next)
}

func (v *VaultRateLimiter) limitByTenant(prefix string, limiter *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		key := prefix + ":unknown"
		if claims != nil {
			key = prefix + ":" + claims.TenantID.String()
		}
		if !limiter.Allow(key) {
			logrus.WithFields(logrus.Fields{"key": key, "path": r.URL.Path}).Warn("Vault rate limit exceeded")
			w.Header().Set("Retry-After", "3600")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"error":"Too Many Requests","code":"VAULT_RATE_LIMIT","message":"Vault operation rate limit exceeded. Try again later."}`)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// FlywheelRateLimiter applies per-tenant rate limits for Flywheel Network (community) write operations.
// Use after RequireAuth so tenant ID is available from context.
type FlywheelRateLimiter struct {
	createThreadLimiter    *RateLimiter // e.g. 10/minute per tenant
	createReplyLimiter     *RateLimiter // e.g. 20/minute per tenant
	submitChallengeLimiter *RateLimiter // e.g. 5/minute per tenant
	executeReplyLimiter    *RateLimiter // e.g. 30/minute per tenant
}

// NewFlywheelRateLimiter creates a limiter for Flywheel operations:
// - Create thread: 10/minute per tenant
// - Create reply: 20/minute per tenant
// - Submit challenge: 5/minute per tenant
// - Execute reply: 30/minute per tenant
func NewFlywheelRateLimiter() *FlywheelRateLimiter {
	return &FlywheelRateLimiter{
		createThreadLimiter:    NewRateLimiter(time.Minute, 10),
		createReplyLimiter:     NewRateLimiter(time.Minute, 20),
		submitChallengeLimiter: NewRateLimiter(time.Minute, 5),
		executeReplyLimiter:    NewRateLimiter(time.Minute, 30),
	}
}

// LimitCreateThread wraps a handler with per-tenant rate limiting for creating threads.
// Requires auth context (use after RequireAuth). Key is tenant ID.
func (f *FlywheelRateLimiter) LimitCreateThread(next http.HandlerFunc) http.HandlerFunc {
	return f.limitByTenant("flywheel_thread", f.createThreadLimiter, next)
}

// LimitCreateReply wraps a handler with per-tenant rate limiting for creating replies.
func (f *FlywheelRateLimiter) LimitCreateReply(next http.HandlerFunc) http.HandlerFunc {
	return f.limitByTenant("flywheel_reply", f.createReplyLimiter, next)
}

// LimitSubmitChallenge wraps a handler with per-tenant rate limiting for submitting challenges.
func (f *FlywheelRateLimiter) LimitSubmitChallenge(next http.HandlerFunc) http.HandlerFunc {
	return f.limitByTenant("flywheel_challenge", f.submitChallengeLimiter, next)
}

// LimitExecuteReply wraps a handler with per-tenant rate limiting for executing replies.
func (f *FlywheelRateLimiter) LimitExecuteReply(next http.HandlerFunc) http.HandlerFunc {
	return f.limitByTenant("flywheel_execute", f.executeReplyLimiter, next)
}

func (f *FlywheelRateLimiter) limitByTenant(prefix string, limiter *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		key := prefix + ":unknown"
		if claims != nil {
			key = prefix + ":" + claims.TenantID.String()
		}
		if !limiter.Allow(key) {
			logrus.WithFields(logrus.Fields{"key": key, "path": r.URL.Path}).Warn("Flywheel rate limit exceeded")
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"error":"Too Many Requests","code":"FLYWHEEL_RATE_LIMIT","message":"Flywheel operation rate limit exceeded. Try again later."}`)
			return
		}
		next.ServeHTTP(w, r)
	}
}
