// Package unified provides a single analytics service that aggregates metrics
// from billing, dashboard, state, factory, agent, and registry.
package unified

import (
	"time"

	"github.com/google/uuid"
)

// MetricKind is the type of metric for time series queries.
type MetricKind string

const (
	MetricKindExecutions   MetricKind = "executions"   // function_logs
	MetricKindStateOps     MetricKind = "state_ops"    // state read+write
	MetricKindBilling      MetricKind = "billing"      // usage_events/rollups
	MetricKindAgentCalls   MetricKind = "agent_calls"  // agent_execution_records
	MetricKindRegistryRuns  MetricKind = "registry_runs" // registry_function_executions by tenant
)

// Granularity is the time bucket for time series.
type Granularity string

const (
	GranularityHour  Granularity = "hour"
	GranularityDay   Granularity = "day"
	GranularityWeek  Granularity = "week"
	GranularityMonth Granularity = "month"
)

// TenantSummary is the unified summary for a tenant over a time range.
type TenantSummary struct {
	TenantID    uuid.UUID    `json:"tenant_id"`
	Start       time.Time    `json:"start"`
	End         time.Time    `json:"end"`
	GeneratedAt time.Time    `json:"generated_at"`

	// Function executions (from function_logs)
	FunctionExecutions int64 `json:"function_executions"`

	// State usage (from state_usage_metrics / AggregationService)
	StateStorageBytes int64 `json:"state_storage_bytes"`
	StateReadOps      int64 `json:"state_read_ops"`
	StateWriteOps     int64 `json:"state_write_ops"`
	StateActiveStates int64 `json:"state_active_states,omitempty"`

	// Billing usage (from usage_events/usage_rollups)
	BillingQuantity int `json:"billing_quantity"`

	// Agent runs (from agent_execution_records)
	AgentCalls    int64   `json:"agent_calls"`
	AgentCostUSD  float64 `json:"agent_cost_usd"`
	AgentSuccessCount int64 `json:"agent_success_count"`
	AgentErrorCount   int64 `json:"agent_error_count"`

	// Registry executions (registry_function_executions for tenant's functions)
	RegistryExecutions int64 `json:"registry_executions,omitempty"`
}

// TimeSeriesPoint is one bucket in a time series.
type TimeSeriesPoint struct {
	Bucket time.Time `json:"bucket"` // start of the period
	Value  float64   `json:"value"`
	Count  int64     `json:"count,omitempty"`
}

// TimeSeriesResponse is the response for tenant time series.
type TimeSeriesResponse struct {
	TenantID    uuid.UUID    `json:"tenant_id"`
	MetricKind  MetricKind   `json:"metric_kind"`
	Granularity Granularity  `json:"granularity"`
	Start       time.Time    `json:"start"`
	End         time.Time    `json:"end"`
	Points      []TimeSeriesPoint `json:"points"`
}

// PlatformSummary is the platform-wide rollup for admin.
type PlatformSummary struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	GeneratedAt time.Time `json:"generated_at"`

	TotalTenantsActive   int64 `json:"total_tenants_active"`
	TotalFunctionExecs   int64 `json:"total_function_executions"`
	TotalStateReadOps    int64 `json:"total_state_read_ops"`
	TotalStateWriteOps   int64 `json:"total_state_write_ops"`
	TotalAgentCalls      int64 `json:"total_agent_calls"`
	TotalRegistryExecs   int64 `json:"total_registry_executions,omitempty"`
}

// ResourceType and EventType for the canonical event store
const (
	ResourceTypeFunction = "function"
	ResourceTypeState    = "state"
	ResourceTypeBilling  = "billing"
	ResourceTypeAgent    = "agent"
	ResourceTypeRegistry = "registry"
)

const (
	EventTypeExecution   = "execution"
	EventTypeStateRead   = "state_read"
	EventTypeStateWrite  = "state_write"
	EventTypeUsage       = "usage"
	EventTypeAgentRun    = "agent_run"
	EventTypeRegistryRun = "registry_run"
)

// AnalyticsEvent is a row in the canonical analytics_events table.
type AnalyticsEvent struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ResourceType string     `json:"resource_type" gorm:"not null;index"`
	EventType    string     `json:"event_type" gorm:"not null;index"`
	OccurredAt   time.Time  `json:"occurred_at" gorm:"not null;index"`
	Quantity     int64      `json:"quantity" gorm:"not null;default:1"`
	LatencyMs    *int       `json:"latency_ms,omitempty"`
	CostUSD      *float64   `json:"cost_usd,omitempty" gorm:"type:decimal(12,6)"`
	ResourceID   *uuid.UUID         `json:"resource_id,omitempty" gorm:"type:uuid"`
	Payload      map[string]any    `json:"payload,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time         `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name.
func (AnalyticsEvent) TableName() string { return "analytics_events" }

// AnalyticsRollup is a row in the analytics_rollups table.
type AnalyticsRollup struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Period     string    `json:"period" gorm:"not null"` // hour, day, month
	PeriodStart time.Time `json:"period_start" gorm:"not null;index"`
	MetricName string    `json:"metric_name" gorm:"not null;index"`
	Value      float64   `json:"value" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"not null;default:now();autoUpdateTime"`
}

// TableName returns the table name.
func (AnalyticsRollup) TableName() string { return "analytics_rollups" }

// Rollup metric names (must match aggregation job and service reads)
const (
	RollupMetricFunctionExecutions = "function_executions"
	RollupMetricStateReadOps       = "state_read_ops"
	RollupMetricStateWriteOps      = "state_write_ops"
	RollupMetricStateStorageBytes  = "state_storage_bytes"
	RollupMetricBillingQuantity    = "billing_quantity"
	RollupMetricAgentCalls         = "agent_calls"
	RollupMetricAgentCostUSD       = "agent_cost_usd"
	RollupMetricRegistryExecutions = "registry_executions"
)
