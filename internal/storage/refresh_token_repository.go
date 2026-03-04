package storage

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefreshTokenRepository handles refresh token-related database operations
type RefreshTokenRepository struct {
	db *PostgresDB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *PostgresDB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// CreateRefreshToken creates a new refresh token
func (r *RefreshTokenRepository) CreateRefreshToken(userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*RefreshToken, error) {
	token := &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO refresh_tokens (id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		token.ID, token.UserID, token.TokenHash, token.IPAddress, token.UserAgent, token.ExpiresAt, token.Revoked, token.CreatedAt, token.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return token, nil
}

// GetRefreshTokenByHash retrieves a refresh token by its hash
func (r *RefreshTokenRepository) GetRefreshTokenByHash(tokenHash string) (*RefreshToken, error) {
	var token RefreshToken
	var revokedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, revoked_at, created_at, updated_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked = false AND expires_at > NOW()`,
		tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.IPAddress, &token.UserAgent,
		&token.ExpiresAt, &token.Revoked, &revokedAt, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}

	return &token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *RefreshTokenRepository) RevokeRefreshToken(tokenID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE refresh_tokens
		SET revoked = true, revoked_at = NOW(), updated_at = NOW()
		WHERE id = $1`,
		tokenID)

	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeUserRefreshTokens revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeUserRefreshTokens(userID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE refresh_tokens
		SET revoked = true, revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND revoked = false`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens removes expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpiredRefreshTokens() (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR (revoked = true AND revoked_at < NOW() - INTERVAL '30 days')`)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// ListUserRefreshTokens lists all refresh tokens for a user
func (r *RefreshTokenRepository) ListUserRefreshTokens(userID uuid.UUID) ([]*RefreshToken, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, revoked, revoked_at, created_at, updated_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list user refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*RefreshToken
	for rows.Next() {
		var token RefreshToken
		var revokedAt sql.NullTime

		err := rows.Scan(
			&token.ID, &token.UserID, &token.TokenHash, &token.IPAddress, &token.UserAgent,
			&token.ExpiresAt, &token.Revoked, &revokedAt, &token.CreatedAt, &token.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}

		if revokedAt.Valid {
			token.RevokedAt = &revokedAt.Time
		}

		tokens = append(tokens, &token)
	}

	return tokens, nil
}

// HashRefreshToken creates a SHA-256 hash of a refresh token
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}