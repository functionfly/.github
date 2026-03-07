// Package webauthn provides WebAuthn/Passkeys authentication support for GoBetterAuth
// This is Phase 3 of the Better Auth migration plan - Passwordless authentication
package webauthn

import (
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebAuthnCredential stores a user's passkey credential
type WebAuthnCredential struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index"`
	CredentialID    []byte         `gorm:"type:bytea;not null;uniqueIndex:idx_gba_webauthn_cred_id"`
	PublicKey       []byte         `gorm:"type:bytea;not null"`
	AttestationType string         `gorm:"type:varchar(50);not null"`
	Transport       []string       `gorm:"type:text[]"`
	Flags           int            `gorm:"default:0"`
	Authenticator   []byte         `gorm:"type:jsonb"` // JSON encoded authenticator data
	SignCount       uint32         `gorm:"default:0"`
	Name            string         `gorm:"type:varchar(255);not null;default:'Passkey'"`
	CreatedAt       time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	LastUsedAt      *time.Time     `gorm:"default:null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the WebAuthnCredential model
func (WebAuthnCredential) TableName() string {
	return "gba_webauthn_credentials"
}

// BeforeCreate hook to set timestamps and ID
func (c *WebAuthnCredential) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.Must(uuid.NewRandom())
	}
	c.CreatedAt = time.Now()
	return nil
}

// UpdateSignCount updates the sign count and last used timestamp
func (c *WebAuthnCredential) UpdateSignCount(newCount uint32) {
	c.SignCount = newCount
	now := time.Now()
	c.LastUsedAt = &now
}

// GetAuthenticatorData returns the authenticator data as a map
func (c *WebAuthnCredential) GetAuthenticatorData() (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(c.Authenticator, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// SetAuthenticatorData stores authenticator data as JSON
func (c *WebAuthnCredential) SetAuthenticatorData(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.Authenticator = jsonData
	return nil
}

// ToWebAuthnCredential converts to webauthn.Credential
func (c *WebAuthnCredential) ToWebAuthnCredential() webauthn.Credential {
	transports := make([]protocol.AuthenticatorTransport, len(c.Transport))
	for i, t := range c.Transport {
		transports[i] = protocol.AuthenticatorTransport(t)
	}

	return webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Flags: webauthn.CredentialFlags{
			UserPresent:    c.Flags&1 != 0,
			UserVerified:   c.Flags&2 != 0,
			BackupEligible: c.Flags&4 != 0,
			BackupState:    c.Flags&8 != 0,
		},
		Authenticator: webauthn.Authenticator{
			SignCount: c.SignCount,
		},
	}
}

// WebAuthnSession stores registration/authentication session data
type WebAuthnSession struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	Challenge string         `gorm:"type:varchar(255);not null"`
	Operation string         `gorm:"type:varchar(20);not null"` // "registration" or "authentication"
	ExpiresAt time.Time      `gorm:"not null"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the WebAuthnSession model
func (WebAuthnSession) TableName() string {
	return "gba_webauthn_sessions"
}

// BeforeCreate hook to set timestamps and ID
func (s *WebAuthnSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.Must(uuid.NewRandom())
	}
	s.CreatedAt = time.Now()
	return nil
}

// IsExpired checks if the session has expired
func (s *WebAuthnSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// WebAuthnUser implements the webauthn.User interface
type WebAuthnUser struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Credentials []WebAuthnCredential
}

// WebAuthnID returns the user's ID as bytes
func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.ID.String())
}

// WebAuthnName returns the user's email
func (u *WebAuthnUser) WebAuthnName() string {
	return u.Email
}

// WebAuthnDisplayName returns the user's display name
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Email
}

// WebAuthnIcon returns the user's icon URL (optional)
func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns the user's credentials as webauthn.Credential slice
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, len(u.Credentials))
	for i, cred := range u.Credentials {
		credentials[i] = cred.ToWebAuthnCredential()
	}
	return credentials
}

// AddCredential adds a credential to the user
func (u *WebAuthnUser) AddCredential(cred WebAuthnCredential) {
	u.Credentials = append(u.Credentials, cred)
}

// UpdateCredential updates an existing credential's sign count
func (u *WebAuthnUser) UpdateCredential(credID []byte, signCount uint32) bool {
	for i := range u.Credentials {
		if string(u.Credentials[i].CredentialID) == string(credID) {
			u.Credentials[i].UpdateSignCount(signCount)
			return true
		}
	}
	return false
}

// Request/Response types for API

