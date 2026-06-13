package vault

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// Phase 4.3: Namespaces
// ============================================================================

// VaultNamespace is a hierarchical path under a tenant. Paths use
// lowercase + /-separators (e.g. "production/api-gateway").
type VaultNamespace struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Path        string     `gorm:"size:512;not null"`
	Description string     `gorm:"type:text"`
	ParentID    *uuid.UUID `gorm:"type:uuid"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

func (VaultNamespace) TableName() string { return "vault_namespaces" }

func (n *VaultNamespace) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// IsAncestorOf reports whether the receiver is an ancestor of `child`.
// Two namespaces are considered identical, not ancestor/descendant.
func (n *VaultNamespace) IsAncestorOf(child *VaultNamespace) bool {
	if n == nil || child == nil {
		return false
	}
	if n.Path == child.Path {
		return false
	}
	return hasPathPrefix(child.Path, n.Path+"/")
}

// ============================================================================
// Phase 4.1: RBAC
// ============================================================================

// VaultRole is a named set of JSONB permissions.
type VaultRole struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name        string     `gorm:"size:100;not null"`
	Description string     `gorm:"type:text"`
	Permissions JSONMap    `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	IsBuiltin   bool       `gorm:"column:is_builtin;not null;default:false"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

func (VaultRole) TableName() string { return "vault_roles" }

func (r *VaultRole) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Permissions == nil {
		r.Permissions = JSONMap{}
	}
	return nil
}

// VaultRoleAssignment binds a user (or team) to a role, optionally
// scoped to a namespace path. An empty Scope means "all namespaces".
type VaultRoleAssignment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	RoleID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	TeamID    *uuid.UUID `gorm:"type:uuid"`
	Scope     string     `gorm:"size:512;not null;default:'all'"`
	CreatedBy uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (VaultRoleAssignment) TableName() string { return "vault_role_assignments" }

// EffectiveScope returns the scope that gates this assignment, defaulting
// to "all" when the field is empty.
func (a *VaultRoleAssignment) EffectiveScope() string {
	if a == nil || a.Scope == "" {
		return "all"
	}
	return a.Scope
}

func (a *VaultRoleAssignment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Phase 4.4: Cross-tenant shares
// ============================================================================

// VaultShare grants another tenant read or read-write access to a
// secret. Grantees see the secret in their `shared/` namespace.
type VaultShare struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SecretID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	SourceTenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	GrantedToTenantID uuid.UUID  `gorm:"type:uuid;not null;index"`
	GrantedByUser     uuid.UUID  `gorm:"type:uuid;not null"`
	Permissions       string     `gorm:"size:20;not null;default:'read'"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
	RevokedAt         *time.Time `gorm:"column:revoked_at"`
	RevokedBy         *uuid.UUID `gorm:"type:uuid"`
	CreatedAt         time.Time  `gorm:"autoCreateTime"`
}

func (VaultShare) TableName() string { return "vault_shares" }

func (s *VaultShare) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsActive reports whether the share is still usable.
func (s *VaultShare) IsActive(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	if s.ExpiresAt != nil && now.After(*s.ExpiresAt) {
		return false
	}
	return true
}

// ============================================================================
// Phase 4.5: SSO
// ============================================================================

// VaultSSOConfig is the per-tenant SAML SSO configuration.
type VaultSSOConfig struct {
	TenantID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Enabled                bool       `gorm:"not null;default:false"`
	SAMLMetadataURL        string     `gorm:"column:saml_metadata_url;type:text"`
	SAMLEntityID           string     `gorm:"column:saml_entity_id;size:255"`
	SAMLSSOURL             string     `gorm:"column:saml_sso_url;type:text"`
	SAMLSLOURL             string     `gorm:"column:saml_slo_url;type:text"`
	SAMLX509Cert           string     `gorm:"column:saml_x509_cert;type:text"`
	DefaultRoleID          *uuid.UUID `gorm:"type:uuid"`
	JITProvisioningEnabled bool       `gorm:"column:jit_provisioning_enabled;not null;default:true"`
	AttributeRoleMapping   JSONMap    `gorm:"column:attribute_role_mapping;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt              time.Time  `gorm:"autoCreateTime"`
	UpdatedAt              time.Time  `gorm:"autoUpdateTime"`
}

func (VaultSSOConfig) TableName() string { return "vault_sso_config" }

func (s *VaultSSOConfig) BeforeCreate(tx *gorm.DB) error {
	if s.AttributeRoleMapping == nil {
		s.AttributeRoleMapping = JSONMap{}
	}
	return nil
}

// ============================================================================
// Phase 4.2: SIEM webhooks
// ============================================================================

// VaultSIEMWebhook is an outbound push target for real-time audit
// event delivery.
type VaultSIEMWebhook struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name               string     `gorm:"size:100;not null"`
	URL                string     `gorm:"type:text;not null"`
	SecretHMAC         []byte     `gorm:"column:secret_hmac;type:bytea;not null"`
	Format             string     `gorm:"size:20;not null;default:'json'"`
	Enabled            bool       `gorm:"not null;default:true"`
	LastDeliveryAt     *time.Time `gorm:"column:last_delivery_at"`
	LastDeliveryStatus *int       `gorm:"column:last_delivery_status"`
	LastDeliveryError  string     `gorm:"column:last_delivery_error;type:text"`
	CreatedBy          uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`
}

func (VaultSIEMWebhook) TableName() string { return "vault_siem_webhooks" }

func (w *VaultSIEMWebhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
