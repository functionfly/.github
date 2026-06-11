package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ConnectorRepository struct {
	db *sql.DB
}

func NewConnectorRepository(db *sql.DB) *ConnectorRepository {
	return &ConnectorRepository{db: db}
}

func (r *ConnectorRepository) GetDB() *sql.DB {
	return r.db
}

func (r *ConnectorRepository) ListCatalog(ctx context.Context) ([]*Connector, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, slug, name, COALESCE(icon_url,''), COALESCE(oauth_url,''), COALESCE(scopes,''), is_active, created_at
		FROM connectors WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list connectors catalog: %w", err)
	}
	defer rows.Close()

	var connectors []*Connector
	for rows.Next() {
		c := &Connector{}
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.IconURL, &c.OAuthURL, &c.Scopes, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan connector: %w", err)
		}
		connectors = append(connectors, c)
	}
	return connectors, nil
}

func (r *ConnectorRepository) GetConnectorBySlug(ctx context.Context, slug string) (*Connector, error) {
	c := &Connector{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, name, COALESCE(icon_url,''), COALESCE(oauth_url,''), COALESCE(scopes,''), is_active, created_at
		FROM connectors WHERE slug = $1`, slug).Scan(
		&c.ID, &c.Slug, &c.Name, &c.IconURL, &c.OAuthURL, &c.Scopes, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get connector by slug: %w", err)
	}
	return c, nil
}

// GetConnectorByID retrieves a connector by its UUID.
func (r *ConnectorRepository) GetConnectorByID(ctx context.Context, id uuid.UUID) (*Connector, error) {
	c := &Connector{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, name, COALESCE(icon_url,''), COALESCE(oauth_url,''), COALESCE(scopes,''), is_active, created_at
		FROM connectors WHERE id = $1`, id).Scan(
		&c.ID, &c.Slug, &c.Name, &c.IconURL, &c.OAuthURL, &c.Scopes, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get connector by id: %w", err)
	}
	return c, nil
}

func (r *ConnectorRepository) GetUserConnectors(ctx context.Context, tenantID uuid.UUID) ([]*UserConnector, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT uc.id, uc.tenant_id, uc.connector_id, c.slug, c.name, COALESCE(c.icon_url,''),
		       COALESCE(uc.display_name,''), uc.status, uc.encrypted_credentials,
		       uc.last_sync_at, COALESCE(uc.sync_error,''), uc.created_at, uc.updated_at
		FROM user_connectors uc
		JOIN connectors c ON c.id = uc.connector_id
		WHERE uc.tenant_id = $1
		ORDER BY uc.created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list user connectors: %w", err)
	}
	defer rows.Close()

	var connectors []*UserConnector
	for rows.Next() {
		uc := &UserConnector{}
		var lastSync sql.NullTime
		if err := rows.Scan(
			&uc.ID, &uc.TenantID, &uc.ConnectorID, &uc.ConnectorSlug, &uc.ConnectorName, &uc.ConnectorIconURL,
			&uc.DisplayName, &uc.Status, &uc.EncryptedCredentials,
			&lastSync, &uc.SyncError, &uc.CreatedAt, &uc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user connector: %w", err)
		}
		if lastSync.Valid {
			uc.LastSyncAt = &lastSync.Time
		}
		connectors = append(connectors, uc)
	}
	return connectors, nil
}

