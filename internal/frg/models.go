// Package frg implements the Function Registry + Live Runtime Graph system.
// It provides versioned, composable function graphs with streaming execution,
// DRE (Deterministic Replay Execution) support, and AI-powered optimization.
package frg

import (
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// ExecutionMode defines how a graph executes
type ExecutionMode string

const (
	ExecutionModeSync        ExecutionMode = "sync"         // Request/response
	ExecutionModeAsync       ExecutionMode = "async"        // Fire-and-forget with execution ID
	ExecutionModeStreaming   ExecutionMode = "streaming"    // Continuous streaming between nodes
	ExecutionModeEventDriven ExecutionMode = "event_driven" // Trigger-based reactive execution
)

// GraphDefinition represents a versioned, forkable function graph
type GraphDefinition struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Author   string    `json:"author" gorm:"not null;index"`
	Name     string    `json:"name" gorm:"not null;index"`
	Version  string    `json:"version" gorm:"not null"`
	FullName string    `json:"full_name" gorm:"-"` // author/name@version

	// Graph structure
	NodeRefs json.RawMessage `json:"node_refs" gorm:"type:jsonb;not null"` // []GraphNodeRef
	Edges    json.RawMessage `json:"edges" gorm:"type:jsonb;not null"`     // []GraphEdge

	// Execution configuration
	ExecutionMode ExecutionMode   `json:"execution_mode" gorm:"not null;default:'sync'"`
	TriggerConfig json.RawMessage `json:"trigger_config,omitempty" gorm:"type:jsonb"`
	InputSchema   json.RawMessage `json:"input_schema,omitempty" gorm:"type:jsonb"`
	OutputSchema  json.RawMessage `json:"output_schema,omitempty" gorm:"type:jsonb"`

	// AI/Discovery metadata
	AIDescription    string  `json:"ai_description,omitempty" gorm:"type:text"`
	AIEmbedding      []byte  `json:"-" gorm:"type:vector(384)"` // pgvector, not serialized
	CompositionScore float64 `json:"composition_score" gorm:"default:0"`

	// Trust/DRE
	TrustScore    float64         `json:"trust_score" gorm:"default:0"`
	Deterministic bool            `json:"deterministic" gorm:"default:false"`
	DREPassport   json.RawMessage `json:"dre_passport,omitempty" gorm:"type:jsonb"`

	// Fork lineage
	ForkedFromAuthor  *string `json:"forked_from_author,omitempty" gorm:"type:varchar(50)"`
	ForkedFromName    *string `json:"forked_from_name,omitempty" gorm:"type:varchar(100)"`
	ForkedFromVersion *string `json:"forked_from_version,omitempty" gorm:"type:varchar(20)"`

	// Ownership
	TenantID    *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid"`
	OwnerUserID *uuid.UUID `json:"owner_user_id,omitempty" gorm:"type:uuid"`
	Visibility  string     `json:"visibility" gorm:"not null;default:'public'"`

	// Monetization
	PricingType  string  `json:"pricing_type" gorm:"not null;default:'free'"`
	BasePrice    float64 `json:"base_price" gorm:"default:0"`
	RevenueShare float64 `json:"revenue_share" gorm:"default:80.00"` // Creator keeps 80%

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	PublishedAt *time.Time `json:"published_at,omitempty" gorm:"type:timestamptz"`
}

// TableName returns the GORM table name
func (GraphDefinition) TableName() string {
	return "graph_definitions"
}

// GraphNodeRef references a function in the graph
type GraphNodeRef struct {
	NodeID   string                 `json:"node_id"`  // Unique within this graph
	Author   string                 `json:"author"`   // Function author
	Name     string                 `json:"name"`     // Function name
	Version  string                 `json:"version"`  // Specific version or "latest"
	Config   map[string]interface{} `json:"config"`   // Node-specific configuration
	Metadata map[string]interface{} `json:"metadata"` // Node metadata
}

// GraphEdge represents a data flow between nodes
type GraphEdge struct {
	ID             string       `json:"id"` // Unique edge ID
	SourceNodeID   string       `json:"source_node_id"`
	TargetNodeID   string       `json:"target_node_id"`
	Mapping        DataMapping  `json:"mapping"`             // Output→input transformation
	Condition      *Condition   `json:"condition,omitempty"` // Optional conditional routing
	Type           EdgeType     `json:"type"`                // sync, async, stream
	RetryPolicy    *RetryPolicy `json:"retry_policy,omitempty"`
	FallbackNodeID *string      `json:"fallback_node_id,omitempty"`
	BufferSize     int          `json:"buffer_size,omitempty"` // For stream buffering
}

