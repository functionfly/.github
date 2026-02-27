package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/sirupsen/logrus"
)

// ExecutionCache provides caching for function execution results
type ExecutionCache struct {
	redisCache *RegistryRedisCache
	keyGen     *RegistryCacheKey
}


// NewExecutionCache creates a new execution cache
func NewExecutionCache(redisCache *RegistryRedisCache) *ExecutionCache {
	return &ExecutionCache{
		redisCache: redisCache,
		keyGen:     NewRegistryCacheKey(),
	}
}

// computeInputHash generates a SHA256 hash of the input for cache keying
func (c *ExecutionCache) computeInputHash(input interface{}) (string, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input: %w", err)
	}

	hash := sha256.Sum256(inputBytes)
	return hex.EncodeToString(hash[:]), nil
}

// GetExecutionResult retrieves a cached execution result
func (c *ExecutionCache) GetExecutionResult(ctx context.Context, functionID, version string, input interface{}) (*types.ExecutionResult, error) {
	inputHash, err := c.computeInputHash(input)
	if err != nil {
		return nil, err
	}

	cacheKey := c.keyGen.ExecutionResult(functionID, version, inputHash)
	var result types.ExecutionResult
	if err := c.redisCache.GetJSON(ctx, cacheKey, &result); err != nil {
		if err == ErrNotFound {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached execution result: %w", err)
	}

	// Check if result has expired
	if time.Now().After(result.ExpiresAt) {
		// Clean up expired result
		c.redisCache.Delete(ctx, cacheKey)
		return nil, nil
	}

	// Update access statistics
	result.LastAccessed = time.Now()
	result.HitCount++
	if err := c.redisCache.SetJSON(ctx, cacheKey, result); err != nil {
		logrus.WithError(err).Warn("Failed to update execution result access stats")
	}

	return &result, nil
}

// SetExecutionResult caches an execution result
func (c *ExecutionCache) SetExecutionResult(ctx context.Context, functionID, version string, input interface{}, result *types.ExecutionResult) error {
	inputHash, err := c.computeInputHash(input)
	if err != nil {
		return err
	}

	result.InputHash = inputHash
	result.ExecutedAt = time.Now()
	result.LastAccessed = time.Now()
	result.HitCount = 0 // First execution, so hit count starts at 0

	cacheKey := c.keyGen.ExecutionResult(functionID, version, inputHash)
	return c.redisCache.SetJSONWithTTL(ctx, cacheKey, result, result.TTL)
}

// InvalidateFunctionResults invalidates all cached results for a function
func (c *ExecutionCache) InvalidateFunctionResults(ctx context.Context, functionID string) error {
	// Use pattern matching to delete all execution results for this function
	pattern := fmt.Sprintf("registry:execution:result:%s:*", functionID)
	return c.redisCache.DeleteByPattern(ctx, pattern)
}

// InvalidateVersionResults invalidates all cached results for a specific function version
func (c *ExecutionCache) InvalidateVersionResults(ctx context.Context, functionID, version string) error {
	pattern := fmt.Sprintf("registry:execution:result:%s:%s:*", functionID, version)
	return c.redisCache.DeleteByPattern(ctx, pattern)
}

// GetCacheStats returns statistics about the execution cache
func (c *ExecutionCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Count cached execution results
	iter := c.redisCache.client.Scan(ctx, 0, "registry:execution:result:*", 0).Iterator()
	resultCount := 0
	for iter.Next(ctx) {
		resultCount++
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to count cached results: %w", err)
	}

	// Count TTL entries
	iter = c.redisCache.client.Scan(ctx, 0, "registry:execution:ttl:*", 0).Iterator()
	ttlCount := 0
	for iter.Next(ctx) {
		ttlCount++
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to count TTL entries: %w", err)
	}

	return map[string]interface{}{
		"cached_results": resultCount,
		"ttl_entries":    ttlCount,
		"total_entries":  resultCount + ttlCount,
	}, nil
}

// IsFunctionDeterministic checks if a function version is deterministic and cacheable
func (c *ExecutionCache) IsFunctionDeterministic(ctx context.Context, functionID, version string) (bool, error) {
	cacheKey := c.keyGen.FunctionDeterministic(functionID, version)

	// Try cache first
	var deterministic bool
	if err := c.redisCache.GetJSON(ctx, cacheKey, &deterministic); err == nil {
		return deterministic, nil
	}

	// Cache miss - would need to query database here, but for now return false
	// In real implementation, this would check the function metadata
	return false, nil
}

// GetFunctionCacheTTL retrieves the cache TTL for a function version
func (c *ExecutionCache) GetFunctionCacheTTL(ctx context.Context, functionID, version string) (time.Duration, error) {
	cacheKey := c.keyGen.FunctionCacheTTL(functionID, version)

	var ttlSeconds int
	if err := c.redisCache.GetJSON(ctx, cacheKey, &ttlSeconds); err != nil {
		if err == ErrNotFound {
			return 0, nil // No TTL set
		}
		return 0, fmt.Errorf("failed to get function cache TTL: %w", err)
	}

	return time.Duration(ttlSeconds) * time.Second, nil
}

// SetFunctionCacheTTL sets the cache TTL for a function version
func (c *ExecutionCache) SetFunctionCacheTTL(ctx context.Context, functionID, version string, ttl time.Duration) error {
	cacheKey := c.keyGen.FunctionCacheTTL(functionID, version)
	ttlSeconds := int(ttl.Seconds())
	return c.redisCache.SetJSON(ctx, cacheKey, ttlSeconds)
}

// RecordExecution records an execution in the cache for analytics
func (c *ExecutionCache) RecordExecution(ctx context.Context, executionID string, functionID string, version string, outcome string, durationMs int, cached bool, timestamp time.Time) error {
	// Store execution metadata for analytics
	executionKey := fmt.Sprintf("registry:execution:record:%s:%d", executionID, timestamp.Unix())

	executionData := map[string]interface{}{
		"function_id": functionID,
		"version":     version,
		"outcome":     outcome,
		"duration_ms": durationMs,
		"cached":      cached,
		"timestamp":   timestamp,
	}

	// Keep execution records for 30 days
	return c.redisCache.SetJSONWithTTL(ctx, executionKey, executionData, 30*24*time.Hour)
}