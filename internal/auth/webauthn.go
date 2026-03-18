package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WebAuthnService handles WebAuthn/Passkey authentication operations
type WebAuthnService struct {
	webAuthn *webauthn.WebAuthn
	repo     *storage.WebAuthnRepository
	logger   *logrus.Logger
}

// Credential represents a WebAuthn credential for the application
type Credential struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	CredentialID   []byte
	PublicKey      []byte
	SignCount      uint32
	BackupEligible bool
	BackupState    bool
	CreatedAt      time.Time
	LastUsedAt     *time.Time
}

// WebAuthnConfig holds the configuration for WebAuthn
type WebAuthnConfig struct {
	RPID          string
	RPOrigins     []string
	RPDisplayName string
	Timeout       time.Duration
}

// webauthnUser implements the webauthn.User interface
type webauthnUser struct {
	id          []byte
	name        string
	displayName string
	credentials []*webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte          { return u.id }
func (u *webauthnUser) WebAuthnName() string        { return u.name }
func (u *webauthnUser) WebAuthnDisplayName() string { return u.displayName }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	result := make([]webauthn.Credential, len(u.credentials))
	for i, c := range u.credentials {
		result[i] = *c
	}
	return result
}
func (u *webauthnUser) WebAuthnIcon() string { return "" }

