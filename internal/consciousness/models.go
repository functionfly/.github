package consciousness

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInsightNotFound = errors.New("insight not found or not active")

// InsightCategory classifies the type of insight.
type InsightCategory string

const (
	CategoryTraffic     InsightCategory = "traffic"
	CategoryCost        InsightCategory = "cost"
	CategoryRedundancy  InsightCategory = "redundancy"
	CategoryHealth      InsightCategory = "health"
	CategoryMarketplace InsightCategory = "marketplace"
	CategoryScaling     InsightCategory = "scaling"
)

// InsightSeverity indicates urgency.
type InsightSeverity string

const (
	SeverityInfo       InsightSeverity = "info"
	SeverityWarning    InsightSeverity = "warning"
	SeverityCritical   InsightSeverity = "critical"
	SeverityOpportunity InsightSeverity = "opportunity"
)

// InsightStatus tracks lifecycle.
type InsightStatus string

const (
	StatusActive      InsightStatus = "active"
	StatusDismissed   InsightStatus = "dismissed"
	StatusApplied     InsightStatus = "applied"
	StatusExpired     InsightStatus = "expired"
	StatusSuperseded  InsightStatus = "superseded"
)

// InsightTrajectory indicates direction.
type InsightTrajectory string

const (
	TrajectoryImproving InsightTrajectory = "improving"
	TrajectoryStable    InsightTrajectory = "stable"
	TrajectoryDegrading InsightTrajectory = "degrading"
	TrajectoryCritical  InsightTrajectory = "critical"
)

// ActionType identifies what action can be taken.
type ActionType string

const (
	ActionNone            ActionType = ""
	ActionMergeFunctions  ActionType = "merge_functions"
	ActionScaleConfig     ActionType = "scale_config"
	ActionSwapMarketplace ActionType = "swap_marketplace"
	ActionOptimize        ActionType = "optimize"
)

// Insight is the core consciousness data model — a single actionable observation.
type Insight struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	TenantID          uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Category          InsightCategory `json:"category" db:"category"`
	Severity          InsightSeverity `json:"severity" db:"severity"`
	Priority          int             `json:"priority" db:"priority"`
	Title             string          `json:"title" db:"title"`
	Message           string          `json:"message" db:"message"`
	Summary           *string         `json:"summary,omitempty" db:"summary"`
	FunctionID        *uuid.UUID      `json:"function_id,omitempty" db:"function_id"`
	GraphID           *uuid.UUID      `json:"graph_id,omitempty" db:"graph_id"`
	AgentID           *uuid.UUID      `json:"agent_id,omitempty" db:"agent_id"`
	RelatedFunctionIDs []uuid.UUID    `json:"related_function_ids" db:"related_function_ids"`
	InsightData       JSONMap         `json:"insight_data" db:"insight_data"`
	ActionType        ActionType      `json:"action_type" db:"action_type"`
	ActionData        JSONMap         `json:"action_data" db:"action_data"`
	ActionPreview     JSONMap         `json:"action_preview" db:"action_preview"`
	Trajectory        *InsightTrajectory `json:"trajectory,omitempty" db:"trajectory"`
	ProjectedDays     *int            `json:"projected_days,omitempty" db:"projected_days"`
	Confidence        *float64        `json:"confidence,omitempty" db:"confidence"`
	Status            InsightStatus   `json:"status" db:"status"`
	DismissedAt       *time.Time      `json:"dismissed_at,omitempty" db:"dismissed_at"`
	AppliedAt         *time.Time      `json:"applied_at,omitempty" db:"applied_at"`
	ExpiresAt         *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
	SupersededBy      *uuid.UUID      `json:"superseded_by,omitempty" db:"superseded_by"`
	ChannelsSent      []string        `json:"channels_sent" db:"channels_sent"`
	ReadAt            *time.Time      `json:"read_at,omitempty" db:"read_at"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`

	// Joined fields (not stored in this table)
	FunctionName *string `json:"function_name,omitempty" db:"-"`
}

