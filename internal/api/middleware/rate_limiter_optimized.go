package middleware

import (
	"sync"
	"time"
)

// BucketRateLimiter implements an optimized in-memory rate limiter using time buckets.
// This avoids O(n*m) cleanup by using fixed-size time buckets that can be deleted in O(1).
// WARNING: This implementation is NOT safe for distributed deployments with multiple instances.
type BucketRateLimiter struct {
	mu          sync.Mutex
	buckets     map[int64]map[string]int // bucketTime -> key -> count
	window      time.Duration
	limit       int
	bucketSize  time.Duration
}

// NewBucketRateLimiter creates a new bucket-based rate limiter
// bucketSize is the size of each time bucket (e.g., 1 second)
func NewBucketRateLimiter(window time.Duration, limit int, bucketSize time.Duration) *BucketRateLimiter {
	if bucketSize <= 0 {
		bucketSize = time.Second
	}
	rl := &BucketRateLimiter{
		buckets:    make(map[int64]map[string]int),
		window:     window,
		limit:      limit,
		bucketSize: bucketSize,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *BucketRateLimiter) cleanupLoop() {
	// Cleanup more frequently than the window to ensure no gaps
	// Use 1/10th of bucket size or 1 second, whichever is larger
	interval := rl.bucketSize / 10
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		rl.cleanup()
	}
}

func (rl *BucketRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	oldestValidBucket := now.Add(-rl.window).UnixNano() / rl.bucketSize.Nanoseconds()

	for bucketTime := range rl.buckets {
		if bucketTime < oldestValidBucket {
			delete(rl.buckets, bucketTime)
		}
	}
}

// Allow checks if a request from the given key should be allowed
func (rl *BucketRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	currentBucket := now.UnixNano() / rl.bucketSize.Nanoseconds()
	windowStart := now.Add(-rl.window).UnixNano() / rl.bucketSize.Nanoseconds()

	// Initialize current bucket if needed
	if rl.buckets[currentBucket] == nil {
		rl.buckets[currentBucket] = make(map[string]int)
	}

	// Count requests across all valid buckets
	totalRequests := 0
	for bucketTime := range rl.buckets {
		if bucketTime >= windowStart {
			if count, exists := rl.buckets[bucketTime][key]; exists {
				totalRequests += count
			}
		}
	}

	// Check if under limit
	if totalRequests < rl.limit {
		rl.buckets[currentBucket][key]++
		return true
	}

	return false
}

// GetRemaining returns the number of remaining requests for a key
func (rl *BucketRateLimiter) GetRemaining(key string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window).UnixNano() / rl.bucketSize.Nanoseconds()

	totalRequests := 0
	for bucketTime := range rl.buckets {
		if bucketTime >= windowStart {
			if count, exists := rl.buckets[bucketTime][key]; exists {
				totalRequests += count
			}
		}
	}

	remaining := rl.limit - totalRequests
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Stats returns current rate limiter statistics
func (rl *BucketRateLimiter) Stats() (bucketCount, keyCount, totalRequests int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucketCount = len(rl.buckets)
	for _, keys := range rl.buckets {
		keyCount += len(keys)
		for _, count := range keys {
			totalRequests += count
		}
	}
	return
}
