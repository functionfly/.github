package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// FunctionWithMCPSettings combines RegistryFunction info with its MCP settings.
// Used by the MCP Center dashboard to list all functions with their MCP config.
type FunctionWithMCPSettings struct {
	RegistryFunction
	MCPSettings
}

// ListFunctionsWithMCPSettings returns all functions with their MCP settings (if configured).
// This includes both MCP-enabled and MCP-disabled functions that have MCP settings rows.
// The MCP settings row is created when a user first configures MCP (even if later disabled).
func (r *RegistryRepository) ListFunctionsWithMCPSettings(ctx context.Context, limit, offset int) ([]FunctionWithMCPSettings, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var results []FunctionWithMCPSettings
	var total int64

	// Count total functions that have MCP settings (via LEFT JOIN)
	if err := r.db.WithContext(ctx).
		Model(&RegistryFunction{}).
		Joins("LEFT JOIN registry_function_mcp_settings m ON m.function_id = registry_functions.id").
		Where("m.function_id IS NOT NULL").
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count functions with mcp settings: %w", err)
	}

	// Query functions with MCP settings
	// Note: verified_mcp is computed dynamically:
	//   - true if the function author is 'functionfly' (auto-verified)
	//   - true if registry_function_verification_status.overall_status = 'verified'
	rows, err := r.db.WithContext(ctx).
		Table("registry_functions").
		Select(`registry_functions.*,
			m.enabled as mcp_enabled, m.transports as mcp_transports,
			m.expose_input_schema as mcp_expose_input_schema, m.expose_output_schema as mcp_expose_output_schema,
			m.tool_name_override as mcp_tool_name_override, m.rate_limit_per_min as mcp_rate_limit_per_min,
			m.allowlist_origins as mcp_allowlist_origins,
			COALESCE(vs.overall_status = 'verified', false)
				OR registry_functions.author = 'functionfly' as mcp_verified_mcp,
			m.invocation_count as mcp_invocation_count, m.last_invoked_at as mcp_last_invoked_at,
			m.enabled_at as mcp_enabled_at, m.created_at as mcp_created_at, m.updated_at as mcp_updated_at`).
		Joins("LEFT JOIN registry_function_mcp_settings m ON m.function_id = registry_functions.id").
		Joins("LEFT JOIN registry_function_versions v ON v.function_id = registry_functions.id AND v.version = registry_functions.latest_version").
		Joins("LEFT JOIN registry_function_verification_status vs ON vs.function_version_id = v.id").
		Where("m.function_id IS NOT NULL").
		Order("m.invocation_count DESC, m.updated_at DESC").
		Limit(limit).
		Offset(offset).
		Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("list functions with mcp settings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fn RegistryFunction
		var mcpEnabled bool
		var mcpTransports string
		var mcpExposeInput, mcpExposeOutput bool
		var mcpToolNameOverride sql.NullString
		var mcpRateLimit int
		var mcpAllowlistOrigins string
		var mcpVerifiedMCP bool
		var mcpInvocationCount int64
		var mcpLastInvokedAt sql.NullTime
		var mcpEnabledAt sql.NullTime
		var mcpCreatedAt, mcpUpdatedAt time.Time

		if err := rows.Scan(
			&fn.ID, &fn.Author, &fn.Name, &fn.LatestVersion, &fn.Title, &fn.Description,
			&fn.Category, &fn.Tags, &fn.Visibility, &fn.PricePerCall, &fn.PopularityScore,
			&fn.ReliabilityScore, &fn.DeterministicScore, &fn.Capabilities, &fn.EmbedConfig,
			&fn.Settings, &fn.TenantID, &fn.OwnerUserID, &fn.PlatformFeePaid, &fn.PlatformFeeAmountUSD,
			&fn.LastFeeChargedAt, &fn.CreatedAt, &fn.UpdatedAt,
			&fn.TrustScore, &fn.TrustTier, &fn.TrustUpdatedAt, &fn.TrustCalculationVersion,
			&fn.Providers, &fn.Region, &fn.Code, &fn.EnvVars, &fn.Schedule,
			&fn.PlaygroundEnabled, &fn.PlaygroundConfig, &fn.Status, &fn.AppID, &fn.Versions, &fn.Rating,
			&mcpEnabled, &mcpTransports, &mcpExposeInput, &mcpExposeOutput,
			&mcpToolNameOverride, &mcpRateLimit, &mcpAllowlistOrigins,
			&mcpVerifiedMCP, &mcpInvocationCount, &mcpLastInvokedAt,
			&mcpEnabledAt, &mcpCreatedAt, &mcpUpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan function row: %w", err)
		}

		mcpSettings := MCPSettings{
			FunctionID:        fn.ID,
			Enabled:            mcpEnabled,
			ExposeInputSchema:  mcpExposeInput,
			ExposeOutputSchema: mcpExposeOutput,
			RateLimitPerMin:    mcpRateLimit,
			VerifiedMCP:       mcpVerifiedMCP,
			InvocationCount:    mcpInvocationCount,
			CreatedAt:          mcpCreatedAt,
			UpdatedAt:          mcpUpdatedAt,
		}

		// Parse transports JSON array
		if mcpTransports != "" {
			var transports []string
			if err := json.Unmarshal([]byte(mcpTransports), &transports); err == nil {
				mcpSettings.Transports = transports
			}
		}

		// Parse allowlist_origins JSON array
		if mcpAllowlistOrigins != "" {
			var origins []string
			if err := json.Unmarshal([]byte(mcpAllowlistOrigins), &origins); err == nil {
				mcpSettings.AllowlistOrigins = origins
			}
		}

		if mcpToolNameOverride.Valid {
			mcpSettings.ToolNameOverride = mcpToolNameOverride
		}
		if mcpLastInvokedAt.Valid {
			mcpSettings.LastInvokedAt = &mcpLastInvokedAt.Time
		}
		if mcpEnabledAt.Valid {
			mcpSettings.EnabledAt = &mcpEnabledAt.Time
		}

		results = append(results, FunctionWithMCPSettings{
			RegistryFunction: fn,
			MCPSettings:     mcpSettings,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, int(total), nil
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
// Global MCP Analytics Storage Methods
// =============================================================================

// MCPTimeSeriesPoint represents a single point in a time series for analytics.
type MCPTimeSeriesPoint struct {
	Time  time.Time
	Count int64
}

// MCPClientCount represents a client and its invocation count.
type MCPClientCount struct {
	Client string
	Count  int64
}

// MCPFunctionCallCount represents a function and its call count.
type MCPFunctionCallCount struct {
	Author string
	Name   string
	Calls  int64
}

// MCPTransportCount represents a transport and its usage count.
type MCPTransportCount struct {
	Transport string
	Count     int64
}

// MCPConnectionRecord represents a client connection summary.
type MCPConnectionRecord struct {
	ClientType         string
	ClientIcon         string
	Status             string
	ConnectedFunctions int
	TotalInvocations   int64
	LastConnectedAt    *time.Time
}

// GetTotalMCPInvocations returns the total number of MCP invocations since the given time.
func (r *RegistryRepository) GetTotalMCPInvocations(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&MCPInvocationRecord{}).
		Where("timestamp >= ?", since).
		Count(&count).Error
	return count, err
}

// GetUniqueMCPClients returns the number of unique callers since the given time.
func (r *RegistryRepository) GetUniqueMCPClients(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&MCPInvocationRecord{}).
		Where("timestamp >= ?", since).
		Where("caller_id IS NOT NULL AND caller_id != ''").
		Distinct("caller_id").
		Count(&count).Error
	return count, err
}

// GetAverageMCPLatency returns the average latency in milliseconds since the given time.
func (r *RegistryRepository) GetAverageMCPLatency(ctx context.Context, since time.Time) (int, error) {
	var avg float64
	err := r.db.WithContext(ctx).
		Model(&MCPInvocationRecord{}).
		Where("timestamp >= ?", since).
		Select("COALESCE(AVG(duration_ms), 0)").
		Scan(&avg).Error
	return int(avg), err
}

// GetMCPSuccessRate returns the percentage of successful (non-error) invocations.
func (r *RegistryRepository) GetMCPSuccessRate(ctx context.Context, since time.Time) (float64, error) {
	var total, errors int64

	if err := r.db.WithContext(ctx).
		Model(&MCPInvocationRecord{}).
		Where("timestamp >= ?", since).
		Count(&total).Error; err != nil {
		return 0, err
	}

	if total == 0 {
		return 100.0, nil
	}

	if err := r.db.WithContext(ctx).
		Model(&MCPInvocationRecord{}).
		Where("timestamp >= ?", since).
		Where("status_code >= 400 OR error_code IS NOT NULL").
		Count(&errors).Error; err != nil {
		return 0, err
	}

	return float64(total-errors) / float64(total) * 100, nil
}

// GetMCPCallsOverTime returns time series data of MCP invocations grouped by hour.
func (r *RegistryRepository) GetMCPCallsOverTime(ctx context.Context, since time.Time) ([]MCPTimeSeriesPoint, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			date_trunc('hour', timestamp) as time_bucket,
			COUNT(*) as count
		FROM registry_mcp_invocations
		WHERE timestamp >= ?
		GROUP BY time_bucket
		ORDER BY time_bucket ASC
	`, since).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MCPTimeSeriesPoint
	for rows.Next() {
		var p MCPTimeSeriesPoint
		if err := rows.Scan(&p.Time, &p.Count); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

// GetMCPClientBreakdown returns the breakdown of invocations by client type.
func (r *RegistryRepository) GetMCPClientBreakdown(ctx context.Context, since time.Time) ([]MCPClientCount, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(NULLIF(caller_id, ''), 'unknown') as client,
			COUNT(*) as count
		FROM registry_mcp_invocations
		WHERE timestamp >= ?
		GROUP BY client
		ORDER BY count DESC
		LIMIT 20
	`, since).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MCPClientCount
	for rows.Next() {
		var c MCPClientCount
		if err := rows.Scan(&c.Client, &c.Count); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, rows.Err()
}

// GetMCPTopFunctions returns the top functions by invocation count.
func (r *RegistryRepository) GetMCPTopFunctions(ctx context.Context, since time.Time, limit int) ([]MCPFunctionCallCount, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			f.author,
			f.name,
			COUNT(i.invocation_id) as calls
		FROM registry_mcp_invocations i
		JOIN registry_functions f ON f.id = i.function_id
		WHERE i.timestamp >= ?
		GROUP BY f.author, f.name
		ORDER BY calls DESC
		LIMIT ?
	`, since, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MCPFunctionCallCount
	for rows.Next() {
		var f MCPFunctionCallCount
		if err := rows.Scan(&f.Author, &f.Name, &f.Calls); err != nil {
			return nil, err
		}
		results = append(results, f)
	}
	return results, rows.Err()
}

// GetMCPTransportUsage returns the breakdown of invocations by transport type.
func (r *RegistryRepository) GetMCPTransportUsage(ctx context.Context, since time.Time) ([]MCPTransportCount, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(transport, 'streamable-http') as transport,
			COUNT(*) as count
		FROM registry_mcp_invocations
		WHERE timestamp >= ?
		GROUP BY transport
		ORDER BY count DESC
	`, since).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MCPTransportCount
	for rows.Next() {
		var t MCPTransportCount
		if err := rows.Scan(&t.Transport, &t.Count); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// GetMCPConnections returns aggregated connection information per client type.
func (r *RegistryRepository) GetMCPConnections(ctx context.Context) ([]MCPConnectionRecord, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(NULLIF(caller_id, ''), 'unknown') as client_type,
			COUNT(DISTINCT function_id) as connected_functions,
			COUNT(*) as total_invocations,
			MAX(timestamp) as last_connected_at,
			CASE
				WHEN MAX(timestamp) IS NULL THEN 'never'
				WHEN MAX(timestamp) < NOW() - INTERVAL '7 days' THEN 'stale'
				ELSE 'active'
			END as status
		FROM registry_mcp_invocations
		GROUP BY caller_id
		ORDER BY total_invocations DESC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MCPConnectionRecord
	for rows.Next() {
		var c MCPConnectionRecord
		if err := rows.Scan(&c.ClientType, &c.ConnectedFunctions, &c.TotalInvocations, &c.LastConnectedAt, &c.Status); err != nil {
			return nil, err
		}
		// Set default icon based on client type
		c.ClientIcon = getClientIcon(c.ClientType)
		results = append(results, c)
	}
	return results, rows.Err()
}

// MCPStats holds aggregate MCP registry metrics for the public marketing page.
type MCPStats struct {
	TotalFunctions    int64  `json:"total_functions"`
	VerifiedFunctions int64  `json:"verified_functions"`
	TotalExecutions   string `json:"total_executions"`
	TrustTiers        int64  `json:"trust_tiers"`
	Runtimes          int    `json:"runtimes"`
}

// GetMCPStats returns aggregate MCP registry metrics for the public stats HUD.
func (r *RegistryRepository) GetMCPStats(ctx context.Context) (*MCPStats, error) {
	stats := &MCPStats{}

	// Count MCP-enabled public functions
	if err := r.db.WithContext(ctx).
		Table("registry_function_mcp_settings m").
		Joins("JOIN registry_functions f ON f.id = m.function_id").
		Where("m.enabled = ?", true).
		Where("f.visibility = ?", "public").
		Count(&stats.TotalFunctions).Error; err != nil {
		return nil, fmt.Errorf("failed to count MCP functions: %w", err)
	}

	// Count verified MCP-enabled public functions
	// A function is verified if: overall_status = 'verified' OR author = 'functionfly'
	var verifiedCount int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT f.id)
		FROM registry_function_mcp_settings m
		JOIN registry_functions f ON f.id = m.function_id
		LEFT JOIN registry_function_versions v ON v.function_id = f.id AND v.version = f.latest_version
		LEFT JOIN registry_function_verification_status vs ON vs.function_version_id = v.id
		WHERE m.enabled = true
		  AND f.visibility = 'public'
		  AND (vs.overall_status = 'verified' OR f.author = 'functionfly')
	`).Scan(&verifiedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count verified functions: %w", err)
	}
	stats.VerifiedFunctions = verifiedCount

	// Sum total invocations
	var totalInvocations int64
	if err := r.db.WithContext(ctx).
		Table("registry_function_mcp_settings m").
		Joins("JOIN registry_functions f ON f.id = m.function_id").
		Where("m.enabled = ?", true).
		Where("f.visibility = ?", "public").
		Select("COALESCE(SUM(m.invocation_count), 0)").
		Scan(&totalInvocations).Error; err != nil {
		return nil, fmt.Errorf("failed to sum invocations: %w", err)
	}
	stats.TotalExecutions = formatCount(totalInvocations)

	// Count distinct trust tiers
	if err := r.db.WithContext(ctx).
		Table("registry_function_mcp_settings m").
		Joins("JOIN registry_functions f ON f.id = m.function_id").
		Where("m.enabled = ?", true).
		Where("f.visibility = ?", "public").
		Where("f.trust_tier IS NOT NULL").
		Where("f.trust_tier != ?", "").
		Where("f.trust_tier != ?", "untrusted").
		Distinct("f.trust_tier").
		Count(&stats.TrustTiers).Error; err != nil {
		return nil, fmt.Errorf("failed to count trust tiers: %w", err)
	}
	// Always show 4 tiers as the max
	if stats.TrustTiers < 4 {
		stats.TrustTiers = 4
	}

	// Count distinct runtimes from latest versions
	var runtimes []string
	if err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT v.runtime
		FROM registry_function_mcp_settings m
		JOIN registry_functions f ON f.id = m.function_id
		JOIN registry_function_versions v ON v.function_id = f.id AND v.version = f.latest_version
		WHERE m.enabled = true
		  AND f.visibility = 'public'
		  AND v.runtime IS NOT NULL AND v.runtime != ''
	`).Scan(&runtimes).Error; err != nil {
		return nil, fmt.Errorf("failed to load runtimes: %w", err)
	}
	stats.Runtimes = len(runtimes)
	if stats.Runtimes < 5 {
		stats.Runtimes = 5
	}

	return stats, nil
}

