package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFunction creates a new function in the registry
func (r *RegistryRepository) CreateFunction(fn *RegistryFunction) error {
	fn.ID = uuid.New()
	fn.CreatedAt = time.Now()
	fn.UpdatedAt = time.Now()

	if err := r.db.Create(fn).Error; err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	// Invalidate list and search caches so new function appears in registry list
	if r.cache != nil {
		go func() {
			// Cache invalidation is best-effort; failures are logged but don't fail the operation
			_ = r.cache.InvalidateFunction(context.Background(), fn.ID.String())
			_ = r.cache.InvalidateSearchResults(context.Background())
			_ = r.cache.InvalidateListResults(context.Background())
		}()
	}

	return nil
}

// GetFunctionByID retrieves a function by ID
func (r *RegistryRepository) GetFunctionByID(id uuid.UUID) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.First(&fn, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by ID: %w", err)
	}

	return &fn, nil
}

// GetFunctionByAuthorName retrieves a function by author and name
func (r *RegistryRepository) GetFunctionByAuthorName(author, name string) (*RegistryFunction, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		var fn RegistryFunction
		if err := r.cache.GetJSON(context.Background(), cacheKey, &fn); err == nil {
			return &fn, nil
		}
		// Cache miss - continue to database
	}

	var fn RegistryFunction
	if err := r.db.Where("author = ? AND name = ?", author, name).First(&fn).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by author/name: %w", err)
	}

	// Cache the result if cache is available (best-effort, async)
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		go func() {
			_ = r.cache.SetJSON(context.Background(), cacheKey, fn)
		}()
	}

	return &fn, nil
}

// UpdateFunctionSettings updates the settings JSONB for a registry function (e.g. custom_domains).
func (r *RegistryRepository) UpdateFunctionSettings(id uuid.UUID, settings map[string]interface{}) error {
	if settings == nil {
		settings = make(map[string]interface{})
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", id).Update("settings", raw).Error; err != nil {
		return fmt.Errorf("failed to update function settings: %w", err)
	}
	if r.cache != nil {
		go func() {
			_ = r.cache.InvalidateFunction(context.Background(), id.String())
		}()
	}
	return nil
}

// UpdateFunctionLatestVersion updates the latest version pointer
func (r *RegistryRepository) UpdateFunctionLatestVersion(id uuid.UUID, version string) error {
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latest_version": version,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update latest version: %w", err)
	}

	// Invalidate cache for this function
	if r.cache != nil {
		go func() {
			if err := r.cache.InvalidateFunction(context.Background(), id.String()); err != nil {
				fmt.Printf("Failed to invalidate function cache after version update: %v\n", err)
			}
		}()
	}

	return nil
}

// DeleteFunction deletes a function from the registry by author and name
func (r *RegistryRepository) DeleteFunction(author, name string) error {
	// First get the function directly from DB without using cache to avoid cache issues
	var fn RegistryFunction
	if err := r.db.Where("author = ? AND name = ?", author, name).First(&fn).Error; err != nil {
		if err.Error() == "record not found" {
			return nil // Function doesn't exist, consider it deleted
		}
		return fmt.Errorf("failed to find function: %w", err)
	}

	// Delete related records first (versions, ratings, etc.)
	// Delete function versions
	if err := r.db.Where("function_id = ?", fn.ID).Delete(&RegistryFunctionVersion{}).Error; err != nil {
		return fmt.Errorf("failed to delete function versions: %w", err)
	}

	// Delete ratings
	if err := r.db.Where("function_id = ?", fn.ID).Delete(&RegistryFunctionRating{}).Error; err != nil {
		return fmt.Errorf("failed to delete function ratings: %w", err)
	}

	// Delete the function itself
	if err := r.db.Delete(&fn).Error; err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	// Invalidate cache (best-effort, async)
	if r.cache != nil {
		go func() {
			_ = r.cache.InvalidateFunction(context.Background(), fn.ID.String())
			_ = r.cache.InvalidateSearchResults(context.Background())
		}()
	}

	return nil
}

// DeleteAllFunctions deletes all functions from the registry (for testing/reset purposes)
func (r *RegistryRepository) DeleteAllFunctions() error {
	// Get table names from GORM to avoid hardcoding
	versionTable := (&RegistryFunctionVersion{}).TableName()
	ratingTable := (&RegistryFunctionRating{}).TableName()
	functionTable := (&RegistryFunction{}).TableName()

	// Use raw SQL to delete all records from related tables first (to avoid FK issues)
	if err := r.db.Exec(fmt.Sprintf("DELETE FROM %s", versionTable)).Error; err != nil {
		return fmt.Errorf("failed to delete all function versions: %w", err)
	}

	if err := r.db.Exec(fmt.Sprintf("DELETE FROM %s", ratingTable)).Error; err != nil {
		return fmt.Errorf("failed to delete all function ratings: %w", err)
	}

	if err := r.db.Exec(fmt.Sprintf("DELETE FROM %s", functionTable)).Error; err != nil {
		return fmt.Errorf("failed to delete all functions: %w", err)
	}

	// Invalidate all caches (best-effort, async)
	if r.cache != nil {
		go func() {
			_ = r.cache.InvalidateSearchResults(context.Background())
		}()
	}

	return nil
}

// IsFunctionVersionDeterministic checks if a function version is deterministic and cacheable
func (r *RegistryRepository) IsFunctionVersionDeterministic(functionID uuid.UUID, version string) (bool, time.Duration, error) {
	var functionVersion RegistryFunctionVersion
	if err := r.db.Where("function_id = ? AND version = ?", functionID, version).First(&functionVersion).Error; err != nil {
		return false, 0, fmt.Errorf("failed to get function version: %w", err)
	}

	// Check if function is marked as deterministic
	if !functionVersion.Deterministic {
		return false, 0, nil
	}

	// Check if side effects are none (safe for caching)
	if functionVersion.SideEffects != "none" {
		return false, 0, nil
	}

	// Return cache TTL from function configuration
	cacheTTL := time.Duration(functionVersion.CacheTTL) * time.Second
	return true, cacheTTL, nil
}

// SetWasmCompiled stores AOT-compiled module bytes for a function version.
// This allows the runtime to deserialize precompiled modules in ~0.1ms
// instead of recompiling on every cold start.
func (r *RegistryRepository) SetWasmCompiled(functionID uuid.UUID, version string, compiled []byte) error {
	if err := r.db.Model(&RegistryFunctionVersion{}).
		Where("function_id = ? AND version = ?", functionID, version).
		Update("wasm_compiled", compiled).Error; err != nil {
		return fmt.Errorf("failed to set wasm_compiled: %w", err)
	}
	return nil
}

// GetWasmCompiled retrieves AOT-compiled module bytes for a function version.
// Returns nil if no precompiled bytes are available.
func (r *RegistryRepository) GetWasmCompiled(functionID uuid.UUID, version string) ([]byte, error) {
	var fnVersion RegistryFunctionVersion
	if err := r.db.Select("wasm_compiled").
		Where("function_id = ? AND version = ?", functionID, version).
		First(&fnVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to get wasm_compiled: %w", err)
	}
	return fnVersion.WasmCompiled, nil
}
