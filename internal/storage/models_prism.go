package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PrismCell represents a registered Prism execution cell.
type PrismCell struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CellID         string          `gorm:"type:text;not null;uniqueIndex" json:"cell_id"`
	Name           string          `gorm:"type:text;not null" json:"name"`
	Runtime        string          `gorm:"type:text;not null;default:'prism'" json:"runtime"`
	Status         string          `gorm:"type:text;not null;default:'registered'" json:"status"`
	Capabilities   []string        `gorm:"type:text[];default:'{}'" json:"capabilities"`
	Metadata       json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	TenantID       *uuid.UUID      `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	RegisteredAt   time.Time       `gorm:"not null;default:now()" json:"registered_at"`
	LastHeartbeat  *time.Time      `json:"last_heartbeat,omitempty"`
	TerminatedAt   *time.Time      `json:"terminated_at,omitempty"`
	CreatedAt      time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"not null;default:now()" json:"updated_at"`
}

func (PrismCell) TableName() string { return "prism_cells" }

// PrismHeartbeat represents a heartbeat received from a Prism cell.
type PrismHeartbeat struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CellID           string    `gorm:"type:text;not null;index" json:"cell_id"`
	Status           string    `gorm:"type:text;not null" json:"status"`
	ActiveExecutions int       `gorm:"not null;default:0" json:"active_executions"`
	ReceivedAt       time.Time `gorm:"not null;default:now()" json:"received_at"`
}

func (PrismHeartbeat) TableName() string { return "prism_heartbeats" }

// PrismExecutionResult represents an execution result from a Prism cell.
type PrismExecutionResult struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID string          `gorm:"type:text;not null;index" json:"execution_id"`
	CellID      string          `gorm:"type:text;not null;index" json:"cell_id"`
	Status      string          `gorm:"type:text;not null" json:"status"`
	Error       string          `gorm:"type:text" json:"error,omitempty"`
	Result      json.RawMessage `gorm:"type:jsonb" json:"result,omitempty"`
	DurationMs  *int            `json:"duration_ms,omitempty"`
	ReceivedAt  time.Time       `gorm:"not null;default:now()" json:"received_at"`
}

func (PrismExecutionResult) TableName() string { return "prism_execution_results" }

// PrismCapability represents a capability announced by a Prism cell.
type PrismCapability struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CellID      string          `gorm:"type:text;not null" json:"cell_id"`
	Capability  string          `gorm:"type:text;not null" json:"capability"`
	TrustScore  float64         `gorm:"not null;default:0" json:"trust_score"`
	Metadata    json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	AnnouncedAt time.Time       `gorm:"not null;default:now()" json:"announced_at"`
}

func (PrismCapability) TableName() string { return "prism_capabilities" }

// PrismRuntimeStatus represents a runtime status report.
type PrismRuntimeStatus struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Healthy      bool            `gorm:"not null;default:true" json:"healthy"`
	ActiveCells  int             `gorm:"not null;default:0" json:"active_cells"`
	ActiveSwarms int             `gorm:"not null;default:0" json:"active_swarms"`
	Metadata     json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	ReceivedAt   time.Time       `gorm:"not null;default:now()" json:"received_at"`
}

func (PrismRuntimeStatus) TableName() string { return "prism_runtime_status" }
