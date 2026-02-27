package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Trigger Operations
// ============================================

// CreateTrigger creates a new state trigger
func (r *StateRepository) CreateTrigger(ctx context.Context, trigger *StateTrigger) (*StateTrigger, error) {
	if trigger.ID == uuid.Nil {
		trigger.ID = uuid.New()
	}
	trigger.CreatedAt = time.Now()
	trigger.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Create(trigger).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger: %w", err)
	}

	return trigger, nil
}

// GetTriggers retrieves triggers for a state
func (r *StateRepository) GetTriggers(ctx context.Context, stateID uuid.UUID) ([]*StateTrigger, error) {
	var triggers []*StateTrigger
	err := r.db.WithContext(ctx).Where("source_state_id = ?", stateID).Find(&triggers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get triggers: %w", err)
	}
	return triggers, nil
}

// ListTriggersByTenant retrieves all triggers for a tenant with pagination
func (r *StateRepository) ListTriggersByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*StateTrigger, int64, error) {
	var triggers []*StateTrigger
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&StateTrigger{}).Where("tenant_id = ?", tenantID).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count triggers: %w", err)
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Find(&triggers).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list triggers: %w", err)
	}

	return triggers, total, nil
}

// GetActiveTriggers retrieves all active triggers for a state and event type
func (r *StateRepository) GetActiveTriggers(ctx context.Context, stateID uuid.UUID, eventType string) ([]*StateTrigger, error) {
	var triggers []*StateTrigger
	err := r.db.WithContext(ctx).
		Where("source_state_id = ? AND trigger_type = ? AND is_active = ?", stateID, eventType, true).
		Find(&triggers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active triggers: %w", err)
	}
	return triggers, nil
}

// UpdateTrigger updates a trigger
func (r *StateRepository) UpdateTrigger(ctx context.Context, trigger *StateTrigger) (*StateTrigger, error) {
	trigger.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Save(trigger).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update trigger: %w", err)
	}

	return trigger, nil
}

// DeleteTrigger deletes a trigger
func (r *StateRepository) DeleteTrigger(ctx context.Context, triggerID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&StateTrigger{}, "id = ?", triggerID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete trigger: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("trigger not found")
	}
	return nil
}

// MarkTriggerFired updates the last triggered timestamp
func (r *StateRepository) MarkTriggerFired(ctx context.Context, triggerID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&StateTrigger{}).Where("id = ?", triggerID).Update("last_triggered_at", time.Now()).Error
}