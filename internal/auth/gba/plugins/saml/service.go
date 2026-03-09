// Package saml provides SAML 2.0 SSO authentication support for GoBetterAuth
package saml

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SAMLService handles SAML Service Provider operations
type SAMLService struct {
	db       *gorm.DB
	logger   *logrus.Logger
	spConfig *SPConfig
}

// SPConfig holds Service Provider configuration
type SPConfig struct {
	EntityID       string
	ACSURL         string
	SLOURL         string
	PrivateKey     *rsa.PrivateKey
	Certificate    *x509.Certificate
	CertificatePEM string
	PrivateKeyPEM  string
}

// NewSAMLService creates a new SAML service
func NewSAMLService(db *gorm.DB, logger *logrus.Logger, entityID, acsURL string) (*SAMLService, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Generate or load SP key pair
	spConfig, err := generateSPConfig(entityID, acsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SP config: %w", err)
	}

	return &SAMLService{
		db:       db,
		logger:   logger,
		spConfig: spConfig,
	}, nil
}

// generateSPConfig generates a self-signed certificate for the SP
func generateSPConfig(entityID, acsURL string) (*SPConfig, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"FunctionFly"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"functionfly.com"},
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return &SPConfig{
		EntityID:       entityID,
		ACSURL:         acsURL,
		PrivateKey:     privateKey,
		Certificate:    cert,
		CertificatePEM: string(certPEM),
		PrivateKeyPEM:  string(keyPEM),
	}, nil
}

// GetSPConfig returns the Service Provider configuration
func (s *SAMLService) GetSPConfig() *SPConfig {
	return s.spConfig
}

// GetConfig retrieves SAML configuration for a tenant
func (s *SAMLService) GetConfig(ctx context.Context, tenantID uuid.UUID) (*SAMLConfig, error) {
	var config SAMLConfig
	result := s.db.WithContext(ctx).Where("tenant_id = ? AND enabled = ? AND deleted_at IS NULL", tenantID, true).First(&config)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("SAML not configured for tenant")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get SAML config: %w", result.Error)
	}
	return &config, nil
}

// GetConfigByID retrieves SAML configuration by ID
func (s *SAMLService) GetConfigByID(ctx context.Context, configID uuid.UUID) (*SAMLConfig, error) {
	var config SAMLConfig
	result := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", configID).First(&config)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("SAML config not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get SAML config: %w", result.Error)
	}
	return &config, nil
}

// CreateConfig creates a new SAML configuration for a tenant
func (s *SAMLService) CreateConfig(ctx context.Context, tenantID uuid.UUID, req *SAMLConfigRequest) (*SAMLConfig, error) {
	// Validate the IdP certificate
	if _, err := parseCertificate(req.IDPCertificate); err != nil {
		return nil, fmt.Errorf("invalid IdP certificate: %w", err)
	}

	// Use defaults if not provided
	spEntityID := req.SPEntityID
	if spEntityID == "" {
		spEntityID = s.spConfig.EntityID
	}

	acsURL := req.ACSURL
	if acsURL == "" {
		acsURL = s.spConfig.ACSURL
	}

	nameIDFormat := req.NameIDFormat
	if nameIDFormat == "" {
		nameIDFormat = "emailAddress"
	}

	config := &SAMLConfig{
		TenantID:       tenantID,
		Enabled:        true,
		IDPEntityID:    req.IDPEntityID,
		IDPSSOURL:      req.IDPSSOURL,
		IDPSLOURL:      req.IDPSLOURL,
		IDPCertificate: req.IDPCertificate,
		SPEntityID:     spEntityID,
		ACSURL:         acsURL,
		NameIDFormat:   nameIDFormat,
	}

	if err := s.db.WithContext(ctx).Create(config).Error; err != nil {
		return nil, fmt.Errorf("failed to create SAML config: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"config_id": config.ID,
	}).Info("SAML configuration created")

	return config, nil
}

