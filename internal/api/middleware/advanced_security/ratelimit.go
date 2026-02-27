package advanced_security

import (
	"math"
	"sync"
	"time"
)

// SlidingWindowRateLimiter implements sliding window rate limiting
type SlidingWindowRateLimiter struct {
	mu       sync.RWMutex
	windows  map[string][]time.Time
	window   time.Duration
	limit    int
	cleanup  time.Duration
}

// TokenBucketRateLimiter implements token bucket algorithm
type TokenBucketRateLimiter struct {
	mu          sync.RWMutex
	buckets     map[string]*TokenBucket
	rate        float64
	burst       int
	lastRefill  time.Time
}

// TokenBucket represents a token bucket
type TokenBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// SlidingWindowRateLimiter implementation
func (swrl *SlidingWindowRateLimiter) Allow(key string) bool {
	swrl.mu.Lock()
	defer swrl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-swrl.window)

	// Clean up old entries
	if timestamps, exists := swrl.windows[key]; exists {
		validTimestamps := make([]time.Time, 0)
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		swrl.windows[key] = validTimestamps
	}

	// Check limit
	if len(swrl.windows[key]) >= swrl.limit {
		return false
	}

	// Add current request
	swrl.windows[key] = append(swrl.windows[key], now)
	return true
}

// TokenBucketRateLimiter implementation
func (tbrl *TokenBucketRateLimiter) Allow(key string) bool {
	tbrl.mu.Lock()
	defer tbrl.mu.Unlock()

	now := time.Now()
	bucket, exists := tbrl.buckets[key]

	if !exists {
		bucket = &TokenBucket{
			tokens:     float64(tbrl.burst),
			lastUpdate: now,
		}
		tbrl.buckets[key] = bucket
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	tokensToAdd := elapsed * tbrl.rate

	bucket.tokens = math.Min(float64(tbrl.burst), bucket.tokens + tokensToAdd)
	bucket.lastUpdate = now

	// Check if we have enough tokens
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}

	return false
}