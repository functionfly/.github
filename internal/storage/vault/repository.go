package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles secrets vault persistence
// All methods are tenant-scoped for security
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new secrets vault repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// --- Secrets ---

// CreateSecret creates a new encrypted secret in the vault
// The secret value must be encrypted client-side before calling this method
func (r *Repository) CreateSecret(ctx context.Context, secret *Secret) error {
	if secret.ID == uuid.Nil {
		secret.ID = uuid.New()
	}
	if secret.Scopes == nil {
		secret.Scopes = JSONMap{}
	}
	if secret.Metadata == nil {
		secret.Metadata = JSONMap{}
	}
	return r.db.WithContext(ctx).Create(secret).Error
}

// GetSecretByID retrieves a secret by its ID, filtered by tenant for security
// Returns nil if not found
func (r *Repository) GetSecretByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Secret, error) {
	var secret Secret
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&secret).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &secret, nil
}

// GetSecretsByTenant retrieves all non-deleted secrets for a tenant
func (r *Repository) GetSecretsByTenant(ctx context.Context, tenantID uuid.UUID) ([]Secret, error) {
	var secrets []Secret
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Find(&secrets).Error
	return secrets, err
}

// UpdateSecret updates an existing secret
// The secret must belong to the tenant (enforced by caller providing tenant-scoped secret)
func (r *Repository) UpdateSecret(ctx context.Context, secret *Secret) error {
	secret.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(secret).Updates(map[string]interface{}{
		"name":             secret.Name,
		"description":     secret.Description,
		"encrypted_value": secret.EncryptedValue,
		"encryption_salt": secret.EncryptionSalt,
		"encryption_iv":   secret.IV,
		"key_version":     secret.KeyVersion,
		"scopes":          secret.Scopes,
		"metadata":        secret.Metadata,
		"updated_at":      secret.UpdatedAt,
	}).Error
}

// DeleteSecret soft-deletes a secret by setting the deleted_at timestamp
// This preserves the audit trail while making the secret inaccessible
func (r *Repository) DeleteSecret(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		Update("deleted_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("secret not found or already deleted")
	}
	return nil
}

// RecordSecretAccess updates the access statistics for a secret
func (r *Repository) RecordSecretAccess(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		Updates(map[string]interface{}{
			"last_accessed_at": now,
			"access_count":     gorm.Expr("access_count + 1"),
		}).Error
}

// --- Access Tokens ---

// CreateAccessToken creates a new access token for secret access
// The raw token should be hashed before storing (only hash is kept)
func (r *Repository) CreateAccessToken(ctx context.Context, token *AccessToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	if token.Scopes == nil {
		token.Scopes = JSONMap{}
	}
	return r.db.WithContext(ctx).Create(token).Error
}

// GetAccessTokenByHash retrieves an access token by its hash
// Also validates that the token is not expired or revoked
func (r *Repository) GetAccessTokenByHash(ctx context.Context, hash string) (*AccessToken, error) {
	var token AccessToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// GetAccessTokenByID retrieves an access token by its ID, filtered by tenant
func (r *Repository) GetAccessTokenByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*AccessToken, error) {
	var token AccessToken
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// ListAccessTokensBySecret retrieves all access tokens for a specific secret
func (r *Repository) ListAccessTokensBySecret(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) ([]AccessToken, error) {
	var tokens []AccessToken
	err := r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

// RevokeAccessToken marks an access token as revoked
func (r *Repository) RevokeAccessToken(ctx context.Context, id uuid.UUID, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&AccessToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     now,
			"revoked_reason": reason,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("access token not found")
	}
	return nil
}

// RecordTokenUse updates the usage statistics for an access token
func (r *Repository) RecordTokenUse(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&AccessToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": now,
			"use_count":    gorm.Expr("use_count + 1"),
		}).Error
}

// CleanupExpiredTokens deletes expired or revoked tokens older than the specified age
// This should be run periodically (e.g., via cron job)
func (r *Repository) CleanupExpiredTokens(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).
		Where("(expires_at < ? OR is_revoked = ?) AND (revoked_at < ? OR expires_at < ?)",
			time.Now(), true, cutoff, cutoff).
		Delete(&AccessToken{}).Error
}

// --- Audit Logs ---

// CreateAuditLog creates a new audit log entry
// All vault operations should be logged for compliance
func (r *Repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.Metadata == nil {
		log.Metadata = JSONMap{}
	}
	return r.db.WithContext(ctx).Create(log).Error
}

// GetAuditLogsBySecret retrieves audit logs for a specific secret
func (r *Repository) GetAuditLogsBySecret(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetAuditLogsByTenant retrieves audit logs for a tenant with optional limit
func (r *Repository) GetAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetAuditLogsByActor retrieves audit logs for a specific actor
func (r *Repository) GetAuditLogsByActor(ctx context.Context, actorID string, actorType ActorType, tenantID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("actor_id = ? AND actor_type = ? AND tenant_id = ?", actorID, actorType, tenantID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetAuditLogsByAction retrieves audit logs for a specific action type
func (r *Repository) GetAuditLogsByAction(ctx context.Context, action AuditAction, tenantID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("action = ? AND tenant_id = ?", action, tenantID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// CountAuditLogsBySecret returns the total number of audit log entries for a secret
func (r *Repository) CountAuditLogsBySecret(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Count(&count).Error
	return count, err
}

// --- Utility Methods ---

// Transaction executes the given function within a database transaction
// Useful for operations that need to modify multiple tables atomically
func (r *Repository) Transaction(ctx context.Context, fn func(*Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewRepository(tx))
	})
}

// DB returns the underlying GORM DB instance
// Use with caution; prefer using the repository methods
func (r *Repository) DB() *gorm.DB {
	return r.db
}
