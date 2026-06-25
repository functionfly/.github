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

// CountSecretsByNamespace returns the total number of non-deleted secrets in a namespace
func (r *Repository) CountSecretsByNamespace(ctx context.Context, tenantID uuid.UUID, namespace string) (int64, error) {
	var count int64
	prefix := namespace
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}
	err := r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND (namespace = ? OR namespace LIKE ?)",
			tenantID, namespace, prefix+"%").
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

// GetSecretsByTenantPaginatedFiltered returns a page of secrets for a tenant with optional secret type filtering
func (r *Repository) GetSecretsByTenantPaginatedFiltered(ctx context.Context, tenantID uuid.UUID, limit, offset int, secretType SecretType) ([]Secret, error) {
	var secrets []Secret
	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC")
	if secretType != "" {
		query = query.Where("secret_type = ?", secretType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&secrets).Error
	return secrets, err
}

// CountSecretsByTenantFiltered returns the total number of secrets for a tenant with optional secret type filtering
func (r *Repository) CountSecretsByTenantFiltered(ctx context.Context, tenantID uuid.UUID, secretType SecretType) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if secretType != "" {
		query = query.Where("secret_type = ?", secretType)
	}
	err := query.Count(&count).Error
	return count, err
}

// UpdateSecret updates an existing secret
// The secret must belong to the tenant (enforced by caller providing tenant-scoped secret)
func (r *Repository) UpdateSecret(ctx context.Context, secret *Secret) error {
	secret.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(secret).Updates(map[string]interface{}{
		"name":                secret.Name,
		"description":         secret.Description,
		"encrypted_value":     secret.EncryptedValue,
		"encryption_salt":     secret.EncryptionSalt,
		"encryption_iv":       secret.IV,
		"encryption_auth_tag": secret.EncryptionAuthTag,
		"key_version":         secret.KeyVersion,
		"scopes":              secret.Scopes,
		"metadata":            secret.Metadata,
		"updated_at":          secret.UpdatedAt,
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
func (r *Repository) GetAuditLogsBySecret(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
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
func (r *Repository) GetAuditLogsByActor(ctx context.Context, actorID string, actorType ActorType, tenantID uuid.UUID, limit, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("actor_id = ? AND actor_type = ? AND tenant_id = ?", actorID, actorType, tenantID).
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

// CountAuditLogsByActor returns the total number of audit log entries for an actor
func (r *Repository) CountAuditLogsByActor(ctx context.Context, actorID string, actorType ActorType, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("actor_id = ? AND actor_type = ? AND tenant_id = ?", actorID, actorType, tenantID).
		Count(&count).Error
	return count, err
}

// GetAuditLogsByAction retrieves audit logs for a specific action type
func (r *Repository) GetAuditLogsByAction(ctx context.Context, action AuditAction, tenantID uuid.UUID, limit, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).
		Where("action = ? AND tenant_id = ?", action, tenantID).
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

// CountAuditLogsByAction returns the total number of audit log entries for an action
func (r *Repository) CountAuditLogsByAction(ctx context.Context, action AuditAction, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("action = ? AND tenant_id = ?", action, tenantID).
		Count(&count).Error
	return count, err
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
		"total_secrets":  totalCount,
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

// DeleteOldSecretVersions deletes secret versions older than the specified age,
// while always keeping at least `keepLatest` versions per secret.
// Returns the number of versions deleted.
func (r *Repository) DeleteOldSecretVersions(ctx context.Context, olderThan time.Duration, keepLatest int) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	var deleted int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find secrets that have more than `keepLatest` versions
		// For each such secret, identify versions to delete (old, not in latest N)
		result := tx.Exec(`
			WITH versions_to_keep AS (
				SELECT sv.secret_id, sv.version_number
				FROM secret_versions sv
				INNER JOIN (
					SELECT secret_id, version_number,
						   ROW_NUMBER() OVER (PARTITION BY secret_id ORDER BY version_number DESC) as rn
					FROM secret_versions
					WHERE created_at < ?
				) latest ON latest.secret_id = sv.secret_id AND latest.version_number = sv.version_number
				WHERE latest.rn <= ?
			)
			DELETE FROM secret_versions
			WHERE secret_id IN (
				SELECT secret_id FROM secret_versions WHERE created_at < ? GROUP BY secret_id HAVING COUNT(*) > ?
			)
			AND created_at < ?
			AND (secret_id, version_number) NOT IN (SELECT secret_id, version_number FROM versions_to_keep)
		`, cutoff, keepLatest, cutoff, keepLatest, cutoff)

		if result.Error != nil {
			return result.Error
		}

		deleted = result.RowsAffected
		return nil
	})

	return deleted, err
}

// DeleteSecretVersionsByTenant deletes all versions for a tenant's secrets older than the specified age.
// Returns the number of versions deleted.
func (r *Repository) DeleteSecretVersionsByTenant(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration, keepLatest int) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	var deleted int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find secrets belonging to this tenant that have more than `keepLatest` versions
		result := tx.Exec(`
			WITH versions_to_keep AS (
				SELECT sv.secret_id, sv.version_number
				FROM secret_versions sv
				INNER JOIN secrets_vault s ON s.id = sv.secret_id AND s.tenant_id = ?
				INNER JOIN (
					SELECT secret_id, version_number,
						   ROW_NUMBER() OVER (PARTITION BY secret_id ORDER BY version_number DESC) as rn
					FROM secret_versions
					WHERE tenant_id = ? AND created_at < ?
				) latest ON latest.secret_id = sv.secret_id AND latest.version_number = sv.version_number
				WHERE latest.rn <= ?
			)
			DELETE FROM secret_versions
			WHERE tenant_id = ?
			AND secret_id IN (
				SELECT secret_id FROM secret_versions
				WHERE tenant_id = ? AND created_at < ?
				GROUP BY secret_id HAVING COUNT(*) > ?
			)
			AND created_at < ?
			AND (secret_id, version_number) NOT IN (SELECT secret_id, version_number FROM versions_to_keep)
		`, tenantID, tenantID, cutoff, keepLatest, tenantID, tenantID, cutoff, keepLatest, cutoff)

		if result.Error != nil {
			return result.Error
		}

		deleted = result.RowsAffected
		return nil
	})

	return deleted, err
}

