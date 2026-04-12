package privacy

import (
	"time"

	"github.com/google/uuid"
)

// PrivacyLevel defines the level of privacy protection
type PrivacyLevel string

const (
	// PrivacyLevelStandard - basic PII logging with retention
	PrivacyLevelStandard PrivacyLevel = "standard"
	// PrivacyLevelEnhanced - anonymized PII, shorter retention
	PrivacyLevelEnhanced PrivacyLevel = "enhanced"
	// PrivacyLevelMaximum - no PII logging, crypto-shredding available
	PrivacyLevelMaximum PrivacyLevel = "maximum"
	// PrivacyLevelGDPR - full GDPR compliance mode
	PrivacyLevelGDPR PrivacyLevel = "gdpr"
)

// PIIMaskType defines how PII should be masked
type PIIMaskType string

const (
	// PIIMaskTypeNone - no masking, store full value
	PIIMaskTypeNone PIIMaskType = "none"
	// PIIMaskTypeHash - one-way hash (irreversible)
	PIIMaskTypeHash PIIMaskType = "hash"
	// PIIMaskTypePartial - partial masking (e.g., 192.168.x.x)
	PIIMaskTypePartial PIIMaskType = "partial"
	// PIIMaskTypeRedact - completely redacted
	PIIMaskTypeRedact PIIMaskType = "redact"
)

// PrivacySettings represents tenant/user privacy configuration
type PrivacySettings struct {
	ID                uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID          *uuid.UUID    `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	UserID            *uuid.UUID    `json:"user_id,omitempty" gorm:"type:uuid;index"`
	PrivacyLevel      PrivacyLevel  `json:"privacy_level" gorm:"not null;default:'standard'"`
	AnonymizeIP       bool          `json:"anonymize_ip" gorm:"not null;default:false"`
	AnonymizeUserAgent bool         `json:"anonymize_user_agent" gorm:"not null;default:false"`
	LogGeoData        bool          `json:"log_geo_data" gorm:"not null;default:true"`
	LogEmbedOrigin    bool          `json:"log_embed_origin" gorm:"not null;default:true"`
	StoreInputOutput  bool          `json:"store_input_output" gorm:"not null;default:true"`
	RetentionDays     int           `json:"retention_days" gorm:"not null;default:90"`
	GDPRMode          bool          `json:"gdpr_mode" gorm:"not null;default:false"`
	AutoDeleteEnabled bool          `json:"auto_delete_enabled" gorm:"not null;default:false"`
	ConsentRequired   bool          `json:"consent_required" gorm:"not null;default:false"`
	ConsentGivenAt    *time.Time    `json:"consent_given_at,omitempty"`
	ConsentVersion    string        `json:"consent_version,omitempty"`
	IPMaskType        PIIMaskType   `json:"ip_mask_type" gorm:"not null;default:'none'"`
	UserAgentMaskType PIIMaskType   `json:"user_agent_mask_type" gorm:"not null;default:'none'"`
	CreatedAt         time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	UpdatedBy         *uuid.UUID    `json:"updated_by,omitempty"`
}

func (PrivacySettings) TableName() string {
	return "privacy_settings"
}

// PrivacyConsentRecord tracks user consent for data processing
type PrivacyConsentRecord struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ConsentType     string    `json:"consent_type" gorm:"not null;index"` // "execution_logging", "analytics", "marketing"
	ConsentGiven    bool      `json:"consent_given" gorm:"not null"`
	ConsentVersion  string    `json:"consent_version" gorm:"not null"`
	ConsentText     string    `json:"consent_text" gorm:"type:text"`
	IPHash          string    `json:"ip_hash" gorm:"size:64"` // Hashed IP for audit
	UserAgentHash   string    `json:"user_agent_hash" gorm:"size:64"` // Hashed UA for audit
	GivenAt         time.Time `json:"given_at" gorm:"not null"`
	WithdrawnAt     *time.Time `json:"withdrawn_at,omitempty"`
	WithdrawnReason string    `json:"withdrawn_reason,omitempty"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (PrivacyConsentRecord) TableName() string {
	return "privacy_consent_records"
}

// DataExportRequest represents a user's GDPR data export request
type DataExportRequest struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	TenantID        *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	Status          string     `json:"status" gorm:"not null;default:'pending';index"` // "pending", "processing", "completed", "failed"
	RequestType     string     `json:"request_type" gorm:"not null;default:'full'"` // "full", "executions", "profile", "audit"
	RequestedAt     time.Time  `json:"requested_at" gorm:"not null"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DownloadURL     string     `json:"download_url,omitempty"`
	DownloadToken   string     `json:"download_token,omitempty"`
	FileSize        int64      `json:"file_size,omitempty"`
	RecordCount     int64      `json:"record_count,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (DataExportRequest) TableName() string {
	return "data_export_requests"
}

// DataDeletionRequest represents a GDPR right-to-erasure request
type DataDeletionRequest struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	TenantID          *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	Status            string     `json:"status" gorm:"not null;default:'pending';index"` // "pending", "processing", "completed", "failed", "partial"
	RequestType       string     `json:"request_type" gorm:"not null;default:'full'"` // "full", "executions", "audit_logs"
	RequestedAt       time.Time  `json:"requested_at" gorm:"not null"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	RecordsDeleted    int64      `json:"records_deleted,omitempty"`
	RecordsAnonymized int64      `json:"records_anonymized,omitempty"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	VerificationHash  string     `json:"verification_hash,omitempty"` // Hash to verify deletion
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (DataDeletionRequest) TableName() string {
	return "data_deletion_requests"
}

