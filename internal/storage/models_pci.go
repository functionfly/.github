package storage

import (
	"time"

	"github.com/google/uuid"
)

// PCIAuditEventType represents types of PCI DSS audit events
const (
	// Cardholder Data Access Events
	PCIAuditCardDataRead        = "card_data_read"
	PCIAuditCardDataWrite       = "card_data_write"
	PCIAuditCardDataDelete      = "card_data_delete"
	PCIAuditCardDataTokenized   = "card_data_tokenized"
	PCIAuditCardDataDetokenized = "card_data_detokenized"

	// Encryption Key Events
	PCIAuditKeyCreated        = "encryption_key_created"
	PCIAuditKeyRotated        = "encryption_key_rotated"
	PCIAuditKeyRetired        = "encryption_key_retired"
	PCIAuditKeyAccessed       = "encryption_key_accessed"
	PCIAuditKeyBackupCreated  = "encryption_key_backup_created"
	PCIAuditKeyBackupRestored = "encryption_key_backup_restored"

	// Payment Flow Events
	PCIAuditPaymentInitiated   = "payment_initiated"
	PCIAuditPaymentProcessed   = "payment_processed"
	PCIAuditPaymentFailed      = "payment_failed"
	PCIAuditRefundProcessed    = "refund_processed"
	PCIAuditChargebackReceived = "chargeback_received"

	// Authentication Events
	PCIAuditAuthSuccess       = "authentication_success"
	PCIAuditAuthFailure       = "authentication_failure"
	PCIAuditSessionCreated    = "session_created"
	PCIAuditSessionTerminated = "session_terminated"

	// Access Control Events
	PCIAuditAccessGranted     = "access_granted"
	PCIAuditAccessRevoked     = "access_revoked"
	PCIAuditPrivilegeElevated = "privilege_elevated"
	PCIAuditAdminAction       = "admin_action"
)

// PCIAuditEventSeverity represents severity levels for PCI audit events
const (
	PCISeverityInfo      = "info"
	PCISeverityWarning   = "warning"
	PCISeverityCritical  = "critical"
	PCISeverityEmergency = "emergency"
)