// UpdateConfig updates an existing SAML configuration
func (s *SAMLService) UpdateConfig(ctx context.Context, configID uuid.UUID, req *SAMLConfigRequest) (*SAMLConfig, error) {
	config, err := s.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	// Validate the IdP certificate
	if _, err := parseCertificate(req.IDPCertificate); err != nil {
		return nil, fmt.Errorf("invalid IdP certificate: %w", err)
	}

	config.IDPEntityID = req.IDPEntityID
	config.IDPSSOURL = req.IDPSSOURL
	config.IDPCertificate = req.IDPCertificate
	config.IDPSLOURL = req.IDPSLOURL

	if req.SPEntityID != "" {
		config.SPEntityID = req.SPEntityID
	}
	if req.ACSURL != "" {
		config.ACSURL = req.ACSURL
	}
	if req.NameIDFormat != "" {
		config.NameIDFormat = req.NameIDFormat
	}

	if err := s.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, fmt.Errorf("failed to update SAML config: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"config_id": configID,
	}).Info("SAML configuration updated")

	return config, nil
}

// DeleteConfig soft-deletes a SAML configuration
func (s *SAMLService) DeleteConfig(ctx context.Context, configID uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&SAMLConfig{}, "id = ?", configID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete SAML config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("SAML config not found")
	}

	s.logger.WithFields(logrus.Fields{
		"config_id": configID,
	}).Info("SAML configuration deleted")

	return nil
}

// BuildAuthnRequest creates a SAML AuthnRequest for the given tenant
func (s *SAMLService) BuildAuthnRequest(ctx context.Context, tenantID uuid.UUID) (*AuthnRequest, string, error) {
	config, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, "", err
	}

	requestID := generateRequestID()
	issueInstant := time.Now().UTC().Format(time.RFC3339)

	authnRequest := &AuthnRequest{
		XMLName: xml.Name{
			Local: "samlp:AuthnRequest",
		},
		ID:                          requestID,
		Version:                     "2.0",
		IssueInstant:                issueInstant,
		Destination:                 config.IDPSSOURL,
		ProtocolBinding:             "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
		AssertionConsumerServiceURL: config.ACSURL,
		Issuer: Issuer{
			Value: config.SPEntityID,
		},
		NameIDPolicy: NameIDPolicy{
			Format:      formatNameID(config.NameIDFormat),
			AllowCreate: "true",
		},
	}

	return authnRequest, requestID, nil
}

