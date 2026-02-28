package registry

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EmbedConfig holds per-function embed configuration
type EmbedConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
	RequireAPIKey    bool     `json:"require_api_key"`
	UIEnabled        bool     `json:"ui_enabled"`
	UITheme          string   `json:"ui_theme"` // "light", "dark", "auto"
	RateLimitPerHour int      `json:"rate_limit_per_hour"`
}

// RegistryFunction represents a function in the public registry
type RegistryFunction struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Author             string          `json:"author" gorm:"not null;index"`
	Name               string          `json:"name" gorm:"not null;index"`
	LatestVersion      sql.NullString  `json:"latest_version" gorm:"type:text"`
	Title              sql.NullString  `json:"title" gorm:"type:text"`
	Description        sql.NullString  `json:"description" gorm:"type:text"`
	Category           sql.NullString  `json:"category" gorm:"type:text"`
	Tags               json.RawMessage `json:"tags" gorm:"type:jsonb"`
	Visibility         string          `json:"visibility" gorm:"not null;default:'public'"`
	PricePerCall       float64         `json:"price_per_call" gorm:"default:0"`
	PopularityScore    int             `json:"popularity_score" gorm:"default:0;index"`
	ReliabilityScore   float64         `json:"reliability_score" gorm:"default:0"`
	DeterministicScore float64         `json:"deterministic_score" gorm:"default:0"`
	Capabilities       json.RawMessage `json:"capabilities" gorm:"type:jsonb"` // Declared capabilities for sandbox enforcement
	EmbedConfig        json.RawMessage `json:"embed_config,omitempty" gorm:"type:jsonb"` // Per-function embed configuration
	TenantID           *uuid.UUID      `json:"tenant_id,omitempty" gorm:"type:uuid"`
	OwnerUserID        *uuid.UUID      `json:"owner_user_id,omitempty" gorm:"type:uuid"`
	CreatedAt          time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Versions []RegistryFunctionVersion `json:"versions,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
	Rating   *RegistryFunctionRating   `json:"rating,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
}

// RegistryFunctionVersion represents a specific version of a function
type RegistryFunctionVersion struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID    uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	Version       string          `json:"version" gorm:"not null;index"`
	Manifest      json.RawMessage `json:"manifest" gorm:"type:jsonb;not null"`
	Runtime       string          `json:"runtime" gorm:"not null"`
	TimeoutMs     int             `json:"timeout_ms" gorm:"default:30000"`
	MemoryMB      int             `json:"memory_mb" gorm:"default:128"`
	Deterministic bool            `json:"deterministic" gorm:"default:false"`
	CacheTTL      int             `json:"cache_ttl" gorm:"default:0"`
	// Capabilities declared in manifest - determines sandbox permissions
	Capabilities json.RawMessage `json:"capabilities" gorm:"type:jsonb"`
	// SideEffects indicates what side effects the function has
	// none = no side effects, safe for caching/replay
	// network = makes external network calls
	// external_state = modifies external state (files, databases, etc.)
	SideEffects string `json:"side_effects" gorm:"default:'none'"`
	// Idempotent indicates if the function is idempotent (safe to retry)
	Idempotent   bool           `json:"idempotent" gorm:"default:false"`
	DeploymentID *uuid.UUID     `json:"deployment_id,omitempty" gorm:"type:uuid"`
	BackendID    *uuid.UUID     `json:"backend_id,omitempty" gorm:"type:uuid"`
	ContentHash  sql.NullString `json:"content_hash" gorm:"type:text"`
	WasmBinary   []byte         `json:"wasm_binary,omitempty" gorm:"type:bytea"`
	SourceHash   sql.NullString `json:"source_hash" gorm:"type:text"`
	SourceCode   sql.NullString `json:"source_code,omitempty" gorm:"type:text"` // Source code for lazy bundling
	BundleSize   sql.NullInt32  `json:"bundle_size"`
	PublishedAt  time.Time      `json:"published_at" gorm:"autoCreateTime"`

	// Relationships
	Function           *RegistryFunction                   `json:"function,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
	Signatures         []RegistryFunctionSignature         `json:"signatures,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	MalwareScans       []RegistryFunctionMalwareScan       `json:"malware_scans,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	Approvals          []RegistryFunctionApproval          `json:"approvals,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	VerificationStatus *RegistryFunctionVerificationStatus `json:"verification_status,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
}

