package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MultiLayerCacheLayer represents a cache layer
type MultiLayerCacheLayer int

const (
	// MultiLayerLayerL1 is in-memory cache (fastest)
	MultiLayerLayerL1 MultiLayerCacheLayer = iota
	// MultiLayerLayerL2 is Redis cache (distributed)
	MultiLayerLayerL2
	// MultiLayerLayerL3 is disk cache (persistent)
	MultiLayerLayerL3
)

// String returns the string representation of cache layer
func (l MultiLayerCacheLayer) String() string {
	switch l {
	case MultiLayerLayerL1:
		return "l1_memory"
	case MultiLayerLayerL2:
		return "l2_redis"
	case MultiLayerLayerL3:
		return "l3_disk"
	default:
		return "unknown"
	}
}

// MultiLayerCacheEntry represents a cached item
type MultiLayerCacheEntry struct {
	Key        string
	Value      interface{}
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AccessedAt time.Time
	HitCount   int64
	Size       int64
}

// MultiLayerCacheStats tracks cache statistics
type MultiLayerCacheStats struct {
	Hits      int64
	Misses    int64
	Sets      int64
	Deletes   int64
	Evictions int64
	Size      int64
	mu        sync.RWMutex
}

// HitRate returns the cache hit rate
func (s *MultiLayerCacheStats) HitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// MultiLayerCache implements multi-layer caching
type MultiLayerCache struct {
	l1Cache      map[string]*MultiLayerCacheEntry
	l2Cache      MultiLayerRedisCache
	l3Cache      MultiLayerDiskCache
	l1Mutex      sync.RWMutex
	stats        map[MultiLayerCacheLayer]*MultiLayerCacheStats
	statsMutex   sync.RWMutex
	defaultTTL   time.Duration
	maxL1Size    int64
	currentL1Size int64
}

// MultiLayerRedisCache interface for L2 cache
type MultiLayerRedisCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// MultiLayerDiskCache interface for L3 cache
type MultiLayerDiskCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// NewMultiLayerCache creates a new multi-layer cache
func NewMultiLayerCache(l2Cache MultiLayerRedisCache, l3Cache MultiLayerDiskCache, defaultTTL time.Duration, maxL1Size int64) *MultiLayerCache {
	return &MultiLayerCache{
		l1Cache:    make(map[string]*MultiLayerCacheEntry),
		l2Cache:    l2Cache,
		l3Cache:    l3Cache,
		stats:      make(map[MultiLayerCacheLayer]*MultiLayerCacheStats),
		defaultTTL: defaultTTL,
		maxL1Size:  maxL1Size,
	}
}

// Get retrieves a value from cache, checking all layers
func (c *MultiLayerCache) Get(ctx context.Context, key string) (interface{}, error) {
	// Try L1 (in-memory)
	if value, found := c.getL1(key); found {
		c.recordHit(MultiLayerLayerL1)
		return value, nil
	}

	// Try L2 (Redis)
	if c.l2Cache != nil {
		data, err := c.l2Cache.Get(ctx, key)
		if err == nil {
			c.recordHit(MultiLayerLayerL2)
			// Promote to L1
			c.setL1(key, data, c.defaultTTL)
			return data, nil
		}
	}

	// Try L3 (Disk)
	if c.l3Cache != nil {
		data, err := c.l3Cache.Get(ctx, key)
		if err == nil {
			c.recordHit(MultiLayerLayerL3)
			// Promote to L1 and L2
			c.setL1(key, data, c.defaultTTL)
			if c.l2Cache != nil {
				_ = c.l2Cache.Set(ctx, key, data, c.defaultTTL)
			}
			return data, nil
		}
	}

	c.recordMiss()
	return nil, fmt.Errorf("cache miss for key: %s", key)
}

// Set stores a value in all cache layers
func (c *MultiLayerCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Set in L1
	c.setL1(key, value, ttl)

	// Set in L2
	if c.l2Cache != nil {
		data, ok := value.([]byte)
		if !ok {
			data = []byte(fmt.Sprintf("%v", value))
		}
		if err := c.l2Cache.Set(ctx, key, data, ttl); err != nil {
			logrus.WithError(err).Warn("Failed to set L2 cache")
		}
	}

	// Set in L3
	if c.l3Cache != nil {
		data, ok := value.([]byte)
		if !ok {
			data = []byte(fmt.Sprintf("%v", value))
		}
		if err := c.l3Cache.Set(ctx, key, data, ttl); err != nil {
			logrus.WithError(err).Warn("Failed to set L3 cache")
		}
	}

	c.recordSet()
	return nil
}

