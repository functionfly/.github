package storage

import (
	"time"

	"github.com/google/uuid"
)

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	ID          uuid.UUID              `json:"id"`
	MetricType  string                 `json:"metric_type"` // 'response_time', 'error_rate', 'throughput', 'health_score', 'circuit_state'
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID       *uuid.UUID             `json:"app_id,omitempty"`
	BackendID   *uuid.UUID             `json:"backend_id,omitempty"`
	Value       float64                `json:"value"`
	StringValue string                 `json:"string_value,omitempty"` // For string-based metrics like circuit states
	Unit        string                 `json:"unit"`                   // 'ms', 'percent', 'requests_per_second', 'score', 'state'
	Labels      map[string]interface{} `json:"labels,omitempty"`       // Additional metadata
	Timestamp   time.Time              `json:"timestamp"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Alert represents a monitoring alert or incident
type Alert struct {
	ID         uuid.UUID              `json:"id"`
	AlertType  string                 `json:"alert_type"` // 'health_degraded', 'backend_down', 'high_error_rate', 'circuit_open'
	Severity   string                 `json:"severity"`   // 'info', 'warning', 'error', 'critical'
	TenantID   *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID      *uuid.UUID             `json:"app_id,omitempty"`
	BackendID  *uuid.UUID             `json:"backend_id,omitempty"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message,omitempty"`
	Status     string                 `json:"status"` // 'active', 'acknowledged', 'resolved'
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy *uuid.UUID             `json:"resolved_by,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Additional alert-specific data
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SystemHealthCheck represents a system health check result
type SystemHealthCheck struct {
	ID             uuid.UUID              `json:"id"`
	CheckType      string                 `json:"check_type"` // 'database', 'api', 'external_service', 'disk_space', 'memory'
	ComponentName  string                 `json:"component_name"`
	Status         string                 `json:"status"` // 'healthy', 'degraded', 'unhealthy', 'unknown'
	ResponseTimeMs *int                   `json:"response_time_ms,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CheckedAt      time.Time              `json:"checked_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

// DatabaseMetric represents a historical database performance metric
type DatabaseMetric struct {
	ID         uuid.UUID              `json:"id"`
	MetricType string                 `json:"metric_type"` // 'connections', 'size_gb', 'query_time', 'cache_hit_ratio', 'throughput'
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`               // 'count', 'gb', 'ms', 'ratio', 'qps'
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Additional context
	RecordedAt time.Time              `json:"recorded_at"`
	CreatedAt  time.Time              `json:"created_at"`
}

// MonitoringEvent represents a real-time monitoring event
type MonitoringEvent struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"` // 'request_completed', 'backend_failover', 'circuit_breaker_transition'
	TenantID  *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID     *uuid.UUID             `json:"app_id,omitempty"`
	BackendID *uuid.UUID             `json:"backend_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"` // Event-specific payload
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`
}

// DashboardConfig represents a dashboard configuration
type DashboardConfig struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"` // NULL for tenant-wide configs
	ConfigType string                 `json:"config_type"`       // 'metric_panel', 'alert_rule', 'chart_config'
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"` // Configuration data specific to the type
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Incident represents a system incident or operational issue
type Incident struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Severity    string     `json:"severity" db:"severity"` // "critical", "high", "medium", "low"
	Status      string     `json:"status" db:"status"`     // "resolved", "investigating", "monitoring"
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
