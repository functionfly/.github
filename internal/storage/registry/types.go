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
	ID                   uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Author               string          `json:"author" gorm:"not null;index"`
	Name                 string          `json:"name" gorm:"not null;index"`
	LatestVersion        sql.NullString  `json:"latest_version" gorm:"type:text"`
	Title                sql.NullString  `json:"title" gorm:"type:text"`
	Description          sql.NullString  `json:"description" gorm:"type:text"`
	Category             sql.NullString  `json:"category" gorm:"type:text"`
	Tags                 json.RawMessage `json:"tags" gorm:"type:jsonb"`
	Visibility           string          `json:"visibility" gorm:"not null;default:'public'"`
	PricePerCall         float64         `json:"price_per_call" gorm:"default:0"`
	PopularityScore      int             `json:"popularity_score" gorm:"default:0;index"`
	ReliabilityScore     float64         `json:"reliability_score" gorm:"default:0"`
	DeterministicScore   float64         `json:"deterministic_score" gorm:"default:0"`
	Capabilities         json.RawMessage `json:"capabilities" gorm:"type:jsonb"`           // Declared capabilities for sandbox enforcement
	EmbedConfig          json.RawMessage `json:"embed_config,omitempty" gorm:"type:jsonb"` // Per-function embed configuration
	Settings             json.RawMessage `json:"settings,omitempty" gorm:"type:jsonb"`     // Per-function settings (custom_domains, etc.)
	TenantID             *uuid.UUID      `json:"tenant_id,omitempty" gorm:"type:uuid"`
	OwnerUserID          *uuid.UUID      `json:"owner_user_id,omitempty" gorm:"type:uuid"`
	PlatformFeePaid      bool            `json:"platform_fee_paid" gorm:"default:false"`
	PlatformFeeAmountUSD float64         `json:"platform_fee_amount_usd" gorm:"default:0"`
	LastFeeChargedAt     *time.Time      `json:"last_fee_charged_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt            time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	// Trust Score fields (denormalized from trust_history for quick access)
	TrustScore              float64    `json:"trust_score" gorm:"default:0"`
	TrustTier               TrustTier  `json:"trust_tier" gorm:"size:20;default:'untrusted'"`
	TrustUpdatedAt          *time.Time `json:"trust_updated_at,omitempty" gorm:"type:timestamptz"`
	TrustCalculationVersion int        `json:"trust_calculation_version" gorm:"default:0"`

	// Relationships
	Versions []RegistryFunctionVersion `json:"versions,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
	Rating   *RegistryFunctionRating   `json:"rating,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
}

// TableName returns the database table name for RegistryFunction.
func (RegistryFunction) TableName() string {
	return "registry_functions"
}

// PlatformFee - audit trail for all platform fees
type PlatformFee struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID      uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	FeeType         string    `json:"fee_type" gorm:"not null;index"` // 'publish', 'version_update', 'commission'
	AmountUSD       float64   `json:"amount_usd" gorm:"type:decimal(14,4);not null"`
	ChargedAt       time.Time `json:"charged_at" gorm:"type:timestamptz;not null;index"`
	StripePaymentID string    `json:"stripe_payment_id,omitempty" gorm:"type:text;index"`
	Status          string    `json:"status" gorm:"not null;default:'pending';index"` // 'pending', 'completed', 'failed', 'refunded'
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Function *RegistryFunction `json:"function,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
}

// TableName returns the legacy publish/version/commission audit table (see 20260402000000_revenue_phase1).
func (PlatformFee) TableName() string {
	return "platform_fees_legacy_publish_audit"
}

// UserWallet - user balance for fee payments
type UserWallet struct {
	UserID              uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	BalanceUSD          float64   `json:"balance_usd" gorm:"type:decimal(14,4);not null;default:0"`
	LifetimeEarningsUSD float64   `json:"lifetime_earnings_usd" gorm:"type:decimal(14,4);not null;default:0"`
	LifetimeFeesUSD     float64   `json:"lifetime_fees_usd" gorm:"type:decimal(14,4);not null;default:0"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the database table name for UserWallet.
func (UserWallet) TableName() string {
	return "user_wallets"
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
	WasmCompiled []byte         `json:"wasm_compiled,omitempty" gorm:"type:bytea"` // AOT-compiled module bytes (.cwasm)
	PublishedAt  time.Time      `json:"published_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	IsActive     bool           `json:"is_active" gorm:"default:true;index"`

	// Relationships
	Function           *RegistryFunction                   `json:"function,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
	Signatures         []RegistryFunctionSignature         `json:"signatures,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	MalwareScans       []RegistryFunctionMalwareScan       `json:"malware_scans,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	Approvals          []RegistryFunctionApproval          `json:"approvals,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
	VerificationStatus *RegistryFunctionVerificationStatus `json:"verification_status,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
}

