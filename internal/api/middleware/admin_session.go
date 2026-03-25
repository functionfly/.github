package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// AdminSessionIdleTimeout is the maximum idle time before a session is considered invalid
	AdminSessionIdleTimeout = 30 * time.Minute
	// AdminSessionMaxLifetime is the maximum lifetime of a session
	AdminSessionMaxLifetime = 24 * time.Hour
)

// AdminSession represents an admin session stored in the database
type AdminSession struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"user_id"`
	TokenHash             string     `json:"token_hash"`
	IPAddress             string     `json:"ip_address"`
	UserAgent             string     `json:"user_agent"`
	DeviceFingerprint     string     `json:"device_fingerprint"`
	CreatedAt             time.Time  `json:"created_at"`
	ExpiresAt             time.Time  `json:"expires_at"`
	LastActivityAt        time.Time  `json:"last_activity_at"`
	IsRevoked             bool       `json:"is_revoked"`
	RevokedAt             *time.Time `json:"revoked_at"`
	FingerprintMismatchWarnings int   `json:"fingerprint_mismatch_warnings"`
}

// AdminSessionMiddleware handles admin session validation
type AdminSessionMiddleware struct {
	db       *sql.DB
	authSvc  *auth.AuthService
	logger   *logrus.Entry
}

// NewAdminSessionMiddleware creates a new admin session middleware
func NewAdminSessionMiddleware(db *sql.DB, authSvc *auth.AuthService) *AdminSessionMiddleware {
	return &AdminSessionMiddleware{
		db:      db,
		authSvc: authSvc,
		logger:  logrus.WithField("middleware", "admin_session"),
	}
}

// hashToken creates a SHA256 hash of a token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateSession validates an admin session against the database
func (m *AdminSessionMiddleware) ValidateSession(token string, clientIP, userAgent, deviceFingerprint string) (*AdminSession, error) {
	tokenHash := hashToken(token)

	session, err := m.getSessionByTokenHash(tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			m.logger.WithFields(logrus.Fields{
				"token_hash": tokenHash[:8] + "...",
			}).Warn("Admin session not found")
			return nil, fmt.Errorf("session not found or expired")
		}
		m.logger.WithError(err).Error("Failed to get admin session")
		return nil, fmt.Errorf("failed to validate session")
	}

	// Check if session is revoked
	if session.IsRevoked {
		m.logger.WithFields(logrus.Fields{
			"session_id": session.ID,
			"user_id":    session.UserID,
		}).Warn("Attempt to use revoked admin session")
		return nil, fmt.Errorf("session has been revoked")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		m.logger.WithFields(logrus.Fields{
			"session_id": session.ID,
			"expires_at": session.ExpiresAt,
		}).Info("Admin session has expired")
		return nil, fmt.Errorf("session has expired")
	}

	// Check idle timeout
	idleDuration := time.Since(session.LastActivityAt)
	if idleDuration > AdminSessionIdleTimeout {
		m.logger.WithFields(logrus.Fields{
			"session_id":      session.ID,
			"idle_duration":   idleDuration,
			"idle_timeout":     AdminSessionIdleTimeout,
		}).Info("Admin session idle timeout exceeded")
		return nil, fmt.Errorf("session idle timeout exceeded")
	}

	// Validate IP address (allow some flexibility for proxy scenarios)
	if !m.validateIP(session.IPAddress, clientIP) {
		m.logger.WithFields(logrus.Fields{
			"session_id":    session.ID,
			"session_ip":    session.IPAddress,
			"request_ip":    clientIP,
		}).Warn("IP address mismatch for admin session")
		// Log warning but don't block - could be proxy rotation
	}

	// Validate device fingerprint if provided
	if deviceFingerprint != "" && session.DeviceFingerprint != "" {
		if !m.validateFingerprint(session.DeviceFingerprint, deviceFingerprint) {
			session.FingerprintMismatchWarnings++
			m.logger.WithFields(logrus.Fields{
				"session_id": session.ID,
				"warnings":   session.FingerprintMismatchWarnings,
			}).Warn("Device fingerprint mismatch for admin session")

			// Update warning count in database
			if err := m.updateFingerprintWarnings(session.ID, session.FingerprintMismatchWarnings); err != nil {
				m.logger.WithError(err).Warn("Failed to update fingerprint warnings")
			}
		}
	}

	return session, nil
}

// validateIP checks if the client IP matches the session IP
// Allows for some flexibility with proxies by checking the /24 subnet
func (m *AdminSessionMiddleware) validateIP(sessionIP, clientIP string) bool {
	// Exact match
	if sessionIP == clientIP {
		return true
	}

	// Check if behind same proxy - compare first two octets for IPv4
	sessionIPObj := net.ParseIP(sessionIP)
	clientIPObj := net.ParseIP(clientIP)
	if sessionIPObj == nil || clientIPObj == nil {
		return false
	}

	// Both IPv4
	if sessionIPObj.To4() != nil && clientIPObj.To4() != nil {
		sessionOctets := sessionIPObj.To4()
		clientOctets := clientIPObj.To4()
		// Allow /24 subnet match
		if sessionOctets[0] == clientOctets[0] && sessionOctets[1] == clientOctets[1] && sessionOctets[2] == clientOctets[2] {
			return true
		}
	}

	return false
}

// validateFingerprint compares device fingerprints
func (m *AdminSessionMiddleware) validateFingerprint(stored, provided string) bool {
	// Direct comparison (fingerprints are already hashed)
	return stored == provided
}

// getSessionByTokenHash retrieves a session by its token hash
func (m *AdminSessionMiddleware) getSessionByTokenHash(tokenHash string) (*AdminSession, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, device_fingerprint,
			   created_at, expires_at, last_activity_at, is_revoked, revoked_at,
			   fingerprint_mismatch_warnings
		FROM admin_sessions
		WHERE token_hash = $1`

	var session AdminSession
	var deviceFingerprint, revokedAt sql.NullString
	var userAgent sql.NullString

	err := m.db.QueryRow(query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.IPAddress,
		&userAgent,
		&deviceFingerprint,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivityAt,
		&session.IsRevoked,
		&revokedAt,
		&session.FingerprintMismatchWarnings,
	)

	if err != nil {
		return nil, err
	}

	if userAgent.Valid {
		session.UserAgent = userAgent.String
	}
	if deviceFingerprint.Valid {
		session.DeviceFingerprint = deviceFingerprint.String
	}
	if revokedAt.Valid {
		revokedTime, _ := time.Parse(time.RFC3339, revokedAt.String)
		session.RevokedAt = &revokedTime
	}

	return &session, nil
}

