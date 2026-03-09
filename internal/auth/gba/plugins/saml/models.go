// Package saml provides SAML 2.0 SSO authentication support for GoBetterAuth
// This is Phase 4 of the Better Auth migration plan - Enterprise SSO
package saml

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SAMLConfig stores SAML Identity Provider configuration for a tenant
type SAMLConfig struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_gba_saml_config_tenant"`
	Enabled        bool           `gorm:"default:false"`
	IDPEntityID    string         `gorm:"type:varchar(500);not null;column:idp_entity_id"`
	IDPSSOURL      string         `gorm:"type:varchar(500);not null;column:idp_sso_url"`
	IDPSLOURL      string         `gorm:"type:varchar(500);column:idp_slo_url"` // optional Single Logout URL
	IDPCertificate string         `gorm:"type:text;not null;column:idp_certificate"` // PEM encoded
	SPEntityID     string         `gorm:"type:varchar(500);not null;column:sp_entity_id"`
	ACSURL         string         `gorm:"type:varchar(500);not null;column:acs_url"`
	NameIDFormat   string         `gorm:"type:varchar(100);default:'emailAddress';column:name_id_format"`
	CreatedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the SAMLConfig model
func (SAMLConfig) TableName() string {
	return "gba_saml_configs"
}

// BeforeCreate hook to set timestamps and ID
func (c *SAMLConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.Must(uuid.NewRandom())
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update the updated_at timestamp
func (c *SAMLConfig) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// SAMLSession stores active SAML sessions for single logout support
type SAMLSession struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index;column:tenant_id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index;column:user_id"`
	NameID       string         `gorm:"type:varchar(255);not null;column:name_id"`
	SessionIndex string         `gorm:"type:varchar(255);not null;column:session_index"`
	ExpiresAt    time.Time      `gorm:"not null;column:expires_at"`
	CreatedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the SAMLSession model
func (SAMLSession) TableName() string {
	return "gba_saml_sessions"
}

// BeforeCreate hook to set timestamps and ID
func (s *SAMLSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.Must(uuid.NewRandom())
	}
	s.CreatedAt = time.Now()
	return nil
}

