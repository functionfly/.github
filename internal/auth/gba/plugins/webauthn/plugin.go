// Package webauthn provides WebAuthn/Passkeys authentication support for GoBetterAuth
package webauthn

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserGetter defines the interface for retrieving user information
type UserGetter interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*WebAuthnUser, error)
}

// DBUserGetter implements UserGetter using direct database access
type DBUserGetter struct {
	db *gorm.DB
}

// GetUserByID retrieves a user by ID from the database
func (g *DBUserGetter) GetUserByID(ctx context.Context, userID uuid.UUID) (*WebAuthnUser, error) {
	var user GBAUser
	result := g.db.WithContext(ctx).Table("gba_users").
		Select("id, email, COALESCE(display_name, email) as display_name").
		Where("id = ?", userID).
		First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}

	return &WebAuthnUser{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}, nil
}

// WebAuthnPlugin provides WebAuthn/Passkeys authentication functionality
type WebAuthnPlugin struct {
	db         *gorm.DB
	webauthn   *webauthn.WebAuthn
	config     *WebAuthnConfig
	logger     *logrus.Logger
	userGetter UserGetter
}

// GBAUser represents a minimal user interface for GoBetterAuth integration
type GBAUser struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
}

// New creates a new WebAuthn plugin instance
func New(db *gorm.DB, config *WebAuthnConfig, logger *logrus.Logger, userGetter UserGetter) (*WebAuthnPlugin, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if config == nil {
		config = DefaultWebAuthnConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	if userGetter == nil {
		userGetter = &DBUserGetter{db: db}
	}

	// Load from environment if not provided
	if config.RPDisplayName == "" {
		config.RPDisplayName = getEnvOrDefault("WEBAUTHN_RP_DISPLAY_NAME", "FunctionFly")
	}
	if config.RPID == "" {
		config.RPID = getEnvOrDefault("WEBAUTHN_RP_ID", "localhost")
	}
	if config.RPOrigin == "" {
		config.RPOrigin = getEnvOrDefault("WEBAUTHN_RP_ORIGIN", "http://localhost:3000")
	}

	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigin:      config.RPOrigin,
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	plugin := &WebAuthnPlugin{
		db:         db,
		webauthn:   w,
		config:     config,
		logger:     logger,
		userGetter: userGetter,
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&WebAuthnCredential{}, &WebAuthnSession{}); err != nil {
		return nil, fmt.Errorf("failed to migrate WebAuthn models: %w", err)
	}

	logger.Info("WebAuthn plugin initialized")
	return plugin, nil
}

// IsEnabled returns true if WebAuthn is enabled
func (p *WebAuthnPlugin) IsEnabled() bool {
	// WebAuthn is enabled if the plugin is initialized
	return p != nil && p.webauthn != nil
}

// IsEnabledForUser checks if WebAuthn is enabled for a specific user
func (p *WebAuthnPlugin) IsEnabledForUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	result := p.db.WithContext(ctx).Model(&WebAuthnCredential{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check WebAuthn status: %w", result.Error)
	}
	return count > 0, nil
}

// GetStatus returns the WebAuthn status for a user
func (p *WebAuthnPlugin) GetStatus(ctx context.Context, userID uuid.UUID) (*WebAuthnStatusResponse, error) {
	var count int64
	if err := p.db.WithContext(ctx).Model(&WebAuthnCredential{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to get credential count: %w", err)
	}

	return &WebAuthnStatusResponse{
		Enabled:         count > 0,
		CredentialCount: int(count),
	}, nil
}

// GetUserCredentials returns all credentials for a user
func (p *WebAuthnPlugin) GetUserCredentials(ctx context.Context, userID uuid.UUID) ([]CredentialInfo, error) {
	var credentials []WebAuthnCredential
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Find(&credentials).Error; err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	result := make([]CredentialInfo, len(credentials))
	for i, cred := range credentials {
		result[i] = CredentialInfo{
			ID:         cred.ID,
			Name:       cred.Name,
			CreatedAt:  cred.CreatedAt,
			LastUsedAt: cred.LastUsedAt,
			SignCount:  cred.SignCount,
		}
	}

	return result, nil
}

// GetCredentialByID retrieves a credential by its ID
func (p *WebAuthnPlugin) GetCredentialByID(ctx context.Context, credentialID uuid.UUID, userID uuid.UUID) (*WebAuthnCredential, error) {
	var cred WebAuthnCredential
	result := p.db.WithContext(ctx).Where("id = ? AND user_id = ?", credentialID, userID).First(&cred)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("credential not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get credential: %w", result.Error)
	}
	return &cred, nil
}

// UpdateCredentialName updates the name of a credential
func (p *WebAuthnPlugin) UpdateCredentialName(ctx context.Context, credentialID uuid.UUID, userID uuid.UUID, name string) error {
	result := p.db.WithContext(ctx).Model(&WebAuthnCredential{}).
		Where("id = ? AND user_id = ?", credentialID, userID).
		Update("name", name)

	if result.Error != nil {
		return fmt.Errorf("failed to update credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"credential_id": credentialID,
		"name":          name,
	}).Info("Credential name updated")

	return nil
}

// DeleteCredential deletes a credential
func (p *WebAuthnPlugin) DeleteCredential(ctx context.Context, credentialID uuid.UUID, userID uuid.UUID) error {
	result := p.db.WithContext(ctx).Where("id = ? AND user_id = ?", credentialID, userID).Delete(&WebAuthnCredential{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"credential_id": credentialID,
	}).Info("Credential deleted")

	return nil
}

// getUserByID retrieves a user by ID using the injected user service
func (p *WebAuthnPlugin) getUserByID(ctx context.Context, userID uuid.UUID) (*WebAuthnUser, error) {
	return p.userGetter.GetUserByID(ctx, userID)
}

// getUserByIDString retrieves a user by ID string for authentication
func (p *WebAuthnPlugin) getUserByIDString(ctx context.Context, userID string) (*WebAuthnUser, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return p.getUserByID(ctx, id)
}

// loadUserCredentials loads credentials for a user
func (p *WebAuthnPlugin) loadUserCredentials(ctx context.Context, userID uuid.UUID) ([]WebAuthnCredential, error) {
	var credentials []WebAuthnCredential
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

// createSession creates a new WebAuthn session
func (p *WebAuthnPlugin) createSession(ctx context.Context, userID uuid.UUID, challenge, operation string) (*WebAuthnSession, error) {
	session := &WebAuthnSession{
		UserID:    userID,
		Challenge: challenge,
		Operation: operation,
		ExpiresAt: time.Now().Add(p.config.SessionTimeout),
	}

	if err := p.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// getSession retrieves a session by ID
func (p *WebAuthnPlugin) getSession(ctx context.Context, sessionID uuid.UUID) (*WebAuthnSession, error) {
	var session WebAuthnSession
	result := p.db.WithContext(ctx).Where("id = ?", sessionID).First(&session)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("session not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get session: %w", result.Error)
	}

	if session.IsExpired() {
		// Clean up expired session
		p.db.Delete(&session)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// deleteSession deletes a session
func (p *WebAuthnPlugin) deleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return p.db.WithContext(ctx).Where("id = ?", sessionID).Delete(&WebAuthnSession{}).Error
}

// storeCredential stores a new credential
func (p *WebAuthnPlugin) storeCredential(ctx context.Context, cred *WebAuthnCredential) error {
	return p.db.WithContext(ctx).Create(cred).Error
}

// updateCredentialSignCount updates the sign count for a credential
func (p *WebAuthnPlugin) updateCredentialSignCount(ctx context.Context, credentialID []byte, signCount uint32) error {
	return p.db.WithContext(ctx).Model(&WebAuthnCredential{}).
		Where("credential_id = ?", credentialID).
		Updates(map[string]interface{}{
			"sign_count":   signCount,
			"last_used_at": time.Now(),
		}).Error
}

// getCredentialByCredentialID retrieves a credential by its credential ID
func (p *WebAuthnPlugin) getCredentialByCredentialID(ctx context.Context, credentialID []byte) (*WebAuthnCredential, error) {
	var cred WebAuthnCredential
	result := p.db.WithContext(ctx).Where("credential_id = ?", credentialID).First(&cred)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("credential not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get credential: %w", result.Error)
	}
	return &cred, nil
}

// cleanExpiredSessions removes expired sessions
func (p *WebAuthnPlugin) cleanExpiredSessions(ctx context.Context) error {
	return p.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&WebAuthnSession{}).Error
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CredentialToDescriptor converts a WebAuthnCredential to a protocol.CredentialDescriptor
func CredentialToDescriptor(cred *WebAuthnCredential) *protocol.CredentialDescriptor {
	transports := make([]protocol.AuthenticatorTransport, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = protocol.AuthenticatorTransport(t)
	}

	return &protocol.CredentialDescriptor{
		Type:         protocol.PublicKeyCredentialType,
		CredentialID: cred.CredentialID,
		Transport:    transports,
	}
}

// base64URLEncode encodes bytes to base64url string
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// base64URLDecode decodes a base64url string to bytes
func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// toCredentialTransport converts string slice to protocol.AuthenticatorTransport slice
func toCredentialTransport(transports []string) []protocol.AuthenticatorTransport {
	result := make([]protocol.AuthenticatorTransport, len(transports))
	for i, t := range transports {
		result[i] = protocol.AuthenticatorTransport(t)
	}
	return result
}
