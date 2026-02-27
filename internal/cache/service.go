package cache

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	// ErrNotEligible is returned when a function is not eligible for caching
	ErrNotEligible = errors.New("function not eligible for caching")
	// ErrNotFound is returned when a cache entry is not found
	ErrNotFound = errors.New("cache entry not found")
)

// CacheService provides the main caching interface
// It coordinates between memory (L1), disk (L2), Redis (L3), and CDN (L4) cache layers
type CacheService struct {
	memory        *MemoryCache
	disk          *DiskCache
	redisRegistry *RegistryRedisCache
	config        *CacheConfig
}

// CacheConfig holds configuration for the cache service
type CacheConfig struct {
	MaxMemoryMB         int64 // Maximum memory for L1 cache in MB
	EnableDiskCache     bool  // Whether to enable L2 disk cache
	EnableRedisCache    bool  // Whether to enable L3 Redis cache for registry data
	EnableCDNCaching    bool  // Whether to set CDN headers
	DefaultTTL          int   // Default TTL in seconds when function doesn't specify
	RedisRegistryTTL    int   // TTL for registry data in Redis (seconds)
}

// CacheResult contains the result of a cache get or execute operation
type CacheResult struct {
	Output    json.RawMessage // The cached or computed output
	FromCache bool            // Whether this was from cache
	Layer     string          // Which layer: "memory", "disk", "none"
	Hit       bool            // Whether this was a cache hit
}

// NewCacheService creates a new cache service
func NewCacheService(db *gorm.DB, redisClient *redis.Client, config *CacheConfig) (*CacheService, error) {
	// Create memory cache
	memory, err := NewMemoryCache(config.MaxMemoryMB)
	if err != nil {
		return nil, err
	}

	// Create disk cache
	var disk *DiskCache
	if config.EnableDiskCache {
		disk, err = NewDiskCache(db)
		if err != nil {
			return nil, err
		}
		// Start background cleanup
		disk.StartCleanupJob(1 * time.Hour)
	}

	// Create Redis registry cache
	var redisRegistry *RegistryRedisCache
	if config.EnableRedisCache && redisClient != nil {
		redisRegistryTTL := time.Duration(config.RedisRegistryTTL) * time.Second
		if redisRegistryTTL == 0 {
			redisRegistryTTL = 10 * time.Minute // Default 10 minutes
		}
		redisRegistry = NewRegistryRedisCache(redisClient, redisRegistryTTL)
	}

	return &CacheService{
		memory:        memory,
		disk:          disk,
		redisRegistry: redisRegistry,
		config:        config,
	}, nil
}