// EdgeType defines how data flows between nodes
type EdgeType string

const (
	EdgeTypeSync   EdgeType = "sync"   // Request/response, waits for completion
	EdgeTypeAsync  EdgeType = "async"  // Fire-and-forget
	EdgeTypeStream EdgeType = "stream" // Continuous streaming
)

// DataMapping transforms output to input between nodes
type DataMapping struct {
	SourcePath string `json:"source_path,omitempty"` // JSONPath or "*" for all
	TargetPath string `json:"target_path,omitempty"` // JSONPath in target input
	Transform  string `json:"transform,omitempty"`   // "map", "filter", "reduce", "flat"
	Script     string `json:"script,omitempty"`      // Custom transformation script
}

// Condition for conditional routing
type Condition struct {
	Operator string      `json:"operator"` // "eq", "ne", "gt", "lt", "contains", "regex", "exists"
	Field    string      `json:"field"`    // JSONPath to field
	Value    interface{} `json:"value"`    // Comparison value
}

// RetryPolicy defines retry behavior for edge failures
type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialBackoff  time.Duration `json:"initial_backoff"`
	MaxBackoff      time.Duration `json:"max_backoff"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"` // Error codes to retry
}

// TriggerConfig defines what triggers graph execution
type TriggerConfig struct {
	Type   string                 `json:"type"` // "webhook", "schedule", "state_trigger", "manual"
	Config map[string]interface{} `json:"config"`
}

// GraphInstance represents a live running graph
type GraphInstance struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DefinitionID uuid.UUID `json:"definition_id" gorm:"type:uuid;not null;index"`

	// Runtime state
	Status       InstanceStatus  `json:"status" gorm:"not null;default:'pending'"`
	InputData    json.RawMessage `json:"input_data,omitempty" gorm:"type:jsonb"`
	OutputData   json.RawMessage `json:"output_data,omitempty" gorm:"type:jsonb"`
	ErrorMessage *string         `json:"error_message,omitempty" gorm:"type:text"`

	// Frozen snapshot of definition at start time
	FrozenNodes json.RawMessage `json:"frozen_nodes,omitempty" gorm:"type:jsonb"`
	FrozenEdges json.RawMessage `json:"frozen_edges,omitempty" gorm:"type:jsonb"`

	// Live runtime tracking
	NodeStates            json.RawMessage `json:"node_states,omitempty" gorm:"type:jsonb"`             // map[node_id]NodeState
	CurrentExecutionOrder json.RawMessage `json:"current_execution_order,omitempty" gorm:"type:jsonb"` // []string

	// DRE traceability
	ExecutionRootHash string     `json:"execution_root_hash,omitempty" gorm:"type:varchar(64)"`
	MEGRecordID       *uuid.UUID `json:"meg_record_id,omitempty" gorm:"type:uuid"`

	// Event streaming
	EventStreamID string     `json:"event_stream_id,omitempty" gorm:"type:varchar(100)"`
	LastEventAt   *time.Time `json:"last_event_at,omitempty" gorm:"type:timestamptz"`

	// State namespace for persistence
	StateNamespace *uuid.UUID `json:"state_namespace,omitempty" gorm:"type:uuid"`

	// Resource tracking
	TotalDurationMs   int     `json:"total_duration_ms" gorm:"default:0"`
	TotalComputeUnits float64 `json:"total_compute_units" gorm:"default:0"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	StartedAt   *time.Time `json:"started_at,omitempty" gorm:"type:timestamptz"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:timestamptz"`
}

// TableName returns the GORM table name
func (GraphInstance) TableName() string {
	return "graph_instances"
}

// InstanceStatus represents the runtime state of a graph
type InstanceStatus string

const (
	InstanceStatusPending   InstanceStatus = "pending"
	InstanceStatusRunning   InstanceStatus = "running"
	InstanceStatusStreaming InstanceStatus = "streaming"
	InstanceStatusPaused    InstanceStatus = "paused"
	InstanceStatusCompleted InstanceStatus = "completed"
	InstanceStatusFailed    InstanceStatus = "failed"
)

