package gateway

import (
	"context"
	"sync"
	"time"
)

// RateLimiter is the shared rate-limit surface used by GatewayCore.
// The existing per-scope rate limiters in middleware/*_rate_limit.go
// are the production implementation; this interface provides the
// GatewayCore with a protocol-agnostic check.
type RateLimiter interface {
	// Allow returns true if the request is within the rate limit.
	// The key is a composite of caller identity + target function/agent.
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// noopRateLimiter always allows. Used when rate limiting is disabled.
type noopRateLimiter struct{}

func (n *noopRateLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) (bool, error) {
	return true, nil
}

// NewNoopRateLimiter returns a rate limiter that always allows.
func NewNoopRateLimiter() RateLimiter {
	return &noopRateLimiter{}
}

// InMemoryRateLimiter is a simple token-bucket rate limiter for
// development and testing. Production uses Redis-backed limiters.
type InMemoryRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens    float64
	maxTokens float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		buckets: make(map[string]*bucket),
	}
}

// Allow checks and consumes a token from the bucket.
func (r *InMemoryRateLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.buckets[key]
	if !ok {
		b = &bucket{
			tokens:     float64(limit),
			maxTokens:  float64(limit),
			refillRate: float64(limit) / window.Seconds(),
			lastRefill: time.Now(),
		}
		r.buckets[key] = b
	}

	// Refill tokens based on elapsed time.
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true, nil
	}
	return false, nil
}

// RateLimitKey builds a composite rate-limit key from caller and target.
func RateLimitKey(callerID, target string) string {
	return callerID + ":" + target
}
