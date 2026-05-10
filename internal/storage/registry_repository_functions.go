package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFunction creates a new function in the registry
func (r *RegistryRepository) CreateFunction(fn *RegistryFunction) error {
	if fn.ID == uuid.Nil {
		fn.ID = uuid.New()
	}
	now := time.Now()
	if fn.CreatedAt.IsZero() {
		fn.CreatedAt = now
	}
	if fn.UpdatedAt.IsZero() {
		fn.UpdatedAt = now
	}

	if err := r.db.Create(fn).Error; err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	// Invalidate any related cache entries (though new functions won't have cache entries yet)
	if r.cache != nil {
		go func() {
			if err := r.cache.InvalidateSearchResults(context.Background()); err != nil {
				fmt.Printf("Failed to invalidate search cache after function creation: %v\n", err)
			}
		}()
	}

	return nil
}

// GetFunctionByID retrieves a function by ID with preloaded relationships
func (r *RegistryRepository) GetFunctionByID(id uuid.UUID) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.Preload("Versions").
		Preload("Versions.Signatures").
		Preload("Versions.MalwareScans").
		Preload("Versions.Approvals").
		Preload("Versions.Approvals.Comments").
		Preload("Versions.VerificationStatus").
		Preload("Rating").
		First(&fn, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by ID: %w", err)
	}

	return &fn, nil
}

// GetFunctionByIDMinimal retrieves a function by ID without preloaded relationships (for performance)
func (r *RegistryRepository) GetFunctionByIDMinimal(id uuid.UUID) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.First(&fn, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by ID: %w", err)
	}

	return &fn, nil
}

// GetFunctionByAuthorName retrieves a function by author and name with preloaded relationships
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
	if err := r.db.Preload("Versions").
		Preload("Versions.Signatures").
		Preload("Versions.MalwareScans").
		Preload("Versions.Approvals").
		Preload("Versions.Approvals.Comments").
		Preload("Versions.VerificationStatus").
		Preload("Rating").
		Where("author = ? AND name = ?", author, name).First(&fn).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by author/name: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, fn); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to cache function info: %v\n", err)
			}
		}()
	}

	return &fn, nil
}

// GetFunctionByAuthorNameMinimal retrieves a function by author and name without preloaded relationships (for performance)
func (r *RegistryRepository) GetFunctionByAuthorNameMinimal(author, name string) (*RegistryFunction, error) {
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

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, fn); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to cache function info: %v\n", err)
			}
		}()
	}

	return &fn, nil
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
