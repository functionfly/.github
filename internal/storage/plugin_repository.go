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

type PluginType string

const (
	PluginTypeUI            PluginType = "ui"
	PluginTypeGraph         PluginType = "graph"
	PluginTypeAITool        PluginType = "ai_tool"
	PluginTypeRuntime       PluginType = "runtime"
	PluginTypeInfrastructure PluginType = "infrastructure"
	PluginTypeMarketplace   PluginType = "marketplace"
)

type PluginStatus string

const (
	PluginStatusEnabled  PluginStatus = "enabled"
	PluginStatusDisabled PluginStatus = "disabled"
	PluginStatusError    PluginStatus = "error"
	PluginStatusPaused    PluginStatus = "paused"
)

type SandboxTier string

const (
	SandboxTierWASM      SandboxTier = "wasm"
	SandboxTierWorker    SandboxTier = "worker"
	SandboxTierMicroVM   SandboxTier = "microvm"
	SandboxTierEnterprise SandboxTier = "enterprise"
)

type Plugin struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"tenant_id"`
	Manifest      map[string]interface{} `json:"manifest"`
	PluginType    PluginType        `json:"plugin_type"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Description   string            `json:"description,omitempty"`
	AuthorName    string            `json:"author_name"`
	AuthorEmail   string            `json:"author_email,omitempty"`
	AuthorWebsite string            `json:"author_website,omitempty"`
	Category      string            `json:"category"`
	Status        PluginStatus      `json:"status"`
	IconURL       string            `json:"icon_url,omitempty"`
	HomepageURL   string            `json:"homepage_url,omitempty"`
	RepositoryURL string            `json:"repository_url,omitempty"`
	License       string            `json:"license,omitempty"`
	SizeBytes     int               `json:"size_bytes"`
	Signature     string            `json:"signature,omitempty"`
	Verified      bool              `json:"verified"`
	Config        map[string]string `json:"config,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	InstalledAt   time.Time         `json:"installed_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	EnabledAt     *time.Time        `json:"enabled_at,omitempty"`
	ErrorMessage  *string           `json:"error,omitempty"`
}