// TableName returns the database table name for RegistryFunctionVersion.
func (RegistryFunctionVersion) TableName() string {
	return "registry_function_versions"
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
	DeterminismScore          float64    `json:"determinism_score" gorm:"default:0"`
	ReplayIntegrityScore      float64    `json:"replay_integrity_score" gorm:"default:0"`
	PerformanceStabilityScore float64    `json:"performance_stability_score" gorm:"default:0"`
	DriftScore                float64    `json:"drift_score" gorm:"default:1"`
	TrustScoreV2              float64    `json:"trust_score_v2" gorm:"default:0"`
	TrustV2UpdatedAt          *time.Time `json:"trust_v2_updated_at" gorm:"type:timestamp"`

	// Relationships
	Function *RegistryFunction `json:"function,omitempty" gorm:"foreignKey:FunctionID;references:ID"`
}

// TableName returns the database table name for RegistryFunctionRating.
func (RegistryFunctionRating) TableName() string {
	return "registry_function_ratings"
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
	// Blocking information
	BlockReason string     `json:"block_reason,omitempty" gorm:"type:text"` // Reason for blocking
	BlockedAt   *time.Time `json:"blocked_at,omitempty"`                    // When the function was blocked
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	FunctionVersion *RegistryFunctionVersion `json:"function_version,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
}

// TableName returns the database table name for RegistryFunctionVerificationStatus.
func (RegistryFunctionVerificationStatus) TableName() string {
	return "registry_function_verification_status"
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
	TraceHash         string `json:"trace_hash"` // empty in lite tier
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
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CertificateID string    `json:"certificate_id" gorm:"not null;uniqueIndex"` // "fxc_01H..."
	ExecutionID   uuid.UUID `json:"execution_id" gorm:"type:uuid;not null;index"`
	MEGRecordID   uuid.UUID `json:"meg_record_id" gorm:"type:uuid;not null"`
	FunctionID    uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`

	CertLevel         string          `json:"cert_level" gorm:"not null;default:'standard'"` // "lite"|"standard"|"legal_grade"
	CertJSON          json.RawMessage `json:"cert_json" gorm:"type:jsonb;not null"`
	ExecutionRootHash string          `json:"execution_root_hash" gorm:"not null;index"`
	CertificateHash   string          `json:"certificate_hash" gorm:"not null"`

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
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID uuid.UUID `json:"execution_id" gorm:"type:uuid;not null;index"`
	FunctionID  uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	Version     string    `json:"version" gorm:"not null"`

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
	DeterminismScore          float64 `json:"determinism_score" gorm:"default:0"`
	ReplayIntegrityScore      float64 `json:"replay_integrity_score" gorm:"default:0"`
	PerformanceStabilityScore float64 `json:"performance_stability_score" gorm:"default:0"`
	DriftScore                float64 `json:"drift_score" gorm:"default:1"` // 1.0 = no drift

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
	IncrementVerified     bool
	IncrementTotal        bool
	IncrementDrift        bool
	TrustPenalty          float64
	CapsuleDescriptorHash string
	LastVerifiedAt        *time.Time
	ResourceHash          string // Resource hash for performance stability tracking
}

// DREScores contains the 4 DRE sub-scores used in TrustScore v2 calculation.
type DREScores struct {
	DeterminismScore          float64 `json:"determinism_score"`
	ReplayIntegrityScore      float64 `json:"replay_integrity_score"`
	PerformanceStabilityScore float64 `json:"performance_stability_score"`
	DriftScore                float64 `json:"drift_score"`
}

// ============================================
// Performance Stability Tracking
// ============================================

// ResourceHashHistory stores resource hashes for performance stability computation.
type ResourceHashHistory struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID     uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	ResourceHashes json.RawMessage `json:"resource_hashes" gorm:"type:jsonb;default:'[]'"`
	UpdatedAt      time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for ResourceHashHistory.
func (ResourceHashHistory) TableName() string { return "resource_hash_history" }

// ============================================
// Function Version Changelog
// ============================================

// ChangeType represents the type of change in a version
type ChangeType string

const (
	ChangeTypeAdded      ChangeType = "added"
	ChangeTypeChanged    ChangeType = "changed"
	ChangeTypeFixed      ChangeType = "fixed"
	ChangeTypeDeprecated ChangeType = "deprecated"
	ChangeTypeRemoved    ChangeType = "removed"
	ChangeTypeSecurity   ChangeType = "security"
	ChangeTypeBreaking   ChangeType = "breaking"
)

