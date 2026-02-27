package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// State CRUD Operations
// ============================================

// CreateState creates a new state container
func (r *StateRepository) CreateState(ctx context.Context, state *State) (*State, error) {
	if state.ID == uuid.Nil {
		state.ID = uuid.New()
	}
	if state.StorageType == "" {
		state.StorageType = "keyvalue"
	}
	state.CreatedAt = time.Now()
	state.UpdatedAt = time.Now()
	state.LastAccessedAt = time.Now()

	err := r.db.WithContext(ctx).Create(state).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create state: %w", err)
	}

	return state, nil
}

// GetStateByID retrieves a state by ID
func (r *StateRepository) GetStateByID(ctx context.Context, stateID uuid.UUID) (*State, error) {
	var state State
	err := r.db.WithContext(ctx).First(&state, "id = ?", stateID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("state not found")
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	return &state, nil
}

// GetStateByPath retrieves a state by its path (tenant/name)
func (r *StateRepository) GetStateByPath(ctx context.Context, tenantID uuid.UUID, path string) (*State, error) {
	var state State
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND full_path = ?", tenantID, path).First(&state).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("state not found")
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	return &state, nil
}

// ListStatesByTenant retrieves all states for a tenant
func (r *StateRepository) ListStatesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*State, int64, error) {
	var states []*State
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&State{}).Where("tenant_id = ?", tenantID).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count states: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&states).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list states: %w", err)
	}

	return states, total, nil
}

// UpdateState updates an existing state
func (r *StateRepository) UpdateState(ctx context.Context, state *State) (*State, error) {
	// Use Updates() for selective field updates, which is more efficient than Save()
	state.UpdatedAt = time.Now()
	state.LastAccessedAt = time.Now()

	err := r.db.WithContext(ctx).Model(state).Updates(state).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}

	return state, nil
}

// DeleteState deletes a state and all its values
func (r *StateRepository) DeleteState(ctx context.Context, stateID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&State{}, "id = ?", stateID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("state not found")
	}

	return nil
}