package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SAMLService handles SAML SSO operations
type SAMLService struct {
	configRepo  *storage.SAMLConfigRepository
	sessionRepo *storage.SAMLSessionRepository
	stateRepo   *storage.SAMLStateRepository
	repo        storage.Repository
	authService *AuthService
	baseURL     string
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
}

// SAMLServiceConfig holds SAML service configuration
type SAMLServiceConfig struct {
	ConfigRepo  *storage.SAMLConfigRepository
	SessionRepo *storage.SAMLSessionRepository
	StateRepo   *storage.SAMLStateRepository
	Repo        storage.Repository
	AuthService *AuthService
	BaseURL     string
}

// NewSAMLService creates a new SAML service
func NewSAMLService(config SAMLServiceConfig) (*SAMLService, error) {
	// Generate a default key pair if not provided
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create a self-signed certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "functionfly-saml",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &SAMLService{
		configRepo:  config.ConfigRepo,
		sessionRepo: config.SessionRepo,
		stateRepo:   config.StateRepo,
		repo:        config.Repo,
		authService: config.AuthService,
		baseURL:     config.BaseURL,
		privateKey:  privateKey,
		certificate: cert,
	}, nil
}

// GetSPMetadata returns the Service Provider metadata XML
func (s *SAMLService) GetSPMetadata(tenantID uuid.UUID) (string, error) {
	config, err := s.configRepo.GetByTenantID(tenantID)
	if err != nil {
		return "", err
	}

	// Generate metadata XML manually
	spEntityID := config.SPEntityID
	if spEntityID == "" {
		spEntityID = "functionfly"
	}

	acsURL := s.baseURL + "/v1/auth/saml/sso"
	if config.SPACSURL != nil && *config.SPACSURL != "" {
		acsURL = *config.SPACSURL
	}

	metadata := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <md:SPSSODescriptor AuthnRequestsSigned="false" WantAssertionsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:NameIDFormat>%s</md:NameIDFormat>
    <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s" index="0" isDefault="true"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>`, spEntityID, config.NameIDFormat, acsURL)

	return metadata, nil
}