type PluginVersion struct {
	ID         string    `json:"id"`
	PluginID   string    `json:"plugin_id"`
	Version    string    `json:"version"`
	Changelog  string    `json:"changelog,omitempty"`
	Manifest   map[string]interface{} `json:"manifest"`
	SizeBytes  int       `json:"size_bytes"`
	Signature  string    `json:"signature,omitempty"`
	ReleaseAt  time.Time `json:"release_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type PluginPermission struct {
	ID              string     `json:"id"`
	PluginID        string     `json:"plugin_id"`
	PermissionType  string     `json:"permission_type"`
	PermissionAction string    `json:"permission_action"`
	Resource        string     `json:"resource,omitempty"`
	Granted         bool       `json:"granted"`
	GrantedAt       *time.Time `json:"granted_at,omitempty"`
	GrantedBy       *string    `json:"granted_by,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type PluginSandbox struct {
	ID                string     `json:"id"`
	PluginID          string     `json:"plugin_id"`
	Tier              SandboxTier `json:"tier"`
	CPULimit          float64    `json:"cpu_limit"`
	MemoryLimitMB     int        `json:"memory_limit_mb"`
	TimeoutSeconds    int        `json:"timeout_seconds"`
	NetworkIsolated   bool       `json:"network_isolated"`
	FilesystemScope   string     `json:"filesystem_scope"`
	MaxInstances      int        `json:"max_instances"`
	EnvVars           map[string]string `json:"env_vars,omitempty"`
	AllowedDomains    []string   `json:"allowed_domains,omitempty"`
	BlockedDomains    []string   `json:"blocked_domains,omitempty"`
	RateLimitRPM      *int       `json:"rate_limit_rpm,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type PluginAnalytics struct {
	ID                string    `json:"id"`
	PluginID          string    `json:"plugin_id"`
	TenantID          string    `json:"tenant_id"`
	EventType         string    `json:"event_type"`
	ExecutionsCount   int       `json:"executions_count"`
	ErrorsCount       int       `json:"errors_count"`
	TotalLatencyMs    int64     `json:"total_latency_ms"`
	CPUUsageSeconds   float64   `json:"cpu_usage_seconds"`
	MemoryUsageMBAvg  float64   `json:"memory_usage_mb_avg"`
	NetworkBytes      int64     `json:"network_bytes"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type PluginRepository struct {
	db *sql.DB
}

func NewPluginRepository(db *sql.DB) *PluginRepository {
	return &PluginRepository{db: db}
}

type ListPluginsParams struct {
	TenantID   string
	PluginType *PluginType
	Status     *PluginStatus
	Category   *string
	Search     *string
	Limit      int
	Offset     int
}

func (r *PluginRepository) List(ctx context.Context, params ListPluginsParams) ([]Plugin, error) {
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

	if params.PluginType != nil {
		conditions = append(conditions, fmt.Sprintf("plugin_type = $%d", argIdx))
		args = append(args, string(*params.PluginType))
		argIdx++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*params.Status))
		argIdx++
	}

	if params.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *params.Category)
		argIdx++
	}

	if params.Search != nil && *params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*params.Search+"%")
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, manifest, plugin_type, name, version, description, author_name,
		       author_email, author_website, category, status, icon_url, homepage_url, repository_url,
		       license, size_bytes, signature, verified, config, metadata, installed_at, updated_at,
		       enabled_at, error_message
		FROM plugins
		WHERE %s
		ORDER BY installed_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	defer rows.Close()

	var plugins []Plugin
	for rows.Next() {
		plugin, err := scanPlugin(rows)
		if err != nil {
			return nil, fmt.Errorf("scan plugin: %w", err)
		}
		plugins = append(plugins, *plugin)
	}

	return plugins, rows.Err()
}

func (r *PluginRepository) Get(ctx context.Context, tenantID, pluginID string) (*Plugin, error) {
	query := `
		SELECT id, tenant_id, manifest, plugin_type, name, version, description, author_name,
		       author_email, author_website, category, status, icon_url, homepage_url, repository_url,
		       license, size_bytes, signature, verified, config, metadata, installed_at, updated_at,
		       enabled_at, error_message
		FROM plugins
		WHERE tenant_id = $1 AND id = $2
	`
	var plugin Plugin
	var manifest, config, metadata []byte
	var desc, authorEmail, authorWebsite, iconURL, homepageURL, repositoryURL, license, signature sql.NullString
	var sizeBytes sql.NullInt64
	var enabledAt sql.NullTime
	var errorMsg sql.NullString
	var verified sql.NullBool

	err := r.db.QueryRowContext(ctx, query, tenantID, pluginID).Scan(
		&plugin.ID, &plugin.TenantID, &manifest, &plugin.PluginType, &plugin.Name, &plugin.Version,
		&desc, &plugin.AuthorName, &authorEmail, &authorWebsite, &plugin.Category, &plugin.Status,
		&iconURL, &homepageURL, &repositoryURL, &license, &sizeBytes, &signature, &verified,
		&config, &metadata, &plugin.InstalledAt, &plugin.UpdatedAt, &enabledAt, &errorMsg,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get plugin: %w", err)
	}

	if desc.Valid {
		plugin.Description = desc.String
	}
	if authorEmail.Valid {
		plugin.AuthorEmail = authorEmail.String
	}
	if authorWebsite.Valid {
		plugin.AuthorWebsite = authorWebsite.String
	}
	if iconURL.Valid {
		plugin.IconURL = iconURL.String
	}
	if homepageURL.Valid {
		plugin.HomepageURL = homepageURL.String
	}
	if repositoryURL.Valid {
		plugin.RepositoryURL = repositoryURL.String
	}
	if license.Valid {
		plugin.License = license.String
	}
	if signature.Valid {
		plugin.Signature = signature.String
	}
	if sizeBytes.Valid {
		plugin.SizeBytes = int(sizeBytes.Int64)
	}
	if verified.Valid {
		plugin.Verified = verified.Bool
	}
	if enabledAt.Valid {
		plugin.EnabledAt = &enabledAt.Time
	}
	if errorMsg.Valid {
		plugin.ErrorMessage = &errorMsg.String
	}
	if len(manifest) > 0 {
		_ = json.Unmarshal(manifest, &plugin.Manifest)
	}
	if len(config) > 0 {
		_ = json.Unmarshal(config, &plugin.Config)
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &plugin.Metadata)
	}

	return &plugin, nil
}