// RegistryFunctionExecution represents an execution record
type RegistryFunctionExecution struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID uuid.UUID      `json:"function_id" gorm:"type:uuid;not null;index"`
	Version    string         `json:"version" gorm:"not null"`
	DurationMs int            `json:"duration_ms" gorm:"not null"`
	StatusCode int            `json:"status_code" gorm:"not null"`
	Cached     bool           `json:"cached" gorm:"default:false"`
	Outcome    string         `json:"outcome" gorm:"not null;index"`
	ErrorCode  sql.NullString `json:"error_code" gorm:"type:text"`
	CallerIP   sql.NullString `json:"caller_ip" gorm:"type:text;index"`
	UserAgent  sql.NullString `json:"user_agent" gorm:"type:text"`
	GeoCountry sql.NullString `json:"geo_country" gorm:"type:text;index"`
	TenantID   *uuid.UUID     `json:"tenant_id" gorm:"type:uuid;index"`
	UserID     *uuid.UUID     `json:"user_id" gorm:"type:uuid;index"`
	Timestamp  time.Time      `json:"timestamp" gorm:"autoCreateTime;index"`
	// Embed tracking
	EmbedOrigin sql.NullString `json:"embed_origin,omitempty" gorm:"type:text"` // Domain that triggered embed execution
	// Replay verification fields
	VerifiedAt         sql.NullTime   `json:"verified_at" gorm:"index"`
	VerificationStatus sql.NullString `json:"verification_status" gorm:"type:text"` // "verified", "failed", "pending"
	VerificationError  sql.NullString `json:"verification_error" gorm:"type:text"`
	ReplayedDurationMs sql.NullInt32  `json:"replayed_duration_ms"`
}

// RegistryExecutionPublic represents a shareable execution for the playground/replay feature
type RegistryExecutionPublic struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PublicID   string          `json:"public_id" gorm:"uniqueIndex;not null"`
	FunctionID uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	Version    string          `json:"version" gorm:"not null"`
	InputJSON  json.RawMessage `json:"input_json" gorm:"type:jsonb;not null"`
	OutputJSON json.RawMessage `json:"output_json" gorm:"type:jsonb;not null"`
	DurationMs int             `json:"duration_ms" gorm:"not null"`
	Cached     bool            `json:"cached" gorm:"default:false"`
	Shareable  bool            `json:"shareable" gorm:"default:true"`
	CreatedAt  time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
	// Replay verification fields
	VerifiedAt         sql.NullTime    `json:"verified_at"`
	VerificationStatus sql.NullString  `json:"verification_status" gorm:"type:text"` // "verified", "failed", "pending"
	VerificationError  sql.NullString  `json:"verification_error" gorm:"type:text"`
	ReplayedOutputJSON json.RawMessage `json:"replayed_output_json" gorm:"type:jsonb"`
	ReplayedDurationMs sql.NullInt32   `json:"replayed_duration_ms"`
}

// ExecutionResourceUsage represents resource usage for an execution
type ExecutionResourceUsage struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID *uuid.UUID `json:"execution_id,omitempty" gorm:"type:uuid;index"` // Links to RegistryFunctionExecution

	// Resource limits (from function configuration)
	MaxMemoryMB  int `json:"max_memory_mb"`
	MaxCPUTimeMs int `json:"max_cpu_time_ms"`

	// Actual usage
	MemoryUsedMB   float64 `json:"memory_used_mb"`
	CPUTimeUsedMs  int     `json:"cpu_time_used_ms"`
	WallTimeUsedMs int     `json:"wall_time_used_ms"`

	// Termination reason
	TerminatedBy string `json:"terminated_by,omitempty" gorm:"size:50"` // "timeout", "memory_limit", "cpu_limit", "normal"

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RegistryFunctionRating represents aggregated ratings for a function
type RegistryFunctionRating struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID         uuid.UUID `json:"function_id" gorm:"type:uuid;not null;uniqueIndex"`
	OverallScore       float64   `json:"overall_score" gorm:"default:0"`
	ReliabilityScore   float64   `json:"reliability_score" gorm:"default:0"`
	LatencyScore       float64   `json:"latency_score" gorm:"default:0"`
	DocumentationScore float64   `json:"documentation_score" gorm:"default:0"`
	TotalRatings       int       `json:"total_ratings" gorm:"default:0"`
	SuccessRate        float64   `json:"success_rate" gorm:"default:0"`
	P95LatencyMs       int       `json:"p95_latency_ms" gorm:"default:0"`
	AvgLatencyMs       int       `json:"avg_latency_ms" gorm:"default:0"`
	// Trust Score fields
	P50LatencyMs      int        `json:"p50_latency_ms" gorm:"default:0"`
	TimeoutRate       float64    `json:"timeout_rate" gorm:"default:0"`
	ErrorRate         float64    `json:"error_rate" gorm:"default:0"`
	ConsumerDiversity float64    `json:"consumer_diversity" gorm:"default:0"`
	TenantDiversity   int        `json:"tenant_diversity" gorm:"default:0"`
	UserDiversity     int        `json:"user_diversity" gorm:"default:0"`
	TrustScore        float64    `json:"trust_score" gorm:"default:0;index"`
	TrustUpdatedAt    *time.Time `json:"trust_updated_at" gorm:"type:timestamp"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// DRE v2 sub-scores (added by migration 0004_dre_v2)
	DeterminismScore           float64    `json:"determinism_score" gorm:"default:0"`
	ReplayIntegrityScore       float64    `json:"replay_integrity_score" gorm:"default:0"`
	PerformanceStabilityScore  float64    `json:"performance_stability_score" gorm:"default:0"`
	DriftScore                 float64    `json:"drift_score" gorm:"default:1"`
	TrustScoreV2               float64    `json:"trust_score_v2" gorm:"default:0"`
	TrustV2UpdatedAt           *time.Time `json:"trust_v2_updated_at" gorm:"type:timestamp"`

	// Relationships
	Function *RegistryFunction `json:"function,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
}

