package security

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// RequestValidator validates incoming requests
type RequestValidator struct {
	maxBodySize     int64
	allowedMethods  map[string]bool
	allowedHeaders  map[string]bool
	requiredHeaders map[string]bool
	rateLimiter     *RateLimiter
	mu              sync.RWMutex
}

// NewRequestValidator creates a new request validator
func NewRequestValidator(maxBodySize int64) *RequestValidator {
	return &RequestValidator{
		maxBodySize: maxBodySize,
		allowedMethods: map[string]bool{
			http.MethodGet:     true,
			http.MethodPost:    true,
			http.MethodPut:     true,
			http.MethodPatch:   true,
			http.MethodDelete:  true,
			http.MethodHead:    true,
			http.MethodOptions: true,
		},
		allowedHeaders: map[string]bool{
			"Authorization":        true,
			"Content-Type":         true,
			"X-Request-ID":         true,
			"X-Trace-ID":           true,
			"X-Span-ID":            true,
			"X-App-Key":            true,
			"X-FFLY-Timestamp":     true,
			"X-FFLY-Signature":     true,
			"Accept":               true,
			"Accept-Encoding":      true,
			"User-Agent":           true,
		},
		requiredHeaders: map[string]bool{
			"Content-Type": true,
		},
		rateLimiter: NewRateLimiter(100, time.Minute), // 100 requests per minute
	}
}

// ValidateRequest validates an HTTP request
func (v *RequestValidator) ValidateRequest(r *http.Request) error {
	// Validate method
	if !v.allowedMethods[r.Method] {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}

	// Validate content length
	if r.ContentLength > v.maxBodySize {
		return fmt.Errorf("request body too large: %d bytes (max: %d)", r.ContentLength, v.maxBodySize)
	}

	// Validate required headers
	for header := range v.requiredHeaders {
		if r.Header.Get(header) == "" {
			return fmt.Errorf("missing required header: %s", header)
		}
	}

	// Validate headers
	for header := range r.Header {
		if !v.allowedHeaders[header] {
			logrus.WithField("header", header).Warn("Unexpected header in request")
		}
	}

	// Validate content type for POST/PUT/PATCH
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			return fmt.Errorf("Content-Type header required for %s requests", r.Method)
		}
		if !strings.HasPrefix(contentType, "application/json") && !strings.HasPrefix(contentType, "multipart/form-data") {
			return fmt.Errorf("unsupported Content-Type: %s", contentType)
		}
	}

	return nil
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Refill tokens
	tokensToAdd := int(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	// Check if request is allowed
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// GetTokens returns the current number of tokens
func (rl *RateLimiter) GetTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.tokens
}

// IPAllowlist manages IP allowlisting
type IPAllowlist struct {
	allowedIPs map[string]bool
	blockedIPs map[string]bool
	mu         sync.RWMutex
}

// NewIPAllowlist creates a new IP allowlist
func NewIPAllowlist() *IPAllowlist {
	return &IPAllowlist{
		allowedIPs: make(map[string]bool),
		blockedIPs: make(map[string]bool),
	}
}

// AddAllowedIP adds an IP to the allowlist
func (a *IPAllowlist) AddAllowedIP(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.allowedIPs[ip] = true
}

// AddBlockedIP adds an IP to the blocklist
func (a *IPAllowlist) AddBlockedIP(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.blockedIPs[ip] = true
}

// IsAllowed checks if an IP is allowed
func (a *IPAllowlist) IsAllowed(ip string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Check blocklist first
	if a.blockedIPs[ip] {
		return false
	}

	// If allowlist is empty, allow all
	if len(a.allowedIPs) == 0 {
		return true
	}

	// Check allowlist
	return a.allowedIPs[ip]
}

// RemoveAllowedIP removes an IP from the allowlist
func (a *IPAllowlist) RemoveAllowedIP(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.allowedIPs, ip)
}

// RemoveBlockedIP removes an IP from the blocklist
func (a *IPAllowlist) RemoveBlockedIP(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.blockedIPs, ip)
}

// SQLInjectionDetector detects SQL injection attempts
type SQLInjectionDetector struct {
	patterns []*regexp.Regexp
}

// NewSQLInjectionDetector creates a new SQL injection detector
func NewSQLInjectionDetector() *SQLInjectionDetector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|select\s+\*\s+from)`),
		regexp.MustCompile(`(?i)(insert\s+into|update\s+\w+\s+set)`),
		regexp.MustCompile(`(?i)(delete\s+from|drop\s+table)`),
		regexp.MustCompile(`(?i)(--|;|\/\*|\*\/)`),
		regexp.MustCompile(`(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1)`),
		regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\()`),
	}

	return &SQLInjectionDetector{
		patterns: patterns,
	}
}

// Detect checks for SQL injection in input
func (d *SQLInjectionDetector) Detect(input string) bool {
	for _, pattern := range d.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// XSSDetector detects XSS attempts
type XSSDetector struct {
	patterns []*regexp.Regexp
}

// NewXSSDetector creates a new XSS detector
func NewXSSDetector() *XSSDetector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)<iframe[^>]*>`),
		regexp.MustCompile(`(?i)<object[^>]*>`),
		regexp.MustCompile(`(?i)<embed[^>]*>`),
	}

	return &XSSDetector{
		patterns: patterns,
	}
}

// Detect checks for XSS in input
func (d *XSSDetector) Detect(input string) bool {
	for _, pattern := range d.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// SecurityMiddleware provides security middleware for HTTP handlers
type SecurityMiddleware struct {
	validator    *RequestValidator
	ipAllowlist  *IPAllowlist
	sqlDetector  *SQLInjectionDetector
	xssDetector  *XSSDetector
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(maxBodySize int64) *SecurityMiddleware {
	return &SecurityMiddleware{
		validator:    NewRequestValidator(maxBodySize),
		ipAllowlist:  NewIPAllowlist(),
		sqlDetector:  NewSQLInjectionDetector(),
		xssDetector:  NewXSSDetector(),
	}
}

// Middleware returns an HTTP middleware function
func (m *SecurityMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = strings.Split(forwarded, ",")[0]
		}

		// Check IP allowlist
		if !m.ipAllowlist.IsAllowed(clientIP) {
			logrus.WithField("ip", clientIP).Warn("Blocked request from IP")
			apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
			return
		}

		// Validate request
		if err := m.validator.ValidateRequest(r); err != nil {
			logrus.WithError(err).Warn("Request validation failed")
			apierror.WriteError(w, apierror.NewBadRequest("Invalid request"))
			return
		}

		// Check rate limit
		if !m.validator.rateLimiter.Allow() {
			logrus.Warn("Rate limit exceeded")
			apierror.WriteError(w, apierror.NewRateLimited("Too Many Requests"))
			return
		}

		// Check for SQL injection in query params
		for key, values := range r.URL.Query() {
			for _, value := range values {
				if m.sqlDetector.Detect(value) {
					logrus.WithFields(logrus.Fields{
						"param": key,
						"value": value,
					}).Warn("SQL injection attempt detected")
					apierror.WriteError(w, apierror.NewBadRequest("Bad Request"))
					return
				}
			}
		}

		// Check for XSS in query params
		for key, values := range r.URL.Query() {
			for _, value := range values {
				if m.xssDetector.Detect(value) {
					logrus.WithFields(logrus.Fields{
						"param": key,
						"value": value,
					}).Warn("XSS attempt detected")
					apierror.WriteError(w, apierror.NewBadRequest("Bad Request"))
					return
				}
			}
		}

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
