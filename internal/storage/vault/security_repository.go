package vault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Phase 1.1: Vault MFA config ---

// GetMFAConfig returns the MFA policy for a tenant, creating a default row
// on first access.
func (r *Repository) GetMFAConfig(ctx context.Context, tenantID uuid.UUID) (*VaultMFAConfig, error) {
	var cfg VaultMFAConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&cfg).Error
	if err == nil {
		return &cfg, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	cfg = VaultMFAConfig{
		TenantID:             tenantID,
		MFARequired:          false,
		MFAMethod:            "totp",
		EnforceForTokens:     false,
		EnforceForAPI:        false,
		MFASessionTTLSeconds: 900,
	}
	if err := r.db.WithContext(ctx).Create(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpdateMFAConfig updates the MFA policy for a tenant.
func (r *Repository) UpdateMFAConfig(ctx context.Context, cfg *VaultMFAConfig) error {
	cfg.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(cfg).Error
}

// --- Phase 1.2: Token IP allowlist helpers ---

// UpdateAccessTokenIPPolicy sets the IP allow/deny lists for a token.
func (r *Repository) UpdateAccessTokenIPPolicy(ctx context.Context, tokenID uuid.UUID, allowedIPs, deniedIPs []string, enabled bool) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&AccessToken{}).
		Where("id = ?", tokenID).
		Updates(map[string]interface{}{
			"allowed_ips":            StringArray(allowedIPs),
			"denied_ips":             StringArray(deniedIPs),
			"ip_restriction_enabled": enabled,
			"updated_at":             now,
		})
	return res.Error
}

// --- Phase 1.3: Secret expiration ---

// ListExpiringSoon returns secrets whose expires_at is within the next window
// but has not yet been notified, used by the background expiry sweeper.
func (r *Repository) ListExpiringSoon(ctx context.Context, window time.Duration, limit int) ([]Secret, error) {
	if limit <= 0 {
		limit = 100
	}
	cutoff := time.Now().Add(window)
	var secrets []Secret
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL AND expires_at IS NOT NULL AND expires_at <= ? AND expired_notified_at IS NULL",
			SecretStatusActive, cutoff).
		Order("expires_at ASC").
		Limit(limit).
		Find(&secrets).Error
	return secrets, err
}

// ListExpired returns secrets past their expires_at and still in the active
// state — i.e. needing to be transitioned to expired and have their tokens
// revoked by the background worker.
func (r *Repository) ListExpired(ctx context.Context, limit int) ([]Secret, error) {
	if limit <= 0 {
		limit = 100
	}
	now := time.Now()
	var secrets []Secret
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL AND expires_at IS NOT NULL AND expires_at <= ?",
			SecretStatusActive, now).
		Order("expires_at ASC").
		Limit(limit).
		Find(&secrets).Error
	return secrets, err
}

// MarkSecretExpired transitions a secret to the expired state and stamps
// the notification timestamp in a single transaction. Returns true on a
// real state change, false if the secret was already expired.
func (r *Repository) MarkSecretExpired(ctx context.Context, secretID, tenantID uuid.UUID) (bool, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&Secret{}).
		Where("id = ? AND tenant_id = ? AND status = ?", secretID, tenantID, SecretStatusActive).
		Updates(map[string]interface{}{
			"status":              SecretStatusExpired,
			"expired_notified_at": now,
			"updated_at":          now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// MarkSecretExpiringSoon flags a secret as expiring_soon for dashboard
// surfacing. The secret remains usable.
func (r *Repository) MarkSecretExpiringSoon(ctx context.Context, secretID, tenantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&Secret{}).
		Where("id = ? AND tenant_id = ? AND status = ?", secretID, tenantID, SecretStatusActive).
		Updates(map[string]interface{}{
			"status":                 SecretStatusExpiringSoon,
			"last_expiry_warning_at": now,
			"updated_at":             now,
		}).Error
}

// RevokeAllTokensForSecret revokes all non-revoked tokens for a secret.
func (r *Repository) RevokeAllTokensForSecret(ctx context.Context, secretID, tenantID uuid.UUID, reason string) (int64, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&AccessToken{}).
		Where("secret_id = ? AND tenant_id = ? AND is_revoked = ?", secretID, tenantID, false).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     now,
			"revoked_reason": reason,
		})
	return res.RowsAffected, res.Error
}

// SetSecretExpiration sets expires_at for a secret, optionally recomputing
// from expire_after_days when expires_at is not provided.
func (r *Repository) SetSecretExpiration(ctx context.Context, secretID, tenantID uuid.UUID, expiresAt *time.Time, expireAfterDays *int) error {
	updates := map[string]interface{}{"updated_at": time.Now()}
	if expiresAt != nil {
		updates["expires_at"] = expiresAt
	}
	if expireAfterDays != nil {
		updates["expire_after_days"] = *expireAfterDays
		if expiresAt == nil {
			computed := time.Now().Add(time.Duration(*expireAfterDays) * 24 * time.Hour)
			updates["expires_at"] = computed
		}
	}
	res := r.db.WithContext(ctx).Model(&Secret{}).
		Where("id = ? AND tenant_id = ?", secretID, tenantID).
		Updates(updates)
	return res.Error
}

// --- Phase 1.4: Break-glass ---

// CreateBreakGlassRequest inserts a new break-glass request.
func (r *Repository) CreateBreakGlassRequest(ctx context.Context, req *BreakGlassRequest) error {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	if req.Metadata == nil {
		req.Metadata = JSONMap{}
	}
	return r.db.WithContext(ctx).Create(req).Error
}