func (r *ConnectorRepository) GetUserConnector(ctx context.Context, tenantID, connectorID uuid.UUID) (*UserConnector, error) {
	uc := &UserConnector{}
	var lastSync sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT uc.id, uc.tenant_id, uc.connector_id, c.slug, c.name, COALESCE(c.icon_url,''),
		       COALESCE(uc.display_name,''), uc.status, uc.encrypted_credentials,
		       uc.last_sync_at, COALESCE(uc.sync_error,''), uc.created_at, uc.updated_at
		FROM user_connectors uc
		JOIN connectors c ON c.id = uc.connector_id
		WHERE uc.tenant_id = $1 AND uc.id = $2`, tenantID, connectorID).Scan(
		&uc.ID, &uc.TenantID, &uc.ConnectorID, &uc.ConnectorSlug, &uc.ConnectorName, &uc.ConnectorIconURL,
		&uc.DisplayName, &uc.Status, &uc.EncryptedCredentials,
		&lastSync, &uc.SyncError, &uc.CreatedAt, &uc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user connector: %w", err)
	}
	if lastSync.Valid {
		uc.LastSyncAt = &lastSync.Time
	}
	return uc, nil
}

func (r *ConnectorRepository) GetUserConnectorBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*UserConnector, error) {
	uc := &UserConnector{}
	var lastSync sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT uc.id, uc.tenant_id, uc.connector_id, c.slug, c.name, COALESCE(c.icon_url,''),
		       COALESCE(uc.display_name,''), uc.status, uc.encrypted_credentials,
		       uc.last_sync_at, COALESCE(uc.sync_error,''), uc.created_at, uc.updated_at
		FROM user_connectors uc
		JOIN connectors c ON c.id = uc.connector_id
		WHERE uc.tenant_id = $1 AND c.slug = $2`, tenantID, slug).Scan(
		&uc.ID, &uc.TenantID, &uc.ConnectorID, &uc.ConnectorSlug, &uc.ConnectorName, &uc.ConnectorIconURL,
		&uc.DisplayName, &uc.Status, &uc.EncryptedCredentials,
		&lastSync, &uc.SyncError, &uc.CreatedAt, &uc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user connector by slug: %w", err)
	}
	if lastSync.Valid {
		uc.LastSyncAt = &lastSync.Time
	}
	return uc, nil
}

func (r *ConnectorRepository) CreateUserConnector(ctx context.Context, uc *UserConnector) (*UserConnector, error) {
	if uc.ID == uuid.Nil {
		uc.ID = uuid.New()
	}
	now := time.Now().UTC()
	uc.CreatedAt = now
	uc.UpdatedAt = now
	if uc.Status == "" {
		uc.Status = "active"
	}
	if uc.EncryptedCredentials == nil {
		uc.EncryptedCredentials = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_connectors (id, tenant_id, connector_id, display_name, status, encrypted_credentials, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uc.ID, uc.TenantID, uc.ConnectorID, uc.DisplayName, uc.Status, uc.EncryptedCredentials, uc.CreatedAt, uc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user connector: %w", err)
	}
	return uc, nil
}

func (r *ConnectorRepository) UpdateUserConnectorStatus(ctx context.Context, id uuid.UUID, status, syncError string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_connectors SET status = $1, sync_error = $2, updated_at = NOW()
		WHERE id = $3`, status, syncError, id)
	return err
}

func (r *ConnectorRepository) UpdateLastSyncAt(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_connectors SET last_sync_at = NOW(), status = 'active', sync_error = '', updated_at = NOW()
		WHERE id = $1`, id)
	return err
}

func (r *ConnectorRepository) DeleteUserConnector(ctx context.Context, tenantID, connectorID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM user_connectors WHERE tenant_id = $1 AND id = $2`, tenantID, connectorID)
	if err != nil {
		return fmt.Errorf("delete user connector: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user connector not found")
	}
	return nil
}

func (r *ConnectorRepository) CountUserConnectors(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_connectors WHERE tenant_id = $1`, tenantID).Scan(&count)
	return count, err
}

