package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// TenantRateLimiter provides per-tenant per-second rate limiting using Redis.
// Limits are plan-aware: Starter=100 req/s, Professional=1000 req/s, Enterprise=10000 req/s.
type TenantRateLimiter struct {
	redis   *redis.Client
	enabled bool
	limits  map[string]int // plan → requests per second
}

// NewTenantRateLimiter creates a new per-tenant rate limiter.
func NewTenantRateLimiter(redisClient *redis.Client) *TenantRateLimiter {
	enabled := redisClient != nil && os.Getenv("TENANT_RATE_LIMITER_DISABLED") != "true"

	limits := map[string]int{
		"free":         10,
		"starter":      100,
		"professional": 1000,
		"enterprise":   10000,
		"agent":        10000,
	}

	// Allow override via env
	if v := os.Getenv("TENANT_RATE_LIMIT_STARTER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits["starter"] = n
		}
	}
	if v := os.Getenv("TENANT_RATE_LIMIT_PROFESSIONAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits["professional"] = n
		}
	}
	if v := os.Getenv("TENANT_RATE_LIMIT_ENTERPRISE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits["enterprise"] = n
		}
	}

	return &TenantRateLimiter{
		redis:   redisClient,
		enabled: enabled,
		limits:  limits,
	}
}

// Allow checks if a request from the given tenant should be allowed.
// Returns the limit and remaining quota for response headers.
func (rl *TenantRateLimiter) Allow(tenantID, plan string) (allowed bool, limit int, remaining int) {
	if !rl.enabled || rl.redis == nil {
		return true, 0, 0
	}

	limit = rl.limitForPlan(plan)
	if limit <= 0 {
		return true, 0, 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	key := fmt.Sprintf("ratelimit:tenant:%s", tenantID)
	now := time.Now().UnixMilli()
	windowStart := now - 1000 // 1-second sliding window

	pipe := rl.redis.Pipeline()
	// Remove entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// Count entries in the window
	countCmd := pipe.ZCard(ctx, key)
	// Add the current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	// Expire the key after 2 seconds (cleanup)
	pipe.Expire(ctx, key, 2*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Tenant rate limiter Redis error")
		// Fail open: allow the request if Redis is unavailable
		return true, limit, limit
	}

	currentCount := int(countCmd.Val())
	remaining = limit - currentCount - 1
	if remaining < 0 {
		remaining = 0
	}

	return currentCount < limit, limit, remaining
}

func (rl *TenantRateLimiter) limitForPlan(plan string) int {
	if limit, ok := rl.limits[plan]; ok {
		return limit
	}
	return rl.limits["free"] // default to free tier
}

// TenantRateLimitMiddleware returns HTTP middleware that rate-limits by tenant ID.
// The tenant ID must be set in the request context by upstream auth middleware.
// For edge-routed requests (public routes), the tenant is resolved in handlePublicRoute,
// so this middleware is applied selectively.
func (rl *TenantRateLimiter) TenantRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant ID from header (set by auth middleware or edge proxy)
		tenantID := r.Header.Get("X-Tenant-ID")
		plan := r.Header.Get("X-Tenant-Plan")

		if tenantID == "" {
			next.ServeHTTP(w, r)
			return
		}

		allowed, limit, remaining := rl.Allow(tenantID, plan)

		// Always set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"rate_limit_exceeded","code":"TENANT_RATE_LIMIT","message":"Rate limit exceeded for tenant. Limit: %d req/s","retry_after":1}`, limit)
			return
		}

		next.ServeHTTP(w, r)
	})
}
