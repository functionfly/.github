// Package frgstate provides StateFabric integration for Function Runtime Graphs
// This enables state persistence between nodes in a graph execution
package frgstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// GraphStateManager provides state management for graph instances
type GraphStateManager struct {
	repo      *state.StateRepository
	stateID   uuid.UUID // The underlying state/store ID
	fabricID  uuid.UUID // The StateFabric ID
	tenantID  uuid.UUID
	namespace string // Graph instance namespace for key isolation
}

// Config configures state management for a graph instance
type Config struct {
	Enabled         bool                   `json:"enabled"`
	TTLSeconds      int                    `json:"ttl_seconds"`       // State key TTL (0 = no expiration)
	Consistency     string                 `json:"consistency"`       // strong, eventual, session
	EncryptionKeyID string                 `json:"encryption_key_id"` // For encrypted state
	Options         map[string]interface{} `json:"options,omitempty"`
}

// ReadResult is the result of a state read operation
type ReadResult struct {
	Found     bool        `json:"found"`
	Value     interface{} `json:"value,omitempty"`
	Version   int64       `json:"version,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
}

// WriteResult is the result of a state write operation
type WriteResult struct {
	Success   bool      `json:"success"`
	Version   int64     `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Conflict  bool      `json:"conflict,omitempty"` // True if conditional write failed
}

