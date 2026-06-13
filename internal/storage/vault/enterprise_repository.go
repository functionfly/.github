package vault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// 4.3 Namespaces
// ============================================================================

// CreateNamespace inserts a new namespace.
func (r *Repository) CreateNamespace(ctx context.Context, n *VaultNamespace) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(n).Error
}

// GetNamespace fetches a namespace by ID.
func (r *Repository) GetNamespace(ctx context.Context, id, tenantID uuid.UUID) (*VaultNamespace, error) {
	var n VaultNamespace
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// GetNamespaceByPath fetches a namespace by its (tenant, path) tuple.
func (r *Repository) GetNamespaceByPath(ctx context.Context, tenantID uuid.UUID, path string) (*VaultNamespace, error) {
	var n VaultNamespace
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND path = ?", tenantID, path).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// ListNamespaces returns all namespaces for a tenant.
func (r *Repository) ListNamespaces(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]VaultNamespace, error) {
	var ns []VaultNamespace
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("path ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return ns, q.Find(&ns).Error
}

// DeleteNamespace removes a namespace. Secrets in the namespace are
// moved to the tenant default namespace.
func (r *Repository) DeleteNamespace(ctx context.Context, id, tenantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n VaultNamespace
		if err := tx.Where("id = ? AND tenant_id = ?", id, tenantID).First(&n).Error; err != nil {
			return err
		}
		if err := tx.Model(&Secret{}).
			Where("tenant_id = ? AND namespace = ?", tenantID, n.Path).
			Update("namespace", "default").Error; err != nil {
			return err
		}
		res := tx.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&VaultNamespace{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("namespace not found")
		}
		_ = now
		return nil
	})
}

// ListSecretsInNamespace returns secrets under a namespace path or any
// of its descendants.
func (r *Repository) ListSecretsInNamespace(ctx context.Context, tenantID uuid.UUID, namespace string, limit, offset int) ([]Secret, error) {
	var secrets []Secret
	prefix := namespace
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}
	q := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL AND (namespace = ? OR namespace LIKE ?)",
			tenantID, namespace, prefix+"%").
		Order("name ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return secrets, q.Find(&secrets).Error
}

// ============================================================================
// 4.1 Roles
// ============================================================================

// CreateRole inserts a new role.
func (r *Repository) CreateRole(ctx context.Context, role *VaultRole) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	if role.Permissions == nil {
		role.Permissions = JSONMap{}
	}
	return r.db.WithContext(ctx).Create(role).Error
}

// GetRole fetches a role by ID.
func (r *Repository) GetRole(ctx context.Context, id, tenantID uuid.UUID) (*VaultRole, error) {
	var role VaultRole
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// ListRoles returns all roles for a tenant.
func (r *Repository) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]VaultRole, error) {
	var rs []VaultRole
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name ASC").Find(&rs).Error
	return rs, err
}

// UpdateRole updates an existing role.
func (r *Repository) UpdateRole(ctx context.Context, role *VaultRole) error {
	role.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(role).Error
}

// DeleteRole removes a role and its assignments.
func (r *Repository) DeleteRole(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&VaultRoleAssignment{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&VaultRole{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("role not found")
		}
		return nil
	})
}

// AssignRole creates a new role assignment.
func (r *Repository) AssignRole(ctx context.Context, a *VaultRoleAssignment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(a).Error
}

// ListAssignmentsForUser returns all role assignments for a user, in
// a tenant, that haven't been deleted.
func (r *Repository) ListAssignmentsForUser(ctx context.Context, tenantID, userID uuid.UUID) ([]VaultRoleAssignment, error) {
	var as []VaultRoleAssignment
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("created_at DESC").
		Find(&as).Error
	return as, err
}

// DeleteAssignment removes a role assignment.
func (r *Repository) DeleteAssignment(ctx context.Context, id, tenantID uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&VaultRoleAssignment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("assignment not found")
	}
	return nil
}

// GetRolesForUser returns the resolved VaultRole rows for a user.
func (r *Repository) GetRolesForUser(ctx context.Context, tenantID, userID uuid.UUID) ([]VaultRole, error) {
	var roles []VaultRole
	err := r.db.WithContext(ctx).
		Table("vault_roles").
		Joins("JOIN vault_role_assignments ON vault_role_assignments.role_id = vault_roles.id").
		Where("vault_roles.tenant_id = ? AND vault_role_assignments.user_id = ?", tenantID, userID).
		Find(&roles).Error
	return roles, err
}