// ChangeCategory represents the category of change
type ChangeCategory string

const (
	ChangeCategoryFeature       ChangeCategory = "feature"
	ChangeCategoryBugFix        ChangeCategory = "bug_fix"
	ChangeCategoryPerformance   ChangeCategory = "performance"
	ChangeCategorySecurity      ChangeCategory = "security"
	ChangeCategoryDocumentation ChangeCategory = "documentation"
	ChangeCategoryDependency    ChangeCategory = "dependency"
	ChangeCategoryBreaking      ChangeCategory = "breaking_change"
	ChangeCategoryOther         ChangeCategory = "other"
)

// FunctionVersionChangelog represents a changelog entry for a function version
type FunctionVersionChangelog struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID        uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	FunctionVersionID uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;index"`
	Version           string          `json:"version" gorm:"not null;index"`
	PreviousVersion   *string         `json:"previous_version,omitempty" gorm:"type:text"`
	ChangeType        ChangeType      `json:"change_type" gorm:"not null"`
	Category          ChangeCategory  `json:"category" gorm:"not null"`
	Title             string          `json:"title" gorm:"not null"`
	Description       string          `json:"description" gorm:"type:text"`
	Changes           json.RawMessage `json:"changes" gorm:"type:jsonb"` // Detailed changes (array of change items)
	Author            string          `json:"author" gorm:"not null"`
	AuthorID          *uuid.UUID      `json:"author_id,omitempty" gorm:"type:uuid"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	FunctionVersion *RegistryFunctionVersion `json:"function_version,omitempty" gorm:"foreignKey:FunctionVersionID;references:ID"`
}

// TableName returns the database table name for FunctionVersionChangelog.
func (FunctionVersionChangelog) TableName() string {
	return "function_version_changelogs"
}

// FunctionChangelogChange represents a single change item within a function version changelog entry
type FunctionChangelogChange struct {
	Component   string `json:"component"`   // e.g., "input schema", "output schema", "runtime"
	Field       string `json:"field"`       // e.g., "timeout", "memory"
	Before      any    `json:"before"`      // previous value
	After       any    `json:"after"`       // new value
	Description string `json:"description"` // human-readable description
}

// ============================================
// Trust Scoring System Types
// ============================================

// TrustTier represents the trust tier classification
type TrustTier string

const (
	TrustTierUntrusted     TrustTier = "untrusted"
	TrustTierTrusted       TrustTier = "trusted"
	TrustTierVerified      TrustTier = "verified"
	TrustTierHighlyTrusted TrustTier = "highly_trusted"
)

// TrustScoreWeights holds the weights for trust score components
type TrustScoreWeights struct {
	Reliability  float64 `json:"reliability"`
	Latency      float64 `json:"latency"`
	ErrorRate    float64 `json:"error_rate"`
	UserRating   float64 `json:"user_rating"`
	Verification float64 `json:"verification"`
}

// DefaultTrustScoreWeights returns the default trust score weights
func DefaultTrustScoreWeights() TrustScoreWeights {
	return TrustScoreWeights{
		Reliability:  0.30,
		Latency:      0.20,
		ErrorRate:    0.20,
		UserRating:   0.15,
		Verification: 0.15,
	}
}

// TrustHistory represents a trust score history entry
type TrustHistory struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID         uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	TrustScore         float64   `json:"trust_score" gorm:"default:0"`
	ReliabilityScore   float64   `json:"reliability_score" gorm:"default:0"`
	LatencyScore       float64   `json:"latency_score" gorm:"default:0"`
	ErrorRateScore     float64   `json:"error_rate_score" gorm:"default:0"`
	UserRatingScore    float64   `json:"user_rating_score" gorm:"default:0"`
	VerificationBonus  float64   `json:"verification_bonus" gorm:"default:0"`
	TotalCalls         int       `json:"total_calls" gorm:"default:0"`
	SuccessRate        float64   `json:"success_rate" gorm:"default:0"`
	P50LatencyMs       int       `json:"p50_latency_ms" gorm:"default:0"`
	P95LatencyMs       int       `json:"p95_latency_ms" gorm:"default:0"`
	P99LatencyMs       int       `json:"p99_latency_ms" gorm:"default:0"`
	ErrorRate          float64   `json:"error_rate" gorm:"default:0"`
	TimeoutRate        float64   `json:"timeout_rate" gorm:"default:0"`
	ConsumerDiversity  int       `json:"consumer_diversity" gorm:"default:0"`
	TenantDiversity    int       `json:"tenant_diversity" gorm:"default:0"`
	UserDiversity      int       `json:"user_diversity" gorm:"default:0"`
	IsVerified         bool      `json:"is_verified" gorm:"default:false"`
	VerificationLevel  string    `json:"verification_level" gorm:"size:50;default:'none'"`
	TrustTier          TrustTier `json:"trust_tier" gorm:"size:20;default:'untrusted'"`
	CalculatedAt       time.Time `json:"calculated_at" gorm:"type:timestamptz"`
	WindowStart        time.Time `json:"window_start" gorm:"type:timestamptz"`
	WindowEnd          time.Time `json:"window_end" gorm:"type:timestamptz"`
	CalculationVersion int       `json:"calculation_version" gorm:"default:1"`
}

// TableName returns the database table name for TrustHistory.
func (TrustHistory) TableName() string {
	return "trust_history"
}

// ExecutionMetrics represents aggregated execution metrics
type ExecutionMetrics struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID      uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	WindowStart     time.Time       `json:"window_start" gorm:"type:timestamptz"`
	WindowEnd       time.Time       `json:"window_end" gorm:"type:timestamptz"`
	WindowType      string          `json:"window_type" gorm:"size:20;default:'hourly'"`
	TotalCalls      int             `json:"total_calls" gorm:"default:0"`
	SuccessfulCalls int             `json:"successful_calls" gorm:"default:0"`
	FailedCalls     int             `json:"failed_calls" gorm:"default:0"`
	TimeoutCalls    int             `json:"timeout_calls" gorm:"default:0"`
	ErrorCalls      int             `json:"error_calls" gorm:"default:0"`
	CachedCalls     int             `json:"cached_calls" gorm:"default:0"`
	LatencyMin      int             `json:"latency_min" gorm:"default:0"`
	LatencyMax      int             `json:"latency_max" gorm:"default:0"`
	LatencySum      int64           `json:"latency_sum" gorm:"default:0"`
	LatencyAvg      float64         `json:"latency_avg" gorm:"default:0"`
	LatencyP50      int             `json:"latency_p50" gorm:"default:0"`
	LatencyP95      int             `json:"latency_p95" gorm:"default:0"`
	LatencyP99      int             `json:"latency_p99" gorm:"default:0"`
	Error4xxCount   int             `json:"error_4xx_count" gorm:"default:0"`
	Error5xxCount   int             `json:"error_5xx_count" gorm:"default:0"`
	UniqueIPs       int             `json:"unique_ips" gorm:"default:0"`
	UniqueTenants   int             `json:"unique_tenants" gorm:"default:0"`
	UniqueUsers     int             `json:"unique_users" gorm:"default:0"`
	GeoDistribution json.RawMessage `json:"geo_distribution" gorm:"type:jsonb"`
	SuccessRate     float64         `json:"success_rate" gorm:"default:0"`
	ErrorRate       float64         `json:"error_rate" gorm:"default:0"`
	TimeoutRate     float64         `json:"timeout_rate" gorm:"default:0"`
	CreatedAt       time.Time       `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for ExecutionMetrics.
