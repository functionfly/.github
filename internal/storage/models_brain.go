package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BrainSignal struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	ConnectorSlug  string          `json:"connector_slug"`
	SignalType     string          `json:"signal_type"`
	EntityID       string          `json:"entity_id,omitempty"`
	EntityName     string          `json:"entity_name,omitempty"`
	Fact           string          `json:"fact"`
	Importance     int             `json:"importance"`
	SourceURL      string          `json:"source_url,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	LastSeenAt     time.Time       `json:"last_seen_at"`
	TTLHours       int             `json:"ttl_hours"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

type BrainStats struct {
	TotalSignals     int              `json:"total_signals"`
	SignalsByType    map[string]int   `json:"signals_by_type"`
	SignalsByConnector map[string]int `json:"signals_by_connector"`
	OldestSignal     *time.Time       `json:"oldest_signal,omitempty"`
	NewestSignal     *time.Time       `json:"newest_signal,omitempty"`
	MemoryUsed       int              `json:"memory_used"`
	MemoryMax        int              `json:"memory_max"`
	RetentionDays    int              `json:"retention_days"`
}

type BrainFeedbackRequest struct {
	SignalID uuid.UUID `json:"signal_id"`
	Helpful  bool      `json:"helpful"`
	Context  string    `json:"context,omitempty"`
}

type BrainSearchResult struct {
	Signal   *BrainSignal `json:"signal"`
	Score    float64      `json:"score"`
	Distance float64      `json:"distance,omitempty"`
}

type SignalListParams struct {
	TenantID      uuid.UUID
	ConnectorSlug string
	SignalType    string
	Limit         int
	Offset        int
	SortBy        string
	SortDir       string
}

type BrainComposer struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	Name          string          `json:"name"`
	Schedule      string          `json:"schedule"`
	SignalFilters json.RawMessage `json:"signal_filters"`
	OutputFormat  string          `json:"output_format"`
	Actions       json.RawMessage `json:"actions"`
	IsActive      bool            `json:"is_active"`
	LastRunAt     *time.Time      `json:"last_run_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type SignalFilter struct {
	ConnectorSlugs []string `json:"connector_slugs"`
	SignalTypes    []string `json:"signal_types"`
	ImportanceMin  int      `json:"importance_min"`
	TimeWindow     string   `json:"time_window"`
}

type ComposerAction struct {
	Type       string          `json:"type"`
	Config     json.RawMessage `json:"config"`
}

type BrainTrigger struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	AgentID        *uuid.UUID      `json:"agent_id,omitempty"`
	Name           string          `json:"name"`
	SignalTypes    []string        `json:"signal_types"`
	ConnectorSlugs []string        `json:"connector_slugs"`
	MinImportance  int             `json:"min_importance"`
	Schedule       string          `json:"schedule"`
	Action         string          `json:"action"`
	ActionConfig   json.RawMessage `json:"action_config,omitempty"`
	IsActive       bool            `json:"is_active"`
	LastFiredAt    *time.Time      `json:"last_fired_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type AnalyticsEvent struct {
	ID            uuid.UUID       `json:"id"`
	EventType     string          `json:"event_type"`
	TenantTier    string          `json:"tenant_tier"`
	ConnectorSlug string          `json:"connector_slug,omitempty"`
	SignalType    string          `json:"signal_type,omitempty"`
	Importance    int             `json:"importance,omitempty"`
	SignalsCount  int             `json:"signals_count,omitempty"`
	FactLength    int             `json:"fact_length,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}
