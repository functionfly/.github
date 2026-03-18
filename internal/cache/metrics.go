package cache

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// CacheHits tracks total cache hits by layer
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_cache_hits_total",
			Help: "Total number of cache hits by layer (memory, disk)",
		},
		[]string{"layer"},
	)

	// CacheMisses tracks total cache misses
	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	// CacheSize tracks current cache size in bytes
	CacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_cache_size_bytes",
			Help: "Current cache size in bytes",
		},
	)

	// CacheMemoryCost tracks memory used by cache
	CacheMemoryCost = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_cache_memory_bytes",
			Help: "Memory used by cache in bytes",
		},
	)

	// CacheEvictions tracks total entries evicted
	CacheEvictions = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_cache_evictions_total",
			Help: "Total number of cache entries evicted",
		},
	)

	// CacheErrors tracks cache-related errors
	CacheErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_cache_errors_total",
			Help: "Total number of cache-related errors",
		},
		[]string{"operation"}, // "get", "set", "invalidate"
	)

	// CacheHitRatio tracks cache hit ratio
	CacheHitRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_cache_hit_ratio",
			Help: "Cache hit ratio (0-1)",
		},
	)

	// CacheTTLExpired tracks entries that expired due to TTL
	CacheTTLExpired = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_cache_ttl_expired_total",
			Help: "Total number of cache entries that expired due to TTL",
		},
	)

	RedisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_redis_cache_hits_total",
			Help: "Total number of Redis registry cache hits",
		},
	)

	RedisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_redis_cache_misses_total",
			Help: "Total number of Redis registry cache misses",
		},
	)

	RedisCacheErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_redis_cache_errors_total",
			Help: "Total number of Redis cache errors",
		},
		[]string{"operation"},
	)

	RedisCacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_redis_cache_entries",
			Help: "Number of entries in Redis registry cache",
		},
	)

	CDNHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_cdn_hits_total",
			Help: "Total number of CDN cache hits",
		},
	)

	CDNMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_cdn_misses_total",
			Help: "Total number of CDN cache misses",
		},
	)

	CDNPurges = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_cdn_purges_total",
			Help: "Total number of CDN cache purges",
		},
	)

	EdgeCacheFunctions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_edge_cache_functions",
			Help: "Number of functions currently cached at edge locations",
		},
	)

	EdgeCachePopularity = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_edge_cache_popularity_total",
			Help: "Total popularity score of functions cached at edge",
		},
	)
)

// RecordCacheHit records a cache hit in metrics
func RecordCacheHit(layer string) {
	CacheHits.WithLabelValues(layer).Inc()
}

// RecordCacheMiss records a cache miss in metrics
func RecordCacheMiss() {
	CacheMisses.Inc()
}

// RecordCacheError records a cache error
func RecordCacheError(operation string) {
	CacheErrors.WithLabelValues(operation).Inc()
}

// UpdateCacheMetrics updates general cache metrics
func UpdateCacheMetrics(metrics *CacheMetrics) {
	if metrics != nil {
		CacheHitRatio.Set(metrics.Ratio)
		CacheEvictions.Add(float64(metrics.Evictions))
	}
}

// Redis cache metrics functions
func RecordRedisCacheHit() {
	RedisCacheHits.Inc()
}

func RecordRedisCacheMiss() {
	RedisCacheMisses.Inc()
}

func RecordRedisCacheError(operation string) {
	RedisCacheErrors.WithLabelValues(operation).Inc()
}

func UpdateRedisCacheSize(size int64) {
	RedisCacheSize.Set(float64(size))
}

// CDN metrics functions
func RecordCDNHit() {
	CDNHits.Inc()
}

func RecordCDNMiss() {
	CDNMisses.Inc()
}

func RecordCDNPurge() {
	CDNPurges.Inc()
}

func UpdateEdgeCacheMetrics(functionCount int, totalPopularity int) {
	EdgeCacheFunctions.Set(float64(functionCount))
	EdgeCachePopularity.Set(float64(totalPopularity))
}

// CacheMonitor provides comprehensive cache monitoring
type CacheMonitor struct {
	cacheService *CacheService
	cdnService   *CDNService
	edgeCache    *EdgeCacheService
}

// NewCacheMonitor creates a new cache monitor
func NewCacheMonitor(cacheService *CacheService, cdnService *CDNService, edgeCache *EdgeCacheService) *CacheMonitor {
	return &CacheMonitor{
		cacheService: cacheService,
		cdnService:   cdnService,
		edgeCache:    edgeCache,
	}
}

// CollectMetrics collects and updates all cache metrics.
func (m *CacheMonitor) CollectMetrics(ctx context.Context) error {

	// Update memory cache metrics
	if memMetrics := m.cacheService.GetMemoryStats(); memMetrics != nil {
		UpdateCacheMetrics(memMetrics)
	}

	// Update Redis registry cache metrics
	if registryCache := m.cacheService.GetRegistryCache(); registryCache != nil {
		if stats, err := registryCache.GetStats(ctx); err == nil {
			UpdateRedisCacheSize(stats.TotalKeys)
		}
	}

	// Update edge cache metrics
	if m.edgeCache != nil {
		if stats, err := m.edgeCache.GetEdgeCacheStats(ctx); err == nil {
			functionCount := 0
			totalPopularity := 0
			if fc, ok := stats["active_functions"].(int); ok {
				functionCount = fc
			}
			if tp, ok := stats["total_popularity"].(int); ok {
				totalPopularity = tp
			}
			UpdateEdgeCacheMetrics(functionCount, totalPopularity)
		}
	}

	return nil
}

// GetComprehensiveStats returns comprehensive cache statistics.
func (m *CacheMonitor) GetComprehensiveStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Memory cache stats
	if memMetrics := m.cacheService.GetMemoryStats(); memMetrics != nil {
		stats["memory"] = map[string]interface{}{
			"hits":   memMetrics.Hits,
			"misses": memMetrics.Misses,
			"ratio":  memMetrics.Ratio,
		}
	}

	// Disk cache stats
	if diskStats, err := m.cacheService.GetDiskStats(); err == nil && diskStats != nil {
		stats["disk"] = map[string]interface{}{
			"total_entries": diskStats.TotalEntries,
			"total_size":    diskStats.TotalSizeBytes,
			"total_hits":    diskStats.TotalHits,
			"expired":       diskStats.ExpiredEntries,
		}
	}

	// Redis registry cache stats
	if registryCache := m.cacheService.GetRegistryCache(); registryCache != nil {
		if redisStats, err := registryCache.GetStats(ctx); err == nil {
			stats["redis_registry"] = map[string]interface{}{
				"total_keys": redisStats.TotalKeys,
				"memory_info": redisStats.MemoryInfo,
			}
		}
	}

	// CDN stats
	if m.cdnService != nil {
		stats["cdn"] = m.cdnService.GetCDNStats()
	}

	// Edge cache stats
	if m.edgeCache != nil {
		if edgeStats, err := m.edgeCache.GetEdgeCacheStats(ctx); err == nil {
			stats["edge"] = edgeStats
		}
	}

	return stats, nil
}
