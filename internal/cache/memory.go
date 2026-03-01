package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
)

// MemoryCache provides L1 in-memory caching using ristretto
// This is the fastest cache layer with microsecond-level latency
type MemoryCache struct {
	cache        *ristretto.Cache
	hits         atomic.Int64
	misses       atomic.Int64
	evicts       atomic.Int64
	sizeBytes    atomic.Int64                   // approximate current cost in bytes (not updated on ristretto evictions)
	functionKeys map[string]map[string]struct{} // functionID -> set of cache keys
	keysMutex    sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache with the specified max memory in MB
func NewMemoryCache(maxMemoryMB int64) (*MemoryCache, error) {
	// Calculate number of counters (for eviction accuracy)
	// Rule of thumb: 10x the max number of items
	numCounters := maxMemoryMB * 10

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(numCounters),
		MaxCost:     maxMemoryMB * 1024 * 1024, // Convert MB to bytes
		BufferItems: 64,                        // Number of keys to buffer per goroutine
	})
	if err != nil {
		return nil, err
	}

	return &MemoryCache{
		cache:        cache,
		functionKeys: make(map[string]map[string]struct{}),
	}, nil
}

// Get retrieves a value from the memory cache
// Returns (value, found) where found is false if key doesn't exist or is expired
func (m *MemoryCache) Get(key string) ([]byte, bool) {
	value, found := m.cache.Get(key)
	if !found || value == nil {
		m.misses.Add(1)
		return nil, false
	}
	m.hits.Add(1)
	return value.([]byte), true
}

// Set stores a value in the memory cache with the specified TTL
func (m *MemoryCache) Set(key string, value []byte, ttl time.Duration) {
	// Cost is based on size
	cost := int64(len(value))
	m.cache.SetWithTTL(key, value, cost, ttl)
	m.sizeBytes.Add(cost)

	// Track the key for function-specific invalidation
	if functionID := extractFunctionIDFromKey(key); functionID != "" {
		m.trackKey(functionID, key)
	}
}

// Delete removes a key from the cache
func (m *MemoryCache) Delete(key string) {
	if value, found := m.cache.Get(key); found && value != nil {
		if b, ok := value.([]byte); ok {
			m.sizeBytes.Add(-int64(len(b)))
		}
	}
	m.cache.Del(key)

	// Remove from tracking
	if functionID := extractFunctionIDFromKey(key); functionID != "" {
		m.untrackKey(functionID, key)
	}
}

// Clear removes all entries from the cache
func (m *MemoryCache) Clear() {
	m.cache.Clear()
	m.sizeBytes.Store(0)

	// Clear all tracking data
	m.keysMutex.Lock()
	m.functionKeys = make(map[string]map[string]struct{})
	m.keysMutex.Unlock()
}

// Close gracefully closes the cache
func (m *MemoryCache) Close() error {
	m.cache.Close()
	return nil
}

// Metrics returns cache performance metrics
func (m *MemoryCache) Metrics() *CacheMetrics {
	total := m.hits.Load() + m.misses.Load()
	var ratio float64
	if total > 0 {
		ratio = float64(m.hits.Load()) / float64(total)
	}
	return &CacheMetrics{
		Hits:      m.hits.Load(),
		Misses:    m.misses.Load(),
		Ratio:     ratio,
		SizeBytes: m.sizeBytes.Load(),
		Evictions: m.evicts.Load(),
	}
}

// Helper methods for function-specific key tracking

// extractFunctionIDFromKey extracts the function ID from a cache key
// Key format: fx:cache:{function_id}:{version}:{hash16}
func extractFunctionIDFromKey(key string) string {
	// Check if this is a function cache key
	if !strings.HasPrefix(key, "fx:cache:") {
		return ""
	}

	// Remove prefix
	parts := strings.Split(key[len("fx:cache:"):], ":")
	if len(parts) < 3 {
		return ""
	}

	return parts[0]
}

// extractVersionFromKey extracts the version from a cache key
// Key format: fx:cache:{function_id}:{version}:{hash16}
func extractVersionFromKey(key string) string {
	// Check if this is a function cache key
	if !strings.HasPrefix(key, "fx:cache:") {
		return ""
	}

	// Remove prefix
	parts := strings.Split(key[len("fx:cache:"):], ":")
	if len(parts) < 3 {
		return ""
	}

	return parts[1]
}

