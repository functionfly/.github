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