// updateFingerprintWarnings updates the fingerprint mismatch warning count
func (m *AdminSessionMiddleware) updateFingerprintWarnings(sessionID uuid.UUID, warnings int) error {
	if m.db == nil {
		return nil
	}

	query := `UPDATE admin_sessions SET fingerprint_mismatch_warnings = $1 WHERE id = $2`
	_, err := m.db.Exec(query, warnings, sessionID)
	return err
}

// UpdateLastActivity updates the last activity timestamp for a session
func (m *AdminSessionMiddleware) UpdateLastActivity(sessionID uuid.UUID) error {
	if m.db == nil {
		return nil
	}

	query := `UPDATE admin_sessions SET last_activity_at = NOW() WHERE id = $1`
	_, err := m.db.Exec(query, sessionID)
	return err
}

// RevokeSession revokes an admin session
func (m *AdminSessionMiddleware) RevokeSession(sessionID uuid.UUID, userID uuid.UUID) error {
	if m.db == nil {
		return fmt.Errorf("database not configured")
	}

	query := `
		UPDATE admin_sessions
		SET is_revoked = true, revoked_at = NOW()
		WHERE id = $1 AND user_id = $2`

	result, err := m.db.Exec(query, sessionID, userID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to revoke admin session")
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or access denied")
	}

	m.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	}).Info("Admin session revoked")

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (m *AdminSessionMiddleware) RevokeAllUserSessions(userID uuid.UUID) error {
	if m.db == nil {
		return fmt.Errorf("database not configured")
	}

	query := `
		UPDATE admin_sessions
		SET is_revoked = true, revoked_at = NOW()
		WHERE user_id = $1 AND is_revoked = false`

	result, err := m.db.Exec(query, userID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to revoke all user sessions")
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	m.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"sessions_revoked": rowsAffected,
	}).Info("All user sessions revoked")

	return nil
}

