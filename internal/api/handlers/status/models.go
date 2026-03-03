// Package status provides the FunctionFly System Status Page API handlers.
// This package implements REST endpoints for platform status, incidents,
// metrics, and maintenance windows.
package status

import (
	"time"

	"github.com/google/uuid"
)

// PlatformStatus represents the overall platform health status
type PlatformStatus struct {
	Status      string               `json:"status"`    // operational|degraded|major_outage|maintenance
	Indicator   string               `json:"indicator"` // none|minor|major|critical
	Description string               `json:"description"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Components  []Component          `json:"components"`
	Incidents   []Incident           `json:"incidents"`
	Maintenance []MaintenanceSummary `json:"maintenance"`
}

// Component represents a system component's status
type Component struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Status       string               `json:"status"` // operational|degraded_performance|partial_outage|major_outage|maintenance
	Type         string               `json:"type"`   // api|database|cache|provider|monitoring
	Description  string               `json:"description"`
	Uptime24h    float64              `json:"uptime_24h"`
	Uptime7d     float64              `json:"uptime_7d"`
	Uptime30d    float64              `json:"uptime_30d"`
	ResponseTime int                  `json:"response_time_ms"`
	LastChecked  time.Time            `json:"last_checked"`
	History      []StatusHistoryPoint `json:"history,omitempty"`
}

// StatusHistoryPoint represents a single point in component status history
type StatusHistoryPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Status         string    `json:"status"`
	ResponseTimeMs int       `json:"response_time_ms"`
}

// Incident represents a system incident or operational issue
type Incident struct {
	ID                 string           `json:"id"`
	Title              string           `json:"title"`
	Severity           string           `json:"severity"` // critical|high|medium|low
	Status             string           `json:"status"`   // investigating|identified|monitoring|resolved
	Description        string           `json:"description"`
	AffectedComponents []string         `json:"affected_components"`
	CreatedAt          time.Time        `json:"created_at"`
	ResolvedAt         *time.Time       `json:"resolved_at,omitempty"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DurationMinutes    *int             `json:"duration_minutes,omitempty"`
	Updates            []IncidentUpdate `json:"updates"`
}

// IncidentUpdate represents an update to an incident
type IncidentUpdate struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // investigating|identified|monitoring|resolved
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *UserRef  `json:"created_by,omitempty"`
}

// UserRef represents a user reference
type UserRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IncidentSummary represents a summarized incident for lists
type IncidentSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MaintenanceWindow represents a scheduled maintenance window
type MaintenanceWindow struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Status             string     `json:"status"` // scheduled|in_progress|completed|cancelled
	ScheduledStart     time.Time  `json:"scheduled_start"`
	ScheduledEnd       time.Time  `json:"scheduled_end"`
	ActualStart        *time.Time `json:"actual_start,omitempty"`
	ActualEnd          *time.Time `json:"actual_end,omitempty"`
	AffectedComponents []string   `json:"affected_components"`
	AffectedProviders  []string   `json:"affected_providers"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// MaintenanceSummary represents a summarized maintenance window
type MaintenanceSummary struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	ScheduledStart time.Time `json:"scheduled_start"`
	ScheduledEnd   time.Time `json:"scheduled_end"`
}

// ProviderStatus represents status for a specific provider
type ProviderStatus struct {
	Name          string          `json:"name"`
	DisplayName   string          `json:"display_name"`
	OverallStatus string          `json:"overall_status"` // operational|degraded|outage
	Regions       []RegionStatus  `json:"regions"`
	Summary       ProviderSummary `json:"summary"`
}

// ProviderSummary contains aggregated provider statistics
type ProviderSummary struct {
	TotalBackends     int     `json:"total_backends"`
	HealthyBackends   int     `json:"healthy_backends"`
	DegradedBackends  int     `json:"degraded_backends"`
	UnhealthyBackends int     `json:"unhealthy_backends"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	ErrorRate         float64 `json:"error_rate"`
}

// RegionStatus represents status for a specific region
type RegionStatus struct {
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	Status    string          `json:"status"` // operational|degraded_performance|partial_outage|major_outage
	LatencyMs float64         `json:"latency_ms"`
	ErrorRate float64         `json:"error_rate"`
	Uptime24h float64         `json:"uptime_24h"`
	Backends  []BackendStatus `json:"backends,omitempty"`
}

