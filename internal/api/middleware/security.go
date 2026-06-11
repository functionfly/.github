package middleware

import (
	"encoding/json"
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
	lastCleanup time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	rl := &RateLimiter{
		requests:    make(map[string][]time.Time),
		window:      window,
		limit:       limit,
		lastCleanup: time.Now(),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	windowStart := now.Add(-rl.window)
	for key, timestamps := range rl.requests {
		validCount := 0
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validCount++
			}
		}
		if validCount == 0 {
			delete(rl.requests, key)
		}
	}
}

// Allow checks if a request from the given key should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean up old entries for this key
	if timestamps, exists := rl.requests[key]; exists {
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
			logrus.Error("API_SHARED_SECRET not configured — rejecting HMAC-required request")
			http.Error(w, "Service misconfigured: API_SHARED_SECRET required", http.StatusInternalServerError)
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

// SetCORSHeaders sets Access-Control-Allow-Origin on the response based on the
// configured allowlist. Handlers that set CORS headers outside the global
// middleware should call this instead of hardcoding "*".
func SetCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	allowed := getCORSAllowedOrigins()
	if origin != "" && IsOriginAllowedForRequest(r) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if len(allowed) == 1 && allowed[0] == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
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
		// SECURITY: If no CORS_ALLOWED_ORIGINS is set, deny all cross-origin requests in production.
		// Only allow localhost in development mode.
		if isDev {
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "https://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") ||
				strings.HasPrefix(origin, "https://127.0.0.1:")
		}
		return false
	}
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

// CORSMiddleware adds CORS headers to responses
func (sm *SecurityMiddleware) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := getCORSAllowedOrigins()
		origin := r.Header.Get("Origin")
		// SECURITY: In production, empty allowlist must deny all cross-origin requests.
		// localhost exception only allowed in DEVELOPMENT mode.
		if len(allowedOrigins) == 0 && origin != "" {
			isDev := os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development"
			isLocalhost := strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
			if !isDev {
				logrus.Error("CORS_ALLOWED_ORIGINS is empty in production — rejecting cross-origin request")
				http.Error(w, "CORS not configured", http.StatusForbidden)
				return
			}
			// In dev mode, only allow localhost
			if !isLocalhost {
				logrus.Error("CORS_ALLOWED_ORIGINS is empty in development — rejecting non-localhost origin")
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
		allowedHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature, x-neon-client-info, X-Device-Fingerprint, X-Environment"
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

// BodySizeLimitMiddleware creates middleware that limits request body size.
// Use this to prevent DoS attacks via large request bodies.
func BodySizeLimitMiddleware(maxBytes int64) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
next.ServeHTTP(w, r)
	}
}

// ConsciousnessRateLimiter applies per-tenant rate limits for consciousness operations.
type ConsciousnessRateLimiter struct {
	scoreLimiter      *RateLimiter // GET /score - 100/hour per tenant
	listInsightsLimiter *RateLimiter // GET /insights - 60/minute per tenant
	runAnalysisLimiter *RateLimiter // POST /run - 5/hour per tenant (expensive operation)
}

// NewConsciousnessRateLimiter creates a limiter for consciousness operations per tenant.
func NewConsciousnessRateLimiter() *ConsciousnessRateLimiter {
	return &ConsciousnessRateLimiter{
		scoreLimiter:       NewRateLimiter(time.Hour, 100),    // 100 score checks per hour per tenant
		listInsightsLimiter: NewRateLimiter(time.Minute, 60), // 60 insight queries per minute per tenant
		runAnalysisLimiter: NewRateLimiter(time.Hour, 5),     // 5 analysis runs per hour per tenant
	}
}

// LimitScore wraps a handler with per-tenant rate limiting for awareness score.
func (c *ConsciousnessRateLimiter) LimitScore(next http.HandlerFunc) http.HandlerFunc {
	return c.limitByTenant("consciousness_score", c.scoreLimiter, next)
}

