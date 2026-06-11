package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FunctionInputSchema represents JSON schema for input validation
type FunctionInputSchema struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionVersionID uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;uniqueIndex"`
	Schema            json.RawMessage `json:"schema" gorm:"type:jsonb;not null"`
	IsStrict          bool            `json:"is_strict" gorm:"default:false"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateFunction creates a new function in the registry
func (r *RegistryRepository) CreateFunction(ctx context.Context, fn *RegistryFunction) error {
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

	logrus.WithFields(logrus.Fields{
		"function_id": fn.ID,
		"author":      fn.Author,
		"name":        fn.Name,
	}).Debug("CreateFunction: about to insert")

	if err := r.db.WithContext(ctx).Create(fn).Error; err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function_id": fn.ID,
		"author":      fn.Author,
		"name":        fn.Name,
	}).Debug("CreateFunction: inserted successfully")

	// Invalidate list and search caches so new function appears in registry list
	if r.cache != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = r.cache.InvalidateFunction(ctx, fn.ID.String())
			_ = r.cache.InvalidateSearchResults(ctx)
			_ = r.cache.InvalidateListResults(ctx)
		}()
	}

	return nil
}

// GetFunctionByID retrieves a function by ID
func (r *RegistryRepository) GetFunctionByID(ctx context.Context, id uuid.UUID) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.WithContext(ctx).First(&fn, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by ID: %w", err)
	}

	return &fn, nil
}

// GetFunctionByAuthorName retrieves a function by author and name
func (r *RegistryRepository) GetFunctionByAuthorName(ctx context.Context, author, name string) (*RegistryFunction, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		var fn RegistryFunction
		if err := r.cache.GetJSON(ctx, cacheKey, &fn); err == nil {
			return &fn, nil
		}
		// Cache miss - continue to database
	}

	var fn RegistryFunction
	var err error
	err = r.db.WithContext(ctx).Where("author = ? AND name = ?", author, name).First(&fn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get function by author/name: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"author":      fn.Author,
		"name":        fn.Name,
		"function_id": fn.ID,
	}).Debug("GetFunctionByAuthorName found function")

	// Cache the result if cache is available (best-effort, async)
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = r.cache.SetJSON(ctx, cacheKey, fn)
		}()
	}

	return &fn, nil
}

// UpdateFunctionSettings updates the settings JSONB for a registry function (e.g. custom_domains).
func (r *RegistryRepository) UpdateFunctionSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error {
	if settings == nil {
		settings = make(map[string]interface{})
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := r.db.WithContext(ctx).Model(&RegistryFunction{}).Where("id = ?", id).Update("settings", raw).Error; err != nil {
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
func (r *RegistryRepository) UpdateFunctionLatestVersion(ctx context.Context, id uuid.UUID, version string) error {
	if err := r.db.WithContext(ctx).Model(&RegistryFunction{}).Where("id = ?", id).Updates(map[string]interface{}{
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
func (r *RegistryRepository) DeleteFunction(ctx context.Context, author, name string) error {
	// First get the function directly from DB without using cache to avoid cache issues
	var fn RegistryFunction
	if err := r.db.WithContext(ctx).Where("author = ? AND name = ?", author, name).First(&fn).Error; err != nil {
		if err.Error() == "record not found" {
			return nil // Function doesn't exist, consider it deleted
		}
		return fmt.Errorf("failed to find function: %w", err)
	}

	// Delete related records first (versions, ratings, etc.)
	// Delete function versions
	if err := r.db.WithContext(ctx).Where("function_id = ?", fn.ID).Delete(&RegistryFunctionVersion{}).Error; err != nil {
		return fmt.Errorf("failed to delete function versions: %w", err)
	}

	// Delete ratings
	if err := r.db.WithContext(ctx).Where("function_id = ?", fn.ID).Delete(&RegistryFunctionRating{}).Error; err != nil {
		return fmt.Errorf("failed to delete function ratings: %w", err)
	}

	// Delete the function itself
	if err := r.db.WithContext(ctx).Delete(&fn).Error; err != nil {
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
func (r *RegistryRepository) DeleteAllFunctions(ctx context.Context) error {
	// Get table names from GORM to avoid hardcoding
	versionTable := (&RegistryFunctionVersion{}).TableName()
	ratingTable := (&RegistryFunctionRating{}).TableName()
	functionTable := (&RegistryFunction{}).TableName()

	// Use raw SQL to delete all records from related tables first (to avoid FK issues)
	if err := r.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s", versionTable)).Error; err != nil {
		return fmt.Errorf("failed to delete all function versions: %w", err)
	}

	if err := r.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s", ratingTable)).Error; err != nil {
		return fmt.Errorf("failed to delete all function ratings: %w", err)
	}

	if err := r.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s", functionTable)).Error; err != nil {
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
func (r *RegistryRepository) IsFunctionVersionDeterministic(ctx context.Context, functionID uuid.UUID, version string) (bool, time.Duration, error) {
	var functionVersion RegistryFunctionVersion
	if err := r.db.WithContext(ctx).Where("function_id = ? AND version = ?", functionID, version).First(&functionVersion).Error; err != nil {
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
func (r *RegistryRepository) SetWasmCompiled(ctx context.Context, functionID uuid.UUID, version string, compiled []byte) error {
	if err := r.db.WithContext(ctx).Model(&RegistryFunctionVersion{}).
		Where("function_id = ? AND version = ?", functionID, version).
		Update("wasm_compiled", compiled).Error; err != nil {
		return fmt.Errorf("failed to set wasm_compiled: %w", err)
	}
	return nil
}

// GetWasmCompiled retrieves AOT-compiled module bytes for a function version.
// Returns nil if no precompiled bytes are available.
func (r *RegistryRepository) GetWasmCompiled(ctx context.Context, functionID uuid.UUID, version string) ([]byte, error) {
	var fnVersion RegistryFunctionVersion
	if err := r.db.WithContext(ctx).Select("wasm_compiled").
		Where("function_id = ? AND version = ?", functionID, version).
		First(&fnVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to get wasm_compiled: %w", err)
	}
	return fnVersion.WasmCompiled, nil
}

// CheckFunctionConflicts checks if any of the proposed function names already exist
// for the given tenant. Returns a map of function name -> existing function info.
func (r *RegistryRepository) CheckFunctionConflicts(ctx context.Context, tenantID uuid.UUID, functionNames []string) (map[string]RegistryFunction, error) {
	if len(functionNames) == 0 {
		return nil, nil
	}

	var conflicts []RegistryFunction
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name IN ?", tenantID, functionNames).
		Find(&conflicts).Error; err != nil {
		return nil, fmt.Errorf("failed to check function conflicts: %w", err)
	}

	result := make(map[string]RegistryFunction, len(conflicts))
	for _, fn := range conflicts {
		result[fn.Name] = fn
	}
	return result, nil
}

// GetFunctionByTenantAndName retrieves a function by tenant and name for conflict checking.
func (r *RegistryRepository) GetFunctionByTenantAndName(ctx context.Context, tenantID uuid.UUID, name string) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", tenantID, name).
		First(&fn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get function by tenant and name: %w", err)
	}
	return &fn, nil
}

// UpsertFunctionInputSchema creates or updates the input JSON schema for a function version.
// This auto-generates a permissive schema from the function's source code when none is provided.
func (r *RegistryRepository) UpsertFunctionInputSchema(ctx context.Context, functionVersionID uuid.UUID, schema json.RawMessage, isStrict bool) error {
	now := time.Now()

	// Try to find existing schema
	var existing FunctionInputSchema
	err := r.db.WithContext(ctx).Where("function_version_id = ?", functionVersionID).First(&existing).Error
	if err == nil {
		// Update existing
		existing.Schema = schema
		existing.IsStrict = isStrict
		existing.UpdatedAt = now
		return r.db.WithContext(ctx).Save(&existing).Error
	}

	// Check if it's a not-found error vs other error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check input schema: %w", err)
	}

	// Create new
	newSchema := FunctionInputSchema{
		FunctionVersionID: functionVersionID,
		Schema:            schema,
		IsStrict:          isStrict,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return r.db.WithContext(ctx).Create(&newSchema).Error
}

// GetFunctionInputSchema retrieves the input schema for a function version.
func (r *RegistryRepository) GetFunctionInputSchema(ctx context.Context, functionVersionID uuid.UUID) (*FunctionInputSchema, error) {
	var schema FunctionInputSchema
	err := r.db.WithContext(ctx).Where("function_version_id = ?", functionVersionID).First(&schema).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get input schema: %w", err)
	}
	return &schema, nil
}
