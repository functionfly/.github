package vault

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
)

// EncryptedDataPayload represents the encrypted secret data sent by the client
// The actual encryption is performed client-side; this is the encrypted payload
type EncryptedDataPayload struct {
	Ciphertext string `json:"ciphertext"`  // base64 encoded encrypted data
	IV         string `json:"iv"`          // base64 encoded initialization vector
	Salt       string `json:"salt"`        // base64 encoded PBKDF2 salt
	Tag        string `json:"tag"`         // base64 encoded authentication tag
	KeyVersion int    `json:"key_version"` // encryption key version (1=passphrase, 2=KMS, 3=HSM)
}

// CreateSecretRequest represents a request to create a new secret
type CreateSecretRequest struct {
	Name          string                 `json:"name" validate:"required"`
	Description   string                 `json:"description,omitempty"`
	SecretType    vault.SecretType       `json:"secret_type" validate:"required"`
	EncryptedData EncryptedDataPayload   `json:"encrypted_data" validate:"required"`
	Scopes        []string               `json:"scopes,omitempty"` // access scopes for the secret
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateSecretRequest represents a request to partially update a secret
// All fields are optional for partial updates
type UpdateSecretRequest struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Scopes      *[]string `json:"scopes,omitempty"`
}

// SecretResponse represents a secret in API responses
type SecretResponse struct {
	ID             uuid.UUID              `json:"id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	SecretType     vault.SecretType       `json:"secret_type"`
	EncryptedData  EncryptedDataPayload   `json:"encrypted_data"`
	Scopes         []string               `json:"scopes,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	LastAccessedAt *time.Time             `json:"last_accessed_at,omitempty"`
	AccessCount    int                    `json:"access_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CurrentVersion int                    `json:"current_version,omitempty"`
	LastModifiedAt *time.Time             `json:"last_modified_at,omitempty"`
}

// SecretMetadataResponse represents a secret without encrypted data (for list views)
type SecretMetadataResponse struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	SecretType     vault.SecretType       `json:"secret_type"`
	Scopes         []string               `json:"scopes,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	LastAccessedAt *time.Time             `json:"last_accessed_at,omitempty"`
	AccessCount    int                    `json:"access_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CurrentVersion int                    `json:"current_version,omitempty"`
	LastModifiedAt *time.Time             `json:"last_modified_at,omitempty"`
}

// ListSecretsResponse represents the response for listing secrets
type ListSecretsResponse struct {
	Secrets []SecretMetadataResponse `json:"secrets"`
	Total   int64                    `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
}

// GenerateTokenRequest represents a request to generate a new access token
type GenerateTokenRequest struct {
	SecretID       uuid.UUID `json:"secret_id" validate:"required"`
	Scopes         []string  `json:"scopes,omitempty"`
	ExpiresInHours int       `json:"expires_in_hours" validate:"min=1,max=8760"` // max 1 year
	Name           string    `json:"name,omitempty"`
}

