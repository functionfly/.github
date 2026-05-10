package consciousness

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Repository handles persistence for consciousness data.
type Repository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewRepository creates a new consciousness repository.
func NewRepository(db *sql.DB, logger *logrus.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

// --- Insights ---

// CreateInsight inserts a new insight.
func (r *Repository) CreateInsight(ctx context.Context, insight *Insight) error {
	insightDataJSON, _ := json.Marshal(insight.InsightData)
	actionDataJSON, _ := json.Marshal(insight.ActionData)
	actionPreviewJSON, _ := json.Marshal(insight.ActionPreview)

	query := `
		INSERT INTO consciousness_insights (
			tenant_id, category, severity, priority, title, message, summary,
			function_id, graph_id, agent_id, related_function_ids,
			insight_data, action_type, action_data, action_preview,
			trajectory, projected_days, confidence,
			status, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18,
			$19, $20
		)
		RETURNING id, created_at, updated_at`

	relatedIDs := pq.Array(insight.RelatedFunctionIDs)

	return r.db.QueryRowContext(ctx, query,
		insight.TenantID, insight.Category, insight.Severity, insight.Priority,
		insight.Title, insight.Message, insight.Summary,
		insight.FunctionID, insight.GraphID, insight.AgentID, relatedIDs,
		insightDataJSON, insight.ActionType, actionDataJSON, actionPreviewJSON,
		insight.Trajectory, insight.ProjectedDays, insight.Confidence,
		insight.Status, insight.ExpiresAt,
	).Scan(&insight.ID, &insight.CreatedAt, &insight.UpdatedAt)
}

// ListInsights returns insights for a tenant with filtering.
func (r *Repository) ListInsights(ctx context.Context, params ListInsightsParams) ([]*Insight, int, error) {
	where := []string{"tenant_id = $1"}
	args := []interface{}{params.TenantID}
	argIdx := 2

	if params.Category != nil {
		where = append(where, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *params.Category)
		argIdx++
	}
	if params.Severity != nil {
		where = append(where, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, *params.Severity)
		argIdx++
	}
	if params.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *params.Status)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM consciousness_insights WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count insights: %w", err)
	}

	// Fetch page
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset

	query := fmt.Sprintf(`
		SELECT id, tenant_id, category, severity, priority, title, message, summary,
			function_id, graph_id, agent_id, related_function_ids,
			insight_data, action_type, action_data, action_preview,
			trajectory, projected_days, confidence,
			status, dismissed_at, applied_at, expires_at, superseded_by,
			channels_sent, read_at, created_at, updated_at
		FROM consciousness_insights
		WHERE %s
		ORDER BY priority DESC, created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list insights: %w", err)
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		insight, err := scanInsight(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan insight: %w", err)
		}
		insights = append(insights, insight)
	}

	return insights, total, rows.Err()
}

// GetInsight returns a single insight by ID.
func (r *Repository) GetInsight(ctx context.Context, id, tenantID uuid.UUID) (*Insight, error) {
	query := `
		SELECT id, tenant_id, category, severity, priority, title, message, summary,
			function_id, graph_id, agent_id, related_function_ids,
			insight_data, action_type, action_data, action_preview,
			trajectory, projected_days, confidence,
			status, dismissed_at, applied_at, expires_at, superseded_by,
			channels_sent, read_at, created_at, updated_at
		FROM consciousness_insights
		WHERE id = $1 AND tenant_id = $2`

	row := r.db.QueryRowContext(ctx, query, id, tenantID)
	insight := &Insight{}
	var relatedIDs pq.Int64Array
	var channelsSent pq.StringArray
	var insightData, actionData, actionPreview []byte

	err := row.Scan(
		&insight.ID, &insight.TenantID, &insight.Category, &insight.Severity, &insight.Priority,
		&insight.Title, &insight.Message, &insight.Summary,
		&insight.FunctionID, &insight.GraphID, &insight.AgentID, &relatedIDs,
		&insightData, &insight.ActionType, &actionData, &actionPreview,
		&insight.Trajectory, &insight.ProjectedDays, &insight.Confidence,
		&insight.Status, &insight.DismissedAt, &insight.AppliedAt, &insight.ExpiresAt, &insight.SupersededBy,
		&channelsSent, &insight.ReadAt, &insight.CreatedAt, &insight.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get insight: %w", err)
	}

	// Unmarshal JSONB fields
	if len(insightData) > 0 {
		_ = json.Unmarshal(insightData, &insight.InsightData)
	}
	if len(actionData) > 0 {
		_ = json.Unmarshal(actionData, &insight.ActionData)
	}
	if len(actionPreview) > 0 {
		_ = json.Unmarshal(actionPreview, &insight.ActionPreview)
	}

	return insight, nil
}

// DismissInsight marks an insight as dismissed.
func (r *Repository) DismissInsight(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `
		UPDATE consciousness_insights
		SET status = 'dismissed', dismissed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND status = 'active'`
	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("dismiss insight: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insight not found or already dismissed")
	}
	return nil
}

// ApplyInsight marks an insight as applied.
func (r *Repository) ApplyInsight(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `
		UPDATE consciousness_insights
		SET status = 'applied', applied_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND status = 'active'`
	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("apply insight: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insight not found or already applied")
	}
	return nil
}

// MarkChannelsSent records which channels an insight was delivered to.
func (r *Repository) MarkChannelsSent(ctx context.Context, id uuid.UUID, channels []string) error {
	query := `
		UPDATE consciousness_insights
		SET channels_sent = array_cat(channels_sent, $2), updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, pq.Array(channels))
	return err
}