// IsExpired checks if the session has expired
func (s *SAMLSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SAMLAttributeMapping defines how SAML attributes map to user fields
type SAMLAttributeMapping struct {
	Email      string `json:"email"`      // Attribute name for email
	FirstName  string `json:"first_name"` // Attribute name for first name
	LastName   string `json:"last_name"`  // Attribute name for last name
	Groups     string `json:"groups"`     // Attribute name for groups/roles
	Department string `json:"department"` // Attribute name for department
}

// DefaultAttributeMapping returns the default SAML attribute mapping
func DefaultAttributeMapping() *SAMLAttributeMapping {
	return &SAMLAttributeMapping{
		Email:      "email",
		FirstName:  "firstName",
		LastName:   "lastName",
		Groups:     "groups",
		Department: "department",
	}
}

// SAMLAssertion represents a parsed SAML assertion
type SAMLAssertion struct {
	NameID       string
	SessionIndex string
	Attributes   map[string][]string
	NotBefore    time.Time
	NotOnOrAfter time.Time
	Audience     string
	Issuer       string
	InResponseTo string
}

// GetEmail extracts the email from SAML attributes
func (a *SAMLAssertion) GetEmail(mapping *SAMLAttributeMapping) string {
	if mapping == nil {
		mapping = DefaultAttributeMapping()
	}

	// Try mapped attribute first
	if emails, ok := a.Attributes[mapping.Email]; ok && len(emails) > 0 {
		return emails[0]
	}

	// Fallback to common attribute names
	for _, attr := range []string{"email", "mail", "Email", "emailAddress", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"} {
		if emails, ok := a.Attributes[attr]; ok && len(emails) > 0 {
			return emails[0]
		}
	}

	// If NameID is an email format, use that
	if a.NameID != "" && contains(a.NameID, "@") {
		return a.NameID
	}

	return ""
}

// GetFirstName extracts the first name from SAML attributes
func (a *SAMLAssertion) GetFirstName(mapping *SAMLAttributeMapping) string {
	if mapping == nil {
		mapping = DefaultAttributeMapping()
	}

	if names, ok := a.Attributes[mapping.FirstName]; ok && len(names) > 0 {
		return names[0]
	}

	// Fallback to common attribute names
	for _, attr := range []string{"firstName", "first_name", "givenName", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"} {
		if names, ok := a.Attributes[attr]; ok && len(names) > 0 {
			return names[0]
		}
	}

	return ""
}

// GetLastName extracts the last name from SAML attributes
func (a *SAMLAssertion) GetLastName(mapping *SAMLAttributeMapping) string {
	if mapping == nil {
		mapping = DefaultAttributeMapping()
	}

	if names, ok := a.Attributes[mapping.LastName]; ok && len(names) > 0 {
		return names[0]
	}

	// Fallback to common attribute names
	for _, attr := range []string{"lastName", "last_name", "surname", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"} {
		if names, ok := a.Attributes[attr]; ok && len(names) > 0 {
			return names[0]
		}
	}

	return ""
}

// GetGroups extracts groups from SAML attributes
func (a *SAMLAssertion) GetGroups(mapping *SAMLAttributeMapping) []string {
	if mapping == nil {
		mapping = DefaultAttributeMapping()
	}

	if groups, ok := a.Attributes[mapping.Groups]; ok {
		return groups
	}

	// Fallback to common attribute names
	for _, attr := range []string{"groups", "roles", "memberOf", "http://schemas.xmlsoap.org/claims/Group"} {
		if groups, ok := a.Attributes[attr]; ok {
			return groups
		}
	}

	return nil
}

// IsValid checks if the assertion is valid (not expired, correct audience)
func (a *SAMLAssertion) IsValid(expectedAudience string) bool {
	now := time.Now()

	// Check NotBefore
	if now.Before(a.NotBefore) {
		return false
	}

	// Check NotOnOrAfter
	if now.After(a.NotOnOrAfter) {
		return false
	}

	// Check audience if provided
	if expectedAudience != "" && a.Audience != expectedAudience {
		return false
	}

	return true
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SAMLConfigRequest represents a request to configure SAML for a tenant
type SAMLConfigRequest struct {
	IDPEntityID    string `json:"idp_entity_id" validate:"required"`
	IDPSSOURL      string `json:"idp_sso_url" validate:"required,url"`
	IDPSLOURL      string `json:"idp_slo_url,omitempty"`
	IDPCertificate string `json:"idp_certificate" validate:"required"`
	SPEntityID     string `json:"sp_entity_id,omitempty"`
	ACSURL         string `json:"acs_url,omitempty"`
	NameIDFormat   string `json:"name_id_format,omitempty"`
}

// SAMLConfigResponse represents the SAML configuration response
type SAMLConfigResponse struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Enabled      bool      `json:"enabled"`
	IDPEntityID  string    `json:"idp_entity_id"`
	IDPSSOURL    string    `json:"idp_sso_url"`
	IDPSLOURL    string    `json:"idp_slo_url,omitempty"`
	SPEntityID   string    `json:"sp_entity_id"`
	ACSURL       string    `json:"acs_url"`
	NameIDFormat string    `json:"name_id_format"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SAMLMetadataResponse represents the Service Provider metadata
type SAMLMetadataResponse struct {
	EntityID    string `json:"entity_id"`
	ACSURL      string `json:"acs_url"`
	SLOURL      string `json:"slo_url,omitempty"`
	Certificate string `json:"certificate,omitempty"`
	XML         string `json:"xml"`
}

// SAMLLoginResponse represents the response from initiating SAML login
type SAMLLoginResponse struct {
	AuthURL     string `json:"auth_url"`
	RequestID   string `json:"request_id"`
	Binding     string `json:"binding"` // "redirect" or "post"
	SAMLRequest string `json:"saml_request,omitempty"`
	RelayState  string `json:"relay_state,omitempty"`
}

// SAMLACSResponse represents the response after ACS callback
type SAMLACSResponse struct {
	Success     bool   `json:"success"`
	UserID      string `json:"user_id,omitempty"`
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
	Message     string `json:"message,omitempty"`
	Token       string `json:"token,omitempty"`
}

// SAMLStatusResponse represents the SAML status for a tenant
type SAMLStatusResponse struct {
	Enabled     bool   `json:"enabled"`
	Configured  bool   `json:"configured"`
	IDPEntityID string `json:"idp_entity_id,omitempty"`
	SPEntityID  string `json:"sp_entity_id,omitempty"`
}
