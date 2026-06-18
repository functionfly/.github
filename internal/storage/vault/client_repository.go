package vault

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- vault_tenant_keys ---

// UpsertTenantKey creates or updates the current user's wrapped DEK
// for a tenant. Used at first vault unlock and on rotation.
func (r *Repository) UpsertTenantKey(ctx context.Context, k *VaultTenantKey) error {
	if k == nil {
		return errors.New("nil tenant key")
	}
	now := time.Now()
	k.RotatedAt = &now
	return r.db.WithContext(ctx).Save(k).Error
}

// GetTenantKey returns the current user's wrapped DEK, or nil if the
// user has not yet set one up.
func (r *Repository) GetTenantKey(ctx context.Context, tenantID, userID uuid.UUID) (*VaultTenantKey, error) {
	var k VaultTenantKey
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&k).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

// DeleteTenantKey removes the wrapped DEK for a (tenant, user). The
// caller is responsible for ensuring all client-mode rows for that
// user are unwrapped first; this is a destructive operation.
func (r *Repository) DeleteTenantKey(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&VaultTenantKey{}).Error
}

// --- dynamic_wrapped_access_tokens ---

// CreateDynamicToken inserts a new dynamic credential access token.
func (r *Repository) CreateDynamicToken(ctx context.Context, t *DynamicWrappedAccessToken) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Scopes == nil {
		t.Scopes = JSONMap{"scopes": []string{}}
	}
	if t.AllowedIPs == nil {
		t.AllowedIPs = StringArray{}
	}
	if t.DeniedIPs == nil {
		t.DeniedIPs = StringArray{}
	}
	return r.db.WithContext(ctx).Create(t).Error
}

// GetDynamicTokenByHash returns the token row for a raw token's
// SHA-256 hash. Used by the bearer auth middleware to look up
// ff_dyn_<token> bearer tokens.
func (r *Repository) GetDynamicTokenByHash(ctx context.Context, hash string) (*DynamicWrappedAccessToken, error) {
	var t DynamicWrappedAccessToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetDynamicTokenByID returns a single token by ID, tenant-scoped.
func (r *Repository) GetDynamicTokenByID(ctx context.Context, id, tenantID uuid.UUID) (*DynamicWrappedAccessToken, error) {
	var t DynamicWrappedAccessToken
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// ListDynamicTokens lists all tokens for a tenant (hash-only; raw
// values are not stored).
func (r *Repository) ListDynamicTokens(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]DynamicWrappedAccessToken, error) {
	var ts []DynamicWrappedAccessToken
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return ts, q.Find(&ts).Error
}

// ListDynamicTokensByCredential lists tokens for a single credential.
func (r *Repository) ListDynamicTokensByCredential(ctx context.Context, tenantID, credentialID uuid.UUID) ([]DynamicWrappedAccessToken, error) {
	var ts []DynamicWrappedAccessToken
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND credential_id = ?", tenantID, credentialID).
		Order("created_at DESC").
		Find(&ts).Error
	return ts, err
}

// RevokeDynamicToken marks a token as revoked.
func (r *Repository) RevokeDynamicToken(ctx context.Context, id, tenantID uuid.UUID, reason string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&DynamicWrappedAccessToken{}).
		Where("id = ? AND tenant_id = ? AND is_revoked = false", id, tenantID).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     now,
			"revoked_reason": reason,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("token not found or already revoked")
	}
	return nil
}

// RecordDynamicTokenUse updates last_used_at + use_count.
func (r *Repository) RecordDynamicTokenUse(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&DynamicWrappedAccessToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": now,
			"use_count":    gorm.Expr("use_count + 1"),
		}).Error
}

// --- dynamic_target_shares (v2 stub) ---

// CreateDynamicTargetShare inserts a share row.
func (r *Repository) CreateDynamicTargetShare(ctx context.Context, s *DynamicTargetShare) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(s).Error
}

// ListDynamicTargetShares lists active (non-revoked) shares for a
// grantee tenant (inbound shares granted TO this tenant).
func (r *Repository) ListDynamicTargetShares(ctx context.Context, granteeTenantID uuid.UUID) ([]DynamicTargetShare, error) {
	var ss []DynamicTargetShare
	err := r.db.WithContext(ctx).
		Where("granted_to_tenant_id = ? AND revoked_at IS NULL", granteeTenantID).
		Order("created_at DESC").
		Find(&ss).Error
	return ss, err
}

// ListTargetSharesBySource lists active (non-revoked) outbound shares for
// a specific target that belong to the source tenant.
func (r *Repository) ListTargetSharesBySource(ctx context.Context, targetID, sourceTenantID uuid.UUID) ([]DynamicTargetShare, error) {
	var ss []DynamicTargetShare
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND source_tenant_id = ? AND revoked_at IS NULL", targetID, sourceTenantID).
		Order("created_at DESC").
		Find(&ss).Error
	return ss, err
}

// GetDynamicTargetShare fetches a single share by ID.
func (r *Repository) GetDynamicTargetShare(ctx context.Context, id uuid.UUID) (*DynamicTargetShare, error) {
	var s DynamicTargetShare
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// RevokeDynamicTargetShare marks a share as revoked.
func (r *Repository) RevokeDynamicTargetShare(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&DynamicTargetShare{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(map[string]interface{}{
			"revoked_at": now,
			"revoked_by": revokedBy,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("share not found or already revoked")
	}
	return nil
}