func (ExecutionMetrics) TableName() string {
	return "execution_metrics"
}

// TrustScoreWeightsConfig represents the trust score weights configuration
type TrustScoreWeightsConfig struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Component   string     `json:"component" gorm:"uniqueIndex;not null"`
	Weight      float64    `json:"weight" gorm:"default:0"`
	Description string     `json:"description" gorm:"type:text"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"type:timestamptz"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty" gorm:"type:uuid"`
}

// TableName returns the database table name for TrustScoreWeightsConfig.
func (TrustScoreWeightsConfig) TableName() string {
	return "trust_score_weights"
}

// TrustScoreJob represents a trust score recalculation job
type TrustScoreJob struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	JobType            string          `json:"job_type" gorm:"size:50;default:'scheduled'"`
	Status             string          `json:"status" gorm:"size:20;default:'pending'"`
	FunctionsProcessed int             `json:"functions_processed" gorm:"default:0"`
	FunctionsTotal     int             `json:"functions_total" gorm:"default:0"`
	Errors             json.RawMessage `json:"errors" gorm:"type:jsonb"`
	StartedAt          *time.Time      `json:"started_at,omitempty" gorm:"type:timestamptz"`
	CompletedAt        *time.Time      `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt          time.Time       `json:"created_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for TrustScoreJob.
func (TrustScoreJob) TableName() string {
	return "trust_score_jobs"
}

