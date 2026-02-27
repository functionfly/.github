package cache

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WriteThroughCache provides write-through caching strategy
type WriteThroughCache struct {
	redisCache     *RegistryRedisCache
	cacheKeyGen    *RegistryCacheKey
	db             *gorm.DB
	warmupEnabled  bool
	warmupInterval time.Duration
	warmupTicker   *time.Ticker
	stopWarmup     chan bool
	warmupMutex    sync.RWMutex
}

// CacheWriteOperation represents a cache write operation
type CacheWriteOperation struct {
	Key      string
	Data     interface{}
	TTL      time.Duration
	Priority int // Higher priority = process first
}

// NewWriteThroughCache creates a new write-through cache
func NewWriteThroughCache(redisCache *RegistryRedisCache, db *gorm.DB) *WriteThroughCache {
	return &WriteThroughCache{
		redisCache:     redisCache,
		cacheKeyGen:    NewRegistryCacheKey(),
		db:             db,
		warmupEnabled:  true,
		warmupInterval: 30 * time.Minute, // Warmup cache every 30 minutes
		stopWarmup:     make(chan bool),
	}
}

// EnableCacheWarmup enables periodic cache warming
func (wtc *WriteThroughCache) EnableCacheWarmup(interval time.Duration) {
	wtc.warmupMutex.Lock()
	defer wtc.warmupMutex.Unlock()

	wtc.warmupEnabled = true
	wtc.warmupInterval = interval

	if wtc.warmupTicker != nil {
		wtc.warmupTicker.Stop()
	}

	wtc.warmupTicker = time.NewTicker(interval)
	go wtc.warmupRoutine()
}

// DisableCacheWarmup disables periodic cache warming
func (wtc *WriteThroughCache) DisableCacheWarmup() {
	wtc.warmupMutex.Lock()
	defer wtc.warmupMutex.Unlock()

	wtc.warmupEnabled = false
	if wtc.warmupTicker != nil {
		wtc.warmupTicker.Stop()
		wtc.warmupTicker = nil
	}
	wtc.stopWarmup <- true
}

// WriteThrough executes a database operation and updates cache
func (wtc *WriteThroughCache) WriteThrough(ctx context.Context, operation func(*gorm.DB) error, cacheOps []CacheWriteOperation) error {
	// Execute database operation first
	if err := operation(wtc.db); err != nil {
		return err
	}

	// Update cache (write-through strategy)
	return wtc.updateCache(ctx, cacheOps)
}

// ReadThrough checks cache first, then database if cache miss
func (wtc *WriteThroughCache) ReadThrough(ctx context.Context, cacheKey string, dest interface{}, dbQuery func(*gorm.DB) error) error {
	// Try cache first
	if err := wtc.redisCache.GetJSON(ctx, cacheKey, dest); err == nil {
		return nil // Cache hit
	}

	// Cache miss - query database
	if err := dbQuery(wtc.db); err != nil {
		return err
	}

	// Update cache asynchronously
	go func() {
		if err := wtc.redisCache.SetJSON(context.Background(), cacheKey, dest); err != nil {
			logrus.WithError(err).WithField("key", cacheKey).Warn("Failed to update cache after read-through")
		}
	}()

	return nil
}

// InvalidateCache invalidates cache entries related to a database change
func (wtc *WriteThroughCache) InvalidateCache(ctx context.Context, pattern string, relatedKeys []string) error {
	// Invalidate the main key
	if err := wtc.redisCache.Delete(ctx, pattern); err != nil {
		logrus.WithError(err).WithField("key", pattern).Warn("Failed to invalidate cache key")
	}

	// Invalidate related keys
	for _, key := range relatedKeys {
		if err := wtc.redisCache.Delete(ctx, key); err != nil {
			logrus.WithError(err).WithField("key", key).Warn("Failed to invalidate related cache key")
		}
	}

	return nil
}

// WarmupCache warms up frequently accessed cache entries
func (wtc *WriteThroughCache) WarmupCache(ctx context.Context) error {
	logrus.Info("Starting cache warmup")

	start := time.Now()
	warmed := 0

	// Warmup popular functions
	if err := wtc.warmupPopularFunctions(ctx); err != nil {
		logrus.WithError(err).Warn("Failed to warmup popular functions")
	} else {
		warmed++
	}

	// Warmup recent executions
	if err := wtc.warmupRecentExecutions(ctx); err != nil {
		logrus.WithError(err).Warn("Failed to warmup recent executions")
	} else {
		warmed++
	}

	// Warmup function ratings
	if err := wtc.warmupFunctionRatings(ctx); err != nil {
		logrus.WithError(err).Warn("Failed to warmup function ratings")
	} else {
		warmed++
	}

	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"duration": duration,
		"items_warmed": warmed,
	}).Info("Cache warmup completed")

	return nil
}