// GenerateTokenResponse represents the response with the generated token
// The token is shown only once at creation time
type GenerateTokenResponse struct {
	TokenID   uuid.UUID `json:"token_id"`
	Token     string    `json:"token"` // plaintext token, shown once
	SecretID  uuid.UUID `json:"secret_id"`
	Name      string    `json:"name,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	Scopes    []string  `json:"scopes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenResponse represents an access token in API responses (without the actual token)
type TokenResponse struct {
	ID            uuid.UUID  `json:"id"`
	SecretID      uuid.UUID  `json:"secret_id"`
	Name          string     `json:"name,omitempty"`
	Scopes        []string   `json:"scopes,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	IsRevoked     bool       `json:"is_revoked"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason string     `json:"revoked_reason,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	UseCount      int        `json:"use_count"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ListTokensResponse represents the response for listing tokens
type ListTokensResponse struct {
	Tokens []TokenResponse `json:"tokens"`
	Total  int64           `json:"total"`
}

// AuditLogEntryResponse represents a single audit log entry
type AuditLogEntryResponse struct {
	ID           uuid.UUID              `json:"id"`
	SecretID     *uuid.UUID             `json:"secret_id,omitempty"`
	Action       vault.AuditAction      `json:"action"`
	ActorID      string                 `json:"actor_id"`
	ActorType    vault.ActorType        `json:"actor_type"`
	RequestID    string                 `json:"request_id,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ListAuditLogResponse represents the response for listing audit logs
type ListAuditLogResponse struct {
	Entries []AuditLogEntryResponse `json:"entries"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SecretVersionResponse represents a secret version in API responses
type SecretVersionResponse struct {
	ID            uuid.UUID              `json:"id"`
	SecretID      uuid.UUID              `json:"secret_id"`
	VersionNumber int                    `json:"version_number"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	SecretType    vault.SecretType       `json:"secret_type"`
	EncryptedData EncryptedDataPayload   `json:"encrypted_data,omitempty"` // Only included when explicitly requested
	Scopes        []string               `json:"scopes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ChangeType    string                 `json:"change_type"`
	ChangeSummary string                 `json:"change_summary,omitempty"`
	ActorID       uuid.UUID              `json:"actor_id"`
	ActorType     vault.ActorType        `json:"actor_type"`
	CreatedAt     time.Time              `json:"created_at"`
}

// SecretVersionMetadataResponse represents a version without encrypted data (for list views)
type SecretVersionMetadataResponse struct {
	ID            uuid.UUID              `json:"id"`
	SecretID      uuid.UUID              `json:"secret_id"`
	VersionNumber int                    `json:"version_number"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	SecretType    vault.SecretType       `json:"secret_type"`
	Scopes        []string               `json:"scopes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ChangeType    string                 `json:"change_type"`
	ChangeSummary string                 `json:"change_summary,omitempty"`
	ActorID       uuid.UUID              `json:"actor_id"`
	ActorType     vault.ActorType        `json:"actor_type"`
	CreatedAt     time.Time              `json:"created_at"`
}

// ListSecretVersionsResponse represents the response for listing secret versions
type ListSecretVersionsResponse struct {
	Versions []SecretVersionMetadataResponse `json:"versions"`
	Total    int64                           `json:"total"`
	Limit    int                             `json:"limit"`
	Offset   int                             `json:"offset"`
}

// SecretVersionDiffResponse represents a diff between two versions
type SecretVersionDiffResponse struct {
	FromVersion      int             `json:"from_version"`
	ToVersion        int             `json:"to_version"`
	HasChanges       bool            `json:"has_changes"`
	NameChanged      bool            `json:"name_changed"`
	NameFrom         string          `json:"name_from,omitempty"`
	NameTo           string          `json:"name_to,omitempty"`
	DescChanged      bool            `json:"description_changed"`
	DescFrom         string          `json:"description_from,omitempty"`
	DescTo           string          `json:"description_to,omitempty"`
	ScopesChanged    bool            `json:"scopes_changed"`
	ScopesFrom       []string        `json:"scopes_from,omitempty"`
	ScopesTo         []string        `json:"scopes_to,omitempty"`
	EncryptedChanged bool            `json:"encrypted_value_changed"`
	ChangeSummary    string          `json:"change_summary,omitempty"`
	ActorID          uuid.UUID       `json:"actor_id"`
	ActorType        vault.ActorType `json:"actor_type"`
	CreatedAt        time.Time       `json:"created_at"`
}

// RollbackSecretRequest represents a request to rollback a secret to a previous version
type RollbackSecretRequest struct {
	TargetVersion int    `json:"target_version" validate:"required,min=1"`
	Reason        string `json:"reason,omitempty"`
}

// RollbackSecretResponse represents the response after a successful rollback
type RollbackSecretResponse struct {
	Secret       SecretResponse        `json:"secret"`
	NewVersion   SecretVersionResponse `json:"new_version"`
	RolledBackTo int                   `json:"rolled_back_to"`
	Message      string                `json:"message"`
}

// RotateSecretRequest represents a request to rotate a secret's encrypted value
type RotateSecretRequest struct {
	EncryptedData EncryptedDataPayload `json:"encrypted_data" validate:"required"`
	Reason        string               `json:"reason,omitempty"`
}

// BulkDeleteRequest represents a request to delete multiple secrets
type BulkDeleteRequest struct {
	SecretIDs []uuid.UUID `json:"secret_ids" validate:"required,min=1"`
	DryRun    bool        `json:"dry_run,omitempty"`
}

// BulkDeleteResponse represents the response from a bulk delete operation
type BulkDeleteResponse struct {
	Deleted  int64                           `json:"deleted"`
	Failed   int64                           `json:"failed"`
	Errors   []BulkDeleteError               `json:"errors,omitempty"`
	Previews map[uuid.UUID]BulkDeletePreview `json:"previews,omitempty"`
}

// BulkDeletePreview represents preview information for a single secret in bulk delete
type BulkDeletePreview struct {
	SecretID     uuid.UUID        `json:"secret_id"`
	SecretName   string           `json:"secret_name"`
	Found        bool             `json:"found"`
	TokensCount  int              `json:"tokens_count"`
	Dependencies []DependencyInfo `json:"dependencies"`
}

// DependencyInfo describes a single dependency
type DependencyInfo struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Criticality string    `json:"criticality"`
}

// BulkDeleteError represents an error that occurred during bulk delete
type BulkDeleteError struct {
	SecretID uuid.UUID `json:"secret_id"`
	Error    string    `json:"error"`
}

// ExportSecretsResponse represents the response for exporting secrets
type ExportSecretsResponse struct {
	Secrets    []SecretExport `json:"secrets"`
	Total      int            `json:"total"`
	ExportedAt time.Time      `json:"exported_at"`
}

// SecretExport represents secret metadata for export (no encrypted values)
type SecretExport struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	SecretType  vault.SecretType       `json:"secret_type"`
	KeyVersion  int                    `json:"key_version"`
	Scopes      []string               `json:"scopes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SecretDependenciesResponse represents the response for getting secret dependencies
type SecretDependenciesResponse struct {
	SecretID     uuid.UUID          `json:"secret_id"`
	Dependencies []SecretDependency `json:"dependencies"`
	Total        int64              `json:"total"`
}

// SecretDependency represents a dependency in API responses
type SecretDependency struct {
	ID            uuid.UUID              `json:"id"`
	DependentID   uuid.UUID              `json:"dependent_id"`
	DependentType string                 `json:"dependent_type"`
	DependentName string                 `json:"dependent_name"`
	Criticality   string                 `json:"criticality"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// CreateSecretDependencyRequest represents a request to create a secret dependency
type CreateSecretDependencyRequest struct {
	SecretID      uuid.UUID              `json:"secret_id" validate:"required"`
	DependentID   uuid.UUID              `json:"dependent_id" validate:"required"`
	DependentType string                 `json:"dependent_type" validate:"required"`
	DependentName string                 `json:"dependent_name" validate:"required"`
	Criticality   string                 `json:"criticality,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ============================================================================
// Phase 1.1: Vault MFA
// ============================================================================

// MFAConfigResponse is the response shape for vault MFA config.
type MFAConfigResponse struct {
	TenantID             uuid.UUID `json:"tenant_id"`
	MFARequired          bool      `json:"mfa_required"`
	MFAMethod            string    `json:"mfa_method"`
	EnforceForTokens     bool      `json:"enforce_for_tokens"`
	EnforceForAPI        bool      `json:"enforce_for_api"`
	MFASessionTTLSeconds int       `json:"mfa_session_ttl_seconds"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// UpdateMFAConfigRequest is a partial-update payload for MFA config.
type UpdateMFAConfigRequest struct {
	MFARequired          *bool  `json:"mfa_required,omitempty"`
	MFAMethod            string `json:"mfa_method,omitempty"`
	EnforceForTokens     *bool  `json:"enforce_for_tokens,omitempty"`
	EnforceForAPI        *bool  `json:"enforce_for_api,omitempty"`
	MFASessionTTLSeconds int    `json:"mfa_session_ttl_seconds,omitempty"`
}

// ============================================================================
// Phase 1.2: Token IP allowlist
// ============================================================================

// UpdateTokenIPPolicyRequest sets IP restrictions on a token.
type UpdateTokenIPPolicyRequest struct {
	AllowedIPs []string `json:"allowed_ips"`
	DeniedIPs  []string `json:"denied_ips"`
	Enabled    bool     `json:"enabled"`
}

// ============================================================================
// Phase 1.3: Secret expiration
// ============================================================================

// SetExpirationRequest sets a TTL on a secret.
type SetExpirationRequest struct {
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	ExpireAfterDays *int       `json:"expire_after_days,omitempty"`
}

// ============================================================================
// Phase 1.4: Break-glass
// ============================================================================

// BreakGlassRequestBody is the body of a break-glass request.
type BreakGlassRequestBody struct {
	Reason          string `json:"reason"`
	DurationMinutes int    `json:"duration_minutes"`
}

// BreakGlassResponse is the API response for break-glass endpoints.
type BreakGlassResponse struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	RequestedBy     uuid.UUID  `json:"requested_by"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty"`
	Reason          string     `json:"reason"`
	Status          string     `json:"status"`
	DurationMinutes int        `json:"duration_minutes"`
	ExpiresAt       time.Time  `json:"expires_at"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// BreakGlassConfigResponse is the per-tenant policy.
type BreakGlassConfigResponse struct {
	TenantID              uuid.UUID `json:"tenant_id"`
	MaxDurationMinutes    int       `json:"max_duration_minutes"`
	RequiredApproverCount int       `json:"required_approver_count"`
	Enabled               bool      `json:"enabled"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ============================================================================
// Phase 1.4b: Escrow
// ============================================================================

// EnableEscrowRequest registers a new escrow row.
type EnableEscrowRequest struct {
	SecurityQuestionHashes []string `json:"security_question_hashes"`
	KDFSalt                string   `json:"kdf_salt"`       // base64
	EncryptedBlob          string   `json:"encrypted_blob"` // base64
	BlobIV                 string   `json:"blob_iv"`        // base64
	BlobAuthTag            string   `json:"blob_auth_tag"`  // base64
}

// EscrowStatusResponse is the API response for escrow.
type EscrowStatusResponse struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	Enabled        bool      `json:"enabled"`
	KDFMethod      string    `json:"kdf_method"`
	BlobKeyVersion int       `json:"blob_key_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ============================================================================
// Phase 2: Dynamic secrets
// ============================================================================

// CreateTargetRequest creates a new dynamic secret target (database
// admin connection).
type CreateTargetRequest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	DBType            string   `json:"db_type"` // postgres | mysql
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	DatabaseName      string   `json:"database_name"`
	AdminUsername     string   `json:"admin_username"`
	AdminPassword     string   `json:"admin_password"`
	SSLMode           string   `json:"ssl_mode,omitempty"`
	AllowedRoles      []string `json:"allowed_roles,omitempty"`
	DefaultTTLSeconds int      `json:"default_ttl_seconds,omitempty"`
	MaxTTLSeconds     int      `json:"max_ttl_seconds,omitempty"`
}

// TargetResponse is the public projection of a DynamicSecretTarget
// (never includes the encrypted admin password).
type TargetResponse struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	DBType            string     `json:"db_type"`
	Host              string     `json:"host"`
	Port              int        `json:"port"`
	DatabaseName      string     `json:"database_name"`
	AdminUsername     string     `json:"admin_username"`
	SSLMode           string     `json:"ssl_mode"`
	AllowedRoles      []string   `json:"allowed_roles,omitempty"`
	DefaultTTLSeconds int        `json:"default_ttl_seconds"`
	MaxTTLSeconds     int        `json:"max_ttl_seconds"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
}

// CreateCredentialRequest creates a new credential template.
type CreateCredentialRequest struct {
	TargetID      string `json:"target_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	RoleTemplate  string `json:"role_template,omitempty"`
	TTLSeconds    int    `json:"ttl_seconds,omitempty"`
	MaxTTLSeconds int    `json:"max_ttl_seconds,omitempty"`
}

// CredentialResponse is the public projection of a DynamicCredential
// (template).
type CredentialResponse struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	TargetID      uuid.UUID `json:"target_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	RoleTemplate  string    `json:"role_template,omitempty"`
	TTLSeconds    int       `json:"ttl_seconds"`
	MaxTTLSeconds int       `json:"max_ttl_seconds"`
	Status        string    `json:"status"`
	CreatedBy     uuid.UUID `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GeneratedCredentialResponse is the API response when a credential
// is freshly generated. The password is returned exactly once.
type GeneratedCredentialResponse struct {
	LeaseID    string             `json:"lease_id"`
	Username   string             `json:"username"`
	Password   string             `json:"password"`
	Host       string             `json:"host"`
	Port       int                `json:"port"`
	Database   string             `json:"database"`
	ExpiresAt  time.Time          `json:"expires_at"`
	Credential CredentialResponse `json:"credential"`
	Target     TargetResponse     `json:"target"`
}

// ============================================================================
// Phase 4: Enterprise features
// ============================================================================

// CreateNamespaceRequest registers a new hierarchical namespace.
type CreateNamespaceRequest struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
}

// NamespaceResponse is the public projection of a VaultNamespace.
type NamespaceResponse struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Path        string     `json:"path"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateRoleRequest creates a new RBAC role.
type CreateRoleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions"`
}

// UpdateRoleRequest partially updates an RBAC role.
type UpdateRoleRequest struct {
	Description *string                `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// RoleResponse is the public projection of a VaultRole.
type RoleResponse struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions"`
	IsBuiltin   bool                   `json:"is_builtin"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AssignRoleRequest binds a user to a role, optionally scoped.
type AssignRoleRequest struct {
	UserID string `json:"user_id"`
	Scope  string `json:"scope,omitempty"`
}

// AssignmentResponse is the public projection of a VaultRoleAssignment.
type AssignmentResponse struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	RoleID    uuid.UUID  `json:"role_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Scope     string     `json:"scope"`
	CreatedBy uuid.UUID  `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
}

// ShareSecretRequest shares a secret with another tenant.
type ShareSecretRequest struct {
	GranteeTenantID string     `json:"grantee_tenant_id"`
	Permissions     string     `json:"permissions,omitempty"` // read | read-write
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// ShareResponse is the public projection of a VaultShare.
type ShareResponse struct {
	ID                uuid.UUID  `json:"id"`
	SecretID          uuid.UUID  `json:"secret_id"`
	SourceTenantID    uuid.UUID  `json:"source_tenant_id"`
	GrantedToTenantID uuid.UUID  `json:"granted_to_tenant_id"`
	GrantedByUser     uuid.UUID  `json:"granted_by_user"`
	Permissions       string     `json:"permissions"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// UpdateSSORequest updates the SAML/SSO configuration.
type UpdateSSORequest struct {
	Enabled                *bool                  `json:"enabled,omitempty"`
	SAMLMetadataURL        string                 `json:"saml_metadata_url,omitempty"`
	SAMLEntityID           string                 `json:"saml_entity_id,omitempty"`
	SAMLSSOURL             string                 `json:"saml_sso_url,omitempty"`
	SAMLSLOURL             string                 `json:"saml_slo_url,omitempty"`
	SAMLX509Cert           string                 `json:"saml_x509_cert,omitempty"`
	JITProvisioningEnabled *bool                  `json:"jit_provisioning_enabled,omitempty"`
	AttributeRoleMapping   map[string]interface{} `json:"attribute_role_mapping,omitempty"`
}

// SSOConfigResponse is the public projection of a VaultSSOConfig
// (the X.509 cert is omitted; fetch separately if needed).
type SSOConfigResponse struct {
	TenantID               uuid.UUID              `json:"tenant_id"`
	Enabled                bool                   `json:"enabled"`
	SAMLMetadataURL        string                 `json:"saml_metadata_url,omitempty"`
	SAMLEntityID           string                 `json:"saml_entity_id,omitempty"`
	SAMLSSOURL             string                 `json:"saml_sso_url,omitempty"`
	SAMLSLOURL             string                 `json:"saml_slo_url,omitempty"`
	JITProvisioningEnabled bool                   `json:"jit_provisioning_enabled"`
	AttributeRoleMapping   map[string]interface{} `json:"attribute_role_mapping,omitempty"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// CreateSIEMWebhookRequest registers a new SIEM webhook.
type CreateSIEMWebhookRequest struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Format string `json:"format,omitempty"` // json | cef
}

// SIEMWebhookResponse is the public projection of a VaultSIEMWebhook.
type SIEMWebhookResponse struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	Name               string     `json:"name"`
	URL                string     `json:"url"`
	Format             string     `json:"format"`
	Enabled            bool       `json:"enabled"`
	LastDeliveryAt     *time.Time `json:"last_delivery_at,omitempty"`
	LastDeliveryStatus *int       `json:"last_delivery_status,omitempty"`
	LastDeliveryError  string     `json:"last_delivery_error,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	// SecretHMAC is only populated on the create response; it is the
	// shared secret the receiver uses to verify X-Signature.
	SecretHMAC string `json:"secret_hmac,omitempty"`
}