// Delete removes a value from all cache layers
func (c *MultiLayerCache) Delete(ctx context.Context, key string) error {
	// Delete from L1
	c.deleteL1(key)

	// Delete from L2
	if c.l2Cache != nil {
		if err := c.l2Cache.Delete(ctx, key); err != nil {
			logrus.WithError(err).Warn("Failed to delete from L2 cache")
		}
	}

	// Delete from L3
	if c.l3Cache != nil {
		if err := c.l3Cache.Delete(ctx, key); err != nil {
			logrus.WithError(err).Warn("Failed to delete from L3 cache")
		}
	}

	c.recordDelete()
	return nil
}

// getL1 retrieves from L1 cache
func (c *MultiLayerCache) getL1(key string) (interface{}, bool) {
	c.l1Mutex.RLock()
	defer c.l1Mutex.RUnlock()

	entry, exists := c.l1Cache[key]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	// Update access time and hit count
	entry.AccessedAt = time.Now()
	entry.HitCount++

	return entry.Value, true
}

// setL1 stores in L1 cache
func (c *MultiLayerCache) setL1(key string, value interface{}, ttl time.Duration) {
	c.l1Mutex.Lock()
	defer c.l1Mutex.Unlock()

	// Check if we need to evict
	if c.currentL1Size >= c.maxL1Size {
		c.evictL1()
	}

	now := time.Now()
	entry := &MultiLayerCacheEntry{
		Key:        key,
		Value:      value,
		ExpiresAt:  now.Add(ttl),
		CreatedAt:  now,
		AccessedAt: now,
		HitCount:   0,
		Size:       int64(len(fmt.Sprintf("%v", value))),
	}

	c.l1Cache[key] = entry
	c.currentL1Size += entry.Size
}

// deleteL1 removes from L1 cache
func (c *MultiLayerCache) deleteL1(key string) {
	c.l1Mutex.Lock()
	defer c.l1Mutex.Unlock()

	if entry, exists := c.l1Cache[key]; exists {
		c.currentL1Size -= entry.Size
		delete(c.l1Cache, key)
	}
}

// evictL1 evicts least recently used entries from L1
func (c *MultiLayerCache) evictL1() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.l1Cache {
		if oldestKey == "" || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}

	if oldestKey != "" {
		if entry, exists := c.l1Cache[oldestKey]; exists {
			c.currentL1Size -= entry.Size
			delete(c.l1Cache, oldestKey)
			c.recordEviction(MultiLayerLayerL1)
		}
	}
}

// recordHit records a cache hit
func (c *MultiLayerCache) recordHit(layer MultiLayerCacheLayer) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	if c.stats[layer] == nil {
		c.stats[layer] = &MultiLayerCacheStats{}
	}
	c.stats[layer].Hits++
}

// recordMiss records a cache miss
func (c *MultiLayerCache) recordMiss() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	for layer := range c.stats {
		c.stats[layer].Misses++
	}
}

// recordSet records a cache set
func (c *MultiLayerCache) recordSet() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	for layer := range c.stats {
		c.stats[layer].Sets++
	}
}

// recordDelete records a cache delete
func (c *MultiLayerCache) recordDelete() {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	for layer := range c.stats {
		c.stats[layer].Deletes++
	}
}

// recordEviction records a cache eviction
func (c *MultiLayerCache) recordEviction(layer MultiLayerCacheLayer) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	if c.stats[layer] == nil {
		c.stats[layer] = &MultiLayerCacheStats{}
	}
	c.stats[layer].Evictions++
}

// GetStats returns cache statistics for all layers
func (c *MultiLayerCache) GetStats() map[MultiLayerCacheLayer]*MultiLayerCacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	stats := make(map[MultiLayerCacheLayer]*MultiLayerCacheStats)
	for layer, s := range c.stats {
		stats[layer] = &MultiLayerCacheStats{
			Hits:      s.Hits,
			Misses:    s.Misses,
			Sets:      s.Sets,
			Deletes:   s.Deletes,
			Evictions: s.Evictions,
			Size:      s.Size,
		}
	}
	return stats
}

// Clear clears all cache layers
func (c *MultiLayerCache) Clear(ctx context.Context) error {
	c.l1Mutex.Lock()
	c.l1Cache = make(map[string]*MultiLayerCacheEntry)
	c.currentL1Size = 0
	c.l1Mutex.Unlock()

	logrus.Info("All cache layers cleared")
	return nil
}
