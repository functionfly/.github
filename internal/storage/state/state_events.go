package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// State Event Operations
// ============================================

// recordEventTX records an event in the event log within a transaction
func (r *StateRepository) recordEventTX(tx *gorm.DB, stateID uuid.UUID, eventType string, key string, prevValue, newValue *JSONMap, sourceType, sourceID string) error {
	// Get sequence number
	var seqNum int64
	err := tx.Model(&StateEvent{}).Where("state_id = ?", stateID).Select("COALESCE(MAX(sequence_num), 0) + 1").Scan(&seqNum).Error
	if err != nil {
		return err
	}

	var k *string
	if key != "" {
		k = &key
	}

	event := &StateEvent{
		ID:            uuid.New(),
		StateID:       stateID,
		EventType:     eventType,
		Key:           k,
		SourceType:    sourceType,
		SourceID:      sourceID,
		Deterministic: false,
		SequenceNum:   seqNum,
		Timestamp:     time.Now(),
	}

	if prevValue != nil {
		event.PreviousValue = prevValue
	}
	if newValue != nil {
		event.NewValue = newValue
	}

	return tx.Create(event).Error
}

// recordEvent records an event in the event log
func (r *StateRepository) recordEvent(ctx context.Context, stateID uuid.UUID, eventType string, key string, prevValue, newValue interface{}, sourceType, sourceID string) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	var prevVal, newVal *JSONMap
	if prevValue != nil {
		pv, _ := prevValue.(*JSONMap)
		prevVal = pv
	}
	if newValue != nil {
		nv, _ := newValue.(*JSONMap)
		newVal = nv
	}

	err := r.recordEventTX(tx, stateID, eventType, key, prevVal, newVal, sourceType, sourceID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetStateHistory retrieves event history for a state
func (r *StateRepository) GetStateHistory(ctx context.Context, stateID uuid.UUID, key string, limit, offset int) ([]*StateEvent, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&StateEvent{}).Where("state_id = ?", stateID)

	if key != "" {
		query = query.Where("key = ?", key)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	var events []*StateEvent
	err = query.
		Order("sequence_num DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	return events, total, nil
}

// TimeTravelQuery queries state at a specific point in time
func (r *StateRepository) TimeTravelQuery(ctx context.Context, stateID uuid.UUID, timestamp time.Time) (map[string]interface{}, error) {
	// Use window function for better performance
	var values []StateValue
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (key) key, value
			FROM state_values
			WHERE state_id = ? AND created_at <= ?
			ORDER BY key, version DESC
		`, stateID, timestamp).
		Scan(&values).Error

	if err != nil {
		return nil, fmt.Errorf("failed to time travel: %w", err)
	}

	result := make(map[string]interface{})
	for _, v := range values {
		result[v.Key] = v.Value
	}

	return result, nil
}