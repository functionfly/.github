package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LoginAttemptRepository handles login attempt-related database operations
type LoginAttemptRepository struct {
	db *PostgresDB
}

// NewLoginAttemptRepository creates a new login attempt repository
func NewLoginAttemptRepository(db *PostgresDB) *LoginAttemptRepository {
	return &LoginAttemptRepository{db: db}
}

// CreateLoginAttempt creates a new login attempt record
func (r *LoginAttemptRepository) CreateLoginAttempt(userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*LoginAttempt, error) {
	attempt := &LoginAttempt{
		ID:           uuid.New(),
		UserID:       userID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Successful:   successful,
		AttemptedAt:  time.Now(),
		LockoutUntil: lockoutUntil,
		CreatedAt:    time.Now(),
	}

	query := `
		INSERT INTO login_attempts (id, user_id, ip_address, user_agent, successful, attempted_at, lockout_until, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query,
		attempt.ID, attempt.UserID, attempt.IPAddress, attempt.UserAgent,
		attempt.Successful, attempt.AttemptedAt, attempt.LockoutUntil, attempt.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create login attempt: %w", err)
	}

	return attempt, nil
}

// GetRecentFailedLoginAttempts counts failed login attempts for a user within a time window
func (r *LoginAttemptRepository) GetRecentFailedLoginAttempts(userID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM login_attempts
		WHERE user_id = $1 AND successful = false AND attempted_at >= $2`,
		userID, since).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count failed login attempts: %w", err)
	}

	return count, nil
}

// GetUserLockoutStatus returns the lockout expiration time for a user if they are currently locked out
func (r *LoginAttemptRepository) GetUserLockoutStatus(userID uuid.UUID) (*time.Time, error) {
	var lockoutUntil sql.NullTime
	err := r.db.QueryRow(`
		SELECT lockout_until
		FROM login_attempts
		WHERE user_id = $1 AND lockout_until IS NOT NULL AND lockout_until > NOW()
		ORDER BY lockout_until DESC
		LIMIT 1`,
		userID).Scan(&lockoutUntil)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active lockout
		}
		return nil, fmt.Errorf("failed to get user lockout status: %w", err)
	}

	if lockoutUntil.Valid {
		return &lockoutUntil.Time, nil
	}

	return nil, nil
}

// ClearUserLockout removes all lockout records for a user
func (r *LoginAttemptRepository) ClearUserLockout(userID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE login_attempts
		SET lockout_until = NULL
		WHERE user_id = $1 AND lockout_until IS NOT NULL`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to clear user lockout: %w", err)
	}

	return nil
}

// DeleteOldLoginAttempts removes login attempts older than the specified time
func (r *LoginAttemptRepository) DeleteOldLoginAttempts(before time.Time) (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM login_attempts
		WHERE created_at < $1`,
		before)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetLastSuccessfulLogin returns the most recent successful login attempt for a user.
// Returns nil if no successful login is found.
func (r *LoginAttemptRepository) GetLastSuccessfulLogin(userID uuid.UUID) (*LoginAttempt, error) {
	var attempt LoginAttempt
	err := r.db.QueryRow(`
		SELECT id, user_id, ip_address, user_agent, successful, attempted_at, lockout_until, created_at
		FROM login_attempts
		WHERE user_id = $1 AND successful = true
		ORDER BY attempted_at DESC
		LIMIT 1`,
		userID).Scan(
		&attempt.ID, &attempt.UserID, &attempt.IPAddress, &attempt.UserAgent,
		&attempt.Successful, &attempt.AttemptedAt, &attempt.LockoutUntil, &attempt.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last successful login: %w", err)
	}

	return &attempt, nil
}