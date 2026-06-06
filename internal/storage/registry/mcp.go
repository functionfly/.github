package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// =============================================================================
// MCP Settings
// =============================================================================

// MCPSettings represents the per-function MCP configuration stored in
// registry_function_mcp_settings. A row exists only for functions that the
// owner has explicitly enabled for MCP exposure.
type MCPSettings struct {
	FunctionID          uuid.UUID       `json:"function_id"          gorm:"type:uuid;primaryKey"`
	Enabled             bool            `json:"enabled"              gorm:"not null;default:false"`
	Transports          StringArray     `json:"transports"           gorm:"type:text[];not null;default:'{streamable-http}'"`
	ExposeInputSchema   bool            `json:"expose_input_schema"  gorm:"not null;default:true"`
	ExposeOutputSchema  bool            `json:"expose_output_schema" gorm:"not null;default:false"`
	ToolNameOverride    sql.NullString  `json:"tool_name_override,omitempty" gorm:"type:text"`
	RateLimitPerMin     int             `json:"rate_limit_per_min"   gorm:"not null;default:60"`
	AllowlistOrigins    StringArray     `json:"allowlist_origins"    gorm:"type:text[];not null;default:'{}'"`
	VerifiedMCP         bool            `json:"verified_mcp"         gorm:"not null;default:false"`
	VerifiedAt          *time.Time      `json:"verified_at,omitempty"           gorm:"type:timestamptz"`
	VerifiedBy          *uuid.UUID      `json:"verified_by,omitempty"           gorm:"type:uuid"`
	EnabledAt           *time.Time      `json:"enabled_at,omitempty"            gorm:"type:timestamptz"`
	EnabledBy           *uuid.UUID      `json:"enabled_by,omitempty"            gorm:"type:uuid"`
	LastInvokedAt       *time.Time      `json:"last_invoked_at,omitempty"       gorm:"type:timestamptz"`
	InvocationCount     int64           `json:"invocation_count"     gorm:"not null;default:0"`
	CreatedAt           time.Time       `json:"created_at"           gorm:"autoCreateTime"`
	UpdatedAt           time.Time       `json:"updated_at"           gorm:"autoUpdateTime"`
}

// TableName pins the table to avoid drift from struct name changes.
func (MCPSettings) TableName() string { return "registry_function_mcp_settings" }

// MCPSettingsInput is the validated shape accepted by the publish flow
// (PublishRequest.MCP) and by the dashboard settings panel.
type MCPSettingsInput struct {
	Enabled            *bool    `json:"enabled"`
	Transports         []string `json:"transports,omitempty"`
	ExposeInputSchema  *bool    `json:"expose_input_schema,omitempty"`
	ExposeOutputSchema *bool    `json:"expose_output_schema,omitempty"`
	ToolNameOverride   string   `json:"tool_name_override,omitempty"`
	RateLimitPerMin    int      `json:"rate_limit_per_min,omitempty"`
	AllowlistOrigins   []string `json:"allowlist_origins,omitempty"`
}

// Validate enforces invariants the DB CHECK constraints also enforce.
// We re-validate in Go so callers get a 400 (not a 500 from a DB error).
func (in *MCPSettingsInput) Validate() error {
	if in == nil {
		return nil
	}
	if in.ToolNameOverride != "" {
		if l := len(in.ToolNameOverride); l < 1 || l > 64 {
			return fmt.Errorf("tool_name_override must be 1-64 chars")
		}
		for _, r := range in.ToolNameOverride {
			ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == '_' || r == '-'
			if !ok {
				return fmt.Errorf("tool_name_override may only contain [a-zA-Z0-9_-]")
			}
		}
	}
	if in.RateLimitPerMin < 0 || in.RateLimitPerMin > 10000 {
		return fmt.Errorf("rate_limit_per_min must be between 0 and 10000")
	}
	allowed := map[string]bool{"streamable-http": true, "stdio": true}
	for _, t := range in.Transports {
		if !allowed[t] {
			return fmt.Errorf("unsupported transport %q (allowed: streamable-http, stdio)", t)
		}
	}
	return nil
}

// ApplyDefaults fills missing values with the platform defaults so the
// stored row is always complete.
func (in *MCPSettingsInput) ApplyDefaults() {
	if in == nil {
		return
	}
	if in.Transports == nil || len(in.Transports) == 0 {
		in.Transports = []string{"streamable-http"}
	}
	if in.RateLimitPerMin == 0 {
		in.RateLimitPerMin = 60
	}
	if in.AllowlistOrigins == nil {
		in.AllowlistOrigins = []string{}
	}
}

// (The publish-time conversion from functionregistry.MCPPublishSettings to
// MCPSettingsInput is performed in the publish handler to avoid an import
// cycle between this storage package and functionregistry.)
// =============================================================================
// Repository methods
// =============================================================================

