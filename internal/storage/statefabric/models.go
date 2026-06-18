package statefabric

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONMap for JSONB columns
type JSONMap map[string]interface{}

// Scan implements sql.Scanner so the DB driver can read JSONB into JSONMap.
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	if len(bytes) == 0 {
		*m = JSONMap{}
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}
	*m = out
	return nil
}

// Value implements driver.Valuer so JSONMap is stored as JSON in JSONB columns.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// StateFabric is the top-level fabric container
type StateFabric struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name          string     `gorm:"not null;size:255"`
	Description   string     `gorm:"type:text"`
	Status        string     `gorm:"not null;size:50;default:pending"`
	Type          string     `gorm:"not null;size:50;default:custom"`
	Settings      JSONMap    `gorm:"type:jsonb;not null;default:'{}'"`
	Throughput    int64      `gorm:"not null;default:0"`
	LatencyMs     int64      `gorm:"not null;default:0"`

	// Retention
	TTLDays   int        `gorm:"not null;default:0"` // 0 = forever
	ExpiresAt *time.Time `gorm:"column:expires_at;index"`

	// Timestamps
	LastUpdated  time.Time  `gorm:"autoUpdateTime"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
	SuspendedAt  *time.Time `gorm:"column:suspended_at"`
	SuspendReason *string    `gorm:"column:suspend_reason;type:text"`
}

func (StateFabric) TableName() string { return "state_fabrics" }

// StateFabricStore is a store within a fabric
type StateFabricStore struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Name       string    `gorm:"not null;size:255"`
	Type       string    `gorm:"not null;size:50"`
	Status     string    `gorm:"not null;size:50;default:active"`
	Size       int64     `gorm:"not null;default:0"`
	MaxSize    int64     `gorm:"not null;default:0"`
	Region     string    `gorm:"not null;size:100;default:local"`
	Provider   string    `gorm:"not null;size:100;default:local"`
	Throughput int64     `gorm:"not null;default:0"`
	LatencyMs  int64     `gorm:"not null;default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`

	// R2 Storage Configuration for memory blobs
	R2MemoryBucket  *string `gorm:"column:r2_memory_bucket;type:varchar(255)"`       // R2 bucket for memory storage
	R2MemoryEnabled bool    `gorm:"column:r2_memory_enabled;not null;default:false"` // Whether R2 memory storage is enabled
}

func (StateFabricStore) TableName() string { return "state_fabric_stores" }

// StateFabricPipeline is a pipeline within a fabric
type StateFabricPipeline struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name           string     `gorm:"not null;size:255"`
	Description    string     `gorm:"type:text"`
	Status         string     `gorm:"not null;size:50;default:draft"`
	Steps          JSONMap    `gorm:"type:jsonb;not null;default:'[]'"`
	InputSchema    JSONMap    `gorm:"type:jsonb;column:input_schema"`
	OutputSchema   JSONMap    `gorm:"type:jsonb;column:output_schema"`
	Throughput     int64      `gorm:"not null;default:0"`
	ErrorRate      float64    `gorm:"not null;default:0"`
	LastExecutedAt *time.Time `gorm:"column:last_executed_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`
}

func (StateFabricPipeline) TableName() string { return "state_fabric_pipelines" }

// StateFabricEvent is an event log entry
type StateFabricEvent struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	StoreID        *uuid.UUID `gorm:"type:uuid;index"`
	EventType      string     `gorm:"not null;size:50"`
	Payload        JSONMap    `gorm:"type:jsonb;not null;default:'{}'"`
	SequenceNumber int64      `gorm:"not null"`
	CorrelationID  string     `gorm:"size:255"`
	Timestamp      time.Time  `gorm:"default:now()"`

	// R2 Storage Reference (for archived event batch storage)
	R2ObjectKey *string    `gorm:"column:r2_object_key;type:text;index"`      // R2 key for archived event batch
	R2Bucket    *string    `gorm:"column:r2_bucket;type:varchar(255)"`        // R2 bucket name
	BatchID     *string    `gorm:"column:batch_id;type:varchar(255);index"`   // Batch ID for grouped events in R2
	IsArchived  bool       `gorm:"column:is_archived;not null;default:false"` // Whether event is archived to R2
	ArchivedAt  *time.Time `gorm:"column:archived_at"`                        // When event was archived
}

func (StateFabricEvent) TableName() string { return "state_fabric_events" }

// StateFabricSnapshot is a snapshot of fabric/store state
type StateFabricSnapshot struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	StoreID     *uuid.UUID `gorm:"type:uuid;index"`
	Name        string     `gorm:"not null;size:255"`
	Description string     `gorm:"type:text"`
	StateData   JSONMap    `gorm:"type:jsonb;column:state_data;not null;default:'{}'"`
	EventCount  int        `gorm:"not null;default:0"`
	SizeBytes   int64      `gorm:"not null;default:0"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`

	// R2 Storage Reference (optional - for large snapshot offloading)
	R2ObjectKey   *string `gorm:"column:r2_object_key;type:text;index"`    // R2 key for offloaded snapshot data
	R2Bucket      *string `gorm:"column:r2_bucket;type:varchar(255)"`      // R2 bucket name
	R2ContentHash *string `gorm:"column:r2_content_hash;type:varchar(64)"` // SHA256 hash of R2 content
}