// BeginRegistrationRequest represents the request to start registration
type BeginRegistrationRequest struct {
	Name              string `json:"name"`                         // Device name (e.g., "MacBook Pro")
	AuthenticatorType string `json:"authenticator_type,omitempty"` // "platform" or "cross-platform"
	ResidentKey       bool   `json:"resident_key,omitempty"`       // Discoverable credential (passkey)
}

// BeginRegistrationResponse represents the response for registration start
type BeginRegistrationResponse struct {
	SessionID string                       `json:"sessionId"`
	Options   *protocol.CredentialCreation `json:"options"`
}

// FinishRegistrationRequest represents the request to complete registration
type FinishRegistrationRequest struct {
	SessionID string                              `json:"sessionId"`
	Response  protocol.CredentialCreationResponse `json:"response"`
}

// FinishRegistrationResponse represents the response for registration completion
type FinishRegistrationResponse struct {
	CredentialID string    `json:"credentialId"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	Message      string    `json:"message"`
}

// BeginAuthenticationResponse represents the response for authentication start
type BeginAuthenticationResponse struct {
	SessionID string      `json:"sessionId"`
	Options   interface{} `json:"options"`
}

// FinishAuthenticationRequest represents the request to complete authentication
type FinishAuthenticationRequest struct {
	SessionID string                               `json:"sessionId"`
	Response  protocol.CredentialAssertionResponse `json:"response"`
}

// FinishAuthenticationResponse represents the response for authentication completion
type FinishAuthenticationResponse struct {
	UserID  uuid.UUID `json:"userId"`
	Email   string    `json:"email"`
	Message string    `json:"message"`
}

// CredentialListResponse represents a list of credentials
type CredentialListResponse struct {
	Credentials []CredentialInfo `json:"credentials"`
}

// CredentialInfo represents credential information for listing
type CredentialInfo struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	SignCount  uint32     `json:"signCount"`
}

// UpdateCredentialRequest represents a request to update a credential name
type UpdateCredentialRequest struct {
	Name string `json:"name"`
}

// UpdateCredentialResponse represents the response for credential update
type UpdateCredentialResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// DeleteCredentialResponse represents the response for credential deletion
type DeleteCredentialResponse struct {
	Deleted bool   `json:"deleted"`
	Message string `json:"message"`
}

// WebAuthnStatusResponse represents the WebAuthn status for a user
type WebAuthnStatusResponse struct {
	Enabled         bool `json:"enabled"`
	CredentialCount int  `json:"credentialCount"`
}

// CredentialDescriptor represents a credential descriptor for the frontend
type CredentialDescriptor struct {
	Type         string   `json:"type"`
	CredentialID string   `json:"credentialId"`
	Transport    []string `json:"transport,omitempty"`
}

// WebAuthnConfig holds WebAuthn configuration
type WebAuthnConfig struct {
	RPDisplayName           string        // Display name for the Relying Party (e.g., "FunctionFly")
	RPID                    string        // Relying Party ID (e.g., "functionfly.com")
	RPOrigin                string        // Relying Party Origin (e.g., "https://app.functionfly.com")
	SessionTimeout          time.Duration // Timeout for WebAuthn sessions (default: 5m)
	AttestationPreference   string        // Attestation preference: "none", "indirect", "direct"
	AuthenticatorAttachment string        // "platform" or "cross-platform" (empty for no preference)
}

// DefaultWebAuthnConfig returns default WebAuthn configuration
func DefaultWebAuthnConfig() *WebAuthnConfig {
	return &WebAuthnConfig{
		RPDisplayName:           "FunctionFly",
		RPID:                    "localhost",
		RPOrigin:                "http://localhost:3000",
		SessionTimeout:          5 * time.Minute,
		AttestationPreference:   "none",
		AuthenticatorAttachment: "",
	}
}

// WebAuthnRepository defines the interface for WebAuthn data operations
type WebAuthnRepository interface {
	// Credential operations
	CreateCredential(cred *WebAuthnCredential) error
	GetCredentialByID(id uuid.UUID) (*WebAuthnCredential, error)
	GetCredentialByCredentialID(credentialID []byte) (*WebAuthnCredential, error)
	GetCredentialsByUserID(userID uuid.UUID) ([]WebAuthnCredential, error)
	UpdateCredential(cred *WebAuthnCredential) error
	DeleteCredential(id uuid.UUID) error

	// Session operations
	CreateSession(session *WebAuthnSession) error
	GetSessionByID(id uuid.UUID) (*WebAuthnSession, error)
	DeleteSession(id uuid.UUID) error
	DeleteExpiredSessions() error
}