// ExpireOldInsights marks expired insights.
func (r *Repository) ExpireOldInsights(ctx context.Context) (int64, error) {
	query := `
		UPDATE consciousness_insights
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'active' AND expires_at IS NOT NULL AND expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("expire insights: %w", err)
	}
	return result.RowsAffected()
}

// CountActiveInsights returns the count of active insights for a tenant.
func (r *Repository) CountActiveInsights(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'active'),
			COUNT(*) FILTER (WHERE status = 'active' AND severity = 'critical')
		FROM consciousness_insights
		WHERE tenant_id = $1`
	var active, critical int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&active, &critical)
	return active, critical, err
}

// HasRecentInsight checks if a similar insight already exists (dedup).
func (r *Repository) HasRecentInsight(ctx context.Context, tenantID uuid.UUID, category InsightCategory, functionID *uuid.UUID, window time.Duration) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM consciousness_insights
			WHERE tenant_id = $1 AND category = $2 AND status = 'active'
			AND ($3::uuid IS NULL OR function_id = $3)
			AND created_at > $4
		)`
	var exists bool
	since := time.Now().Add(-window)
	err := r.db.QueryRowContext(ctx, query, tenantID, category, functionID, since).Scan(&exists)
	return exists, err
}

// --- Awareness Scores ---

// UpsertScore inserts or updates the awareness score for a tenant.
func (r *Repository) UpsertScore(ctx context.Context, score *SystemAwarenessScore) error {
	query := `
		INSERT INTO system_awareness_scores (
			tenant_id, overall_score,
			health_score, efficiency_score, scalability_score, reliability_score, optimization_score,
			functions_analyzed, graphs_analyzed, agents_analyzed,
			active_insights, critical_insights,
			previous_score, trend, computed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (tenant_id) DO UPDATE SET
			overall_score = EXCLUDED.overall_score,
			health_score = EXCLUDED.health_score,
			efficiency_score = EXCLUDED.efficiency_score,
			scalability_score = EXCLUDED.scalability_score,
			reliability_score = EXCLUDED.reliability_score,
			optimization_score = EXCLUDED.optimization_score,
			functions_analyzed = EXCLUDED.functions_analyzed,
			graphs_analyzed = EXCLUDED.graphs_analyzed,
			agents_analyzed = EXCLUDED.agents_analyzed,
			active_insights = EXCLUDED.active_insights,
			critical_insights = EXCLUDED.critical_insights,
			previous_score = EXCLUDED.previous_score,
			trend = EXCLUDED.trend,
			computed_at = EXCLUDED.computed_at,
			created_at = system_awareness_scores.created_at
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		score.TenantID, score.OverallScore,
		score.HealthScore, score.EfficiencyScore, score.ScalabilityScore,
		score.ReliabilityScore, score.OptimizationScore,
		score.FunctionsAnalyzed, score.GraphsAnalyzed, score.AgentsAnalyzed,
		score.ActiveInsights, score.CriticalInsights,
		score.PreviousScore, score.Trend, score.ComputedAt,
	).Scan(&score.ID)
}