// LimitListInsights wraps a handler with per-tenant rate limiting for listing insights.
func (c *ConsciousnessRateLimiter) LimitListInsights(next http.HandlerFunc) http.HandlerFunc {
	return c.limitByTenant("consciousness_insights", c.listInsightsLimiter, next)
}

// LimitRunAnalysis wraps a handler with per-tenant rate limiting for running analysis.
func (c *ConsciousnessRateLimiter) LimitRunAnalysis(next http.HandlerFunc) http.HandlerFunc {
	return c.limitByTenant("consciousness_run", c.runAnalysisLimiter, next)
}

func (c *ConsciousnessRateLimiter) limitByTenant(prefix string, limiter *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		key := fmt.Sprintf("%s:tenant:%s", prefix, claims.TenantID.String())

		if !limiter.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(limiter.window.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			logrus.WithFields(logrus.Fields{
				"tenant_id": claims.TenantID.String(),
				"prefix":    prefix,
				"ip":        getClientIP(r),
			}).Warn("Consciousness operation rate limit exceeded")
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Too many consciousness operations. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
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

// getClientIP extracts the client IP address from the request.
// SECURITY: For rate limiting and security purposes, X-Forwarded-For is NOT trusted
// because it can be easily spoofed by attackers. Only trust proxy headers when the
// connection comes from a known trusted proxy (127.0.0.1 or configured proxy IPs).
func getClientIP(r *http.Request) string {
	// Always validate against trusted proxy list for forwarded headers
	remoteAddr := r.RemoteAddr
	xff := r.Header.Get("X-Forwarded-For")

	// Only trust X-Forwarded-For from known trusted proxies
	// In production, this should be configured to only trust your load balancer/CDN IPs
	if xff != "" && isTrustedProxy(r) {
		// X-Forwarded-For can contain multiple IPs, take the first one (original client)
		s := strings.TrimSpace(xff)
		if idx := strings.Index(s, ","); idx > 0 {
			s = strings.TrimSpace(s[:idx])
		}
		return stripPortFromHost(s)
	}

	// Check X-Real-IP header only from trusted proxy
	xri := r.Header.Get("X-Real-IP")
	if xri != "" && isTrustedProxy(r) {
		return stripPortFromHost(strings.TrimSpace(xri))
	}

	// Fall back to RemoteAddr (e.g. "127.0.0.1:50286") — must strip port for INET
	return stripPortFromHost(remoteAddr)
}

// isTrustedProxy checks if the request comes from a known trusted proxy
func isTrustedProxy(r *http.Request) bool {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}

	// Trust loopback addresses (localhost connections from nginx, apache, etc.)
	if remoteHost == "127.0.0.1" || remoteHost == "::1" || remoteHost == "localhost" {
		return true
	}

	// Check for explicitly configured trusted proxies
	trustedProxies := os.Getenv("TRUSTED_PROXY_IPS")
	if trustedProxies != "" {
		for _, proxy := range strings.Split(trustedProxies, ",") {
			proxy = strings.TrimSpace(proxy)
			if remoteHost == proxy {
				return true
			}
		}
	}

	return false
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

// MFARateLimiter applies rate limiting to MFA verification endpoints to prevent brute force attacks
// on TOTP codes and backup codes.
type MFARateLimiter struct {
	limiter *RateLimiter
}

// NewMFARateLimiter creates an MFA rate limiter with stricter limits than auth endpoints.
// Defaults to 5 attempts per minute per user to prevent brute forcing 6-digit TOTP codes.
func NewMFARateLimiter() *MFARateLimiter {
	limit := 5
	window := 60 // seconds

	if v := os.Getenv("MFA_RATE_LIMIT_REQUESTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := os.Getenv("MFA_RATE_LIMIT_WINDOW_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			window = parsed
		}
	}

	return &MFARateLimiter{
		limiter: NewRateLimiter(time.Duration(window)*time.Second, limit),
	}
}

// LimitVerify wraps an MFA verification handler with rate limiting.
// Uses user ID from context (set by auth middleware) as the rate limit key.
func (m *MFARateLimiter) LimitVerify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by auth middleware)
		claims := GetUserFromContext(r)
		if claims == nil {
			// No auth context, allow through (auth middleware will reject)
			next.ServeHTTP(w, r)
			return
		}

		// Use user ID as rate limit key to prevent per-user brute force
		key := claims.UserID.String()

		if !m.limiter.Allow(key) {
			logrus.WithFields(logrus.Fields{
				"user_id": claims.UserID.String(),
				"path":    r.URL.Path,
			}).Warn("MFA verification rate limit exceeded")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(m.limiter.window.Seconds())))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"message":"Too many MFA verification attempts. Please wait before trying again."}`)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// VaultRateLimiter applies per-tenant rate limits for vault operations.
// Use after RequireAuth so tenant ID is available from context.
type VaultRateLimiter struct {
	createSecretLimiter  *RateLimiter // e.g. 30/hour per tenant
	generateTokenLimiter *RateLimiter // e.g. 60/hour per tenant
	readSecretLimiter    *RateLimiter // e.g. 200/hour per tenant
	listSecretsLimiter   *RateLimiter // e.g. 100/hour per tenant
}

// NewVaultRateLimiter creates a limiter for vault operations per tenant.
func NewVaultRateLimiter() *VaultRateLimiter {
	return &VaultRateLimiter{
		createSecretLimiter:  NewRateLimiter(time.Hour, 30),
		generateTokenLimiter: NewRateLimiter(time.Hour, 60),
		readSecretLimiter:    NewRateLimiter(time.Hour, 200),
		listSecretsLimiter:   NewRateLimiter(time.Hour, 100),
	}
}

// LimitCreate wraps a handler with per-tenant rate limiting for creating secrets.
func (v *VaultRateLimiter) LimitCreate(next http.HandlerFunc) http.HandlerFunc {
	return v.limitByTenant("vault_create", v.createSecretLimiter, next)
}

// LimitGenerateToken wraps a handler with per-tenant rate limiting for generating tokens.
func (v *VaultRateLimiter) LimitGenerateToken(next http.HandlerFunc) http.HandlerFunc {
	return v.limitByTenant("vault_token", v.generateTokenLimiter, next)
}

// LimitRead wraps a handler with per-tenant rate limiting for reading a secret.
func (v *VaultRateLimiter) LimitRead(next http.HandlerFunc) http.HandlerFunc {
	return v.limitByTenant("vault_read", v.readSecretLimiter, next)
}

// LimitList wraps a handler with per-tenant rate limiting for listing secrets.
func (v *VaultRateLimiter) LimitList(next http.HandlerFunc) http.HandlerFunc {
	return v.limitByTenant("vault_list", v.listSecretsLimiter, next)
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

// ProviderRateLimiter limits provider connect/disconnect/test operations per tenant.
type ProviderRateLimiter struct {
	connectLimiter    *RateLimiter // Limit: 10/hour per tenant for connecting new providers
	disconnectLimiter *RateLimiter // Limit: 20/hour per tenant for disconnecting providers
	testLimiter       *RateLimiter // Limit: 30/minute per tenant for testing connections
}

// NewProviderRateLimiter creates a limiter for provider operations with sensible defaults.
func NewProviderRateLimiter() *ProviderRateLimiter {
	return &ProviderRateLimiter{
		connectLimiter:    NewRateLimiter(time.Hour, 10),   // 10 connects per hour per tenant
		disconnectLimiter: NewRateLimiter(time.Hour, 20),   // 20 disconnects per hour per tenant
		testLimiter:       NewRateLimiter(time.Minute, 30), // 30 tests per minute per tenant
	}
}

// LimitConnect wraps a handler with per-tenant rate limiting for connecting providers.
func (p *ProviderRateLimiter) LimitConnect(next http.HandlerFunc) http.HandlerFunc {
	return p.limitByTenant("provider_connect", p.connectLimiter, next)
}

// LimitDisconnect wraps a handler with per-tenant rate limiting for disconnecting providers.
func (p *ProviderRateLimiter) LimitDisconnect(next http.HandlerFunc) http.HandlerFunc {
	return p.limitByTenant("provider_disconnect", p.disconnectLimiter, next)
}

// LimitTest wraps a handler with per-tenant rate limiting for testing provider connections.
func (p *ProviderRateLimiter) LimitTest(next http.HandlerFunc) http.HandlerFunc {
	return p.limitByTenant("provider_test", p.testLimiter, next)
}

func (p *ProviderRateLimiter) limitByTenant(prefix string, limiter *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		key := fmt.Sprintf("%s:tenant:%s", prefix, claims.TenantID.String())

		if !limiter.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(limiter.window.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			logrus.WithFields(logrus.Fields{
				"tenant_id": claims.TenantID.String(),
				"prefix":    prefix,
				"ip":        getClientIP(r),
			}).Warn("Provider operation rate limit exceeded")
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Too many provider operations. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// WalletRateLimiter limits wallet operations per user/agent to prevent abuse
// and protect financial operations from brute force or spam
type WalletRateLimiter struct {
	balanceCheckLimiter *RateLimiter // Limit: 60/minute per wallet for balance checks
	topUpLimiter        *RateLimiter // Limit: 5/hour per wallet for top-up requests
	adjustmentLimiter   *RateLimiter // Limit: 10/hour per admin for adjustments
}

// NewWalletRateLimiter creates a limiter for wallet operations with sensible defaults
func NewWalletRateLimiter() *WalletRateLimiter {
	return &WalletRateLimiter{
		balanceCheckLimiter: NewRateLimiter(time.Minute, 60), // 60 balance checks per minute per wallet
		topUpLimiter:        NewRateLimiter(time.Hour, 5),    // 5 top-ups per hour per wallet
		adjustmentLimiter:   NewRateLimiter(time.Hour, 10),   // 10 adjustments per hour per admin
	}
}

// LimitBalanceCheck wraps a handler with rate limiting for balance check operations
// Uses wallet ID as the key
func (wr *WalletRateLimiter) LimitBalanceCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Use user ID as the wallet identifier key
		key := fmt.Sprintf("wallet_balance_check:%s", claims.UserID.String())

		if !wr.balanceCheckLimiter.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			logrus.WithFields(logrus.Fields{
				"user_id": claims.UserID.String(),
				"path":    r.URL.Path,
			}).Warn("Wallet balance check rate limit exceeded")
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"code":    "WALLET_RATE_LIMIT",
				"message": "Too many wallet balance checks. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// LimitTopUp wraps a handler with rate limiting for wallet top-up operations
// Uses wallet/user ID as the key
func (wr *WalletRateLimiter) LimitTopUp(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Use tenant + user ID as the key
		key := fmt.Sprintf("wallet_topup:%s:%s", claims.TenantID.String(), claims.UserID.String())

		if !wr.topUpLimiter.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(wr.topUpLimiter.window.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			logrus.WithFields(logrus.Fields{
				"user_id":   claims.UserID.String(),
				"tenant_id": claims.TenantID.String(),
				"path":      r.URL.Path,
			}).Warn("Wallet top-up rate limit exceeded")
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"code":    "WALLET_TOPUP_RATE_LIMIT",
				"message": "Too many wallet top-up attempts. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// LimitAdminAdjustment wraps a handler with rate limiting for admin wallet adjustments
// Uses admin user ID as the key
func (wr *WalletRateLimiter) LimitAdminAdjustment(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Use admin user ID as the key
		key := fmt.Sprintf("wallet_admin_adjustment:%s", claims.UserID.String())

		if !wr.adjustmentLimiter.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(wr.adjustmentLimiter.window.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			logrus.WithFields(logrus.Fields{
				"admin_id": claims.UserID.String(),
				"path":     r.URL.Path,
			}).Warn("Wallet admin adjustment rate limit exceeded")
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"code":    "WALLET_ADJUSTMENT_RATE_LIMIT",
				"message": "Too many wallet adjustments. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}
