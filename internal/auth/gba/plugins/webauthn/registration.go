// Package webauthn provides WebAuthn/Passkeys authentication support for GoBetterAuth
package webauthn

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// BeginRegistration starts the WebAuthn registration ceremony
// This generates the options that the frontend needs to create a new credential
func (p *WebAuthnPlugin) BeginRegistration(ctx context.Context, userID uuid.UUID, req BeginRegistrationRequest) (*BeginRegistrationResponse, error) {
	// Validate request
	if err := p.ValidateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Get user from database
	user, err := p.getUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load existing credentials for the user
	credentials, err := p.loadUserCredentials(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	user.Credentials = credentials

	// Configure registration options
	opts := make([]webauthn.RegistrationOption, 0)

	// Set authenticator selection based on request
	selection := protocol.AuthenticatorSelection{
		UserVerification: protocol.VerificationPreferred,
	}

	if req.AuthenticatorType == "platform" {
		selection.AuthenticatorAttachment = protocol.Platform
	} else if req.AuthenticatorType == "cross-platform" {
		selection.AuthenticatorAttachment = protocol.CrossPlatform
	}

	if req.ResidentKey {
		selection.RequireResidentKey = &req.ResidentKey
	}

	opts = append(opts, webauthn.WithAuthenticatorSelection(selection))

	// Exclude existing credentials to prevent duplicate registration
	if len(credentials) > 0 {
		excludeList := make([]protocol.CredentialDescriptor, len(credentials))
		for i, cred := range credentials {
			excludeList[i] = *CredentialToDescriptor(&cred)
		}
		opts = append(opts, webauthn.WithExclusions(excludeList))
	}

	// Begin registration
	creation, sessionData, err := p.webauthn.BeginRegistration(user, opts...)
	if err != nil {
		p.logger.WithError(err).Error("Failed to begin WebAuthn registration")
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Serialize session data to JSON for storage
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Create session in database
	session := &WebAuthnSession{
		UserID:    userID,
		Challenge: base64URLEncode(sessionDataJSON),
		Operation: "registration",
		ExpiresAt: time.Now().Add(p.config.SessionTimeout),
	}

	if err := p.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": session.ID,
		"name":       req.Name,
	}).Info("WebAuthn registration started")

	return &BeginRegistrationResponse{
		SessionID: session.ID.String(),
		Options:   creation,
	}, nil
}

// FinishRegistration completes the WebAuthn registration ceremony
// This verifies the credential creation response and stores the credential
func (p *WebAuthnPlugin) FinishRegistration(ctx context.Context, userID uuid.UUID, req FinishRegistrationRequest) (*FinishRegistrationResponse, error) {
	// Parse session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Get session from database
	session, err := p.getSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify session operation
	if session.Operation != "registration" {
		return nil, fmt.Errorf("invalid session operation")
	}

	// Verify session belongs to user
	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Decode session data
	sessionDataJSON, err := base64URLDecode(session.Challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session data: %w", err)
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataJSON, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Get user
	user, err := p.getUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load existing credentials
	credentials, err := p.loadUserCredentials(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	user.Credentials = credentials

	// Parse the credential response from the request
	parsedResponse, err := req.Response.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential response: %w", err)
	}

	// Finish registration
	credential, err := p.webauthn.CreateCredential(user, sessionData, parsedResponse)
	if err != nil {
		p.logger.WithError(err).Warn("WebAuthn registration verification failed")
		return nil, fmt.Errorf("registration verification failed: %w", err)
	}

	// Use default device name
	deviceName := "Passkey"

	// Store credential in database
	// Convert transport to string slice
	transport := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transport[i] = string(t)
	}

	// Calculate flags
	flags := 0
	if credential.Flags.UserPresent {
		flags |= 1
	}
	if credential.Flags.UserVerified {
		flags |= 2
	}
	if credential.Flags.BackupEligible {
		flags |= 4
	}
	if credential.Flags.BackupState {
		flags |= 8
	}

	// Store authenticator data as JSON
	authenticatorData := map[string]interface{}{
		"aaguid":    base64URLEncode(credential.Authenticator.AAGUID),
		"signCount": credential.Authenticator.SignCount,
	}
	authenticatorJSON, err := json.Marshal(authenticatorData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authenticator data: %w", err)
	}

	newCredential := &WebAuthnCredential{
		UserID:          userID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       transport,
		Flags:           flags,
		Authenticator:   authenticatorJSON,
		SignCount:       credential.Authenticator.SignCount,
		Name:            deviceName,
	}

	if err := p.storeCredential(ctx, newCredential); err != nil {
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	// Delete session
	if err := p.deleteSession(ctx, sessionID); err != nil {
		p.logger.WithError(err).Warn("Failed to delete session after registration")
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"credential_id": newCredential.ID,
	}).Info("WebAuthn registration completed")

	return &FinishRegistrationResponse{
		CredentialID: newCredential.ID.String(),
		Name:         newCredential.Name,
		CreatedAt:    newCredential.CreatedAt,
		Message:      "Passkey registered successfully",
	}, nil
}

// RegistrationOptions provides options for customizing the registration ceremony
type RegistrationOptions struct {
	// RequireResidentKey creates a discoverable credential (passkey)
	RequireResidentKey bool

	// UserVerification determines if user verification is required
	UserVerification string // "required", "preferred", "discouraged"

	// AuthenticatorAttachment specifies the authenticator type
	AuthenticatorAttachment string // "platform", "cross-platform", ""

	// ExcludeCredentials prevents duplicate registration
	ExcludeCredentials []protocol.CredentialDescriptor

	// Attestation determines the attestation conveyance preference
	Attestation string // "none", "indirect", "direct"
}

// ValidateRegistrationRequest validates the registration request
func (p *WebAuthnPlugin) ValidateRegistrationRequest(req BeginRegistrationRequest) error {
	if req.AuthenticatorType != "" &&
		req.AuthenticatorType != "platform" &&
		req.AuthenticatorType != "cross-platform" {
		return fmt.Errorf("invalid authenticator_type: must be 'platform', 'cross-platform', or empty")
	}

	// Device name validation
	if len(req.Name) > 255 {
		return fmt.Errorf("name too long: maximum 255 characters")
	}

	return nil
}