// NodeState tracks the execution state of a node within an instance
type NodeState struct {
	Status            string          `json:"status"` // "pending", "executing", "waiting", "completed", "failed", "retrying", "failed_with_fallback"
	Output            json.RawMessage `json:"output,omitempty"`
	Error             *string         `json:"error,omitempty"`
	AttemptCount      int             `json:"attempt_count"`
	ExecCertID        *uuid.UUID      `json:"exec_cert_id,omitempty"`
	DurationMs        int             `json:"duration_ms"`
	FallbackTriggered *bool           `json:"fallback_triggered,omitempty"`
	FallbackNodeID    *string         `json:"fallback_node_id,omitempty"`
}

// GraphNodeExecution tracks individual node runs
type GraphNodeExecution struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InstanceID uuid.UUID `json:"instance_id" gorm:"type:uuid;not null;index"`
	NodeID     string    `json:"node_id" gorm:"not null"`

	// Function reference
	FunctionAuthor  string `json:"function_author" gorm:"not null"`
	FunctionName    string `json:"function_name" gorm:"not null"`
	FunctionVersion string `json:"function_version" gorm:"not null"`

	// Execution
	Status       string          `json:"status" gorm:"not null;default:'pending'"`
	InputData    json.RawMessage `json:"input_data,omitempty" gorm:"type:jsonb"`
	OutputData   json.RawMessage `json:"output_data,omitempty" gorm:"type:jsonb"`
	ErrorMessage *string         `json:"error_message,omitempty" gorm:"type:text"`

	// Retry tracking
	AttemptCount int `json:"attempt_count" gorm:"default:0"`
	MaxAttempts  int `json:"max_attempts" gorm:"default:3"`

	// DRE
	ExecutionCertID *uuid.UUID `json:"execution_cert_id,omitempty" gorm:"type:uuid"`

	// Performance
	DurationMs   int     `json:"duration_ms"`
	ComputeUnits float64 `json:"compute_units"`
	MemoryPeakMB int     `json:"memory_peak_mb"`

	// Streaming state
	StreamPosition int64  `json:"stream_position" gorm:"default:0"`
	StreamChecksum string `json:"stream_checksum,omitempty"`

	// Timestamps
	StartedAt   *time.Time `json:"started_at,omitempty" gorm:"type:timestamptz"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (GraphNodeExecution) TableName() string {
	return "graph_node_executions"
}

// GraphEdgeExecution tracks data flow between nodes
type GraphEdgeExecution struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InstanceID   uuid.UUID `json:"instance_id" gorm:"type:uuid;not null;index"`
	SourceNodeID string    `json:"source_node_id" gorm:"not null"`
	TargetNodeID string    `json:"target_node_id" gorm:"not null"`

	// Execution
	Status   string   `json:"status" gorm:"not null;default:'idle'"`
	EdgeType EdgeType `json:"edge_type" gorm:"not null"`

	// Data transfer metrics
	RecordsTransferred int        `json:"records_transferred" gorm:"default:0"`
	BytesTransferred   int64      `json:"bytes_transferred" gorm:"default:0"`
	LastDataAt         *time.Time `json:"last_data_at,omitempty" gorm:"type:timestamptz"`

	// Retry/fallback
	RetryCount     int     `json:"retry_count" gorm:"default:0"`
	FallbackUsed   bool    `json:"fallback_used" gorm:"default:false"`
	FallbackNodeID *string `json:"fallback_node_id,omitempty"`
	ErrorMessage   *string `json:"error_message,omitempty" gorm:"type:text"`

	// Timestamps
	StartedAt   *time.Time `json:"started_at,omitempty" gorm:"type:timestamptz"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (GraphEdgeExecution) TableName() string {
	return "graph_edge_executions"
}

// GraphEvent represents an event in the event stream
type GraphEvent struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InstanceID uuid.UUID `json:"instance_id" gorm:"type:uuid;not null;index"`

	EventType string  `json:"event_type" gorm:"not null"` // trigger, complete, error, stream, checkpoint, retry
	NodeID    *string `json:"node_id,omitempty"`

	// Payload
	Payload     json.RawMessage `json:"payload,omitempty" gorm:"type:jsonb"`
	ContentType string          `json:"content_type" gorm:"default:'json'"`

	// Ordering
	SequenceNum int64     `json:"sequence_num" gorm:"not null"`
	Timestamp   time.Time `json:"timestamp" gorm:"autoCreateTime"`

	// DRE
	InputHash  *string `json:"input_hash,omitempty"`
	OutputHash *string `json:"output_hash,omitempty"`
}