// GetMCPSettings returns MCP settings for a function, or nil if not enabled
// / not configured. Returns ErrRecordNotFound only when the function_id does
// not exist; an absent MCP row is signalled by (nil, nil) to match the
// repository convention used elsewhere.
func (r *RegistryRepository) GetMCPSettings(ctx context.Context, functionID uuid.UUID) (*MCPSettings, error) {
	var s MCPSettings
	err := r.db.WithContext(ctx).Where("function_id = ?", functionID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get mcp settings: %w", err)
	}
	return &s, nil
}

// UpsertMCPSettings inserts or updates the MCP settings row for a function.
// When `enabled` flips to true, `enabled_at` is set; when it flips to false,
// `enabled_at` is cleared.
func (r *RegistryRepository) UpsertMCPSettings(ctx context.Context, functionID uuid.UUID, in MCPSettingsInput, actorID *uuid.UUID) (*MCPSettings, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	in.ApplyDefaults()

	now := time.Now().UTC()
	existing, err := r.GetMCPSettings(ctx, functionID)
	if err != nil {
		return nil, err
	}

	transports := StringArray(in.Transports)
	origins := StringArray(in.AllowlistOrigins)
	if transports == nil {
		transports = StringArray{"streamable-http"}
	}
	if origins == nil {
		origins = StringArray{}
	}

	// Build overrides map. Only override non-default fields so existing
	// non-toggled state is preserved.
	updates := map[string]interface{}{
		"transports":            transports,
		"expose_input_schema":   boolOrDefault(in.ExposeInputSchema, true),
		"expose_output_schema":  boolOrDefault(in.ExposeOutputSchema, false),
		"rate_limit_per_min":    in.RateLimitPerMin,
		"allowlist_origins":     origins,
		"updated_at":            now,
	}

	var enabled *bool
	if in.Enabled != nil {
		enabled = in.Enabled
		updates["enabled"] = *enabled
	}

	if in.ToolNameOverride != "" {
		updates["tool_name_override"] = in.ToolNameOverride
	} else if existing != nil && existing.ToolNameOverride.Valid {
		// keep existing
	}

	// enabled_at bookkeeping
	switch {
	case existing == nil && enabled != nil && *enabled:
		updates["enabled_at"] = now
		if actorID != nil {
			updates["enabled_by"] = *actorID
		}
	case existing != nil && !existing.Enabled && enabled != nil && *enabled:
		updates["enabled_at"] = now
		if actorID != nil {
			updates["enabled_by"] = *actorID
		}
	case existing != nil && existing.Enabled && enabled != nil && !*enabled:
		updates["enabled_at"] = nil
	}

	if existing == nil {
		// Insert path
		s := &MCPSettings{
			FunctionID:         functionID,
			Enabled:            enabled != nil && *enabled,
			Transports:         transports,
			ExposeInputSchema:  boolOrDefault(in.ExposeInputSchema, true),
			ExposeOutputSchema: boolOrDefault(in.ExposeOutputSchema, false),
			RateLimitPerMin:    in.RateLimitPerMin,
			AllowlistOrigins:   origins,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if in.ToolNameOverride != "" {
			s.ToolNameOverride = sqlNullString(in.ToolNameOverride)
		}
		if enabled != nil && *enabled {
			s.EnabledAt = &now
			if actorID != nil {
				s.EnabledBy = actorID
			}
		}
		if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
			return nil, fmt.Errorf("insert mcp settings: %w", err)
		}
		return s, nil
	}

	// Update path
	if err := r.db.WithContext(ctx).Model(&MCPSettings{}).
		Where("function_id = ?", functionID).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update mcp settings: %w", err)
	}
	return r.GetMCPSettings(ctx, functionID)
}

// DisableMCPSettings removes the enabled flag and clears enabled_at. The row
// is preserved (for audit) so it shows up in admin dashboards.
func (r *RegistryRepository) DisableMCPSettings(ctx context.Context, functionID uuid.UUID, actorID *uuid.UUID) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"enabled":     false,
		"enabled_at":  nil,
		"updated_at":  now,
	}
	return r.db.WithContext(ctx).Model(&MCPSettings{}).
		Where("function_id = ?", functionID).
		Updates(updates).Error
}

// ListEnabledMCPSettings returns all functions that are currently MCP-enabled.
// Used by the public `/v1/mcp/tools` index and by the marketing site registry
// page. Pagination is by function_id (cursor pagination when needed).
func (r *RegistryRepository) ListEnabledMCPSettings(ctx context.Context, category, runtime string, minTrust float64, limit, offset int) ([]MCPSettings, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	base := r.db.WithContext(ctx).Table("registry_function_mcp_settings m").
		Select(`m.*`).
		Joins("JOIN registry_functions f ON f.id = m.function_id").
		Where("m.enabled = ?", true).
		Where("f.visibility = ?", "public")

	if category != "" {
		base = base.Where("f.category = ?", category)
	}
	if runtime != "" {
		base = base.Where("f.latest_version IS NOT NULL")
	}
	if minTrust > 0 {
		base = base.Where("f.trust_score >= ?", minTrust)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count mcp settings: %w", err)
	}

	var rows []MCPSettings
	if err := base.Order("m.verified_mcp DESC, m.invocation_count DESC, m.updated_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list mcp settings: %w", err)
	}
	return rows, int(total), nil
}