// BuildRedirectURL creates a redirect URL with SAMLRequest parameter
func (s *SAMLService) BuildRedirectURL(authnRequest *AuthnRequest, idpSSOURL string) (string, error) {
	// Marshal the AuthnRequest to XML
	xmlData, err := xml.MarshalIndent(authnRequest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal AuthnRequest: %w", err)
	}

	// Add XML declaration and namespace
	xmlData = append([]byte(xml.Header), xmlData...)
	xmlData = append([]byte("<samlp:AuthnRequest xmlns:samlp=\"urn:oasis:names:tc:SAML:2.0:protocol\" xmlns:saml=\"urn:oasis:names:tc:SAML:2.0:assertion\">\n"), xmlData[38:]...)

	// Deflate and base64 encode
	var compressedBuf strings.Builder
	flateWriter, _ := flate.NewWriter(&compressedBuf, flate.DefaultCompression)
	flateWriter.Write(xmlData)
	flateWriter.Close()

	encodedRequest := base64.StdEncoding.EncodeToString([]byte(compressedBuf.String()))

	// Build URL
	u, err := url.Parse(idpSSOURL)
	if err != nil {
		return "", fmt.Errorf("invalid IdP SSO URL: %w", err)
	}

	q := u.Query()
	q.Set("SAMLRequest", encodedRequest)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ParseSAMLResponse parses and validates a SAML response
func (s *SAMLService) ParseSAMLResponse(ctx context.Context, tenantID uuid.UUID, samlResponse string) (*SAMLAssertion, error) {
	config, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse the SAML response XML
	var response SAMLResponse
	if err := xml.Unmarshal(decoded, &response); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Check status
	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return nil, fmt.Errorf("SAML authentication failed: %s", response.Status.StatusCode.Value)
	}

	// Parse the assertion
	if len(response.Assertions) == 0 {
		return nil, fmt.Errorf("no assertions found in SAML response")
	}

	assertion := response.Assertions[0]

	// Validate Issuer matches IdP Entity ID
	if assertion.Issuer.Value != config.IDPEntityID {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", config.IDPEntityID, assertion.Issuer.Value)
	}

	// Extract attributes
	attributes := make(map[string][]string)
	for _, attr := range assertion.AttributeStatement.Attributes {
		values := make([]string, 0, len(attr.AttributeValues))
		for _, val := range attr.AttributeValues {
			values = append(values, val.Value)
		}
		attributes[attr.Name] = values
	}

	// Parse conditions
	notBefore, _ := time.Parse(time.RFC3339, assertion.Conditions.NotBefore)
	notOnOrAfter, _ := time.Parse(time.RFC3339, assertion.Conditions.NotOnOrAfter)

	samlAssertion := &SAMLAssertion{
		NameID:       assertion.Subject.NameID.Value,
		SessionIndex: assertion.AuthnStatement.SessionIndex,
		Attributes:   attributes,
		NotBefore:    notBefore,
		NotOnOrAfter: notOnOrAfter,
		Audience:     "",
		Issuer:       assertion.Issuer.Value,
		InResponseTo: response.InResponseTo,
	}

	// Extract audience
	if len(assertion.Conditions.AudienceRestrictions) > 0 &&
		len(assertion.Conditions.AudienceRestrictions[0].Audiences) > 0 {
		samlAssertion.Audience = assertion.Conditions.AudienceRestrictions[0].Audiences[0].Value
	}

	// Validate the assertion
	if !samlAssertion.IsValid(config.SPEntityID) {
		return nil, fmt.Errorf("SAML assertion is not valid (expired or wrong audience)")
	}

	return samlAssertion, nil
}

// CreateSession creates a SAML session for a user
func (s *SAMLService) CreateSession(ctx context.Context, tenantID, userID uuid.UUID, assertion *SAMLAssertion) (*SAMLSession, error) {
	session := &SAMLSession{
		TenantID:     tenantID,
		UserID:       userID,
		NameID:       assertion.NameID,
		SessionIndex: assertion.SessionIndex,
		ExpiresAt:    assertion.NotOnOrAfter,
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create SAML session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id":     tenantID,
		"user_id":       userID,
		"session_index": assertion.SessionIndex,
	}).Info("SAML session created")

	return session, nil
}

// DeleteSession deletes a SAML session
func (s *SAMLService) DeleteSession(ctx context.Context, sessionIndex string) error {
	result := s.db.WithContext(ctx).Where("session_index = ?", sessionIndex).Delete(&SAMLSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete SAML session: %w", result.Error)
	}
	return nil
}

// ParseLogoutRequest decodes and parses a SAML LogoutRequest (base64, optionally deflate-compressed).
// Returns the NameID and zero or more SessionIndex values from the request.
func (s *SAMLService) ParseLogoutRequest(samlRequest string) (nameID string, sessionIndices []string, err error) {
	if samlRequest == "" {
		return "", nil, fmt.Errorf("empty SAMLRequest")
	}
	decoded, err := base64.StdEncoding.DecodeString(samlRequest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode SAMLRequest: %w", err)
	}
	raw := decoded
	// Try deflate decompression if XML unmarshal fails (many IdPs send deflate-compressed)
	var req LogoutRequest
	if err := xml.Unmarshal(raw, &req); err != nil {
		// Try raw deflate (no zlib header)
		deflateReader := flate.NewReader(bytes.NewReader(decoded))
		decompressed, errRead := io.ReadAll(deflateReader)
		deflateReader.Close()
		if errRead != nil {
			return "", nil, fmt.Errorf("failed to parse SAML LogoutRequest: %w", err)
		}
		raw = decompressed
		if err := xml.Unmarshal(raw, &req); err != nil {
			return "", nil, fmt.Errorf("failed to parse SAML LogoutRequest: %w", err)
		}
	}
	nameID = strings.TrimSpace(req.NameID.Value)
	for _, si := range req.SessionIndex {
		if v := strings.TrimSpace(si.Value); v != "" {
			sessionIndices = append(sessionIndices, v)
		}
	}
	return nameID, sessionIndices, nil
}

