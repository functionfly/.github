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
// If the state's IsEncrypted flag is true and encryption is enabled, the value will be encrypted.
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

	// Prepare the new value
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

	// Check if state requires encryption
	if state.IsEncrypted && r.IsEncryptionEnabled() {
		// Encrypt the value
		encryptedData, _, encErr := r.encryptJSONValue(JSONMap(value))
		if encErr != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to encrypt state value: %w", encErr)
		}
		if encryptedData != nil {
			newValue.Value = nil // Clear plaintext value
			newValue.EncryptedVal = encryptedData
			newValue.IsEncrypted = true
		}
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
// If the value is encrypted, it will be decrypted automatically.
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

	// Decrypt the value if it's encrypted
	if stateValue.IsEncrypted && len(stateValue.EncryptedVal) > 0 {
		decryptedValue, _, decErr := r.decryptJSONValue(stateValue.EncryptedVal)
		if decErr != nil {
			return nil, fmt.Errorf("failed to decrypt state value: %w", decErr)
		}
		if decryptedValue != nil {
			stateValue.Value = decryptedValue
		}
	}

	// Update last accessed
	r.db.WithContext(ctx).Model(&State{}).Where("id = ?", stateID).Update("last_accessed_at", time.Now())

	return &stateValue, nil
}

// GetAllStateValues retrieves all latest values for a state
// Encrypted values are automatically decrypted.
func (r *StateRepository) GetAllStateValues(ctx context.Context, stateID uuid.UUID) ([]*StateValue, error) {
	// Use window function for better performance than subquery with JOIN
	var values []*StateValue
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (key) id, state_id, key, value, version, previous_value,
				   content_hash, expires_at, created_by, created_at, is_encrypted, encrypted_val
			FROM state_values
			WHERE state_id = ? AND (expires_at IS NULL OR expires_at > NOW())
			ORDER BY key, version DESC
		`, stateID).
		Scan(&values).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get state values: %w", err)
	}

	// Decrypt encrypted values
	for _, v := range values {
		if v.IsEncrypted && len(v.EncryptedVal) > 0 {
			decryptedValue, _, decErr := r.decryptJSONValue(v.EncryptedVal)
			if decErr != nil {
				// Log but don't fail - just return encrypted value as-is
				continue
			}
			if decryptedValue != nil {
				v.Value = decryptedValue
			}
		}
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
// If the state's IsEncrypted flag is true and encryption is enabled, values will be encrypted.
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

	// Check if state requires encryption
	encryptValues := state.IsEncrypted && r.IsEncryptionEnabled()

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

		// Encrypt value if required
		if encryptValues {
			encryptedData, _, encErr := r.encryptJSONValue(JSONMap(value))
			if encErr != nil {
				tx.Rollback()
				return fmt.Errorf("failed to encrypt state value for key %s: %w", key, encErr)
			}
			if encryptedData != nil {
				newValue.Value = nil
				newValue.EncryptedVal = encryptedData
				newValue.IsEncrypted = true
			}
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
