package storage

import (
	"time"

	"github.com/google/uuid"
)

// App represents an application
type App struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Name      string    `json:"name" gorm:"not null"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;not null"`
	Backends  []Backend `json:"backends,omitempty" gorm:"foreignKey:AppID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Backend represents a backend server for an app
type Backend struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AppID        uuid.UUID `json:"app_id" gorm:"type:uuid;not null"`
	App          *App      `json:"app,omitempty" gorm:"foreignKey:AppID"`
	Provider     string    `json:"provider" gorm:"not null"`
	Region       string    `json:"region" gorm:"not null"`
	URL          string    `json:"url" gorm:"not null"`
	SharedSecret string    `json:"shared_secret" gorm:"column:shared_secret;not null"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	Priority     *int      `json:"priority,omitempty"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID           uuid.UUID `json:"id"`
	BackendID    uuid.UUID `json:"backend_id"`
	Timestamp    time.Time `json:"timestamp"`
	OK           bool      `json:"ok"`
	StatusCode   int       `json:"status_code"`
	LatencyMs    int       `json:"latency_ms"`
	ErrorMessage string    `json:"error_message"`
}

// CircuitState represents circuit breaker state
type CircuitState struct {
	BackendID     uuid.UUID  `json:"backend_id"`
	State         string     `json:"state"` // "closed", "open", "half-open"
	SinceTs       time.Time  `json:"since_ts"`
	FailCount     int        `json:"fail_count"`
	SuccessCount  int        `json:"success_count"`
	LastFailureTs *time.Time `json:"last_failure_ts"`
	LastSuccessTs *time.Time `json:"last_success_ts"`
}

// BackendStatus represents the combined status of a backend
type BackendStatus struct {
	Backend           *Backend      `json:"backend"`
	CircuitState      *CircuitState `json:"circuit_state"`
	LatestHealthCheck *HealthCheck  `json:"latest_health_check"`
}

// RoutingEvent represents a routing decision and its outcome
type RoutingEvent struct {
	ID        uuid.UUID `json:"id"`
	AppID     uuid.UUID `json:"app_id"`
	BackendID uuid.UUID `json:"backend_id"`
	Timestamp time.Time `json:"timestamp"`
	LatencyMs int       `json:"latency_ms"`
	Outcome   string    `json:"outcome"` // "success", "failure", "timeout"
	RequestID string    `json:"request_id"`
}

// Deployment represents a deployment of an app
type Deployment struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AppID        uuid.UUID `json:"app_id" gorm:"type:uuid;not null"`
	App          *App      `json:"app,omitempty" gorm:"foreignKey:AppID"`
	Provider     string    `json:"provider" gorm:"not null"`
	Region       string    `json:"region" gorm:"not null"`
	DeploymentID string    `json:"deployment_id" gorm:"not null"`            // Provider-specific deployment ID
	Status       string    `json:"status" gorm:"not null;default:'pending'"` // "pending", "deploying", "success", "failed", "rollback"
	ArtifactKey  string    `json:"artifact_key" gorm:"not null"`             // Reference to stored artifact
	Routes       []string  `json:"routes" gorm:"type:jsonb"`                 // Route patterns bound to this deployment
	Message      string    `json:"message"`                                  // Status message or error details
	Metadata     string    `json:"metadata" gorm:"type:json"`                // JSON metadata from provider
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DeploymentArtifact represents a stored deployment artifact
type DeploymentArtifact struct {
	Key         string    `json:"key"`
	AppID       uuid.UUID `json:"app_id"`
	Provider    string    `json:"provider"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
}
