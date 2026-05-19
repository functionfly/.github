package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExtensionStatus represents valid extension statuses
type ExtensionStatus string

const (
	ExtensionStatusEnabled  ExtensionStatus = "enabled"
	ExtensionStatusDisabled ExtensionStatus = "disabled"
	ExtensionStatusError    ExtensionStatus = "error"
)

// Extension represents a studio extension
type Extension struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description,omitempty"`
	AuthorName   string            `json:"author_name"`
	Category     string            `json:"category"`
	Status       ExtensionStatus   `json:"status"`
	Permissions  []string          `json:"permissions"`
	Hooks        []string          `json:"hooks"`
	SizeKB       int               `json:"size_kb"`
	Config       map[string]string `json:"config,omitempty"`
	InstalledAt  time.Time         `json:"installed_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	EnabledAt    *time.Time        `json:"enabled_at,omitempty"`
	ErrorMessage *string           `json:"error,omitempty"`
}

// Validate validates the extension data
func (e *Extension) Validate() error {
	if e.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	if e.Version == "" {
		return fmt.Errorf("version is required")
	}
	if len(e.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}
	return nil
}

// ExtensionRepository handles database operations for extensions
type ExtensionRepository struct {
	db *sql.DB
}

// NewExtensionRepository creates a new extension repository
func NewExtensionRepository(db *sql.DB) *ExtensionRepository {
	return &ExtensionRepository{db: db}
}

// ListExtensionsParams contains parameters for listing extensions
type ListExtensionsParams struct {
	TenantID  string
	Category  *string
	Status    *ExtensionStatus
	Limit     int
	Offset    int
}

// ListExtensions returns extensions filtered by tenant and optional filters
func (r *ExtensionRepository) ListExtensions(ctx context.Context, params ListExtensionsParams) ([]Extension, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, params.TenantID)
	argIdx++

	if params.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *params.Category)
		argIdx++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*params.Status))
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, version, description, author_name, category,
		       status, permissions, hooks, size_kb, config, installed_at, updated_at,
		       enabled_at, error_message
		FROM extension_registry
		WHERE %s
		ORDER BY installed_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list extensions: %w", err)
	}
	defer rows.Close()

	var extensions []Extension
	for rows.Next() {
		ext, err := scanExtension(rows)
		if err != nil {
			return nil, fmt.Errorf("scan extension: %w", err)
		}
		extensions = append(extensions, *ext)
	}

	return extensions, rows.Err()
}

// GetExtension returns a single extension by ID
func (r *ExtensionRepository) GetExtension(ctx context.Context, tenantID, extID string) (*Extension, error) {
	query := `
		SELECT id, tenant_id, name, version, description, author_name, category,
		       status, permissions, hooks, size_kb, config, installed_at, updated_at,
		       enabled_at, error_message
		FROM extension_registry
		WHERE tenant_id = $1 AND id = $2
	`
	var ext Extension
	var desc, authorName sql.NullString
	var permissions, hooks, config []byte
	var sizeKB sql.NullInt64
	var enabledAt sql.NullTime
	var errorMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, extID).Scan(
		&ext.ID, &ext.TenantID, &ext.Name, &ext.Version, &desc, &authorName,
		&ext.Category, &ext.Status, &permissions, &hooks, &sizeKB, &config,
		&ext.InstalledAt, &ext.UpdatedAt, &enabledAt, &errorMsg,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get extension: %w", err)
	}

	if desc.Valid {
		ext.Description = desc.String
	}
	if authorName.Valid {
		ext.AuthorName = authorName.String
	}
	if len(permissions) > 0 {
		_ = json.Unmarshal(permissions, &ext.Permissions)
	}
	if len(hooks) > 0 {
		_ = json.Unmarshal(hooks, &ext.Hooks)
	}
	if len(config) > 0 {
		_ = json.Unmarshal(config, &ext.Config)
	}
	if sizeKB.Valid {
		ext.SizeKB = int(sizeKB.Int64)
	}
	if enabledAt.Valid {
		ext.EnabledAt = &enabledAt.Time
	}
	if errorMsg.Valid {
		ext.ErrorMessage = &errorMsg.String
	}

	return &ext, nil
}

// InstallExtension creates a new extension installation
func (r *ExtensionRepository) InstallExtension(ctx context.Context, ext *Extension) error {
	if err := ext.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if ext.ID == "" {
		ext.ID = uuid.New().String()
	}

	permissionsRaw, _ := json.Marshal(ext.Permissions)
	hooksRaw, _ := json.Marshal(ext.Hooks)
	configRaw, _ := json.Marshal(ext.Config)
	now := time.Now()

	query := `
		INSERT INTO extension_registry (id, tenant_id, name, version, description, author_name, category, status, permissions, hooks, size_kb, config, installed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (tenant_id, name, version) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
		RETURNING installed_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		ext.ID, ext.TenantID, ext.Name, ext.Version, ext.Description, ext.AuthorName,
		ext.Category, ext.Status, permissionsRaw, hooksRaw, ext.SizeKB, configRaw, now, now,
	).Scan(&ext.InstalledAt, &ext.UpdatedAt)

	if err != nil {
		return fmt.Errorf("install extension: %w", err)
	}

	return nil
}