// BackendStatus represents status for a specific backend
type BackendStatus struct {
	ID                  string    `json:"id"`
	URL                 string    `json:"url"`
	Status              string    `json:"status"`        // healthy|degraded|unhealthy|unknown
	CircuitState        string    `json:"circuit_state"` // closed|half-open|open
	LatencyMs           int       `json:"latency_ms"`
	LastCheck           time.Time `json:"last_check"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
}

// UptimeMetricsResponse represents uptime metrics response
type UptimeMetricsResponse struct {
	Period        string            `json:"period"`
	Resolution    string            `json:"resolution"`
	OverallUptime float64           `json:"overall_uptime"`
	DataPoints    []UptimeDataPoint `json:"data_points"`
}

// UptimeDataPoint represents a single uptime data point
type UptimeDataPoint struct {
	Timestamp          time.Time          `json:"timestamp"`
	UptimePercent      float64            `json:"uptime_percent"`
	TotalChecks        int                `json:"total_checks"`
	FailedChecks       int                `json:"failed_checks"`
	ComponentBreakdown map[string]float64 `json:"component_breakdown,omitempty"`
}

// LatencyMetricsResponse represents latency metrics response
type LatencyMetricsResponse struct {
	Period       string                  `json:"period"`
	Percentile   string                  `json:"percentile"`
	OverallAvgMs float64                 `json:"overall_avg_ms"`
	DataPoints   []LatencyDataPoint      `json:"data_points"`
	ByProvider   map[string]LatencyStats `json:"by_provider,omitempty"`
}

// LatencyDataPoint represents a single latency data point
type LatencyDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	ValueMs   float64   `json:"value_ms"`
	Provider  string    `json:"provider,omitempty"`
}

// LatencyStats represents latency statistics
type LatencyStats struct {
	AvgMs float64 `json:"avg_ms"`
	MinMs float64 `json:"min_ms"`
	MaxMs float64 `json:"max_ms"`
	P50Ms float64 `json:"p50_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

// ComponentStatusResponse represents components status response
type ComponentStatusResponse struct {
	Components  []Component `json:"components"`
	GeneratedAt time.Time   `json:"generated_at"`
}

// ProvidersStatusResponse represents providers status response
type ProvidersStatusResponse struct {
	Providers   []ProviderStatus `json:"providers"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// IncidentsListResponse represents incidents list response
type IncidentsListResponse struct {
	Incidents  []Incident `json:"incidents"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// MaintenanceListResponse represents maintenance list response
type MaintenanceListResponse struct {
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
}

// CreateIncidentRequest represents request to create an incident
type CreateIncidentRequest struct {
	Title              string         `json:"title"`
	Severity           string         `json:"severity"` // critical|high|medium|low
	Description        string         `json:"description"`
	AffectedComponents []string       `json:"affected_components,omitempty"`
	InitialUpdate      *InitialUpdate `json:"initial_update,omitempty"`
}

// InitialUpdate represents an initial incident update
type InitialUpdate struct {
	Message string `json:"message"`
	Status  string `json:"status"` // investigating|identified|monitoring
}

// UpdateIncidentRequest represents request to update an incident
type UpdateIncidentRequest struct {
	Title              string                 `json:"title,omitempty"`
	Severity           string                 `json:"severity,omitempty"`
	Status             string                 `json:"status,omitempty"`
	Description        string                 `json:"description,omitempty"`
	AffectedComponents []string               `json:"affected_components,omitempty"`
	NewUpdate          *IncidentUpdateRequest `json:"new_update,omitempty"`
}

// IncidentUpdateRequest represents a request to add an incident update
type IncidentUpdateRequest struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// CreateMaintenanceRequest represents request to create a maintenance window
type CreateMaintenanceRequest struct {
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	ScheduledStart     time.Time `json:"scheduled_start"`
	ScheduledEnd       time.Time `json:"scheduled_end"`
	AffectedComponents []string  `json:"affected_components"`
	AffectedProviders  []string  `json:"affected_providers,omitempty"`
}

// ListIncidentsQuery represents query parameters for listing incidents
type ListIncidentsQuery struct {
	Status    string
	Severity  string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// ListMaintenanceQuery represents query parameters for listing maintenance
type ListMaintenanceQuery struct {
	Status   string
	Upcoming bool
	Limit    int
}

// ProviderHealthQuery represents query parameters for provider health
type ProviderHealthQuery struct {
	Provider string
	Region   string
	Detailed bool
}

// UptimeMetricsQuery represents query parameters for uptime metrics
type UptimeMetricsQuery struct {
	Component  string
	Provider   string
	Period     string
	Resolution string
}

// LatencyMetricsQuery represents query parameters for latency metrics
type LatencyMetricsQuery struct {
	Provider   string
	Region     string
	Period     string
	Percentile string
}

// ComponentHealthQuery represents query parameters for component health
type ComponentHealthQuery struct {
	IncludeHistory bool
	ComponentType  string
}

// PrometheusQueryResult represents a Prometheus query result
type PrometheusQueryResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value,omitempty"`  // For vector results [timestamp, value]
	Values [][]interface{}   `json:"values,omitempty"` // For matrix results
}

// PrometheusResponse represents a Prometheus API response
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   *struct {
		ResultType string                  `json:"resultType"`
		Result     []PrometheusQueryResult `json:"result"`
	} `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// HealthMonitorStatus represents health monitor status from health service
type HealthMonitorStatus struct {
	LastProbeTime   time.Time             `json:"last_probe_time"`
	ProbeInterval   time.Duration         `json:"probe_interval"`
	Backends        []BackendHealthStatus `json:"backends"`
	CircuitBreakers []CircuitBreakerState `json:"circuit_breakers"`
}

// BackendHealthStatus represents individual backend health
type BackendHealthStatus struct {
	BackendID    string    `json:"backend_id"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	LastCheck    time.Time `json:"last_check"`
	Healthy      bool      `json:"healthy"`
	LatencyMs    int       `json:"latency_ms"`
	StatusCode   int       `json:"status_code,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// CircuitBreakerState represents circuit breaker state
type CircuitBreakerState struct {
	BackendID       string     `json:"backend_id"`
	State           string     `json:"state"` // closed, open, half-open
	Since           time.Time  `json:"since"`
	FailureCount    int        `json:"failure_count"`
	SuccessCount    int        `json:"success_count"`
	LastFailureTime *time.Time `json:"last_failure_time,omitempty"`
	LastSuccessTime *time.Time `json:"last_success_time,omitempty"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// SubscribeMessage represents a client subscription request
type SubscribeMessage struct {
	Type     string   `json:"type"`
	Channels []string `json:"channels"`
}

// StatusUpdateMessage represents a status update broadcast
type StatusUpdateMessage struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ProviderUpdate represents a real-time provider status update
type ProviderUpdate struct {
	Provider     string  `json:"provider"`
	Region       string  `json:"region"`
	Status       string  `json:"status"`
	LatencyMs    float64 `json:"latency_ms"`
	CircuitState string  `json:"circuit_state,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// DatabaseIncident represents the database incident model
type DatabaseIncident struct {
	ID          uuid.UUID  `db:"id"`
	Title       string     `db:"title"`
	Severity    string     `db:"severity"`
	Status      string     `db:"status"`
	Description string     `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
	ResolvedAt  *time.Time `db:"resolved_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// DatabaseIncidentUpdate represents an incident update from database
type DatabaseIncidentUpdate struct {
	ID         uuid.UUID  `db:"id"`
	IncidentID uuid.UUID  `db:"incident_id"`
	Status     string     `db:"status"`
	Message    string     `db:"message"`
	CreatedAt  time.Time  `db:"created_at"`
	CreatedBy  *uuid.UUID `db:"created_by"`
	UserName   string     `db:"user_name"`
}

// DatabaseMaintenance represents the database maintenance model
type DatabaseMaintenance struct {
	ID                 uuid.UUID  `db:"id"`
	Title              string     `db:"title"`
	Description        string     `db:"description"`
	ScheduledStart     time.Time  `db:"scheduled_start"`
	ScheduledEnd       time.Time  `db:"scheduled_end"`
	ActualStart        *time.Time `db:"actual_start"`
	ActualEnd          *time.Time `db:"actual_end"`
	Status             string     `db:"status"`
	AffectedComponents []string   `db:"affected_components"`
	AffectedProviders  []string   `db:"affected_providers"`
	CreatedBy          uuid.UUID  `db:"created_by"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

// DatabaseAlert represents an alert from the database
type DatabaseAlert struct {
	ID         uuid.UUID  `db:"id"`
	AlertType  string     `db:"alert_type"`
	Severity   string     `db:"severity"`
	TenantID   *uuid.UUID `db:"tenant_id"`
	AppID      *uuid.UUID `db:"app_id"`
	BackendID  *uuid.UUID `db:"backend_id"`
	Title      string     `db:"title"`
	Message    string     `db:"message"`
	Status     string     `db:"status"`
	ResolvedAt *time.Time `db:"resolved_at"`
	Provider   string     `db:"provider"`
	Region     string     `db:"region"`
	AppName    string     `db:"app_name"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

// DatabaseHealthCheck represents a health check from database
type DatabaseHealthCheck struct {
	ID             uuid.UUID `db:"id"`
	CheckType      string    `db:"check_type"`
	ComponentName  string    `db:"component_name"`
	Status         string    `db:"status"`
	ResponseTimeMs *int      `db:"response_time_ms"`
	Message        string    `db:"message"`
	Metadata       []byte    `db:"metadata"`
	CheckedAt      time.Time `db:"checked_at"`
	CreatedAt      time.Time `db:"created_at"`
}

// DatabaseBackendStatus represents backend status from database
type DatabaseBackendStatus struct {
	ID           uuid.UUID  `db:"id"`
	Provider     string     `db:"provider"`
	Region       string     `db:"region"`
	URL          string     `db:"url"`
	State        string     `db:"state"` // closed, half-open, open
	FailCount    int        `db:"fail_count"`
	SuccessCount int        `db:"success_count"`
	LastFailure  *time.Time `db:"last_failure_ts"`
	LastSuccess  *time.Time `db:"last_success_ts"`
	Healthy      bool       `db:"healthy"`
	LatencyMs    int        `db:"latency_ms"`
	StatusCode   int        `db:"status_code"`
	LastCheck    time.Time  `db:"last_check"`
}