// warmupPopularFunctions warms up cache for popular functions
func (wtc *WriteThroughCache) warmupPopularFunctions(ctx context.Context) error {
	var functions []struct {
		ID             uuid.UUID `json:"id"`
		Author         string    `json:"author"`
		Name           string    `json:"name"`
		PopularityScore int      `json:"popularity_score"`
	}

	// Get top 100 popular functions
	if err := wtc.db.Table("registry_functions").
		Select("id, author, name, popularity_score").
		Where("visibility = ? AND popularity_score > ?", "public", 10).
		Order("popularity_score DESC").
		Limit(100).
		Find(&functions).Error; err != nil {
		return err
	}

	// Cache function info
	for _, fn := range functions {
		cacheKey := wtc.cacheKeyGen.FunctionInfo(fn.Author, fn.Name)
		if err := wtc.redisCache.SetJSON(ctx, cacheKey, fn); err != nil {
			logrus.WithError(err).WithField("function", fn.Author+"/"+fn.Name).Warn("Failed to cache function info")
		}
	}

	return nil
}

// warmupRecentExecutions warms up cache for recent executions
func (wtc *WriteThroughCache) warmupRecentExecutions(ctx context.Context) error {
	var executions []struct {
		FunctionID uuid.UUID `json:"function_id"`
		Version    string    `json:"version"`
		Outcome    string    `json:"outcome"`
		Timestamp  time.Time `json:"timestamp"`
	}

	// Get recent executions (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	if err := wtc.db.Table("registry_function_executions").
		Select("function_id, version, outcome, timestamp").
		Where("timestamp > ?", since).
		Order("timestamp DESC").
		Limit(1000).
		Find(&executions).Error; err != nil {
		return err
	}

	// Cache execution stats by function
	stats := make(map[string]int)
	for _, exec := range executions {
		key := exec.FunctionID.String()
		stats[key]++
	}

	// Cache the stats
	for functionID, count := range stats {
		cacheKey := wtc.cacheKeyGen.ExecutionCount(functionID, "")
		if err := wtc.redisCache.SetJSON(ctx, cacheKey, map[string]int{"count": count}); err != nil {
			logrus.WithError(err).WithField("function_id", functionID).Warn("Failed to cache execution count")
		}
	}

	return nil
}

// warmupFunctionRatings warms up cache for function ratings
func (wtc *WriteThroughCache) warmupFunctionRatings(ctx context.Context) error {
	var ratings []struct {
		FunctionID     uuid.UUID `json:"function_id"`
		OverallScore   float64   `json:"overall_score"`
		TrustScore     float64   `json:"trust_score"`
		TotalRatings   int       `json:"total_ratings"`
	}

	// Get all function ratings
	if err := wtc.db.Table("registry_function_ratings").
		Select("function_id, overall_score, trust_score, total_ratings").
		Find(&ratings).Error; err != nil {
		return err
	}

	// Cache ratings
	for _, rating := range ratings {
		cacheKey := wtc.cacheKeyGen.FunctionRating(rating.FunctionID.String())
		if err := wtc.redisCache.SetJSON(ctx, cacheKey, rating); err != nil {
			logrus.WithError(err).WithField("function_id", rating.FunctionID).Warn("Failed to cache function rating")
		}
	}

	return nil
}

// updateCache updates multiple cache entries
func (wtc *WriteThroughCache) updateCache(ctx context.Context, operations []CacheWriteOperation) error {
	for _, op := range operations {
		if err := wtc.redisCache.SetJSONWithTTL(ctx, op.Key, op.Data, op.TTL); err != nil {
			logrus.WithError(err).WithField("key", op.Key).Warn("Failed to update cache")
			// Continue with other operations even if one fails
		}
	}
	return nil
}

// warmupRoutine runs periodic cache warming
func (wtc *WriteThroughCache) warmupRoutine() {
	for {
		select {
		case <-wtc.warmupTicker.C:
			wtc.warmupMutex.RLock()
			enabled := wtc.warmupEnabled
			wtc.warmupMutex.RUnlock()

			if enabled {
				if err := wtc.WarmupCache(context.Background()); err != nil {
					logrus.WithError(err).Error("Cache warmup failed")
				}
			}
		case <-wtc.stopWarmup:
			return
		}
	}
}

// GetCacheStats returns cache performance statistics
func (wtc *WriteThroughCache) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"warmup_enabled":  wtc.warmupEnabled,
		"warmup_interval": wtc.warmupInterval.String(),
		"cache_backend":   "redis",
	}
}

// CacheAside provides cache-aside pattern for read operations
func (wtc *WriteThroughCache) CacheAside(ctx context.Context, cacheKey string, dest interface{}, loader func() error) error {
	// Try cache first
	if err := wtc.redisCache.GetJSON(ctx, cacheKey, dest); err == nil {
		return nil // Cache hit
	}

	// Cache miss - load from source
	if err := loader(); err != nil {
		return err
	}

	// Update cache
	return wtc.redisCache.SetJSON(ctx, cacheKey, dest)
}