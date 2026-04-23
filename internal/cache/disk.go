package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FunctionCache is the GORM model for cached function execution results
type FunctionCache struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CacheKey   string          `json:"cache_key" gorm:"type:varchar(255);uniqueIndex;not null"`
	FunctionID uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index:idx_function_version"`
	Version    string          `json:"version" gorm:"type:varchar(50);not null;index:idx_function_version"`
	InputHash  string          `json:"input_hash" gorm:"type:varchar(64);not null"`
	OutputJSON json.RawMessage `json:"output_json" gorm:"type:jsonb;not null"`
	OutputSize int             `json:"output_size" gorm:"not null"`
	CreatedAt  time.Time       `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt  time.Time       `json:"expires_at" gorm:"not null;index:idx_expires_at"`
	HitCount   int             `json:"hit_count" gorm:"default:0"`
	LastHitAt  *time.Time      `json:"last_hit_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for GORM
func (FunctionCache) TableName() string {
	return "function_cache"
}

// DiskCache provides L2 persistent caching using GORM
// This survives restarts and handles repeated calls efficiently
type DiskCache struct {
	db   *gorm.DB
	stop chan struct{} // Stop signal for background cleanup goroutine
}

// NewDiskCache creates a new disk cache with the given GORM database
// Note: The function_cache table should be created via migrations (see migrations/ directory)
// This function assumes the table already exists
func NewDiskCache(db *gorm.DB) (*DiskCache, error) {
	return &DiskCache{db: db, stop: make(chan struct{})}, nil
}

// Get retrieves a cache entry by key
// Returns nil if not found or expired
func (d *DiskCache) Get(cacheKey string) (*FunctionCache, error) {
	var record FunctionCache
	err := d.db.Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).
		First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Update hit count asynchronously
	go d.incrementHitCount(record.ID)

	return &record, nil
}

// Set stores a cache entry
// Uses upsert to handle duplicates
func (d *DiskCache) Set(record *FunctionCache) error {
	record.ID = uuid.New()
	record.CreatedAt = time.Now()

	// Upsert using GORM - First try to find existing record
	existing := &FunctionCache{}
	err := d.db.Where("cache_key = ?", record.CacheKey).First(existing).Error

	if err == nil && existing.ID != uuid.Nil {
		// Record exists - update it
		return d.db.Model(existing).Updates(map[string]interface{}{
			"output_json": record.OutputJSON,
			"output_size": record.OutputSize,
			"expires_at":  record.ExpiresAt,
			"hit_count":   gorm.Expr("hit_count + 1"),
			"last_hit_at": time.Now(),
		}).Error
	}

	// Record doesn't exist - create new
	return d.db.Create(record).Error
}

// SetWithExpiry creates a cache record with the specified TTL
func (d *DiskCache) SetWithExpiry(cacheKey, functionID, version, inputHash string, output json.RawMessage, ttlSeconds int) error {
	fnUUID, err := uuid.Parse(functionID)
	if err != nil {
		return fmt.Errorf("invalid function ID %q: %w", functionID, err)
	}

	record := &FunctionCache{
		CacheKey:   cacheKey,
		FunctionID: fnUUID,
		Version:    version,
		InputHash:  inputHash,
		OutputJSON: output,
		OutputSize: len(output),
		ExpiresAt:  time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
	return d.Set(record)
}

// incrementHitCount updates the hit count for a cache entry
// This is called asynchronously to avoid blocking the read path
func (d *DiskCache) incrementHitCount(id uuid.UUID) {
	d.db.Model(&FunctionCache{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"hit_count":   gorm.Expr("hit_count + 1"),
			"last_hit_at": time.Now(),
		})
}

// Delete removes a specific cache entry
func (d *DiskCache) Delete(cacheKey string) error {
	return d.db.Where("cache_key = ?", cacheKey).Delete(&FunctionCache{}).Error
}

// DeleteByFunction removes all cache entries for a function
func (d *DiskCache) DeleteByFunction(functionID string) error {
	return d.db.Where("function_id = ?", functionID).Delete(&FunctionCache{}).Error
}

// DeleteByVersion removes all cache entries for a function version
func (d *DiskCache) DeleteByVersion(functionID, version string) error {
	return d.db.Where("function_id = ? AND version = ?", functionID, version).Delete(&FunctionCache{}).Error
}

// Cleanup removes all expired entries
// Should be called periodically (e.g., via cron or background goroutine)
func (d *DiskCache) Cleanup() (int64, error) {
	result := d.db.Where("expires_at < ?", time.Now()).Delete(&FunctionCache{})
	return result.RowsAffected, result.Error
}

// CleanupOlderThan removes all entries older than the specified duration
func (d *DiskCache) CleanupOlderThan(duration time.Duration) (int64, error) {
	cutoff := time.Now().Add(-duration)
	result := d.db.Where("created_at < ?", cutoff).Delete(&FunctionCache{})
	return result.RowsAffected, result.Error
}

// GetStats returns statistics about the cache
func (d *DiskCache) GetStats() (*DiskCacheStats, error) {
	stats := &DiskCacheStats{}

	// Total entries
	if err := d.db.Model(&FunctionCache{}).Count(&stats.TotalEntries).Error; err != nil {
		return nil, err
	}

	// Total size
	if err := d.db.Model(&FunctionCache{}).Select("COALESCE(SUM(output_size), 0)").Scan(&stats.TotalSizeBytes).Error; err != nil {
		return nil, err
	}

	// Total hits
	if err := d.db.Model(&FunctionCache{}).Select("COALESCE(SUM(hit_count), 0)").Scan(&stats.TotalHits).Error; err != nil {
		return nil, err
	}

	// Expired entries
	if err := d.db.Model(&FunctionCache{}).Where("expires_at < ?", time.Now()).Count(&stats.ExpiredEntries).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// DiskCacheStats holds statistics about the disk cache
type DiskCacheStats struct {
	TotalEntries   int64
	TotalSizeBytes int64
	TotalHits      int64
	ExpiredEntries int64
}

// StartCleanupJob starts a background job that periodically cleans up expired entries
func (d *DiskCache) StartCleanupJob(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d.Cleanup()
			case <-d.stop:
				return
			}
		}
	}()
}

// Stop stops the background cleanup job
func (d *DiskCache) Stop() {
	close(d.stop)
}

// GetByFunctionAndVersion retrieves all cache entries for a function version
func (d *DiskCache) GetByFunctionAndVersion(functionID, version string) ([]FunctionCache, error) {
	var records []FunctionCache
	err := d.db.Where("function_id = ? AND version = ? AND expires_at > ?",
		functionID, version, time.Now()).Find(&records).Error
	return records, err
}