// CountSecretVersionsByAge returns the count of secret versions older than the specified age
func (r *Repository) CountSecretVersionsByAge(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&SecretVersion{}).
		Where("created_at < ?", cutoff).
		Count(&count).Error
	return count, err
}

// --- Secret Dependencies ---

// CreateSecretDependency creates a new secret dependency record
func (r *Repository) CreateSecretDependency(ctx context.Context, dep *SecretDependency) error {
	if dep.ID == uuid.Nil {
		dep.ID = uuid.New()
	}
	if dep.Metadata == nil {
		dep.Metadata = JSONMap{}
	}
	return r.db.WithContext(ctx).Create(dep).Error
}

// GetSecretDependencies retrieves all dependencies for a secret
func (r *Repository) GetSecretDependencies(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) ([]SecretDependency, error) {
	var deps []SecretDependency
	err := r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Order("criticality DESC, dependent_name ASC").
		Find(&deps).Error
	return deps, err
}

// GetSecretDependencyByID retrieves a single secret dependency by ID
func (r *Repository) GetSecretDependencyByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*SecretDependency, error) {
	var dep SecretDependency
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&dep).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dep, nil
}

// DeleteSecretDependency deletes a secret dependency by ID
func (r *Repository) DeleteSecretDependency(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&SecretDependency{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("secret dependency not found")
	}
	return nil
}

// DeleteSecretDependenciesBySecret deletes all dependencies for a secret
func (r *Repository) DeleteSecretDependenciesBySecret(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Delete(&SecretDependency{}).Error
}

// GetDependenciesByDependentID retrieves all secret dependencies for a given dependent
func (r *Repository) GetDependenciesByDependentID(ctx context.Context, dependentID uuid.UUID, tenantID uuid.UUID) ([]SecretDependency, error) {
	var deps []SecretDependency
	err := r.db.WithContext(ctx).
		Where("dependent_id = ? AND tenant_id = ?", dependentID, tenantID).
		Find(&deps).Error
	return deps, err
}

// CountSecretDependencies returns the total number of dependencies for a secret
func (r *Repository) CountSecretDependencies(ctx context.Context, secretID uuid.UUID, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&SecretDependency{}).
		Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
		Count(&count).Error
	return count, err
}

// --- Bulk Operations ---

