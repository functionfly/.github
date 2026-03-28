package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UpdateUserMFA updates MFA fields for a user.
func (r *UserRepository) UpdateUserMFA(userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error {
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET mfa_secret = $1, mfa_enabled = $2, mfa_backup_codes = $3, mfa_last_used = $4, updated_at = NOW() WHERE id = $5`,
		secret, enabled, string(backupCodesJSON), lastUsed, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateUserMFAEnabled(userID uuid.UUID, enabled bool) error {
	_, err := r.db.Exec(`
		UPDATE users SET mfa_enabled = $1, updated_at = NOW() WHERE id = $2`,
		enabled, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA enabled: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateUserMFABackupCodes(userID uuid.UUID, backupCodes []string) error {
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET mfa_backup_codes = $1, updated_at = NOW() WHERE id = $2`,
		string(backupCodesJSON), userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA backup codes: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateUserMFALastUsed(userID uuid.UUID, lastUsed *time.Time) error {
	_, err := r.db.Exec(`
		UPDATE users SET mfa_last_used = $1, updated_at = NOW() WHERE id = $2`,
		lastUsed, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA last used: %w", err)
	}

	return nil
}

// UpdateUserSettings updates the settings JSONB field for a user
func (r *UserRepository) UpdateUserSettings(userID uuid.UUID, settings map[string]interface{}) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal user settings: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET settings = $1, updated_at = NOW() WHERE id = $2`,
		settingsJSON, userID)

	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

// GetUserSettings retrieves the settings JSONB field for a user
func (r *UserRepository) GetUserSettings(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Settings == nil {
		return make(map[string]interface{}), nil
	}

	return user.Settings, nil
}

// HashPassword securely hashes a password using bcrypt
func (r *UserRepository) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against the stored hash
func (r *UserRepository) VerifyPassword(userID uuid.UUID, password string) (bool, error) {
	var storedHash string
	err := r.db.QueryRow(`
		SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&storedHash)

	if err == sql.ErrNoRows {
		return false, fmt.Errorf("user not found")
	}
	if err != nil {
		return false, fmt.Errorf("failed to get password hash: %w", err)
	}

	// Verify the password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return true, nil
}
