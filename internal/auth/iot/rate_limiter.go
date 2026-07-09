package iot

import (
	"sync"
	"time"
)

type RateLimiter struct {
	max    int
	window time.Duration
	mu     sync.Mutex
	buckets map[string]*tokenBucket
}

type tokenBucket struct {
	tokens    int
	lastCheck time.Time
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		max:     max,
		window:  window,
		buckets: make(map[string]*tokenBucket),
	}
}

func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[identifier]

	if !exists {
		rl.buckets[identifier] = &tokenBucket{
			tokens:    rl.max - 1,
			lastCheck: now,
		}
		return true
	}

	elapsed := now.Sub(bucket.lastCheck)
	tokensToAdd := int(elapsed.Seconds() * float64(rl.max) / rl.window.Seconds())

	if tokensToAdd > 0 {
		bucket.tokens = min(rl.max, bucket.tokens+tokensToAdd)
		bucket.lastCheck = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) Reset(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.buckets, identifier)
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for identifier, bucket := range rl.buckets {
		if now.Sub(bucket.lastCheck) > rl.window*2 {
			delete(rl.buckets, identifier)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
