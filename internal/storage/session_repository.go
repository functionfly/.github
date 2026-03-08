package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionRepository handles session-related database operations
type SessionRepository struct {
	db *PostgresDB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *PostgresDB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession creates a new session
func (r *SessionRepository) CreateSession(userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error) {
	session := &Session{
		ID:           uuid.New(),
		UserID:       userID,
		SessionToken: sessionToken,
		MFAVerified:  false,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expiresAt,
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO sessions (id, user_id, session_token, mfa_verified, mfa_last_used, ip_address, user_agent, expires_at, last_activity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		session.ID, session.UserID, session.SessionToken, session.MFAVerified, session.MFALastUsed, session.IPAddress, session.UserAgent, session.ExpiresAt, session.LastActivity, session.CreatedAt, session.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSessionByToken retrieves a session by its token
func (r *SessionRepository) GetSessionByToken(sessionToken string) (*Session, error) {
	var session Session
	var mfaLastUsed sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, session_token, mfa_verified, mfa_last_used, ip_address, user_agent, expires_at, last_activity, created_at, updated_at
		FROM sessions
		WHERE session_token = $1 AND expires_at > NOW()`,
		sessionToken).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.MFAVerified,
		&mfaLastUsed, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.LastActivity, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if mfaLastUsed.Valid {
		session.MFALastUsed = &mfaLastUsed.Time
	}

	return &session, nil
}

// GetSessionByID retrieves a session by its ID
func (r *SessionRepository) GetSessionByID(sessionID uuid.UUID) (*Session, error) {
	var session Session
	var mfaLastUsed sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, session_token, mfa_verified, mfa_last_used, ip_address, user_agent, expires_at, last_activity, created_at, updated_at
		FROM sessions
		WHERE id = $1 AND expires_at > NOW()`,
		sessionID).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.MFAVerified,
		&mfaLastUsed, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.LastActivity, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if mfaLastUsed.Valid {
		session.MFALastUsed = &mfaLastUsed.Time
	}

	return &session, nil
}

// UpdateSessionMFAStatus updates the MFA verification status of a session
func (r *SessionRepository) UpdateSessionMFAStatus(sessionToken string, mfaVerified bool) error {
	now := time.Now()
	// Use distinct parameter indices: $1=mfaVerified, $2=mfa_last_used, $3=last_activity,
	// $4=updated_at, $5=sessionToken — avoids $1 being reused in the WHERE clause.
	_, err := r.db.Exec(`
		UPDATE sessions
		SET mfa_verified = $1, mfa_last_used = $2, last_activity = $3, updated_at = $4
		WHERE session_token = $5 AND expires_at > NOW()`,
		mfaVerified, now, now, now, sessionToken)

	if err != nil {
		return fmt.Errorf("failed to update session MFA status: %w", err)
	}

	return nil
}

// UpdateSessionActivity updates the last activity timestamp of a session
func (r *SessionRepository) UpdateSessionActivity(sessionToken string) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE sessions
		SET last_activity = $1, updated_at = $2
		WHERE session_token = $3 AND expires_at > NOW()`,
		now, now, sessionToken)

	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	return nil
}

// DeleteSession deletes a session by its token
func (r *SessionRepository) DeleteSession(sessionToken string) error {
	_, err := r.db.Exec(`
		DELETE FROM sessions
		WHERE session_token = $1`,
		sessionToken)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions removes all expired sessions and returns the count
func (r *SessionRepository) DeleteExpiredSessions() (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM sessions
		WHERE expires_at <= NOW()`)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// DeleteUserSessions deletes all sessions for a specific user
func (r *SessionRepository) DeleteUserSessions(userID uuid.UUID) error {
	_, err := r.db.Exec(`
		DELETE FROM sessions
		WHERE user_id = $1`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// DeleteSessionByID deletes a single session by ID if it belongs to the given user
func (r *SessionRepository) DeleteSessionByID(sessionID, userID uuid.UUID) error {
	result, err := r.db.Exec(`
		DELETE FROM sessions
		WHERE id = $1 AND user_id = $2`,
		sessionID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("session not found or access denied")
	}

	return nil
}

// ListUserSessions retrieves all active sessions for a user
func (r *SessionRepository) ListUserSessions(userID uuid.UUID) ([]*Session, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, session_token, mfa_verified, mfa_last_used, ip_address, user_agent, expires_at, last_activity, created_at, updated_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_activity DESC`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var mfaLastUsed sql.NullTime

		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken, &session.MFAVerified,
			&mfaLastUsed, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
			&session.LastActivity, &session.CreatedAt, &session.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if mfaLastUsed.Valid {
			session.MFALastUsed = &mfaLastUsed.Time
		}

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// CountActiveUserSessions counts the number of active sessions for a user
func (r *SessionRepository) CountActiveUserSessions(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()`,
		userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count active user sessions: %w", err)
	}

	return count, nil
}

// ListTenantSessions retrieves sessions for all users in a tenant
func (r *SessionRepository) ListTenantSessions(tenantID uuid.UUID, limit, offset int) ([]*Session, error) {
	query := `
		SELECT s.id, s.user_id, s.session_token, s.mfa_verified, s.mfa_last_used, s.ip_address, s.user_agent, s.expires_at, s.last_activity, s.created_at, s.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE u.tenant_id = $1 AND s.expires_at > NOW()
		ORDER BY s.last_activity DESC`

	args := []interface{}{tenantID}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var mfaLastUsed sql.NullTime

		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken, &session.MFAVerified,
			&mfaLastUsed, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
			&session.LastActivity, &session.CreatedAt, &session.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if mfaLastUsed.Valid {
			session.MFALastUsed = &mfaLastUsed.Time
		}

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// DeleteSessionByIDOnly deletes a session by ID without checking user ownership
func (r *SessionRepository) DeleteSessionByIDOnly(sessionID uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.Exec(`
		DELETE FROM sessions
		WHERE id = $1 AND user_id = $2`,
		sessionID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete session by ID: %w", err)
	}

	return nil
}