// GetOrExecute checks the cache and executes the function if needed
// This is the main entry point for cached function execution
func (c *CacheService) GetOrExecute(
	eligibility EligibilityResult,
	input []byte,
	executeFn func() (json.RawMessage, error),
) (*CacheResult, error) {
	// If not eligible, just execute without caching
	if !eligibility.Eligible {
		output, err := executeFn()
		if err != nil {
			return nil, err
		}
		return &CacheResult{
			Output:    output,
			FromCache: false,
			Layer:     "none",
			Hit:       false,
		}, nil
	}

	// Normalize input for consistent cache key
	normalizedInput, err := NormalizeInput(input)
	if err != nil {
		// If normalization fails, execute without caching
		output, err := executeFn()
		if err != nil {
			return nil, err
		}
		return &CacheResult{
			Output:    output,
			FromCache: false,
			Layer:     "none",
			Hit:       false,
		}, nil
	}

	// Generate cache key
	cacheKey := GenerateCacheKey(eligibility.FunctionID, eligibility.Version, normalizedInput)

	// Check L1: Memory cache
	if cached, found := c.memory.Get(cacheKey); found {
		RecordCacheHit("memory")
		return &CacheResult{
			Output:    cached,
			FromCache: true,
			Layer:     "memory",
			Hit:       true,
		}, nil
	}

	// Check L2: Disk cache
	if c.disk != nil {
		record, err := c.disk.Get(cacheKey)
		if err == nil && record != nil {
			// Populate L1 cache
			ttl := time.Duration(eligibility.TTL) * time.Second
			c.memory.Set(cacheKey, record.OutputJSON, ttl)

			RecordCacheHit("disk")
			return &CacheResult{
				Output:    record.OutputJSON,
				FromCache: true,
				Layer:     "disk",
				Hit:       true,
			}, nil
		}
	}

	// Execute function (cache miss)
	output, err := executeFn()
	if err != nil {
		RecordCacheError("execute")
		return nil, err
	}

	RecordCacheMiss()

	// Validate and re-serialize output to prevent poisoning
	validated, err := ValidateOutput(output)
	if err != nil {
		// If validation fails, execute without caching
		return &CacheResult{
			Output:    output,
			FromCache: false,
			Layer:     "none",
			Hit:       false,
		}, nil
	}

	// Store in L1 cache
	ttl := time.Duration(eligibility.TTL) * time.Second
	c.memory.Set(cacheKey, validated, ttl)

	// Store in L2 cache
	if c.disk != nil {
		inputHash := HashInput(normalizedInput)
		_ = c.disk.SetWithExpiry(cacheKey, eligibility.FunctionID, eligibility.Version, inputHash, validated, eligibility.TTL)
	}

	return &CacheResult{
		Output:    validated,
		FromCache: false,
		Layer:     "none",
		Hit:       false,
	}, nil
}

// InvalidateFunction invalidates all cache entries for a function
func (c *CacheService) InvalidateFunction(functionID string) error {
	// Clear from memory - now selective invalidation
	if c.memory != nil {
		c.memory.DeleteByFunction(functionID)
	}

	// Clear from disk
	if c.disk != nil {
		return c.disk.DeleteByFunction(functionID)
	}
	return nil
}

// InvalidateVersion invalidates cache for a specific version
func (c *CacheService) InvalidateVersion(functionID, version string) error {
	// Clear from memory - now selective invalidation
	if c.memory != nil {
		c.memory.DeleteByVersion(functionID, version)
	}

	// Clear from disk
	if c.disk != nil {
		return c.disk.DeleteByVersion(functionID, version)
	}
	return nil
}

// PurgeAll clears all cache entries
func (c *CacheService) PurgeAll() error {
	if c.memory != nil {
		c.memory.Clear()
	}

	if c.disk != nil {
		_, err := c.disk.Cleanup()
		return err
	}
	return nil
}

// GetMemoryStats returns memory cache statistics
func (c *CacheService) GetMemoryStats() *CacheMetrics {
	if c.memory == nil {
		return nil
	}
	return c.memory.Metrics()
}

// GetDiskStats returns disk cache statistics
func (c *CacheService) GetDiskStats() (*DiskCacheStats, error) {
	if c.disk == nil {
		return nil, nil
	}
	return c.disk.GetStats()
}

// Close gracefully closes all cache resources
func (c *CacheService) Close() error {
	if c.memory != nil {
		return c.memory.Close()
	}
	return nil
}

// Memory returns the memory cache for direct access
func (c *CacheService) Memory() *MemoryCache {
	return c.memory
}

// Disk returns the disk cache for direct access
func (c *CacheService) Disk() *DiskCache {
	return c.disk
}

// GetRegistryCache returns the Redis registry cache
func (c *CacheService) GetRegistryCache() *RegistryRedisCache {
	return c.redisRegistry
}

// IsRedisCacheEnabled returns whether Redis registry cache is enabled
func (c *CacheService) IsRedisCacheEnabled() bool {
	return c.redisRegistry != nil
}

// DefaultCacheConfig returns a sensible default configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxMemoryMB:         100,  // 100MB default
		EnableDiskCache:     true,
		EnableRedisCache:    false, // Redis disabled by default (requires explicit setup)
		EnableCDNCaching:    true,
		DefaultTTL:          3600, // 1 hour default
		RedisRegistryTTL:    600,  // 10 minutes default for registry data
	}
}
