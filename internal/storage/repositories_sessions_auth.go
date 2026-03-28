package storage

import (
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: sessions, refresh tokens, login attempts, auth events.

// Session operations
func (db *PostgresDB) CreateSession(userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error) {
	return db.sessionRepository.CreateSession(userID, sessionToken, ipAddress, userAgent, expiresAt)
}

func (db *PostgresDB) GetSessionByToken(sessionToken string) (*Session, error) {
	return db.sessionRepository.GetSessionByToken(sessionToken)
}

func (db *PostgresDB) GetSessionByID(sessionID uuid.UUID) (*Session, error) {
	return db.sessionRepository.GetSessionByID(sessionID)
}

func (db *PostgresDB) UpdateSessionMFAStatus(sessionToken string, mfaVerified bool) error {
	return db.sessionRepository.UpdateSessionMFAStatus(sessionToken, mfaVerified)
}

func (db *PostgresDB) UpdateSessionActivity(sessionToken string) error {
	return db.sessionRepository.UpdateSessionActivity(sessionToken)
}

func (db *PostgresDB) DeleteSession(sessionToken string) error {
	return db.sessionRepository.DeleteSession(sessionToken)
}

func (db *PostgresDB) DeleteExpiredSessions() (int64, error) {
	return db.sessionRepository.DeleteExpiredSessions()
}

func (db *PostgresDB) DeleteUserSessions(userID uuid.UUID) error {
	return db.sessionRepository.DeleteUserSessions(userID)
}

func (db *PostgresDB) DeleteSessionByID(sessionID, userID uuid.UUID) error {
	return db.sessionRepository.DeleteSessionByID(sessionID, userID)
}

func (db *PostgresDB) ListUserSessions(userID uuid.UUID) ([]*Session, error) {
	return db.sessionRepository.ListUserSessions(userID)
}

func (db *PostgresDB) CountActiveUserSessions(userID uuid.UUID) (int, error) {
	return db.sessionRepository.CountActiveUserSessions(userID)
}

func (db *PostgresDB) ListTenantSessions(tenantID uuid.UUID, limit, offset int) ([]*Session, error) {
	return db.sessionRepository.ListTenantSessions(tenantID, limit, offset)
}

func (db *PostgresDB) DeleteSessionByIDOnly(sessionID uuid.UUID, userID uuid.UUID) error {
	return db.sessionRepository.DeleteSessionByIDOnly(sessionID, userID)
}

// Refresh token operations
func (db *PostgresDB) CreateRefreshToken(userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*RefreshToken, error) {
	return db.refreshTokenRepository.CreateRefreshToken(userID, tokenHash, ipAddress, userAgent, expiresAt)
}

func (db *PostgresDB) GetRefreshTokenByHash(tokenHash string) (*RefreshToken, error) {
	return db.refreshTokenRepository.GetRefreshTokenByHash(tokenHash)
}

func (db *PostgresDB) RevokeRefreshToken(tokenID uuid.UUID) error {
	return db.refreshTokenRepository.RevokeRefreshToken(tokenID)
}

func (db *PostgresDB) RevokeUserRefreshTokens(userID uuid.UUID) error {
	return db.refreshTokenRepository.RevokeUserRefreshTokens(userID)
}

func (db *PostgresDB) DeleteExpiredRefreshTokens() (int64, error) {
	return db.refreshTokenRepository.DeleteExpiredRefreshTokens()
}

func (db *PostgresDB) ListUserRefreshTokens(userID uuid.UUID) ([]*RefreshToken, error) {
	return db.refreshTokenRepository.ListUserRefreshTokens(userID)
}

// Login attempt operations
func (db *PostgresDB) CreateLoginAttempt(userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*LoginAttempt, error) {
	return db.loginAttemptRepository.CreateLoginAttempt(userID, ipAddress, userAgent, successful, lockoutUntil)
}

func (db *PostgresDB) GetRecentFailedLoginAttempts(userID uuid.UUID, since time.Time) (int, error) {
	return db.loginAttemptRepository.GetRecentFailedLoginAttempts(userID, since)
}

func (db *PostgresDB) GetUserLockoutStatus(userID uuid.UUID) (*time.Time, error) {
	return db.loginAttemptRepository.GetUserLockoutStatus(userID)
}

func (db *PostgresDB) ClearUserLockout(userID uuid.UUID) error {
	return db.loginAttemptRepository.ClearUserLockout(userID)
}

func (db *PostgresDB) DeleteOldLoginAttempts(before time.Time) (int64, error) {
	return db.loginAttemptRepository.DeleteOldLoginAttempts(before)
}

// Auth event operations
func (db *PostgresDB) LogAuthEvent(event *AuthEvent) error {
	return db.authEventRepository.LogAuthEvent(event)
}

func (db *PostgresDB) GetAuthEventsForUser(userID uuid.UUID, limit, offset int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetAuthEventsForUser(userID, limit, offset)
}

func (db *PostgresDB) GetAuthEventsByType(eventType string, limit, offset int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetAuthEventsByType(eventType, limit, offset)
}

func (db *PostgresDB) GetRecentAuthEvents(since time.Time, limit int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetRecentAuthEvents(since, limit)
}

func (db *PostgresDB) DeleteOldAuthEvents(before time.Time) (int64, error) {
	return db.authEventRepository.DeleteOldAuthEvents(before)
}