func (r *ConnectorRepository) GetActiveConnectorsForSync(ctx context.Context, frequency string) ([]*UserConnector, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT uc.id, uc.tenant_id, uc.connector_id, c.slug, c.name, COALESCE(c.icon_url,''),
		       COALESCE(uc.display_name,''), uc.status, uc.encrypted_credentials,
		       uc.last_sync_at, COALESCE(uc.sync_error,''), uc.created_at, uc.updated_at
		FROM user_connectors uc
		JOIN connectors c ON c.id = uc.connector_id
		WHERE uc.status = 'active'
		  AND (uc.last_sync_at IS NULL OR uc.last_sync_at < NOW() - $1::interval)
		ORDER BY uc.last_sync_at ASC NULLS FIRST
		LIMIT 100`, frequency)
	if err != nil {
		return nil, fmt.Errorf("get active connectors for sync: %w", err)
	}
	defer rows.Close()

	var connectors []*UserConnector
	for rows.Next() {
		uc := &UserConnector{}
		var lastSync sql.NullTime
		if err := rows.Scan(
			&uc.ID, &uc.TenantID, &uc.ConnectorID, &uc.ConnectorSlug, &uc.ConnectorName, &uc.ConnectorIconURL,
			&uc.DisplayName, &uc.Status, &uc.EncryptedCredentials,
			&lastSync, &uc.SyncError, &uc.CreatedAt, &uc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan connector for sync: %w", err)
		}
		if lastSync.Valid {
			uc.LastSyncAt = &lastSync.Time
		}
		connectors = append(connectors, uc)
	}
	return connectors, nil
}

func (r *ConnectorRepository) UpdateUserConnectorSettings(ctx context.Context, tenantID, connectorID uuid.UUID, enabled *bool, displayName *string, syncFrequency *string, autoSync *bool) error {
	// Check if connector exists
	uc, err := r.GetUserConnector(ctx, tenantID, connectorID)
	if err != nil {
		return fmt.Errorf("user connector not found")
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if enabled != nil {
		status := "active"
		if !*enabled {
			status = "disabled"
		}
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if displayName != nil && *displayName != "" {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, *displayName)
		argIndex++
	}

	if syncFrequency != nil {
		validFrequencies := map[string]bool{"5m": true, "15m": true, "1h": true, "6h": true, "24h": true}
		if validFrequencies[*syncFrequency] {
			setClauses = append(setClauses, fmt.Sprintf("sync_frequency = $%d", argIndex))
			args = append(args, *syncFrequency)
			argIndex++
		}
	}

	if autoSync != nil {
		setClauses = append(setClauses, fmt.Sprintf("auto_sync = $%d", argIndex))
		args = append(args, *autoSync)
		argIndex++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf("UPDATE user_connectors SET %s WHERE tenant_id = $%d AND id = $%d",
		strings.Join(setClauses, ", "), argIndex, argIndex+1)
	args = append(args, tenantID, connectorID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update user connector settings: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user connector not found")
	}

	_ = uc // silence unused variable warning
	return nil
}

func (r *ConnectorRepository) StoreOAuthState(ctx context.Context, state string, tenantID, connectorID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO connector_oauth_states (state, tenant_id, connector_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT DO NOTHING`, state, tenantID, connectorID)
	return err
}

// ConnectorOAuthState represents a stored OAuth state for CSRF validation
type ConnectorOAuthState struct {
	State       string
	TenantID    uuid.UUID
	ConnectorID uuid.UUID
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// GetOAuthState retrieves an OAuth state without consuming it (for validation)
func (r *ConnectorRepository) GetOAuthState(ctx context.Context, state string) (*ConnectorOAuthState, error) {
	s := &ConnectorOAuthState{}
	err := r.db.QueryRowContext(ctx, `
		SELECT state, tenant_id, connector_id, expires_at, created_at
		FROM connector_oauth_states
		WHERE state = $1 AND expires_at > NOW()`, state).Scan(
		&s.State, &s.TenantID, &s.ConnectorID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get oauth state: %w", err)
	}
	return s, nil
}

// ConsumeOAuthState retrieves and deletes an OAuth state (one-time use)
func (r *ConnectorRepository) ConsumeOAuthState(ctx context.Context, state string) (*ConnectorOAuthState, error) {
	s := &ConnectorOAuthState{}
	err := r.db.QueryRowContext(ctx, `
		DELETE FROM connector_oauth_states
		WHERE state = $1 AND expires_at > NOW()
		RETURNING state, tenant_id, connector_id, expires_at, created_at`, state).Scan(
		&s.State, &s.TenantID, &s.ConnectorID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("consume oauth state: %w", err)
	}
	return s, nil
}

// CleanupExpiredOAuthStates removes expired OAuth states
func (r *ConnectorRepository) CleanupExpiredOAuthStates(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM connector_oauth_states WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired oauth states: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}