// NewWebAuthnService creates a new WebAuthn service
func NewWebAuthnService(repo *storage.WebAuthnRepository, config WebAuthnConfig) (*WebAuthnService, error) {
	logger := logrus.New()

	webAuthnConfig := &webauthn.Config{
		RPID:          config.RPID,
		RPOrigins:     config.RPOrigins,
		RPDisplayName: config.RPDisplayName,
		Timeouts: webauthn.TimeoutsConfig{
			Registration: webauthn.TimeoutConfig{Timeout: config.Timeout},
			Login:        webauthn.TimeoutConfig{Timeout: config.Timeout},
		},
	}

	wa, err := webauthn.New(webAuthnConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	return &WebAuthnService{
		webAuthn: wa,
		repo:     repo,
		logger:   logger,
	}, nil
}

// BeginRegistration starts the registration ceremony for a new passkey
func (s *WebAuthnService) BeginRegistration(userID uuid.UUID, displayName, email string) ([]byte, []byte, error) {
	// Get existing credentials for the user
	existingCreds, _ := s.repo.GetAllCredentialsForUser(userID)

	// Create WebAuthn user
	user := &webauthnUser{
		id:          userID[:],
		name:        email,
		displayName: displayName,
		credentials: s.convertCredentials(existingCreds),
	}

	// Generate registration options
	credentialOptions, sessionData, err := s.webAuthn.BeginRegistration(user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to begin registration")
		return nil, nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Marshal the options for the frontend
	optionsJSON, err := json.Marshal(credentialOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	// Marshal session data for the caller to store (e.g. API handler stores in Redis via WebAuthnSessionStore)
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"userID": userID,
		"email":  email,
	}).Info("Started WebAuthn registration")

	return optionsJSON, sessionJSON, nil
}

// FinishRegistration completes the registration ceremony
// responseData should be the JSON-encoded attestation response from the client
func (s *WebAuthnService) FinishRegistration(userID uuid.UUID, sessionDataJSON []byte, responseData json.RawMessage) (*Credential, error) {
	// Parse the session data
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataJSON, &sessionData); err != nil {
		s.logger.WithError(err).Error("Failed to parse session data")
		return nil, fmt.Errorf("failed to parse session data: %w", err)
	}

	// Create user for verification
	user := &webauthnUser{
		id: userID[:],
	}

	// Create a fake http.Request with the response data
	req, err := http.NewRequest("POST", "/", strings.NewReader(string(responseData)))
	if err != nil {
		s.logger.WithError(err).Error("Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Validate the registration response
	credential, err := s.webAuthn.FinishRegistration(user, sessionData, req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to finish registration")
		return nil, fmt.Errorf("failed to finish registration: %w", err)
	}

	// Extract backup flags from credential flags
	backupEligible := credential.Flags.BackupEligible
	backupState := credential.Flags.BackupState

	// Store the credential in the database
	dbCredential := &storage.WebAuthnCredential{
		UserID:         userID,
		CredentialID:   credential.ID,
		PublicKey:      credential.PublicKey,
		SignCount:      credential.Authenticator.SignCount,
		BackupEligible: backupEligible,
		BackupState:    backupState,
	}

	if err := s.repo.Create(dbCredential); err != nil {
		s.logger.WithError(err).Error("Failed to store credential")
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"userID":       userID,
		"credentialID": base64.StdEncoding.EncodeToString(credential.ID),
	}).Info("Completed WebAuthn registration")

	return &Credential{
		ID:             dbCredential.ID,
		UserID:         dbCredential.UserID,
		CredentialID:   dbCredential.CredentialID,
		PublicKey:      dbCredential.PublicKey,
		SignCount:      dbCredential.SignCount,
		BackupEligible: dbCredential.BackupEligible,
		BackupState:    dbCredential.BackupState,
		CreatedAt:      dbCredential.CreatedAt,
		LastUsedAt:     dbCredential.LastUsedAt,
	}, nil
}

// BeginLogin starts the authentication ceremony
func (s *WebAuthnService) BeginLogin(userID uuid.UUID) ([]byte, []byte, error) {
	// Get credentials for the user
	credentials, err := s.repo.GetAllCredentialsForUser(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credentials for user")
		return nil, nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	if len(credentials) == 0 {
		s.logger.WithField("userID", userID).Warn("No credentials found for user")
		return nil, nil, errors.New("no credentials found for user")
	}

	// Create WebAuthn user
	user := &webauthnUser{
		id:          userID[:],
		credentials: s.convertCredentials(credentials),
	}

	// Generate authentication options
	credentialOptions, sessionData, err := s.webAuthn.BeginLogin(user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to begin login")
		return nil, nil, fmt.Errorf("failed to begin login: %w", err)
	}

	// Return options as JSON
	optionsJSON, err := json.Marshal(credentialOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	// Marshal session data for the caller to store (e.g. API handler stores in Redis via WebAuthnSessionStore)
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"userID":      userID,
		"credentials": len(credentials),
	}).Info("Started WebAuthn login")

	return optionsJSON, sessionJSON, nil
}

// FinishLogin completes the authentication ceremony
// responseData should be the JSON-encoded assertion response from the client
func (s *WebAuthnService) FinishLogin(userID uuid.UUID, sessionDataJSON []byte, responseData json.RawMessage) (bool, error) {
	// Get credentials for the user
	credentials, err := s.repo.GetAllCredentialsForUser(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credentials for user")
		return false, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Create WebAuthn user
	user := &webauthnUser{
		id:          userID[:],
		credentials: s.convertCredentials(credentials),
	}

	// Parse the session data
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataJSON, &sessionData); err != nil {
		s.logger.WithError(err).Error("Failed to parse session data")
		return false, fmt.Errorf("failed to parse session data: %w", err)
	}

	// Create a fake http.Request with the response data
	req, err := http.NewRequest("POST", "/", strings.NewReader(string(responseData)))
	if err != nil {
		s.logger.WithError(err).Error("Failed to create request")
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Verify the authentication response
	credential, err := s.webAuthn.FinishLogin(user, sessionData, req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to finish login")
		return false, fmt.Errorf("failed to finish login: %w", err)
	}

	// Find and update the credential in the database
	credID := base64.StdEncoding.EncodeToString(credential.ID)
	for _, cred := range credentials {
		if base64.StdEncoding.EncodeToString(cred.CredentialID) == credID {
			// Update sign count
			if err := s.repo.UpdateSignCount(cred.ID, credential.Authenticator.SignCount); err != nil {
				s.logger.WithError(err).Error("Failed to update sign count")
			}
			// Update last used
			if err := s.repo.UpdateLastUsed(cred.ID); err != nil {
				s.logger.WithError(err).Error("Failed to update last used")
			}
			break
		}
	}

	s.logger.WithFields(logrus.Fields{
		"userID":       userID,
		"credentialID": credID,
	}).Info("Completed WebAuthn login")

	return true, nil
}

// DeleteCredential deletes a credential by its UUID
func (s *WebAuthnService) DeleteCredential(credentialID uuid.UUID) error {
	if err := s.repo.Delete(credentialID); err != nil {
		s.logger.WithError(err).Error("Failed to delete credential")
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	s.logger.WithField("credentialID", credentialID).Info("Deleted WebAuthn credential")
	return nil
}

// GetCredentialsForUser returns all credentials for a user
func (s *WebAuthnService) GetCredentialsForUser(userID uuid.UUID) ([]*Credential, error) {
	credentials, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credentials for user")
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	result := make([]*Credential, len(credentials))
	for i, cred := range credentials {
		result[i] = &Credential{
			ID:             cred.ID,
			UserID:         cred.UserID,
			CredentialID:   cred.CredentialID,
			PublicKey:      cred.PublicKey,
			SignCount:      cred.SignCount,
			BackupEligible: cred.BackupEligible,
			BackupState:    cred.BackupState,
			CreatedAt:      cred.CreatedAt,
			LastUsedAt:     cred.LastUsedAt,
		}
	}

	return result, nil
}

// GetWebAuthn returns the underlying WebAuthn instance
func (s *WebAuthnService) GetWebAuthn() *webauthn.WebAuthn {
	return s.webAuthn
}

// Helper function to convert storage credentials to WebAuthn credentials
func (s *WebAuthnService) convertCredentials(credentials []*storage.WebAuthnCredential) []*webauthn.Credential {
	result := make([]*webauthn.Credential, len(credentials))
	for i, cred := range credentials {
		result[i] = &webauthn.Credential{
			ID:        cred.CredentialID,
			PublicKey: cred.PublicKey,
			Flags: webauthn.CredentialFlags{
				BackupEligible: cred.BackupEligible,
				BackupState:    cred.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				SignCount: cred.SignCount,
			},
		}
	}
	return result
}