// RegistryFunctionSignature represents a digital signature for function content verification
type RegistryFunctionSignature struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionVersionID uuid.UUID      `json:"function_version_id" gorm:"type:uuid;not null;index"`
	Algorithm         string         `json:"algorithm" gorm:"not null;size:50"`   // "rsa-sha256", "ecdsa-p256-sha256", etc.
	KeyID             string         `json:"key_id" gorm:"not null;size:255"`     // Identifier for the signing key
	Signature         string         `json:"signature" gorm:"not null;type:text"` // Base64 encoded signature
	SignedBy          string         `json:"signed_by" gorm:"not null;size:255"`  // Email or identifier of signer
	SignedAt          time.Time      `json:"signed_at" gorm:"autoCreateTime"`
	VerifiedAt        *time.Time     `json:"verified_at,omitempty"`
	VerificationError sql.NullString `json:"verification_error" gorm:"type:text"`
	IsValid           bool           `json:"is_valid" gorm:"default:false"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// RegistryFunctionMalwareScan represents malware scan results for a function version
type RegistryFunctionMalwareScan struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionVersionID uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;index"`
	ScanEngine        string          `json:"scan_engine" gorm:"not null;size:100"` // "clamav", "yara", "custom", etc.
	ScanVersion       string          `json:"scan_version" gorm:"not null;size:50"` // Version of scan engine used
	Status            string          `json:"status" gorm:"not null;size:50"`       // "clean", "suspicious", "malicious", "error"
	ThreatsFound      json.RawMessage `json:"threats_found" gorm:"type:jsonb"`      // JSON array of detected threats
	RiskScore         float64         `json:"risk_score" gorm:"default:0"`          // 0.0-1.0 risk assessment
	ScanMetadata      json.RawMessage `json:"scan_metadata" gorm:"type:jsonb"`      // Additional scan details
	ScannedAt         time.Time       `json:"scanned_at" gorm:"autoCreateTime"`
	ScanDurationMs    int             `json:"scan_duration_ms"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// RegistryFunctionApproval represents an approval/review workflow for high-trust functions
type RegistryFunctionApproval struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionVersionID uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;index"`
	ApprovalType      string          `json:"approval_type" gorm:"not null;size:50"`  // "security_review", "code_review", "compliance"
	RequestedBy       uuid.UUID       `json:"requested_by" gorm:"type:uuid;not null"` // User who requested approval
	Status            string          `json:"status" gorm:"not null;size:50"`         // "pending", "approved", "rejected", "requires_changes"
	Priority          string          `json:"priority" gorm:"not null;size:20"`       // "low", "medium", "high", "critical"
	TrustLevel        string          `json:"trust_level" gorm:"not null;size:20"`    // "standard", "high", "enterprise"
	ReviewDeadline    *time.Time      `json:"review_deadline,omitempty"`
	AssignedTo        *uuid.UUID      `json:"assigned_to,omitempty" gorm:"type:uuid"` // Assigned reviewer
	ApprovedBy        *uuid.UUID      `json:"approved_by,omitempty" gorm:"type:uuid"` // User who approved/rejected
	ApprovedAt        *time.Time      `json:"approved_at,omitempty"`
	ReviewComments    string          `json:"review_comments" gorm:"type:text"`
	RequiredActions   json.RawMessage `json:"required_actions" gorm:"type:jsonb"`  // JSON array of required actions
	CompletedActions  json.RawMessage `json:"completed_actions" gorm:"type:jsonb"` // JSON array of completed actions
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	FunctionVersion *RegistryFunctionVersion          `json:"function_version,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	Comments        []RegistryFunctionApprovalComment `json:"comments,omitempty" gorm:"foreignKey:ApprovalID;references:ID"`
}

// RegistryFunctionApprovalComment represents comments on approval reviews
type RegistryFunctionApprovalComment struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ApprovalID uuid.UUID `json:"approval_id" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Comment    string    `json:"comment" gorm:"type:text;not null"`
	IsInternal bool      `json:"is_internal" gorm:"default:false"` // Internal reviewer comments
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Approval *RegistryFunctionApproval `json:"approval,omitempty" gorm:"foreignKey:ApprovalID;references:ID"`
}