// ============================================================================
// 4.4 Shares
// ============================================================================

// CreateShare inserts a new cross-tenant share.
func (r *Repository) CreateShare(ctx context.Context, s *VaultShare) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(s).Error
}

// ListSharesForSecret returns active shares for a secret.
func (r *Repository) ListSharesForSecret(ctx context.Context, secretID uuid.UUID) ([]VaultShare, error) {
	var shares []VaultShare
	err := r.db.WithContext(ctx).Where("secret_id = ?", secretID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

// ListSharesForGrantee returns active shares granted to a tenant.
func (r *Repository) ListSharesForGrantee(ctx context.Context, tenantID uuid.UUID) ([]VaultShare, error) {
	var shares []VaultShare
	err := r.db.WithContext(ctx).Where("granted_to_tenant_id = ? AND revoked_at IS NULL", tenantID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

// GetShare fetches a single share by ID.
func (r *Repository) GetShare(ctx context.Context, id uuid.UUID) (*VaultShare, error) {
	var s VaultShare
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// RevokeShare marks a share as revoked.
func (r *Repository) RevokeShare(ctx context.Context, id, revokedBy uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&VaultShare{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"revoked_at": now,
			"revoked_by": revokedBy,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("share not found")
	}
	return nil
}

// GetActiveShareForGrantee returns an active share of a secret to a
// grantee tenant, or nil if none.
func (r *Repository) GetActiveShareForGrantee(ctx context.Context, secretID, granteeTenantID uuid.UUID) (*VaultShare, error) {
	var s VaultShare
	err := r.db.WithContext(ctx).
		Where("secret_id = ? AND granted_to_tenant_id = ? AND revoked_at IS NULL", secretID, granteeTenantID).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// ============================================================================
// 4.5 SSO
// ============================================================================

// GetSSOConfig returns the SSO config for a tenant, creating a default
// row on first read.
func (r *Repository) GetSSOConfig(ctx context.Context, tenantID uuid.UUID) (*VaultSSOConfig, error) {
	var s VaultSSOConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&s).Error
	if err == nil {
		return &s, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	s = VaultSSOConfig{
		TenantID:               tenantID,
		Enabled:                false,
		JITProvisioningEnabled: true,
		AttributeRoleMapping:   JSONMap{},
	}
	if err := r.db.WithContext(ctx).Create(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSSOConfig persists the SSO config.
func (r *Repository) UpdateSSOConfig(ctx context.Context, s *VaultSSOConfig) error {
	s.UpdatedAt = time.Now()
	if s.AttributeRoleMapping == nil {
		s.AttributeRoleMapping = JSONMap{}
	}
	return r.db.WithContext(ctx).Save(s).Error
}

// ============================================================================
// 4.2 SIEM webhooks
// ============================================================================

// CreateSIEMWebhook inserts a new SIEM webhook.
func (r *Repository) CreateSIEMWebhook(ctx context.Context, w *VaultSIEMWebhook) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(w).Error
}

// ListSIEMWebhooks lists webhooks for a tenant.
func (r *Repository) ListSIEMWebhooks(ctx context.Context, tenantID uuid.UUID) ([]VaultSIEMWebhook, error) {
	var ws []VaultSIEMWebhook
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&ws).Error
	return ws, err
}

// GetSIEMWebhook fetches a webhook by ID.
func (r *Repository) GetSIEMWebhook(ctx context.Context, id, tenantID uuid.UUID) (*VaultSIEMWebhook, error) {
	var w VaultSIEMWebhook
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

// DeleteSIEMWebhook removes a webhook.
func (r *Repository) DeleteSIEMWebhook(ctx context.Context, id, tenantID uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&VaultSIEMWebhook{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// MarkSIEMDelivery records the result of a webhook delivery attempt.
func (r *Repository) MarkSIEMDelivery(ctx context.Context, id uuid.UUID, status int, errMsg string) {
	now := time.Now()
	r.db.WithContext(ctx).Model(&VaultSIEMWebhook{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_delivery_at":     now,
			"last_delivery_status": status,
			"last_delivery_error":  errMsg,
		})
}
