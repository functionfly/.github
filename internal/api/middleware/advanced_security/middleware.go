package advanced_security

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// SecurityMiddlewareInterface defines the interface for basic security middleware
type SecurityMiddlewareInterface interface {
	SecurityHeaders(next http.HandlerFunc) http.HandlerFunc
	CORSMiddleware(next http.HandlerFunc) http.HandlerFunc
	RequireHMACSignature(next http.HandlerFunc) http.HandlerFunc
}

// AdvancedSecurityMiddleware provides enhanced security features for the API gateway
type AdvancedSecurityMiddleware struct {
	securityMiddleware SecurityMiddlewareInterface

	// Rate limiting enhancements
	slidingWindowLimiter *SlidingWindowRateLimiter
	tokenBucketLimiter   *TokenBucketRateLimiter

	// DDoS protection
	requestFingerprinting *RequestFingerprinting
	botDetection          *BotDetection
	trafficAnalyzer       *TrafficAnalyzer

	// Traffic management
	circuitBreaker *CircuitBreaker
	requestQueue   *RequestQueue

	// Geo-blocking and reputation
	ipReputation *IPReputationChecker
	geoBlocker   *GeoBlocker

	// Advanced filtering
	sqlInjectionFilter  *SQLInjectionFilter
	xssFilter           *XSSFilter
	pathTraversalFilter *PathTraversalFilter

	// Configuration
	config     *AdvancedSecurityConfig
	allowedIPs map[string]bool

	logger *logrus.Logger
	stop   chan struct{} // Stop signal for background goroutines
}

