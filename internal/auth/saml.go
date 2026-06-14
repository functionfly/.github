package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"encoding/hex"
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
	PrivateKey  string // PEM-encoded RSA private key (optional; generated if empty)
	Certificate string // PEM-encoded X.509 certificate (optional; generated if empty)
}

// NewSAMLService creates a new SAML service
// SP key pair is loaded from: 1) environment config, 2) database, or 3) generated and persisted
func NewSAMLService(config SAMLServiceConfig) (*SAMLService, error) {
	var privateKey *rsa.PrivateKey
	var cert *x509.Certificate
	var keySource string

	// 1. Try to load from environment config (highest priority)
	if config.PrivateKey != "" && config.Certificate != "" {
		block, _ := pem.Decode([]byte(config.PrivateKey))
		if block != nil {
			if pk, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				privateKey = pk
				keySource = "environment"
				logrus.Info("Loaded SAML SP key pair from environment configuration")
			}
		}
		if privateKey != nil {
			certBlock, _ := pem.Decode([]byte(config.Certificate))
			if certBlock != nil {
				if c, err := x509.ParseCertificate(certBlock.Bytes); err == nil {
					cert = c
				}
			}
		}
	}

	// 2. Try to load from database (if not found in environment)
	if privateKey == nil && config.ConfigRepo != nil {
		// Use a nil/empty tenant ID to look for platform-wide keys
		// Try to get any config that has SP keys stored
		dbPrivateKey, dbCertificate, err := config.ConfigRepo.GetSPKeyPair(context.Background(), uuid.Nil)
		if err != nil {
			logrus.WithError(err).Warn("Failed to load SAML SP key pair from database")
		}
		if dbPrivateKey != nil && dbCertificate != nil {
			block, _ := pem.Decode([]byte(*dbPrivateKey))
			if block != nil {
				if pk, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
					privateKey = pk
					keySource = "database"
					logrus.Info("Loaded SAML SP key pair from database")
				}
			}
			if privateKey != nil {
				certBlock, _ := pem.Decode([]byte(*dbCertificate))
				if certBlock != nil {
					if c, err := x509.ParseCertificate(certBlock.Bytes); err == nil {
						cert = c
					}
				}
			}
		}
	}

	// 3. Generate a new key pair if none was loaded
	if privateKey == nil {
		logrus.Info("Generating new SAML SP key pair (not found in environment or database)")
		var err error
		privateKey, err = rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		// Create a self-signed certificate
		template := x509.Certificate{
			SerialNumber: big.NewInt(time.Now().UnixNano()),
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

		cert, err = x509.ParseCertificate(certDER)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		keySource = "generated"

		// Persist the generated keys to database for future restarts
		if config.ConfigRepo != nil {
			privateKeyPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			})
			certPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: certDER,
			})

			if err := config.ConfigRepo.SaveSPKeyPair(context.Background(), uuid.Nil, string(privateKeyPEM), string(certPEM)); err != nil {
				logrus.WithError(err).Warn("Failed to persist generated SAML SP key pair to database")
			} else {
				logrus.Info("Persisted generated SAML SP key pair to database")
			}
		}
	}

	logrus.WithField("key_source", keySource).Info("SAML service initialized with SP key pair")

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
	config, err := s.configRepo.GetByTenantID(context.Background(), tenantID)
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
	config, err := s.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get SAML config: %w", err)
	}

	if !config.Enabled {
		return "", fmt.Errorf("SAML is not enabled for this tenant")
	}

	// Generate AuthnRequest ID (must be unique and non-guessable)
	requestID := "_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	// Create cryptographically secure state token (replaces UUID-based state)
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate secure state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	expiresAt := time.Now().Add(10 * time.Minute)
	// Store state with requestID for InResponseTo validation on callback
	if err := s.stateRepo.SaveAuthnRequestState(context.Background(), state, tenantID, requestID, relayState, expiresAt); err != nil {
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
	XMLName       xml.Name        `xml:"samlp:Response"`
	ID            string          `xml:"ID,attr"`
	InResponseTo  string          `xml:"InResponseTo,attr"`
	Signature     *SAMLSignature  `xml:"ds:Signature,omitempty"`
	Assertion     *SAMLAssertion   `xml:"saml:Assertion"`
}