func (r *PluginRepository) Create(ctx context.Context, plugin *Plugin) error {
	if plugin.ID == "" {
		plugin.ID = uuid.New().String()
	}

	manifestRaw, _ := json.Marshal(plugin.Manifest)
	configRaw, _ := json.Marshal(plugin.Config)
	metadataRaw, _ := json.Marshal(plugin.Metadata)
	now := time.Now()

	query := `
		INSERT INTO plugins (id, tenant_id, manifest, plugin_type, name, version, description, author_name,
		                      author_email, author_website, category, status, icon_url, homepage_url,
		                      repository_url, license, size_bytes, signature, verified, config, metadata,
		                      installed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		ON CONFLICT (tenant_id, name) DO UPDATE SET
			manifest = EXCLUDED.manifest,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
		RETURNING installed_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		plugin.ID, plugin.TenantID, manifestRaw, plugin.PluginType, plugin.Name, plugin.Version,
		plugin.Description, plugin.AuthorName, plugin.AuthorEmail, plugin.AuthorWebsite,
		plugin.Category, plugin.Status, plugin.IconURL, plugin.HomepageURL, plugin.RepositoryURL,
		plugin.License, plugin.SizeBytes, plugin.Signature, plugin.Verified, configRaw, metadataRaw,
		now, now,
	).Scan(&plugin.InstalledAt, &plugin.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create plugin: %w", err)
	}

	return nil
}

func (r *PluginRepository) Update(ctx context.Context, plugin *Plugin) error {
	manifestRaw, _ := json.Marshal(plugin.Manifest)
	configRaw, _ := json.Marshal(plugin.Config)
	metadataRaw, _ := json.Marshal(plugin.Metadata)
	now := time.Now()

	query := `
		UPDATE plugins SET
			manifest = $1, plugin_type = $2, name = $3, version = $4, description = $5,
			author_name = $6, author_email = $7, author_website = $8, category = $9, status = $10,
			icon_url = $11, homepage_url = $12, repository_url = $13, license = $14,
			size_bytes = $15, signature = $16, verified = $17, config = $18, metadata = $19,
			updated_at = $20, error_message = $21
		WHERE tenant_id = $22 AND id = $23
	`

	result, err := r.db.ExecContext(ctx, query,
		manifestRaw, plugin.PluginType, plugin.Name, plugin.Version, plugin.Description,
		plugin.AuthorName, plugin.AuthorEmail, plugin.AuthorWebsite, plugin.Category, plugin.Status,
		plugin.IconURL, plugin.HomepageURL, plugin.RepositoryURL, plugin.License,
		plugin.SizeBytes, plugin.Signature, plugin.Verified, configRaw, metadataRaw,
		now, plugin.ErrorMessage, plugin.TenantID, plugin.ID,
	)
	if err != nil {
		return fmt.Errorf("update plugin: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	plugin.UpdatedAt = now
	return nil
}

func (r *PluginRepository) Delete(ctx context.Context, tenantID, pluginID string) error {
	query := `DELETE FROM plugins WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, pluginID)
	if err != nil {
		return fmt.Errorf("delete plugin: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

func (r *PluginRepository) SetStatus(ctx context.Context, tenantID, pluginID string, status PluginStatus) error {
	now := time.Now()
	var query string
	var args []interface{}

	switch status {
	case PluginStatusEnabled:
		query = `
			UPDATE plugins SET status = $1, enabled_at = $2, updated_at = $3, error_message = NULL
			WHERE tenant_id = $4 AND id = $5
		`
		args = []interface{}{status, now, now, tenantID, pluginID}
	case PluginStatusDisabled:
		query = `
			UPDATE plugins SET status = $1, updated_at = $2
			WHERE tenant_id = $3 AND id = $4
		`
		args = []interface{}{status, now, tenantID, pluginID}
	case PluginStatusPaused:
		query = `
			UPDATE plugins SET status = $1, updated_at = $2
			WHERE tenant_id = $3 AND id = $4
		`
		args = []interface{}{status, now, tenantID, pluginID}
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("set plugin status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

func (r *PluginRepository) GetEnabledByType(ctx context.Context, tenantID string, pluginType PluginType) (*Plugin, error) {
	query := `
		SELECT id, tenant_id, manifest, plugin_type, name, version, description, author_name,
		       author_email, author_website, category, status, icon_url, homepage_url, repository_url,
		       license, size_bytes, signature, verified, config, metadata, installed_at, updated_at,
		       enabled_at, error_message
		FROM plugins
		WHERE tenant_id = $1 AND plugin_type = $2 AND status = $3
		LIMIT 1
	`
	var plugin Plugin
	var manifest, config, metadata []byte
	var desc, authorEmail, authorWebsite, iconURL, homepageURL, repositoryURL, license, signature sql.NullString
	var sizeBytes sql.NullInt64
	var enabledAt sql.NullTime
	var errorMsg sql.NullString
	var verified sql.NullBool

	err := r.db.QueryRowContext(ctx, query, tenantID, string(pluginType), PluginStatusEnabled).Scan(
		&plugin.ID, &plugin.TenantID, &manifest, &plugin.PluginType, &plugin.Name, &plugin.Version,
		&desc, &plugin.AuthorName, &authorEmail, &authorWebsite, &plugin.Category, &plugin.Status,
		&iconURL, &homepageURL, &repositoryURL, &license, &sizeBytes, &signature, &verified,
		&config, &metadata, &plugin.InstalledAt, &plugin.UpdatedAt, &enabledAt, &errorMsg,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get enabled plugin by type: %w", err)
	}

	if desc.Valid {
		plugin.Description = desc.String
	}
	if authorEmail.Valid {
		plugin.AuthorEmail = authorEmail.String
	}
	if authorWebsite.Valid {
		plugin.AuthorWebsite = authorWebsite.String
	}
	if iconURL.Valid {
		plugin.IconURL = iconURL.String
	}
	if homepageURL.Valid {
		plugin.HomepageURL = homepageURL.String
	}
	if repositoryURL.Valid {
		plugin.RepositoryURL = repositoryURL.String
	}
	if license.Valid {
		plugin.License = license.String
	}
	if signature.Valid {
		plugin.Signature = signature.String
	}
	if sizeBytes.Valid {
		plugin.SizeBytes = int(sizeBytes.Int64)
	}
	if verified.Valid {
		plugin.Verified = verified.Bool
	}
	if enabledAt.Valid {
		plugin.EnabledAt = &enabledAt.Time
	}
	if errorMsg.Valid {
		plugin.ErrorMessage = &errorMsg.String
	}
	if len(manifest) > 0 {
		_ = json.Unmarshal(manifest, &plugin.Manifest)
	}
	if len(config) > 0 {
		_ = json.Unmarshal(config, &plugin.Config)
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &plugin.Metadata)
	}

	return &plugin, nil
}

func (r *PluginRepository) SetError(ctx context.Context, tenantID, pluginID string, errMsg string) error {
	now := time.Now()
	query := `
		UPDATE plugins SET status = $1, error_message = $2, updated_at = $3
		WHERE tenant_id = $4 AND id = $5
	`
	result, err := r.db.ExecContext(ctx, query, PluginStatusError, errMsg, now, tenantID, pluginID)
	if err != nil {
		return fmt.Errorf("set plugin error: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

func (r *PluginRepository) UpdateConfig(ctx context.Context, tenantID, pluginID string, config map[string]string) error {
	configRaw, _ := json.Marshal(config)
	now := time.Now()

	query := `
		UPDATE plugins SET config = $1, updated_at = $2
		WHERE tenant_id = $3 AND id = $4
	`
	result, err := r.db.ExecContext(ctx, query, configRaw, now, tenantID, pluginID)
	if err != nil {
		return fmt.Errorf("update plugin config: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

func (r *PluginRepository) GetSandbox(ctx context.Context, pluginID string) (*PluginSandbox, error) {
	query := `
		SELECT id, plugin_id, tier, cpu_limit, memory_limit_mb, timeout_seconds,
		       network_isolated, filesystem_scope, max_instances, env_vars,
		       allowed_domains, blocked_domains, rate_limit_rpm, created_at, updated_at
		FROM plugin_sandboxes
		WHERE plugin_id = $1
	`
	var sandbox PluginSandbox
	var envVars []byte
	var allowedDomains, blockedDomains []string
	var rateLimitRPM sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, pluginID).Scan(
		&sandbox.ID, &sandbox.PluginID, &sandbox.Tier, &sandbox.CPULimit,
		&sandbox.MemoryLimitMB, &sandbox.TimeoutSeconds, &sandbox.NetworkIsolated,
		&sandbox.FilesystemScope, &sandbox.MaxInstances, &envVars,
		&allowedDomains, &blockedDomains, &rateLimitRPM, &sandbox.CreatedAt, &sandbox.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sandbox: %w", err)
	}

	sandbox.AllowedDomains = allowedDomains
	sandbox.BlockedDomains = blockedDomains
	if len(envVars) > 0 {
		_ = json.Unmarshal(envVars, &sandbox.EnvVars)
	}
	if rateLimitRPM.Valid {
		rpm := int(rateLimitRPM.Int64)
		sandbox.RateLimitRPM = &rpm
	}

	return &sandbox, nil
}

func (r *PluginRepository) UpsertSandbox(ctx context.Context, sandbox *PluginSandbox) error {
	if sandbox.ID == "" {
		sandbox.ID = uuid.New().String()
	}

	envVarsRaw, _ := json.Marshal(sandbox.EnvVars)
	now := time.Now()

	query := `
		INSERT INTO plugin_sandboxes (id, plugin_id, tier, cpu_limit, memory_limit_mb, timeout_seconds,
		                               network_isolated, filesystem_scope, max_instances, env_vars,
		                               allowed_domains, blocked_domains, rate_limit_rpm, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (plugin_id) DO UPDATE SET
			tier = EXCLUDED.tier, cpu_limit = EXCLUDED.cpu_limit, memory_limit_mb = EXCLUDED.memory_limit_mb,
			timeout_seconds = EXCLUDED.timeout_seconds, network_isolated = EXCLUDED.network_isolated,
			filesystem_scope = EXCLUDED.filesystem_scope, max_instances = EXCLUDED.max_instances,
			env_vars = EXCLUDED.env_vars, allowed_domains = EXCLUDED.allowed_domains,
			blocked_domains = EXCLUDED.blocked_domains, rate_limit_rpm = EXCLUDED.rate_limit_rpm,
			updated_at = EXCLUDED.updated_at
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		sandbox.ID, sandbox.PluginID, sandbox.Tier, sandbox.CPULimit, sandbox.MemoryLimitMB,
		sandbox.TimeoutSeconds, sandbox.NetworkIsolated, sandbox.FilesystemScope, sandbox.MaxInstances,
		envVarsRaw, sandbox.AllowedDomains, sandbox.BlockedDomains, sandbox.RateLimitRPM,
		now, now,
	).Scan(&sandbox.CreatedAt, &sandbox.UpdatedAt)

	if err != nil {
		return fmt.Errorf("upsert sandbox: %w", err)
	}

	return nil
}

func (r *PluginRepository) ListPermissions(ctx context.Context, pluginID string) ([]PluginPermission, error) {
	query := `
		SELECT id, plugin_id, permission_type, permission_action, resource,
		       granted, granted_at, granted_by, expires_at, created_at
		FROM plugin_permissions
		WHERE plugin_id = $1
		ORDER BY permission_type, permission_action
	`
	rows, err := r.db.QueryContext(ctx, query, pluginID)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []PluginPermission
	for rows.Next() {
		var perm PluginPermission
		var resource sql.NullString
		var grantedAt, expiresAt sql.NullTime
		var grantedBy sql.NullString

		err := rows.Scan(
			&perm.ID, &perm.PluginID, &perm.PermissionType, &perm.PermissionAction,
			&resource, &perm.Granted, &grantedAt, &grantedBy, &expiresAt, &perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}

		if resource.Valid {
			perm.Resource = resource.String
		}
		if grantedAt.Valid {
			perm.GrantedAt = &grantedAt.Time
		}
		if grantedBy.Valid {
			perm.GrantedBy = &grantedBy.String
		}
		if expiresAt.Valid {
			perm.ExpiresAt = &expiresAt.Time
		}

		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

func (r *PluginRepository) SetPermission(ctx context.Context, perm *PluginPermission) error {
	if perm.ID == "" {
		perm.ID = uuid.New().String()
	}
	if perm.Granted && perm.GrantedAt == nil {
		now := time.Now()
		perm.GrantedAt = &now
	}

	query := `
		INSERT INTO plugin_permissions (id, plugin_id, permission_type, permission_action, resource,
		                                granted, granted_at, granted_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (plugin_id, permission_type, COALESCE(resource, '')) DO UPDATE SET
			granted = EXCLUDED.granted, granted_at = EXCLUDED.granted_at,
			granted_by = EXCLUDED.granted_by, expires_at = EXCLUDED.expires_at
	`

	_, err := r.db.ExecContext(ctx, query,
		perm.ID, perm.PluginID, perm.PermissionType, perm.PermissionAction,
		perm.Resource, perm.Granted, perm.GrantedAt, perm.GrantedBy, perm.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("set permission: %w", err)
	}

	return nil
}

func (r *PluginRepository) CreateVersion(ctx context.Context, version *PluginVersion) error {
	if version.ID == "" {
		version.ID = uuid.New().String()
	}
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now()
	}

	manifestRaw, _ := json.Marshal(version.Manifest)

	query := `
		INSERT INTO plugin_versions (id, plugin_id, version, changelog, manifest, size_bytes, signature, release_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (plugin_id, version) DO UPDATE SET
			changelog = EXCLUDED.changelog, manifest = EXCLUDED.manifest,
			size_bytes = EXCLUDED.size_bytes, signature = EXCLUDED.signature
		RETURNING created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		version.ID, version.PluginID, version.Version, version.Changelog,
		manifestRaw, version.SizeBytes, version.Signature, version.ReleaseAt, version.CreatedAt,
	).Scan(&version.CreatedAt)

	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}

	return nil
}

func (r *PluginRepository) ListVersions(ctx context.Context, pluginID string) ([]PluginVersion, error) {
	query := `
		SELECT id, plugin_id, version, changelog, manifest, size_bytes, signature, release_at, created_at
		FROM plugin_versions
		WHERE plugin_id = $1
		ORDER BY release_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, pluginID)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	defer rows.Close()

	var versions []PluginVersion
	for rows.Next() {
		var v PluginVersion
		var changelog, signature sql.NullString
		var manifest []byte

		err := rows.Scan(
			&v.ID, &v.PluginID, &v.Version, &changelog, &manifest,
			&v.SizeBytes, &signature, &v.ReleaseAt, &v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}

		if changelog.Valid {
			v.Changelog = changelog.String
		}
		if signature.Valid {
			v.Signature = signature.String
		}
		if len(manifest) > 0 {
			_ = json.Unmarshal(manifest, &v.Manifest)
		}

		versions = append(versions, v)
	}

	return versions, rows.Err()
}

func (r *PluginRepository) GetPreviousVersion(ctx context.Context, pluginID, currentVersion string) (*PluginVersion, error) {
	query := `
		SELECT id, plugin_id, version, changelog, manifest, size_bytes, signature, release_at, created_at
		FROM plugin_versions
		WHERE plugin_id = $1 AND version != $2
		ORDER BY release_at DESC
		LIMIT 1
	`
	var v PluginVersion
	var changelog, signature sql.NullString
	var manifest []byte

	err := r.db.QueryRowContext(ctx, query, pluginID, currentVersion).Scan(
		&v.ID, &v.PluginID, &v.Version, &changelog, &manifest,
		&v.SizeBytes, &signature, &v.ReleaseAt, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get previous version: %w", err)
	}

	if changelog.Valid {
		v.Changelog = changelog.String
	}
	if signature.Valid {
		v.Signature = signature.String
	}
	if len(manifest) > 0 {
		_ = json.Unmarshal(manifest, &v.Manifest)
	}

	return &v, nil
}

func scanPlugin(rows interface{ Scan(dst ...interface{}) error }) (*Plugin, error) {
	var plugin Plugin
	var manifest, config, metadata []byte
	var desc, authorEmail, authorWebsite, iconURL, homepageURL, repositoryURL, license, signature sql.NullString
	var sizeBytes sql.NullInt64
	var enabledAt sql.NullTime
	var errorMsg sql.NullString
	var verified sql.NullBool

	err := rows.Scan(
		&plugin.ID, &plugin.TenantID, &manifest, &plugin.PluginType, &plugin.Name, &plugin.Version,
		&desc, &plugin.AuthorName, &authorEmail, &authorWebsite, &plugin.Category, &plugin.Status,
		&iconURL, &homepageURL, &repositoryURL, &license, &sizeBytes, &signature, &verified,
		&config, &metadata, &plugin.InstalledAt, &plugin.UpdatedAt, &enabledAt, &errorMsg,
	)
	if err != nil {
		return nil, err
	}

	if desc.Valid {
		plugin.Description = desc.String
	}
	if authorEmail.Valid {
		plugin.AuthorEmail = authorEmail.String
	}
	if authorWebsite.Valid {
		plugin.AuthorWebsite = authorWebsite.String
	}
	if iconURL.Valid {
		plugin.IconURL = iconURL.String
	}
	if homepageURL.Valid {
		plugin.HomepageURL = homepageURL.String
	}
	if repositoryURL.Valid {
		plugin.RepositoryURL = repositoryURL.String
	}
	if license.Valid {
		plugin.License = license.String
	}
	if signature.Valid {
		plugin.Signature = signature.String
	}
	if sizeBytes.Valid {
		plugin.SizeBytes = int(sizeBytes.Int64)
	}
	if verified.Valid {
		plugin.Verified = verified.Bool
	}
	if enabledAt.Valid {
		plugin.EnabledAt = &enabledAt.Time
	}
	if errorMsg.Valid {
		plugin.ErrorMessage = &errorMsg.String
	}
	if len(manifest) > 0 {
		_ = json.Unmarshal(manifest, &plugin.Manifest)
	}
	if len(config) > 0 {
		_ = json.Unmarshal(config, &plugin.Config)
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &plugin.Metadata)
	}

	return &plugin, nil
}