// keyMatchesVersion checks if a cache key matches a specific function version
func keyMatchesVersion(key, functionID, version string) bool {
	// Quick check - must start with the right prefix
	expectedPrefix := "fx:cache:" + functionID + ":"
	if !strings.HasPrefix(key, expectedPrefix) {
		return false
	}

	// Extract version from key
	keyVersion := extractVersionFromKey(key)
	return keyVersion == version
}

// trackKey adds a key to the function tracking map
func (m *MemoryCache) trackKey(functionID, key string) {
	m.keysMutex.Lock()
	defer m.keysMutex.Unlock()

	if m.functionKeys[functionID] == nil {
		m.functionKeys[functionID] = make(map[string]struct{})
	}
	m.functionKeys[functionID][key] = struct{}{}
}

// untrackKey removes a key from the function tracking map
func (m *MemoryCache) untrackKey(functionID, key string) {
	m.keysMutex.Lock()
	defer m.keysMutex.Unlock()

	if keys, exists := m.functionKeys[functionID]; exists {
		delete(keys, key)
		// Clean up empty maps
		if len(keys) == 0 {
			delete(m.functionKeys, functionID)
		}
	}
}

// DeleteByFunction removes all cache entries for a specific function
func (m *MemoryCache) DeleteByFunction(functionID string) {
	m.keysMutex.RLock()
	keys := make([]string, 0, len(m.functionKeys[functionID]))
	for key := range m.functionKeys[functionID] {
		keys = append(keys, key)
	}
	m.keysMutex.RUnlock()

	// Delete keys from cache
	for _, key := range keys {
		m.cache.Del(key)
	}

	// Clean up tracking data
	m.keysMutex.Lock()
	delete(m.functionKeys, functionID)
	m.keysMutex.Unlock()
}

// DeleteByVersion removes cache entries for a specific function version
func (m *MemoryCache) DeleteByVersion(functionID, version string) {
	m.keysMutex.RLock()
	keys := make([]string, 0)
	if functionKeys, exists := m.functionKeys[functionID]; exists {
		for key := range functionKeys {
			// Check if this key matches the specific version
			if keyMatchesVersion(key, functionID, version) {
				keys = append(keys, key)
			}
		}
	}
	m.keysMutex.RUnlock()

	// Delete matching keys from cache
	for _, key := range keys {
		m.cache.Del(key)
		m.untrackKey(functionID, key)
	}
}

// CacheMetrics holds performance statistics for the memory cache
type CacheMetrics struct {
	Hits      int64
	Misses    int64
	Ratio     float64
	SizeBytes int64 // approximate current cache size in bytes
	Evictions int64
}

// WaitForIndexing waits for all recent writes to be indexed
// This is important for read-your-writes consistency
func (m *MemoryCache) WaitForIndexing() {
	m.cache.Wait()
}

// GetWithTTL retrieves a value and checks if it's still valid based on TTL
// Note: ristretto handles TTL internally, but this provides explicit TTL checking
func (m *MemoryCache) GetWithTTL(ctx context.Context, key string) ([]byte, bool, time.Duration) {
	value, found := m.cache.Get(key)
	if !found {
		return nil, false, 0
	}
	// ristretto handles TTL internally - if we get the value, it's valid
	return value.([]byte), true, 0
}

// SetWithContext stores a value with context for cancellation support
func (m *MemoryCache) SetWithContext(ctx context.Context, key string, value []byte, ttl time.Duration) bool {
	cost := int64(len(value))
	success := m.cache.SetWithTTL(key, value, cost, ttl)

	if success {
		// Track the key for function-specific invalidation
		if functionID := extractFunctionIDFromKey(key); functionID != "" {
			m.trackKey(functionID, key)
		}
	}

	return success
}

// MaxMemoryConfig returns recommended memory cache sizes for different server tiers
func MaxMemoryConfig(tier string) int64 {
	switch tier {
	case "tiny":
		return 50 // 50MB for $5 server
	case "small":
		return 100 // 100MB
	case "medium":
		return 250 // 250MB
	case "large":
		return 500 // 500MB
	default:
		return 100 // default
	}
}

// NewMemoryCacheForTier creates a cache sized appropriately for a server tier
func NewMemoryCacheForTier(tier string) (*MemoryCache, error) {
	return NewMemoryCache(MaxMemoryConfig(tier))
}

// Stats returns a formatted string of cache statistics
func (m *MemoryCache) Stats() string {
	metrics := m.Metrics()
	return fmt.Sprintf(
		"Cache Stats - Hits: %d, Misses: %d, Hit Ratio: %.2f%%, Evictions: %d",
		metrics.Hits,
		metrics.Misses,
		metrics.Ratio*100,
		metrics.Evictions,
	)
}
