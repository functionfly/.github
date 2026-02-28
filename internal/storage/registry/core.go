package registry

import (
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistryRepository handles function registry database operations
type RegistryRepository struct {
	db     *gorm.DB
	cache  *cache.RegistryRedisCache
	keyGen *cache.RegistryCacheKey
}

// NewRegistryRepository creates a new registry repository
func NewRegistryRepository(db *gorm.DB, cacheClient *cache.RegistryRedisCache) *RegistryRepository {
	var keyGen *cache.RegistryCacheKey

	if cacheClient != nil {
		keyGen = cache.NewRegistryCacheKey()
	}

	return &RegistryRepository{
		db:     db,
		cache:  cacheClient,
		keyGen: keyGen,
	}
}

// GetRecentExecutions retrieves the most recent executions for a function
func (r *RegistryRepository) GetRecentExecutions(functionID uuid.UUID, limit int) ([]RegistryFunctionExecution, error) {
	var executions []RegistryFunctionExecution
	err := r.db.Where("function_id = ?", functionID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// GetRecentPublicExecutions retrieves the most recent public executions for a function
func (r *RegistryRepository) GetRecentPublicExecutions(functionID uuid.UUID, limit int) ([]RegistryExecutionPublic, error) {
	var executions []RegistryExecutionPublic
	err := r.db.Where("function_id = ?", functionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}