// UninstallExtension removes an extension
func (r *ExtensionRepository) UninstallExtension(ctx context.Context, tenantID, extID string) error {
	query := `DELETE FROM extension_registry WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, extID)
	if err != nil {
		return fmt.Errorf("uninstall extension: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("extension not found")
	}

	return nil
}

// EnableExtension enables an extension
func (r *ExtensionRepository) EnableExtension(ctx context.Context, tenantID, extID string) error {
	now := time.Now()
	query := `
		UPDATE extension_registry
		SET status = 'enabled', enabled_at = $1, updated_at = $2, error_message = NULL
		WHERE tenant_id = $3 AND id = $4
	`
	result, err := r.db.ExecContext(ctx, query, now, now, tenantID, extID)
	if err != nil {
		return fmt.Errorf("enable extension: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("extension not found")
	}

	return nil
}

// DisableExtension disables an extension
func (r *ExtensionRepository) DisableExtension(ctx context.Context, tenantID, extID string) error {
	now := time.Now()
	query := `
		UPDATE extension_registry
		SET status = 'disabled', updated_at = $1
		WHERE tenant_id = $2 AND id = $3
	`
	result, err := r.db.ExecContext(ctx, query, now, tenantID, extID)
	if err != nil {
		return fmt.Errorf("disable extension: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("extension not found")
	}

	return nil
}

// UpdateExtensionConfig updates extension configuration
func (r *ExtensionRepository) UpdateExtensionConfig(ctx context.Context, tenantID, extID string, config map[string]string) error {
	configRaw, _ := json.Marshal(config)
	now := time.Now()

	query := `
		UPDATE extension_registry
		SET config = $1, updated_at = $2
		WHERE tenant_id = $3 AND id = $4
	`
	result, err := r.db.ExecContext(ctx, query, configRaw, now, tenantID, extID)
	if err != nil {
		return fmt.Errorf("update extension config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("extension not found")
	}

	return nil
}

// ListHooks returns all enabled hooks for a tenant
func (r *ExtensionRepository) ListHooks(ctx context.Context, tenantID string) ([]ExtensionHook, error) {
	query := `
		SELECT id, name, description, extension_id, events, enabled
		FROM extension_registry
		WHERE tenant_id = $1 AND status = 'enabled' AND hooks != '[]'::jsonb
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list hooks: %w", err)
	}
	defer rows.Close()

	var hooks []ExtensionHook
	for rows.Next() {
		var hook ExtensionHook
		var events []byte
		err := rows.Scan(&hook.ID, &hook.Name, &hook.Description, &hook.ExtensionID, &events, &hook.Enabled)
		if err != nil {
			return nil, fmt.Errorf("scan hook: %w", err)
		}
		if len(events) > 0 {
			_ = json.Unmarshal(events, &hook.Events)
		}
		hooks = append(hooks, hook)
	}

	return hooks, rows.Err()
}

// ExtensionHook represents a hook exposed by an extension
type ExtensionHook struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	ExtensionID  string   `json:"extension_id"`
	Events       []string `json:"events"`
	Enabled      bool     `json:"enabled"`
}

func scanExtension(rows interface{ Scan(dst ...interface{}) error }) (*Extension, error) {
	var ext Extension
	var desc, authorName sql.NullString
	var permissions, hooks, config []byte
	var sizeKB sql.NullInt64
	var enabledAt sql.NullTime
	var errorMsg sql.NullString

	err := rows.Scan(
		&ext.ID, &ext.TenantID, &ext.Name, &ext.Version, &desc, &authorName,
		&ext.Category, &ext.Status, &permissions, &hooks, &sizeKB, &config,
		&ext.InstalledAt, &ext.UpdatedAt, &enabledAt, &errorMsg,
	)
	if err != nil {
		return nil, err
	}

	if desc.Valid {
		ext.Description = desc.String
	}
	if authorName.Valid {
		ext.AuthorName = authorName.String
	}
	if len(permissions) > 0 {
		_ = json.Unmarshal(permissions, &ext.Permissions)
	}
	if len(hooks) > 0 {
		_ = json.Unmarshal(hooks, &ext.Hooks)
	}
	if len(config) > 0 {
		_ = json.Unmarshal(config, &ext.Config)
	}
	if sizeKB.Valid {
		ext.SizeKB = int(sizeKB.Int64)
	}
	if enabledAt.Valid {
		ext.EnabledAt = &enabledAt.Time
	}
	if errorMsg.Valid {
		ext.ErrorMessage = &errorMsg.String
	}

	return &ext, nil
}