// NewAdvancedSecurityMiddleware creates a new advanced security middleware
func NewAdvancedSecurityMiddleware(securityMiddleware SecurityMiddlewareInterface, db storage.Repository) *AdvancedSecurityMiddleware {
	config := &AdvancedSecurityConfig{
		SlidingWindowLimit:        getEnvInt("ADVANCED_SECURITY_SLIDING_WINDOW_LIMIT", 100),
		SlidingWindowWindow:       time.Duration(getEnvInt("ADVANCED_SECURITY_SLIDING_WINDOW_MINUTES", 1)) * time.Minute,
		TokenBucketRate:           getEnvFloat("ADVANCED_SECURITY_TOKEN_BUCKET_RATE", 10.0),
		TokenBucketBurst:          getEnvInt("ADVANCED_SECURITY_TOKEN_BUCKET_BURST", 20),
		EnableBotDetection:        getEnvBool("ADVANCED_SECURITY_ENABLE_BOT_DETECTION", true),
		EnableTrafficAnalysis:     getEnvBool("ADVANCED_SECURITY_ENABLE_TRAFFIC_ANALYSIS", true),
		SuspiciousThreshold:       getEnvInt("ADVANCED_SECURITY_SUSPICIOUS_THRESHOLD", 10),
		BlockDuration:             time.Duration(getEnvInt("ADVANCED_SECURITY_BLOCK_MINUTES", 15)) * time.Minute,
		CircuitBreakerThreshold:   getEnvFloat("ADVANCED_SECURITY_CIRCUIT_BREAKER_THRESHOLD", 0.5),
		CircuitBreakerTimeout:     time.Duration(getEnvInt("ADVANCED_SECURITY_CIRCUIT_BREAKER_MINUTES", 1)) * time.Minute,
		QueueSize:                 getEnvInt("ADVANCED_SECURITY_QUEUE_SIZE", 1000),
		QueueTimeout:              time.Duration(getEnvInt("ADVANCED_SECURITY_QUEUE_SECONDS", 30)) * time.Second,
		BlockedCountries:          getEnvStringSlice("ADVANCED_SECURITY_BLOCKED_COUNTRIES", ""),
		BlockedIPs:                getEnvStringSlice("ADVANCED_SECURITY_BLOCKED_IPS", ""),
		AllowedIPs:                getEnvStringSlice("ADVANCED_SECURITY_ALLOWED_IPS", ""),
		EnableSQLInjectionFilter:  getEnvBool("ADVANCED_SECURITY_ENABLE_SQL_INJECTION_FILTER", true),
		EnableXSSFilter:           getEnvBool("ADVANCED_SECURITY_ENABLE_XSS_FILTER", true),
		EnablePathTraversalFilter: getEnvBool("ADVANCED_SECURITY_ENABLE_PATH_TRAVERSAL_FILTER", true),
		MetricsEnabled:            getEnvBool("ADVANCED_SECURITY_METRICS_ENABLED", true),
	}

	// Initialize allowed IPs map
	allowedIPs := make(map[string]bool)
	for _, ip := range config.AllowedIPs {
		allowedIPs[ip] = true
	}

	asm := &AdvancedSecurityMiddleware{
		securityMiddleware: securityMiddleware,
		config:             config,
		allowedIPs:         allowedIPs,
		logger:             logrus.New(),
		stop:               make(chan struct{}),
	}

	// Initialize rate limiters
	asm.slidingWindowLimiter = &SlidingWindowRateLimiter{
		windows: make(map[string][]time.Time),
		window:  config.SlidingWindowWindow,
		limit:   config.SlidingWindowLimit,
		cleanup: time.Minute,
	}

	asm.tokenBucketLimiter = &TokenBucketRateLimiter{
		buckets:    make(map[string]*TokenBucket),
		rate:       config.TokenBucketRate,
		burst:      config.TokenBucketBurst,
		lastRefill: time.Now(),
	}

	// Initialize DDoS protection
	asm.requestFingerprinting = &RequestFingerprinting{
		fingerprints: make(map[string]*RequestPattern),
		window:       time.Minute,
	}

	asm.botDetection = &BotDetection{
		botSignatures:  make(map[string]bool),
		suspiciousIPs:  make(map[string]*BotActivity),
		detectionRules: asm.initBotDetectionRules(),
		rateWindows:    make(map[string][]time.Time),
		rateLimit:      50,               // Allow 50 requests per window
		rateWindow:     10 * time.Second, // 10-second window
	}

	asm.trafficAnalyzer = &TrafficAnalyzer{
		trafficStats:     make(map[string]*TrafficStats),
		window:           time.Minute * 5,
		anomalyThreshold: 3.0,
		stop:             asm.stop,
	}

	// Initialize traffic management
	asm.circuitBreaker = &CircuitBreaker{
		state:               "closed",
		failureCount:        0,
		successCount:        0,
		threshold:           config.CircuitBreakerThreshold,
		timeout:             config.CircuitBreakerTimeout,
		halfOpenMaxRequests: 3,
	}

	asm.requestQueue = &RequestQueue{
		queue:   make(chan *QueuedRequest, config.QueueSize),
		workers: 10,
		timeout: config.QueueTimeout,
	}

	// Initialize geo-blocking
	asm.ipReputation = &IPReputationChecker{
		reputationDB:  make(map[string]int),
		cacheDuration: time.Hour,
	}

	asm.geoBlocker = &GeoBlocker{
		blockedCountries: make(map[string]bool),
		blockedIPs:       make(map[string]bool),
		allowedIPs:       allowedIPs, // propagate from config
	}

	// Initialize filters
	asm.sqlInjectionFilter = &SQLInjectionFilter{
		patterns: asm.initSQLInjectionPatterns(),
	}

	asm.xssFilter = &XSSFilter{
		patterns: asm.initXSSPatterns(),
	}

	asm.pathTraversalFilter = &PathTraversalFilter{
		patterns: asm.initPathTraversalPatterns(),
	}

	// Start background processes
	go asm.cleanupRoutine()
	go asm.requestQueue.processQueue()
	go asm.trafficAnalyzer.monitorTraffic()
	go asm.botDetectionCleanupRoutine()

	return asm
}

// Stop stops all background goroutines in the advanced security middleware
func (asm *AdvancedSecurityMiddleware) Stop() {
	close(asm.stop)
	close(asm.requestQueue.queue)
}