// TableName returns the GORM table name
func (GraphEvent) TableName() string {
	return "graph_events"
}

// GraphOptimizationSuggestion stores AI-generated recommendations
type GraphOptimizationSuggestion struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DefinitionID uuid.UUID `json:"definition_id" gorm:"type:uuid;not null;index"`

	SuggestionType  string  `json:"suggestion_type" gorm:"not null"` // parallel, cache, runtime, replacement, structure
	Description     string  `json:"description" gorm:"not null"`
	EstimatedImpact float64 `json:"estimated_impact" gorm:"not null"` // % improvement

	// Action
	ActionConfig json.RawMessage `json:"action_config" gorm:"type:jsonb;not null"`

	// AI metadata
	AIConfidence float64   `json:"ai_confidence" gorm:"default:0.5"`
	GeneratedAt  time.Time `json:"generated_at" gorm:"autoCreateTime"`

	// User interaction
	Dismissed   bool       `json:"dismissed" gorm:"default:false"`
	DismissedAt *time.Time `json:"dismissed_at,omitempty" gorm:"type:timestamptz"`
	DismissedBy *uuid.UUID `json:"dismissed_by,omitempty" gorm:"type:uuid"`
	Applied     bool       `json:"applied" gorm:"default:false"`
	AppliedAt   *time.Time `json:"applied_at,omitempty" gorm:"type:timestamptz"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (GraphOptimizationSuggestion) TableName() string {
	return "graph_optimization_suggestions"
}

// ExecutionResult represents the outcome of a graph execution
type ExecutionResult struct {
	InstanceID        uuid.UUID                       `json:"instance_id"`
	Status            InstanceStatus                  `json:"status"`
	Output            json.RawMessage                 `json:"output,omitempty"`
	Error             *string                         `json:"error,omitempty"`
	NodeResults       map[string]*NodeExecutionResult `json:"node_results,omitempty"`
	ExecutionRootHash string                          `json:"execution_root_hash,omitempty"`
	DurationMs        int                             `json:"duration_ms"`
	ComputeUnits      float64                         `json:"compute_units"`
}

// NodeExecutionResult represents a single node's execution result
type NodeExecutionResult struct {
	Status     string          `json:"status"`
	Output     json.RawMessage `json:"output,omitempty"`
	Error      *string         `json:"error,omitempty"`
	DurationMs int             `json:"duration_ms"`
	CertID     *uuid.UUID      `json:"cert_id,omitempty"`
}

// CostBreakdown details the cost of a graph execution
type CostBreakdown struct {
	Total            float64            `json:"total"`
	NodeCosts        map[string]float64 `json:"node_costs"`       // Per-node costs
	PlatformFee      float64            `json:"platform_fee"`     // 15%
	CreatorEarnings  float64            `json:"creator_earnings"` // 85%
	RuntimeCost      float64            `json:"runtime_cost"`     // CPU/memory
	DataTransferCost float64            `json:"data_transfer_cost,omitempty"`
}

// GraphRuntime represents the live runtime state of a graph (in-memory only)
type GraphRuntime struct {
	Instance      *GraphInstance
	Definition    *GraphDefinition
	Nodes         map[string]*RuntimeNode
	Edges         []*RuntimeEdge
	InputChannel  chan *GraphEvent
	OutputChannel chan *ExecutionResult
	EventStream   EventStream
}

// RuntimeNode wraps a node with runtime state
type RuntimeNode struct {
	Ref          *GraphNodeRef
	Definition   *registry.RegistryFunctionVersion // Resolved function
	State        *NodeState
	InputBuffer  chan interface{}
	OutputBuffer chan interface{}
}

// RuntimeEdge wraps an edge with runtime state
type RuntimeEdge struct {
	Definition *GraphEdge
	Buffer     chan interface{}
}

// EventStream interface for event bus (NATS/Kafka abstraction)
type EventStream interface {
	Publish(event *GraphEvent) error
	Subscribe(instanceID uuid.UUID, handler func(*GraphEvent)) error
	Close() error
}
