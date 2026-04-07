package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AdminSessionModel represents an admin session stored in the database (storage layer copy)
type AdminSessionModel struct {
	ID                          uuid.UUID
	UserID                      uuid.UUID
	TokenHash                   string
	IPAddress                   string
	UserAgent                   string
	DeviceFingerprint           string
	CreatedAt                   time.Time
	ExpiresAt                   time.Time
	LastActivityAt              time.Time
	IsRevoked                   bool
	RevokedAt                   *time.Time
	FingerprintMismatchWarnings int
}

// AdminSessionRepository handles admin session-related database operations
type AdminSessionRepository struct {
	db *PostgresDB
}

// NewAdminSessionRepository creates a new admin session repository
func NewAdminSessionRepository(db *PostgresDB) *AdminSessionRepository {
	return &AdminSessionRepository{db: db}
}

// hashToken creates a SHA256 hash of a token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateAdminSession creates a new admin session
func (r *AdminSessionRepository) CreateAdminSession(userID uuid.UUID, token string, ipAddress, userAgent, deviceFingerprint string, expiresAt time.Time) (*AdminSessionModel, error) {
	tokenHash := hashToken(token)
	session := &AdminSessionModel{
		ID:                          uuid.New(),
		UserID:                      userID,
		TokenHash:                   tokenHash,
		IPAddress:                   ipAddress,
		UserAgent:                   userAgent,
		DeviceFingerprint:           deviceFingerprint,
		CreatedAt:                   time.Now(),
		ExpiresAt:                   expiresAt,
		LastActivityAt:              time.Now(),
		IsRevoked:                   false,
		FingerprintMismatchWarnings: 0,
	}

	_, err := r.db.Exec(`
		INSERT INTO admin_sessions (id, user_id, token_hash, ip_address, user_agent, device_fingerprint, created_at, expires_at, last_activity_at, is_revoked, fingerprint_mismatch_warnings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		session.ID, session.UserID, session.TokenHash, session.IPAddress, session.UserAgent, session.DeviceFingerprint, session.CreatedAt, session.ExpiresAt, session.LastActivityAt, session.IsRevoked, session.FingerprintMismatchWarnings)

	if err != nil {
		return nil, fmt.Errorf("failed to create admin session: %w", err)
	}

	return session, nil
}

// GetAdminSessionByToken retrieves an admin session by its token
func (r *AdminSessionRepository) GetAdminSessionByToken(token string) (*AdminSessionModel, error) {
	tokenHash := hashToken(token)

	var session AdminSessionModel
	var deviceFingerprint, revokedAt sql.NullString
	var userAgent sql.NullString

	err := r.db.QueryRow(`
		SELECT id, user_id, token_hash, ip_address, user_agent, device_fingerprint,
			   created_at, expires_at, last_activity_at, is_revoked, revoked_at,
			   fingerprint_mismatch_warnings
		FROM admin_sessions
		WHERE token_hash = $1`,
		tokenHash).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin session not found")
		}
		return nil, fmt.Errorf("failed to get admin session: %w", err)
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

// UpdateAdminSessionLastActivity updates the last activity timestamp for an admin session
func (r *AdminSessionRepository) UpdateAdminSessionLastActivity(sessionID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE admin_sessions
		SET last_activity_at = NOW()
		WHERE id = $1`,
		sessionID)

	if err != nil {
		return fmt.Errorf("failed to update admin session activity: %w", err)
	}

	return nil
}

// RevokeAdminSession revokes an admin session
func (r *AdminSessionRepository) RevokeAdminSession(sessionID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE admin_sessions
		SET is_revoked = true, revoked_at = NOW()
		WHERE id = $1`,
		sessionID)

	if err != nil {
		return fmt.Errorf("failed to revoke admin session: %w", err)
	}

	return nil
}

// RevokeAllAdminUserSessions revokes all admin sessions for a user
func (r *AdminSessionRepository) RevokeAllAdminUserSessions(userID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE admin_sessions
		SET is_revoked = true, revoked_at = NOW()
		WHERE user_id = $1 AND is_revoked = false`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to revoke all admin user sessions: %w", err)
	}

	return nil
}

// ListAdminUserSessions lists all active admin sessions for a user
func (r *AdminSessionRepository) ListAdminUserSessions(userID uuid.UUID) ([]*AdminSessionModel, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, token_hash, ip_address, user_agent, device_fingerprint,
			   created_at, expires_at, last_activity_at, is_revoked, revoked_at,
			   fingerprint_mismatch_warnings
		FROM admin_sessions
		WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()
		ORDER BY last_activity_at DESC`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list admin user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*AdminSessionModel
	for rows.Next() {
		var session AdminSessionModel
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
			return nil, fmt.Errorf("failed to scan admin session: %w", err)
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating admin sessions: %w", err)
	}

	return sessions, nil
}

// DeleteExpiredAdminSessions removes all expired admin sessions and returns the count
func (r *AdminSessionRepository) DeleteExpiredAdminSessions() (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM admin_sessions
		WHERE expires_at <= NOW()`)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired admin sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
