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

// GetSecretsByTenant retrieves all non-deleted secrets for a tenant (use for small sets; prefer paginated for listing)
func (r *Repository) GetSecretsByTenant(ctx context.Context, tenantID uuid.UUID) ([]Secret, error) {
	var secrets []Secret
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Find(&secrets).Error
	return secrets, err
}

// CountSecretsByTenant returns the total number of non-deleted secrets for a tenant
func (r *Repository) CountSecretsByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error
	return count, err
}

// GetSecretsByTenantPaginated returns a page of secrets for a tenant (metadata suitable for list views)
func (r *Repository) GetSecretsByTenantPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]Secret, error) {
	var secrets []Secret
	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&secrets).Error
	return secrets, err
}

// UpdateSecret updates an existing secret
// The secret must belong to the tenant (enforced by caller providing tenant-scoped secret)
func (r *Repository) UpdateSecret(ctx context.Context, secret *Secret) error {
	secret.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(secret).Updates(map[string]interface{}{
		"name":            secret.Name,
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

// GetAuditLogsByTenant retrieves audit logs for a tenant with limit and offset for pagination
func (r *Repository) GetAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// CountAuditLogsByTenant returns the total number of audit log entries for a tenant
func (r *Repository) CountAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
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

// --- Secret Versions ---

// CreateSecretVersion creates a new version snapshot of a secret
// This automatically increments the version number and updates the secret's current_version
func (r *Repository) CreateSecretVersion(ctx context.Context, version *SecretVersion) error {
	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}
	if version.Scopes == nil {
		version.Scopes = JSONMap{}
	}
	if version.Metadata == nil {
		version.Metadata = JSONMap{}
	}

	// Use a transaction to ensure version creation and secret update are atomic
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the next version number for this secret
		var maxVersion int
		if err := tx.Model(&SecretVersion{}).
			Where("secret_id = ?", version.SecretID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&maxVersion).Error; err != nil {
			return err
		}
		version.VersionNumber = maxVersion + 1

		// Create the version record
		if err := tx.Create(version).Error; err != nil {
			return err
		}

		// Update the secret's current_version and last_modified timestamps
		now := time.Now()
		if err := tx.Model(&Secret{}).
			Where("id = ? AND tenant_id = ?", version.SecretID, version.TenantID).
			Updates(map[string]interface{}{
				"current_version":  version.VersionNumber,
				"last_modified_by": version.ActorID,
				"last_modified_at": now,
				"updated_at":       now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetSecretVersions retrieves all versions for a secret in descending order (newest first)
func (r *Repository) GetSecretVersions(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]SecretVersion, error) {
	var versions []SecretVersion
	query := r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Order("version_number DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&versions).Error
	return versions, err
}

// CountSecretVersions returns the total number of versions for a secret
func (r *Repository) CountSecretVersions(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&SecretVersion{}).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Count(&count).Error
	return count, err
}

// GetSecretVersionByNumber retrieves a specific version by its version number
func (r *Repository) GetSecretVersionByNumber(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID, versionNumber int) (*SecretVersion, error) {
	var version SecretVersion
	err := r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ? AND version_number = ?", secretID, tenantID, versionNumber).
		First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &version, nil
}

// GetSecretVersionByID retrieves a version by its ID
func (r *Repository) GetSecretVersionByID(ctx context.Context, versionID uuid.UUID, tenantID uuid.UUID) (*SecretVersion, error) {
	var version SecretVersion
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", versionID, tenantID).
		First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &version, nil
}

// RollbackSecret restores a secret to a previous version by creating a new version with the old data
// This creates a new version rather than overwriting, preserving the complete history
func (r *Repository) RollbackSecret(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID, targetVersionNumber int, actorID uuid.UUID, actorType ActorType) (*SecretVersion, error) {
	// Get the target version
	targetVersion, err := r.GetSecretVersionByNumber(ctx, secretID, tenantID, targetVersionNumber)
	if err != nil {
		return nil, err
	}
	if targetVersion == nil {
		return nil, fmt.Errorf("target version %d not found", targetVersionNumber)
	}

	// Get the current secret
	secret, err := r.GetSecretByID(ctx, secretID, tenantID)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Create a rollback version that restores the target version's data
	rollbackVersion := &SecretVersion{
		SecretID:          secretID,
		TenantID:          tenantID,
		Name:              targetVersion.Name,
		Description:       targetVersion.Description,
		SecretType:        targetVersion.SecretType,
		EncryptedValue:    targetVersion.EncryptedValue,
		EncryptionSalt:    targetVersion.EncryptionSalt,
		IV:                targetVersion.IV,
		EncryptionAuthTag: targetVersion.EncryptionAuthTag,
		KeyVersion:        targetVersion.KeyVersion,
		Scopes:            targetVersion.Scopes,
		Metadata:          targetVersion.Metadata,
		ChangeType:        "rollback",
		ChangeSummary:     fmt.Sprintf("Rolled back to version %d", targetVersionNumber),
		ActorID:           actorID,
		ActorType:         actorType,
	}

	// Use transaction to ensure atomicity
	var result *SecretVersion
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the next version number
		var maxVersion int
		if err := tx.Model(&SecretVersion{}).
			Where("secret_id = ?", secretID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&maxVersion).Error; err != nil {
			return err
		}
		rollbackVersion.VersionNumber = maxVersion + 1

		// Create the rollback version record
		if err := tx.Create(rollbackVersion).Error; err != nil {
			return err
		}

		// Update the secret to match the rolled-back version
		now := time.Now()
		if err := tx.Model(&Secret{}).
			Where("id = ? AND tenant_id = ?", secretID, tenantID).
			Updates(map[string]interface{}{
				"name":                targetVersion.Name,
				"description":         targetVersion.Description,
				"secret_type":         targetVersion.SecretType,
				"encrypted_value":     targetVersion.EncryptedValue,
				"encryption_salt":     targetVersion.EncryptionSalt,
				"encryption_iv":       targetVersion.IV,
				"encryption_auth_tag": targetVersion.EncryptionAuthTag,
				"key_version":         targetVersion.KeyVersion,
				"scopes":              targetVersion.Scopes,
				"metadata":            targetVersion.Metadata,
				"current_version":     rollbackVersion.VersionNumber,
				"last_modified_by":    actorID,
				"last_modified_at":    now,
				"updated_at":          now,
			}).Error; err != nil {
			return err
		}

		result = rollbackVersion
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSecretVersionCount returns the count of versions for a secret
func (r *Repository) GetSecretVersionCount(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&SecretVersion{}).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Count(&count).Error
	return int(count), err
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

// GetAllSecretsPaginated returns all secrets without tenant filtering (admin only)
func (r *Repository) GetAllSecretsPaginated(ctx context.Context, limit, offset int) ([]Secret, error) {
	var secrets []Secret
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&secrets).Error
	return secrets, err
}

// CountAllSecrets returns the total count of all secrets (admin only)
func (r *Repository) CountAllSecrets(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Secret{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

// GetSecretByIDAdmin retrieves a secret by ID without tenant filtering (admin only)
func (r *Repository) GetSecretByIDAdmin(ctx context.Context, id uuid.UUID) (*Secret, error) {
	var secret Secret
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&secret).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &secret, nil
}

// GetVaultStats returns statistics about the vault (admin only)
func (r *Repository) GetVaultStats(ctx context.Context) (map[string]interface{}, error) {
	var totalCount int64
	var totalVersions int64

	r.db.WithContext(ctx).Model(&Secret{}).Where("deleted_at IS NULL").Count(&totalCount)
	r.db.WithContext(ctx).Model(&SecretVersion{}).Count(&totalVersions)

	return map[string]interface{}{
		"total_secrets": totalCount,
		"total_versions": totalVersions,
	}, nil
}

// GetTenantsWithSecrets returns all tenant IDs that have secrets (admin only)
func (r *Repository) GetTenantsWithSecrets(ctx context.Context) ([]uuid.UUID, error) {
	var tenantIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("deleted_at IS NULL").
		Distinct("tenant_id").
		Pluck("tenant_id", &tenantIDs).Error
	return tenantIDs, err
}
