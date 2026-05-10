package timemachine

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Replay struct {
	ID                      uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID                uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null;index"`
	UserID                  uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	FunctionID              uuid.UUID      `json:"function_id" gorm:"type:uuid;not null;index"`
	WindowStart             time.Time      `json:"window_start" gorm:"not null"`
	WindowEnd               time.Time      `json:"window_end" gorm:"not null"`
	TargetVersionID         uuid.UUID      `json:"target_version_id" gorm:"type:uuid;not null"`
	TargetVersion           string         `json:"target_version" gorm:"not null"`
	MaxExecutions           int            `json:"max_executions" gorm:"default:1000"`
	ReconciliationMode      string         `json:"reconciliation_mode" gorm:"default:dry_run"`
	AutoReconcile           bool           `json:"auto_reconcile" gorm:"default:false"`
	Status                  string         `json:"status" gorm:"default:pending;index"`
	ProgressPercent         float64        `json:"progress_percent" gorm:"default:0"`
	CurrentPhase            sql.NullString `json:"current_phase"`
	ErrorMessage            sql.NullString `json:"error_message"`
	TotalExecutionsFound    int            `json:"total_executions_found" gorm:"default:0"`
	TotalExecutionsReplayed int            `json:"total_executions_replayed" gorm:"default:0"`
	TotalExecutionsChanged  int            `json:"total_executions_changed" gorm:"default:0"`
	TotalExecutionsFailed   int            `json:"total_executions_failed" gorm:"default:0"`
	Reason                  string         `json:"reason" gorm:"not null"`
	IncidentURL             sql.NullString `json:"incident_url"`
	StartedAt               *time.Time     `json:"started_at"`
	CompletedAt             *time.Time     `json:"completed_at"`
	CreatedAt               time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Replay) TableName() string {
	return "time_machine_replays"
}

type ReplayItem struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReplayID              uuid.UUID       `json:"replay_id" gorm:"type:uuid;not null;index"`
	OriginalExecutionID   uuid.UUID       `json:"original_execution_id" gorm:"type:uuid;not null"`
	OriginalInput         json.RawMessage `json:"original_input" gorm:"type:jsonb;not null"`
	OriginalOutput        json.RawMessage `json:"original_output" gorm:"type:jsonb;not null"`
	OriginalVersion       string          `json:"original_version" gorm:"not null"`
	OriginalDurationMs    int             `json:"original_duration_ms" gorm:"not null"`
	OriginalTimestamp     time.Time       `json:"original_timestamp" gorm:"not null"`
	OriginalMEGRootHash   sql.NullString  `json:"original_meg_root_hash"`
	OriginalCertificateID sql.NullString  `json:"original_certificate_id"`
	NewOutput             json.RawMessage `json:"new_output" gorm:"type:jsonb"`
	NewDurationMs         sql.NullInt32   `json:"new_duration_ms"`
	NewMEGRootHash        sql.NullString  `json:"new_meg_root_hash"`
	NewStatusCode         sql.NullInt32   `json:"new_status_code"`
	OutputChanged         sql.NullBool    `json:"output_changed"`
	DiffType              sql.NullString  `json:"diff_type"`
	DiffSummary           sql.NullString  `json:"diff_summary"`
	DiffDetail            json.RawMessage `json:"diff_detail" gorm:"type:jsonb"`
	ReconciliationStatus  string          `json:"reconciliation_status" gorm:"default:pending"`
	ReconciliationActions json.RawMessage `json:"reconciliation_actions" gorm:"type:jsonb"`
	ReconciledAt          *time.Time      `json:"reconciled_at"`
	ReplayError           sql.NullString  `json:"replay_error"`
	ReplayErrorCode       sql.NullString  `json:"replay_error_code"`
	Status                string          `json:"status" gorm:"default:pending;index"`
	CreatedAt             time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ReplayItem) TableName() string {
	return "time_machine_replay_items"
}

type Reconciliation struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReplayID       uuid.UUID       `json:"replay_id" gorm:"type:uuid;not null;index"`
	ReplayItemID   uuid.UUID       `json:"replay_item_id" gorm:"type:uuid;not null"`
	ActionType     string          `json:"action_type" gorm:"not null"`
	TargetResource string          `json:"target_resource" gorm:"not null"`
	OldValue       json.RawMessage `json:"old_value" gorm:"type:jsonb"`
	NewValue       json.RawMessage `json:"new_value" gorm:"type:jsonb"`
	Status         string          `json:"status" gorm:"default:pending;index"`
	AppliedAt      *time.Time      `json:"applied_at"`
	ErrorMessage   sql.NullString  `json:"error_message"`
	DryRun         bool            `json:"dry_run" gorm:"default:false"`
	Reversible     bool            `json:"reversible" gorm:"default:true"`
	ReversalData   json.RawMessage `json:"reversal_data" gorm:"type:jsonb"`
	CreatedAt      time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

func (Reconciliation) TableName() string {
	return "time_machine_reconciliations"
}

type AuditCertificate struct {
	ID                   uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReplayID             uuid.UUID       `json:"replay_id" gorm:"type:uuid;not null;index"`
	CertificateID        string          `json:"certificate_id" gorm:"uniqueIndex;not null"`
	CertJSON             json.RawMessage `json:"cert_json" gorm:"type:jsonb;not null"`
	CertHash             string          `json:"cert_hash" gorm:"not null"`
	PreviousCertHash     sql.NullString  `json:"previous_cert_hash"`
	MerkleRoot           sql.NullString  `json:"merkle_root"`
	Signature            sql.NullString  `json:"signature"`
	ComplianceFrameworks []string        `json:"compliance_frameworks" gorm:"type:text[];default:'{}'"`
	LegalHoldRef         sql.NullString  `json:"legal_hold_ref"`
	RetentionPolicy      string          `json:"retention_policy" gorm:"default:7_years"`
	Anchored             bool            `json:"anchored" gorm:"default:false"`
	AnchorChain          sql.NullString  `json:"anchor_chain"`
	AnchorTxHash         sql.NullString  `json:"anchor_tx_hash"`
	AnchoredAt           *time.Time      `json:"anchored_at"`
	CreatedAt            time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

func (AuditCertificate) TableName() string {
	return "time_machine_audit_certificates"
}

type Schedule struct {
	ID                    uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	FunctionID            uuid.UUID  `json:"function_id" gorm:"type:uuid;not null"`
	CronExpression        string     `json:"cron_expression" gorm:"not null"`
	Timezone              string     `json:"timezone" gorm:"default:UTC"`
	ReplayWindowHours     int        `json:"replay_window_hours" gorm:"default:24"`
	TargetVersionStrategy string     `json:"target_version_strategy" gorm:"default:latest"`
	PinnedVersionID       *uuid.UUID `json:"pinned_version_id" gorm:"type:uuid"`
	ReconciliationMode    string     `json:"reconciliation_mode" gorm:"default:dry_run"`
	AutoReconcile         bool       `json:"auto_reconcile" gorm:"default:false"`
	ReasonTemplate        string     `json:"reason_template" gorm:"default:Scheduled replay"`
	Enabled               bool       `json:"enabled" gorm:"default:true"`
	LastRunAt             *time.Time `json:"last_run_at"`
	NextRunAt             *time.Time `json:"next_run_at"`
	TotalRuns             int        `json:"total_runs" gorm:"default:0"`
	LastReplayID          *uuid.UUID `json:"last_replay_id" gorm:"type:uuid"`
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Schedule) TableName() string {
	return "time_machine_schedules"
}
