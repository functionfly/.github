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

	// Secret identification
	Name        string `gorm:"not null;size:255"`
	Description string `gorm:"type:text"`

	// Secret type categorization
	SecretType SecretType `gorm:"not null;size:50;default:'api_key'"`

	// Encrypted data - never store plaintext
	// This is the AES-256-GCM encrypted blob (ciphertext + auth tag)
	EncryptedValue []byte `gorm:"type:bytea;not null"`

	// Encryption metadata
	EncryptionSalt string `gorm:"not null;size:255"`  // PBKDF2 salt (base64)
	IV             string `gorm:"not null;size:255"`  // AES-GCM IV/nonce (base64)
	KeyVersion     int    `gorm:"not null;default:1"` // 1=passphrase, 2=KMS, 3=HSM

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
}

// TableName specifies the table name for Secret
func (Secret) TableName() string {
	return "secrets_vault"
}

// IsDeleted returns true if the secret has been soft-deleted
func (s *Secret) IsDeleted() bool {
	return s.DeletedAt != nil
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