// SystemAwarenessScore represents the computed health of a tenant's backend.
type SystemAwarenessScore struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OverallScore      float64    `json:"overall_score" db:"overall_score"`
	HealthScore       float64    `json:"health_score" db:"health_score"`
	EfficiencyScore   float64    `json:"efficiency_score" db:"efficiency_score"`
	ScalabilityScore  float64    `json:"scalability_score" db:"scalability_score"`
	ReliabilityScore  float64    `json:"reliability_score" db:"reliability_score"`
	OptimizationScore float64    `json:"optimization_score" db:"optimization_score"`
	FunctionsAnalyzed int        `json:"functions_analyzed" db:"functions_analyzed"`
	GraphsAnalyzed    int        `json:"graphs_analyzed" db:"graphs_analyzed"`
	AgentsAnalyzed    int        `json:"agents_analyzed" db:"agents_analyzed"`
	ActiveInsights    int        `json:"active_insights" db:"active_insights"`
	CriticalInsights  int        `json:"critical_insights" db:"critical_insights"`
	PreviousScore     *float64   `json:"previous_score,omitempty" db:"previous_score"`
	Trend             *string    `json:"trend,omitempty" db:"trend"`
	ComputedAt        time.Time  `json:"computed_at" db:"computed_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// Preferences holds per-tenant consciousness notification settings.
type Preferences struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	TenantID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	EmailEnabled        bool       `json:"email_enabled" db:"email_enabled"`
	SlackEnabled        bool       `json:"slack_enabled" db:"slack_enabled"`
	SlackWebhookURL     *string    `json:"slack_webhook_url,omitempty" db:"slack_webhook_url"`
	InAppEnabled        bool       `json:"inapp_enabled" db:"inapp_enabled"`
	WebhookEnabled      bool       `json:"webhook_enabled" db:"webhook_enabled"`
	WebhookURL          *string    `json:"webhook_url,omitempty" db:"webhook_url"`
	WebhookSecret       *string    `json:"-" db:"webhook_secret"`
	DigestFrequency     string     `json:"digest_frequency" db:"digest_frequency"`
	QuietHoursStart     *string    `json:"quiet_hours_start,omitempty" db:"quiet_hours_start"`
	QuietHoursEnd       *string    `json:"quiet_hours_end,omitempty" db:"quiet_hours_end"`
	Timezone            string     `json:"timezone" db:"timezone"`
	EnabledCategories   []string   `json:"enabled_categories" db:"enabled_categories"`
	MinNotifySeverity   string     `json:"min_notify_severity" db:"min_notify_severity"`
	AutoApplyEnabled    bool       `json:"auto_apply_enabled" db:"auto_apply_enabled"`
	AutoApplyCategories []string  `json:"auto_apply_categories" db:"auto_apply_categories"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// DeliveryLog tracks notification delivery attempts.
type DeliveryLog struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	InsightID   uuid.UUID  `json:"insight_id" db:"insight_id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Channel     string     `json:"channel" db:"channel"`
	Status      string     `json:"status" db:"status"`
	ErrorMsg    *string    `json:"error_message,omitempty" db:"error_message"`
	SentAt      time.Time  `json:"sent_at" db:"sent_at"`
}

// ScoreWeights controls component contribution to the overall score.
type ScoreWeights struct {
	Health       float64 `json:"health"`
	Efficiency   float64 `json:"efficiency"`
	Scalability  float64 `json:"scalability"`
	Reliability  float64 `json:"reliability"`
	Optimization float64 `json:"optimization"`
}

// DefaultScoreWeights returns sensible defaults.
func DefaultScoreWeights() ScoreWeights {
	return ScoreWeights{
		Health:       0.25,
		Efficiency:   0.20,
		Scalability:  0.20,
		Reliability:  0.20,
		Optimization: 0.15,
	}
}

// ScoreLabel returns a human-readable label for the score.
func ScoreLabel(score float64) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 70:
		return "good"
	case score >= 50:
		return "needs_attention"
	case score >= 30:
		return "at_risk"
	default:
		return "critical"
	}
}

// ListInsightsParams controls insight listing/filtering.
type ListInsightsParams struct {
	TenantID  uuid.UUID
	Category  *InsightCategory
	Severity  *InsightSeverity
	Status    *InsightStatus
	Limit     int
	Offset    int
}

// JSONMap is a wrapper for JSONB fields that marshals to/from JSON.
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface for JSONB.
func (j *JSONMap) Scan(src interface{}) error {
	if src == nil {
		*j = make(JSONMap)
		return nil
	}
	switch v := src.(type) {
	case []byte:
		if len(v) == 0 {
			*j = make(JSONMap)
			return nil
		}
		return json.Unmarshal(v, j)
	case string:
		if v == "" {
			*j = make(JSONMap)
			return nil
		}
		return json.Unmarshal([]byte(v), j)
	}
	*j = make(JSONMap)
	return nil
}

// Digest represents a compiled set of insights for periodic delivery.
type Digest struct {
	TenantID    uuid.UUID             `json:"tenant_id"`
	Period      string                `json:"period"`
	Insights    []*Insight            `json:"insights"`
	Score       *SystemAwarenessScore `json:"score"`
	GeneratedAt time.Time             `json:"generated_at"`
}

// SeverityWeight returns a numeric weight for severity-based sorting.
func SeverityWeight(s InsightSeverity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityWarning:
		return 3
	case SeverityOpportunity:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string { return &s }

// timePtr returns a pointer to the given time.
func timePtr(t time.Time) *time.Time { return &t }

// planRequestLimit returns the monthly request limit for a plan.
func planRequestLimit(plan string) int {
	switch plan {
	case "free":
		return 100_000
	case "starter":
		return 1_000_000
	case "professional":
		return 10_000_000
	case "enterprise", "agent_enterprise":
		return -1
	default:
		return 100_000
	}
}

// Ensure sql is used (compile-time check).
var _ sql.Scanner = (*JSONMap)(nil)
