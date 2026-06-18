package statefabric

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	statestore "github.com/functionfly/functionfly/internal/storage/state"
)

// GetFabricValue reads a key from the fabric's durable state store.
func (r *Repository) GetFabricValue(ctx context.Context, tenantID, fabricID uuid.UUID, key string) (map[string]interface{}, error) {
	if _, err := r.GetFabric(ctx, tenantID, fabricID); err != nil {
		return nil, err
	}
	sv, err := r.stateRepo.GetStateValue(ctx, fabricID, key)
	if err != nil {
		return nil, err
	}
	if sv.Value == nil {
		return nil, nil
	}
	return map[string]interface{}(sv.Value), nil
}

// SetFabricValue writes a key to the fabric's durable state store.
func (r *Repository) SetFabricValue(ctx context.Context, tenantID, fabricID uuid.UUID, key string, value map[string]interface{}, sourceID string) (*statestore.StateValue, error) {
	if _, err := r.GetFabric(ctx, tenantID, fabricID); err != nil {
		return nil, err
	}
	if sourceID == "" {
		sourceID = "edge"
	}

	oldValue, _ := r.stateRepo.GetStateValue(ctx, fabricID, key)
	var oldValPtr *statestore.JSONMap
	if oldValue != nil {
		oldVal := statestore.JSONMap(oldValue.Value)
		oldValPtr = &oldVal
	}

	result, err := r.stateRepo.SetStateValue(ctx, fabricID, key, value, "edge", sourceID)
	if err != nil {
		return nil, err
	}

	if r.triggerEngine != nil {
		newVal := statestore.JSONMap(value)
		go r.triggerEngine.ProcessStateChange(ctx, fabricID, key, "set", oldValPtr, &newVal)
	}

	return result, nil
}

// DeleteFabricValue removes a key from the fabric's durable state store.
func (r *Repository) DeleteFabricValue(ctx context.Context, tenantID, fabricID uuid.UUID, key string, sourceID string) error {
	if _, err := r.GetFabric(ctx, tenantID, fabricID); err != nil {
		return err
	}
	if sourceID == "" {
		sourceID = "edge"
	}

	oldValue, _ := r.stateRepo.GetStateValue(ctx, fabricID, key)
	var oldValPtr *statestore.JSONMap
	if oldValue != nil {
		oldVal := statestore.JSONMap(oldValue.Value)
		oldValPtr = &oldVal
	}

	err := r.stateRepo.DeleteStateValue(ctx, fabricID, key, "edge", sourceID)
	if err != nil {
		return err
	}

	if r.triggerEngine != nil {
		go r.triggerEngine.ProcessStateChange(ctx, fabricID, key, "delete", oldValPtr, nil)
	}

	return nil
}

// GetFabricValueOrNil returns nil when the key is missing instead of an error.
func (r *Repository) GetFabricValueOrNil(ctx context.Context, tenantID, fabricID uuid.UUID, key string) (map[string]interface{}, error) {
	if _, err := r.GetFabric(ctx, tenantID, fabricID); err != nil {
		return nil, err
	}
	sv, err := r.stateRepo.GetStateValue(ctx, fabricID, key)
	if err != nil {
		if err.Error() == "state value not found" {
			return nil, nil
		}
		return nil, err
	}
	if sv.Value == nil {
		return nil, nil
	}
	return map[string]interface{}(sv.Value), nil
}

