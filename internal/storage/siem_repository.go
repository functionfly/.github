package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SIEMConfig represents a SIEM integration configuration
type SIEMConfig struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	TenantID        uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	Enabled         bool                   `json:"enabled" db:"enabled"`
	ExportFormat    string                 `json:"export_format" db:"export_format"`
	DestinationType string                 `json:"destination_type" db:"destination_type"`
	Config          map[string]interface{} `json:"config,omitempty" db:"config"`
	LastExportAt    *time.Time             `json:"last_export_at,omitempty" db:"last_export_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// SIEMExportLog represents a SIEM export operation log
type SIEMExportLog struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SIEMConfigID uuid.UUID `json:"siem_config_id" db:"siem_config_id"`
	Status       string    `json:"status" db:"status"`
	RecordsSent  int       `json:"records_sent" db:"records_sent"`
	ErrorMessage string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// SIEMRepository handles SIEM configuration and export log operations
type SIEMRepository struct {
	db *PostgresDB
}

// NewSIEMRepository creates a new SIEM repository
func NewSIEMRepository(db *PostgresDB) *SIEMRepository {
	return &SIEMRepository{db: db}
}

// Create inserts a new SIEM config
func (r *SIEMRepository) Create(ctx context.Context, config *SIEMConfig) error {
	config.ID = uuid.New()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if config.ExportFormat == "" {
		config.ExportFormat = "json"
	}
	if config.Config == nil {
		config.Config = make(map[string]interface{})
	}

	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	query := `
		INSERT INTO siem_configs (id, tenant_id, name, enabled, export_format, destination_type, config, last_export_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.db.ExecContext(ctx, query,
		config.ID, config.TenantID, config.Name, config.Enabled, config.ExportFormat,
		config.DestinationType, configJSON, config.LastExportAt, config.CreatedAt, config.UpdatedAt)

	return err
}

// GetByID retrieves a SIEM config by ID
func (r *SIEMRepository) GetByID(ctx context.Context, id uuid.UUID) (*SIEMConfig, error) {
	query := `
		SELECT id, tenant_id, name, enabled, export_format, destination_type, config, last_export_at, created_at, updated_at
		FROM siem_configs
		WHERE id = $1`

	var config SIEMConfig
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID, &config.TenantID, &config.Name, &config.Enabled, &config.ExportFormat,
		&config.DestinationType, &configJSON, &config.LastExportAt, &config.CreatedAt, &config.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SIEM config: %w", err)
	}

	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &config.Config)
	}

	return &config, nil
}

// GetByTenant retrieves all SIEM configs for a tenant
func (r *SIEMRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]*SIEMConfig, error) {
	query := `
		SELECT id, tenant_id, name, enabled, export_format, destination_type, config, last_export_at, created_at, updated_at
		FROM siem_configs
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query SIEM configs: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// GetEnabled retrieves all enabled SIEM configs (for background export)
func (r *SIEMRepository) GetEnabled(ctx context.Context) ([]*SIEMConfig, error) {
	query := `
		SELECT id, tenant_id, name, enabled, export_format, destination_type, config, last_export_at, created_at, updated_at
		FROM siem_configs
		WHERE enabled = true
		ORDER BY tenant_id, created_at`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled SIEM configs: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// Update updates an existing SIEM config
func (r *SIEMRepository) Update(ctx context.Context, config *SIEMConfig) error {
	config.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	query := `
		UPDATE siem_configs
		SET name = $1, enabled = $2, export_format = $3, destination_type = $4, config = $5, last_export_at = $6, updated_at = $7
		WHERE id = $8`

	result, err := r.db.ExecContext(ctx, query,
		config.Name, config.Enabled, config.ExportFormat, config.DestinationType,
		configJSON, config.LastExportAt, config.UpdatedAt, config.ID)

	if err != nil {
		return fmt.Errorf("failed to update SIEM config: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("SIEM config not found")
	}

	return nil
}

// Delete deletes a SIEM config
func (r *SIEMRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM siem_configs WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete SIEM config: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("SIEM config not found")
	}

	return nil
}

// UpdateLastExportAt updates the last export timestamp
func (r *SIEMRepository) UpdateLastExportAt(ctx context.Context, id uuid.UUID, exportTime time.Time) error {
	query := `UPDATE siem_configs SET last_export_at = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, exportTime, time.Now(), id)
	return err
}

// CreateExportLog inserts a new export log entry
func (r *SIEMRepository) CreateExportLog(ctx context.Context, log *SIEMExportLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO siem_export_logs (id, siem_config_id, status, records_sent, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.SIEMConfigID, log.Status, log.RecordsSent, log.ErrorMessage, log.CreatedAt)

	return err
}

// GetExportLogs retrieves export logs for a SIEM config
func (r *SIEMRepository) GetExportLogs(ctx context.Context, siemConfigID uuid.UUID, limit int) ([]*SIEMExportLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	query := `
		SELECT id, siem_config_id, status, records_sent, error_message, created_at
		FROM siem_export_logs
		WHERE siem_config_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, siemConfigID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query export logs: %w", err)
	}
	defer rows.Close()

	return r.scanExportLogs(rows)
}

// GetLatestExportLog retrieves the most recent export log for a config
func (r *SIEMRepository) GetLatestExportLog(ctx context.Context, siemConfigID uuid.UUID) (*SIEMExportLog, error) {
	query := `
		SELECT id, siem_config_id, status, records_sent, error_message, created_at
		FROM siem_export_logs
		WHERE siem_config_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var log SIEMExportLog

	err := r.db.QueryRowContext(ctx, query, siemConfigID).Scan(
		&log.ID, &log.SIEMConfigID, &log.Status, &log.RecordsSent, &log.ErrorMessage, &log.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest export log: %w", err)
	}

	return &log, nil
}

func (r *SIEMRepository) scanConfigs(rows *sql.Rows) ([]*SIEMConfig, error) {
	var configs []*SIEMConfig

	for rows.Next() {
		var config SIEMConfig
		var configJSON []byte

		err := rows.Scan(
			&config.ID, &config.TenantID, &config.Name, &config.Enabled, &config.ExportFormat,
			&config.DestinationType, &configJSON, &config.LastExportAt, &config.CreatedAt, &config.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan SIEM config: %w", err)
		}

		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &config.Config)
		}

		configs = append(configs, &config)
	}

	return configs, rows.Err()
}

func (r *SIEMRepository) scanExportLogs(rows *sql.Rows) ([]*SIEMExportLog, error) {
	var logs []*SIEMExportLog

	for rows.Next() {
		var log SIEMExportLog

		err := rows.Scan(
			&log.ID, &log.SIEMConfigID, &log.Status, &log.RecordsSent, &log.ErrorMessage, &log.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan export log: %w", err)
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}
