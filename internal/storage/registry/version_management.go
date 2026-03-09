package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// VersionConflictStrategy defines how to handle re-publishing an existing version
type VersionConflictStrategy string

const (
	VersionConflictError     VersionConflictStrategy = "error"
	VersionConflictOverwrite VersionConflictStrategy = "overwrite"
	VersionConflictCreateNew VersionConflictStrategy = "create_new"
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

// GetFunctionVersionByID retrieves a function version by its primary key (used for verification status lookup).
func (r *RegistryRepository) GetFunctionVersionByID(id uuid.UUID) (*RegistryFunctionVersion, error) {
	var v RegistryFunctionVersion
	if err := r.db.First(&v, id).Error; err != nil {
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

func isVersionNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "record not found") || strings.Contains(s, "no rows in result set")
}

// UpsertFunctionVersion creates or updates a function version based on the conflict strategy.
func (r *RegistryRepository) UpsertFunctionVersion(v *RegistryFunctionVersion, strategy VersionConflictStrategy) (created bool, err error) {
	existing, lookupErr := r.GetFunctionVersion(v.FunctionID, v.Version)
	if lookupErr != nil && !isVersionNotFound(lookupErr) {
		return false, fmt.Errorf("failed to check existing version: %w", lookupErr)
	}

	if existing != nil {
		switch strategy {
		case VersionConflictError:
			return false, fmt.Errorf("version %s already exists for function %s; use conflict_strategy=overwrite to update it",
				v.Version, v.FunctionID)
		case VersionConflictOverwrite:
			v.ID = existing.ID
			v.PublishedAt = existing.PublishedAt
			v.UpdatedAt = time.Now()
			if err := r.db.Model(existing).Updates(map[string]interface{}{
				"manifest":      v.Manifest,
				"runtime":       v.Runtime,
				"timeout_ms":    v.TimeoutMs,
				"memory_mb":     v.MemoryMB,
				"deterministic": v.Deterministic,
				"side_effects":  v.SideEffects,
				"idempotent":    v.Idempotent,
				"cache_ttl":     v.CacheTTL,
				"capabilities":  v.Capabilities,
				"wasm_binary":   v.WasmBinary,
				"source_hash":   v.SourceHash,
				"bundle_size":   v.BundleSize,
				"source_code":   v.SourceCode,
				"updated_at":    v.UpdatedAt,
			}).Error; err != nil {
				return false, fmt.Errorf("failed to overwrite function version: %w", err)
			}
			if r.cache != nil && r.keyGen != nil {
				go func() {
					cacheKey := r.keyGen.FunctionVersion(v.FunctionID.String(), v.Version)
					_ = r.cache.Delete(context.Background(), cacheKey)
					latestKey := r.keyGen.FunctionVersion(v.FunctionID.String(), "latest")
					_ = r.cache.Delete(context.Background(), latestKey)
				}()
			}
			return false, nil
		case VersionConflictCreateNew:
			// Fall through to create new
		}
	}

	v.ID = uuid.New()
	v.PublishedAt = time.Now()
	if err := r.db.Create(v).Error; err != nil {
		return false, fmt.Errorf("failed to create function version: %w", err)
	}
	return true, nil
}

// ============================================
// Function Version Changelog Methods
// ============================================

// CreateFunctionVersionChangelog creates a new changelog entry for a function version
func (r *RegistryRepository) CreateFunctionVersionChangelog(c *FunctionVersionChangelog) error {
	c.ID = uuid.New()
	c.CreatedAt = time.Now()

	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("failed to create function version changelog: %w", err)
	}

	return nil
}

// GetFunctionVersionChangelogs retrieves all changelogs for a function
func (r *RegistryRepository) GetFunctionVersionChangelogs(functionID uuid.UUID) ([]FunctionVersionChangelog, error) {
	var changelogs []FunctionVersionChangelog
	if err := r.db.Where("function_id = ?", functionID).Order("created_at DESC").Find(&changelogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get function version changelogs: %w", err)
	}

	return changelogs, nil
}

// GetFunctionVersionChangelog retrieves a specific changelog by ID
func (r *RegistryRepository) GetFunctionVersionChangelog(id uuid.UUID) (*FunctionVersionChangelog, error) {
	var changelog FunctionVersionChangelog
	if err := r.db.First(&changelog, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function version changelog: %w", err)
	}

	return &changelog, nil
}

// GetChangelogByFunctionVersionID retrieves changelog for a specific function version
func (r *RegistryRepository) GetChangelogByFunctionVersionID(functionVersionID uuid.UUID) (*FunctionVersionChangelog, error) {
	var changelog FunctionVersionChangelog
	if err := r.db.Where("function_version_id = ?", functionVersionID).First(&changelog).Error; err != nil {
		return nil, fmt.Errorf("failed to get changelog by function version ID: %w", err)
	}

	return &changelog, nil
}

// GetChangelogsByVersion retrieves all changelogs for a specific version
func (r *RegistryRepository) GetChangelogsByVersion(functionID uuid.UUID, version string) ([]FunctionVersionChangelog, error) {
	var changelogs []FunctionVersionChangelog
	if err := r.db.Where("function_id = ? AND version = ?", functionID, version).Find(&changelogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get changelogs by version: %w", err)
	}

	return changelogs, nil
}

// GetChangelogsByCategory retrieves changelogs filtered by category
func (r *RegistryRepository) GetChangelogsByCategory(functionID uuid.UUID, category ChangeCategory) ([]FunctionVersionChangelog, error) {
	var changelogs []FunctionVersionChangelog
	if err := r.db.Where("function_id = ? AND category = ?", functionID, category).Order("created_at DESC").Find(&changelogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get changelogs by category: %w", err)
	}

	return changelogs, nil
}

// GetPreviousVersion retrieves the previous version for a function
func (r *RegistryRepository) GetPreviousVersion(functionID uuid.UUID, currentVersion string) (*RegistryFunctionVersion, error) {
	var version RegistryFunctionVersion
	// Get the version immediately before the current version based on published_at
	if err := r.db.Where("function_id = ? AND published_at < (SELECT published_at FROM registry_function_versions WHERE function_id = ? AND version = ?)",
		functionID, functionID, currentVersion).
		Order("published_at DESC").
		First(&version).Error; err != nil {
		return nil, fmt.Errorf("failed to get previous version: %w", err)
	}

	return &version, nil
}
