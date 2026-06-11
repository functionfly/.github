package gba

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SessionManager handles session creation and validation
type SessionManager struct {
	secret string
	config *SessionConfig
	logger *logrus.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(secret string, config *SessionConfig) *SessionManager {
	return &SessionManager{
		secret: secret,
		config: config,
		logger: logrus.New(),
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(db *gorm.DB, userID, tenantID uuid.UUID, ipAddress, userAgent string) (*Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &Session{
		UserID:       userID,
		TenantID:     tenantID,
		SessionToken: token,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(sm.config.MaxAge),
	}

	if err := db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session token
func (sm *SessionManager) ValidateSession(db *gorm.DB, token string) (*Session, error) {
	var session Session
	if err := db.Where("session_token = ?", token).First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.IsExpired() {
		// Clean up expired session
		db.Delete(&session)
		return nil, fmt.Errorf("session has expired")
	}

	return &session, nil
}

// InvalidateSession deletes a session
func (sm *SessionManager) InvalidateSession(db *gorm.DB, token string) error {
	return db.Where("session_token = ?", token).Delete(&Session{}).Error
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateAllUserSessions(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("user_id = ?", userID).Delete(&Session{}).Error
}

// GetUserSessions returns all active sessions for a user
func (sm *SessionManager) GetUserSessions(db *gorm.DB, userID uuid.UUID) ([]Session, error) {
	var sessions []Session
	err := db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// CreateSessionTrusted creates a new session for a trusted device with extended expiry (30 days).
// trustedToken is a device-level token used to recognize and trust this device on future logins.
func (sm *SessionManager) CreateSessionTrusted(db *gorm.DB, userID, tenantID uuid.UUID, ipAddress, userAgent string, trustedToken string) (*Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days for trusted devices

	session := &Session{
		UserID:             userID,
		TenantID:           tenantID,
		SessionToken:       token,
		TrustedDeviceToken: trustedToken,
		IPAddress:          ipAddress,
		UserAgent:          userAgent,
		ExpiresAt:          expiresAt,
	}

	if err := db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create trusted session: %w", err)
	}

	return session, nil
}

// GetSessionByTrustedToken returns an active session matching the given trusted device token.
// Used during login to "promote" a device to a trusted session if the user has rememberDevices enabled.
func (sm *SessionManager) GetSessionByTrustedToken(db *gorm.DB, userID uuid.UUID, trustedToken string) (*Session, error) {
	var session Session
	err := db.Where("user_id = ? AND trusted_device_token = ? AND expires_at > ?", userID, trustedToken, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, fmt.Errorf("trusted session not found: %w", err)
	}
	return &session, nil
}

// GetOrCreateTrustedSession returns an existing trusted session or creates a new one.
// This is called during login when rememberDevices is enabled.
func (sm *SessionManager) GetOrCreateTrustedSession(db *gorm.DB, userID, tenantID uuid.UUID, ipAddress, userAgent string, trustedToken string) (*Session, error) {
	// Try to get existing trusted session
	session, err := sm.GetSessionByTrustedToken(db, userID, trustedToken)
	if err == nil {
		// Update activity and return existing session
		db.Model(&session).Updates(map[string]interface{}{
			"last_activity": time.Now(),
			"ip_address":    ipAddress,
			"user_agent":    userAgent,
			"expires_at":    time.Now().Add(30 * 24 * time.Hour), // refresh expiry
		})
		return session, nil
	}

	// Create new trusted session
	return sm.CreateSessionTrusted(db, userID, tenantID, ipAddress, userAgent, trustedToken)
}

// DeleteTrustedSessions deletes all trusted sessions for a user (used when disabling rememberDevices).
func (sm *SessionManager) DeleteTrustedSessions(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("user_id = ? AND trusted_device_token IS NOT NULL AND trusted_device_token != ''", userID).Delete(&Session{}).Error
}

// SetSessionCookie sets the session cookie on an HTTP response
func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sm.config.MaxAge.Seconds()),
		HttpOnly: sm.config.CookieHTTPOnly,
		Secure:   sm.config.CookieSecure,
		SameSite: parseSameSite(sm.config.CookieSameSite),
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: sm.config.CookieHTTPOnly,
		Secure:   sm.config.CookieSecure,
		SameSite: parseSameSite(sm.config.CookieSameSite),
	}
	http.SetCookie(w, cookie)
}

// GetSessionTokenFromRequest extracts the session token from a request
func (sm *SessionManager) GetSessionTokenFromRequest(r *http.Request) string {
	// Check cookie first
	cookie, err := r.Cookie(sm.config.CookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}

// generateSessionToken generates a cryptographically secure session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// parseSameSite parses a SameSite string to http.SameSite
func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
