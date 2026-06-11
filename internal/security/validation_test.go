package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestValidator(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024) // 1MB

	assert.NotNil(t, validator)
	assert.Equal(t, int64(1024*1024), validator.maxBodySize)
	assert.NotNil(t, validator.allowedMethods)
	assert.NotNil(t, validator.allowedHeaders)
	assert.NotNil(t, validator.requiredHeaders)
	assert.NotNil(t, validator.rateLimiter)
}

func TestRequestValidator_ValidateRequest_ValidGet(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_ValidPost(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_UnsupportedMethod(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodTrace, "/test", nil)

	err := validator.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method not allowed")
}

func TestRequestValidator_ValidateRequest_BodyTooLarge(t *testing.T) {
	validator := NewRequestValidator(10) // 10 bytes max

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("this is a long body"))
	req.Header.Set("Content-Type", "application/json")

	err := validator.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request body too large")
}

func TestRequestValidator_ValidateRequest_MissingContentType(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"key":"value"}`))

	err := validator.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Content-Type header required")
}

func TestRequestValidator_ValidateRequest_UnsupportedContentType(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "text/plain")

	err := validator.ValidateRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported Content-Type")
}

func TestRequestValidator_ValidateRequest_ValidMultipart(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary")

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_PutMethod(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodPut, "/test", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_PatchMethod(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodPatch, "/test", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_DeleteMethod(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_HeadMethod(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodHead, "/test", nil)

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestRequestValidator_ValidateRequest_OptionsMethod(t *testing.T) {
	validator := NewRequestValidator(1024 * 1024)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)

	err := validator.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(100, time.Minute)

	assert.NotNil(t, limiter)
	assert.Equal(t, 100, limiter.tokens)
	assert.Equal(t, 100, limiter.maxTokens)
	assert.Equal(t, time.Minute, limiter.refillRate)
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)

	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(), "Request %d should be allowed", i+1)
	}

	assert.False(t, limiter.Allow(), "6th request should be denied")
}

func TestRateLimiter_GetTokens(t *testing.T) {
	limiter := NewRateLimiter(10, time.Minute)

	assert.Equal(t, 10, limiter.GetTokens())

	limiter.Allow()
	limiter.Allow()

	assert.Equal(t, 8, limiter.GetTokens())
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := NewRateLimiter(5, 100*time.Millisecond)

	for i := 0; i < 5; i++ {
		limiter.Allow()
	}

	assert.False(t, limiter.Allow())

	time.Sleep(150 * time.Millisecond)

	assert.True(t, limiter.Allow())
}

func TestRateLimiter_MaxTokensNotExceeded(t *testing.T) {
	limiter := NewRateLimiter(5, 50*time.Millisecond)

	for i := 0; i < 10; i++ {
		limiter.Allow()
		time.Sleep(60 * time.Millisecond)
	}

	tokens := limiter.GetTokens()
	assert.LessOrEqual(t, tokens, 5)
}

func TestNewIPAllowlist(t *testing.T) {
	allowlist := NewIPAllowlist()

	assert.NotNil(t, allowlist)
	assert.NotNil(t, allowlist.allowedIPs)
	assert.NotNil(t, allowlist.blockedIPs)
}

func TestIPAllowlist_AddAllowedIP(t *testing.T) {
	allowlist := NewIPAllowlist()

	allowlist.AddAllowedIP("192.168.1.1")

	assert.True(t, allowlist.allowedIPs["192.168.1.1"])
}

func TestIPAllowlist_AddBlockedIP(t *testing.T) {
	allowlist := NewIPAllowlist()

	allowlist.AddBlockedIP("10.0.0.1")

	assert.True(t, allowlist.blockedIPs["10.0.0.1"])
}

func TestIPAllowlist_IsAllowed_NoRestrictions(t *testing.T) {
	allowlist := NewIPAllowlist()

	assert.True(t, allowlist.IsAllowed("192.168.1.1"))
	assert.True(t, allowlist.IsAllowed("10.0.0.1"))
	assert.True(t, allowlist.IsAllowed("8.8.8.8"))
}

func TestIPAllowlist_IsAllowed_BlockedIP(t *testing.T) {
	allowlist := NewIPAllowlist()

	allowlist.AddBlockedIP("10.0.0.1")

	assert.False(t, allowlist.IsAllowed("10.0.0.1"))
	assert.True(t, allowlist.IsAllowed("192.168.1.1"))
}

func TestIPAllowlist_IsAllowed_WithAllowlist(t *testing.T) {
	allowlist := NewIPAllowlist()

	allowlist.AddAllowedIP("192.168.1.1")
	allowlist.AddAllowedIP("192.168.1.2")

	assert.True(t, allowlist.IsAllowed("192.168.1.1"))
	assert.True(t, allowlist.IsAllowed("192.168.1.2"))
	assert.False(t, allowlist.IsAllowed("192.168.1.3"))
	assert.False(t, allowlist.IsAllowed("10.0.0.1"))
}

func TestIPAllowlist_RemoveAllowedIP(t *testing.T) {
	allowlist := NewIPAllowlist()

	allowlist.AddAllowedIP("192.168.1.1")
	allowlist.RemoveAllowedIP("192.168.1.1")

	assert.False(t, allowlist.IsAllowed("192.168.1.1"))
}

func TestIPAllowlist_RemoveBlockedIP(t *testing.T) {
	allowlist := NewIPAllowlist()

	allowlist.AddBlockedIP("10.0.0.1")
	allowlist.RemoveBlockedIP("10.0.0.1")

	assert.True(t, allowlist.IsAllowed("10.0.0.1"))
}

func TestNewSQLInjectionDetector(t *testing.T) {
	detector := NewSQLInjectionDetector()

	assert.NotNil(t, detector)
	assert.Len(t, detector.patterns, 6)
}

func TestSQLInjectionDetector_Detect(t *testing.T) {
	detector := NewSQLInjectionDetector()

	tests := []struct {
		input    string
		detected bool
	}{
		{"' OR '1'='1", true},
		{"'; DROP TABLE users; --", true},
		{" UNION SELECT * FROM passwords", true},
		{"1 AND 1=1", true},
		{"'; exec sp_executesql", true},
		{"normal text", false},
		{"hello world", false},
		{"12345", false},
		{"user@email.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := detector.Detect(tt.input)
			assert.Equal(t, tt.detected, result)
		})
	}
}

func TestNewXSSDetector(t *testing.T) {
	detector := NewXSSDetector()

	assert.NotNil(t, detector)
	assert.Len(t, detector.patterns, 6)
}

func TestXSSDetector_Detect(t *testing.T) {
	detector := NewXSSDetector()

	tests := []struct {
		input    string
		detected bool
	}{
		{"<script>alert('xss')</script>", true},
		{"javascript:alert('xss')", true},
		{"<iframe src='evil.com'></iframe>", true},
		{"<object data='evil.exe'></object>", true},
		{"onclick=alert('xss')", true},
		{"<embed src='evil.swf'></embed>", true},
		{"normal text", false},
		{"hello world", false},
		{"user@email.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := detector.Detect(tt.input)
			assert.Equal(t, tt.detected, result)
		})
	}
}

func TestNewSecurityMiddleware(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)

	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.validator)
	assert.NotNil(t, middleware.ipAllowlist)
	assert.NotNil(t, middleware.sqlDetector)
	assert.NotNil(t, middleware.xssDetector)
}

func TestSecurityMiddleware_Middleware(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityMiddleware_Middleware_BlockedIP(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)
	middleware.ipAllowlist.AddBlockedIP("192.168.1.1")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSecurityMiddleware_Middleware_RateLimitExceeded(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)
	middleware.validator = NewRequestValidator(1024 * 1024)

	limiter := NewRateLimiter(1, time.Minute)
	middleware.validator.rateLimiter = limiter

	limiter.Allow()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestSecurityMiddleware_Middleware_SQLInjection(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test?param=' OR '1'='1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurityMiddleware_Middleware_XSSAttempt(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test?param=<script>alert('xss')</script>", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurityMiddleware_Middleware_ValidRequest(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecurityMiddleware_Middleware_WithXForwardedFor(t *testing.T) {
	middleware := NewSecurityMiddleware(1024 * 1024)
	middleware.ipAllowlist.AddBlockedIP("192.168.1.100")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
