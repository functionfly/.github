package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AnalyticsSettings struct {
	ID          uuid.UUID              `json:"id"`
	ServiceName string                 `json:"service_name"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type AnalyticsRepository struct {
	db *PostgresDB
}

func NewAnalyticsRepository(db *PostgresDB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// CreateAnalyticsSettings creates new analytics settings for a service
func (r *AnalyticsRepository) CreateAnalyticsSettings(settings *AnalyticsSettings) (*AnalyticsSettings, error) {
	configJSON, err := json.Marshal(settings.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if settings.ID == uuid.Nil {
		settings.ID = uuid.New()
	}
	now := time.Now()
	settings.CreatedAt = now
	settings.UpdatedAt = now

	var returnedConfigJSON []byte
	err = r.db.QueryRow(`
		INSERT INTO analytics_settings (id, service_name, enabled, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (service_name) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			config = EXCLUDED.config,
			updated_at = EXCLUDED.updated_at
		RETURNING id, service_name, enabled, config, created_at, updated_at`,
		settings.ID, settings.ServiceName, settings.Enabled, configJSON, settings.CreatedAt, settings.UpdatedAt).Scan(
		&settings.ID, &settings.ServiceName, &settings.Enabled, &returnedConfigJSON,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create analytics settings: %w", err)
	}

	if err := json.Unmarshal(returnedConfigJSON, &settings.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal returned config: %w", err)
	}

	return settings, nil
}

// InitializeTenantAnalytics creates default analytics configuration for a tenant
func (r *AnalyticsRepository) InitializeTenantAnalytics(tenantID uuid.UUID) error {
	// Create tenant-specific analytics tracking entry
	settings := &AnalyticsSettings{
		ID:          uuid.New(),
		ServiceName: fmt.Sprintf("tenant_%s_analytics", tenantID.String()),
		Enabled:     true,
		Config: map[string]interface{}{
			"tenant_id":         tenantID.String(),
			"tracking_enabled":  true,
			"track_executions":  true,
			"track_errors":      true,
			"track_performance": true,
			"retention_days":    90,
		},
	}

	_, err := r.CreateAnalyticsSettings(settings)
	return err
}

func (r *AnalyticsRepository) GetAnalyticsSettings(serviceName string) (*AnalyticsSettings, error) {
	settings := &AnalyticsSettings{}
	var configJSON []byte
	var config map[string]interface{}

	err := r.db.QueryRow(`
		SELECT id, service_name, enabled, config, created_at, updated_at
		FROM analytics_settings
		WHERE service_name = $1`, serviceName).Scan(
		&settings.ID, &settings.ServiceName, &settings.Enabled, &configJSON,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analytics settings not found for service: %s", serviceName)
		}
		return nil, fmt.Errorf("failed to get analytics settings: %w", err)
	}

	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	settings.Config = config

	return settings, nil
}

func (r *AnalyticsRepository) GetAllAnalyticsSettings() ([]AnalyticsSettings, error) {
	rows, err := r.db.Query(`
		SELECT id, service_name, enabled, config, created_at, updated_at
		FROM analytics_settings
		ORDER BY service_name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query analytics settings: %w", err)
	}
	defer rows.Close()

	var settings []AnalyticsSettings
	for rows.Next() {
		s := AnalyticsSettings{}
		var configJSON []byte
		var config map[string]interface{}

		err := rows.Scan(&s.ID, &s.ServiceName, &s.Enabled, &configJSON, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analytics settings: %w", err)
		}

		if err := json.Unmarshal(configJSON, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		s.Config = config
		settings = append(settings, s)
	}

	return settings, nil
}

func (r *AnalyticsRepository) UpdateAnalyticsSettings(serviceName string, enabled bool, config map[string]interface{}) (*AnalyticsSettings, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	settings := &AnalyticsSettings{}
	var returnedConfigJSON []byte

	err = r.db.QueryRow(`
		UPDATE analytics_settings
		SET enabled = $2, config = $3, updated_at = NOW()
		WHERE service_name = $1
		RETURNING id, service_name, enabled, config, created_at, updated_at`,
		serviceName, enabled, configJSON).Scan(
		&settings.ID, &settings.ServiceName, &settings.Enabled, &returnedConfigJSON,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analytics settings not found for service: %s", serviceName)
		}
		return nil, fmt.Errorf("failed to update analytics settings: %w", err)
	}

	if err := json.Unmarshal(returnedConfigJSON, &settings.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal returned config: %w", err)
	}

	return settings, nil
}
