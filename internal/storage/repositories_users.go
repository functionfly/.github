package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostgresDB methods: users and OAuth CSRF state.

// User operations
func (db *PostgresDB) CreateUser(email, passwordHash string, tenantID uuid.UUID) (*User, error) {
	return db.userRepository.CreateUser(email, passwordHash, tenantID)
}

func (db *PostgresDB) CreateUserWithSocialAuth(email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*User, error) {
	return db.userRepository.CreateUserWithSocialAuth(email, tenantID, provider, providerID, providerData)
}

func (db *PostgresDB) CreateUserWithRole(ctx context.Context, user *User) (*User, error) {
	return db.userRepository.CreateUserWithRole(ctx, user)
}

func (db *PostgresDB) GetUserByEmail(email string) (*User, error) {
	return db.userRepository.GetUserByEmail(email)
}

func (db *PostgresDB) GetUserByID(userID uuid.UUID) (*User, error) {
	return db.userRepository.GetUserByID(userID)
}

func (db *PostgresDB) GetUserByUsername(username string) (*User, error) {
	return db.userRepository.GetUserByUsername(username)
}

func (db *PostgresDB) GetUserForPublicProfile(login string) (*User, error) {
	return db.userRepository.GetUserForPublicProfile(login)
}

func (db *PostgresDB) SearchUsersByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]UserSearchHit, error) {
	return db.userRepository.SearchUsersByUsernamePrefix(ctx, prefix, limit)
}

// IsUsernameReserved checks if a username is reserved
func (db *PostgresDB) IsUsernameReserved(username string) (bool, error) {
	return db.userRepository.IsUsernameReserved(username)
}

func (db *PostgresDB) GetUserByVerificationToken(token string) (*User, error) {
	return db.userRepository.GetUserByVerificationToken(token)
}

func (db *PostgresDB) GetUserBySocialProvider(provider, providerID string) (*User, error) {
	return db.userRepository.GetUserBySocialProvider(provider, providerID)
}

func (db *PostgresDB) ListUsers() ([]*User, error) {
	return db.userRepository.ListUsers()
}

func (db *PostgresDB) ListUserIDsByTenant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error) {
	return db.userRepository.ListUserIDsByTenant(ctx, tenantID)
}

func (db *PostgresDB) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*User, error) {
	return db.userRepository.UpdateUser(ctx, userID, updates)
}

func (db *PostgresDB) UpdateUserEmailVerification(ctx context.Context, userID uuid.UUID, verified bool, token *string, expiresAt *time.Time) error {
	return db.userRepository.UpdateUserEmailVerification(ctx, userID, verified, token, expiresAt)
}

func (db *PostgresDB) UpdateUserProviderData(userID uuid.UUID, providerData map[string]interface{}) error {
	return db.userRepository.UpdateUserProviderData(userID, providerData)
}

func (db *PostgresDB) UpdateUserSettings(userID uuid.UUID, settings map[string]interface{}) error {
	return db.userRepository.UpdateUserSettings(userID, settings)
}

func (db *PostgresDB) GetUserSettings(userID uuid.UUID) (map[string]interface{}, error) {
	return db.userRepository.GetUserSettings(userID)
}

// ListActiveUsersByTenant lists all active (non-deactivated) users for a tenant
func (db *PostgresDB) ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*User, error) {
	return db.userRepository.ListActiveUsersByTenant(ctx, tenantID)
}

// CountActiveUsersByTenant counts all active (non-deactivated) users for a tenant
func (db *PostgresDB) CountActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return db.userRepository.CountActiveUsersByTenant(ctx, tenantID)
}

// DeactivateUser soft-deletes a user (sets deactivated_at and deactivated_by)
func (db *PostgresDB) DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID) error {
	return db.userRepository.DeactivateUser(ctx, userID, deactivatedBy)
}

// ReactivateUser reactivates a previously deactivated user
func (db *PostgresDB) ReactivateUser(ctx context.Context, userID uuid.UUID) error {
	return db.userRepository.ReactivateUser(ctx, userID)
}