// IncrementMCPInvocationCount atomically increments invocation_count and sets
// last_invoked_at. Called from the tools/call handler.
func (r *RegistryRepository) IncrementMCPInvocationCount(ctx context.Context, functionID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&MCPSettings{}).
		Where("function_id = ?", functionID).
		Updates(map[string]interface{}{
			"invocation_count": gorm.Expr("invocation_count + 1"),
			"last_invoked_at":  gorm.Expr("now()"),
		}).Error
}

// SetMCPSettingsVerified is admin-only: marks a function as "Verified MCP"
// (the curated quality bar). Audited via the verified_by / verified_at columns.
func (r *RegistryRepository) SetMCPSettingsVerified(ctx context.Context, functionID uuid.UUID, verified bool, actorID uuid.UUID) error {
	updates := map[string]interface{}{
		"verified_mcp": verified,
		"updated_at":   time.Now().UTC(),
	}
	if verified {
		updates["verified_at"] = time.Now().UTC()
		updates["verified_by"] = actorID
	} else {
		updates["verified_at"] = nil
		updates["verified_by"] = nil
	}
	return r.db.WithContext(ctx).Model(&MCPSettings{}).
		Where("function_id = ?", functionID).
		Updates(updates).Error
}

// =============================================================================
// MCP Invocations (observability)
// =============================================================================

// MCPInvocationRecord is a single row in registry_mcp_invocations.
type MCPInvocationRecord struct {
	FunctionID    uuid.UUID
	InvocationID  uuid.UUID
	CallerID      string
	CallerOrigin  string
	Transport     string
	Method        string
	DurationMs    int
	StatusCode    int
	ErrorCode     string
	RequestBytes  int
	ResponseBytes int
	Timestamp     time.Time
}

// RecordMCPInvocation inserts a single invocation row. The partitioned parent
// table routes the write to the correct monthly partition. This is fire-and-
// safe; callers should never fail a request because logging failed.
func (r *RegistryRepository) RecordMCPInvocation(ctx context.Context, rec MCPInvocationRecord) error {
	if rec.InvocationID == uuid.Nil {
		rec.InvocationID = uuid.New()
	}
	if rec.Timestamp.IsZero() {
		rec.Timestamp = time.Now().UTC()
	}
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO registry_mcp_invocations
		    (function_id, invocation_id, caller_id, caller_origin, transport, method,
		     duration_ms, status_code, error_code, request_bytes, response_bytes, timestamp)
		VALUES (?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, -1), NULLIF(?, -1), ?)
	`, rec.FunctionID, rec.InvocationID, rec.CallerID, rec.CallerOrigin, rec.Transport, rec.Method,
		rec.DurationMs, rec.StatusCode, rec.ErrorCode,
		intOrMinusOne(rec.RequestBytes), intOrMinusOne(rec.ResponseBytes),
		rec.Timestamp).Error
	if err != nil {
		logrus.WithError(err).WithField("function_id", rec.FunctionID).Warn("failed to record mcp invocation")
	}
	return err
}

// AggregateMCPInvocationsByFunction returns simple counters used by the
// dashboard. Bounded lookback to keep the query fast.
func (r *RegistryRepository) AggregateMCPInvocationsByFunction(ctx context.Context, functionID uuid.UUID, since time.Time) (count int64, errorCount int64, p95Ms int, err error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
		    COUNT(*) AS total,
		    COUNT(*) FILTER (WHERE status_code >= 400 OR error_code IS NOT NULL) AS errors,
		    COALESCE(percentile_disc(0.95) WITHIN GROUP (ORDER BY duration_ms), 0)::int AS p95
		FROM registry_mcp_invocations
		WHERE function_id = ? AND timestamp >= ?
	`, functionID, since).Rows()
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&count, &errorCount, &p95Ms); err != nil {
			return 0, 0, 0, err
		}
	}
	return
}

// =============================================================================
// helpers
// =============================================================================

func boolOrDefault(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// jsonRaw is a tiny helper to keep callers terse (currently unused, reserved
// for future config merging).
var _ = json.Marshal

func intOrMinusOne(v int) interface{} {
	if v == 0 {
		return -1
	}
	return v
}

// SQLNullString is a small alias to avoid a hard dep cycle. (types.go imports
// database/sql, so we re-export.)
type sqlNullStringT = sql.NullString

// sqlNullString converts a non-empty string to a sql.NullString.
func sqlNullString(s string) sqlNullStringT {
	if s == "" {
		return sqlNullStringT{}
	}
	return sqlNullStringT{String: s, Valid: true}
}