// botDetectionCleanupRoutine periodically cleans up stale bot detection data
func (asm *AdvancedSecurityMiddleware) botDetectionCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			asm.botDetection.CleanupOldData()
		case <-asm.stop:
			return
		}
	}
}

// AdvancedRateLimit applies multiple rate limiting strategies.
// Rate limit keys are scoped per-tenant when a tenant ID is available in the JWT,
// falling back to IP-based limiting for unauthenticated requests.
// This prevents a single noisy tenant from degrading service for all others.
func (asm *AdvancedSecurityMiddleware) AdvancedRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// Build per-tenant rate limit key when possible
		rateLimitKey := buildPerTenantRateLimitKey(r, clientIP)

		// Apply sliding window rate limiting (per-tenant key)
		if !asm.slidingWindowLimiter.Allow(rateLimitKey) {
			asm.logRateLimit(rateLimitKey, "sliding_window", r)
			asm.respondRateLimited(w, r)
			return
		}

		// Apply token bucket rate limiting (per-tenant key)
		if !asm.tokenBucketLimiter.Allow(rateLimitKey) {
			asm.logRateLimit(rateLimitKey, "token_bucket", r)
			asm.respondRateLimited(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// buildPerTenantRateLimitKey constructs a rate limit key scoped per-tenant when possible.
// For authenticated requests: "tenant:{tenantID}:{normalizedPath}"
// For unauthenticated requests: "ip:{clientIP}"
func buildPerTenantRateLimitKey(r *http.Request, clientIP string) string {
	if tenantID := extractTenantIDFromJWT(r); tenantID != "" {
		return fmt.Sprintf("tenant:%s:%s", tenantID, normalizePathForRateLimit(r.URL.Path))
	}
	return fmt.Sprintf("ip:%s", clientIP)
}

// extractTenantIDFromJWT extracts the tenant_id claim from the JWT Authorization header.
// Returns empty string if no valid JWT is found. Does NOT verify the signature
// (verification is done by the auth middleware; here we only need the tenant ID for rate limiting).
func extractTenantIDFromJWT(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) < 8 || auth[:7] != "Bearer " {
		return ""
	}
	token := auth[7:]
	// JWT: header.payload.signature
	dot1 := indexByte(token, '.')
	if dot1 < 0 {
		return ""
	}
	dot2 := indexByte(token[dot1+1:], '.')
	if dot2 < 0 {
		return ""
	}
	payload := token[dot1+1 : dot1+1+dot2]

	// Decode base64url payload
	decoded := base64URLDecodeSimple(payload)
	if decoded == "" {
		return ""
	}
	return extractJSONField(decoded, "tenant_id")
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func base64URLDecodeSimple(s string) string {
	// Add padding
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	// Replace URL-safe chars
	b := []byte(s)
	for i, c := range b {
		if c == '-' {
			b[i] = '+'
		} else if c == '_' {
			b[i] = '/'
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return ""
	}
	return string(decoded)
}

func extractJSONField(json, field string) string {
	key := `"` + field + `":"`
	idx := 0
	for idx <= len(json)-len(key) {
		if json[idx:idx+len(key)] == key {
			start := idx + len(key)
			end := start
			for end < len(json) && json[end] != '"' {
				end++
			}
			return json[start:end]
		}
		idx++
	}
	return ""
}

// normalizePathForRateLimit normalizes a URL path to avoid per-resource key explosion.
// e.g., /v1/functions/abc-123/versions/1.0.0 -> /v1/functions
func normalizePathForRateLimit(path string) string {
	count := 0
	for i, c := range path {
		if c == '/' {
			count++
			if count == 3 {
				return path[:i]
			}
		}
	}
	return path
}

// DDoSProtection applies DDoS protection mechanisms
func (asm *AdvancedSecurityMiddleware) DDoSProtection(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip for health checks so Fly, K8s, and load balancers always get 200
		if r.URL.Path == "/health" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// The coming-soon waitlist endpoint must be publicly accessible.
		// Exempt it from bot/traffic blocking to avoid false positives.
		if r.Method == http.MethodPost && r.URL.Path == "/v1/feedback" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)

		// Never block or track loopback (localhost) so local dev works
		if isLoopbackIP(clientIP) {
			next.ServeHTTP(w, r)
			return
		}

		// Check if IP is blocked due to suspicious activity
		if asm.isIPBlocked(clientIP) {
			asm.logger.WithFields(logrus.Fields{
				"ip":     clientIP,
				"reason": "blocked_due_to_suspicious_activity",
			}).Warn("Blocked request from suspicious IP")
			apierror.WriteError(w, apierror.NewForbidden("Access denied"))
			return
		}

		// Fingerprint request for pattern analysis
		asm.requestFingerprinting.AnalyzeRequest(r)

		// Bot detection (skip for allowed IPs to avoid false positives on dev/CI machines)
		if asm.config.EnableBotDetection && !asm.allowedIPs[clientIP] {
			if isBot, reason := asm.botDetection.DetectBot(r); isBot {
				asm.logger.WithFields(logrus.Fields{
					"ip":         clientIP,
					"bot_reason": reason,
				}).Warn("Bot detected and blocked")
				asm.blockIP(clientIP, reason)
				apierror.WriteError(w, apierror.NewForbidden("Access denied"))
				return
			}
		}

		// Traffic analysis (skip for allowed IPs to avoid false positives on dev/CI machines)
		if asm.config.EnableTrafficAnalysis && !asm.allowedIPs[clientIP] {
			if isAttack, attackType := asm.trafficAnalyzer.DetectAttack(clientIP); isAttack {
				asm.logger.WithFields(logrus.Fields{
					"ip":          clientIP,
					"attack_type": attackType,
				}).Warn("Attack pattern detected")
				asm.blockIP(clientIP, attackType)
				apierror.WriteError(w, apierror.NewForbidden("Access denied"))
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

// TrafficManagement applies traffic management with circuit breaker and queuing
func (asm *AdvancedSecurityMiddleware) TrafficManagement(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// Skip circuit breaker and queuing for WebSocket upgrade requests
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		// Check circuit breaker
		if !asm.circuitBreaker.Allow() {
			asm.logger.Warn("Circuit breaker open, queuing request")
			asm.requestQueue.QueueRequest(w, r, next)
			return
		}

		// Wrap response writer to track success/failure and traffic
		rw := &responseWriterTracker{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			asm:            asm,
			startTime:      time.Now(),
			clientIP:       clientIP,
		}

		next.ServeHTTP(rw, r)
	}
}

// GeoBlocking blocks requests based on geographic location and IP reputation
func (asm *AdvancedSecurityMiddleware) GeoBlocking(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip for health checks so Fly, K8s, and load balancers always get 200
		if r.URL.Path == "/health" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// Keep coming-soon waitlist publicly accessible.
		if r.Method == http.MethodPost && r.URL.Path == "/v1/feedback" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)

		// Check IP allowlist/blocklist
		if !asm.geoBlocker.IsAllowed(clientIP) {
			asm.logger.WithFields(logrus.Fields{
				"ip":     clientIP,
				"reason": "ip_blocked",
			}).Warn("Request blocked by geo-blocking rules")
			apierror.WriteError(w, apierror.NewForbidden("Access denied"))
			return
		}

		// Check IP reputation
		if reputation := asm.ipReputation.GetReputation(clientIP); reputation < -50 {
			asm.logger.WithFields(logrus.Fields{
				"ip":         clientIP,
				"reputation": reputation,
			}).Warn("Request blocked due to poor IP reputation")
			apierror.WriteError(w, apierror.NewForbidden("Access denied"))
			return
		}

		next.ServeHTTP(w, r)
	}
}

// AdvancedInputValidation applies advanced input validation and filtering
func (asm *AdvancedSecurityMiddleware) AdvancedInputValidation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip input validation for registry publish: request body is function source code,
		// which often triggers false positives (e.g. "select", "<script>"). Match path with or without /v1 prefix.
		if r.Method == http.MethodPost && (r.URL.Path == "/v1/registry/publish" || strings.HasSuffix(r.URL.Path, "/registry/publish")) {
			next.ServeHTTP(w, r)
			return
		}
		// SQL injection filtering
		if asm.config.EnableSQLInjectionFilter {
			if asm.sqlInjectionFilter.Detect(r) {
				asm.logger.WithFields(logrus.Fields{
					"ip":          getClientIP(r),
					"attack_type": "sql_injection",
				}).Warn("SQL injection attempt detected")
				apierror.WriteError(w, apierror.NewBadRequest("Bad request"))
				return
			}
		}

		// XSS filtering
		if asm.config.EnableXSSFilter {
			if asm.xssFilter.Detect(r) {
				asm.logger.WithFields(logrus.Fields{
					"ip":          getClientIP(r),
					"attack_type": "xss",
				}).Warn("XSS attempt detected")
				apierror.WriteError(w, apierror.NewBadRequest("Bad request"))
				return
			}
		}

		// Path traversal filtering
		if asm.config.EnablePathTraversalFilter {
			if asm.pathTraversalFilter.Detect(r) {
				asm.logger.WithFields(logrus.Fields{
					"ip":          getClientIP(r),
					"attack_type": "path_traversal",
				}).Warn("Path traversal attempt detected")
				apierror.WriteError(w, apierror.NewBadRequest("Bad request"))
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

// Initialization methods
func (asm *AdvancedSecurityMiddleware) initBotDetectionRules() []BotDetectionRule {
	return []BotDetectionRule{
		{
			name:        "empty_user_agent",
			pattern:     regexp.MustCompile(`^$`),
			score:       20,
			description: "empty_user_agent",
		},
		{
			name:        "suspicious_user_agent",
			pattern:     regexp.MustCompile(`(?i)(bot|crawler|spider|scanner|python|wget)`),
			score:       15,
			description: "suspicious_user_agent",
		},
		// Note: rapid_requests detection is now handled by rate-based logic in DetectBotWithRateLimit
	}
}

func (asm *AdvancedSecurityMiddleware) initSQLInjectionPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|update\s+.*\s+set|drop\s+table|alter\s+table|--|#|/\*|\*/)`),
		regexp.MustCompile(`(?i)(script|javascript|vbscript|onload|onerror|onclick)`),
		regexp.MustCompile(`(?i)(\%27|\%22|\%3C|\%3E|\%28|\%29)`), // URL encoded
	}
}

func (asm *AdvancedSecurityMiddleware) initXSSPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?i)(<script|<iframe|<object|<embed)`),
		regexp.MustCompile(`(?i)(javascript:|vbscript:|data:)`),
		// Word boundary so "component=all" does not match as on+ent= (event-handler XSS).
		regexp.MustCompile(`(?i)\bon\w+\s*=`),
	}
}

func (asm *AdvancedSecurityMiddleware) initPathTraversalPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(\.\./|\.\.\\)`),
		regexp.MustCompile(`(\.\.)`),
		regexp.MustCompile(`(~|\.\.)`),
	}
}

