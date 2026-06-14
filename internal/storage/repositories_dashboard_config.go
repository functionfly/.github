package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: dashboard_configs (inline SQL).

// Dashboard configuration operations
func (db *PostgresDB) CreateDashboardConfig(ctx context.Context, config *DashboardConfig) (*DashboardConfig, error) {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	query := `
		INSERT INTO dashboard_configs (id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := db.ExecContext(ctx, query,
		config.ID, config.TenantID, config.UserID, config.ConfigType,
		config.Name, config.Config, config.IsActive, config.CreatedAt, config.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard config: %w", err)
	}

	return config, nil
}

func (db *PostgresDB) GetDashboardConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*DashboardConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at
		FROM dashboard_configs
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(context.Background(), query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dashboard configs: %w", err)
	}
	defer rows.Close()

	var configs []*DashboardConfig
	for rows.Next() {
		config := &DashboardConfig{}
		err := rows.Scan(
			&config.ID, &config.TenantID, &config.UserID, &config.ConfigType,
			&config.Name, &config.Config, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dashboard config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (db *PostgresDB) GetDashboardConfigsByUser(ctx context.Context, userID uuid.UUID) ([]*DashboardConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at
		FROM dashboard_configs
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user dashboard configs: %w", err)
	}
	defer rows.Close()

	var configs []*DashboardConfig
	for rows.Next() {
		config := &DashboardConfig{}
		err := rows.Scan(
			&config.ID, &config.TenantID, &config.UserID, &config.ConfigType,
			&config.Name, &config.Config, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dashboard config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (db *PostgresDB) GetDashboardConfigByID(ctx context.Context, configID uuid.UUID) (*DashboardConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at
		FROM dashboard_configs
		WHERE id = $1
	`

	config := &DashboardConfig{}
	err := db.QueryRowContext(context.Background(), query, configID).Scan(
		&config.ID, &config.TenantID, &config.UserID, &config.ConfigType,
		&config.Name, &config.Config, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dashboard config not found")
		}
		return nil, fmt.Errorf("failed to get dashboard config: %w", err)
	}

	return config, nil
}

func (db *PostgresDB) UpdateDashboardConfig(ctx context.Context, configID uuid.UUID, updates map[string]interface{}) (*DashboardConfig, error) {
	// Get current config
	config, err := db.GetDashboardConfigByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		config.Name = name
	}
	if configData, ok := updates["config"].(map[string]interface{}); ok {
		config.Config = configData
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		config.IsActive = isActive
	}

	config.UpdatedAt = time.Now()

	// Update in database
	query := `
		UPDATE dashboard_configs
		SET name = $2, config = $3, is_active = $4, updated_at = $5
		WHERE id = $1
	`

	_, err = db.ExecContext(ctx, query,
		configID, config.Name, config.Config, config.IsActive, config.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update dashboard config: %w", err)
	}

	return config, nil
}

func (db *PostgresDB) DeleteDashboardConfig(ctx context.Context, configID uuid.UUID) error {
	query := `DELETE FROM dashboard_configs WHERE id = $1`

	result, err := db.ExecContext(ctx, query, configID)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dashboard config not found")
	}

	return nil
}