// RemixHistory records the relationship between original functions and their remixes
type RemixHistory struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SourceFunctionID uuid.UUID `json:"source_function_id" gorm:"type:uuid;not null;index"`       // Original function
	TargetFunctionID uuid.UUID `json:"target_function_id" gorm:"type:uuid;not null;uniqueIndex"` // Remixed function
	RemixedByUserID  uuid.UUID `json:"remixed_by_user_id" gorm:"type:uuid;not null;index"`
	RemixedAt        time.Time `json:"remixed_at" gorm:"type:timestamptz;not null;index"`
	Customization    string    `json:"customization" gorm:"type:text"` // User's customization description
	CostUSD          float64   `json:"cost_usd" gorm:"default:0"`      // Fee charged for remix
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	SourceFunction *RegistryFunction `json:"source_function,omitempty" gorm:"foreignKey:SourceFunctionID;references:ID"`
	TargetFunction *RegistryFunction `json:"target_function,omitempty" gorm:"foreignKey:TargetFunctionID;references:ID"`
}

// TableName returns the database table name for RemixHistory.
func (RemixHistory) TableName() string {
	return "remix_history"
}

// FunctionLike tracks user likes on registry functions
type FunctionLike struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index:idx_function_user,unique"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index:idx_function_user,unique"`
	LikedAt    time.Time `json:"liked_at" gorm:"type:timestamptz;not null;autoCreateTime;index"`
}

// TableName returns the database table name for FunctionLike.
func (FunctionLike) TableName() string {
	return "function_likes"
}

// TrustScoreResponse is the API response for trust score queries
type TrustScoreResponse struct {
	FunctionID        uuid.UUID `json:"function_id"`
	TrustScore        float64   `json:"trust_score"`
	TrustTier         TrustTier `json:"trust_tier"`
	IsVerified        bool      `json:"is_verified"`
	VerificationLevel string    `json:"verification_level"`
	Components        struct {
		Reliability  float64 `json:"reliability"`
		Latency      float64 `json:"latency"`
		ErrorRate    float64 `json:"error_rate"`
		UserRating   float64 `json:"user_rating"`
		Verification float64 `json:"verification"`
	} `json:"components"`
	Metrics struct {
		TotalCalls   int     `json:"total_calls"`
		SuccessRate  float64 `json:"success_rate"`
		P50LatencyMs int     `json:"p50_latency_ms"`
		P95LatencyMs int     `json:"p95_latency_ms"`
		P99LatencyMs int     `json:"p99_latency_ms"`
		ErrorRate    float64 `json:"error_rate"`
		TimeoutRate  float64 `json:"timeout_rate"`
	} `json:"metrics"`
	Diversity struct {
		Consumers int `json:"consumers"`
		Tenants   int `json:"tenants"`
		Users     int `json:"users"`
	} `json:"diversity"`
	LastUpdated time.Time `json:"last_updated"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
}