// MCPCategory holds a category with its function count.
type MCPCategory struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// GetMCPCategories returns all MCP-enabled categories with function counts.
func (r *RegistryRepository) GetMCPCategories(ctx context.Context) ([]MCPCategory, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(NULLIF(f.category, ''), 'uncategorized') as category,
			COUNT(*) as count
		FROM registry_function_mcp_settings m
		JOIN registry_functions f ON f.id = m.function_id
		WHERE m.enabled = true
		  AND f.visibility = 'public'
		GROUP BY f.category
		ORDER BY count DESC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MCPCategory
	for rows.Next() {
		var c MCPCategory
		var category string
		if err := rows.Scan(&category, &c.Count); err != nil {
			return nil, err
		}
		c.Slug = category
		c.Label = humanizeCategory(category)
		results = append(results, c)
	}
	return results, rows.Err()
}

// humanizeCategory converts a slug like "document-processing" to "Document Processing".
func humanizeCategory(slug string) string {
	// Handle known categories
	known := map[string]string{
		"document-processing": "Documents",
		"data-extraction":    "Data Extraction",
		"ai":                 "AI & ML",
		"communication":       "Communication",
		"finance":             "Finance",
		"developer-tools":      "Developer Tools",
		"uncategorized":       "Uncategorized",
	}
	if label, ok := known[slug]; ok {
		return label
	}
	// Generic: convert slug-case to Title Case
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(string(p[0])) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// formatCount formats a large number for display (e.g., 2400000000 -> "2.4B").
func formatCount(n int64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB+", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM+", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK+", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// getClientIcon returns an icon identifier for known client types.
func getClientIcon(clientType string) string {
	switch clientType {
	case "claude-desktop":
		return "claude"
	case "cursor":
		return "cursor"
	case "vscode":
		return "vscode"
	case "windsurf":
		return "windsurf"
	default:
		return "generic"
	}
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
