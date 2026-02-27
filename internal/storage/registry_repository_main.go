package storage

import (
	"github.com/functionfly/functionfly/internal/cache"
	"gorm.io/gorm"
)

// RegistryRepository handles function registry database operations
type RegistryRepository struct {
	db     *gorm.DB
	cache  *cache.RegistryRedisCache
	keyGen *cache.RegistryCacheKey
}

// NewRegistryRepository creates a new registry repository
func NewRegistryRepository(db *gorm.DB, redisCache *cache.RegistryRedisCache) *RegistryRepository {
	var keyGen *cache.RegistryCacheKey
	if redisCache != nil {
		keyGen = cache.NewRegistryCacheKey()
	}

	return &RegistryRepository{
		db:     db,
		cache:  redisCache,
		keyGen: keyGen,
	}
}