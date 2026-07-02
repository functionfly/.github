package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MaintenanceRepository handles platform maintenance mode data access
type MaintenanceRepository struct {
	db *gorm.DB

	// In-memory cache for GetEnabledMaintenance to avoid hitting DB every 2-3 seconds
	mu                 sync.RWMutex
	cachedEnabled      *types.PlatformMaintenance
	cachedEnabledExpiry time.Time
}

// NewMaintenanceRepository creates a new maintenance repository
func NewMaintenanceRepository(db *gorm.DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

// GetPlatformMaintenance gets the current platform maintenance configuration
func (r *MaintenanceRepository) GetPlatformMaintenance(ctx context.Context) (*types.PlatformMaintenance, error) {
	var maintenance types.PlatformMaintenance
	err := r.db.WithContext(ctx).Order("created_at DESC").First(&maintenance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get platform maintenance: %w", err)
	}
	return &maintenance, nil
}

// GetEnabledMaintenance gets the current enabled maintenance configuration.
// Results are cached in-memory for 30 seconds to reduce DB load from the
// maintenance middleware that polls every 2-3 seconds.
func (r *MaintenanceRepository) GetEnabledMaintenance(ctx context.Context) (*types.PlatformMaintenance, error) {
	r.mu.RLock()
	if r.cachedEnabled != nil && time.Now().Before(r.cachedEnabledExpiry) {
		result := r.cachedEnabled
		r.mu.RUnlock()
		return result, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if r.cachedEnabled != nil && time.Now().Before(r.cachedEnabledExpiry) {
		return r.cachedEnabled, nil
	}

	var maintenance types.PlatformMaintenance
	err := r.db.WithContext(ctx).Where("enabled = ?", true).First(&maintenance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.cachedEnabled = nil
			r.cachedEnabledExpiry = time.Now().Add(30 * time.Second)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get enabled maintenance: %w", err)
	}

	r.cachedEnabled = &maintenance
	r.cachedEnabledExpiry = time.Now().Add(30 * time.Second)
	return &maintenance, nil
}

// UpdatePlatformMaintenance updates the platform maintenance configuration
func (r *MaintenanceRepository) UpdatePlatformMaintenance(ctx context.Context, maintenance *types.PlatformMaintenance) error {
	maintenance.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(maintenance).Error
	if err != nil {
		return fmt.Errorf("failed to update platform maintenance: %w", err)
	}
	r.invalidateEnabledCache()
	return nil
}

// EnableMaintenance enables maintenance mode
func (r *MaintenanceRepository) EnableMaintenance(ctx context.Context, maintenance *types.PlatformMaintenance) error {
	maintenance.Enabled = true
	maintenance.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(maintenance).Error
	if err != nil {
		return fmt.Errorf("failed to enable maintenance: %w", err)
	}
	r.invalidateEnabledCache()
	return nil
}

// DisableMaintenance disables maintenance mode
func (r *MaintenanceRepository) DisableMaintenance(ctx context.Context, maintenance *types.PlatformMaintenance) error {
	maintenance.Enabled = false
	maintenance.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(maintenance).Error
	if err != nil {
		return fmt.Errorf("failed to disable maintenance: %w", err)
	}
	r.invalidateEnabledCache()
	return nil
}

// invalidateEnabledCache clears the cached enabled maintenance result
func (r *MaintenanceRepository) invalidateEnabledCache() {
	r.mu.Lock()
	r.cachedEnabled = nil
	r.cachedEnabledExpiry = time.Time{}
	r.mu.Unlock()
}

// GetMaintenanceTemplate gets a maintenance page template by name
func (r *MaintenanceRepository) GetMaintenanceTemplate(ctx context.Context, name string) (*types.MaintenancePageTemplate, error) {
	var template types.MaintenancePageTemplate
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default template
			return r.GetDefaultTemplate(ctx)
		}
		return nil, fmt.Errorf("failed to get maintenance template: %w", err)
	}
	return &template, nil
}

// GetDefaultTemplate gets the default maintenance page template
func (r *MaintenanceRepository) GetDefaultTemplate(ctx context.Context) (*types.MaintenancePageTemplate, error) {
	var template types.MaintenancePageTemplate
	err := r.db.WithContext(ctx).Where("is_default = ?", true).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return hardcoded default if none in DB
			return &types.MaintenancePageTemplate{
				Name:            "default",
				Title:           stringPtr("We'll be back soon!"),
				MessageHTML:     stringPtr("<p>We're performing scheduled maintenance. We'll be back shortly.</p>"),
				BackgroundColor: "#1a1a2e",
				TextColor:       "#ffffff",
				AccentColor:     "#4ecdc4",
				ShowContactInfo: true,
				ShowSocialLinks: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get default template: %w", err)
	}
	return &template, nil
}

// GetTemplateByID gets a template by ID
func (r *MaintenanceRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*types.MaintenancePageTemplate, error) {
	var template types.MaintenancePageTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &template, nil
}

// ListTemplates lists all maintenance page templates
func (r *MaintenanceRepository) ListTemplates(ctx context.Context) ([]types.MaintenancePageTemplate, error) {
	var templates []types.MaintenancePageTemplate
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}

// CreateTemplate creates a new maintenance page template
func (r *MaintenanceRepository) CreateTemplate(ctx context.Context, template *types.MaintenancePageTemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Create(template).Error
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	return nil
}

// UpdateTemplate updates a maintenance page template
func (r *MaintenanceRepository) UpdateTemplate(ctx context.Context, template *types.MaintenancePageTemplate) error {
	template.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(template).Error
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}
	return nil
}

// DeleteTemplate deletes a maintenance page template
func (r *MaintenanceRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&types.MaintenancePageTemplate{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// GetScheduledMaintenance gets upcoming scheduled maintenance
func (r *MaintenanceRepository) GetScheduledMaintenance(ctx context.Context) ([]types.PlatformMaintenance, error) {
	var maintenances []types.PlatformMaintenance
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("is_scheduled = ? AND scheduled_start > ?", true, now).
		Order("scheduled_start ASC").
		Find(&maintenances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled maintenance: %w", err)
	}
	return maintenances, nil
}

// CreateAuditLog creates an audit log entry for maintenance changes
func (r *MaintenanceRepository) CreateAuditLog(ctx context.Context, log *types.MaintenanceAuditLog) error {
	log.ChangedAt = time.Now()
	err := r.db.WithContext(ctx).Create(log).Error
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLog gets audit log entries for maintenance
func (r *MaintenanceRepository) GetAuditLog(ctx context.Context, limit int) ([]types.MaintenanceAuditLog, error) {
	var logs []types.MaintenanceAuditLog
	err := r.db.WithContext(ctx).
		Order("changed_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return logs, nil
}

// Helper function to convert maintenance config to JSON for audit
func maintenanceToJSON(maintenance *types.PlatformMaintenance) (string, error) {
	if maintenance == nil {
		return "null", nil
	}
	data, err := json.Marshal(maintenance)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// stringPtr is a helper to create string pointers
func stringPtr(s string) *string {
	return &s
}