// Helper methods
func (asm *AdvancedSecurityMiddleware) isIPBlocked(ip string) bool {
	if isLoopbackIP(ip) {
		return false
	}
	// Check if IP is in allowed list first
	if asm.allowedIPs[ip] {
		return false
	}

	asm.botDetection.mu.RLock()
	defer asm.botDetection.mu.RUnlock()

	if activity, exists := asm.botDetection.suspiciousIPs[ip]; exists {
		if activity.blockedUntil != nil && time.Now().Before(*activity.blockedUntil) {
			return true
		}
	}
	return false
}

func (asm *AdvancedSecurityMiddleware) blockIP(ip, reason string) {
	if isLoopbackIP(ip) {
		return
	}
	asm.botDetection.mu.Lock()
	defer asm.botDetection.mu.Unlock()

	activity, exists := asm.botDetection.suspiciousIPs[ip]
	if !exists {
		activity = &BotActivity{ip: ip}
		asm.botDetection.suspiciousIPs[ip] = activity
	}

	blockUntil := time.Now().Add(asm.config.BlockDuration)
	activity.blockedUntil = &blockUntil
	activity.detectionReason = reason
}

func (asm *AdvancedSecurityMiddleware) logRateLimit(ip, limiterType string, r *http.Request) {
	asm.logger.WithFields(logrus.Fields{
		"ip":      ip,
		"limiter": limiterType,
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Warn("Rate limit exceeded")
}

func (asm *AdvancedSecurityMiddleware) respondRateLimited(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-RateLimit-Limit", "100")
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
	w.Header().Set("Retry-After", "60")
	apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
}

func (asm *AdvancedSecurityMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			asm.cleanup()
		case <-asm.stop:
			return
		}
	}
}