// GetSessionsForLogout finds SAML sessions for a tenant matching the logout request.
// If sessionIndices is non-empty, sessions are matched by session_index; otherwise by name_id.
func (s *SAMLService) GetSessionsForLogout(ctx context.Context, tenantID uuid.UUID, nameID string, sessionIndices []string) ([]SAMLSession, error) {
	var sessions []SAMLSession
	q := s.db.WithContext(ctx).Model(&SAMLSession{}).Where("tenant_id = ? AND expires_at > ?", tenantID, time.Now())
	if len(sessionIndices) > 0 {
		q = q.Where("session_index IN ?", sessionIndices)
	} else {
		q = q.Where("name_id = ?", nameID)
	}
	if err := q.Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to find SAML sessions: %w", err)
	}
	return sessions, nil
}

// InvalidateSessionByToken deletes the gba_sessions row for the given session token (used when initiating SLO).
func (s *SAMLService) InvalidateSessionByToken(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	result := s.db.WithContext(ctx).Exec("DELETE FROM gba_sessions WHERE session_token = ?", token)
	if result.Error != nil {
		return fmt.Errorf("failed to invalidate session: %w", result.Error)
	}
	return nil
}

// GetUserFromSessionToken returns user_id and tenant_id for a valid session token, or error if not found.
func (s *SAMLService) GetUserFromSessionToken(ctx context.Context, token string) (userID, tenantID uuid.UUID, err error) {
	if token == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("empty session token")
	}
	var out struct {
		UserID   uuid.UUID
		TenantID uuid.UUID
	}
	err = s.db.WithContext(ctx).Raw(
		"SELECT user_id, tenant_id FROM gba_sessions WHERE session_token = ? AND expires_at > ?",
		token, time.Now(),
	).Scan(&out).Error
	if err != nil || out.UserID == uuid.Nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("session not found or expired")
	}
	return out.UserID, out.TenantID, nil
}

// GetSAMLSessionForUser returns one SAML session for the given tenant and user (for building LogoutRequest).
func (s *SAMLService) GetSAMLSessionForUser(ctx context.Context, tenantID, userID uuid.UUID) (*SAMLSession, error) {
	var sess SAMLSession
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND expires_at > ?", tenantID, userID, time.Now()).First(&sess).Error
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// BuildLogoutRequestRedirectURL builds a SAML LogoutRequest and returns the IdP SLO redirect URL (HTTP-Redirect binding).
func (s *SAMLService) BuildLogoutRequestRedirectURL(ctx context.Context, tenantID uuid.UUID, nameID, sessionIndex string) (string, error) {
	config, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if config.IDPSLOURL == "" {
		return "", fmt.Errorf("IdP SLO URL not configured")
	}
	req := &LogoutRequest{
		XMLName:      xml.Name{Local: "LogoutRequest"},
		ID:           generateRequestID(),
		Version:      "2.0",
		IssueInstant: time.Now().UTC().Format(time.RFC3339),
		Issuer:       Issuer{Value: config.SPEntityID},
		NameID:       NameID{Value: nameID},
		SessionIndex: []LogoutRequestSessionIndex{{Value: sessionIndex}},
	}
	xmlData, err := xml.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LogoutRequest: %w", err)
	}
	// Deflate and base64 encode
	var compressedBuf strings.Builder
	flateWriter, _ := flate.NewWriter(&compressedBuf, flate.DefaultCompression)
	flateWriter.Write(xmlData)
	flateWriter.Close()
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(compressedBuf.String()))
	u, err := url.Parse(config.IDPSLOURL)
	if err != nil {
		return "", fmt.Errorf("invalid IdP SLO URL: %w", err)
	}
	q := u.Query()
	q.Set("SAMLRequest", encodedRequest)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// InvalidateGBASessionsForUsers deletes all gba_sessions for the given user IDs (used on SLO).