// SAMLSignature represents a digital signature (XML DSig)
type SAMLSignature struct {
	XMLName        xml.Name   `xml:"ds:Signature"`
	SignedInfo     *DSSignedInfo `xml:"SignedInfo"`
	KeyInfo        *DSKeyInfo `xml:"KeyInfo"`
	SignatureValue string     `xml:"SignatureValue"`
}

// DSSignedInfo contains the SignedInfo element for signature verification
// Uses proper XML namespace handling to extract canonical XML
type DSSignedInfo struct {
	XMLName       xml.Name    `xml:"ds:SignedInfo"`
	CanonicalMethod string    `xml:"CanonicalizationMethod"`
	SignatureMethod string    `xml:"SignatureMethod"`
	Reference     DSReference `xml:"Reference"`
}

// DSReference contains reference information for the signature
type DSReference struct {
	URI          string `xml:"URI,attr"`
	DigestMethod string `xml:"DigestMethod"`
	DigestValue  string `xml:"DigestValue"`
}

// DSKeyInfo contains key information for signature verification
type DSKeyInfo struct {
	XMLName  xml.Name  `xml:"KeyInfo"`
	X509Data *X509Data `xml:"X509Data"`
}

// X509Data contains X509 certificate data
type X509Data struct {
	XMLCertificate string `xml:"X509Certificate"`
}

// SAMLAssertion represents a SAML assertion
type SAMLAssertion struct {
	XMLName             xml.Name                 `xml:"saml:Assertion"`
	ID                  string                   `xml:"ID,attr"`
	Subject             *SAMLSubject             `xml:"saml:Subject"`
	Conditions          *SAMLConditions          `xml:"saml:Conditions"`
	AttributeStatements []SAMLAttributeStatement `xml:"saml:AttributeStatement"`
}

// SAMLConditions represents the conditions under which the assertion is valid
type SAMLConditions struct {
	NotBefore    time.Time `xml:"NotBefore,attr"`
	NotOnOrAfter time.Time `xml:"NotOnOrAfter,attr"`
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

// verifySAMLSignature verifies the SAML response signature using proper XML parsing
// Returns nil if signature is valid, error if verification fails or signature is missing
func (s *SAMLService) verifySAMLSignature(decodedResponse []byte, config *storage.SAMLConfig) error {
	// Check if IdP certificate is configured
	if config.IDPCertificate == nil || *config.IDPCertificate == "" {
		return fmt.Errorf("IdP certificate not configured - cannot verify SAML signature")
	}

	// Parse the XML to extract signature info using proper XML structures
	var sigInfo struct {
		Signature struct {
			SignatureValue string `xml:"SignatureValue"`
			SignedInfo     struct {
				CanonicalMethod string `xml:"CanonicalizationMethod,attr"`
			} `xml:"SignedInfo"`
			KeyInfo struct {
				X509Data struct {
					X509Certificate string `xml:"X509Certificate"`
				} `xml:"X509Data"`
			} `xml:"KeyInfo"`
		} `xml:"Signature"`
	}

	if err := xml.Unmarshal(decodedResponse, &sigInfo); err != nil {
		return fmt.Errorf("failed to parse SAML signature: %w", err)
	}

	// Extract the signature value
	if sigInfo.Signature.SignatureValue == "" {
		return fmt.Errorf("SAML response has no signature - possible forged response")
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(sigInfo.Signature.SignatureValue)
	if err != nil {
		return fmt.Errorf("failed to decode signature value: %w", err)
	}

	// Get the certificate for verification
	var certPEM string
	if sigInfo.Signature.KeyInfo.X509Data.X509Certificate != "" {
		// Use embedded certificate from response
		certPEM = sigInfo.Signature.KeyInfo.X509Data.X509Certificate
	} else {
		// Fall back to configured IdP certificate
		certPEM = *config.IDPCertificate
	}

	// Parse the certificate
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		// Try adding PEM headers if missing
		certBlock, _ = pem.Decode([]byte("-----BEGIN CERTIFICATE-----\n" + certPEM + "\n-----END CERTIFICATE-----"))
		if certBlock == nil {
			return fmt.Errorf("failed to parse IdP certificate")
		}
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse IdP certificate: %w", err)
	}

	// Extract SignedInfo content using proper XML parsing
	// Use XML token approach to get exact SignedInfo element content
	signedInfoXML, err := extractSignedInfoXML(decodedResponse)
	if err != nil {
		return fmt.Errorf("could not find SignedInfo element in SAML response: %w", err)
	}

	// Hash the SignedInfo with SHA-256 (SHA-1 is cryptographically broken)
	hashed := sha256.Sum256(signedInfoXML)

	// Verify the signature using RSA PKCS1v15
	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("certificate does not contain RSA public key")
	}

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		logrus.WithError(err).Warn("SAML signature verification failed - response may be forged")
		return fmt.Errorf("SAML signature verification failed: %w", err)
	}

	logrus.Debug("SAML signature verified successfully")
	return nil
}

