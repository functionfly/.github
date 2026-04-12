package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RegistryRedisCache provides Redis-based caching for registry metadata
// This is used for caching function info, versions, ratings, and other registry data
type RegistryRedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// RegistryCacheKey generates cache keys for registry data
type RegistryCacheKey struct{}

// NewRegistryCacheKey creates a new cache key generator
func NewRegistryCacheKey() *RegistryCacheKey {
	return &RegistryCacheKey{}
}

// FunctionInfo generates cache key for function info
func (k *RegistryCacheKey) FunctionInfo(author, name string) string {
	return fmt.Sprintf("registry:function:%s/%s", author, name)
}

// FunctionVersions generates cache key for function versions
func (k *RegistryCacheKey) FunctionVersions(functionID string) string {
	return fmt.Sprintf("registry:function:%s:versions", functionID)
}

// FunctionVersion generates cache key for specific function version
func (k *RegistryCacheKey) FunctionVersion(functionID, version string) string {
	return fmt.Sprintf("registry:function:%s:version:%s", functionID, version)
}

// FunctionRating generates cache key for function rating
func (k *RegistryCacheKey) FunctionRating(functionID string) string {
	return fmt.Sprintf("registry:function:%s:rating", functionID)
}

// FunctionStats generates cache key for function stats
func (k *RegistryCacheKey) FunctionStats(functionID string, since time.Time) string {
	return fmt.Sprintf("registry:function:%s:stats:%d", functionID, since.Unix())
}

// FunctionTrustStats generates cache key for function trust stats
func (k *RegistryCacheKey) FunctionTrustStats(functionID string, since time.Time) string {
	return fmt.Sprintf("registry:function:%s:trust_stats:%d", functionID, since.Unix())
}

// ConsumerDiversity generates cache key for consumer diversity stats
func (k *RegistryCacheKey) ConsumerDiversity(functionID string, since time.Time) string {
	return fmt.Sprintf("registry:function:%s:consumer_diversity:%d", functionID, since.Unix())
}

// SearchResults generates cache key for search results
func (k *RegistryCacheKey) SearchResults(query string, filters map[string]interface{}) string {
	// Create a deterministic key from filters
	filterStr := ""
	for k, v := range filters {
		filterStr += fmt.Sprintf("%s:%v:", k, v)
	}
	return fmt.Sprintf("registry:search:%s:%s", query, filterStr)
}

// ListFunctions generates cache key for function lists
func (k *RegistryCacheKey) ListFunctions(author, category, visibility string, limit, offset int) string {
	return fmt.Sprintf("registry:list:%s:%s:%s:%d:%d", author, category, visibility, limit, offset)
}

// SearchFunctions generates cache key for search results
func (k *RegistryCacheKey) SearchFunctions(query, category, runtime string, minRating float64, limit, offset int) string {
	return fmt.Sprintf("registry:search:q=%s:cat=%s:runtime=%s:rating=%.2f:limit=%d:offset=%d",
		query, category, runtime, minRating, limit, offset)
}

// ListFunctionsByTrustScore generates cache key for trust score ordered lists
func (k *RegistryCacheKey) ListFunctionsByTrustScore(category string, tags []string, visibility string, limit, offset int) string {
	tagsStr := ""
	for _, tag := range tags {
		tagsStr += tag + ","
	}
	if len(tagsStr) > 0 {
		tagsStr = tagsStr[:len(tagsStr)-1] // Remove trailing comma
	}
	return fmt.Sprintf("registry:trust_list:cat=%s:tags=%s:vis=%s:limit=%d:offset=%d",
		category, tagsStr, visibility, limit, offset)
}

// ExecutionCount generates cache key for execution count queries
func (k *RegistryCacheKey) ExecutionCount(functionID string, version string) string {
	if version == "" {
		return fmt.Sprintf("registry:function:%s:execution_count", functionID)
	}
	return fmt.Sprintf("registry:function:%s:version:%s:execution_count", functionID, version)
}

// RecentExecutions generates cache key for recent executions
func (k *RegistryCacheKey) RecentExecutions(functionID string, limit int) string {
	return fmt.Sprintf("registry:function:%s:recent_executions:%d", functionID, limit)
}

// ExecutionResult generates cache key for function execution result
func (k *RegistryCacheKey) ExecutionResult(functionID, version, inputHash string) string {
	return fmt.Sprintf("registry:execution:result:%s:%s:%s", functionID, version, inputHash)
}

// ExecutionResultTTL generates cache key for execution result TTL
func (k *RegistryCacheKey) ExecutionResultTTL(functionID, version, inputHash string) string {
	return fmt.Sprintf("registry:execution:ttl:%s:%s:%s", functionID, version, inputHash)
}

// FunctionDeterministic generates cache key to check if function is deterministic
func (k *RegistryCacheKey) FunctionDeterministic(functionID, version string) string {
	return fmt.Sprintf("registry:function:deterministic:%s:%s", functionID, version)
}

// FunctionCacheTTL generates cache key for function cache TTL setting
func (k *RegistryCacheKey) FunctionCacheTTL(functionID, version string) string {
	return fmt.Sprintf("registry:function:cache_ttl:%s:%s", functionID, version)
}