func (s *SAMLService) InvalidateGBASessionsForUsers(ctx context.Context, userIDs []uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}
	// Use raw SQL to avoid importing gba package (avoids circular dependency)
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "DELETE FROM gba_sessions WHERE user_id IN (" + strings.Join(placeholders, ",") + ")"
	result := s.db.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return fmt.Errorf("failed to invalidate GBA sessions: %w", result.Error)
	}
	s.logger.WithField("user_count", len(userIDs)).WithField("rows_affected", result.RowsAffected).Info("Invalidated GBA sessions for SLO")
	return nil
}

// CleanupExpiredSessions removes expired SAML sessions
func (s *SAMLService) CleanupExpiredSessions(ctx context.Context) error {
	result := s.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&SAMLSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}
	s.logger.WithField("count", result.RowsAffected).Info("Cleaned up expired SAML sessions")
	return nil
}

// parseCertificate parses a PEM-encoded certificate
func parseCertificate(pemData string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// generateRequestID generates a unique SAML request ID
func generateRequestID() string {
	return "_" + uuid.Must(uuid.NewRandom()).String()
}

// formatNameID returns the full NameID format URL
func formatNameID(format string) string {
	switch format {
	case "emailAddress":
		return "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	case "unspecified":
		return "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
	case "transient":
		return "urn:oasis:names:tc:SAML:2.0:nameid-format:transient"
	case "persistent":
		return "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
	default:
		return "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	}
}

// GenerateMetadata generates SP metadata XML
func (s *SAMLService) GenerateMetadata(tenantID uuid.UUID, config *SAMLConfig) string {
	entityDescriptor := EntityDescriptor{
		XMLName: xml.Name{
			Local: "md:EntityDescriptor",
		},
		EntityID:   config.SPEntityID,
		ID:         generateRequestID(),
		ValidUntil: time.Now().Add(24 * time.Hour * 365).Format(time.RFC3339),
		SPSSODescriptor: SPSSODescriptor{
			ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
			AuthnRequestsSigned:        "false",
			WantAssertionsSigned:       "true",
			KeyDescriptors: []KeyDescriptor{
				{
					Use: "signing",
					KeyInfo: KeyInfo{
						X509Data: X509Data{
							X509Certificate: strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s.spConfig.CertificatePEM, "-----BEGIN CERTIFICATE-----", ""), "-----END CERTIFICATE-----", "")),
						},
					},
				},
			},
			AssertionConsumerServices: []IndexedEndpoint{
				{
					Binding:   "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
					Location:  config.ACSURL,
					Index:     "0",
					IsDefault: "true",
				},
			},
		},
	}

	xmlData, _ := xml.MarshalIndent(entityDescriptor, "", "  ")
	return xml.Header + string(xmlData)
}

// HashToken creates a bcrypt hash of a token
func HashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyToken verifies a token against a hash
func VerifyToken(token, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil
}

// XML Structures for SAML

// AuthnRequest represents a SAML authentication request
type AuthnRequest struct {
	XMLName                     xml.Name     `xml:"samlp:AuthnRequest"`
	ID                          string       `xml:"ID,attr"`
	Version                     string       `xml:"Version,attr"`
	IssueInstant                string       `xml:"IssueInstant,attr"`
	Destination                 string       `xml:"Destination,attr,omitempty"`
	ProtocolBinding             string       `xml:"ProtocolBinding,attr"`
	AssertionConsumerServiceURL string       `xml:"AssertionConsumerServiceURL,attr"`
	Issuer                      Issuer       `xml:"saml:Issuer"`
	NameIDPolicy                NameIDPolicy `xml:"samlp:NameIDPolicy"`
}

