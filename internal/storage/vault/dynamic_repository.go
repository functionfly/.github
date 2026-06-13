package vault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Dynamic secret targets ---

// CreateTarget inserts a new dynamic secret target.
func (r *Repository) CreateTarget(ctx context.Context, t *DynamicSecretTarget) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.AllowedRoles == nil {
		t.AllowedRoles = StringArray{}
	}
	return r.db.WithContext(ctx).Create(t).Error
}

// GetTarget fetches a target by ID, tenant-scoped.
func (r *Repository) GetTarget(ctx context.Context, id, tenantID uuid.UUID) (*DynamicSecretTarget, error) {
	var t DynamicSecretTarget
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// ListTargets lists active targets for a tenant.
func (r *Repository) ListTargets(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]DynamicSecretTarget, error) {
	var ts []DynamicSecretTarget
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return ts, q.Find(&ts).Error
}

// DeleteTarget soft-removes a target by setting status=disabled.
// We don't hard-delete because audit + leases reference the row.
func (r *Repository) DeleteTarget(ctx context.Context, id, tenantID uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&DynamicSecretTarget{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{"status": "disabled", "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("target not found")
	}
	return nil
}

// MarkTargetError stores the last connection / generation error.
func (r *Repository) MarkTargetError(ctx context.Context, id uuid.UUID, errMsg string) {
	now := time.Now()
	r.db.WithContext(ctx).Model(&DynamicSecretTarget{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"last_error": errMsg, "updated_at": now})
}

// MarkTargetUsed updates last_used_at / clears last_error.
func (r *Repository) MarkTargetUsed(ctx context.Context, id uuid.UUID) {
	now := time.Now()
	r.db.WithContext(ctx).Model(&DynamicSecretTarget{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"last_used_at": now, "last_error": "", "updated_at": now})
}

// --- Dynamic credentials (templates) ---

// CreateCredential inserts a new credential template.
func (r *Repository) CreateCredential(ctx context.Context, c *DynamicCredential) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(c).Error
}

// GetCredential fetches a credential by ID, tenant-scoped.
func (r *Repository) GetCredential(ctx context.Context, id, tenantID uuid.UUID) (*DynamicCredential, error) {
	var c DynamicCredential
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// ListCredentials lists credential templates for a tenant.
func (r *Repository) ListCredentials(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]DynamicCredential, error) {
	var cs []DynamicCredential
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return cs, q.Find(&cs).Error
}

// DeleteCredential marks a credential as disabled.
func (r *Repository) DeleteCredential(ctx context.Context, id, tenantID uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&DynamicCredential{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{"status": "disabled", "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}
	return nil
}

// --- Leases ---

// CreateLease inserts a new lease row.
func (r *Repository) CreateLease(ctx context.Context, l *DynamicCredentialLease) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(l).Error
}

// GetLeaseByLeaseID fetches a lease by its public lease_id string.
func (r *Repository) GetLeaseByLeaseID(ctx context.Context, leaseID string) (*DynamicCredentialLease, error) {
	var l DynamicCredentialLease
	err := r.db.WithContext(ctx).Where("lease_id = ?", leaseID).First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

// GetActiveLeaseByUsername finds a non-revoked lease for a (target,
// username) pair — used by the worker to avoid dropping a user whose
// lease has been renewed.
func (r *Repository) GetActiveLeaseByUsername(ctx context.Context, targetID uuid.UUID, dbUsername string) (*DynamicCredentialLease, error) {
	var l DynamicCredentialLease
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND db_username = ? AND revoked_at IS NULL", targetID, dbUsername).
		Order("expires_at DESC").
		First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

// ListLeasesByCredential lists leases for a credential template.
func (r *Repository) ListLeasesByCredential(ctx context.Context, credentialID uuid.UUID, limit int) ([]DynamicCredentialLease, error) {
	var ls []DynamicCredentialLease
	q := r.db.WithContext(ctx).Where("credential_id = ?", credentialID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return ls, q.Find(&ls).Error
}

// RenewLease extends a lease's expiry and updates renewed_at.
func (r *Repository) RenewLease(ctx context.Context, leaseID string, newExpiresAt time.Time) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&DynamicCredentialLease{}).
		Where("lease_id = ? AND revoked_at IS NULL", leaseID).
		Updates(map[string]interface{}{"expires_at": newExpiresAt, "renewed_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("lease not found or revoked")
	}
	return nil
}

// RevokeLease marks a lease as revoked with a reason.
func (r *Repository) RevokeLease(ctx context.Context, leaseID, reason string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&DynamicCredentialLease{}).
		Where("lease_id = ?", leaseID).
		Updates(map[string]interface{}{
			"revoked_at":        now,
			"revocation_reason": reason,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("lease not found")
	}
	return nil
}

// ListExpiredLeases returns leases past their expires_at that haven't
// been revoked. The background worker uses this to drop the underlying
// DB users.
func (r *Repository) ListExpiredLeases(ctx context.Context, limit int) ([]DynamicCredentialLease, error) {
	if limit <= 0 {
		limit = 100
	}
	var ls []DynamicCredentialLease
	err := r.db.WithContext(ctx).
		Where("revoked_at IS NULL AND expires_at <= ?", time.Now()).
		Order("expires_at ASC").
		Limit(limit).
		Find(&ls).Error
	return ls, err
}