// GetScore returns the current awareness score for a tenant.
func (r *Repository) GetScore(ctx context.Context, tenantID uuid.UUID) (*SystemAwarenessScore, error) {
	query := `
		SELECT id, tenant_id, overall_score,
			health_score, efficiency_score, scalability_score, reliability_score, optimization_score,
			functions_analyzed, graphs_analyzed, agents_analyzed,
			active_insights, critical_insights,
			previous_score, trend, computed_at, created_at
		FROM system_awareness_scores
		WHERE tenant_id = $1`

	score := &SystemAwarenessScore{}
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&score.ID, &score.TenantID, &score.OverallScore,
		&score.HealthScore, &score.EfficiencyScore, &score.ScalabilityScore,
		&score.ReliabilityScore, &score.OptimizationScore,
		&score.FunctionsAnalyzed, &score.GraphsAnalyzed, &score.AgentsAnalyzed,
		&score.ActiveInsights, &score.CriticalInsights,
		&score.PreviousScore, &score.Trend, &score.ComputedAt, &score.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get score: %w", err)
	}
	return score, nil
}

// GetScoreHistory returns the last N scores for trend analysis.
func (r *Repository) GetScoreHistory(ctx context.Context, tenantID uuid.UUID, limit int) ([]*SystemAwarenessScore, error) {
	// Note: current schema only stores latest via upsert on tenant_id.
	// For history, we'd need a separate history table or append-only inserts.
	// For now, return the current score.
	score, err := r.GetScore(ctx, tenantID)
	if err != nil || score == nil {
		return nil, err
	}
	return []*SystemAwarenessScore{score}, nil
}

// --- Preferences ---

// GetPreferences returns consciousness preferences for a tenant.
func (r *Repository) GetPreferences(ctx context.Context, tenantID uuid.UUID) (*Preferences, error) {
	query := `
		SELECT id, tenant_id, email_enabled, slack_enabled, slack_webhook_url,
			inapp_enabled, webhook_enabled, webhook_url, webhook_secret,
			digest_frequency, quiet_hours_start, quiet_hours_end, timezone,
			enabled_categories, min_notify_severity,
			auto_apply_enabled, auto_apply_categories,
			created_at, updated_at
		FROM consciousness_preferences
		WHERE tenant_id = $1`

	prefs := &Preferences{}
	var categories pq.StringArray
	var autoCategories pq.StringArray

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&prefs.ID, &prefs.TenantID, &prefs.EmailEnabled, &prefs.SlackEnabled, &prefs.SlackWebhookURL,
		&prefs.InAppEnabled, &prefs.WebhookEnabled, &prefs.WebhookURL, &prefs.WebhookSecret,
		&prefs.DigestFrequency, &prefs.QuietHoursStart, &prefs.QuietHoursEnd, &prefs.Timezone,
		&categories, &prefs.MinNotifySeverity,
		&prefs.AutoApplyEnabled, &autoCategories,
		&prefs.CreatedAt, &prefs.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults
			return DefaultPreferences(tenantID), nil
		}
		return nil, fmt.Errorf("get preferences: %w", err)
	}

	prefs.EnabledCategories = []string(categories)
	prefs.AutoApplyCategories = []string(autoCategories)
	return prefs, nil
}

