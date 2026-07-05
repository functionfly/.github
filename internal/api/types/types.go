package types

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// API Response types
type AppResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	TenantID       string    `json:"tenant_id"`
	DeployUrl      string    `json:"deploy_url"`
	DeployUrlIntent string   `json:"deploy_url_intent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type BackendResponse struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	URL          string    `json:"url"`
	SharedSecret string    `json:"shared_secret"`
	CreatedAt    time.Time `json:"created_at"`
}

type CircuitStateResponse struct {
	State         string     `json:"state"`
	SinceTs       time.Time  `json:"since_ts"`
	FailCount     int        `json:"fail_count"`
	SuccessCount  int        `json:"success_count"`
	LastFailureTs *time.Time `json:"last_failure_ts"`
}

type HealthCheckResponse struct {
	Timestamp    time.Time `json:"timestamp"`
	OK           bool      `json:"ok"`
	StatusCode   int       `json:"status_code"`
	LatencyMs    int       `json:"latency_ms"`
	ErrorMessage string    `json:"error_message"`
}

type BackendStatusResponse struct {
	Backend           *BackendResponse      `json:"backend"`
	CircuitState      *CircuitStateResponse `json:"circuit_state"`
	LatestHealthCheck *HealthCheckResponse  `json:"latest_health_check"`
}

type StatusResponse struct {
	App      *AppResponse             `json:"app"`
	Backends []*BackendStatusResponse `json:"backends"`
}

// Function-related types
type CreateFunctionRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	Code      string   `json:"code" validate:"required"`
	Providers []string `json:"providers" validate:"required,min=1"`
	Region    string   `json:"region" validate:"required"`
	EnvVars   []storage.EnvironmentVariable `json:"env_vars,omitempty"`
}

type CreateFunctionResponse struct {
	FunctionID string `json:"function_id"`
}

type UpdateFunctionRequest struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Code      *string  `json:"code,omitempty"`
	Providers []string `json:"providers,omitempty"`
	Region    *string  `json:"region,omitempty"`
	EnvVars   []storage.EnvironmentVariable `json:"env_vars,omitempty"`
}

type DeployFunctionRequest struct {
	FunctionId   uuid.UUID `json:"function_id" validate:"required"`
	BackendID    string    `json:"backend_id" validate:"required"`
	Version      string    `json:"version,omitempty"`
	Environment  string    `json:"environment,omitempty" validate:"omitempty,oneof=dev staging prod"`
}

type DeployFunctionResponse struct {
	FunctionID   string                        `json:"function_id"`
	DeploymentID string                        `json:"deployment_id"`
	URL         string                        `json:"url"`
	Region      string                        `json:"region"`
	Providers   []string                      `json:"providers"`
	Status      string                        `json:"status"`
	Deployments []*storage.FunctionDeployment `json:"deployments,omitempty"`
}

type TestFunctionRequest struct {
	FunctionId *uuid.UUID `json:"function_id,omitempty"`
	Input      string     `json:"input"`
}

type TestFunctionResponse struct {
	Success         bool                     `json:"success"`
	Output          interface{}              `json:"output"`
	ExecutionTimeMs int                      `json:"execution_time_ms"`
	Logs            []*storage.FunctionLog   `json:"logs"`
}

type FunctionMetricsResponse struct {
	Requests     int     `json:"requests"`
	LatencyMs   int     `json:"latency_ms"`
	ErrorRate   float64 `json:"error_rate"`
	UptimePercent float64 `json:"uptime_percent"`
}

// Request/Response types
type CreateAppRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateAppRequest struct {
	Name *string `json:"name,omitempty"`
}

type CreateBackendRequest struct {
	Provider     string `json:"provider"`
	Region       string `json:"region"`
	URL          string `json:"url"`
	SharedSecret string `json:"shared_secret"`
	Priority     *int   `json:"priority,omitempty"`
}

type DeployRequest struct {
	Provider       string                 `json:"provider"`
	Region         string                 `json:"region"`
	Artifact       string                 `json:"artifact"` // Base64 encoded
	Routes         []string               `json:"routes"`
	EnvVars        map[string]string      `json:"env_vars"`
	Secrets        map[string]string      `json:"secrets"`
	ProviderConfig map[string]interface{} `json:"provider_config"`
}

// Analytics types
type AnalyticsService struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Status   string `json:"status"`   // "loading", "loaded", "error", "disabled"
	Config   map[string]interface{} `json:"config"`
	LastUsed *time.Time `json:"last_used,omitempty"`
}

type AnalyticsSettings struct {
	GoogleAnalytics *GoogleAnalyticsConfig `json:"google_analytics,omitempty"`
	Hotjar         *HotjarConfig         `json:"hotjar,omitempty"`
	Services       []AnalyticsService    `json:"services"`
}

type GoogleAnalyticsConfig struct {
	MeasurementID string `json:"measurement_id"`
	Enabled       bool   `json:"enabled"`
}

type HotjarConfig struct {
	SiteID  string `json:"site_id"`
	Enabled bool   `json:"enabled"`
}

type UpdateAnalyticsRequest struct {
	GoogleAnalytics *GoogleAnalyticsConfig `json:"google_analytics,omitempty"`
	Hotjar         *HotjarConfig         `json:"hotjar,omitempty"`
}