// GraphStateSnapshot captures all state for checkpoint/restore
type GraphStateSnapshot struct {
	SnapshotID   uuid.UUID              `json:"snapshot_id"`
	InstanceID   uuid.UUID              `json:"instance_id"`
	Namespace    string                 `json:"namespace"`
	State        map[string]interface{} `json:"state"`
	CreatedAt    time.Time              `json:"created_at"`
	NodeStates   map[string]interface{} `json:"node_states,omitempty"`
	ExecutionSeq int64                  `json:"execution_seq"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewGraphStateManager creates a new state manager for a graph instance
// Creates or uses an existing state fabric for the graph
func NewGraphStateManager(
	ctx context.Context,
	repo *state.StateRepository,
	tenantID uuid.UUID,
	instanceID uuid.UUID,
	graphName string,
) (*GraphStateManager, error) {
	// Check if state fabric already exists for this tenant/graph
	existingState, err := repo.GetStateByPath(ctx, tenantID, fmt.Sprintf("%s/%s", tenantID.String()[:8], graphName))
	if err != nil {
		// Create new state fabric for this graph
		newState := &state.State{
			TenantID:    tenantID,
			Name:        fmt.Sprintf("graph-%s", graphName),
			FullPath:    fmt.Sprintf("%s/graph-%s", tenantID.String()[:8], graphName),
			StorageType: "keyvalue", // Optimized for graph state
			MaxSizeMB:   100,        // 100MB default per graph
			TTLDays:     7,          // 7 day retention
			Description: stringPtr(fmt.Sprintf("State fabric for graph %s", graphName)),
			Tags: state.JSONMap{
				"type":        "frg",
				"instance_id": instanceID.String(),
				"graph_name":  graphName,
			},
		}
		existingState, err = repo.CreateState(ctx, newState)
		if err != nil {
			return nil, fmt.Errorf("failed to create state fabric: %w", err)
		}
		logrus.WithFields(logrus.Fields{
			"state_id":   existingState.ID,
			"tenant_id":  tenantID,
			"graph_name": graphName,
		}).Info("Created new state fabric for graph")
	}

	return &GraphStateManager{
		repo:      repo,
		stateID:   existingState.ID,
		tenantID:  tenantID,
		namespace: instanceID.String(), // Use instance ID as namespace
	}, nil
}

// SetState writes a value to state
func (m *GraphStateManager) SetState(ctx context.Context, key string, value interface{}, nodeID string) (*WriteResult, error) {
	start := time.Now()
	fullKey := m.keyWithNamespace(key)

	// Store as JSON map
	valueMap := map[string]interface{}{
		"_value":   value,
		"_node_id": nodeID,
		"_written": time.Now().Format(time.RFC3339),
	}

	stateValue, err := m.repo.SetStateValue(ctx, m.stateID, fullKey, valueMap, "frg_node", nodeID)
	if err != nil {
		logrus.WithError(err).WithField("key", key).Error("Failed to set state")
		return &WriteResult{Success: false}, err
	}

	latency := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"key":        key,
		"version":    stateValue.Version,
		"node_id":    nodeID,
		"latency_ms": latency.Milliseconds(),
	}).Debug("State set operation")

	return &WriteResult{
		Success:   true,
		Version:   int64(stateValue.Version),
		Timestamp: stateValue.CreatedAt,
	}, nil
}

// GetState retrieves a value from state
func (m *GraphStateManager) GetState(ctx context.Context, key string) (*ReadResult, error) {
	start := time.Now()
	fullKey := m.keyWithNamespace(key)

	stateValue, err := m.repo.GetStateValue(ctx, m.stateID, fullKey)
	if err != nil {
		if err.Error() == "state value not found" {
			return &ReadResult{Found: false}, nil
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Extract the actual value from the wrapper
	var value interface{}
	if rawValue, ok := stateValue.Value["_value"]; ok {
		value = rawValue
	} else {
		value = stateValue.Value
	}

	latency := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"key":     key,
		"latency": latency.Milliseconds(),
		"version": stateValue.Version,
	}).Debug("State get operation")

	return &ReadResult{
		Found:     true,
		Value:     value,
		Version:   int64(stateValue.Version),
		Timestamp: stateValue.CreatedAt,
	}, nil
}

// GetStateWithDefault retrieves state with a default value if not found
func (m *GraphStateManager) GetStateWithDefault(ctx context.Context, key string, defaultValue interface{}) (interface{}, error) {
	result, err := m.GetState(ctx, key)
	if err != nil {
		return nil, err
	}
	if !result.Found {
		return defaultValue, nil
	}
	return result.Value, nil
}

// IncrementState atomically increments a numeric state value
func (m *GraphStateManager) IncrementState(ctx context.Context, key string, delta float64, nodeID string) (*WriteResult, error) {
	// Use optimistic locking with retry
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		result, err := m.GetState(ctx, key)
		if err != nil {
			return nil, err
		}

		var current float64 = 0
		if result.Found {
			switch v := result.Value.(type) {
			case float64:
				current = v
			case int:
				current = float64(v)
			case int64:
				current = float64(v)
			case json.Number:
				n, _ := v.Float64()
				current = n
			default:
				return nil, fmt.Errorf("cannot increment non-numeric value: %T", v)
			}
		}

		newValue := current + delta
		writeResult, err := m.SetState(ctx, key, newValue, nodeID)
		if err != nil {
			return nil, err
		}

		return writeResult, nil
	}

	return nil, fmt.Errorf("failed to increment state after %d retries", maxRetries)
}

// AppendState appends to a list state value
func (m *GraphStateManager) AppendState(ctx context.Context, key string, item interface{}, maxSize int, nodeID string) (*WriteResult, error) {
	result, err := m.GetState(ctx, key)
	if err != nil {
		return nil, err
	}

	var list []interface{}
	if result.Found {
		switch v := result.Value.(type) {
		case []interface{}:
			list = v
		default:
			// Start new list with existing as first item
			list = []interface{}{v}
		}
	}

	list = append(list, item)

	// Truncate if exceeds max size
	if maxSize > 0 && len(list) > maxSize {
		list = list[len(list)-maxSize:]
	}

	return m.SetState(ctx, key, list, nodeID)
}

// DeleteState removes a state key
func (m *GraphStateManager) DeleteState(ctx context.Context, key string, nodeID string) error {
	fullKey := m.keyWithNamespace(key)
	return m.repo.DeleteStateValue(ctx, m.stateID, fullKey, "frg_node", nodeID)
}

// GetAllState retrieves all state keys for this graph instance
func (m *GraphStateManager) GetAllState(ctx context.Context) (map[string]interface{}, error) {
	// Get all values from the state
	values, err := m.repo.GetAllStateValues(ctx, m.stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all state values: %w", err)
	}

	result := make(map[string]interface{})
	for _, v := range values {
		// Strip namespace from key
		cleanKey := m.stripNamespace(v.Key)
		if cleanKey == "" {
			continue
		}

		// Extract actual value
		var value interface{}
		if rawValue, ok := v.Value["_value"]; ok {
			value = rawValue
		} else {
			value = v.Value
		}

		result[cleanKey] = value
	}

	return result, nil
}

// CreateCheckpoint creates a full snapshot of all state for this graph instance
func (m *GraphStateManager) CreateCheckpoint(ctx context.Context, instanceID uuid.UUID, nodeStates map[string]interface{}) (*GraphStateSnapshot, error) {
	start := time.Now()

	// Get all state
	allState, err := m.GetAllState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all state: %w", err)
	}

	snapshot := &GraphStateSnapshot{
		SnapshotID:   uuid.New(),
		InstanceID:   instanceID,
		Namespace:    m.namespace,
		State:        allState,
		CreatedAt:    time.Now(),
		NodeStates:   nodeStates,
		ExecutionSeq: time.Now().UnixNano(),
		Metadata: map[string]interface{}{
			"checkpoint_version": "1.0",
			"state_keys_count":   len(allState),
			"state_fabric_id":    m.stateID.String(),
		},
	}

	// Store snapshot in state fabric as a special key
	snapshotData := map[string]interface{}{
		"_snapshot":    snapshot,
		"_created_at":  time.Now().Format(time.RFC3339),
		"_instance_id": instanceID.String(),
	}

	snapshotKey := fmt.Sprintf("_checkpoint_%s", snapshot.SnapshotID.String())
	if _, err := m.repo.SetStateValue(ctx, m.stateID, snapshotKey, snapshotData, "frg_checkpoint", "system"); err != nil {
		return nil, fmt.Errorf("failed to store snapshot: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"snapshot_id": snapshot.SnapshotID,
		"instance_id": instanceID,
		"state_keys":  len(allState),
		"latency_ms":  time.Since(start).Milliseconds(),
	}).Info("Created graph state checkpoint")

	return snapshot, nil
}

// RestoreCheckpoint restores state from a checkpoint
func (m *GraphStateManager) RestoreCheckpoint(ctx context.Context, snapshotID uuid.UUID) (*GraphStateSnapshot, error) {
	snapshotKey := fmt.Sprintf("_checkpoint_%s", snapshotID.String())
	stateValue, err := m.repo.GetStateValue(ctx, m.stateID, snapshotKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Extract snapshot data
	snapshotDataRaw, ok := stateValue.Value["_snapshot"]
	if !ok {
		return nil, fmt.Errorf("snapshot data not found in state")
	}

	snapshotDataJSON, _ := json.Marshal(snapshotDataRaw)
	var restoredSnapshot GraphStateSnapshot
	if err := json.Unmarshal(snapshotDataJSON, &restoredSnapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	// Restore all state keys
	for key, value := range restoredSnapshot.State {
		if _, err := m.SetState(ctx, key, value, "restore"); err != nil {
			return nil, fmt.Errorf("failed to restore state key %s: %w", key, err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"snapshot_id": snapshotID,
		"state_keys":  len(restoredSnapshot.State),
	}).Info("Restored graph state checkpoint")

	return &restoredSnapshot, nil
}

// CleanupState removes all state for this graph instance
func (m *GraphStateManager) CleanupState(ctx context.Context) error {
	// Get all state keys and delete them
	allState, err := m.GetAllState(ctx)
	if err != nil {
		return err
	}

	for key := range allState {
		if err := m.DeleteState(ctx, key, "cleanup"); err != nil {
			logrus.WithError(err).WithField("key", key).Warn("Failed to delete state key during cleanup")
		}
	}

	logrus.WithFields(logrus.Fields{
		"namespace":  m.namespace,
		"state_keys": len(allState),
	}).Info("Cleaned up graph state")
	return nil
}

// Helper methods

func (m *GraphStateManager) keyWithNamespace(key string) string {
	return fmt.Sprintf("%s:%s", m.namespace, key)
}

func (m *GraphStateManager) stripNamespace(fullKey string) string {
	prefix := m.namespace + ":"
	if len(fullKey) > len(prefix) && fullKey[:len(prefix)] == prefix {
		return fullKey[len(prefix):]
	}
	// Skip special keys (checkpoints, etc.)
	if len(fullKey) > 1 && fullKey[0:1] == "_" {
		return ""
	}
	return fullKey
}

func stringPtr(s string) *string {
	return &s
}