// ListFabricTriggers returns event triggers for a fabric, excluding pipeline-backed triggers.
func (r *Repository) ListFabricTriggers(ctx context.Context, tenantID, fabricID uuid.UUID) ([]*statestore.StateTrigger, error) {
	if _, err := r.GetFabric(ctx, tenantID, fabricID); err != nil {
		return nil, err
	}
	triggers, err := r.stateRepo.GetTriggers(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	out := make([]*statestore.StateTrigger, 0, len(triggers))
	for _, trigger := range triggers {
		if len(stepsFromCondition(trigger.Condition)) > 0 {
			continue
		}
		out = append(out, trigger)
	}
	return out, nil
}

type FabricTriggerInput struct {
	TriggerType             string
	KeyPattern              string
	Condition               map[string]interface{}
	TargetFunctionID        *uuid.UUID
	TargetFunction          string
	IncludePrevious         bool
	IncludeNew              bool
	MaxInvocationsPerMinute int
	IsActive                bool
}

// CreateFabricTrigger creates an event trigger scoped to a fabric.
func (r *Repository) CreateFabricTrigger(ctx context.Context, tenantID, fabricID uuid.UUID, input FabricTriggerInput) (*statestore.StateTrigger, error) {
	if _, err := r.GetFabric(ctx, tenantID, fabricID); err != nil {
		return nil, err
	}
	if input.TriggerType == "" {
		input.TriggerType = "on_write"
	}
	if input.MaxInvocationsPerMinute <= 0 {
		input.MaxInvocationsPerMinute = 60
	}
	trigger := &statestore.StateTrigger{
		TenantID:                tenantID,
		SourceStateID:           &fabricID,
		TriggerType:             input.TriggerType,
		KeyPattern:              strPtr(input.KeyPattern),
		Condition:               statestore.JSONMap(input.Condition),
		TargetFunctionID:        input.TargetFunctionID,
		TargetFunction:          input.TargetFunction,
		IncludePrevious:         input.IncludePrevious,
		IncludeNew:              input.IncludeNew,
		MaxInvocationsPerMinute: input.MaxInvocationsPerMinute,
		IsActive:                input.IsActive,
	}
	return r.stateRepo.CreateTrigger(ctx, trigger)
}

// FabricKeyEntry is a key-value row within a fabric's durable store.
type FabricKeyEntry struct {
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	UpdatedAt string                 `json:"updatedAt"`
}

// ListFabricKeys returns latest values for a fabric, optionally filtered by key prefix.
func (r *Repository) ListFabricKeys(ctx context.Context, tenantID, fabricID uuid.UUID, prefix string, limit, offset int) ([]FabricKeyEntry, int, string, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, 0, "", err
	}
	if state.TenantID != tenantID {
		return nil, 0, "", fmt.Errorf("state fabric not found")
	}

	values, err := r.stateRepo.GetAllStateValues(ctx, fabricID)
	if err != nil {
		return nil, 0, "", err
	}

	filtered := make([]*statestore.StateValue, 0, len(values))
	for _, v := range values {
		if prefix == "" || strings.HasPrefix(v.Key, prefix) {
			filtered = append(filtered, v)
		}
	}

	total := len(filtered)
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	page := filtered[offset:end]

	items := make([]FabricKeyEntry, 0, len(page))
	for _, v := range page {
		val := map[string]interface{}{}
		if v.Value != nil {
			val = map[string]interface{}(v.Value)
		}
		items = append(items, FabricKeyEntry{
			Key:       v.Key,
			Value:     val,
			UpdatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return items, total, state.FullPath, nil
}

// UpdateFabricTrigger updates an event trigger scoped to a fabric.
func (r *Repository) UpdateFabricTrigger(ctx context.Context, tenantID, fabricID, triggerID uuid.UUID, input FabricTriggerInput) (*statestore.StateTrigger, error) {
	trigger, err := r.stateRepo.GetTrigger(ctx, triggerID)
	if err != nil {
		return nil, err
	}
	if trigger.TenantID != tenantID || trigger.SourceStateID == nil || *trigger.SourceStateID != fabricID {
		return nil, fmt.Errorf("trigger not found")
	}
	if len(stepsFromCondition(trigger.Condition)) > 0 {
		return nil, fmt.Errorf("trigger not found")
	}

	if input.TriggerType != "" {
		trigger.TriggerType = input.TriggerType
	}
	if input.KeyPattern != "" {
		trigger.KeyPattern = strPtr(input.KeyPattern)
	}
	if input.Condition != nil {
		trigger.Condition = statestore.JSONMap(input.Condition)
	}
	if input.TargetFunctionID != nil {
		trigger.TargetFunctionID = input.TargetFunctionID
	}
	if input.TargetFunction != "" {
		trigger.TargetFunction = input.TargetFunction
	}
	trigger.IncludePrevious = input.IncludePrevious
	trigger.IncludeNew = input.IncludeNew
	if input.MaxInvocationsPerMinute > 0 {
		trigger.MaxInvocationsPerMinute = input.MaxInvocationsPerMinute
	}
	trigger.IsActive = input.IsActive

	return r.stateRepo.UpdateTrigger(ctx, trigger)
}

// RestoreFabricSnapshot restores live fabric state from a snapshot by ID.
// It creates a backup snapshot before restoring to allow rollback if needed.
func (r *Repository) RestoreFabricSnapshot(ctx context.Context, tenantID, fabricID, snapshotID uuid.UUID) error {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return err
	}
	if state.TenantID != tenantID {
		return fmt.Errorf("state fabric not found")
	}

	var snapshot statestore.StateSnapshot
	if err := r.db.WithContext(ctx).Where("id = ? AND state_id = ?", snapshotID, fabricID).First(&snapshot).Error; err != nil {
		return fmt.Errorf("snapshot not found")
	}

	// Create a backup snapshot before restoring
	backupLabel := fmt.Sprintf("pre-restore-backup-%s", snapshotID.String()[:8])
	_, backupErr := r.CreateSnapshot(ctx, tenantID, fabricID, backupLabel)
	if backupErr != nil {
		// Log the error but continue with restore - backup failure shouldn't block restore
		logrus.WithError(backupErr).WithFields(logrus.Fields{
			"fabric_id":   fabricID,
			"snapshot_id": snapshotID,
		}).Warn("Failed to create backup snapshot before restore - continuing with restore anyway")
	}

	return r.stateRepo.RestoreSnapshot(ctx, fabricID, snapshot.SnapshotVersion, "dashboard", snapshotID.String())
}

// DeleteFabricTrigger deletes a trigger after verifying it belongs to the fabric.
func (r *Repository) DeleteFabricTrigger(ctx context.Context, tenantID, fabricID, triggerID uuid.UUID) error {
	trigger, err := r.stateRepo.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}
	if trigger.TenantID != tenantID || trigger.SourceStateID == nil || *trigger.SourceStateID != fabricID {
		return fmt.Errorf("trigger not found")
	}
	if len(stepsFromCondition(trigger.Condition)) > 0 {
		return fmt.Errorf("trigger not found")
	}
	return r.stateRepo.DeleteTrigger(ctx, triggerID)
}

func strPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
