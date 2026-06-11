package vault

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Secret represents an encrypted secret in the vault
// The actual secret value is encrypted client-side before storage
type Secret struct {
	// Primary key using UUID
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Tenant ownership for multi-tenancy
	TenantID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID   uuid.UUID `gorm:"type:uuid;not null"` // User who created the secret (required by DB)

	// Secret identification
	Name        string `gorm:"not null;size:255"`
	Description string `gorm:"type:text"`

	// Secret type categorization
	SecretType SecretType `gorm:"not null;size:50;default:'api_key'"`

	// Encrypted data - never store plaintext
	// This is the AES-256-GCM encrypted blob (ciphertext + auth tag)
	EncryptedValue []byte `gorm:"type:bytea;not null"`

	// Encryption metadata (column names match migration: encryption_iv, encryption_salt; BYTEA in DB)
	EncryptionSalt    []byte `gorm:"column:encryption_salt;type:bytea;not null"`     // PBKDF2 salt
	IV                []byte `gorm:"column:encryption_iv;type:bytea;not null"`       // AES-GCM IV/nonce
	EncryptionAuthTag []byte `gorm:"column:encryption_auth_tag;type:bytea;not null"` // GCM auth tag (required by DB)
	KeyVersion        int    `gorm:"not null;default:1"`                             // 1=passphrase, 2=KMS, 3=HSM

	// Access control scopes (JSONB for flexibility)
	Scopes JSONMap `gorm:"type:jsonb;not null;default:'[]'::jsonb"`

	// Extensibility metadata
	Metadata JSONMap `gorm:"type:jsonb;not null;default:'{}'::jsonb"`

	// Usage tracking
	LastAccessedAt *time.Time `gorm:"column:last_accessed_at"`
	AccessCount    int        `gorm:"not null;default:0"`

	// Soft delete for audit trail
	DeletedAt *time.Time `gorm:"column:deleted_at"`

	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Version tracking
	CurrentVersion *int       `gorm:"column:current_version"`
	LastModifiedBy *uuid.UUID `gorm:"column:last_modified_by"`
	LastModifiedAt *time.Time `gorm:"column:last_modified_at"`
}

// TableName specifies the table name for Secret
func (Secret) TableName() string {
	return "secrets_vault"
}

// IsDeleted returns true if the secret has been soft-deleted
func (s *Secret) IsDeleted() bool {
	return s.DeletedAt != nil
}