func (StateFabricSnapshot) TableName() string { return "state_fabric_snapshots" }

// StateFabricReplay is a replay session
type StateFabricReplay struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	FabricID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	SnapshotID     *uuid.UUID `gorm:"type:uuid"`
	StartEventID   *uuid.UUID `gorm:"type:uuid;column:start_event_id"`
	EndEventID     *uuid.UUID `gorm:"type:uuid;column:end_event_id"`
	Status         string     `gorm:"not null;size:50;default:pending"`
	Progress       int        `gorm:"not null;default:0"`
	EventsReplayed int64      `gorm:"not null;default:0"`
	StartedAt      time.Time  `gorm:"autoCreateTime"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`
	ErrorMessage   *string    `gorm:"column:error_message;type:text"`

	// R2 Storage Reference (for replay data storage)
	R2ObjectKey   *string `gorm:"column:r2_object_key;type:text;index"`    // R2 key for replay session data
	R2Bucket      *string `gorm:"column:r2_bucket;type:varchar(255)"`      // R2 bucket name
	R2ContentHash *string `gorm:"column:r2_content_hash;type:varchar(64)"` // SHA256 hash of R2 content
}

func (StateFabricReplay) TableName() string { return "state_fabric_replays" }

// StateFabricPipelineExecution records a pipeline run
type StateFabricPipelineExecution struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PipelineID uuid.UUID `gorm:"type:uuid;not null;index"`
	Status     string    `gorm:"not null;size:50;default:pending"`
	InputData  JSONMap   `gorm:"type:jsonb;column:input_data"`
	OutputData JSONMap   `gorm:"type:jsonb;column:output_data"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (StateFabricPipelineExecution) TableName() string { return "state_fabric_pipeline_executions" }

// StateFabricDeadLetter represents a failed operation that couldn't be retried
type StateFabricDeadLetter struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	PipelineID    *uuid.UUID `gorm:"type:uuid;index"`
	OperationType string     `gorm:"not null;size:50"` // "pipeline_execution", "snapshot", "replay", "key_operation"
	InputData     JSONMap    `gorm:"type:jsonb;column:input_data"`
	ErrorMessage  string     `gorm:"type:text"`
	ErrorCode    string     `gorm:"size:50"`
	Attempts      int        `gorm:"not null;default:0"`
	MaxAttempts   int        `gorm:"not null;default:3"`
	Status        string     `gorm:"not null;size:50;default:pending"` // "pending", "retrying", "failed", "resolved"
	NextRetryAt   *time.Time `gorm:"column:next_retry_at;index"`
	ResolvedAt    *time.Time `gorm:"column:resolved_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
}

func (StateFabricDeadLetter) TableName() string { return "state_fabric_dead_letters" }

func (s *StateFabricDeadLetter) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// StateFabricVersion represents a version snapshot of fabric configuration
type StateFabricVersion struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	VersionNumber int       `gorm:"not null"`
	Name          string    `gorm:"size:255"`
	Description   string    `gorm:"type:text"`
	Type          string    `gorm:"size:50"`
	Settings      JSONMap   `gorm:"type:jsonb"`
	ChangeType    string    `gorm:"size:50"` // "create", "update", "delete"
	ChangeSummary string    `gorm:"type:text"`
	ActorID       uuid.UUID `gorm:"type:uuid"`
	ActorType     string    `gorm:"size:50"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}

func (StateFabricVersion) TableName() string { return "state_fabric_versions" }

func (s *StateFabricVersion) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// StateFabricKeyVersion represents a version of a key value
type StateFabricKeyVersion struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Key           string    `gorm:"not null;size:1024;index"`
	Value         JSONMap   `gorm:"type:jsonb"`
	VersionNumber int       `gorm:"not null"`
	ChangeType    string    `gorm:"size:50"` // "set", "delete"
	ChangeSummary string    `gorm:"type:text"`
	TTLSeconds    int       `gorm:"default:0"`
	ExpiresAt     *time.Time `gorm:"column:expires_at;index"`
	ActorID       uuid.UUID `gorm:"type:uuid"`
	ActorType     string    `gorm:"size:50"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}

func (StateFabricKeyVersion) TableName() string { return "state_fabric_key_versions" }

func (s *StateFabricKeyVersion) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets UUIDs
func (s *StateFabric) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *StateFabricStore) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *StateFabricPipeline) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *StateFabricEvent) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *StateFabricSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *StateFabricReplay) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *StateFabricPipelineExecution) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
