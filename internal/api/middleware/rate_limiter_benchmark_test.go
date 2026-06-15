package middleware

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// BenchmarkOriginalRateLimiterCleanup benchmarks the original O(n*m) cleanup
func BenchmarkOriginalRateLimiterCleanup(b *testing.B) {
	clientCounts := []int{1000, 5000, 10000, 25000}

	for _, clientCount := range clientCounts {
		b.Run(fmt.Sprintf("%dClients", clientCount), func(b *testing.B) {
			rl := NewRateLimiter(time.Minute, 100)

			// Populate
			var wg sync.WaitGroup
			for i := 0; i < clientCount; i++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					key := fmt.Sprintf("client-%d", clientID)
					for j := 0; j < 50; j++ {
						rl.Allow(key)
					}
				}(i)
			}
			wg.Wait()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl.cleanup()
			}
			b.StopTimer()

			b.ReportMetric(float64(len(rl.requests)), "keys")
			b.ReportAllocs()
		})
	}
}

// BenchmarkBucketRateLimiterCleanup benchmarks the bucket-based O(1) cleanup
func BenchmarkBucketRateLimiterCleanup(b *testing.B) {
	clientCounts := []int{1000, 5000, 10000, 25000}

	for _, clientCount := range clientCounts {
		b.Run(fmt.Sprintf("%dClients", clientCount), func(b *testing.B) {
			rl := NewBucketRateLimiter(time.Minute, 100, time.Second)

			// Populate
			var wg sync.WaitGroup
			for i := 0; i < clientCount; i++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					key := fmt.Sprintf("client-%d", clientID)
					for j := 0; j < 50; j++ {
						rl.Allow(key)
					}
				}(i)
			}
			wg.Wait()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl.cleanup()
			}
			b.StopTimer()

			bucketCount, keyCount, _ := rl.Stats()
			b.ReportMetric(float64(bucketCount), "buckets")
			b.ReportMetric(float64(keyCount), "keys")
			b.ReportAllocs()
		})
	}
}

// BenchmarkBucketRateLimiterVsOriginal_ConcurrentAccess compares concurrent access performance
func BenchmarkBucketRateLimiterVsOriginal_ConcurrentAccess(b *testing.B) {
	clientCount := 5000

	b.Run("Original", func(b *testing.B) {
		rl := NewRateLimiter(time.Minute, 100)

		// Populate
		var wg sync.WaitGroup
		for i := 0; i < clientCount; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()
				key := fmt.Sprintf("client-%d", clientID)
				for j := 0; j < 50; j++ {
					rl.Allow(key)
				}
			}(i)
		}
		wg.Wait()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			clientID := 0
			for pb.Next() {
				key := fmt.Sprintf("client-%d", clientID)
				rl.Allow(key)
				clientID = (clientID + 1) % clientCount
			}
		})
		b.StopTimer()
	})

	b.Run("Bucket", func(b *testing.B) {
		rl := NewBucketRateLimiter(time.Minute, 100, time.Second)

		// Populate
		var wg sync.WaitGroup
		for i := 0; i < clientCount; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()
				key := fmt.Sprintf("client-%d", clientID)
				for j := 0; j < 50; j++ {
					rl.Allow(key)
				}
			}(i)
		}
		wg.Wait()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			clientID := 0
			for pb.Next() {
				key := fmt.Sprintf("client-%d", clientID)
				rl.Allow(key)
				clientID = (clientID + 1) % clientCount
			}
		})
		b.StopTimer()
	})
}

// TestBucketRateLimiterCleanup verifies the bucket implementation correctness
func TestBucketRateLimiterCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	rl := NewBucketRateLimiter(100*time.Millisecond, 100, 10*time.Millisecond)

	clientCount := 1000
	requestsPerClient := 50

	var wg sync.WaitGroup
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			key := fmt.Sprintf("client-%d", clientID)
			for j := 0; j < requestsPerClient; j++ {
				rl.Allow(key)
			}
		}(i)
	}
	wg.Wait()

	bucketCount, keyCount, totalRequests := rl.Stats()
	t.Logf("Initial state: %d buckets, %d keys, %d total requests", bucketCount, keyCount, totalRequests)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Cleanup
	rl.cleanup()

	bucketCount, keyCount, totalRequests = rl.Stats()
	t.Logf("After cleanup: %d buckets, %d keys, %d total requests", bucketCount, keyCount, totalRequests)

	if totalRequests != 0 {
		t.Errorf("Expected 0 total requests after expiration, got %d", totalRequests)
	}
}
