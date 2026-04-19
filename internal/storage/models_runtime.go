package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LocalRuntimeInstance represents a registered local runtime instance
type LocalRuntimeInstance struct {
	ID            uuid.UUID `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeID     string    `json:"runtime_id" gorm:"column:runtime_id;uniqueIndex;not null"` // Unique identifier for this runtime instance
	RuntimeType   string    `json:"runtime_type" gorm:"column:runtime_type;not null"`         // "node18", "node20", "python3.11", etc.
	FunctionName  string    `json:"function_name" gorm:"column:function_name;not null"`
	ManifestPath  string    `json:"manifest_path" gorm:"column:manifest_path;not null"`
	Host          string    `json:"host" gorm:"column:host;not null"`     // Hostname/IP
	Port          int       `json:"port" gorm:"column:port;not null"`     // Port number
	PID           int       `json:"pid" gorm:"column:pid;not null"`       // Process ID
	Status        string    `json:"status" gorm:"column:status;not null"` // "running", "stopped", "error"
	LastHeartbeat time.Time `json:"last_heartbeat" gorm:"column:last_heartbeat;not null"`
	Uptime        int64     `json:"uptime" gorm:"column:uptime;not null"` // Uptime in seconds
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// LocalRuntimeMetric represents a metric snapshot from a local runtime instance
type LocalRuntimeMetric struct {
	ID                uuid.UUID             `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeInstanceID uuid.UUID             `json:"runtime_instance_id" gorm:"column:runtime_instance_id;type:uuid;not null"`
	RuntimeInstance   *LocalRuntimeInstance `json:"runtime_instance,omitempty" gorm:"foreignKey:RuntimeInstanceID"`
	Timestamp         time.Time             `json:"timestamp" gorm:"column:timestamp;not null"`

	// Performance metrics
	MemoryUsage       MemoryStats `json:"memory_usage" gorm:"column:memory_usage;type:jsonb"`
	CPUUsage          float64     `json:"cpu_usage" gorm:"column:cpu_usage;not null"`
	ActiveConnections int         `json:"active_connections" gorm:"column:active_connections;not null"`
	RequestThroughput float64     `json:"request_throughput" gorm:"column:request_throughput;not null"`
	TotalRequests     int64       `json:"total_requests" gorm:"column:total_requests;not null"`
	ErrorRate         float64     `json:"error_rate" gorm:"column:error_rate;not null"`

	// Function execution metrics
	ExecutionCount int64         `json:"execution_count" gorm:"column:execution_count;not null"`
	AverageLatency time.Duration `json:"average_latency" gorm:"column:average_latency;not null;type:bigint"`
	ErrorCount     int64         `json:"error_count" gorm:"column:error_count;not null"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Heap   uint64 `json:"heap" gorm:"column:heap"`
	Stack  uint64 `json:"stack" gorm:"column:stack"`
	System uint64 `json:"system" gorm:"column:system"`
}

// Value implements the driver.Valuer interface for database storage
func (m MemoryStats) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database reading
func (m *MemoryStats) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into MemoryStats", value)
	}

	return json.Unmarshal(bytes, m)
}

// JSONMap represents a JSON map that can be stored in JSONB columns
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database reading
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}

	*j = JSONMap(m)
	return nil
}

// LocalRuntimeHealth represents the health status of a local runtime instance
type LocalRuntimeHealth struct {
	RuntimeInstanceID uuid.UUID             `json:"runtime_instance_id" gorm:"column:runtime_instance_id;type:uuid;primaryKey"`
	RuntimeInstance   *LocalRuntimeInstance `json:"runtime_instance,omitempty" gorm:"foreignKey:RuntimeInstanceID"`
	Timestamp         time.Time             `json:"timestamp" gorm:"column:timestamp;not null"`
	Status            string                `json:"status" gorm:"column:status;not null"` // "healthy", "degraded", "unhealthy"
	ResponseTime      time.Duration         `json:"response_time" gorm:"column:response_time;not null;type:bigint"`
	Checks            JSONMap               `json:"checks" gorm:"column:checks;type:jsonb"`
	Error             *string               `json:"error,omitempty" gorm:"column:error"`
}

func (LocalRuntimeHealth) TableName() string {
	return "local_runtime_health"
}

// EnvironmentVariable represents an environment variable for a function
type EnvironmentVariable struct {
	Key      string `json:"key" db:"key"`
	Value    string `json:"value" db:"value"`
	IsSecret bool   `json:"is_secret" db:"is_secret"`
}

// ScheduleConfig represents a function schedule configuration
type ScheduleConfig struct {
	ID        int64     `json:"id" db:"-"`
	Cron      string    `json:"cron" db:"cron"`
	Timezone  string    `json:"timezone" db:"timezone"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	LastRun   time.Time `json:"last_run" db:"last_run"`
	NextRun   time.Time `json:"next_run" db:"next_run"`
	RunOnDeploy bool   `json:"run_on_deploy" db:"run_on_deploy"`
}

// FunctionConfig represents a user-created function configuration
type FunctionConfig struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	TenantID          uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	AppID             *uuid.UUID             `json:"app_id,omitempty" db:"app_id"`
	Name              string                 `json:"name" db:"name"`
	Providers         []string               `json:"providers" db:"providers"`
	Region            string                 `json:"region" db:"region"`
	Code              string                 `json:"code" db:"code"`
	EnvVars           []EnvironmentVariable  `json:"env_vars" db:"env_vars"`
	Version           string                 `json:"version" db:"version"`
	Status            string                 `json:"status" db:"status"` // "draft", "deploying", "deployed", "failed"
	Schedule          *ScheduleConfig        `json:"schedule,omitempty" db:"schedule"` // Cron schedule configuration
	PlaygroundEnabled bool                   `json:"playground_enabled" db:"playground_enabled"`
	PlaygroundConfig  map[string]interface{} `json:"playground_config" db:"playground_config"`
	Capabilities      []string               `json:"capabilities" db:"capabilities"` // Declared capabilities for sandbox enforcement
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// FunctionDeployment represents a deployment of a function
type FunctionDeployment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	FunctionID   uuid.UUID `json:"function_id" db:"function_id"`
	Version      string    `json:"version" db:"version"`
	Status       string    `json:"status" db:"status"` // "pending", "deploying", "success", "failed"
	Provider     string    `json:"provider" db:"provider"`
	Region       string    `json:"region" db:"region"`
	DeployedURL  *string   `json:"deployed_url,omitempty" db:"deployed_url"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	DurationMs   *int      `json:"duration_ms,omitempty" db:"duration_ms"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// FunctionLog represents a log entry for function operations
type FunctionLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	FunctionID   *uuid.UUID             `json:"function_id,omitempty" db:"function_id"`
	DeploymentID *uuid.UUID             `json:"deployment_id,omitempty" db:"deployment_id"`
	Level        string                 `json:"level" db:"level"` // "info", "warn", "error", "debug"
	Message      string                 `json:"message" db:"message"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	Source       string                 `json:"source" db:"source"` // "deployment", "runtime", "monitoring"
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}
