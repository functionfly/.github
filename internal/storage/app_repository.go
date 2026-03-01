package storage

import (
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
func (r *AppRepository) CreateApp(name, slug string, tenantID uuid.UUID) (*App, error) {
	app := &App{
		TenantID: tenantID,
		Name:     name,
		Slug:     slug,
	}

	if err := r.db.Create(app).Error; err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	return app, nil
}

// GetAppByID retrieves an app by ID
func (r *AppRepository) GetAppByID(id uuid.UUID) (*App, error) {
	var app App
	err := r.db.First(&app, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return &app, nil
}

// GetAppBySlug retrieves an app by slug
func (r *AppRepository) GetAppBySlug(slug string) (*App, error) {
	var app App
	err := r.db.Where("slug = ?", slug).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return &app, nil
}

// ListAppsByTenant lists all apps for a tenant.
// Runs inside a transaction with app.tenant_id set so RLS can use the index
// and avoid a full table scan when the session variable is unset.
func (r *AppRepository) ListAppsByTenant(tenantID uuid.UUID) ([]*App, error) {
	var apps []*App
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL app.tenant_id = ?", tenantID.String()).Error; err != nil {
			return err
		}
		return tx.Model(&App{}).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&apps).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list apps by tenant: %w", err)
	}
	return apps, nil
}