package vault

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// SecretType represents the type of secret being stored
type SecretType string

const (
	// SecretTypeAPIKey for API keys
	SecretTypeAPIKey SecretType = "api_key"
	// SecretTypeOAuthToken for OAuth tokens
	SecretTypeOAuthToken SecretType = "oauth_token"
	// SecretTypePassword for passwords
	SecretTypePassword SecretType = "password"
	// SecretTypeCertificate for certificates
	SecretTypeCertificate SecretType = "certificate"
)

// Valid checks if the SecretType is valid
func (s SecretType) Valid() bool {
	switch s {
	case SecretTypeAPIKey, SecretTypeOAuthToken, SecretTypePassword, SecretTypeCertificate:
		return true
	}
	return false
}

// String returns the string representation
func (s SecretType) String() string {
	return string(s)
}

// AuditAction represents the type of action performed on a secret
type AuditAction string

const (
	// AuditActionCreate for secret creation
	AuditActionCreate AuditAction = "create"
	// AuditActionRead for secret access/reading
	AuditActionRead AuditAction = "read"
	// AuditActionUpdate for secret modification
	AuditActionUpdate AuditAction = "update"
	// AuditActionDelete for secret deletion
	AuditActionDelete AuditAction = "delete"
	// AuditActionUse for secret usage (e.g., in function execution)
	AuditActionUse AuditAction = "use"
	// AuditActionRevoke for token revocation
	AuditActionRevoke AuditAction = "revoke"
	// AuditActionVersion for secret versioning (snapshot created)
	AuditActionVersion AuditAction = "version"
	// AuditActionRollback for secret rollback to a previous version
	AuditActionRollback AuditAction = "rollback"
	// AuditActionExpire for secret expiration (Phase 1.3)
	AuditActionExpire AuditAction = "expire"
	// AuditActionBreakGlass for break-glass emergency access (Phase 1.4)
	AuditActionBreakGlass AuditAction = "break_glass"
	// AuditActionMFAVerify for vault MFA verification (Phase 1.1)
	AuditActionMFAVerify AuditAction = "mfa_verify"
	// AuditActionDYNTokenCreate for creating a dynamic wrapped access token.
	AuditActionDYNTokenCreate AuditAction = "dyn_token_create"
	// AuditActionDYNTokenRevoke for revoking a dynamic wrapped access token.
	AuditActionDYNTokenRevoke AuditAction = "dyn_token_revoke"
	// AuditActionClientDekInit for initializing a client-side data encryption key.
	AuditActionClientDekInit AuditAction = "client_dek_init"
	// AuditActionClientWrapRotate for rotating a wrapped client DEK.
	AuditActionClientWrapRotate AuditAction = "client_dek_wrap_rotate"
	// AuditActionClientDekShare for sharing a wrapped DEK with another user.
	AuditActionClientDekShare AuditAction = "client_dek_share"
)

// Valid checks if the AuditAction is valid
func (a AuditAction) Valid() bool {
	switch a {
	case AuditActionCreate, AuditActionRead, AuditActionUpdate, AuditActionDelete, AuditActionUse,
		AuditActionRevoke, AuditActionVersion, AuditActionRollback, AuditActionExpire,
		AuditActionBreakGlass, AuditActionMFAVerify,
		AuditActionDYNTokenCreate, AuditActionDYNTokenRevoke,
		AuditActionClientDekInit, AuditActionClientWrapRotate, AuditActionClientDekShare:
		return true
	}
	return false
}

// String returns the string representation
func (a AuditAction) String() string {
	return string(a)
}

// ActorType represents the type of actor performing an action
type ActorType string

const (
	// ActorTypeUser for human users
	ActorTypeUser ActorType = "user"
	// ActorTypeToken for access tokens
	ActorTypeToken ActorType = "token"
	// ActorTypeSystem for system operations
	ActorTypeSystem ActorType = "system"
	// ActorTypeAPIKey for API key authentication
	ActorTypeAPIKey ActorType = "api_key"
)

// Valid checks if the ActorType is valid
func (a ActorType) Valid() bool {
	switch a {
	case ActorTypeUser, ActorTypeToken, ActorTypeSystem, ActorTypeAPIKey:
		return true
	}
	return false
}

// String returns the string representation
func (a ActorType) String() string {
	return string(a)
}

// EncryptedData represents the structure of encrypted secret data
// This is the format stored in the encrypted_value column
// The actual encryption happens client-side; this is the serialized form
type EncryptedData struct {
	// Ciphertext is the encrypted data (base64 encoded)
	Ciphertext string `json:"ciphertext"`
	// IV is the initialization vector/nonce used for encryption (base64 encoded)
	IV string `json:"iv"`
	// Salt is the PBKDF2 salt used for key derivation (base64 encoded)
	Salt string `json:"salt"`
	// Version indicates the encryption scheme version
	Version int `json:"version"`
	// Algorithm specifies the encryption algorithm used
	Algorithm string `json:"algorithm"`
}

// JSONMap for JSONB columns (same pattern as statefabric)
type JSONMap map[string]interface{}

// Scan implements sql.Scanner so the DB driver can read JSONB into JSONMap
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	if len(bytes) == 0 {
		*m = JSONMap{}
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}
	*m = out
	return nil
}

// Value implements driver.Valuer so JSONMap is stored as JSON in JSONB columns
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// StringArray for simple string array storage in JSONB
type StringArray []string

// Scan implements sql.Scanner for StringArray
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	if len(bytes) == 0 {
		*a = StringArray{}
		return nil
	}
	var out []string
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}
	*a = out
	return nil
}

// Value implements driver.Valuer for StringArray
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(a)
}
