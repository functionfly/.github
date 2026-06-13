package vault

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DynamicSecretDBType is the underlying database engine of a target.
type DynamicSecretDBType string

const (
	DynamicSecretDBPostgres DynamicSecretDBType = "postgres"
	DynamicSecretDBMySQL    DynamicSecretDBType = "mysql"
)

// Valid reports whether the value is a known DB type.
func (d DynamicSecretDBType) Valid() bool {
	switch d {
	case DynamicSecretDBPostgres, DynamicSecretDBMySQL:
		return true
	}
	return false
}

// DynamicSecretTarget is a managed database connection used to mint
// dynamic credentials. The admin password is encrypted server-side
// using the platform envelope (see internal/crypto/server_encryption.go).
type DynamicSecretTarget struct {
	ID                     uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID               uuid.UUID           `gorm:"type:uuid;not null;index"`
	Name                   string              `gorm:"size:255;not null"`
	Description            string              `gorm:"type:text"`
	DBType                 DynamicSecretDBType `gorm:"column:db_type;size:20;not null"`
	Host                   string              `gorm:"size:255;not null"`
	Port                   int                 `gorm:"not null"`
	DatabaseName           string              `gorm:"column:database_name;size:255;not null"`
	AdminUsername          string              `gorm:"column:admin_username;size:255;not null"`
	EncryptedAdminPassword []byte              `gorm:"column:encrypted_admin_password;type:bytea;not null"`
	PasswordNonce          []byte              `gorm:"column:password_nonce;type:bytea;not null"`
	PasswordKeyVersion     int                 `gorm:"column:password_key_version;not null;default:1"`
	SSLMode                string              `gorm:"column:ssl_mode;size:20;not null;default:'require'"`
	AllowedRoles           StringArray         `gorm:"column:allowed_roles;type:jsonb;not null;default:'[]'::jsonb"`
	DefaultTTLSeconds      int                 `gorm:"column:default_ttl_seconds;not null;default:3600"`
	MaxTTLSeconds          int                 `gorm:"column:max_ttl_seconds;not null;default:86400"`
	Status                 string              `gorm:"size:20;not null;default:'active'"`
	CreatedBy              uuid.UUID           `gorm:"type:uuid;not null"`
	CreatedAt              time.Time           `gorm:"autoCreateTime"`
	UpdatedAt              time.Time           `gorm:"autoUpdateTime"`
	LastUsedAt             *time.Time          `gorm:"column:last_used_at"`
	LastError              string              `gorm:"column:last_error;type:text"`
}

func (DynamicSecretTarget) TableName() string { return "dynamic_secret_targets" }

func (d *DynamicSecretTarget) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.AllowedRoles == nil {
		d.AllowedRoles = StringArray{}
	}
	return nil
}

// DynamicCredential is a named, reusable template that mints temporary
// users against a target. The actual lease is recorded separately in
// DynamicCredentialLease.
type DynamicCredential struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index"`
	TargetID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Name          string    `gorm:"size:255;not null"`
	Description   string    `gorm:"type:text"`
	RoleTemplate  string    `gorm:"column:role_template;size:100"`
	TTLSeconds    int       `gorm:"column:ttl_seconds;not null;default:3600"`
	MaxTTLSeconds int       `gorm:"column:max_ttl_seconds;not null;default:86400"`
	Status        string    `gorm:"size:20;not null;default:'active'"`
	CreatedBy     uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (DynamicCredential) TableName() string { return "dynamic_credentials" }

func (d *DynamicCredential) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// DynamicCredentialLease is a single issuance of a dynamic credential.
// One lease = one DB user. The background worker drops the user when
// the lease expires or is revoked.
type DynamicCredentialLease struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LeaseID          string     `gorm:"column:lease_id;size:64;not null;uniqueIndex"`
	CredentialID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	TargetID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	TenantID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	DBUsername       string     `gorm:"column:db_username;size:128;not null"`
	ExpiresAt        time.Time  `gorm:"not null"`
	RenewedAt        *time.Time `gorm:"column:renewed_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	RevocationReason string     `gorm:"column:revocation_reason;size:100"`
	LastUsedAt       *time.Time `gorm:"column:last_used_at"`
	UseCount         int        `gorm:"not null;default:0"`
	IssuedTo         *uuid.UUID `gorm:"type:uuid"`
	IssuedIP         string     `gorm:"column:issued_ip;size:45"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
}

func (DynamicCredentialLease) TableName() string { return "dynamic_credential_leases" }

func (d *DynamicCredentialLease) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// IsActive reports whether the lease is still valid (not revoked and
// not past its expiry).
func (d *DynamicCredentialLease) IsActive(now time.Time) bool {
	if d.RevokedAt != nil {
		return false
	}
	return now.Before(d.ExpiresAt)
}
