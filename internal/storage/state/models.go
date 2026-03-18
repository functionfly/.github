package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONMap represents flexible JSON data storage
type JSONMap map[string]interface{}

// ============================================
// Core State Models
// ============================================

// State represents a durable state container bound to a function identity
type State struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name       string     `json:"name" gorm:"not null;size:255"`
	FullPath   string     `json:"full_path" gorm:"uniqueIndex;not null;size:500"` // "acme/cart"
	FunctionID *uuid.UUID `json:"function_id,omitempty" gorm:"type:uuid;index"`     // Optional bound function

	// State Configuration
	StorageType string `json:"storage_type" gorm:"not null;default:'keyvalue';size:50"` // "keyvalue" | "document" | "timeseries" | "graph"

	// Retention
	TTLDays   int `json:"ttl_days" gorm:"not null;default:0"`     // 0 = forever
	MaxSizeMB int `json:"max_size_mb" gorm:"not null;default:100"`

	// Versioning
	CurrentVersion int  `json:"current_version" gorm:"not null;default:1"`
	IsVersioned    bool `json:"is_versioned" gorm:"not null;default:true"`

	// Security
	IsEncrypted bool `json:"is_encrypted" gorm:"not null;default:false"` // Enable encryption at rest

	// Permissions
	IsPublic         bool `json:"is_public" gorm:"not null;default:false"`
	AllowCrossTenant bool `json:"allow_cross_tenant" gorm:"not null;default:false"`

	// Metadata
	Description *string `json:"description,omitempty" gorm:"size:1000"`
	Tags        JSONMap `json:"tags" gorm:"type:jsonb;not null;default:'{}'"`

	// Billing
	StorageUsedMB int64 `json:"storage_used_mb" gorm:"not null;default:0"`
	WriteOpsMonth int64 `json:"write_ops_month" gorm:"not null;default:0"`
	ReadOpsMonth  int64 `json:"read_ops_month" gorm:"not null;default:0"`

	// Timestamps
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
	LastAccessedAt time.Time `json:"last_accessed_at" gorm:"not null"`
}

func (State) TableName() string {
	return "states"
}

// BeforeCreate hook for State
func (s *State) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.StorageType == "" {
		s.StorageType = "keyvalue"
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now()
	}
	if s.LastAccessedAt.IsZero() {
		s.LastAccessedAt = time.Now()
	}
	return nil
}

// StateValue represents a single key-value entry in state
type StateValue struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Key (supports hierarchical keys like "user/123/profile")
	Key string `json:"key" gorm:"not null:index;size:500"`

	// Value (JSON for flexibility)
	Value JSONMap `json:"value" gorm:"type:jsonb;not null"`

	// Encryption support
	IsEncrypted bool   `json:"is_encrypted" gorm:"not null;default:false"`
	EncryptedVal []byte `json:"-" gorm:"type:bytea"` // Encrypted value (never serialized to JSON)

	// Versioning
	Version       int      `json:"version" gorm:"not null:index"`
	PreviousValue *JSONMap `json:"previous_value,omitempty" gorm:"type:jsonb"`

	// Content Addressing (for deduplication)
	ContentHash string `json:"content_hash" gorm:"index;size:64"`

	// TTL
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index"`

	// Metadata
	CreatedBy string    `json:"created_by" gorm:"not null;size:255"` // function_id or user_id
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
}

func (StateValue) TableName() string {
	return "state_values"
}

// BeforeCreate hook for StateValue
func (sv *StateValue) BeforeCreate(tx *gorm.DB) error {
	if sv.ID == uuid.Nil {
		sv.ID = uuid.New()
	}
	if sv.CreatedAt.IsZero() {
		sv.CreatedAt = time.Now()
	}
	if sv.ContentHash == "" && sv.Value != nil {
		sv.ContentHash = contentHashJSON(sv.Value)
	}
	return nil
}

// contentHashJSON returns a SHA-256 hex digest of the JSON-serialized value for content addressing and deduplication.
func contentHashJSON(v JSONMap) string {
	if v == nil {
		return ""
	}
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// StateEvent represents an immutable event in state history
type StateEvent struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Event Types: "set" | "delete" | "snapshot" | "restore" | "merge"
	EventType string `json:"event_type" gorm:"not null;index;size:50"`

	// Key affected (null for state-level events)
	Key *string `json:"key,omitempty" gorm:"index;size:500"`

	// Event Data
	PreviousValue *JSONMap `json:"previous_value,omitempty" gorm:"type:jsonb"`
	NewValue      *JSONMap `json:"new_value,omitempty" gorm:"type:jsonb"`

	// Causality
	CausationID   *uuid.UUID `json:"causation_id,omitempty" gorm:"type:uuid;index"`
	CorrelationID string     `json:"correlation_id" gorm:"not null;size:255;index"` // For distributed tracing

	// Source
	SourceType string `json:"source_type" gorm:"not null;size:50;index"` // "function" | "user" | "system" | "trigger"
	SourceID   string `json:"source_id" gorm:"not null;size:255;index"`  // function_id or user_id

	// Determinism Proof (for replay verification)
	InputHash     string `json:"input_hash" gorm:"size:128"`
	OutputHash    string `json:"output_hash" gorm:"size:128"`
	Deterministic bool   `json:"deterministic" gorm:"not null;default:false"`

	// Sequence (for ordering)
	SequenceNum int64 `json:"sequence_num" gorm:"not null;uniqueIndex"`

	Timestamp time.Time `json:"timestamp" gorm:"autoCreateTime;not null;index"`
}

func (StateEvent) TableName() string {
	return "state_events"
}