// GetBreakGlassConfig returns the per-tenant break-glass config or creates
// a default one.
func (r *Repository) GetBreakGlassConfig(ctx context.Context, tenantID uuid.UUID) (*BreakGlassConfig, error) {
	var cfg BreakGlassConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&cfg).Error
	if err == nil {
		return &cfg, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	cfg = BreakGlassConfig{
		TenantID:              tenantID,
		MaxDurationMinutes:    60,
		RequiredApproverCount: 1,
		ApproverUserIDs:       JSONMap{"user_ids": []string{}},
		Enabled:               true,
	}
	if err := r.db.WithContext(ctx).Create(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpdateBreakGlassConfig persists a break-glass policy.
func (r *Repository) UpdateBreakGlassConfig(ctx context.Context, cfg *BreakGlassConfig) error {
	cfg.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(cfg).Error
}

// GetBreakGlassRequest fetches a request by ID (tenant-scoped).
func (r *Repository) GetBreakGlassRequest(ctx context.Context, id, tenantID uuid.UUID) (*BreakGlassRequest, error) {
	var req BreakGlassRequest
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// ListBreakGlassRequestsByTenant returns recent requests for a tenant.
func (r *Repository) ListBreakGlassRequestsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]BreakGlassRequest, error) {
	var reqs []BreakGlassRequest
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return reqs, q.Find(&reqs).Error
}

// ApproveBreakGlassRequest marks the request approved and sets the
// approver + expiry. The active break-glass window is bound by expires_at.
func (r *Repository) ApproveBreakGlassRequest(ctx context.Context, id, tenantID, approverID uuid.UUID, durationMinutes int) error {
	now := time.Now()
	expires := now.Add(time.Duration(durationMinutes) * time.Minute)
	res := r.db.WithContext(ctx).Model(&BreakGlassRequest{}).
		Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, "pending").
		Updates(map[string]interface{}{
			"status":      "approved",
			"approved_by": approverID,
			"approved_at": now,
			"expires_at":  expires,
			"updated_at":  now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("break-glass request not found or not pending")
	}
	return nil
}

// DenyBreakGlassRequest marks a request as denied.
func (r *Repository) DenyBreakGlassRequest(ctx context.Context, id, tenantID, denierID uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&BreakGlassRequest{}).
		Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, "pending").
		Updates(map[string]interface{}{
			"status":      "denied",
			"approved_by": denierID,
			"approved_at": now,
			"updated_at":  now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("break-glass request not found or not pending")
	}
	return nil
}

// RevokeBreakGlassRequest revokes an active grant.
func (r *Repository) RevokeBreakGlassRequest(ctx context.Context, id, tenantID uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&BreakGlassRequest{}).
		Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, "approved").
		Updates(map[string]interface{}{
			"status":     "revoked",
			"revoked_at": now,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("break-glass request not found or not active")
	}
	return nil
}

// ExpireBreakGlassRequests marks approved requests whose window has
// passed as expired. Used by the background worker.
func (r *Repository) ExpireBreakGlassRequests(ctx context.Context) (int64, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&BreakGlassRequest{}).
		Where("status = ? AND expires_at <= ?", "approved", now).
		Updates(map[string]interface{}{
			"status":     "expired",
			"updated_at": now,
		})
	return res.RowsAffected, res.Error
}

// FindActiveBreakGlassRequest returns the active (approved, unexpired)
// request for a tenant, if any. There can be at most one active request
// at a time per tenant.
func (r *Repository) FindActiveBreakGlassRequest(ctx context.Context, tenantID, userID uuid.UUID) (*BreakGlassRequest, error) {
	var req BreakGlassRequest
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND requested_by = ? AND status = ? AND expires_at > ?",
			tenantID, userID, "approved", now).
		Order("approved_at DESC").
		First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// --- Phase 1.4b: Escrow ---

// GetEscrowConfig returns the per-tenant escrow config (may be nil if
// escrow is not enabled).
func (r *Repository) GetEscrowConfig(ctx context.Context, tenantID uuid.UUID) (*VaultEscrowConfig, error) {
	var cfg VaultEscrowConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// UpsertEscrowConfig enables / updates escrow for a tenant.
func (r *Repository) UpsertEscrowConfig(ctx context.Context, cfg *VaultEscrowConfig) error {
	cfg.UpdatedAt = time.Now()
	if cfg.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id required")
	}
	return r.db.WithContext(ctx).Save(cfg).Error
}

// DisableEscrow removes the escrow config for a tenant.
func (r *Repository) DisableEscrow(ctx context.Context, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&VaultEscrowConfig{}).Error
}

// ListActiveBreakGlassForUser returns the most recent active break-glass
// grant for the given user, or nil if none is currently in effect.
//
// "Active" means Status == "approved", ExpiresAt is in the future, and
// the request was not revoked. Used by the vault MFA middleware to
// short-circuit MFA checks for users with an active emergency grant.
func (r *Repository) ListActiveBreakGlassForUser(ctx context.Context, tenantID, userID uuid.UUID) (*BreakGlassRequest, error) {
	var bg BreakGlassRequest
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND requested_by = ? AND status = ? AND expires_at > ? AND revoked_at IS NULL",
			tenantID, userID, "approved", time.Now()).
		Order("expires_at DESC").
		First(&bg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bg, nil
}

// HasRecentMFAVerify reports whether the user has a recent
// AuditActionMFAVerify row within the supplied window. The vault MFA
// middleware uses this to skip re-prompting for users who verified
// within the session TTL.
func (r *Repository) HasRecentMFAVerify(ctx context.Context, tenantID, userID uuid.UUID, window time.Duration) (bool, error) {
	if window <= 0 {
		return false, nil
	}
	cutoff := time.Now().Add(-window)
	var count int64
	err := r.db.WithContext(ctx).
		Table("audit_log").
		Where("tenant_id = ? AND actor_user_id = ? AND action = ? AND created_at > ?",
			tenantID, userID, "vault.mfa.verify", cutoff).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
