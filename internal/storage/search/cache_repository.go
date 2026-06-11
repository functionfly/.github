package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CacheRepository handles database operations for search result caching
type CacheRepository struct {
	db  *gorm.DB
	ttl time.Duration
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(db *gorm.DB, ttlSeconds int) *CacheRepository {
	return &CacheRepository{
		db:  db,
		ttl: time.Duration(ttlSeconds) * time.Second,
	}
}

// Get retrieves a cached result by key
func (r *CacheRepository) Get(ctx context.Context, cacheKey string) (*CacheEntry, error) {
	var entry CacheEntry
	err := r.db.WithContext(ctx).
		Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).
		First(&entry).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	return &entry, nil
}

// Set stores a result in cache
func (r *CacheRepository) Set(ctx context.Context, entry *CacheEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CachedAt.IsZero() {
		entry.CachedAt = time.Now()
	}
	if entry.ExpiresAt.IsZero() {
		entry.ExpiresAt = time.Now().Add(r.ttl)
	}

	// Use upsert to handle concurrent writes
	return r.db.WithContext(ctx).Save(entry).Error
}

// Delete removes a cached entry
func (r *CacheRepository) Delete(ctx context.Context, cacheKey string) error {
	return r.db.WithContext(ctx).
		Where("cache_key = ?", cacheKey).
		Delete(&CacheEntry{}).Error
}

// DeleteByTool removes all cached entries for a tool
func (r *CacheRepository) DeleteByTool(ctx context.Context, toolName string) error {
	return r.db.WithContext(ctx).
		Where("tool_name = ?", toolName).
		Delete(&CacheEntry{}).Error
}

// Cleanup removes all expired entries
func (r *CacheRepository) Cleanup(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&CacheEntry{})

	return result.RowsAffected, result.Error
}

// GetStats returns cache statistics
func (r *CacheRepository) GetStats(ctx context.Context) (*CacheStats, error) {
	var stats CacheStats

	// Total entries
	if err := r.db.WithContext(ctx).Model(&CacheEntry{}).Count(&stats.TotalEntries).Error; err != nil {
		return nil, err
	}

	// Expired entries
	if err := r.db.WithContext(ctx).
		Model(&CacheEntry{}).
		Where("expires_at < ?", time.Now()).
		Count(&stats.ExpiredEntries).Error; err != nil {
		return nil, err
	}

	// Entries by tool
	type toolCount struct {
		ToolName string
		Count    int64
	}

	var toolCounts []toolCount
	if err := r.db.WithContext(ctx).
		Model(&CacheEntry{}).
		Select("tool_name, COUNT(*) as count").
		Group("tool_name").
		Scan(&toolCounts).Error; err != nil {
		return nil, err
	}

	stats.EntriesByTool = make(map[string]int64, len(toolCounts))
	for _, tc := range toolCounts {
		stats.EntriesByTool[tc.ToolName] = tc.Count
	}

	return &stats, nil
}

// CacheStats holds cache statistics
type CacheStats struct {
	TotalEntries    int64
	ExpiredEntries  int64
	EntriesByTool   map[string]int64
}

// ValidateCacheResult is a helper to unmarshal cached results
func ValidateCacheResult[T any](entry *CacheEntry) (*T, error) {
	if entry == nil {
		return nil, nil
	}

	var result T
	if err := json.Unmarshal(entry.Results, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached result: %w", err)
	}

	return &result, nil
}