// Issuer represents the SAML Issuer
type Issuer struct {
	XMLName xml.Name `xml:"saml:Issuer"`
	Value   string   `xml:",chardata"`
}

// NameIDPolicy represents the NameID policy
type NameIDPolicy struct {
	XMLName     xml.Name `xml:"samlp:NameIDPolicy"`
	Format      string   `xml:"Format,attr"`
	AllowCreate string   `xml:"AllowCreate,attr,omitempty"`
}

// LogoutRequest represents a SAML 2.0 LogoutRequest (Single Logout)
type LogoutRequest struct {
	XMLName      xml.Name               `xml:"LogoutRequest"`
	ID           string                 `xml:"ID,attr"`
	Version      string                 `xml:"Version,attr"`
	IssueInstant string                 `xml:"IssueInstant,attr"`
	Issuer       Issuer                 `xml:"Issuer"`
	NameID       NameID                 `xml:"NameID"`
	SessionIndex []LogoutRequestSessionIndex `xml:"SessionIndex"`
}

// LogoutRequestSessionIndex is the SessionIndex element in a LogoutRequest
type LogoutRequestSessionIndex struct {
	Value string `xml:",chardata"`
}

// SAMLResponse represents a SAML response
type SAMLResponse struct {
	XMLName      xml.Name    `xml:"Response"`
	ID           string      `xml:"ID,attr"`
	InResponseTo string      `xml:"InResponseTo,attr"`
	Version      string      `xml:"Version,attr"`
	IssueInstant string      `xml:"IssueInstant,attr"`
	Destination  string      `xml:"Destination,attr,omitempty"`
	Issuer       Issuer      `xml:"Issuer"`
	Status       Status      `xml:"Status"`
	Assertions   []Assertion `xml:"Assertion"`
}

// Status represents the SAML status
type Status struct {
	StatusCode StatusCode `xml:"StatusCode"`
}

// StatusCode represents the SAML status code
type StatusCode struct {
	Value string `xml:"Value,attr"`
}

// Assertion represents a SAML assertion
type Assertion struct {
	XMLName            xml.Name           `xml:"Assertion"`
	ID                 string             `xml:"ID,attr"`
	Version            string             `xml:"Version,attr"`
	IssueInstant       string             `xml:"IssueInstant,attr"`
	Issuer             Issuer             `xml:"Issuer"`
	Subject            Subject            `xml:"Subject"`
	Conditions         Conditions         `xml:"Conditions"`
	AttributeStatement AttributeStatement `xml:"AttributeStatement"`
	AuthnStatement     AuthnStatement     `xml:"AuthnStatement"`
}

// Subject represents the SAML subject
type Subject struct {
	NameID NameID `xml:"NameID"`
}

// NameID represents the SAML NameID
type NameID struct {
	Format string `xml:"Format,attr,omitempty"`
	Value  string `xml:",chardata"`
}

// Conditions represents SAML conditions
type Conditions struct {
	NotBefore            string                `xml:"NotBefore,attr"`
	NotOnOrAfter         string                `xml:"NotOnOrAfter,attr"`
	AudienceRestrictions []AudienceRestriction `xml:"AudienceRestriction"`
}

// AudienceRestriction represents SAML audience restrictions
type AudienceRestriction struct {
	Audiences []Audience `xml:"Audience"`
}

// Audience represents a SAML audience
type Audience struct {
	Value string `xml:",chardata"`
}

// AttributeStatement represents SAML attributes
type AttributeStatement struct {
	Attributes []Attribute `xml:"Attribute"`
}

// Attribute represents a SAML attribute
type Attribute struct {
	Name            string           `xml:"Name,attr"`
	NameFormat      string           `xml:"NameFormat,attr,omitempty"`
	AttributeValues []AttributeValue `xml:"AttributeValue"`
}

// AttributeValue represents a SAML attribute value
type AttributeValue struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

