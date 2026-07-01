package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppRepository handles app-related database operations
type AppRepository struct {
	db *gorm.DB
}

// NewAppRepository creates a new app repository
func NewAppRepository(db *PostgresDB) *AppRepository {
	return &AppRepository{db: db.GORM}
}

// CreateApp creates a new app
func (r *AppRepository) CreateApp(ctx context.Context, name, slug string, tenantID uuid.UUID) (*App, error) {
	app := &App{
		TenantID: tenantID,
		Name:     name,
		Slug:     slug,
	}

	if err := r.db.WithContext(ctx).Create(app).Error; err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	return app, nil
}

// GetAppByID retrieves an app by ID
func (r *AppRepository) GetAppByID(ctx context.Context, id uuid.UUID) (*App, error) {
	var app App
	err := r.db.WithContext(ctx).First(&app, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return &app, nil
}

// GetAppBySlug retrieves an app by slug
func (r *AppRepository) GetAppBySlug(ctx context.Context, slug string) (*App, error) {
	var app App
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return &app, nil
}

// GetAppBySlugAndTenant retrieves an app by slug within a tenant.
func (r *AppRepository) GetAppBySlugAndTenant(ctx context.Context, slug string, tenantID uuid.UUID) (*App, error) {
	var app App
	err := r.db.WithContext(ctx).Where("slug = ? AND tenant_id = ?", slug, tenantID).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return &app, nil
}

// ListAppsByTenant lists all apps for a tenant.
// The apps table does not use RLS; a simple filtered query is used.
func (r *AppRepository) ListAppsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*App, error) {
	var apps []*App
	err := r.db.WithContext(ctx).Model(&App{}).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list apps by tenant: %w", err)
	}
	return apps, nil
}

// UpdateApp updates an app's mutable fields (name).
func (r *AppRepository) UpdateApp(ctx context.Context, id uuid.UUID, name string) (*App, error) {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if len(updates) == 0 {
		return r.GetAppByID(ctx, id)
	}
	if err := r.db.WithContext(ctx).Model(&App{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}
	return r.GetAppByID(ctx, id)
}

// DeleteApp deletes an app by ID. Cascading deletes for backends/deployments
// are handled by FK constraints.
func (r *AppRepository) DeleteApp(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&App{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}
	return nil
}