// UpsertPreferences creates or updates consciousness preferences.
func (r *Repository) UpsertPreferences(ctx context.Context, prefs *Preferences) error {
	query := `
		INSERT INTO consciousness_preferences (
			tenant_id, email_enabled, slack_enabled, slack_webhook_url,
			inapp_enabled, webhook_enabled, webhook_url, webhook_secret,
			digest_frequency, quiet_hours_start, quiet_hours_end, timezone,
			enabled_categories, min_notify_severity,
			auto_apply_enabled, auto_apply_categories
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		ON CONFLICT (tenant_id) DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			slack_enabled = EXCLUDED.slack_enabled,
			slack_webhook_url = EXCLUDED.slack_webhook_url,
			inapp_enabled = EXCLUDED.inapp_enabled,
			webhook_enabled = EXCLUDED.webhook_enabled,
			webhook_url = EXCLUDED.webhook_url,
			webhook_secret = EXCLUDED.webhook_secret,
			digest_frequency = EXCLUDED.digest_frequency,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			timezone = EXCLUDED.timezone,
			enabled_categories = EXCLUDED.enabled_categories,
			min_notify_severity = EXCLUDED.min_notify_severity,
			auto_apply_enabled = EXCLUDED.auto_apply_enabled,
			auto_apply_categories = EXCLUDED.auto_apply_categories,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		prefs.TenantID, prefs.EmailEnabled, prefs.SlackEnabled, prefs.SlackWebhookURL,
		prefs.InAppEnabled, prefs.WebhookEnabled, prefs.WebhookURL, prefs.WebhookSecret,
		prefs.DigestFrequency, prefs.QuietHoursStart, prefs.QuietHoursEnd, prefs.Timezone,
		pq.Array(prefs.EnabledCategories), prefs.MinNotifySeverity,
		prefs.AutoApplyEnabled, pq.Array(prefs.AutoApplyCategories),
	).Scan(&prefs.ID, &prefs.CreatedAt, &prefs.UpdatedAt)
}

// --- Delivery Log ---

// LogDelivery records a notification delivery attempt.
func (r *Repository) LogDelivery(ctx context.Context, log *DeliveryLog) error {
	query := `
		INSERT INTO consciousness_delivery_log (insight_id, tenant_id, channel, status, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sent_at`
	return r.db.QueryRowContext(ctx, query,
		log.InsightID, log.TenantID, log.Channel, log.Status, log.ErrorMsg,
	).Scan(&log.ID, &log.SentAt)
}

// --- Helpers ---

// DefaultPreferences returns default preferences for a new tenant.
func DefaultPreferences(tenantID uuid.UUID) *Preferences {
	return &Preferences{
		TenantID:            tenantID,
		EmailEnabled:        true,
		SlackEnabled:        false,
		InAppEnabled:        true,
		WebhookEnabled:      false,
		DigestFrequency:     "daily",
		Timezone:            "UTC",
		EnabledCategories:   []string{"traffic", "cost", "redundancy", "health", "marketplace", "scaling"},
		MinNotifySeverity:   "warning",
		AutoApplyEnabled:    false,
		AutoApplyCategories: []string{},
	}
}

// scanInsight scans a row into an Insight struct.
func scanInsight(rows *sql.Rows) (*Insight, error) {
	insight := &Insight{}
	var relatedIDs pq.Int64Array
	var channelsSent pq.StringArray
	var insightData, actionData, actionPreview []byte

	err := rows.Scan(
		&insight.ID, &insight.TenantID, &insight.Category, &insight.Severity, &insight.Priority,
		&insight.Title, &insight.Message, &insight.Summary,
		&insight.FunctionID, &insight.GraphID, &insight.AgentID, &relatedIDs,
		&insightData, &insight.ActionType, &actionData, &actionPreview,
		&insight.Trajectory, &insight.ProjectedDays, &insight.Confidence,
		&insight.Status, &insight.DismissedAt, &insight.AppliedAt, &insight.ExpiresAt, &insight.SupersededBy,
		&channelsSent, &insight.ReadAt, &insight.CreatedAt, &insight.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(insightData) > 0 {
		_ = json.Unmarshal(insightData, &insight.InsightData)
	}
	if len(actionData) > 0 {
		_ = json.Unmarshal(actionData, &insight.ActionData)
	}
	if len(actionPreview) > 0 {
		_ = json.Unmarshal(actionPreview, &insight.ActionPreview)
	}

	// Convert pq arrays to Go types
	if insight.InsightData == nil {
		insight.InsightData = make(JSONMap)
	}
	if insight.ActionData == nil {
		insight.ActionData = make(JSONMap)
	}
	if insight.ActionPreview == nil {
		insight.ActionPreview = make(JSONMap)
	}
	insight.ChannelsSent = []string(channelsSent)

	return insight, nil
}
