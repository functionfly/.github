package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// State Value CRUD Operations
// ============================================

// SetStateValue sets a value in state (creates or updates)
func (r *StateRepository) SetStateValue(ctx context.Context, stateID uuid.UUID, key string, value map[string]interface{}, sourceType, sourceID string) (*StateValue, error) {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Get current version
	var state State
	if err := tx.First(&state, "id = ?", stateID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Get previous value for versioning
	var previousValue *JSONMap
	var stateValue StateValue
	err := tx.Where("state_id = ? AND key = ?", stateID, key).Order("version DESC").First(&stateValue).Error
	if err == nil {
		previousValue = &stateValue.Value
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get previous value: %w", err)
	}

	// Create new value
	newValue := &StateValue{
		ID:            uuid.New(),
		StateID:       stateID,
		Key:           key,
		Value:         JSONMap(value),
		Version:       state.CurrentVersion,
		PreviousValue: previousValue,
		CreatedBy:     sourceID,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(newValue).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set state value: %w", err)
	}

	// Update state version
	if err := tx.Model(&state).Updates(map[string]interface{}{
		"current_version":  gorm.Expr("current_version + 1"),
		"updated_at":       time.Now(),
		"last_accessed_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update state version: %w", err)
	}

	// Record event
	jsonMapValue := JSONMap(value)
	r.recordEventTX(tx, stateID, "set", key, previousValue, &jsonMapValue, sourceType, sourceID)

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newValue, nil
}

// GetStateValue retrieves a value from state
func (r *StateRepository) GetStateValue(ctx context.Context, stateID uuid.UUID, key string) (*StateValue, error) {
	var stateValue StateValue
	err := r.db.WithContext(ctx).
		Where("state_id = ? AND key = ?", stateID, key).
		Order("version DESC").
		First(&stateValue).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("state value not found")
		}
		return nil, fmt.Errorf("failed to get state value: %w", err)
	}

	// Update last accessed
	r.db.WithContext(ctx).Model(&State{}).Where("id = ?", stateID).Update("last_accessed_at", time.Now())

	return &stateValue, nil
}

// GetAllStateValues retrieves all latest values for a state
func (r *StateRepository) GetAllStateValues(ctx context.Context, stateID uuid.UUID) ([]*StateValue, error) {
	// Use window function for better performance than subquery with JOIN
	var values []*StateValue
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (key) id, state_id, key, value, version, previous_value,
				   content_hash, expires_at, created_by, created_at
			FROM state_values
			WHERE state_id = ? AND (expires_at IS NULL OR expires_at > NOW())
			ORDER BY key, version DESC
		`, stateID).
		Scan(&values).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get state values: %w", err)
	}

	return values, nil
}

// DeleteStateValue deletes a value from state
func (r *StateRepository) DeleteStateValue(ctx context.Context, stateID uuid.UUID, key string, sourceType, sourceID string) error {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Get previous value for event
	var previousValue *JSONMap
	var stateValue StateValue
	err := tx.Where("state_id = ? AND key = ?", stateID, key).Order("version DESC").First(&stateValue).Error
	if err == nil {
		previousValue = &stateValue.Value
	}

	// Insert deletion marker (version 0)
	delValue := JSONMap{"deleted": true}
	newValue := &StateValue{
		ID:            uuid.New(),
		StateID:       stateID,
		Key:           key,
		Value:         delValue,
		Version:       0,
		PreviousValue: previousValue,
		CreatedBy:     sourceID,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(newValue).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark deletion: %w", err)
	}

	// Record event
	r.recordEventTX(tx, stateID, "delete", key, previousValue, nil, sourceType, sourceID)

	// Update state version
	if err := tx.Model(&State{}).Where("id = ?", stateID).Update("updated_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update state: %w", err)
	}

	return tx.Commit().Error
}

// BulkSetStateValues sets multiple values in state efficiently
func (r *StateRepository) BulkSetStateValues(ctx context.Context, stateID uuid.UUID, values map[string]map[string]interface{}, sourceType, sourceID string) error {
	if len(values) == 0 {
		return nil
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Get current state version
	var state State
	if err := tx.First(&state, "id = ?", stateID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get state: %w", err)
	}

	currentVersion := state.CurrentVersion

	// Prepare batch insert
	stateValues := make([]*StateValue, 0, len(values))
	for key, value := range values {
		newValue := &StateValue{
			ID:        uuid.New(),
			StateID:   stateID,
			Key:       key,
			Value:     JSONMap(value),
			Version:   currentVersion,
			CreatedBy: sourceID,
			CreatedAt: time.Now(),
		}
		stateValues = append(stateValues, newValue)
	}

	// Bulk insert
	if err := tx.CreateInBatches(stateValues, 100).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to bulk insert state values: %w", err)
	}

	// Update state version
	if err := tx.Model(&state).Updates(map[string]interface{}{
		"current_version":  gorm.Expr("current_version + 1"),
		"updated_at":       time.Now(),
		"last_accessed_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update state version: %w", err)
	}

	return tx.Commit().Error
}