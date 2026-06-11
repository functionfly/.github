package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: sessions, refresh tokens, login attempts, auth events.

// Session operations
func (db *PostgresDB) CreateSession(ctx context.Context, userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error) {
	return db.sessionRepository.CreateSession(ctx, userID, sessionToken, ipAddress, userAgent, expiresAt)
}

func (db *PostgresDB) GetSessionByToken(ctx context.Context, sessionToken string) (*Session, error) {
	return db.sessionRepository.GetSessionByToken(ctx, sessionToken)
}

func (db *PostgresDB) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	return db.sessionRepository.GetSessionByID(ctx, sessionID)
}

func (db *PostgresDB) UpdateSessionMFAStatus(ctx context.Context, sessionToken string, mfaVerified bool) error {
	return db.sessionRepository.UpdateSessionMFAStatus(ctx, sessionToken, mfaVerified)
}

func (db *PostgresDB) UpdateSessionActivity(ctx context.Context, sessionToken string) error {
	return db.sessionRepository.UpdateSessionActivity(ctx, sessionToken)
}

func (db *PostgresDB) DeleteSession(ctx context.Context, sessionToken string) error {
	return db.sessionRepository.DeleteSession(ctx, sessionToken)
}

func (db *PostgresDB) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	return db.sessionRepository.DeleteExpiredSessions(ctx)
}

func (db *PostgresDB) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	return db.sessionRepository.DeleteUserSessions(ctx, userID)
}

func (db *PostgresDB) DeleteSessionByID(ctx context.Context, sessionID, userID uuid.UUID) error {
	return db.sessionRepository.DeleteSessionByID(ctx, sessionID, userID)
}

func (db *PostgresDB) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	return db.sessionRepository.ListUserSessions(ctx, userID)
}

func (db *PostgresDB) CountActiveUserSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	return db.sessionRepository.CountActiveUserSessions(ctx, userID)
}

func (db *PostgresDB) ListTenantSessions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Session, error) {
	return db.sessionRepository.ListTenantSessions(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) DeleteSessionByIDOnly(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	return db.sessionRepository.DeleteSessionByIDOnly(ctx, sessionID, userID)
}

// Refresh token operations
func (db *PostgresDB) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*RefreshToken, error) {
	return db.refreshTokenRepository.CreateRefreshToken(ctx, userID, tokenHash, ipAddress, userAgent, expiresAt)
}

func (db *PostgresDB) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	return db.refreshTokenRepository.GetRefreshTokenByHash(ctx, tokenHash)
}

func (db *PostgresDB) RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	return db.refreshTokenRepository.RevokeRefreshToken(ctx, tokenID)
}

func (db *PostgresDB) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return db.refreshTokenRepository.RevokeUserRefreshTokens(ctx, userID)
}

func (db *PostgresDB) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	return db.refreshTokenRepository.DeleteExpiredRefreshTokens(ctx)
}

func (db *PostgresDB) ListUserRefreshTokens(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error) {
	return db.refreshTokenRepository.ListUserRefreshTokens(ctx, userID)
}

// Login attempt operations
func (db *PostgresDB) CreateLoginAttempt(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*LoginAttempt, error) {
	return db.loginAttemptRepository.CreateLoginAttempt(ctx, userID, ipAddress, userAgent, successful, lockoutUntil)
}

func (db *PostgresDB) GetRecentFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	return db.loginAttemptRepository.GetRecentFailedLoginAttempts(ctx, userID, since)
}

func (db *PostgresDB) GetUserLockoutStatus(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	return db.loginAttemptRepository.GetUserLockoutStatus(ctx, userID)
}

func (db *PostgresDB) ClearUserLockout(ctx context.Context, userID uuid.UUID) error {
	return db.loginAttemptRepository.ClearUserLockout(ctx, userID)
}

func (db *PostgresDB) DeleteOldLoginAttempts(ctx context.Context, before time.Time) (int64, error) {
	return db.loginAttemptRepository.DeleteOldLoginAttempts(ctx, before)
}

// Auth event operations
func (db *PostgresDB) LogAuthEvent(ctx context.Context, event *AuthEvent) error {
	return db.authEventRepository.LogAuthEvent(ctx, event)
}

func (db *PostgresDB) GetAuthEventsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetAuthEventsForUser(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetAuthEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetAuthEventsByType(ctx, eventType, limit, offset)
}

func (db *PostgresDB) GetRecentAuthEvents(ctx context.Context, since time.Time, limit int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetRecentAuthEvents(ctx, since, limit)
}

func (db *PostgresDB) DeleteOldAuthEvents(ctx context.Context, before time.Time) (int64, error) {
	return db.authEventRepository.DeleteOldAuthEvents(ctx, before)
}