// InitiateLogin initiates a SAML login by generating an AuthnRequest
func (s *SAMLService) InitiateLogin(tenantID uuid.UUID, relayState string) (string, error) {
	config, err := s.configRepo.GetByTenantID(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get SAML config: %w", err)
	}

	if !config.Enabled {
		return "", fmt.Errorf("SAML is not enabled for this tenant")
	}

	// Create state for CSRF protection
	state := uuid.New().String()
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.stateRepo.SaveAuthnRequestState(state, tenantID, relayState, expiresAt); err != nil {
		return "", fmt.Errorf("failed to save SAML state: %w", err)
	}

	// Build IdP URL
	idpURL := ""
	if config.IDPSSOURL != nil {
		idpURL = *config.IDPSSOURL
	}

	if idpURL == "" {
		return "", fmt.Errorf("IdP SSO URL not configured")
	}

	// Generate AuthnRequest
	requestID := "_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	spEntityID := config.SPEntityID
	if spEntityID == "" {
		spEntityID = "functionfly"
	}

	nameIDFormat := config.NameIDFormat
	if nameIDFormat == "" {
		nameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	}

	authnRequest := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="%s" Version="2.0" IssueInstant="%s" AssertionConsumerServiceURL="%s">
  <saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">%s</saml:Issuer>
  <samlp:NameIDPolicy Format="%s" AllowCreate="true"/>
</samlp:AuthnRequest>`, requestID, time.Now().UTC().Format(time.RFC3339), s.baseURL+"/v1/auth/saml/sso", spEntityID, nameIDFormat)

	// Encode the request
	encodedReq := base64.StdEncoding.EncodeToString([]byte(authnRequest))

	// Build the redirect URL
	redirectURL := fmt.Sprintf("%s?SAMLRequest=%s", idpURL, urlEncode(encodedReq))
	if relayState != "" {
		redirectURL += "&RelayState=" + urlEncode(relayState)
	}

	return redirectURL, nil
}

// SAMLResponse represents a parsed SAML response
type SAMLResponse struct {
	XMLName   xml.Name       `xml:"samlp:Response"`
	Assertion *SAMLAssertion `xml:"saml:Assertion"`
}

// SAMLAssertion represents a SAML assertion
type SAMLAssertion struct {
	XMLName             xml.Name                 `xml:"saml:Assertion"`
	ID                  string                   `xml:"ID,attr"`
	Subject             *SAMLSubject             `xml:"saml:Subject"`
	AttributeStatements []SAMLAttributeStatement `xml:"saml:AttributeStatement"`
}

// SAMLSubject represents the subject of a SAML assertion
type SAMLSubject struct {
	NameID *SAMLNameID `xml:"saml:NameID"`
}

// SAMLNameID represents a SAML NameID
type SAMLNameID struct {
	Value string `xml:",chardata"`
}

// SAMLAttributeStatement represents attribute statements
type SAMLAttributeStatement struct {
	Attributes []SAMLAttribute `xml:"saml:Attribute"`
}

// SAMLAttribute represents a SAML attribute
type SAMLAttribute struct {
	Name         string               `xml:"Name,attr"`
	FriendlyName string               `xml:"FriendlyName,attr"`
	Values       []SAMLAttributeValue `xml:"saml:AttributeValue"`
}

// SAMLAttributeValue represents an attribute value
type SAMLAttributeValue struct {
	Value string `xml:",chardata"`
}

// ProcessResponse processes a SAML Response from the IdP
func (s *SAMLService) ProcessResponse(tenantID uuid.UUID, samlResponse string) (*SAMLLoginResponse, error) {
	config, err := s.configRepo.GetByTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SAML config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("SAML is not enabled for this tenant")
	}

	// Decode the response
	decodedResponse, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse the SAML response
	var resp SAMLResponse
	if err := xml.Unmarshal(decodedResponse, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	if resp.Assertion == nil {
		return nil, fmt.Errorf("no assertion in SAML response")
	}

	// Extract user attributes
	email := ""
	nameID := ""

	// Get NameID
	if resp.Assertion.Subject != nil && resp.Assertion.Subject.NameID != nil {
		nameID = resp.Assertion.Subject.NameID.Value
	}

	// Get email from attributes
	for _, attrStmt := range resp.Assertion.AttributeStatements {
		for _, attr := range attrStmt.Attributes {
			if attr.Name == "email" || attr.Name == "Email" || attr.FriendlyName == "email" {
				if len(attr.Values) > 0 && attr.Values[0].Value != "" {
					email = attr.Values[0].Value
					break
				}
			}
		}
		if email != "" {
			break
		}
	}

	// Fallback to NameID if email not found
	if email == "" && nameID != "" && strings.Contains(nameID, "@") {
		email = nameID
	}

	if email == "" {
		return nil, fmt.Errorf("no email found in SAML response")
	}

	// Find or create the user
	user, err := s.findOrCreateSAMLUser(tenantID, email, nameID, resp.Assertion)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate JWT token
	token, err := s.authService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create SAML session
	sessionIndex := resp.Assertion.ID

	// Extract attributes as map
	attrs := make(map[string]interface{})
	for _, attrStmt := range resp.Assertion.AttributeStatements {
		for _, attr := range attrStmt.Attributes {
			if len(attr.Values) > 0 {
				if len(attr.Values) == 1 {
					attrs[attr.Name] = attr.Values[0].Value
				} else {
					values := make([]string, len(attr.Values))
					for i, v := range attr.Values {
						values[i] = v.Value
					}
					attrs[attr.Name] = values
				}
			}
		}
	}

	session := &storage.SAMLSession{
		ID:           uuid.New(),
		TenantID:     tenantID,
		UserID:       user.ID,
		SAMLNameID:   nameID,
		SessionIndex: sessionIndex,
		NotOnOrAfter: time.Now().Add(24 * time.Hour), // Default 24 hour session
		Attributes:   attrs,
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		logrus.WithError(err).Warn("Failed to create SAML session")
	}

	return &SAMLLoginResponse{
		User:        user,
		Token:       token,
		NameID:      nameID,
		SAMLSession: session,
	}, nil
}

// findOrCreateSAMLUser finds or creates a user based on SAML attributes
func (s *SAMLService) findOrCreateSAMLUser(tenantID uuid.UUID, email, nameID string, assertion *SAMLAssertion) (*storage.User, error) {
	// Try to find existing user by email
	user, err := s.repo.GetUserByEmail(email)
	if err == nil && user != nil {
		return user, nil
	}

	// User not found - create a new one
	// Extract name from attributes
	name := email
	for _, attrStmt := range assertion.AttributeStatements {
		for _, attr := range attrStmt.Attributes {
			if attr.Name == "name" || attr.Name == "displayName" || attr.FriendlyName == "displayName" {
				if len(attr.Values) > 0 && attr.Values[0].Value != "" {
					name = attr.Values[0].Value
					break
				}
			}
		}
	}

	// Create user with SAML provider
	providerData := map[string]interface{}{
		"saml": true,
	}
	user, err = s.repo.CreateUserWithSocialAuth(email, tenantID, "saml", nameID, providerData)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Update name if different from email
	if name != email {
		_, err = s.repo.UpdateUser(nil, user.ID, map[string]interface{}{"name": name})
		if err != nil {
			logrus.WithError(err).Warn("Failed to update user name")
		}
	}

	return user, nil
}

// GetConfig retrieves SAML configuration for a tenant
func (s *SAMLService) GetConfig(tenantID uuid.UUID) (*storage.SAMLConfig, error) {
	return s.configRepo.GetByTenantID(tenantID)
}

// SaveConfig saves SAML configuration for a tenant
func (s *SAMLService) SaveConfig(config *storage.SAMLConfig) error {
	// Set default values
	if config.SPEntityID == "" {
		config.SPEntityID = "functionfly"
	}
	if config.NameIDFormat == "" {
		config.NameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	}

	// Generate SP URLs if not provided
	baseURL := s.baseURL
	if config.SPACSURL == nil || *config.SPACSURL == "" {
		acsURL := baseURL + "/v1/auth/saml/sso"
		config.SPACSURL = &acsURL
	}
	if config.SPMetadataURL == nil || *config.SPMetadataURL == "" {
		metadataURL := baseURL + "/v1/auth/saml/metadata"
		config.SPMetadataURL = &metadataURL
	}

	// Try to get existing config
	existing, err := s.configRepo.GetByTenantID(config.TenantID)
	if err != nil {
		// Config doesn't exist, create new
		config.ID = uuid.New()
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		return s.configRepo.Create(config)
	}

	// Update existing config
	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()
	return s.configRepo.Update(config)
}

// HandleSLO handles Single Logout
func (s *SAMLService) HandleSLO(tenantID, userID uuid.UUID) error {
	// Delete SAML sessions for the user
	return s.sessionRepo.DeleteByUserID(userID)
}

// SAMLLoginResponse represents the response after successful SAML login
type SAMLLoginResponse struct {
	User        *storage.User
	Token       string
	NameID      string
	SAMLSession *storage.SAMLSession
}

// urlEncode performs simple URL encoding
func urlEncode(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			result += string(c)
		} else {
			result += fmt.Sprintf("%%%02X", byte(c))
		}
	}
	return result
}