// RegistryFunctionVerificationStatus represents the overall verification status of a function version
type RegistryFunctionVerificationStatus struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionVersionID uuid.UUID `json:"function_version_id" gorm:"type:uuid;not null;uniqueIndex"`
	// Content verification
	ContentHashVerified bool `json:"content_hash_verified" gorm:"default:false"`
	SignatureVerified   bool `json:"signature_verified" gorm:"default:false"`
	// Security verification
	MalwareScanned   bool    `json:"malware_scanned" gorm:"default:false"`
	MalwareStatus    string  `json:"malware_status" gorm:"size:50"` // "clean", "suspicious", "malicious"
	MalwareRiskScore float64 `json:"malware_risk_score" gorm:"default:0"`
	// Approval workflow
	ApprovalRequired bool       `json:"approval_required" gorm:"default:false"`
	ApprovalStatus   string     `json:"approval_status" gorm:"size:50"` // "not_required", "pending", "approved", "rejected"
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	// Overall status
	OverallStatus      string     `json:"overall_status" gorm:"not null;size:50"` // "unverified", "verifying", "verified", "failed", "blocked"
	LastVerifiedAt     *time.Time `json:"last_verified_at,omitempty"`
	NextVerificationAt *time.Time `json:"next_verification_at,omitempty"` // For periodic re-verification
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	FunctionVersion *RegistryFunctionVersion `json:"function_version,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
}

// ============================================
// DRE 2.0 Models
// ============================================

// MEGRecord represents a Merkle Execution Graph record for a single execution.
// One record is created per execution (for deterministic functions).
type MEGRecord struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID uuid.UUID `json:"execution_id" gorm:"type:uuid;not null;index"`
	FunctionID  uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	Version     string    `json:"version" gorm:"not null"`

	// MEG component hashes (DRE/1.0 leaf ordering — fixed forever)
	ExecutionRootHash string `json:"execution_root_hash" gorm:"not null;index"`
	InputHash         string `json:"input_hash" gorm:"not null"`
	EnvironmentHash   string `json:"environment_hash" gorm:"not null"`
	DependencyHash    string `json:"dependency_hash" gorm:"not null"`
	TraceHash         string `json:"trace_hash"`         // empty in lite tier
	ResourceHash      string `json:"resource_hash" gorm:"not null"`
	OutputHash        string `json:"output_hash" gorm:"not null"`
	MetadataHash      string `json:"metadata_hash" gorm:"not null"`

	// Capsule descriptor
	CapsuleDescriptorHash string `json:"capsule_descriptor_hash" gorm:"not null"`
	DeterminismTier       string `json:"determinism_tier" gorm:"not null;default:'full'"` // "full"|"lite"
	ProtocolVersion       string `json:"protocol_version" gorm:"not null;default:'dre/1.0'"`

	// Replay verification state
	ReplayRootHash   string     `json:"replay_root_hash"`
	ReplayVerifiedAt *time.Time `json:"replay_verified_at"`
	ReplayNodeID     string     `json:"replay_node_id"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the database table name for MEGRecord.
func (MEGRecord) TableName() string { return "execution_meg_records" }

