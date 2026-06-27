package storage

import (
	"context"
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
func (r *SessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error) {
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

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, session_token, mfa_verified, mfa_last_used, ip_address, user_agent, expires_at, last_activity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		session.ID, session.UserID, session.SessionToken, session.MFAVerified, session.MFALastUsed, session.IPAddress, session.UserAgent, session.ExpiresAt, session.LastActivity, session.CreatedAt, session.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSessionByToken retrieves a session by its token
func (r *SessionRepository) GetSessionByToken(ctx context.Context, sessionToken string) (*Session, error) {
	var session Session
	var mfaLastUsed sql.NullTime

	err := r.db.QueryRowContext(ctx, `
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
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	var session Session
	var mfaLastUsed sql.NullTime

	err := r.db.QueryRowContext(ctx, `
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
func (r *SessionRepository) UpdateSessionMFAStatus(ctx context.Context, sessionToken string, mfaVerified bool) error {
	now := time.Now()
	// Use distinct parameter indices: $1=mfaVerified, $2=mfa_last_used, $3=last_activity,
	// $4=updated_at, $5=sessionToken — avoids $1 being reused in the WHERE clause.
	_, err := r.db.ExecContext(ctx, `
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
func (r *SessionRepository) UpdateSessionActivity(ctx context.Context, sessionToken string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
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
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE session_token = $1`,
		sessionToken)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions removes all expired sessions and returns the count
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
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
func (r *SessionRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE user_id = $1`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// DeleteSessionByID deletes a single session by ID if it belongs to the given user
func (r *SessionRepository) DeleteSessionByID(ctx context.Context, sessionID, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
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
func (r *SessionRepository) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	rows, err := r.db.QueryContext(ctx, `
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
func (r *SessionRepository) CountActiveUserSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
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
func (r *SessionRepository) ListTenantSessions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Session, error) {
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

	rows, err := r.db.QueryContext(ctx, query, args...)
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
func (r *SessionRepository) DeleteSessionByIDOnly(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE id = $1 AND user_id = $2`,
		sessionID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete session by ID: %w", err)
	}

	return nil
}

// LoginHistory represents a login event record
type LoginHistory struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	EventType   string     `json:"event_type" db:"event_type"` // 'login', 'logout', 'logout_other', 'session_expired', 'revoked'
	IPAddress   string     `json:"ip_address" db:"ip_address"`
	UserAgent   string     `json:"user_agent,omitempty" db:"user_agent"`
	Device      string     `json:"device,omitempty" db:"device"`
	Location    string     `json:"location,omitempty" db:"location"`
	LoginMethod string     `json:"login_method,omitempty" db:"login_method"`
	MFAUsed     bool       `json:"mfa_used" db:"mfa_used"`
	SessionID   *uuid.UUID `json:"session_id,omitempty" db:"session_id"`
	Metadata    []byte     `json:"metadata,omitempty" db:"metadata"` // JSONB
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// CreateLoginHistory records a login event
func (r *SessionRepository) CreateLoginHistory(ctx context.Context, userID uuid.UUID, eventType, ipAddress, userAgent, device, loginMethod string, mfaUsed bool, sessionID *uuid.UUID) (*LoginHistory, error) {
	loginHistory := &LoginHistory{
		ID:          uuid.New(),
		UserID:      userID,
		EventType:   eventType,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Device:      device,
		LoginMethod: loginMethod,
		MFAUsed:     mfaUsed,
		SessionID:   sessionID,
		CreatedAt:   time.Now(),
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO login_history (id, user_id, event_type, ip_address, user_agent, device, login_method, mfa_used, session_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		loginHistory.ID, loginHistory.UserID, loginHistory.EventType, loginHistory.IPAddress,
		loginHistory.UserAgent, loginHistory.Device, loginHistory.LoginMethod, loginHistory.MFAUsed,
		loginHistory.SessionID, loginHistory.Metadata, loginHistory.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create login history: %w", err)
	}

	return loginHistory, nil
}

// ListUserLoginHistory retrieves login history for a user with pagination
func (r *SessionRepository) ListUserLoginHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*LoginHistory, error) {
	query := `
		SELECT id, user_id, event_type, ip_address, user_agent, device, location, login_method, mfa_used, session_id, metadata, created_at
		FROM login_history
		WHERE user_id = $1
		ORDER BY created_at DESC`

	args := []interface{}{userID}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list login history: %w", err)
	}
	defer rows.Close()

	var history []*LoginHistory
	for rows.Next() {
		var h LoginHistory
		var userAgent, device, location, loginMethod sql.NullString
		var sessionID sql.NullString
		var metadata []byte

		err := rows.Scan(
			&h.ID, &h.UserID, &h.EventType, &h.IPAddress,
			&userAgent, &device, &location, &loginMethod,
			&h.MFAUsed, &sessionID, &metadata, &h.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan login history: %w", err)
		}

		if userAgent.Valid {
			h.UserAgent = userAgent.String
		}
		if device.Valid {
			h.Device = device.String
		}
		if location.Valid {
			h.Location = location.String
		}
		if loginMethod.Valid {
			h.LoginMethod = loginMethod.String
		}
		if sessionID.Valid {
			id, _ := uuid.Parse(sessionID.String)
			h.SessionID = &id
		}
		h.Metadata = metadata

		history = append(history, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating login history: %w", err)
	}

	return history, nil
}

// CountUserLoginHistory counts total login history entries for a user
func (r *SessionRepository) CountUserLoginHistory(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM login_history
		WHERE user_id = $1`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count login history: %w", err)
	}

	return count, nil
}