// AuthnStatement represents SAML authentication statement
type AuthnStatement struct {
	AuthnInstant string       `xml:"AuthnInstant,attr"`
	SessionIndex string       `xml:"SessionIndex,attr"`
	AuthnContext AuthnContext `xml:"AuthnContext"`
}

// AuthnContext represents SAML authentication context
type AuthnContext struct {
	AuthnContextClassRef string `xml:"AuthnContextClassRef"`
}

// EntityDescriptor represents SAML metadata
type EntityDescriptor struct {
	XMLName         xml.Name        `xml:"md:EntityDescriptor"`
	EntityID        string          `xml:"entityID,attr"`
	ID              string          `xml:"ID,attr"`
	ValidUntil      string          `xml:"validUntil,attr"`
	SPSSODescriptor SPSSODescriptor `xml:"md:SPSSODescriptor"`
}

// SPSSODescriptor represents the SP SSO descriptor
type SPSSODescriptor struct {
	ProtocolSupportEnumeration string            `xml:"protocolSupportEnumeration,attr"`
	AuthnRequestsSigned        string            `xml:"AuthnRequestsSigned,attr"`
	WantAssertionsSigned       string            `xml:"WantAssertionsSigned,attr"`
	KeyDescriptors             []KeyDescriptor   `xml:"md:KeyDescriptor"`
	AssertionConsumerServices  []IndexedEndpoint `xml:"md:AssertionConsumerService"`
}

// KeyDescriptor represents a key descriptor
type KeyDescriptor struct {
	Use     string  `xml:"use,attr"`
	KeyInfo KeyInfo `xml:"ds:KeyInfo"`
}

// KeyInfo represents key info
type KeyInfo struct {
	X509Data X509Data `xml:"ds:X509Data"`
}

// X509Data represents X509 data
type X509Data struct {
	X509Certificate string `xml:"ds:X509Certificate"`
}

// IndexedEndpoint represents an indexed endpoint
type IndexedEndpoint struct {
	Binding   string `xml:"Binding,attr"`
	Location  string `xml:"Location,attr"`
	Index     string `xml:"index,attr"`
	IsDefault string `xml:"isDefault,attr,omitempty"`
}

// HashString creates a SHA256 hash of a string
func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SecureRandomString generates a cryptographically secure random string
func SecureRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// GetUserByEmail retrieves a user by email from the database
func (s *SAMLService) GetUserByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*UserInfo, error) {
	var user UserInfo
	result := s.db.WithContext(ctx).Table("gba_users").
		Select("id, email, COALESCE(first_name, '') as first_name, COALESCE(last_name, '') as last_name, tenant_id, created_at").
		Where("tenant_id = ? AND email = ?", tenantID, email).
		First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil // User doesn't exist
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}

	return &user, nil
}

// CreateUser creates a new user from SAML assertion
func (s *SAMLService) CreateUser(ctx context.Context, tenantID uuid.UUID, assertion *SAMLAssertion) (*UserInfo, error) {
	userID := uuid.Must(uuid.NewRandom())
	email := assertion.GetEmail(nil)
	firstName := assertion.GetFirstName(nil)
	lastName := assertion.GetLastName(nil)

	user := &UserInfo{
		ID:        userID,
		TenantID:  tenantID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now(),
	}

	// Insert into database
	result := s.db.WithContext(ctx).Table("gba_users").Create(map[string]interface{}{
		"id":         userID,
		"tenant_id":  tenantID,
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
		"email":     email,
	}).Info("User created via SAML")

	return user, nil
}

// UserInfo represents basic user information
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateUser updates user information from SAML assertion
func (s *SAMLService) UpdateUser(ctx context.Context, user *UserInfo, assertion *SAMLAssertion) error {
	firstName := assertion.GetFirstName(nil)
	lastName := assertion.GetLastName(nil)

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if firstName != "" {
		updates["first_name"] = firstName
	}
	if lastName != "" {
		updates["last_name"] = lastName
	}

	result := s.db.WithContext(ctx).Table("gba_users").
		Where("id = ?", user.ID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
	}).Info("User updated via SAML")

	return nil
}