// AnonymizedExecution represents an execution record with anonymized PII
type AnonymizedExecution struct {
	ID              uuid.UUID  `json:"id"`
	FunctionID      uuid.UUID  `json:"function_id"`
	Version         string     `json:"version"`
	DurationMs      int        `json:"duration_ms"`
	StatusCode      int        `json:"status_code"`
	Cached          bool       `json:"cached"`
	Outcome         string     `json:"outcome"`
	IPHashPrefix    string     `json:"ip_hash_prefix,omitempty"` // First 8 chars of IP hash for rough geo
	UserAgentHash   string     `json:"user_agent_hash,omitempty"` // Hash for uniqueness tracking
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
	Region          string     `json:"region,omitempty"` // Region code instead of exact geo
}

// PrivacyAuditLog tracks privacy-related operations
type PrivacyAuditLog struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Action       string    `json:"action" gorm:"not null;index"` // "export_requested", "export_completed", "deletion_requested", "deletion_completed", "consent_given", "consent_withdrawn"
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	RequestID    *uuid.UUID `json:"request_id,omitempty"` // Export or deletion request ID
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Details      string    `json:"details,omitempty" gorm:"type:text"`
	Success      bool      `json:"success" gorm:"not null"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

func (PrivacyAuditLog) TableName() string {
	return "privacy_audit_log"
}

// PIIDetectionResult contains detected PII in data
type PIIDetectionResult struct {
	HasPII      bool                 `json:"has_pii"`
	Categories  []string             `json:"categories,omitempty"` // "email", "phone", "ssn", "credit_card", etc.
	Confidence  float64              `json:"confidence"`           // 0.0-1.0
	Matches     []PIIMatch           `json:"matches,omitempty"`
	RedactedData string              `json:"redacted_data,omitempty"`
}

// PIIMatch represents a specific PII match
type PIIMatch struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Position   int     `json:"position"`
	Length     int     `json:"length"`
	Confidence float64 `json:"confidence"`
}

// PrivacyHeaders contains privacy-related HTTP headers
type PrivacyHeaders struct {
	DoNotTrack      bool   `json:"do_not_track"`
	GDPRApplies     bool   `json:"gdpr_applies"`
	CCPAApplies     bool   `json:"ccpa_applies"`
	ConsentGiven    bool   `json:"consent_given"`
	PrivacyLevel    string `json:"privacy_level"`
	RequestAnonymization bool `json:"request_anonymization"`
}

// GDPRDataPackage represents all data for a user export
type GDPRDataPackage struct {
	ExportID      string                 `json:"export_id"`
	UserID        string                 `json:"user_id"`
	GeneratedAt   time.Time              `json:"generated_at"`
	Version       string                 `json:"version"`
	Profile       map[string]interface{} `json:"profile,omitempty"`
	Executions    []ExecutionData        `json:"executions,omitempty"`
	AuditLogs     []AuditLogData         `json:"audit_logs,omitempty"`
	ConsentRecords []ConsentData         `json:"consent_records,omitempty"`
}

// ExecutionData contains execution-related data for export
type ExecutionData struct {
	ExecutionID   string    `json:"execution_id"`
	FunctionID    string    `json:"function_id"`
	FunctionName  string    `json:"function_name,omitempty"`
	Version       string    `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
	DurationMs    int       `json:"duration_ms"`
	StatusCode    int       `json:"status_code"`
	Cached        bool      `json:"cached"`
}

// AuditLogData contains audit data for export
type AuditLogData struct {
	EventID      string    `json:"event_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	Timestamp    time.Time `json:"timestamp"`
	Success      bool      `json:"success"`
}

// ConsentData contains consent records for export
type ConsentData struct {
	ConsentType  string    `json:"consent_type"`
	Given        bool      `json:"given"`
	Version      string    `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

// GlobalPrivacySettings represents system-wide privacy defaults
type GlobalPrivacySettings struct {
	ID                        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DefaultPrivacyLevel       PrivacyLevel `json:"default_privacy_level" gorm:"not null;default:'standard'"`
	DefaultIPMaskType         PIIMaskType  `json:"default_ip_mask_type" gorm:"not null;default:'none'"`
	DefaultUserAgentMaskType  PIIMaskType  `json:"default_user_agent_mask_type" gorm:"not null;default:'none'"`
	DefaultRetentionDays      int          `json:"default_retention_days" gorm:"not null;default:90"`
	GDPRModeEnabled           bool         `json:"gdpr_mode_enabled" gorm:"not null;default:false"`
	CCPAModeEnabled           bool         `json:"ccpa_mode_enabled" gorm:"not null;default:false"`
	AutoAnonymizeAfterDays    int          `json:"auto_anonymize_after_days" gorm:"default:0"`
	RequireConsent           bool         `json:"require_consent" gorm:"not null;default:false"`
	PIIScanningEnabled       bool         `json:"pii_scanning_enabled" gorm:"not null;default:false"`
	InputOutputRedaction     bool         `json:"input_output_redaction" gorm:"not null;default:false"`
	CreatedAt                time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	UpdatedBy                *uuid.UUID   `json:"updated_by,omitempty"`
	IsActive                 bool         `json:"is_active" gorm:"not null;default:true;unique"`
}

func (GlobalPrivacySettings) TableName() string {
	return "global_privacy_settings"
}
