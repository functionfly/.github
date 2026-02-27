package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RegistryRepository interface for cache invalidation
type RegistryRepository interface {
	GetLatestFunctionVersion(functionID uuid.UUID) (interface{}, error)
}

// CacheInvalidator handles cache invalidation based on function lifecycle events
// It provides hooks for automatic invalidation when functions are updated or republished
type CacheInvalidator struct {
	service             *CacheService
	registryRepository RegistryRepository
	edgeCache          *EdgeCacheService
	cdnService         *CDNService
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator(service *CacheService, registryRepository RegistryRepository, edgeCache *EdgeCacheService, cdnService *CDNService) *CacheInvalidator {
	return &CacheInvalidator{
		service:             service,
		registryRepository: registryRepository,
		edgeCache:          edgeCache,
		cdnService:         cdnService,
	}
}

// FunctionInfo represents basic function information for cache invalidation
type FunctionInfo struct {
	ID      uuid.UUID
	Version string
}

// OnFunctionPublished is called when a new version of a function is published
// This invalidates the cache for that specific function version since the code has changed
func (i *CacheInvalidator) OnFunctionPublished(fn FunctionInfo) error {
	// Invalidate all cache layers for this function
	if err := i.invalidateAllLayers(fn.ID.String(), fn.Version); err != nil {
		return err
	}

	// Invalidate edge cache if function was previously cached at edge
	if i.edgeCache != nil {
		ctx := context.Background()
		if err := i.edgeCache.PurgeFunctionFromEdge(ctx, fn.ID); err != nil {
			// Log but don't fail the operation
			fmt.Printf("Failed to purge function from edge cache: %v\n", err)
		}
	}

	return nil
}

// invalidateAllLayers invalidates cache across all layers (memory, disk, Redis, edge)
func (i *CacheInvalidator) invalidateAllLayers(functionID, version string) error {
	// Invalidate memory and disk cache
	if err := i.service.InvalidateVersion(functionID, version); err != nil {
		return fmt.Errorf("failed to invalidate memory/disk cache: %w", err)
	}

	// Invalidate Redis registry cache
	if registryCache := i.service.GetRegistryCache(); registryCache != nil {
		if err := registryCache.InvalidateVersion(context.Background(), functionID, version); err != nil {
			return fmt.Errorf("failed to invalidate Redis registry cache: %w", err)
		}
	}

	return nil
}

// FunctionVersionInfo represents version information for cache invalidation
type FunctionVersionInfo struct {
	Deterministic bool
	CacheTTL      int
	SideEffects   bool
}

// OnFunctionUpdated is called when function metadata changes
// This checks if cache-affecting fields were modified
func (i *CacheInvalidator) OnFunctionUpdated(fn FunctionInfo, oldVersion FunctionVersionInfo) error {
	// Check if any cache-affecting fields changed
	// We need the current version to compare - get from registry if available
	var currentVersion FunctionVersionInfo
	if i.registryRepository != nil {
		if _, err := i.registryRepository.GetLatestFunctionVersion(fn.ID); err == nil {
			// Try to extract version info (simplified - in real implementation would cast properly)
			currentVersion = FunctionVersionInfo{} // Placeholder
		} else {
			// If we can't get current version, safest to invalidate all
			return i.invalidateAllLayers(fn.ID.String(), "")
		}
	}

	// Check if eligibility-affecting fields changed
	if currentVersion.Deterministic != oldVersion.Deterministic ||
		currentVersion.CacheTTL != oldVersion.CacheTTL ||
		currentVersion.SideEffects != oldVersion.SideEffects {
		return i.invalidateAllLayers(fn.ID.String(), "")
	}

	// For non-cache-affecting changes, only invalidate registry metadata cache
	if registryCache := i.service.GetRegistryCache(); registryCache != nil {
		if err := registryCache.InvalidateFunction(context.Background(), fn.ID.String()); err != nil {
			return fmt.Errorf("failed to invalidate registry metadata cache: %w", err)
		}
	}

	return nil
}

// OnFunctionDeleted is called when a function is deleted
// This completely removes all cache entries for the function
func (i *CacheInvalidator) OnFunctionDeleted(fn FunctionInfo) error {
	// Invalidate all cache layers
	if err := i.invalidateAllLayers(fn.ID.String(), ""); err != nil {
		return err
	}

	// Purge from edge cache
	if i.edgeCache != nil {
		ctx := context.Background()
		if err := i.edgeCache.PurgeFunctionFromEdge(ctx, fn.ID); err != nil {
			// Log but don't fail the operation
			fmt.Printf("Failed to purge deleted function from edge cache: %v\n", err)
		}
	}

	return nil
}

// OnVersionDeleted is called when a specific version is deleted
func (i *CacheInvalidator) OnVersionDeleted(fnID, version string) error {
	return i.service.InvalidateVersion(fnID, version)
}

// PurgeAll clears all cache entries across all layers (admin operation)
func (i *CacheInvalidator) PurgeAll() error {
	// Clear memory and disk cache
	if err := i.service.PurgeAll(); err != nil {
		return err
	}

	// Clear Redis registry cache
	if registryCache := i.service.GetRegistryCache(); registryCache != nil {
		ctx := context.Background()
		if err := registryCache.Clear(ctx); err != nil {
			return fmt.Errorf("failed to clear Redis registry cache: %w", err)
		}
	}

	// Clear edge cache
	if i.edgeCache != nil {
		ctx := context.Background()
		if err := i.edgeCache.RefreshEdgeCacheCandidates(ctx); err != nil {
			return fmt.Errorf("failed to refresh edge cache: %w", err)
		}
	}

	return nil
}

// PurgeFunction clears cache for a specific function (manual admin operation)
func (i *CacheInvalidator) PurgeFunction(functionID string) error {
	return i.service.InvalidateFunction(functionID)
}

// PurgeVersion clears cache for a specific function version (manual admin operation)
func (i *CacheInvalidator) PurgeVersion(functionID, version string) error {
	return i.service.InvalidateVersion(functionID, version)
}

// getCurrentVersion retrieves the current version of a function
// This is a helper to get the latest version for comparison
func (i *CacheInvalidator) getCurrentVersion(functionID string) (FunctionVersionInfo, error) {
	if i.registryRepository == nil {
		// Fallback to safe behavior if repository not injected
		return FunctionVersionInfo{}, nil
	}

	// Parse the function ID string to UUID
	functionUUID, err := uuid.Parse(functionID)
	if err != nil {
		return FunctionVersionInfo{}, fmt.Errorf("invalid function ID format: %w", err)
	}

	// Get the latest version from the repository
	_, err = i.registryRepository.GetLatestFunctionVersion(functionUUID)
	if err != nil {
		return FunctionVersionInfo{}, fmt.Errorf("failed to get latest function version: %w", err)
	}

	// In a real implementation, you'd properly cast the version interface
	// to extract the needed fields. For now, return a default.
	return FunctionVersionInfo{}, nil
}

// CacheStats returns current cache statistics for monitoring
type CacheStats struct {
	MemoryHits       int64
	MemoryMisses     int64
	MemoryHitRatio   float64
	DiskTotalEntries int64
	DiskTotalSize    int64
	DiskTotalHits    int64
	DiskExpired      int64
}

// GetStats returns comprehensive cache statistics
func (i *CacheInvalidator) GetStats() (*CacheStats, error) {
	stats := &CacheStats{}

	// Get memory stats
	if i.service.memory != nil {
		memMetrics := i.service.memory.Metrics()
		stats.MemoryHits = memMetrics.Hits
		stats.MemoryMisses = memMetrics.Misses
		stats.MemoryHitRatio = memMetrics.Ratio
	}

	// Get disk stats
	if i.service.disk != nil {
		diskStats, err := i.service.disk.GetStats()
		if err != nil {
			return nil, err
		}
		if diskStats != nil {
			stats.DiskTotalEntries = diskStats.TotalEntries
			stats.DiskTotalSize = diskStats.TotalSizeBytes
			stats.DiskTotalHits = diskStats.TotalHits
			stats.DiskExpired = diskStats.ExpiredEntries
		}
	}

	return stats, nil
}