// ExecutionCertificate represents a stored FXCERT execution certificate.
type ExecutionCertificate struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CertificateID string          `json:"certificate_id" gorm:"not null;uniqueIndex"` // "fxc_01H..."
	ExecutionID   uuid.UUID       `json:"execution_id" gorm:"type:uuid;not null;index"`
	MEGRecordID   uuid.UUID       `json:"meg_record_id" gorm:"type:uuid;not null"`
	FunctionID    uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`

	CertLevel          string          `json:"cert_level" gorm:"not null;default:'standard'"` // "lite"|"standard"|"legal_grade"
	CertJSON           json.RawMessage `json:"cert_json" gorm:"type:jsonb;not null"`
	ExecutionRootHash  string          `json:"execution_root_hash" gorm:"not null;index"`
	CertificateHash    string          `json:"certificate_hash" gorm:"not null"`

	// Signatures
	NodeSignature     string `json:"node_signature"`
	PlatformSignature string `json:"platform_signature"`

	// Blockchain anchoring (optional)
	Anchored          bool       `json:"anchored" gorm:"default:false"`
	AnchorChain       string     `json:"anchor_chain"`
	AnchorBlockNumber int64      `json:"anchor_block_number"`
	AnchorTxHash      string     `json:"anchor_tx_hash"`
	AnchorMerkleRoot  string     `json:"anchor_merkle_root"`
	AnchoredAt        *time.Time `json:"anchored_at"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the database table name for ExecutionCertificate.
func (ExecutionCertificate) TableName() string { return "execution_certificates" }

// DriftReportRecord represents a stored drift report when replay diverges.
type DriftReportRecord struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID uuid.UUID       `json:"execution_id" gorm:"type:uuid;not null;index"`
	FunctionID  uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	Version     string          `json:"version" gorm:"not null"`

	OriginalRootHash string          `json:"original_root_hash" gorm:"not null"`
	ReplayRootHash   string          `json:"replay_root_hash" gorm:"not null"`
	DriftCategory    string          `json:"drift_category" gorm:"not null"` // capsule.DriftCategory
	ComponentDiff    json.RawMessage `json:"component_diff" gorm:"type:jsonb"`
	TrustPenalty     float64         `json:"trust_penalty" gorm:"default:0"`

	DetectedAt time.Time `json:"detected_at" gorm:"autoCreateTime;index"`
}

// TableName returns the database table name for DriftReportRecord.
func (DriftReportRecord) TableName() string { return "drift_reports" }

// ExecutionPassport represents the per-function aggregate of DRE statistics.
// It is the public-facing "determinism passport" shown on the marketplace.
type ExecutionPassport struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID uuid.UUID `json:"function_id" gorm:"type:uuid;not null;uniqueIndex"`

	// Determinism statistics
	DeterministicReliability float64 `json:"deterministic_reliability" gorm:"default:0"` // 0.0-1.0
	ReplayDriftIncidents     int     `json:"replay_drift_incidents" gorm:"default:0"`
	VerifiedExecutionsTotal  int64   `json:"verified_executions_total" gorm:"default:0"`
	TotalExecutions          int64   `json:"total_executions" gorm:"default:0"`

	// DRE sub-scores (feed into TrustScore v2)
	DeterminismScore           float64 `json:"determinism_score" gorm:"default:0"`
	ReplayIntegrityScore       float64 `json:"replay_integrity_score" gorm:"default:0"`
	PerformanceStabilityScore  float64 `json:"performance_stability_score" gorm:"default:0"`
	DriftScore                 float64 `json:"drift_score" gorm:"default:1"` // 1.0 = no drift

	// Capsule version history
	CapsuleVersionsUsed json.RawMessage `json:"capsule_versions_used" gorm:"type:jsonb"`

	// Metadata
	LastVerifiedAt *time.Time `json:"last_verified_at"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for ExecutionPassport.
func (ExecutionPassport) TableName() string { return "function_execution_passports" }

// PassportUpdate contains the fields to update in an ExecutionPassport.
type PassportUpdate struct {
	IncrementVerified    bool
	IncrementTotal       bool
	IncrementDrift       bool
	TrustPenalty         float64
	CapsuleDescriptorHash string
	LastVerifiedAt       *time.Time
}

// DREScores contains the 4 DRE sub-scores used in TrustScore v2 calculation.
type DREScores struct {
	DeterminismScore          float64 `json:"determinism_score"`
	ReplayIntegrityScore      float64 `json:"replay_integrity_score"`
	PerformanceStabilityScore float64 `json:"performance_stability_score"`
	DriftScore                float64 `json:"drift_score"`
}