// NewRegistryRedisCache creates a new Redis cache for registry data
func NewRegistryRedisCache(client *redis.Client, defaultTTL time.Duration) *RegistryRedisCache {
	if defaultTTL == 0 {
		defaultTTL = 10 * time.Minute // Default 10 minute TTL for registry data
	}
	return &RegistryRedisCache{
		client: client,
		ttl:    defaultTTL,
	}
}

// Get retrieves data from cache by key
func (c *RegistryRedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if c.client == nil {
		return nil, ErrNotFound
	}
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		RecordRedisCacheMiss()
		return nil, ErrNotFound
	}
	if err != nil {
		RecordRedisCacheError("get")
		return nil, fmt.Errorf("failed to get from registry cache: %w", err)
	}
	RecordRedisCacheHit()
	return data, nil
}

// GetJSON retrieves and unmarshals JSON data from cache
func (c *RegistryRedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return nil
}

// Set stores data in cache with default TTL
func (c *RegistryRedisCache) Set(ctx context.Context, key string, data []byte) error {
	if c.client == nil {
		return nil
	}
	return c.SetWithTTL(ctx, key, data, c.ttl)
}

// SetWithTTL stores data in cache with specified TTL
func (c *RegistryRedisCache) SetWithTTL(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if c.client == nil {
		return nil
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		RecordRedisCacheError("set")
		return fmt.Errorf("failed to set registry cache: %w", err)
	}
	return nil
}

// SetJSON marshals and stores JSON data in cache
func (c *RegistryRedisCache) SetJSON(ctx context.Context, key string, data interface{}) error {
	return c.SetJSONWithTTL(ctx, key, data, c.ttl)
}

// SetJSONWithTTL marshals and stores JSON data in cache with specified TTL
func (c *RegistryRedisCache) SetJSONWithTTL(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for cache: %w", err)
	}
	return c.SetWithTTL(ctx, key, jsonData, ttl)
}

// Delete removes data from cache
func (c *RegistryRedisCache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return nil
	}
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from registry cache: %w", err)
	}
	return nil
}

// DeleteByPattern removes all keys matching a pattern
func (c *RegistryRedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if c.client == nil {
		return nil
	}
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
		logrus.Debugf("Deleted %d cache keys matching pattern: %s", len(keys), pattern)
	}

	return nil
}

// InvalidateFunction invalidates all cache entries for a function
func (c *RegistryRedisCache) InvalidateFunction(ctx context.Context, functionID string) error {
	patterns := []string{
		fmt.Sprintf("registry:function:%s:*", functionID),
		fmt.Sprintf("registry:list:*:%s:*", functionID), // Clear any lists that might contain this function
		"registry:list:*", // List keys are registry:list:author:category:visibility:limit:offset — clear all so new publishes show up
		"registry:trust_list:*",
	}

	for _, pattern := range patterns {
		if err := c.DeleteByPattern(ctx, pattern); err != nil {
			logrus.Warnf("Failed to delete cache pattern %s: %v", pattern, err)
		}
	}

	return nil
}

// InvalidateVersion invalidates cache entries for a specific function version
func (c *RegistryRedisCache) InvalidateVersion(ctx context.Context, functionID, version string) error {
	patterns := []string{
		fmt.Sprintf("registry:function:%s:version:%s", functionID, version),
		fmt.Sprintf("registry:function:%s:versions", functionID),
	}

	for _, pattern := range patterns {
		if err := c.DeleteByPattern(ctx, pattern); err != nil {
			logrus.Warnf("Failed to delete cache pattern %s: %v", pattern, err)
		}
	}

	return nil
}

// InvalidateSearchResults invalidates search result caches
func (c *RegistryRedisCache) InvalidateSearchResults(ctx context.Context) error {
	pattern := "registry:search:*"
	return c.DeleteByPattern(ctx, pattern)
}

// InvalidateListResults invalidates function list caches (so description/category updates appear after publish)
func (c *RegistryRedisCache) InvalidateListResults(ctx context.Context) error {
	for _, pattern := range []string{"registry:list:*", "registry:trust_list:*"} {
		if err := c.DeleteByPattern(ctx, pattern); err != nil {
			return err
		}
	}
	return nil
}

// Clear clears all registry cache entries
func (c *RegistryRedisCache) Clear(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	pattern := "registry:*"
	return c.DeleteByPattern(ctx, pattern)
}

// GetStats returns cache statistics
func (c *RegistryRedisCache) GetStats(ctx context.Context) (*RegistryCacheStats, error) {
	if c.client == nil {
		return &RegistryCacheStats{}, nil
	}
	// Count registry keys
	iter := c.client.Scan(ctx, 0, "registry:*", 0).Iterator()
	var count int64
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan registry cache: %w", err)
	}

	// Get memory info
	info := c.client.Info(ctx, "memory")
	if info.Err() != nil {
		return nil, fmt.Errorf("failed to get Redis memory info: %w", info.Err())
	}

	return &RegistryCacheStats{
		TotalKeys:  count,
		MemoryInfo: info.Val(),
	}, nil
}

// RegistryCacheStats holds statistics about the registry cache
type RegistryCacheStats struct {
	TotalKeys  int64
	MemoryInfo string
}
