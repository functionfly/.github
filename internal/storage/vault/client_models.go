package vault

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// 2026-06-16 design: per-tenant DEK + client-encrypted dynamic credentials
// ============================================================================

// VaultTenantKey is the per-user wrapped DEK for a tenant. The DEK
// itself never leaves the client; the row stores ciphertext produced
// by the user's passphrase-derived KEK (Argon2id).
type VaultTenantKey struct {
	TenantID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	WrappedDEK  []byte    `gorm:"column:wrapped_dek;type:bytea;not null"`
	DEKIV       []byte    `gorm:"column:dek_iv;type:bytea;not null"`
	DEKAuthTag  []byte    `gorm:"column:dek_auth_tag;type:bytea;not null"`
	DEKSalt     []byte    `gorm:"column:dek_salt;type:bytea;not null"`
	KeyVersion  int       `gorm:"column:key_version;not null;default:3"`
	KDFParams   JSONMap   `gorm:"column:kdf_params;type:jsonb;not null;default:'{\"t\":3,\"m\":65536,\"p\":4}'::jsonb"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	RotatedAt   *time.Time `gorm:"column:rotated_at"`
}

func (VaultTenantKey) TableName() string { return "vault_tenant_keys" }

// VaultUserKey reserves per-user public key material for v2 asymmetric
// key wrapping. v1 creates the table but does not use it.
type VaultUserKey struct {
	TenantID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	PublicKey []byte    `gorm:"column:public_key;type:bytea;not null"`
	KeyType   string    `gorm:"column:key_type;size:32;not null;default:'x25519'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (VaultUserKey) TableName() string { return "vault_user_keys" }

// DynamicWrappedAccessToken is a CI/CD-agent token scoped to a single
// dynamic credential. The raw token is returned once at creation; only
// its SHA-256 hash is persisted.
type DynamicWrappedAccessToken struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID             uuid.UUID  `gorm:"type:uuid;not null;index"`
	CredentialID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash            string     `gorm:"column:token_hash;size:255;not null;uniqueIndex"`
	Name                 string     `gorm:"size:255;not null"`
	Scopes               JSONMap    `gorm:"type:jsonb;not null;default:'[]'::jsonb"`
	ExpiresAt            time.Time  `gorm:"not null"`
	IsRevoked            bool       `gorm:"column:is_revoked;not null;default:false"`
	RevokedAt            *time.Time `gorm:"column:revoked_at"`
	RevokedReason        string     `gorm:"column:revoked_reason;type:text"`
	AllowedIPs           StringArray `gorm:"column:allowed_ips;type:jsonb;not null;default:'[]'::jsonb"`
	DeniedIPs            StringArray `gorm:"column:denied_ips;type:jsonb;not null;default:'[]'::jsonb"`
	IPRestrictionEnabled bool        `gorm:"column:ip_restriction_enabled;not null;default:false"`
	LastUsedAt           *time.Time `gorm:"column:last_used_at"`
	UseCount             int        `gorm:"not null;default:0"`
	CreatedBy            uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt            time.Time  `gorm:"autoCreateTime"`
}

func (DynamicWrappedAccessToken) TableName() string { return "dynamic_wrapped_access_tokens" }

func (t *DynamicWrappedAccessToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Scopes == nil {
		t.Scopes = JSONMap{"scopes": []string{}}
	}
	if t.AllowedIPs == nil {
		t.AllowedIPs = StringArray{}
	}
	if t.DeniedIPs == nil {
		t.DeniedIPs = StringArray{}
	}
	return nil
}

// IsValid reports whether the token is not expired and not revoked.
func (t *DynamicWrappedAccessToken) IsValid() bool {
	if t.IsRevoked || t.RevokedAt != nil {
		return false
	}
	return time.Now().Before(t.ExpiresAt)
}

// DynamicTargetShare is the v2 cross-tenant share stub. v1 endpoints
// return 501 / 404 for this entity; the table is created so v2 is
// purely additive.
type DynamicTargetShare struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TargetID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	SourceTenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	GrantedToTenantID uuid.UUID  `gorm:"type:uuid;not null;index"`
	GrantedByUser     uuid.UUID  `gorm:"type:uuid;not null"`
	Permissions       string     `gorm:"size:20;not null;default:'read'"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
	RevokedAt         *time.Time `gorm:"column:revoked_at"`
	RevokedBy         *uuid.UUID `gorm:"type:uuid"`
	CreatedAt         time.Time  `gorm:"autoCreateTime"`
}

func (DynamicTargetShare) TableName() string { return "dynamic_target_shares" }

func (s *DynamicTargetShare) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
