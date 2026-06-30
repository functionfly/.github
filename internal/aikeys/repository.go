package aikeys

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository manages BYOK AI provider keys.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new BYOK repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create stores a new BYOK key.
func (r *Repository) Create(ctx context.Context, key *types.AIProviderKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetByTenantAndProvider returns the active BYOK key for a tenant+provider.
func (r *Repository) GetByTenantAndProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*types.AIProviderKey, error) {
	var key types.AIProviderKey
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider = ?", tenantID, provider).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// ListByTenant returns all BYOK keys for a tenant (without encrypted fields).
func (r *Repository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]types.AIProviderKey, error) {
	var keys []types.AIProviderKey
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("provider ASC").
		Find(&keys).Error
	return keys, err
}

// Update updates specific fields on a BYOK key.
func (r *Repository) Update(ctx context.Context, id string, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&types.AIProviderKey{}).
		Where("id = ?", id).
		Updates(fields).Error
}

// Delete removes a BYOK key by ID and tenant.
func (r *Repository) Delete(ctx context.Context, id string, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&types.AIProviderKey{}).Error
}

// DeleteByTenantAndProvider removes a BYOK key by tenant+provider.
func (r *Repository) DeleteByTenantAndProvider(ctx context.Context, tenantID uuid.UUID, provider string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider = ?", tenantID, provider).
		Delete(&types.AIProviderKey{}).Error
}

// GetPendingHealthCheck returns keys that need a health check (oldest first).
func (r *Repository) GetPendingHealthCheck(ctx context.Context, batchSize int) ([]types.AIProviderKey, error) {
	var keys []types.AIProviderKey
	err := r.db.WithContext(ctx).
		Where("status IN ('active', 'degraded')").
		Where("last_health_check IS NULL OR last_health_check < ?", time.Now().Add(-6*time.Hour)).
		Order("last_health_check ASC NULLS FIRST").
		Limit(batchSize).
		Find(&keys).Error
	return keys, err
}

// UpdateHealthStatus updates the health status and message for a key.
func (r *Repository) UpdateHealthStatus(ctx context.Context, id string, status, message string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&types.AIProviderKey{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":            status,
			"health_message":    message,
			"last_health_check": now,
			"updated_at":        now,
		}).Error
}

// UpdateLastUsed updates the last_used_at timestamp.
func (r *Repository) UpdateLastUsed(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&types.AIProviderKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}