func (db *PostgresDB) UpdateUserMFA(userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error {
	return db.userRepository.UpdateUserMFA(userID, secret, enabled, backupCodes, lastUsed)
}

func (db *PostgresDB) UpdateUserMFAEnabled(userID uuid.UUID, enabled bool) error {
	return db.userRepository.UpdateUserMFAEnabled(userID, enabled)
}

func (db *PostgresDB) UpdateUserMFABackupCodes(userID uuid.UUID, backupCodes []string) error {
	return db.userRepository.UpdateUserMFABackupCodes(userID, backupCodes)
}

func (db *PostgresDB) UpdateUserMFALastUsed(userID uuid.UUID, lastUsed *time.Time) error {
	return db.userRepository.UpdateUserMFALastUsed(userID, lastUsed)
}

func (db *PostgresDB) VerifyPassword(userID uuid.UUID, password string) (bool, error) {
	return db.userRepository.VerifyPassword(userID, password)
}

// Username change history operations
func (db *PostgresDB) CreateUsernameChangeHistory(ctx context.Context, history *UsernameChangeHistory) error {
	return db.userRepository.CreateUsernameChangeHistory(ctx, history)
}

func (db *PostgresDB) ChangeUsernameWithHistory(ctx context.Context, userID uuid.UUID, newUsername string, history *UsernameChangeHistory) error {
	return db.userRepository.ChangeUsernameWithHistory(ctx, userID, newUsername, history)
}

func (db *PostgresDB) GetUsernameChangeHistory(ctx context.Context, userID uuid.UUID) ([]*UsernameChangeHistory, error) {
	return db.userRepository.GetUsernameChangeHistory(ctx, userID)
}

func (db *PostgresDB) CountUsernameChangesInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (int, error) {
	return db.userRepository.CountUsernameChangesInWindow(ctx, userID, windowStart)
}

func (db *PostgresDB) GetLastUsernameChange(ctx context.Context, userID uuid.UUID) (*UsernameChangeHistory, error) {
	return db.userRepository.GetLastUsernameChange(ctx, userID)
}

func (db *PostgresDB) HasUsernameChangedInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (bool, error) {
	return db.userRepository.HasUsernameChangedInWindow(ctx, userID, windowStart)
}

// OAuth state (CSRF) — persisted for multi-instance OAuth flows
func (db *PostgresDB) StoreOAuthState(ctx context.Context, state string, expiresAt time.Time, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string) error {
	row := &OAuthState{
		State:             state,
		ExpiresAt:         expiresAt,
		RedirectURI:       redirectURI,
		InviteCode:        inviteCode,
		CodeVerifier:      codeVerifier,
		LoginHint:         loginHint,
		DeviceFingerprint: deviceFingerprint,
	}
	return db.GORM.WithContext(ctx).Create(row).Error
}

func (db *PostgresDB) ValidateAndConsumeOAuthState(ctx context.Context, state string) (bool, string, string, string, string, string, error) {
	var row OAuthState
	err := db.GORM.WithContext(ctx).Where("state = ? AND expires_at > ?", state, time.Now()).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "", "", "", "", "", nil
		}
		return false, "", "", "", "", "", err
	}
	redirectURI := row.RedirectURI
	inviteCode := row.InviteCode
	codeVerifier := row.CodeVerifier
	loginHint := row.LoginHint
	deviceFingerprint := row.DeviceFingerprint
	// Consume: delete so the state can only be used once
	if err := db.GORM.WithContext(ctx).Where("state = ?", state).Delete(&OAuthState{}).Error; err != nil {
		return false, "", "", "", "", "", err
	}
	return true, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint, nil
}

func (db *PostgresDB) DeleteExpiredOAuthStates() (int64, error) {
	result := db.GORM.Where("expires_at < ?", time.Now()).Delete(&OAuthState{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// Pending username change operations
func (db *PostgresDB) CreatePendingUsernameChange(ctx context.Context, pending *PendingUsernameChange) error {
	return db.userRepository.CreatePendingUsernameChange(ctx, pending)
}

func (db *PostgresDB) GetPendingUsernameChangeByID(ctx context.Context, id uuid.UUID) (*PendingUsernameChange, error) {
	return db.userRepository.GetPendingUsernameChangeByID(ctx, id)
}

func (db *PostgresDB) GetPendingUsernameChangeByCheckoutSession(ctx context.Context, sessionID string) (*PendingUsernameChange, error) {
	return db.userRepository.GetPendingUsernameChangeByCheckoutSession(ctx, sessionID)
}

func (db *PostgresDB) UpdatePendingUsernameChangeStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.userRepository.UpdatePendingUsernameChangeStatus(ctx, id, status)
}

func (db *PostgresDB) DeleteExpiredPendingUsernameChanges(ctx context.Context) (int64, error) {
	return db.userRepository.DeleteExpiredPendingUsernameChanges(ctx)
}

func (db *PostgresDB) ListPendingUsernameChangesForUser(ctx context.Context, userID uuid.UUID) ([]*PendingUsernameChange, error) {
	return db.userRepository.ListPendingUsernameChangesForUser(ctx, userID)
}
