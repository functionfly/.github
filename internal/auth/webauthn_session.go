package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// WebAuthnSessionTTL is the time a WebAuthn session is valid
	WebAuthnSessionTTL = 60 * time.Second

	// webAuthnSessionKeyPrefix is the prefix for Redis keys
	webAuthnSessionKeyPrefix = "webauthn:session:"
)

// WebAuthnSession represents session data stored during WebAuthn ceremony
type WebAuthnSession struct {
	Challenge    string          `json:"challenge"`
	UserHandle   string          `json:"userHandle"`
	UserID       string          `json:"userId"`
	CredentialID []byte          `json:"credentialId,omitempty"`
	Operation    string          `json:"operation"`   // "registration" or "authentication"
	SessionData  json.RawMessage `json:"sessionData"` // The webauthn.SessionData JSON
	ExpiresAt    time.Time       `json:"expiresAt"`
}

// WebAuthnSessionStore handles Redis-based session storage for WebAuthn
type WebAuthnSessionStore struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewWebAuthnSessionStore creates a new WebAuthn session store
func NewWebAuthnSessionStore(client *redis.Client) *WebAuthnSessionStore {
	return &WebAuthnSessionStore{
		client: client,
		logger: logrus.New(),
	}
}

// Create stores a new WebAuthn session and returns the session ID
func (s *WebAuthnSessionStore) Create(ctx context.Context, session *WebAuthnSession) (string, error) {
	// Generate a cryptographically secure session ID using crypto/rand
	sessionID, err := generateSecureSessionID()
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate secure session ID")
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Set expiration
	session.ExpiresAt = time.Now().Add(WebAuthnSessionTTL)

	// Marshal session data
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal WebAuthn session")
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store in Redis with TTL
	key := webAuthnSessionKeyPrefix + sessionID
	if err := s.client.Set(ctx, key, sessionJSON, WebAuthnSessionTTL).Err(); err != nil {
		s.logger.WithError(err).Error("Failed to store WebAuthn session")
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"sessionID": sessionID,
		"operation": session.Operation,
		"userID":    session.UserID,
		"ttl":       WebAuthnSessionTTL,
	}).Debug("Created WebAuthn session")

	return sessionID, nil
}

// Get retrieves a WebAuthn session by ID
func (s *WebAuthnSessionStore) Get(ctx context.Context, sessionID string) (*WebAuthnSession, error) {
	key := webAuthnSessionKeyPrefix + sessionID

	sessionJSON, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		s.logger.Debug("WebAuthn session not found or expired")
		return nil, nil
	}
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve WebAuthn session")
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	var session WebAuthnSession
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal WebAuthn session")
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.logger.Debug("WebAuthn session expired")
		// Try to delete anyway (might already be gone)
		s.client.Del(ctx, key)
		return nil, nil
	}

	return &session, nil
}

// Delete removes a WebAuthn session
func (s *WebAuthnSessionStore) Delete(ctx context.Context, sessionID string) error {
	key := webAuthnSessionKeyPrefix + sessionID

	if err := s.client.Del(ctx, key).Err(); err != nil {
		s.logger.WithError(err).Error("Failed to delete WebAuthn session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.WithField("sessionID", sessionID).Debug("Deleted WebAuthn session")
	return nil
}

// DeleteByUserID removes all WebAuthn sessions for a user
func (s *WebAuthnSessionStore) DeleteByUserID(ctx context.Context, userID string) error {
	pattern := webAuthnSessionKeyPrefix + "*"

	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keysToDelete []string

	for iter.Next(ctx) {
		key := iter.Val()
		// Get the session to check userID
		sessionJSON, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var session WebAuthnSession
		if err := json.Unmarshal(sessionJSON, &session); err != nil {
			continue
		}

		if session.UserID == userID {
			keysToDelete = append(keysToDelete, key)
		}
	}

	if err := iter.Err(); err != nil {
		s.logger.WithError(err).Error("Failed to scan WebAuthn sessions")
	}

	if len(keysToDelete) > 0 {
		if err := s.client.Del(ctx, keysToDelete...).Err(); err != nil {
			s.logger.WithError(err).Error("Failed to delete WebAuthn sessions")
			return fmt.Errorf("failed to delete sessions: %w", err)
		}

		s.logger.WithFields(logrus.Fields{
			"userID":  userID,
			"deleted": len(keysToDelete),
		}).Debug("Deleted WebAuthn sessions for user")
	}

	return nil
}

// generateSecureSessionID generates a cryptographically secure session ID using crypto/rand
func generateSecureSessionID() (string, error) {
	// Generate 16 bytes (128 bits) of cryptographically secure random data
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	// Format: hex encoded random bytes (32 hex characters)
	// Additional uniqueness from timestamp not needed when using crypto/rand
	return hex.EncodeToString(b), nil
}
