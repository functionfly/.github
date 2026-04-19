package middleware

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDistributedRateLimiter(t *testing.T) {
	// Without Redis (should not be enabled)
	limiter := NewDistributedRateLimiter(nil, time.Minute, 10, "test")
	assert.NotNil(t, limiter)
	assert.False(t, limiter.IsEnabled())
	assert.Equal(t, 10, limiter.GetLimit())
	assert.Equal(t, time.Minute, limiter.GetWindow())
}

func TestDistributedRateLimiter_DisabledByEnv(t *testing.T) {
	// Set environment variable to disable
	os.Setenv("DISTRIBUTED_RATE_LIMITER_DISABLED", "true")
	defer os.Unsetenv("DISTRIBUTED_RATE_LIMITER_DISABLED")

	// Even with Redis client, should not be enabled
	limiter := NewDistributedRateLimiter(nil, time.Minute, 10, "test")
	assert.False(t, limiter.IsEnabled())
}

func TestNewDistributedAuthRateLimiter_Defaults(t *testing.T) {
	limiter := NewDistributedAuthRateLimiter(nil)
	assert.NotNil(t, limiter)
	assert.False(t, limiter.IsEnabled())
	assert.Equal(t, 10, limiter.GetLimit())
	assert.Equal(t, 60*time.Second, limiter.GetWindow())
}

func TestNewDistributedAuthRateLimiter_FromEnv(t *testing.T) {
	// Set custom values
	os.Setenv("AUTH_RATE_LIMIT_REQUESTS", "20")
	os.Setenv("AUTH_RATE_LIMIT_WINDOW_SECONDS", "120")
	defer func() {
		os.Unsetenv("AUTH_RATE_LIMIT_REQUESTS")
		os.Unsetenv("AUTH_RATE_LIMIT_WINDOW_SECONDS")
	}()

	limiter := NewDistributedAuthRateLimiter(nil)
	assert.Equal(t, 20, limiter.GetLimit())
	assert.Equal(t, 120*time.Second, limiter.GetWindow())
}

func TestNewDistributedAuthRateLimiter_InvalidEnv(t *testing.T) {
	// Set invalid values (should fall back to defaults)
	os.Setenv("AUTH_RATE_LIMIT_REQUESTS", "invalid")
	os.Setenv("AUTH_RATE_LIMIT_WINDOW_SECONDS", "-1")
	defer func() {
		os.Unsetenv("AUTH_RATE_LIMIT_REQUESTS")
		os.Unsetenv("AUTH_RATE_LIMIT_WINDOW_SECONDS")
	}()

	limiter := NewDistributedAuthRateLimiter(nil)
	assert.Equal(t, 10, limiter.GetLimit())              // Default
	assert.Equal(t, 60*time.Second, limiter.GetWindow()) // Default
}

func TestHybridRateLimiter(t *testing.T) {
	// Without Redis - should use memory
	limiter := NewHybridRateLimiter(nil, time.Minute, 10, "test")
	assert.NotNil(t, limiter)
	assert.False(t, limiter.useRedis)

	// Allow should work (returns true for first requests)
	assert.True(t, limiter.Allow("test-key"))
}

func TestNewMagicLinkRateLimiter_Defaults(t *testing.T) {
	limiter := NewMagicLinkRateLimiter(nil)
	assert.NotNil(t, limiter)

	// Should allow email (no Redis)
	assert.True(t, limiter.AllowEmail("test@example.com"))

	// Should allow IP
	assert.True(t, limiter.AllowIP("192.168.1.1"))
}

func TestNewMagicLinkRateLimiter_FromEnv(t *testing.T) {
	// Set custom max attempts
	os.Setenv("MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR", "10")
	defer os.Unsetenv("MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR")

	limiter := NewMagicLinkRateLimiter(nil)
	assert.NotNil(t, limiter)

	// Email limiter should be configured with 10 attempts
	assert.Equal(t, 10, limiter.emailLimiter.redis.GetLimit())
}

func TestNewMagicLinkRateLimiter_InvalidEnv(t *testing.T) {
	// Set invalid value (should fall back to default)
	os.Setenv("MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR", "invalid")
	defer os.Unsetenv("MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR")

	limiter := NewMagicLinkRateLimiter(nil)
	assert.NotNil(t, limiter)
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test@example.com", "email:test@example.com"},
		{"TEST@EXAMPLE.COM", "email:TEST@EXAMPLE.COM"},
		{"", "email:"},
	}

	for _, tt := range tests {
		result := normalizeEmail(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestDistributedRateLimiter_AllowWithoutRedis(t *testing.T) {
	// Without Redis, should always return true (fail open)
	limiter := NewDistributedRateLimiter(nil, time.Minute, 5, "test")

	// Should allow all requests when not enabled
	for i := 0; i < 100; i++ {
		assert.True(t, limiter.Allow("test-key"))
	}
}

func TestDistributedRateLimiter_GetRemainingWithoutRedis(t *testing.T) {
	// Without Redis, should return full limit
	limiter := NewDistributedRateLimiter(nil, time.Minute, 10, "test")

	// Should return full limit when not enabled
	assert.Equal(t, 10, limiter.GetRemaining("test-key"))
}

func TestDistributedRateLimiter_ResetWithoutRedis(t *testing.T) {
	// Without Redis, should not error
	limiter := NewDistributedRateLimiter(nil, time.Minute, 10, "test")

	// Should not error when not enabled
	err := limiter.Reset("test-key")
	assert.NoError(t, err)
}