func (asm *AdvancedSecurityMiddleware) cleanup() {
	now := time.Now()

	// Clean up old fingerprints
	asm.requestFingerprinting.mu.Lock()
	for ip, pattern := range asm.requestFingerprinting.fingerprints {
		if now.Sub(pattern.lastSeen) > time.Hour {
			delete(asm.requestFingerprinting.fingerprints, ip)
		}
	}
	asm.requestFingerprinting.mu.Unlock()

	// Clean up expired blocks
	asm.botDetection.mu.Lock()
	for ip, activity := range asm.botDetection.suspiciousIPs {
		if activity.blockedUntil != nil && now.After(*activity.blockedUntil) {
			delete(asm.botDetection.suspiciousIPs, ip)
		}
	}
	asm.botDetection.mu.Unlock()
}

// Delegate methods from SecurityMiddleware

// SecurityHeaders delegates to the embedded SecurityMiddleware
func (asm *AdvancedSecurityMiddleware) SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return asm.securityMiddleware.SecurityHeaders(next)
}

// CORSMiddleware delegates to the embedded SecurityMiddleware
func (asm *AdvancedSecurityMiddleware) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return asm.securityMiddleware.CORSMiddleware(next)
}

// RequireHMACSignature delegates to the embedded SecurityMiddleware
func (asm *AdvancedSecurityMiddleware) RequireHMACSignature(next http.HandlerFunc) http.HandlerFunc {
	return asm.securityMiddleware.RequireHMACSignature(next)
}

