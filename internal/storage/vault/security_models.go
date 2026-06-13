package vault

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SecretStatus represents the lifecycle state of a secret in the vault.
type SecretStatus string

const (
	SecretStatusActive       SecretStatus = "active"
	SecretStatusExpiringSoon SecretStatus = "expiring_soon"
	SecretStatusExpired      SecretStatus = "expired"
	SecretStatusRevoked      SecretStatus = "revoked"
)

// KDFMethod identifies the key-derivation function used to derive
// the encryption key from a user's passphrase.
type KDFMethod string

const (
	KDFMethodPBKDF2SHA256 KDFMethod = "pbkdf2-sha256"
	KDFMethodArgon2id     KDFMethod = "argon2id"
)

// KeyVersion constants for secrets_vault.key_version.
// 1 = legacy PBKDF2-SHA256 (backwards compatible).
// 2 = Argon2id (OWASP 2023 recommended).
const (
	KeyVersionPBKDF2 = 1
	KeyVersionArgon2 = 2
)

// VaultMFAConfig holds per-tenant policy for MFA enforcement on vault
// operations. A row is created lazily on first read.
type VaultMFAConfig struct {
	TenantID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	MFARequired          bool      `gorm:"not null;default:false"`
	MFAMethod            string    `gorm:"size:20;not null;default:'totp'"`
	EnforceForTokens     bool      `gorm:"not null;default:false"`
	EnforceForAPI        bool      `gorm:"not null;default:false"`
	MFASessionTTLSeconds int       `gorm:"not null;default:900"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`
}

func (VaultMFAConfig) TableName() string { return "vault_mfa_config" }

func (v *VaultMFAConfig) BeforeCreate(tx *gorm.DB) error {
	if v.TenantID == uuid.Nil {
		v.TenantID = uuid.New()
	}
	return nil
}

// BreakGlassRequest represents an emergency access request used when
// normal authentication is unavailable.
type BreakGlassRequest struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	RequestedBy     uuid.UUID  `gorm:"type:uuid;not null"`
	ApprovedBy      *uuid.UUID `gorm:"type:uuid"`
	Reason          string     `gorm:"type:text;not null"`
	Status          string     `gorm:"size:20;not null;default:'pending'"`
	DurationMinutes int        `gorm:"not null;default:60"`
	ExpiresAt       time.Time  `gorm:"not null"`
	ApprovedAt      *time.Time
	RevokedAt       *time.Time
	Metadata        JSONMap   `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (BreakGlassRequest) TableName() string { return "break_glass_requests" }

func (b *BreakGlassRequest) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.Metadata == nil {
		b.Metadata = JSONMap{}
	}
	return nil
}

// IsActive returns true when a break-glass grant is still valid.
func (b *BreakGlassRequest) IsActive() bool {
	if b.Status != "approved" {
		return false
	}
	return time.Now().Before(b.ExpiresAt)
}

// BreakGlassConfig is the per-tenant break-glass policy.
type BreakGlassConfig struct {
	TenantID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	MaxDurationMinutes    int       `gorm:"not null;default:60"`
	RequiredApproverCount int       `gorm:"not null;default:1"`
	ApproverUserIDs       JSONMap   `gorm:"type:jsonb;not null;default:'[]'::jsonb"`
	Enabled               bool      `gorm:"not null;default:true"`
	CreatedAt             time.Time `gorm:"autoCreateTime"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime"`
}

func (BreakGlassConfig) TableName() string { return "break_glass_config" }

func (b *BreakGlassConfig) BeforeCreate(tx *gorm.DB) error {
	if b.TenantID == uuid.Nil {
		b.TenantID = uuid.New()
	}
	if b.ApproverUserIDs == nil {
		b.ApproverUserIDs = JSONMap{"user_ids": []string{}}
	}
	return nil
}

// VaultEscrowConfig stores the encrypted recovery blob used to reset a
// user's passphrase in the optional enterprise escrow flow.
type VaultEscrowConfig struct {
	TenantID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Enabled         bool       `gorm:"not null;default:false"`
	SecurityQHashes JSONMap    `gorm:"column:security_question_hashes;type:jsonb;not null;default:'[]'::jsonb"`
	KDFSalt         []byte     `gorm:"column:kdf_salt;type:bytea;not null"`
	KDFMethod       string     `gorm:"column:kdf_method;size:20;not null;default:'argon2id'"`
	KDFParams       JSONMap    `gorm:"column:kdf_params;type:jsonb;not null;default:'{}'::jsonb"`
	EncryptedBlob   []byte     `gorm:"column:encrypted_recovery_blob;type:bytea;not null"`
	BlobIV          []byte     `gorm:"column:blob_iv;type:bytea;not null"`
	BlobAuthTag     []byte     `gorm:"column:blob_auth_tag;type:bytea;not null"`
	BlobKeyVersion  int        `gorm:"column:blob_key_version;not null;default:2"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
	LastRecoveredAt *time.Time `gorm:"column:last_recovered_at"`
}

func (VaultEscrowConfig) TableName() string { return "vault_escrow_config" }

func (v *VaultEscrowConfig) BeforeCreate(tx *gorm.DB) error {
	if v.TenantID == uuid.Nil {
		v.TenantID = uuid.New()
	}
	if v.KDFParams == nil {
		v.KDFParams = JSONMap{}
	}
	if v.SecurityQHashes == nil {
		v.SecurityQHashes = JSONMap{"hashes": []string{}}
	}
	return nil
}
