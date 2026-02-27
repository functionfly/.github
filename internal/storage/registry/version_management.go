package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFunctionVersion creates a new function version
func (r *RegistryRepository) CreateFunctionVersion(v *RegistryFunctionVersion) error {
	v.ID = uuid.New()
	v.PublishedAt = time.Now()

	if err := r.db.Create(v).Error; err != nil {
		return fmt.Errorf("failed to create function version: %w", err)
	}

	return nil
}

// GetFunctionVersion retrieves a specific version of a function
func (r *RegistryRepository) GetFunctionVersion(functionID uuid.UUID, version string) (*RegistryFunctionVersion, error) {
	var v RegistryFunctionVersion
	if err := r.db.Where("function_id = ? AND version = ?", functionID, version).First(&v).Error; err != nil {
		return nil, fmt.Errorf("failed to get function version: %w", err)
	}

	return &v, nil
}

// GetLatestFunctionVersion retrieves the latest version of a function
func (r *RegistryRepository) GetLatestFunctionVersion(functionID uuid.UUID) (*RegistryFunctionVersion, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionVersion(functionID.String(), "latest")
		var v RegistryFunctionVersion
		if err := r.cache.GetJSON(context.Background(), cacheKey, &v); err == nil {
			return &v, nil
		}
		// Cache miss - continue to database
	}

	var v RegistryFunctionVersion
	if err := r.db.Where("function_id = ?", functionID).Order("published_at DESC").First(&v).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest function version: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionVersion(functionID.String(), "latest")
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, v); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to cache latest function version: %v\n", err)
			}
		}()
	}

	return &v, nil
}

// ListFunctionVersions lists all versions of a function
func (r *RegistryRepository) ListFunctionVersions(functionID uuid.UUID) ([]RegistryFunctionVersion, error) {
	var versions []RegistryFunctionVersion
	if err := r.db.Where("function_id = ?", functionID).Order("published_at DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list function versions: %w", err)
	}

	return versions, nil
}