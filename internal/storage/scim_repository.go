package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SCIMConfig represents a SCIM configuration for a tenant
type SCIMConfig struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	Enabled    bool       `json:"enabled" gorm:"default:false"`
	IDPURL     *string    `json:"idp_url" gorm:"size:500"`
	IDPToken   *string    `json:"idp_token" gorm:"size:500"`
	SecretKey  []byte     `json:"secret_key" gorm:"type:bytea"`
	SyncGroups bool       `json:"sync_groups" gorm:"default:true"`
	SyncUsers  bool       `json:"sync_users" gorm:"default:true"`
	LastSyncAt *time.Time `json:"last_sync_at"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// SCIMSyncLog represents a SCIM sync operation log entry
type SCIMSyncLog struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Direction    string    `json:"direction" gorm:"size:20"`     // inbound, outbound
	ResourceType string    `json:"resource_type" gorm:"size:50"` // User, Group
	ResourceID   string    `json:"resource_id" gorm:"size:255"`
	Action       string    `json:"action" gorm:"size:20"` // create, update, delete
	Success      bool      `json:"success"`
	ErrorMessage *string   `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SCIMConfigRepository handles SCIM configuration database operations
type SCIMConfigRepository struct {
	db *PostgresDB
}

// NewSCIMConfigRepository creates a new SCIM config repository
func NewSCIMConfigRepository(db *PostgresDB) *SCIMConfigRepository {
	return &SCIMConfigRepository{db: db}
}

// Create creates a new SCIM configuration
func (r *SCIMConfigRepository) Create(config *SCIMConfig) error {
	_, err := r.db.Exec(`
		INSERT INTO scim_configs (id, tenant_id, enabled, idp_url, idp_token, secret_key, sync_groups, sync_users, last_sync_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		config.ID, config.TenantID, config.Enabled, config.IDPURL, config.IDPToken,
		config.SecretKey, config.SyncGroups, config.SyncUsers, config.LastSyncAt,
		config.CreatedAt, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create SCIM config: %w", err)
	}
	return nil
}

// GetByTenantID retrieves SCIM config by tenant ID
func (r *SCIMConfigRepository) GetByTenantID(tenantID uuid.UUID) (*SCIMConfig, error) {
	var config SCIMConfig

	err := r.db.QueryRow(`
		SELECT id, tenant_id, enabled, idp_url, idp_token, secret_key, sync_groups, sync_users, last_sync_at, created_at, updated_at
		FROM scim_configs WHERE tenant_id = $1`, tenantID).Scan(
		&config.ID, &config.TenantID, &config.Enabled, &config.IDPURL, &config.IDPToken,
		&config.SecretKey, &config.SyncGroups, &config.SyncUsers, &config.LastSyncAt,
		&config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("SCIM config not found for tenant")
		}
		return nil, fmt.Errorf("failed to get SCIM config: %w", err)
	}

	return &config, nil
}

// Update updates an existing SCIM configuration
func (r *SCIMConfigRepository) Update(config *SCIMConfig) error {
	_, err := r.db.Exec(`
		UPDATE scim_configs 
		SET enabled = $2, idp_url = $3, idp_token = $4, secret_key = $5, 
		    sync_groups = $6, sync_users = $7, last_sync_at = $8, updated_at = $9
		WHERE id = $1`,
		config.ID, config.Enabled, config.IDPURL, config.IDPToken, config.SecretKey,
		config.SyncGroups, config.SyncUsers, config.LastSyncAt, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update SCIM config: %w", err)
	}
	return nil
}

// Delete deletes a SCIM configuration
func (r *SCIMConfigRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM scim_configs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete SCIM config: %w", err)
	}
	return nil
}

// Enable enables SCIM for a tenant
func (r *SCIMConfigRepository) Enable(tenantID uuid.UUID, idpURL, idpToken string, secretKey []byte) error {
	// Check if config exists
	_, err := r.GetByTenantID(tenantID)
	if err != nil {
		// Create new config
		config := &SCIMConfig{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Enabled:    true,
			IDPURL:     &idpURL,
			IDPToken:   &idpToken,
			SecretKey:  secretKey,
			SyncGroups: true,
			SyncUsers:  true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		return r.Create(config)
	}

	// Update existing config
	_, err = r.db.Exec(`
		UPDATE scim_configs 
		SET enabled = true, idp_url = $2, idp_token = $3, secret_key = $4, updated_at = $5
		WHERE tenant_id = $1`,
		tenantID, idpURL, idpToken, secretKey, time.Now())
	if err != nil {
		return fmt.Errorf("failed to enable SCIM: %w", err)
	}
	return nil
}

// Disable disables SCIM for a tenant
func (r *SCIMConfigRepository) Disable(tenantID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE scim_configs 
		SET enabled = false, updated_at = $2
		WHERE tenant_id = $1`,
		tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to disable SCIM: %w", err)
	}
	return nil
}

// UpdateLastSyncAt updates the last sync timestamp
func (r *SCIMConfigRepository) UpdateLastSyncAt(tenantID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE scim_configs 
		SET last_sync_at = $2, updated_at = $3
		WHERE tenant_id = $1`,
		tenantID, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}
	return nil
}

// SCIMSyncLogRepository handles SCIM sync log database operations
type SCIMSyncLogRepository struct {
	db *PostgresDB
}

// NewSCIMSyncLogRepository creates a new SCIM sync log repository
func NewSCIMSyncLogRepository(db *PostgresDB) *SCIMSyncLogRepository {
	return &SCIMSyncLogRepository{db: db}
}

// Create creates a new SCIM sync log entry
func (r *SCIMSyncLogRepository) Create(log *SCIMSyncLog) error {
	_, err := r.db.Exec(`
		INSERT INTO scim_sync_log (id, tenant_id, direction, resource_type, resource_id, action, success, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		log.ID, log.TenantID, log.Direction, log.ResourceType, log.ResourceID,
		log.Action, log.Success, log.ErrorMessage, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create SCIM sync log: %w", err)
	}
	return nil
}

// GetByTenantID retrieves SCIM sync logs by tenant ID with pagination
func (r *SCIMSyncLogRepository) GetByTenantID(tenantID uuid.UUID, limit, offset int) ([]SCIMSyncLog, error) {
	rows, err := r.db.Query(`
		SELECT id, tenant_id, direction, resource_type, resource_id, action, success, error_message, created_at
		FROM scim_sync_log 
		WHERE tenant_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get SCIM sync logs: %w", err)
	}
	defer rows.Close()

	var logs []SCIMSyncLog
	for rows.Next() {
		var log SCIMSyncLog
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.Direction, &log.ResourceType, &log.ResourceID,
			&log.Action, &log.Success, &log.ErrorMessage, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SCIM sync log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// GetByResourceID retrieves SCIM sync logs for a specific resource
func (r *SCIMSyncLogRepository) GetByResourceID(tenantID uuid.UUID, resourceType, resourceID string) ([]SCIMSyncLog, error) {
	rows, err := r.db.Query(`
		SELECT id, tenant_id, direction, resource_type, resource_id, action, success, error_message, created_at
		FROM scim_sync_log 
		WHERE tenant_id = $1 AND resource_type = $2 AND resource_id = $3
		ORDER BY created_at DESC`, tenantID, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SCIM sync logs: %w", err)
	}
	defer rows.Close()

	var logs []SCIMSyncLog
	for rows.Next() {
		var log SCIMSyncLog
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.Direction, &log.ResourceType, &log.ResourceID,
			&log.Action, &log.Success, &log.ErrorMessage, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SCIM sync log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}