// responseWriterTracker implementation
func (rwt *responseWriterTracker) WriteHeader(code int) {
	if !rwt.written {
		rwt.statusCode = code
		rwt.written = true

		// Track success/failure for circuit breaker
		if code >= 500 {
			rwt.asm.circuitBreaker.RecordFailure()
		} else if code < 400 {
			rwt.asm.circuitBreaker.RecordSuccess()
		}

		// Record traffic statistics
		responseTime := time.Since(rwt.startTime)
		// Exclude client-error codes that are normal during auth flows:
		//   401 = unauthenticated (session bootstrap, CSRF check, last-login with no session)
		//   404 = resource not found (endpoint not deployed, SPA catch-all fallback)
		// These are NOT indicators of an attack and would otherwise inflate error rates
		// to trigger false-positive "high_error_rate" blocks on legitimate browser clients.
		isError := code >= 400 && code != http.StatusUnauthorized && code != http.StatusNotFound
		rwt.asm.trafficAnalyzer.RecordRequest(rwt.clientIP, responseTime, isError)

		rwt.ResponseWriter.WriteHeader(code)
	}
}

func (rwt *responseWriterTracker) Write(data []byte) (int, error) {
	if !rwt.written {
		rwt.WriteHeader(http.StatusOK)
	}
	return rwt.ResponseWriter.Write(data)
}