// RequireAdminSession middleware validates the admin session on each request
func (m *AdminSessionMiddleware) RequireAdminSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip validation if database is not configured (development mode)
		if m.db == nil {
			m.logger.Warn("Admin session validation skipped - database not configured")
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// Extract client information
		clientIP := extractClientIPFromRequest(r)
		userAgent := r.UserAgent()
		deviceFingerprint := r.Header.Get("X-Device-Fingerprint")

		// Validate session
		session, err := m.ValidateSession(token, clientIP, userAgent, deviceFingerprint)
		if err != nil {
			m.logger.WithFields(logrus.Fields{
				"error":     err.Error(),
				"client_ip": clientIP,
				"path":      r.URL.Path,
			}).Warn("Admin session validation failed")

			// Return generic error to client
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Update last activity asynchronously
		go func() {
			if err := m.UpdateLastActivity(session.ID); err != nil {
				m.logger.WithError(err).WithField("session_id", session.ID).Warn("Failed to update session last activity")
			}
		}()

		// Add session to request context
		ctx := context.WithValue(r.Context(), contextKeyAdminSession, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// contextKeyAdminSession is the context key for admin session
var contextKeyAdminSession = "adminSession"

// GetAdminSessionFromContext extracts the admin session from request context
func GetAdminSessionFromContext(r *http.Request) *AdminSession {
	if session, ok := r.Context().Value(contextKeyAdminSession).(*AdminSession); ok {
		return session
	}
	return nil
}

// HandleRevokeSession handles POST /v1/admin/sessions/revoke
func (m *AdminSessionMiddleware) HandleRevokeSession(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	if err := m.RevokeSession(sessionID, claims.UserID); err != nil {
		m.logger.WithError(err).Error("Failed to revoke session")
		http.Error(w, "Failed to revoke session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleRevokeAllSessions handles POST /v1/admin/sessions/revoke-all
func (m *AdminSessionMiddleware) HandleRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := m.RevokeAllUserSessions(claims.UserID); err != nil {
		m.logger.WithError(err).Error("Failed to revoke all sessions")
		http.Error(w, "Failed to revoke sessions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleListSessions handles GET /v1/admin/sessions
func (m *AdminSessionMiddleware) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessions, err := m.listUserSessions(claims.UserID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to list sessions")
		http.Error(w, "Failed to list sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

// listUserSessions lists all active sessions for a user
func (m *AdminSessionMiddleware) listUserSessions(userID uuid.UUID) ([]*AdminSession, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, device_fingerprint,
			   created_at, expires_at, last_activity_at, is_revoked, revoked_at,
			   fingerprint_mismatch_warnings
		FROM admin_sessions
		WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()
		ORDER BY last_activity_at DESC`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*AdminSession
	for rows.Next() {
		var session AdminSession
		var deviceFingerprint, revokedAt sql.NullString
		var userAgent sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.TokenHash,
			&session.IPAddress,
			&userAgent,
			&deviceFingerprint,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.LastActivityAt,
			&session.IsRevoked,
			&revokedAt,
			&session.FingerprintMismatchWarnings,
		)
		if err != nil {
			return nil, err
		}

		if userAgent.Valid {
			session.UserAgent = userAgent.String
		}
		if deviceFingerprint.Valid {
			session.DeviceFingerprint = deviceFingerprint.String
		}
		if revokedAt.Valid {
			revokedTime, _ := time.Parse(time.RFC3339, revokedAt.String)
			session.RevokedAt = &revokedTime
		}

		// Mask token hash for security
		session.TokenHash = session.TokenHash[:8] + "..."

		sessions = append(sessions, &session)
	}

	return sessions, nil
}


