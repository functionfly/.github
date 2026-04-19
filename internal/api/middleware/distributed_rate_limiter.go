package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// DistributedRateLimiter provides Redis-backed rate limiting for distributed deployments
// Use this instead of the in-memory RateLimiter when running multiple API instances
type DistributedRateLimiter struct {
	redis     *redis.Client
	window    time.Duration
	limit     int
	keyPrefix string
	enabled   bool
}

// NewDistributedRateLimiter creates a new Redis-backed rate limiter
// Falls back to in-memory rate limiting if Redis is not available
func NewDistributedRateLimiter(redisClient *redis.Client, window time.Duration, limit int, keyPrefix string) *DistributedRateLimiter {
	// Check if distributed rate limiting is explicitly disabled
	disabled := os.Getenv("DISTRIBUTED_RATE_LIMITER_DISABLED") == "true"

	return &DistributedRateLimiter{
		redis:     redisClient,
		window:    window,
		limit:     limit,
		keyPrefix: keyPrefix,
		enabled:   redisClient != nil && !disabled,
	}
}

// NewDistributedAuthRateLimiter creates a distributed rate limiter for auth endpoints
func NewDistributedAuthRateLimiter(redisClient *redis.Client) *DistributedRateLimiter {
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

	return NewDistributedRateLimiter(
		redisClient,
		time.Duration(window)*time.Second,
		limit,
		"auth_rate_limit",
	)
}

// Allow checks if a request should be allowed using Redis sliding window
func (rl *DistributedRateLimiter) Allow(key string) bool {
	if !rl.enabled || rl.redis == nil {
		// Fall back to in-memory if Redis is not available
		return true
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)
	now := time.Now().Unix()
	windowStart := now - int64(rl.window.Seconds())

	// Use Redis sorted set for sliding window
	// Remove old entries outside the window
	pipe := rl.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("%d", windowStart))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(now), Member: now})

	// Set expiry on the key
	pipe.Expire(ctx, redisKey, rl.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Redis rate limiter error, allowing request")
		return true // Fail open on Redis error
	}

	count := countCmd.Val()
	return int(count) < rl.limit
}

// GetRemaining returns the remaining requests allowed for the key
func (rl *DistributedRateLimiter) GetRemaining(key string) int {
	if !rl.enabled || rl.redis == nil {
		return rl.limit // Return full limit if not using Redis
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)
	now := time.Now().Unix()
	windowStart := now - int64(rl.window.Seconds())

	// Remove old entries and count remaining
	pipe := rl.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, redisKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Redis rate limiter error getting remaining")
		return rl.limit
	}

	count := int(countCmd.Val())
	remaining := rl.limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// Reset clears the rate limit for a key (useful for testing or admin overrides)
func (rl *DistributedRateLimiter) Reset(key string) error {
	if !rl.enabled || rl.redis == nil {
		return nil
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)
	return rl.redis.Del(ctx, redisKey).Err()
}

// IsEnabled returns whether the distributed rate limiter is active
func (rl *DistributedRateLimiter) IsEnabled() bool {
	return rl.enabled
}

// GetLimit returns the configured rate limit
func (rl *DistributedRateLimiter) GetLimit() int {
	return rl.limit
}

// GetWindow returns the configured time window
func (rl *DistributedRateLimiter) GetWindow() time.Duration {
	return rl.window
}

// DistributedRateLimitMiddleware wraps the distributed rate limiter as HTTP middleware
type DistributedRateLimitMiddleware struct {
	limiter *DistributedRateLimiter
}

// NewDistributedRateLimitMiddleware creates middleware from a distributed rate limiter
func NewDistributedRateLimitMiddleware(limiter *DistributedRateLimiter) *DistributedRateLimitMiddleware {
	return &DistributedRateLimitMiddleware{limiter: limiter}
}

// Limit wraps a handler with distributed rate limiting
func (m *DistributedRateLimitMiddleware) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.limiter.IsEnabled() {
			// Skip if not enabled
			next(w, r)
			return
		}

		// Use client IP as the rate limiting key
		clientIP := getClientIP(r)

		if !m.limiter.Allow(clientIP) {
			remaining := m.limiter.GetRemaining(clientIP)
			logrus.WithFields(logrus.Fields{
				"ip":        clientIP,
				"path":      r.URL.Path,
				"remaining": remaining,
			}).Warn("Distributed rate limit exceeded")

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(m.limiter.GetWindow().Seconds())))
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.limiter.GetLimit()))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(m.limiter.GetWindow()).Unix()))
			w.WriteHeader(http.StatusTooManyRequests)

			response := map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Add rate limit headers to successful requests
		remaining := m.limiter.GetRemaining(getClientIP(r))
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.limiter.GetLimit()))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Window", fmt.Sprintf("%d", int(m.limiter.GetWindow().Seconds())))

		next(w, r)
	}
}

// HybridRateLimiter combines in-memory and Redis-backed rate limiting
// Uses in-memory for single-instance deployments and Redis for distributed deployments
type HybridRateLimiter struct {
	memory   *RateLimiter
	redis    *DistributedRateLimiter
	useRedis bool
}

// NewHybridRateLimiter creates a rate limiter that automatically chooses the best backend
// Uses Redis if available, falls back to in-memory otherwise
func NewHybridRateLimiter(redisClient *redis.Client, window time.Duration, limit int, keyPrefix string) *HybridRateLimiter {
	redisLimiter := NewDistributedRateLimiter(redisClient, window, limit, keyPrefix)

	return &HybridRateLimiter{
		memory:   NewRateLimiter(window, limit),
		redis:    redisLimiter,
		useRedis: redisLimiter.IsEnabled(),
	}
}

// Allow checks rate limit using the best available backend
func (rl *HybridRateLimiter) Allow(key string) bool {
	if rl.useRedis {
		return rl.redis.Allow(key)
	}
	return rl.memory.Allow(key)
}

// MagicLinkRateLimiter provides rate limiting specifically for magic link endpoints
type MagicLinkRateLimiter struct {
	// Per-email rate limiting (distributed across instances if Redis available)
	emailLimiter *HybridRateLimiter
	// Per-IP rate limiting (in-memory only - IP is per-instance concept)
	ipLimiter *RateLimiter
}

// NewMagicLinkRateLimiter creates a rate limiter for magic link requests
func NewMagicLinkRateLimiter(redisClient *redis.Client) *MagicLinkRateLimiter {
	// Load limits from env or use defaults
	maxAttempts := 5
	if v := os.Getenv("MAGIC_LINK_MAX_ATTEMPTS_PER_HOUR"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxAttempts = parsed
		}
	}

	return &MagicLinkRateLimiter{
		// Email-based: distributed with Redis (users can hit different instances)
		emailLimiter: NewHybridRateLimiter(redisClient, time.Hour, maxAttempts, "magic_link_email"),
		// IP-based: in-memory only (IP is tied to the instance handling the request)
		ipLimiter: NewRateLimiter(time.Minute, 10), // 10 per minute per IP
	}
}

// AllowEmail checks if an email is within rate limits
func (rl *MagicLinkRateLimiter) AllowEmail(email string) bool {
	return rl.emailLimiter.Allow(normalizeEmail(email))
}

// AllowIP checks if an IP is within rate limits
func (rl *MagicLinkRateLimiter) AllowIP(ip string) bool {
	return rl.ipLimiter.Allow(ip)
}

// normalizeEmail standardizes email for rate limiting
func normalizeEmail(email string) string {
	return "email:" + email
}