// SecretVersion represents a historical snapshot of a secret at a specific version
// This enables versioning, rollback, and audit trails while maintaining zero-knowledge encryption
type SecretVersion struct {
	// Primary key using UUID
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Foreign key to the parent secret
	SecretID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Tenant ownership for multi-tenancy
	TenantID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Version number (sequential, starting at 1)
	VersionNumber int `gorm:"not null"`

	// Snapshot of secret metadata at this version
	Name        string     `gorm:"not null;size:255"`
	Description string     `gorm:"type:text"`
	SecretType  SecretType `gorm:"not null;size:50"`

	// Encrypted data snapshot (zero-knowledge: server never sees plaintext)
	EncryptedValue    []byte `gorm:"type:bytea;not null"`
	EncryptionSalt    []byte `gorm:"column:encryption_salt;type:bytea;not null"`
	IV                []byte `gorm:"column:encryption_iv;type:bytea;not null"`
	EncryptionAuthTag []byte `gorm:"column:encryption_auth_tag;type:bytea;not null"`
	KeyVersion        int    `gorm:"not null;default:1"`

	// Scopes and metadata at this version
	Scopes   JSONMap `gorm:"type:jsonb;not null;default:'[]'::jsonb"`
	Metadata JSONMap `gorm:"type:jsonb;not null;default:'{}'::jsonb"`

	// Change tracking
	ChangeType    string `gorm:"not null;size:20"` // 'create', 'update', 'rollback'
	ChangeSummary string `gorm:"type:text"`

	// Actor who created this version
	ActorID   uuid.UUID `gorm:"type:uuid;not null"`
	ActorType ActorType `gorm:"not null;size:50;default:'user'"`

	// Timestamp when this version was created
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for SecretVersion
func (SecretVersion) TableName() string {
	return "secret_versions"
}

// BeforeCreate hook to ensure UUID is set if not provided
func (v *SecretVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// AccessToken represents a scoped, time-limited access token for secrets
// Raw tokens are never stored; only SHA-256 hashes are kept
type AccessToken struct {
	// Primary key using UUID
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Foreign key to the secret this token grants access to
	SecretID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Tenant ownership for security filtering
	TenantID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Token hash (SHA-256 of the actual token)
	// Raw tokens are returned once at creation and never stored
	TokenHash string `gorm:"not null;size:255;uniqueIndex"`

	// Token metadata
	Name        string `gorm:"size:255"`
	Description string `gorm:"type:text"`

	// Scope and permissions (JSONB for flexible evolution)
	Scopes JSONMap `gorm:"type:jsonb;not null;default:'[]'::jsonb"`

	// Expiration and revocation
	ExpiresAt     time.Time  `gorm:"not null"`
	IsRevoked     bool       `gorm:"not null;default:false"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	RevokedReason string     `gorm:"column:revoked_reason;type:text"`

	// Usage tracking
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	UseCount   int        `gorm:"not null;default:0"`

	// Actor who created this token
	CreatedBy string `gorm:"not null;size:255"`

	// Timestamp
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for AccessToken
func (AccessToken) TableName() string {
	return "secret_access_tokens"
}

// IsValid returns true if the token is not expired and not revoked
func (t *AccessToken) IsValid() bool {
	if t.IsRevoked {
		return false
	}
	if t.RevokedAt != nil {
		return false
	}
	return time.Now().Before(t.ExpiresAt)
}

// RecordUse updates the usage statistics for the token
func (t *AccessToken) RecordUse() {
	t.UseCount++
	now := time.Now()
	t.LastUsedAt = &now
}

// AuditLog represents an immutable audit trail entry for vault operations
// All actions on secrets are logged for compliance and security
type AuditLog struct {
	// Primary key using UUID
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Foreign key to the affected secret (nullable in case secret is deleted)
	SecretID *uuid.UUID `gorm:"type:uuid;index"`

	// Tenant ownership for filtering
	TenantID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Action performed
	Action AuditAction `gorm:"not null;size:50"`

	// Actor information
	ActorID   string    `gorm:"not null;size:255;index"`
	ActorType ActorType `gorm:"not null;size:50"`

	// Request context
	RequestID string `gorm:"size:255"`                  // For correlating with API requests
	IPAddress string `gorm:"column:ip_address;size:45"` // IPv6 max length
	UserAgent string `gorm:"type:text"`

	// Action metadata (redacted - no plaintext secrets)
	Metadata JSONMap `gorm:"type:jsonb;not null;default:'{}'::jsonb"`

	// Success/failure tracking
	Success      bool   `gorm:"not null;default:true"`
	ErrorMessage string `gorm:"column:error_message;type:text"`

	// Timestamp
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "secrets_audit_log"
}

// BeforeCreate hook to ensure UUID is set if not provided
func (s *Secret) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook to ensure UUID is set if not provided
func (t *AccessToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook to ensure UUID is set if not provided
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// SecretDependency represents a link between a secret and a service/function that depends on it
// Used for impact analysis during rotation or deletion
type SecretDependency struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SecretID      uuid.UUID `gorm:"type:uuid;not null;index"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index"`
	DependentID   uuid.UUID `gorm:"type:uuid;not null"`
	DependentType string    `gorm:"not null;size:50"` // 'function', 'service', 'integration', 'workflow'
	DependentName string    `gorm:"not null;size:255"`
	Criticality   string    `gorm:"not null;size:20;default:'medium'"` // 'low', 'medium', 'high', 'critical'
	Metadata      JSONMap   `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for SecretDependency
func (SecretDependency) TableName() string {
	return "secret_dependencies"
}

// BeforeCreate hook to ensure UUID is set if not provided
func (sd *SecretDependency) BeforeCreate(tx *gorm.DB) error {
	if sd.ID == uuid.Nil {
		sd.ID = uuid.New()
	}
	return nil
}

// DependentType constants
const (
	DependentTypeFunction    = "function"
	DependentTypeService      = "service"
	DependentTypeIntegration  = "integration"
	DependentTypeWorkflow     = "workflow"
)

// Criticality constants
const (
	CriticalityLow      = "low"
	CriticalityMedium   = "medium"
	CriticalityHigh     = "high"
	CriticalityCritical = "critical"
)