// TrustHistoryResponse is the API response for trust history queries
type TrustHistoryResponse struct {
	FunctionID uuid.UUID      `json:"function_id"`
	History    []TrustHistory `json:"history"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// WindowType represents the type of calculation window
type WindowType string

const (
	// WindowTypeDiscrete uses fixed time windows (hourly, daily, etc)
	WindowTypeDiscrete WindowType = "discrete"
	// WindowTypeSliding uses a continuous rolling window with exponential smoothing
	WindowTypeSliding WindowType = "sliding"
)

// SlidingWindowConfig holds configuration for sliding window calculations
type SlidingWindowConfig struct {
	WindowDuration  time.Duration `json:"window_duration"`  // e.g., 24h
	SmoothingFactor float64       `json:"smoothing_factor"` // alpha for EMA (0-1), higher = more responsive
	MinDataPoints   int           `json:"min_data_points"`  // minimum executions for calculation
	UpdateInterval  time.Duration `json:"update_interval"`  // how often to recalculate
}

// DefaultSlidingWindowConfig returns default sliding window settings
func DefaultSlidingWindowConfig() SlidingWindowConfig {
	return SlidingWindowConfig{
		WindowDuration:  24 * time.Hour,
		SmoothingFactor: 0.3, // Balanced between responsive and stable
		MinDataPoints:   10,
		UpdateInterval:  5 * time.Minute,
	}
}

// TrustScoreDelta represents a change in trust score
type TrustScoreDelta struct {
	FunctionID         uuid.UUID          `json:"function_id"`
	PreviousScore      float64            `json:"previous_score"`
	CurrentScore       float64            `json:"current_score"`
	ScoreChange        float64            `json:"score_change"`         // absolute change
	ScoreChangePercent float64            `json:"score_change_percent"` // percentage change
	PreviousTier       TrustTier          `json:"previous_tier"`
	CurrentTier        TrustTier          `json:"current_tier"`
	TierChanged        bool               `json:"tier_changed"`
	ComponentChanges   map[string]float64 `json:"component_changes,omitempty"`
	CalculatedAt       time.Time          `json:"calculated_at"`
	WindowType         WindowType         `json:"window_type"`
}

// TrustScoreThresholdConfig holds threshold configuration for alerts
type TrustScoreThresholdConfig struct {
	CriticalThreshold  float64       `json:"critical_threshold"`    // Score below this triggers critical alert (default: 50)
	WarningThreshold   float64       `json:"warning_threshold"`     // Score below this triggers warning (default: 70)
	MinChangeForNotify float64       `json:"min_change_for_notify"` // Min score change to trigger notification (default: 5)
	CooldownPeriod     time.Duration `json:"cooldown_period"`       // Min time between notifications per function (default: 15m)
}

// DefaultThresholdConfig returns default threshold settings
func DefaultThresholdConfig() TrustScoreThresholdConfig {
	return TrustScoreThresholdConfig{
		CriticalThreshold:  50.0,
		WarningThreshold:   70.0,
		MinChangeForNotify: 5.0,
		CooldownPeriod:     15 * time.Minute,
	}
}

// TrustScoreStreamEvent represents an event for SSE/WebSocket streaming
type TrustScoreStreamEvent struct {
	EventType  string           `json:"event_type"` // "score_update", "tier_change", "threshold_breach"
	FunctionID uuid.UUID        `json:"function_id"`
	Score      *TrustHistory    `json:"score,omitempty"`
	Delta      *TrustScoreDelta `json:"delta,omitempty"`
	Timestamp  time.Time        `json:"timestamp"`
	WindowType WindowType       `json:"window_type"`
}

// SlidingWindowState holds the state for a sliding window calculation
type SlidingWindowState struct {
	FunctionID         uuid.UUID `json:"function_id" gorm:"type:uuid;primaryKey"`
	CurrentScore       float64   `json:"current_score"`
	PreviousScore      float64   `json:"previous_score"`
	ReliabilityScore   float64   `json:"reliability_score"`
	LatencyScore       float64   `json:"latency_score"`
	ErrorRateScore     float64   `json:"error_rate_score"`
	UserRatingScore    float64   `json:"user_rating_score"`
	VerificationBonus  float64   `json:"verification_bonus"`
	LastUpdated        time.Time `json:"last_updated" gorm:"type:timestamptz"`
	WindowStart        time.Time `json:"window_start" gorm:"type:timestamptz"`
	WindowEnd          time.Time `json:"window_end" gorm:"type:timestamptz"`
	TotalCallsInWindow int       `json:"total_calls_in_window"`
	LastCalculation    time.Time `json:"last_calculation" gorm:"type:timestamptz"`
	SmoothingFactor    float64   `json:"smoothing_factor"`
}

// TableName returns the database table name for SlidingWindowState.
func (SlidingWindowState) TableName() string {
	return "trust_sliding_window_state"
}

// GetComponentScore retrieves a specific component score from the sliding window state
func (s *SlidingWindowState) GetComponentScore(component string) float64 {
	switch component {
	case "reliability":
		return s.ReliabilityScore
	case "latency":
		return s.LatencyScore
	case "error_rate":
		return s.ErrorRateScore
	case "user_rating":
		return s.UserRatingScore
	case "verification":
		return s.VerificationBonus
	default:
		return 0
	}
}

// SetComponentScore updates a specific component score in the sliding window state
func (s *SlidingWindowState) SetComponentScore(component string, score float64) {
	switch component {
	case "reliability":
		s.ReliabilityScore = score
	case "latency":
		s.LatencyScore = score
	case "error_rate":
		s.ErrorRateScore = score
	case "user_rating":
		s.UserRatingScore = score
	case "verification":
		s.VerificationBonus = score
	}
}

// UpdateComponentScores updates all component scores from a TrustHistory record
func (s *SlidingWindowState) UpdateComponentScores(history *TrustHistory) {
	s.ReliabilityScore = history.ReliabilityScore
	s.LatencyScore = history.LatencyScore
	s.ErrorRateScore = history.ErrorRateScore
	s.UserRatingScore = history.UserRatingScore
	s.VerificationBonus = history.VerificationBonus
}

// ============================================
// Verification Pipeline Types
// ============================================

// VerificationJob represents a verification pipeline job
type VerificationJob struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID        uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	FunctionVersionID uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;index"`
	Level             string          `json:"level" gorm:"not null;size:20;default:'basic'"`
	Status            string          `json:"status" gorm:"not null;size:20;default:'queued'"`
	Priority          string          `json:"priority" gorm:"not null;size:20;default:'normal'"`
	RequestedAt       time.Time       `json:"requested_at" gorm:"type:timestamptz"`
	StartedAt         *time.Time      `json:"started_at,omitempty" gorm:"type:timestamptz"`
	CompletedAt       *time.Time      `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	ResultStatus      string          `json:"result_status" gorm:"size:20"`
	ResultData        json.RawMessage `json:"result_data" gorm:"type:jsonb"`
	Error             string          `json:"error" gorm:"type:text"`
	RequestedBy       *uuid.UUID      `json:"requested_by,omitempty" gorm:"type:uuid"`
	IsAutoVerify      bool            `json:"is_auto_verify" gorm:"default:false"`
	RetryCount        int             `json:"retry_count" gorm:"default:0"`
	MaxRetries        int             `json:"max_retries" gorm:"default:3"`
	CreatedAt         time.Time       `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for VerificationJob.
func (VerificationJob) TableName() string {
	return "verification_jobs"
}

// VerificationResult represents detailed results of a verification pipeline run
type VerificationResult struct {
	ID                   uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	JobID                uuid.UUID       `json:"job_id" gorm:"type:uuid;not null;index"`
	FunctionID           uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	FunctionVersionID    uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;index"`
	Level                string          `json:"level" gorm:"not null;size:20"`
	Status               string          `json:"status" gorm:"not null;size:20"`
	StagesData           json.RawMessage `json:"stages_data" gorm:"type:jsonb"`
	MalwareScanPassed    *bool           `json:"malware_scan_passed"`
	MalwareScanRiskScore float64         `json:"malware_scan_risk_score" gorm:"default:0"`
	DRPassed             *bool           `json:"dre_passed"`
	DRPassRate           float64         `json:"dre_pass_rate" gorm:"default:0"`
	DREIsDeterministic   bool            `json:"dre_is_deterministic" gorm:"default:false"`
	FXCERTPassed         *bool           `json:"fxcert_passed"`
	FXCERTValidUntil     *time.Time      `json:"fxcert_valid_until,omitempty" gorm:"type:timestamptz"`
	ManualReviewStatus   string          `json:"manual_review_status" gorm:"size:20"`
	TotalExecutions      int             `json:"total_executions" gorm:"default:0"`
	SuccessRate          float64         `json:"success_rate" gorm:"default:0"`
	AvgLatencyMs         int             `json:"avg_latency_ms" gorm:"default:0"`
	ErrorRate            float64         `json:"error_rate" gorm:"default:0"`
	TrustScoreImpact     float64         `json:"trust_score_impact" gorm:"default:0"`
	StartedAt            time.Time       `json:"started_at" gorm:"type:timestamptz"`
	CompletedAt          *time.Time      `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	Error                string          `json:"error" gorm:"type:text"`
	CreatedAt            time.Time       `json:"created_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for VerificationResult.
func (VerificationResult) TableName() string {
	return "verification_results"
}

// ManualReviewQueue represents a human review queue entry for Level 3 verification
type ManualReviewQueue struct {
	ID                          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID                  uuid.UUID  `json:"function_id" gorm:"type:uuid;not null;index"`
	FunctionVersionID           uuid.UUID  `json:"function_version_id" gorm:"type:uuid;not null;index"`
	VerificationJobID           *uuid.UUID `json:"verification_job_id,omitempty" gorm:"type:uuid"`
	Status                      string     `json:"status" gorm:"not null;size:20;default:'pending'"`
	Priority                    string     `json:"priority" gorm:"not null;size:20;default:'normal'"`
	AssignedTo                  *uuid.UUID `json:"assigned_to,omitempty" gorm:"type:uuid"`
	ReviewType                  string     `json:"review_type" gorm:"not null;size:50"`
	ReviewNotes                 string     `json:"review_notes" gorm:"type:text"`
	ReviewComments              string     `json:"review_comments" gorm:"type:text"`
	DecisionAt                  *time.Time `json:"decision_at,omitempty" gorm:"type:timestamptz"`
	DecisionBy                  *uuid.UUID `json:"decision_by,omitempty" gorm:"type:uuid"`
	DecisionReason              string     `json:"decision_reason" gorm:"type:text"`
	CreatedAt                   time.Time  `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt                   time.Time  `json:"updated_at" gorm:"type:timestamptz"`
	DueAt                       *time.Time `json:"due_at,omitempty" gorm:"type:timestamptz"`
	CompletedAt                 *time.Time `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	AutoApproveIfNoResponseDays int        `json:"auto_approve_if_no_response_days" gorm:"default:7"`
}

// TableName returns the database table name for ManualReviewQueue.
func (ManualReviewQueue) TableName() string {
	return "manual_review_queue"
}

// VerificationLevelConfig defines requirements for each verification level
type VerificationLevelConfig struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Level                string     `json:"level" gorm:"uniqueIndex;not null;size:20"`
	RequiresMalwareScan  bool       `json:"requires_malware_scan" gorm:"default:true"`
	RequiresDRE          bool       `json:"requires_dre" gorm:"default:false"`
	RequiresFXCERT       bool       `json:"requires_fxcert" gorm:"default:false"`
	RequiresManualReview bool       `json:"requires_manual_review" gorm:"default:false"`
	MinDRuEPassRate      float64    `json:"min_dre_pass_rate" gorm:"default:0.95"`
	MaxLatencyMs         int        `json:"max_latency_ms" gorm:"default:5000"`
	MinSuccessRate       float64    `json:"min_success_rate" gorm:"default:0.99"`
	AutoUpgradeFromLevel string     `json:"auto_upgrade_from_level" gorm:"size:20"`
	AutoUpgradeAfterDays int        `json:"auto_upgrade_after_days"`
	TrustBonus           float64    `json:"trust_bonus" gorm:"default:0"`
	IsActive             bool       `json:"is_active" gorm:"default:true"`
	Description          string     `json:"description" gorm:"type:text"`
	CreatedAt            time.Time  `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"type:timestamptz"`
	UpdatedBy            *uuid.UUID `json:"updated_by,omitempty" gorm:"type:uuid"`
}

// TableName returns the database table name for VerificationLevelConfig.
func (VerificationLevelConfig) TableName() string {
	return "verification_level_config"
}

// VerificationAuditLog represents audit trail for verification activities
type VerificationAuditLog struct {
	ID                   uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID           *uuid.UUID      `json:"function_id,omitempty" gorm:"type:uuid"`
	VerificationJobID    *uuid.UUID      `json:"verification_job_id,omitempty" gorm:"type:uuid"`
	VerificationResultID *uuid.UUID      `json:"verification_result_id,omitempty" gorm:"type:uuid"`
	Action               string          `json:"action" gorm:"not null;size:50"`
	ActorType            string          `json:"actor_type" gorm:"not null;size:20"`
	ActorID              *uuid.UUID      `json:"actor_id,omitempty" gorm:"type:uuid"`
	ActorEmail           string          `json:"actor_email" gorm:"size:255"`
	OldValue             json.RawMessage `json:"old_value" gorm:"type:jsonb"`
	NewValue             json.RawMessage `json:"new_value" gorm:"type:jsonb"`
	IPAddress            string          `json:"ip_address" gorm:"size:45"`
	UserAgent            string          `json:"user_agent" gorm:"type:text"`
	CreatedAt            time.Time       `json:"created_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for VerificationAuditLog.
func (VerificationAuditLog) TableName() string {
	return "verification_audit_log"
}

// VerificationSchedule represents scheduled periodic re-verification
type VerificationSchedule struct {
	ID                     uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID             uuid.UUID  `json:"function_id" gorm:"type:uuid;not null;index"`
	VerificationLevel      string     `json:"verification_level" gorm:"not null;size:20"`
	Frequency              string     `json:"frequency" gorm:"not null;size:20;default:'monthly'"`
	NextVerificationAt     time.Time  `json:"next_verification_at" gorm:"type:timestamptz;index"`
	LastVerificationAt     *time.Time `json:"last_verification_at,omitempty" gorm:"type:timestamptz"`
	LastVerificationResult string     `json:"last_verification_result" gorm:"size:20"`
	IsActive               bool       `json:"is_active" gorm:"default:true"`
	IsPaused               bool       `json:"is_paused" gorm:"default:false"`
	PauseReason            string     `json:"pause_reason" gorm:"type:text"`
	NotifyOnFailure        bool       `json:"notify_on_failure" gorm:"default:true"`
	NotificationEmail      string     `json:"notification_email" gorm:"size:255"`
	CreatedAt              time.Time  `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt              time.Time  `json:"updated_at" gorm:"type:timestamptz"`
	CreatedBy              *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
}

// TableName returns the database table name for VerificationSchedule.
func (VerificationSchedule) TableName() string {
	return "verification_schedule"
}
