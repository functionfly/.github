package advanced_security

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
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
	botDetection         *BotDetection
	trafficAnalyzer      *TrafficAnalyzer

	// Traffic management
	circuitBreaker       *CircuitBreaker
	requestQueue         *RequestQueue

	// Geo-blocking and reputation
	ipReputation         *IPReputationChecker
	geoBlocker           *GeoBlocker

	// Advanced filtering
	sqlInjectionFilter   *SQLInjectionFilter
	xssFilter           *XSSFilter
	pathTraversalFilter *PathTraversalFilter

	// Configuration
	config              *AdvancedSecurityConfig
	allowedIPs          map[string]bool

	logger              *logrus.Logger
}

// NewAdvancedSecurityMiddleware creates a new advanced security middleware
func NewAdvancedSecurityMiddleware(securityMiddleware SecurityMiddlewareInterface, db storage.Repository) *AdvancedSecurityMiddleware {
	config := &AdvancedSecurityConfig{
		SlidingWindowLimit:       getEnvInt("ADVANCED_SECURITY_SLIDING_WINDOW_LIMIT", 100),
		SlidingWindowWindow:      time.Duration(getEnvInt("ADVANCED_SECURITY_SLIDING_WINDOW_MINUTES", 1)) * time.Minute,
		TokenBucketRate:         getEnvFloat("ADVANCED_SECURITY_TOKEN_BUCKET_RATE", 10.0),
		TokenBucketBurst:        getEnvInt("ADVANCED_SECURITY_TOKEN_BUCKET_BURST", 20),
		EnableBotDetection:      getEnvBool("ADVANCED_SECURITY_ENABLE_BOT_DETECTION", true),
		EnableTrafficAnalysis:   getEnvBool("ADVANCED_SECURITY_ENABLE_TRAFFIC_ANALYSIS", true),
		SuspiciousThreshold:     getEnvInt("ADVANCED_SECURITY_SUSPICIOUS_THRESHOLD", 10),
		BlockDuration:           time.Duration(getEnvInt("ADVANCED_SECURITY_BLOCK_MINUTES", 15)) * time.Minute,
		CircuitBreakerThreshold: getEnvFloat("ADVANCED_SECURITY_CIRCUIT_BREAKER_THRESHOLD", 0.5),
		CircuitBreakerTimeout:   time.Duration(getEnvInt("ADVANCED_SECURITY_CIRCUIT_BREAKER_MINUTES", 1)) * time.Minute,
		QueueSize:               getEnvInt("ADVANCED_SECURITY_QUEUE_SIZE", 1000),
		QueueTimeout:            time.Duration(getEnvInt("ADVANCED_SECURITY_QUEUE_SECONDS", 30)) * time.Second,
		BlockedCountries:        getEnvStringSlice("ADVANCED_SECURITY_BLOCKED_COUNTRIES", ""),
		BlockedIPs:             getEnvStringSlice("ADVANCED_SECURITY_BLOCKED_IPS", ""),
		AllowedIPs:             getEnvStringSlice("ADVANCED_SECURITY_ALLOWED_IPS", ""),
		EnableSQLInjectionFilter: getEnvBool("ADVANCED_SECURITY_ENABLE_SQL_INJECTION_FILTER", true),
		EnableXSSFilter:         getEnvBool("ADVANCED_SECURITY_ENABLE_XSS_FILTER", true),
		EnablePathTraversalFilter: getEnvBool("ADVANCED_SECURITY_ENABLE_PATH_TRAVERSAL_FILTER", true),
		MetricsEnabled:         getEnvBool("ADVANCED_SECURITY_METRICS_ENABLED", true),
	}

	// Initialize allowed IPs map
	allowedIPs := make(map[string]bool)
	for _, ip := range config.AllowedIPs {
		allowedIPs[ip] = true
	}

	asm := &AdvancedSecurityMiddleware{
		securityMiddleware: securityMiddleware,
		config:            config,
		allowedIPs:        allowedIPs,
		logger:            logrus.New(),
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
	}

	asm.trafficAnalyzer = &TrafficAnalyzer{
		trafficStats:     make(map[string]*TrafficStats),
		window:           time.Minute * 5,
		anomalyThreshold: 3.0,
	}

	// Initialize traffic management
	asm.circuitBreaker = &CircuitBreaker{
		state:                "closed",
		failureCount:         0,
		successCount:         0,
		threshold:            config.CircuitBreakerThreshold,
		timeout:              config.CircuitBreakerTimeout,
		halfOpenMaxRequests:  3,
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
		allowedIPs:       make(map[string]bool),
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

	return asm
}

// AdvancedRateLimit applies multiple rate limiting strategies
func (asm *AdvancedSecurityMiddleware) AdvancedRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// Apply sliding window rate limiting
		if !asm.slidingWindowLimiter.Allow(clientIP) {
			asm.logRateLimit(clientIP, "sliding_window", r)
			asm.respondRateLimited(w, r)
			return
		}

		// Apply token bucket rate limiting
		if !asm.tokenBucketLimiter.Allow(clientIP) {
			asm.logRateLimit(clientIP, "token_bucket", r)
			asm.respondRateLimited(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// DDoSProtection applies DDoS protection mechanisms
func (asm *AdvancedSecurityMiddleware) DDoSProtection(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// Check if IP is blocked due to suspicious activity
		if asm.isIPBlocked(clientIP) {
			asm.logger.WithFields(logrus.Fields{
				"ip": clientIP,
				"reason": "blocked_due_to_suspicious_activity",
			}).Warn("Blocked request from suspicious IP")
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Fingerprint request for pattern analysis
		asm.requestFingerprinting.AnalyzeRequest(r)

		// Bot detection
		if asm.config.EnableBotDetection {
			if isBot, reason := asm.botDetection.DetectBot(r); isBot {
				asm.logger.WithFields(logrus.Fields{
					"ip": clientIP,
					"bot_reason": reason,
				}).Warn("Bot detected and blocked")
				asm.blockIP(clientIP, reason)
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		}

		// Traffic analysis
		if asm.config.EnableTrafficAnalysis {
			if isAttack, attackType := asm.trafficAnalyzer.DetectAttack(clientIP); isAttack {
				asm.logger.WithFields(logrus.Fields{
					"ip": clientIP,
					"attack_type": attackType,
				}).Warn("Attack pattern detected")
				asm.blockIP(clientIP, attackType)
				http.Error(w, "Access denied", http.StatusForbidden)
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
			asm:           asm,
			startTime:     time.Now(),
			clientIP:      clientIP,
		}

		next.ServeHTTP(rw, r)
	}
}

// GeoBlocking blocks requests based on geographic location and IP reputation
func (asm *AdvancedSecurityMiddleware) GeoBlocking(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// Check IP allowlist/blocklist
		if !asm.geoBlocker.IsAllowed(clientIP) {
			asm.logger.WithFields(logrus.Fields{
				"ip": clientIP,
				"reason": "ip_blocked",
			}).Warn("Request blocked by geo-blocking rules")
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check IP reputation
		if reputation := asm.ipReputation.GetReputation(clientIP); reputation < -50 {
			asm.logger.WithFields(logrus.Fields{
				"ip": clientIP,
				"reputation": reputation,
			}).Warn("Request blocked due to poor IP reputation")
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// AdvancedInputValidation applies advanced input validation and filtering
func (asm *AdvancedSecurityMiddleware) AdvancedInputValidation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// SQL injection filtering
		if asm.config.EnableSQLInjectionFilter {
			if asm.sqlInjectionFilter.Detect(r) {
				asm.logger.WithFields(logrus.Fields{
					"ip": getClientIP(r),
					"attack_type": "sql_injection",
				}).Warn("SQL injection attempt detected")
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}

		// XSS filtering
		if asm.config.EnableXSSFilter {
			if asm.xssFilter.Detect(r) {
				asm.logger.WithFields(logrus.Fields{
					"ip": getClientIP(r),
					"attack_type": "xss",
				}).Warn("XSS attempt detected")
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}

		// Path traversal filtering
		if asm.config.EnablePathTraversalFilter {
			if asm.pathTraversalFilter.Detect(r) {
				asm.logger.WithFields(logrus.Fields{
					"ip": getClientIP(r),
					"attack_type": "path_traversal",
				}).Warn("Path traversal attempt detected")
				http.Error(w, "Bad request", http.StatusBadRequest)
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
			pattern:     regexp.MustCompile(`(?i)(bot|crawler|spider|scanner|python|curl|wget)`),
			score:       15,
			description: "suspicious_user_agent",
		},
		{
			name:        "rapid_requests",
			pattern:     regexp.MustCompile(`.*`),
			score:       10,
			description: "rapid_requests",
		},
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
		regexp.MustCompile(`(?i)(on\w+\s*=)`),
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
		"ip": ip,
		"limiter": limiterType,
		"method": r.Method,
		"path": r.URL.Path,
	}).Warn("Rate limit exceeded")
}

func (asm *AdvancedSecurityMiddleware) respondRateLimited(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-RateLimit-Limit", "100")
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
	w.Header().Set("Retry-After", "60")
	http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
}

func (asm *AdvancedSecurityMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for range ticker.C {
		asm.cleanup()
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
		isError := code >= 400
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