// extractSignedInfoXML extracts the SignedInfo element content using proper XML parsing
// This avoids regex-based extraction which can be vulnerable to malformed XML
func extractSignedInfoXML(xmlData []byte) ([]byte, error) {
	// Use xml.NewDecoder for streaming parse to find SignedInfo
	decoder := xml.NewDecoder(strings.NewReader(string(xmlData)))

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to parse XML: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			// Check for ds:SignedInfo or SignedInfo
			if se.Name.Local == "SignedInfo" || se.Name.Local == "SignedInfo" &&
				(strings.HasSuffix(se.Name.Space, "signature") || se.Name.Space == "") {
				// Found SignedInfo, extract the element
				// var signedInfo xml.Token // removed as unused
				// Get the start tag raw content
				startBytes := decoder.InputOffset()
				// Read until end tag
				// var content []byte // removed as unused
				depth := 1
				for depth > 0 {
					tok, err := decoder.Token()
					if err != nil {
						return nil, fmt.Errorf("failed to read SignedInfo: %w", err)
					}
					switch tok.(type) {
					case xml.StartElement:
						depth++
					case xml.EndElement:
						depth--
						if depth == 0 && tok.(xml.EndElement).Name.Local == "SignedInfo" {
							endBytes := decoder.InputOffset()
							return xmlData[startBytes-1:endBytes], nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("SignedInfo element not found")
}

// validateAssertionTimestamps validates NotBefore and NotOnOrAfter conditions in SAML assertion
func (s *SAMLService) validateAssertionTimestamps(assertion *SAMLAssertion) error {
	if assertion == nil || assertion.Conditions == nil {
		return nil // No conditions to validate
	}

	now := time.Now()

	// Check NotBefore - assertion is not valid yet
	if !assertion.Conditions.NotBefore.IsZero() && now.Before(assertion.Conditions.NotBefore) {
		return fmt.Errorf("assertion not yet valid: NotBefore=%v, now=%v", assertion.Conditions.NotBefore, now)
	}

	// Check NotOnOrAfter - assertion has expired
	if !assertion.Conditions.NotOnOrAfter.IsZero() && now.After(assertion.Conditions.NotOnOrAfter) {
		return fmt.Errorf("assertion has expired: NotOnOrAfter=%v, now=%v", assertion.Conditions.NotOnOrAfter, now)
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ProcessResponse processes a SAML Response from the IdP
// relayState contains our original state token (used to look up the AuthnRequest ID for InResponseTo validation)
func (s *SAMLService) ProcessResponse(tenantID uuid.UUID, samlResponse string, relayState string) (*SAMLLoginResponse, error) {
	config, err := s.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SAML config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("SAML is not enabled for this tenant")
	}

	// Validate relayState and retrieve stored AuthnRequest ID
	_, requestID, _, err := s.stateRepo.GetAuthnRequestState(context.Background(), relayState)
	if err != nil {
		logrus.WithError(err).WithField("relay_state", relayState[:min(8, len(relayState))]+"...").Warn("SAML state validation failed")
		return nil, fmt.Errorf("SAML state validation failed: %w", err)
	}

	// Decode the response
	decodedResponse, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// CRITICAL: Verify SAML response signature before trusting any data
	if err := s.verifySAMLSignature(decodedResponse, config); err != nil {
		logrus.WithError(err).Error("SAML signature verification failed - rejecting response")
		return nil, fmt.Errorf("SAML response signature verification failed: %w", err)
	}

	// Parse the SAML response
	var resp SAMLResponse
	if err := xml.Unmarshal(decodedResponse, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	if resp.Assertion == nil {
		return nil, fmt.Errorf("no assertion in SAML response")
	}

	// SECURITY: Validate InResponseTo matches our original AuthnRequest ID (replay attack prevention)
	if resp.InResponseTo != "" && requestID != "" && resp.InResponseTo != requestID {
		logrus.WithFields(logrus.Fields{
			"in_response_to": resp.InResponseTo,
			"expected_id":    requestID,
		}).Warn("SAML InResponseTo mismatch - possible replay attack")
		return nil, fmt.Errorf("SAML response InResponseTo does not match (replay attack prevention)")
	}

	// SECURITY: Validate assertion timestamps (NotBefore and NotOnOrAfter)
	if err := s.validateAssertionTimestamps(resp.Assertion); err != nil {
		logrus.WithError(err).Warn("SAML assertion timestamp validation failed")
		return nil, fmt.Errorf("SAML assertion timestamp validation failed: %w", err)
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

	// Delete the state after successful validation (one-time use)
	_ = s.stateRepo.DeleteAuthnRequestState(context.Background(), relayState)

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

	// Determine NotOnOrAfter from assertion conditions
	notOnOrAfter := time.Now().Add(24 * time.Hour)
	if resp.Assertion.Conditions != nil && !resp.Assertion.Conditions.NotOnOrAfter.IsZero() {
		notOnOrAfter = resp.Assertion.Conditions.NotOnOrAfter
	}

	session := &storage.SAMLSession{
		ID:           uuid.New(),
		TenantID:     tenantID,
		UserID:       user.ID,
		SAMLNameID:   nameID,
		SessionIndex: sessionIndex,
		NotOnOrAfter: notOnOrAfter,
		Attributes:   attrs,
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(context.Background(), session); err != nil {
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
	user, err := s.repo.GetUserByEmail(context.Background(), email)
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
	user, err = s.repo.CreateUserWithSocialAuth(context.Background(), email, tenantID, "saml", nameID, providerData)
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
	return s.configRepo.GetByTenantID(context.Background(), tenantID)
}

// SaveConfig saves SAML configuration for a tenant
func (s *SAMLService) SaveConfig(ctx context.Context, config *storage.SAMLConfig) error {
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
	existing, err := s.configRepo.GetByTenantID(ctx, config.TenantID)
	if err != nil {
		// Config doesn't exist, create new
		config.ID = uuid.New()
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		return s.configRepo.Create(context.Background(), config)
	}

	// Update existing config
	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()
	return s.configRepo.Update(ctx, config)
}

// HandleSLO handles Single Logout
func (s *SAMLService) HandleSLO(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Delete SAML sessions for the user
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// SAMLLoginResponse represents the response after successful SAML login
type SAMLLoginResponse struct {
	User        *storage.User
	Token       string
	NameID      string
	SAMLSession *storage.SAMLSession
}

// Repo returns the repository for auth event logging
func (s *SAMLService) Repo() storage.Repository {
	return s.repo
}

// AuthService returns the auth service for token generation
func (s *SAMLService) AuthService() *AuthService {
	return s.authService
}

// GenerateRefreshToken generates a refresh token for SAML login
func (s *SAMLService) GenerateRefreshToken() (token, hash string, err error) {
	if s.authService == nil {
		return "", "", fmt.Errorf("auth service not configured")
	}
	return s.authService.GenerateRefreshToken()
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
