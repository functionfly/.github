package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// Snapshot Operations
// ============================================

// CreateSnapshot creates a snapshot of current state
func (r *StateRepository) CreateSnapshot(ctx context.Context, stateID uuid.UUID, label string) (*StateSnapshot, error) {
	// Get all current values
	values, err := r.GetAllStateValues(ctx, stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get values for snapshot: %w", err)
	}

	// Get state info
	state, err := r.GetStateByID(ctx, stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Get sequence range
	var firstSeq, lastSeq int64
	r.db.WithContext(ctx).Model(&StateEvent{}).
		Select("COALESCE(MIN(sequence_num), 0)").
		Where("state_id = ?", stateID).Scan(&firstSeq)
	r.db.WithContext(ctx).Model(&StateEvent{}).
		Select("COALESCE(MAX(sequence_num), 0)").
		Where("state_id = ?", stateID).Scan(&lastSeq)

	// Build state data
	stateData := make(JSONMap)
	for _, v := range values {
		stateData[v.Key] = v.Value
	}

	snapshot := &StateSnapshot{
		ID:              uuid.New(),
		StateID:         stateID,
		SnapshotVersion: state.CurrentVersion,
		StateData:       stateData,
		KeyCount:        len(values),
		FirstSequence:   firstSeq,
		LastSequence:    lastSeq,
		CreatedAt:       time.Now(),
	}

	if label != "" {
		snapshot.Label = &label
	}

	err = r.db.WithContext(ctx).Create(snapshot).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Record snapshot event
	r.recordEvent(ctx, stateID, "snapshot", "", nil, &stateData, "system", "snapshot")

	return snapshot, nil
}

// GetSnapshot retrieves a snapshot by version
func (r *StateRepository) GetSnapshot(ctx context.Context, stateID uuid.UUID, version int) (*StateSnapshot, error) {
	var snapshot StateSnapshot
	err := r.db.WithContext(ctx).Where("state_id = ? AND snapshot_version = ?", stateID, version).First(&snapshot).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("snapshot not found")
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}
	return &snapshot, nil
}

// ListSnapshots lists all snapshots for a state
func (r *StateRepository) ListSnapshots(ctx context.Context, stateID uuid.UUID, limit, offset int) ([]*StateSnapshot, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&StateSnapshot{}).Where("state_id = ?", stateID).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count snapshots: %w", err)
	}

	var snapshots []*StateSnapshot
	err = r.db.WithContext(ctx).
		Where("state_id = ?", stateID).
		Order("snapshot_version DESC").
		Limit(limit).
		Offset(offset).
		Find(&snapshots).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list snapshots: %w", err)
	}

	return snapshots, total, nil
}

// RestoreSnapshot restores state from a snapshot
func (r *StateRepository) RestoreSnapshot(ctx context.Context, stateID uuid.UUID, snapshotVersion int, sourceType, sourceID string) error {
	snapshot, err := r.GetSnapshot(ctx, stateID, snapshotVersion)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Delete all current values
	if err := tx.Where("state_id = ?", stateID).Delete(&StateValue{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete current values: %w", err)
	}

	// Restore values from snapshot
	for key, value := range snapshot.StateData {
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		newValue := &StateValue{
			ID:        uuid.New(),
			StateID:   stateID,
			Key:       key,
			Value:     JSONMap(valueMap),
			Version:   snapshotVersion,
			CreatedBy: sourceID,
			CreatedAt: time.Now(),
		}

		if err := tx.Create(newValue).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to restore value: %w", err)
		}
	}

	// Update state version
	if err := tx.Model(&State{}).Where("id = ?", stateID).Updates(map[string]interface{}{
		"current_version": snapshotVersion,
		"updated_at":      time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update state version: %w", err)
	}

	// Record restore event
	r.recordEventTX(tx, stateID, "restore", "", nil, &snapshot.StateData, sourceType, sourceID)

	return tx.Commit().Error
}