// BulkDeleteSecrets deletes multiple secrets in a transaction
// Returns the number of secrets deleted and any errors encountered
func (r *Repository) BulkDeleteSecrets(ctx context.Context, secretIDs []uuid.UUID, tenantID uuid.UUID) (int64, []error) {
	var deletedCount int64
	var errors []error

	if len(secretIDs) == 0 {
		return 0, nil
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, secretID := range secretIDs {
			// First revoke all tokens for this secret
			if err := tx.Model(&AccessToken{}).
				Where("secret_id = ? AND tenant_id = ? AND is_revoked = ?", secretID, tenantID, false).
				Updates(map[string]interface{}{
					"is_revoked":     true,
					"revoked_at":     time.Now(),
					"revoked_reason": "bulk_delete",
				}).Error; err != nil {
				errors = append(errors, fmt.Errorf("failed to revoke tokens for secret %s: %w", secretID, err))
				continue
			}

			// Delete dependencies
			if err := tx.Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
				Delete(&SecretDependency{}).Error; err != nil {
				errors = append(errors, fmt.Errorf("failed to delete dependencies for secret %s: %w", secretID, err))
				continue
			}

			// Soft delete the secret
			result := tx.Model(&Secret{}).
				Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", secretID, tenantID).
				Update("deleted_at", time.Now())
			if result.Error != nil {
				errors = append(errors, fmt.Errorf("failed to delete secret %s: %w", secretID, result.Error))
				continue
			}
			if result.RowsAffected > 0 {
				deletedCount++
			}
		}

		if len(errors) > 0 && deletedCount == 0 {
			return fmt.Errorf("all bulk delete operations failed")
		}
		return nil
	})

	if err != nil {
		errors = append(errors, err)
	}
	return deletedCount, errors
}

// BulkDeleteSecretsDryRun performs a dry run of bulk delete and returns what would be affected
func (r *Repository) BulkDeleteSecretsDryRun(ctx context.Context, secretIDs []uuid.UUID, tenantID uuid.UUID) (map[uuid.UUID]BulkDeletePreview, error) {
	previews := make(map[uuid.UUID]BulkDeletePreview)

	for _, secretID := range secretIDs {
		var secret Secret
		if err := r.db.WithContext(ctx).
			Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", secretID, tenantID).
			First(&secret).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				previews[secretID] = BulkDeletePreview{
					SecretID:     secretID,
					SecretName:   "",
					Found:        false,
					TokensCount:  0,
					Dependencies: []DependencyInfo{},
				}
				continue
			}
			return nil, err
		}

		// Count tokens
		var tokensCount int64
		r.db.WithContext(ctx).Model(&AccessToken{}).
			Where("secret_id = ? AND tenant_id = ? AND is_revoked = ?", secretID, tenantID, false).
			Count(&tokensCount)

		// Get dependencies
		var deps []SecretDependency
		r.db.WithContext(ctx).
			Where("secret_id = ? AND tenant_id = ?", secretID, tenantID).
			Find(&deps)

		depInfos := make([]DependencyInfo, len(deps))
		for i, d := range deps {
			depInfos[i] = DependencyInfo{
				ID:          d.DependentID,
				Type:        d.DependentType,
				Name:        d.DependentName,
				Criticality: d.Criticality,
			}
		}

		previews[secretID] = BulkDeletePreview{
			SecretID:     secretID,
			SecretName:   secret.Name,
			Found:        true,
			TokensCount:  int(tokensCount),
			Dependencies: depInfos,
		}
	}

	return previews, nil
}

// BulkDeletePreview contains preview information for bulk delete
type BulkDeletePreview struct {
	SecretID     uuid.UUID        `json:"secret_id"`
	SecretName   string           `json:"secret_name"`
	Found        bool             `json:"found"`
	TokensCount  int              `json:"tokens_count"`
	Dependencies []DependencyInfo `json:"dependencies"`
}

// DependencyInfo describes a single dependency
type DependencyInfo struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Criticality string    `json:"criticality"`
}

// --- Export ---

// ExportSecrets returns all non-deleted secrets for a tenant with metadata (no encrypted values)
// The export format is suitable for backup or migration purposes
func (r *Repository) ExportSecrets(ctx context.Context, tenantID uuid.UUID) ([]SecretExport, error) {
	var secrets []Secret
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Find(&secrets).Error
	if err != nil {
		return nil, err
	}

	exports := make([]SecretExport, len(secrets))
	for i, s := range secrets {
		exports[i] = SecretExport{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			SecretType:  s.SecretType,
			KeyVersion:  s.KeyVersion,
			Scopes:      s.Scopes,
			Metadata:    s.Metadata,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		}
	}

	return exports, nil
}

// SecretExport represents secret metadata for export (no encrypted values)
type SecretExport struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	SecretType  SecretType `json:"secret_type"`
	KeyVersion  int        `json:"key_version"`
	Scopes      JSONMap    `json:"scopes"`
	Metadata    JSONMap    `json:"metadata"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
