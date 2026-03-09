package generation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// redisGenerationCachePrefix is the key prefix for generation cache entries
	redisGenerationCachePrefix = "generation:cache:"

	// defaultGenerationCacheTTL is the default TTL for generation cache entries
	defaultGenerationCacheTTL = 30 * time.Minute
)

// RedisGenerationCache is a Redis-backed implementation of GenerationCache
// that enables distributed caching across multiple service instances.
type RedisGenerationCache struct {
	client *redis.Client
	ttl    time.Duration
}

// RedisCacheConfig holds configuration for the Redis generation cache
type RedisCacheConfig struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}

// NewRedisCacheConfig creates a Redis cache config from environment variables
func NewRedisCacheConfig() *RedisCacheConfig {
	ttl := defaultGenerationCacheTTL
	if ttlStr := os.Getenv("GENERATION_CACHE_TTL"); ttlStr != "" {
		if parsed, err := strconv.Atoi(ttlStr); err == nil && parsed > 0 {
			ttl = time.Duration(parsed) * time.Second
		}
	}

	return &RedisCacheConfig{
		Addr:     getEnvString("REDIS_ADDR", "localhost:6379"),
		Password: getEnvString("REDIS_PASSWORD", ""),
		DB:       getEnvInt("GENERATION_CACHE_REDIS_DB", 0),
		TTL:      ttl,
	}
}

// NewRedisGenerationCache creates a new Redis-backed generation cache
func NewRedisGenerationCache(client *redis.Client, ttl time.Duration) *RedisGenerationCache {
	if ttl <= 0 {
		ttl = defaultGenerationCacheTTL
	}
	return &RedisGenerationCache{
		client: client,
		ttl:    ttl,
	}
}

// NewRedisGenerationCacheFromConfig creates a new Redis-backed generation cache from config
func NewRedisGenerationCacheFromConfig(config *RedisCacheConfig) (*RedisGenerationCache, error) {
	if config == nil {
		config = NewRedisCacheConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis for generation cache: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"addr": config.Addr,
		"db":   config.DB,
		"ttl":  config.TTL,
	}).Info("Initialized Redis generation cache")

	return NewRedisGenerationCache(client, config.TTL), nil
}

// Get retrieves a cached generation result by request key
func (c *RedisGenerationCache) Get(ctx context.Context, req *GenerationRequest) (*CachedGeneration, bool) {
	if c == nil || c.client == nil {
		return nil, false
	}

	key := cacheKey(req)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		logrus.WithError(err).WithField("key", key).Warn("Failed to get from generation cache")
		return nil, false
	}

	var cached CachedGeneration
	if err := json.Unmarshal(data, &cached); err != nil {
		logrus.WithError(err).WithField("key", key).Warn("Failed to unmarshal cached generation")
		return nil, false
	}

	// Check if the entry has expired
	if time.Now().UTC().After(cached.ExpiresAt) {
		// Delete the expired entry
		_ = c.client.Del(ctx, key)
		return nil, false
	}

	return &cached, true
}

// Put stores a generation result in the cache
func (c *RedisGenerationCache) Put(ctx context.Context, req *GenerationRequest, value CachedGeneration) {
	if c == nil || c.client == nil {
		return
	}

	value.StoredAt = time.Now().UTC()
	value.ExpiresAt = value.StoredAt.Add(c.ttl)
	value.PromptKey = cacheKey(req)

	data, err := json.Marshal(value)
	if err != nil {
		logrus.WithError(err).WithField("key", value.PromptKey).Warn("Failed to marshal cached generation")
		return
	}

	if err := c.client.Set(ctx, value.PromptKey, data, c.ttl).Err(); err != nil {
		logrus.WithError(err).WithField("key", value.PromptKey).Warn("Failed to put generation cache")
	}
}

// Invalidate removes cache entries matching the predicate
// Note: This is a best-effort operation in Redis since we can't iterate efficiently.
// We scan for keys matching the prefix and filter by predicate.
func (c *RedisGenerationCache) Invalidate(ctx context.Context, predicate func(CachedGeneration) bool) {
	if c == nil || c.client == nil || predicate == nil {
		return
	}

	// Scan for all generation cache keys
	pattern := redisGenerationCachePrefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()

	var keysToDelete []string
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := c.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var cached CachedGeneration
		if err := json.Unmarshal(data, &cached); err != nil {
			continue
		}

		if predicate(cached) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	if err := iter.Err(); err != nil {
		logrus.WithError(err).Warn("Error scanning generation cache keys")
	}

	// Delete matching keys in batch
	if len(keysToDelete) > 0 {
		if err := c.client.Del(ctx, keysToDelete...).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to delete generation cache entries")
		}
		logrus.WithField("count", len(keysToDelete)).Info("Invalidated generation cache entries")
	}
}

// Ping checks the Redis connection
func (c *RedisGenerationCache) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("Redis generation cache not initialized")
	}
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisGenerationCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// NewGenerationCache creates a GenerationCache based on configuration.
// If redisClient is provided and useRedis is true, it returns a Redis-backed cache.
// Otherwise, it returns an in-memory cache as fallback.
func NewGenerationCache(redisClient *redis.Client, useRedis bool, ttl time.Duration) GenerationCache {
	if useRedis && redisClient != nil {
		return NewRedisGenerationCache(redisClient, ttl)
	}
	return NewInMemoryGenerationCache(ttl)
}

// GetEnvString is a helper to get string env var with default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt is a helper to get int env var with default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
