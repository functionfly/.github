package middleware

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Stress test configuration
const (
	stressTestClients    = 10000  // Number of concurrent clients (keys)
	stressTestRequests  = 100    // Requests per client
	stressTestWindow    = time.Minute
	stressTestLimit     = 100
)

// TestRateLimiterCleanupStress tests cleanup performance under high load
func TestRateLimiterCleanupStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	rl := NewRateLimiter(stressTestWindow, stressTestLimit)

	// Simulate thousands of concurrent clients making requests
	var wg sync.WaitGroup
	clientCount := stressTestClients
	requestsPerClient := stressTestRequests

	t.Logf("Populating rate limiter with %d clients x %d requests = %d total entries",
		clientCount, requestsPerClient, clientCount*requestsPerClient)

	start := time.Now()
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
	populateDuration := time.Since(start)

	t.Logf("Population took: %v", populateDuration)
	t.Logf("Map size after population: %d keys", len(rl.requests))

	// Verify map size
	if len(rl.requests) != clientCount {
		t.Errorf("Expected %d keys, got %d", clientCount, len(rl.requests))
	}

	// Measure cleanup performance
	t.Log("Starting cleanup benchmark...")

	// Force cleanup by advancing time (we can't easily do this without modifying the code)
	// Instead, we'll just measure the cleanup as-is
	cleanupStart := time.Now()
	rl.cleanup()
	cleanupDuration := time.Since(cleanupStart)

	t.Logf("Cleanup duration: %v", cleanupDuration)
	t.Logf("Keys remaining after cleanup: %d", len(rl.requests))

	// Cleanup should complete in reasonable time
	// Under normal conditions, 10k keys with 100 timestamps each should clean up quickly
	if cleanupDuration > 5*time.Second {
		t.Errorf("Cleanup took too long: %v (expected < 5s)", cleanupDuration)
	}
}

// BenchmarkRateLimiterCleanup benchmarks cleanup performance
func BenchmarkRateLimiterCleanup(b *testing.B) {
	// Vary the number of clients to see scaling
	clientCounts := []int{1000, 5000, 10000, 25000}

	for _, clientCount := range clientCounts {
		b.Run(fmt.Sprintf("%dClients", clientCount), func(b *testing.B) {
			rl := NewRateLimiter(stressTestWindow, stressTestLimit)

			// Populate with requests
			var wg sync.WaitGroup
			requestsPerClient := stressTestRequests

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

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl.cleanup()
			}
			b.StopTimer()

			b.ReportMetric(float64(len(rl.requests)), "keys")
			b.ReportMetric(float64(clientCount*requestsPerClient), "total_timestamps")
			b.ReportAllocs()
		})
	}
}

// BenchmarkRateLimiterConcurrentAccess benchmarks concurrent access during cleanup
func BenchmarkRateLimiterConcurrentAccess(b *testing.B) {
	clientCount := 5000
	requestsPerClient := 50

	rl := NewRateLimiter(stressTestWindow, stressTestLimit)

	// Populate first
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

	b.ReportMetric(float64(len(rl.requests)), "keys")
}

// TestRateLimiterCleanupWithExpiredEntries tests cleanup when most entries are expired
func TestRateLimiterCleanupWithExpiredEntries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a rate limiter with very short window for testing
	shortWindow := 100 * time.Millisecond
	rl := NewRateLimiter(shortWindow, stressTestLimit)

	clientCount := 5000
	requestsPerClient := 50

	// Populate
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

	initialKeys := len(rl.requests)
	t.Logf("Initial keys: %d", initialKeys)

	// Wait for all entries to expire
	time.Sleep(shortWindow + 50*time.Millisecond)

	// Now cleanup should remove all entries
	rl.cleanup()

	finalKeys := len(rl.requests)
	t.Logf("Final keys after expired cleanup: %d", finalKeys)

	if finalKeys != 0 {
		t.Errorf("Expected 0 keys after expired cleanup, got %d", finalKeys)
	}
}

// TestRateLimiterMemoryUsage tests memory usage with large number of entries
func TestRateLimiterMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	rl := NewRateLimiter(stressTestWindow, stressTestLimit)

	clientCount := 10000
	requestsPerClient := 100

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

	t.Logf("Map size: %d keys", len(rl.requests))
	t.Logf("Total timestamps stored: approximately %d", clientCount*requestsPerClient)

	// Verify all entries were stored
	if len(rl.requests) != clientCount {
		t.Errorf("Expected %d keys, got %d", clientCount, len(rl.requests))
	}
}

// TestRateLimiterCleanupWithNow tests cleanup using actual time.Now() timestamps
func TestRateLimiterCleanupWithNow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Use a real 1-minute window with actual time.Now()
	window := time.Minute
	rl := NewRateLimiter(window, stressTestLimit)

	clientCount := 10000
	requestsPerClient := 50

	t.Logf("Testing with actual time.Now() - %d clients x %d requests", clientCount, requestsPerClient)
	t.Logf("Test started at: %v", time.Now())

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

	t.Logf("Population completed at: %v", time.Now())
	t.Logf("Map size: %d keys", len(rl.requests))

	// Measure cleanup with actual time.Now()
	cleanupStart := time.Now()
	rl.cleanup()
	cleanupDuration := time.Since(cleanupStart)

	t.Logf("Cleanup completed at: %v", time.Now())
	t.Logf("Cleanup duration (wall clock): %v", cleanupDuration)

	// Verify cleanup was fast
	if cleanupDuration > 10*time.Second {
		t.Errorf("Cleanup took too long: %v", cleanupDuration)
	}

	// Verify all entries still present (they haven't expired yet)
	if len(rl.requests) != clientCount {
		t.Errorf("Expected %d keys after cleanup, got %d", clientCount, len(rl.requests))
	}
}