// PCIAuditEvent represents a PCI DSS compliance audit log entry
// These events are immutable and retained per PCI DSS requirements (minimum 1 year, critical data 3 years)
type PCIAuditEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	EventType string    `json:"event_type" db:"event_type"` // Type of PCI event
	Severity  string    `json:"severity" db:"severity"`     // info, warning, critical, emergency

	// Actor Information
	ActorUserID    *uuid.UUID `json:"actor_user_id,omitempty" db:"actor_user_id"`
	ActorEmail     string     `json:"actor_email,omitempty" db:"actor_email"`
	ActorRole      string     `json:"actor_role,omitempty" db:"actor_role"` // admin, system, api, customer
	ActorIP        string     `json:"actor_ip,omitempty" db:"actor_ip"`
	ActorUserAgent string     `json:"actor_user_agent,omitempty" db:"actor_user_agent"`
	SessionID      string     `json:"session_id,omitempty" db:"session_id"`

	// Resource Information (what was accessed)
	ResourceType string     `json:"resource_type" db:"resource_type"` // payment_method, encryption_key, transaction, etc.
	ResourceID   *uuid.UUID `json:"resource_id,omitempty" db:"resource_id"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`

	// Cardholder Data Context (NEVER store full PAN or CVV)
	// Only store last 4 digits and card brand for identification
	CardLastFour    *string `json:"card_last_four,omitempty" db:"card_last_four"` // Last 4 digits only
	CardBrand       *string `json:"card_brand,omitempty" db:"card_brand"`         // visa, mastercard, etc.
	CardExpiryMonth *int    `json:"card_expiry_month,omitempty" db:"card_expiry_month"`
	CardExpiryYear  *int    `json:"card_expiry_year,omitempty" db:"card_expiry_year"`
	TokenID         *string `json:"token_id,omitempty" db:"token_id"` // Reference to tokenized card

	// Encryption Key Context
	KeyID        *uuid.UUID `json:"key_id,omitempty" db:"key_id"`
	KeyAlgorithm *string    `json:"key_algorithm,omitempty" db:"key_algorithm"` // AES-256-GCM, RSA-2048, etc.
	KeyOperation *string    `json:"key_operation,omitempty" db:"key_operation"` // encrypt, decrypt, sign, verify

	// Event Details
	Description   string  `json:"description" db:"description"`
	RequestID     string  `json:"request_id,omitempty" db:"request_id"`
	TransactionID *string `json:"transaction_id,omitempty" db:"transaction_id"`
	StripeEventID *string `json:"stripe_event_id,omitempty" db:"stripe_event_id"`

	// Security Context
	AuthMethod    string  `json:"auth_method,omitempty" db:"auth_method"` // mfa, password, api_key, webhook_signature
	MFAUsed       bool    `json:"mfa_used" db:"mfa_used"`
	Success       bool    `json:"success" db:"success"`
	FailureReason *string `json:"failure_reason,omitempty" db:"failure_reason"`

	// Compliance Metadata (JSON for flexibility)
	ComplianceData interface{} `json:"compliance_data,omitempty" db:"compliance_data"` // Additional PCI-required fields
	Metadata       interface{} `json:"metadata,omitempty" db:"metadata"`

	// Retention and Tamper Protection
	RetentionUntil *time.Time `json:"retention_until,omitempty" db:"retention_until"` // When this record can be deleted
	TamperHash     string     `json:"tamper_hash,omitempty" db:"tamper_hash"`         // Cryptographic hash for integrity
	ChainHash      string     `json:"chain_hash,omitempty" db:"chain_hash"`           // Hash chaining for audit trail integrity

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PCIEncryptionKey represents an encryption key used for cardholder data protection
// Tracks key lifecycle for PCI DSS compliance (req 3.6)
type PCIEncryptionKey struct {
	ID             uuid.UUID `json:"id" db:"id"`
	KeyType        string    `json:"key_type" db:"key_type"`               // data_encryption, key_encryption, signing
	Algorithm      string    `json:"algorithm" db:"algorithm"`             // AES-256-GCM, RSA-4096, etc.
	KeyFingerprint string    `json:"key_fingerprint" db:"key_fingerprint"` // Hash of key for identification (NOT the key itself)
	KeyStatus      string    `json:"key_status" db:"key_status"`           // active, rotated, retired, destroyed

	// Key Lifecycle
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	ActivatedAt    *time.Time `json:"activated_at,omitempty" db:"activated_at"`
	RotatedAt      *time.Time `json:"rotated_at,omitempty" db:"rotated_at"`
	RotatedFromKey *uuid.UUID `json:"rotated_from_key,omitempty" db:"rotated_from_key"`
	RetiredAt      *time.Time `json:"retired_at,omitempty" db:"retired_at"`
	RetiredReason  *string    `json:"retired_reason,omitempty" db:"retired_reason"`
	DestroyedAt    *time.Time `json:"destroyed_at,omitempty" db:"destroyed_at"`

	// Key Splitting/Sharing (for custodian requirements)
	HasKeyShares      bool `json:"has_key_shares" db:"has_key_shares"`
	KeySharesTotal    int  `json:"key_shares_total" db:"key_shares_total"`
	KeySharesRequired int  `json:"key_shares_required" db:"key_shares_required"`

	// HSM/External KMS
	IsHSMBacked bool    `json:"is_hsm_backed" db:"is_hsm_backed"`
	HSMKeyID    *string `json:"hsm_key_id,omitempty" db:"hsm_key_id"`
	KMSProvider *string `json:"kms_provider,omitempty" db:"kms_provider"` // aws, gcp, azure, etc.

	// Rotation Schedule
	RotationDueAt        *time.Time `json:"rotation_due_at,omitempty" db:"rotation_due_at"`
	RotationIntervalDays int        `json:"rotation_interval_days" db:"rotation_interval_days"`

	Metadata  interface{} `json:"metadata,omitempty" db:"metadata"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// PCIKeyAccessLog tracks who accessed encryption keys and why
// Required for PCI DSS key custodian tracking
type PCIKeyAccessLog struct {
	ID           uuid.UUID `json:"id" db:"id"`
	KeyID        uuid.UUID `json:"key_id" db:"key_id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	UserEmail    string    `json:"user_email" db:"user_email"`
	AccessType   string    `json:"access_type" db:"access_type"` // read, use, rotate, backup, restore
	AccessReason string    `json:"access_reason" db:"access_reason"`

	// Context
	IPAddress   string `json:"ip_address" db:"ip_address"`
	UserAgent   string `json:"user_agent" db:"user_agent"`
	SessionID   string `json:"session_id" db:"session_id"`
	MFAVerified bool   `json:"mfa_verified" db:"mfa_verified"`

	// Approval for sensitive operations
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
	ApprovalTicket *string    `json:"approval_ticket,omitempty" db:"approval_ticket"`

	Success       bool    `json:"success" db:"success"`
	FailureReason *string `json:"failure_reason,omitempty" db:"failure_reason"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PCICardholderDataAccessLog tracks all access to cardholder data environment
// Helps demonstrate compliance with PCI DSS requirement 10
type PCICardholderDataAccessLog struct {
	ID         uuid.UUID `json:"id" db:"id"`
	AccessType string    `json:"access_type" db:"access_type"` // read, write, delete, tokenize, detokenize
	DataType   string    `json:"data_type" db:"data_type"`     // pan, expiry, name, service_code

	// Actor
	UserID        *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	SystemProcess string     `json:"system_process,omitempty" db:"system_process"`
	APIKeyID      *uuid.UUID `json:"api_key_id,omitempty" db:"api_key_id"`

	// Context
	TenantID        *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	PaymentMethodID *uuid.UUID `json:"payment_method_id,omitempty" db:"payment_method_id"`
	TransactionID   *string    `json:"transaction_id,omitempty" db:"transaction_id"`

	// Data Identification (safely logged)
	CardLastFour   *string `json:"card_last_four,omitempty" db:"card_last_four"`
	TokenReference *string `json:"token_reference,omitempty" db:"token_reference"`

	// Access Context
	IPAddress string `json:"ip_address" db:"ip_address"`
	RequestID string `json:"request_id" db:"request_id"`
	Purpose   string `json:"purpose" db:"purpose"`

	// CDE (Cardholder Data Environment) access tracking
	CDESection   string `json:"cde_section" db:"cde_section"`
	DataFlowStep string `json:"data_flow_step" db:"data_flow_step"`

	Success       bool    `json:"success" db:"success"`
	FailureReason *string `json:"failure_reason,omitempty" db:"failure_reason"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PCIEnvironmentControl tracks access to the Cardholder Data Environment
// Documents network segmentation and access controls (PCI DSS req 1, 2)
type PCIEnvironmentControl struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ControlType string    `json:"control_type" db:"control_type"` // firewall_rule, access_list, network_segment
	ControlName string    `json:"control_name" db:"control_name"`
	Description string    `json:"description" db:"description"`

	// For network controls
	SourceIPRange      *string `json:"source_ip_range,omitempty" db:"source_ip_range"`
	DestinationIPRange *string `json:"destination_ip_range,omitempty" db:"destination_ip_range"`
	PortRange          *string `json:"port_range,omitempty" db:"port_range"`
	Protocol           *string `json:"protocol,omitempty" db:"protocol"`

	// Status
	IsActive   bool      `json:"is_active" db:"is_active"`
	ApprovedBy uuid.UUID `json:"approved_by" db:"approved_by"`
	ApprovedAt time.Time `json:"approved_at" db:"approved_at"`

	// Review cycle
	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty" db:"last_reviewed_at"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	NextReviewAt   *time.Time `json:"next_review_at,omitempty" db:"next_review_at"`

	Metadata  interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}
