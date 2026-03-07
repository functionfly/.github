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

// BeginAuthentication starts the WebAuthn authentication ceremony
// This generates the options that the frontend needs to assert a credential
func (p *WebAuthnPlugin) BeginAuthentication(ctx context.Context, userID uuid.UUID) (*BeginAuthenticationResponse, error) {
	// Get user
	user, err := p.getUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load user's credentials
	credentials, err := p.loadUserCredentials(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("no credentials found for user")
	}

	user.Credentials = credentials

	// Configure authentication options
	opts := make([]webauthn.LoginOption, 0)

	// Begin authentication
	assertion, sessionData, err := p.webauthn.BeginLogin(user, opts...)
	if err != nil {
		p.logger.WithError(err).Error("Failed to begin WebAuthn authentication")
		return nil, fmt.Errorf("failed to begin authentication: %w", err)
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
		Operation: "authentication",
		ExpiresAt: time.Now().Add(p.config.SessionTimeout),
	}

	if err := p.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": session.ID,
	}).Info("WebAuthn authentication started")

	return &BeginAuthenticationResponse{
		SessionID: session.ID.String(),
		Options:   assertion,
	}, nil
}

// BeginAuthenticationDiscoverable starts a discoverable credential authentication (without user ID)
// This is used for passkeys where the user is not known in advance
func (p *WebAuthnPlugin) BeginAuthenticationDiscoverable(ctx context.Context) (*BeginAuthenticationResponse, error) {
	// Configure authentication options for discoverable credentials
	opts := make([]webauthn.LoginOption, 0)

	// Allow discoverable credentials (resident keys)
	opts = append(opts, webauthn.WithUserVerification(protocol.VerificationPreferred))

	// Begin authentication without a specific user
	assertion, sessionData, err := p.webauthn.BeginDiscoverableLogin(opts...)
	if err != nil {
		p.logger.WithError(err).Error("Failed to begin WebAuthn discoverable authentication")
		return nil, fmt.Errorf("failed to begin authentication: %w", err)
	}

	// Serialize session data to JSON for storage
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Create session in database with a temporary user ID
	// The actual user ID will be determined after authentication
	session := &WebAuthnSession{
		UserID:    uuid.Nil, // Will be updated after authentication
		Challenge: base64URLEncode(sessionDataJSON),
		Operation: "authentication",
		ExpiresAt: time.Now().Add(p.config.SessionTimeout),
	}

	if err := p.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
	}).Info("WebAuthn discoverable authentication started")

	return &BeginAuthenticationResponse{
		SessionID: session.ID.String(),
		Options:   assertion,
	}, nil
}

// FinishAuthentication completes the WebAuthn authentication ceremony
// This verifies the credential assertion response and returns the authenticated user
func (p *WebAuthnPlugin) FinishAuthentication(ctx context.Context, req FinishAuthenticationRequest) (*FinishAuthenticationResponse, error) {
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
	if session.Operation != "authentication" {
		return nil, fmt.Errorf("invalid session operation")
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

	// Parse the assertion response
	parsedResponse, err := req.Response.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse assertion response: %w", err)
	}

	// Get the credential ID from the response
	credentialID := []byte(parsedResponse.Raw.ID)

	// Find the credential in the database
	cred, err := p.getCredentialByCredentialID(ctx, credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	// Get user
	user, err := p.getUserByID(ctx, cred.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load user's credentials
	credentials, err := p.loadUserCredentials(ctx, cred.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	user.Credentials = credentials

	// Finish authentication
	credential, err := p.webauthn.ValidateLogin(user, sessionData, parsedResponse)
	if err != nil {
		p.logger.WithError(err).Warn("WebAuthn authentication verification failed")
		return nil, fmt.Errorf("authentication verification failed: %w", err)
	}

	// Update sign count to prevent replay attacks
	if credential.Authenticator.SignCount > 0 {
		if err := p.updateCredentialSignCount(ctx, credential.ID, credential.Authenticator.SignCount); err != nil {
			p.logger.WithError(err).Warn("Failed to update credential sign count")
		}
	}

	// Delete session
	if err := p.deleteSession(ctx, sessionID); err != nil {
		p.logger.WithError(err).Warn("Failed to delete session after authentication")
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":       cred.UserID,
		"credential_id": cred.ID,
	}).Info("WebAuthn authentication completed")

	return &FinishAuthenticationResponse{
		UserID:  cred.UserID,
		Email:   user.Email,
		Message: "Authentication successful",
	}, nil
}

// FinishAuthenticationDiscoverable completes a discoverable credential authentication
// This is used for passkeys where the user is not known in advance
func (p *WebAuthnPlugin) FinishAuthenticationDiscoverable(ctx context.Context, req FinishAuthenticationRequest) (*FinishAuthenticationResponse, error) {
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
	if session.Operation != "authentication" {
		return nil, fmt.Errorf("invalid session operation")
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

	// Parse the assertion response
	parsedResponse, err := req.Response.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse assertion response: %w", err)
	}

	// Get the credential ID from the response
	credentialID := []byte(parsedResponse.Raw.ID)

	// Find the credential in the database
	cred, err := p.getCredentialByCredentialID(ctx, credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	// Get user
	user, err := p.getUserByID(ctx, cred.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load user's credentials
	credentials, err := p.loadUserCredentials(ctx, cred.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	user.Credentials = credentials

	// Validate discoverable login
	_, err = p.webauthn.ValidateDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			return user, nil
		},
		sessionData,
		parsedResponse,
	)
	if err != nil {
		p.logger.WithError(err).Warn("WebAuthn discoverable authentication verification failed")
		return nil, fmt.Errorf("authentication verification failed: %w", err)
	}

	// Get the credential to update sign count
	credential, err := p.getCredentialByCredentialID(ctx, credentialID)
	if err != nil {
		p.logger.WithError(err).Warn("Failed to get credential for sign count update")
	} else {
		// Update sign count to prevent replay attacks
		if err := p.updateCredentialSignCount(ctx, credential.CredentialID, credential.SignCount+1); err != nil {
			p.logger.WithError(err).Warn("Failed to update credential sign count")
		}
	}

	// Delete session
	if err := p.deleteSession(ctx, sessionID); err != nil {
		p.logger.WithError(err).Warn("Failed to delete session after authentication")
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":       cred.UserID,
		"credential_id": cred.ID,
	}).Info("WebAuthn discoverable authentication completed")

	return &FinishAuthenticationResponse{
		UserID:  cred.UserID,
		Email:   user.Email,
		Message: "Authentication successful",
	}, nil
}
