package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Permission Operations
// ============================================

// GrantPermission grants access to a state
func (r *StateRepository) GrantPermission(ctx context.Context, perm *StatePermission) (*StatePermission, error) {
	if perm.ID == uuid.Nil {
		perm.ID = uuid.New()
	}
	perm.CreatedAt = time.Now()
	perm.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Create(perm).Error
	if err != nil {
		return nil, fmt.Errorf("failed to grant permission: %w", err)
	}

	return perm, nil
}

// GetPermissions retrieves all permissions for a state
func (r *StateRepository) GetPermissions(ctx context.Context, stateID uuid.UUID) ([]*StatePermission, error) {
	var permissions []*StatePermission
	err := r.db.WithContext(ctx).Where("state_id = ?", stateID).Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return permissions, nil
}

// CheckPermission checks if a principal has the required permission
func (r *StateRepository) CheckPermission(ctx context.Context, stateID uuid.UUID, principalType string, principalID uuid.UUID, requiredPermission string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&StatePermission{}).
		Where("state_id = ? AND principal_type = ? AND principal_id = ?", stateID, principalType, principalID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	if count > 0 {
		var hasPermission bool
		err = r.db.WithContext(ctx).
			Model(&StatePermission{}).
			Select(requiredPermission).
			Where("state_id = ? AND principal_type = ? AND principal_id = ?", stateID, principalType, principalID).
			Scan(&hasPermission).Error
		return hasPermission, err
	}

	// Check if state is public
	var state State
	err = r.db.WithContext(ctx).First(&state, "id = ?", stateID).Error
	if err != nil {
		return false, nil
	}
	return state.IsPublic && requiredPermission == "can_read", nil
}

// RevokePermission revokes access from a state
func (r *StateRepository) RevokePermission(ctx context.Context, permissionID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&StatePermission{}, "id = ?", permissionID)
	if result.Error != nil {
		return fmt.Errorf("failed to revoke permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}
	return nil
}