// BeforeCreate hook for StateEvent
func (se *StateEvent) BeforeCreate(tx *gorm.DB) error {
	if se.ID == uuid.Nil {
		se.ID = uuid.New()
	}
	if se.Timestamp.IsZero() {
		se.Timestamp = time.Now()
	}
	if se.SequenceNum == 0 {
		// Get next sequence number for this state
		var maxSeq int64
		tx.Model(&StateEvent{}).Where("state_id = ?", se.StateID).Select("COALESCE(MAX(sequence_num), 0)").Scan(&maxSeq)
		se.SequenceNum = maxSeq + 1
	}
	return nil
}

// StateSnapshot represents a point-in-time snapshot of state
type StateSnapshot struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Snapshot Identification
	SnapshotVersion int     `json:"snapshot_version" gorm:"not null"`
	Label           *string `json:"label,omitempty"` // Optional human-readable label

	// Content
	StateData      JSONMap `json:"state_data" gorm:"type:jsonb;not null"`
	StateSizeBytes int64   `json:"state_size_bytes"`

	// Coverage
	KeyCount      int   `json:"key_count"`
	FirstSequence int64 `json:"first_sequence"`
	LastSequence  int64 `json:"last_sequence"`

	// Determinism
	RootEventID uuid.UUID `json:"root_event_id" gorm:"type:uuid"` // First event in snapshot

	// Compression
	IsCompressed    bool   `json:"is_compressed" gorm:"default:false"`
	CompressionAlgo string `json:"compression_algo"` // "lz4", "zstd", ""

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (StateSnapshot) TableName() string {
	return "state_snapshots"
}

// StatePermission defines access control for state
type StatePermission struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Principal
	PrincipalType string     `json:"principal_type"` // "user" | "team" | "function" | "tenant"
	PrincipalID   *uuid.UUID `json:"principal_id,omitempty" gorm:"type:uuid"`

	// Permissions
	CanRead    bool `json:"can_read" gorm:"default:false"`
	CanWrite   bool `json:"can_write" gorm:"default:false"`
	CanDelete  bool `json:"can_delete" gorm:"default:false"`
	CanAdmin   bool `json:"can_admin" gorm:"default:false"`
	CanTrigger bool `json:"can_trigger" gorm:"default:false"` // For function triggers

	// Constraints
	IPWhitelist      JSONMap `json:"ip_whitelist" gorm:"type:jsonb;default:'[]'"`
	TimeRestrictions JSONMap `json:"time_restrictions" gorm:"type:jsonb"`
	RateLimit        *int    `json:"rate_limit,omitempty"` // Requests per minute

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (StatePermission) TableName() string {
	return "state_permissions"
}

// StateTrigger defines automatic function invocation on state changes
type StateTrigger struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Source
	SourceStateID *uuid.UUID `json:"source_state_id,omitempty" gorm:"type:uuid"`

	// Trigger Configuration
	TriggerType string  `json:"trigger_type" gorm:"not null"` // "on_write" | "on_read" | "on_delete" | "on_condition"
	KeyPattern  *string `json:"key_pattern,omitempty"`         // Glob pattern for keys

	// Condition (for advanced triggers)
	Condition JSONMap `json:"condition" gorm:"type:jsonb"`

	// Target Function
	TargetFunctionID *uuid.UUID `json:"target_function_id,omitempty" gorm:"type:uuid"`
	TargetFunction   string     `json:"target_function"` // "org/function:version"

	// Payload
	IncludePrevious bool `json:"include_previous" gorm:"default:false"`
	IncludeNew      bool `json:"include_new" gorm:"default:true"`

	// Rate Limiting
	MaxInvocationsPerMinute int `json:"max_invocations_per_minute" gorm:"default:60"`

	// Status
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (StateTrigger) TableName() string {
	return "state_triggers"
}

// StateUsageMetric represents usage data for billing and analytics
type StateUsageMetric struct {
	ID       uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	StateID  *uuid.UUID `json:"state_id,omitempty" gorm:"type:uuid;index"`

	// Metric type
	MetricType string `json:"metric_type" gorm:"not null"` // "storage" | "write_ops" | "read_ops"

	// Value
	Value int64  `json:"value" gorm:"not null"`
	Unit  string `json:"unit"` // "bytes" | "ops" | "mb"

	// Time period
	PeriodStart time.Time `json:"period_start" gorm:"not null"`
	PeriodEnd   time.Time `json:"period_end" gorm:"not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (StateUsageMetric) TableName() string {
	return "state_usage_metrics"
}

// AgentMemory represents AI agent memory with embeddings
type AgentMemory struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	AgentID  string    `json:"agent_id" gorm:"not null;index"`

	// Memory Type: "working" | "longterm" | "context" | "episodic"
	MemoryType string `json:"memory_type" gorm:"not null;index"`

	// Content
	Content        *string `json:"content,omitempty"`
	StructuredData JSONMap `json:"structured_data" gorm:"type:jsonb"`

	// Embedding (pgvector) - stored as byte array for PostgreSQL vector type
	Embedding []float32 `json:"embedding" gorm:"type:vector(1536)"`

	// Metadata
	ImportanceScore float32    `json:"importance_score" gorm:"default:0.5"` // 0.0-1.0 for retention
	AccessCount     int        `json:"access_count" gorm:"default:0"`
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty"`

	// Retention
	TTLDays   int        `json:"ttl_days" gorm:"default:30"`    // Auto-expire after N days
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index"`

	// Provenance
	SourceEventID *uuid.UUID `json:"source_event_id,omitempty" gorm:"type:uuid"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (AgentMemory) TableName() string {
	return "agent_memories"
}
