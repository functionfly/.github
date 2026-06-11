//go:build cgo

package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage/statefabric"
	"github.com/google/uuid"
)

// StateFabricHostHandler implements host functions for StateFabric operations
// This handler is used by the WASM runtime to provide state access to edge functions
type StateFabricHostHandler struct {
	*DefaultHostHandler
	repo     *statefabric.Repository
	tenantID uuid.UUID
	fabricID uuid.UUID
	ctx      context.Context
}

// NewStateFabricHostHandler creates a new handler with StateFabric support
func NewStateFabricHostHandler(
	baseHandler *DefaultHostHandler,
	repo *statefabric.Repository,
	tenantID uuid.UUID,
	fabricID uuid.UUID,
) *StateFabricHostHandler {
	return &StateFabricHostHandler{
		DefaultHostHandler: baseHandler,
		repo:               repo,
		tenantID:           tenantID,
		fabricID:           fabricID,
		ctx:                context.Background(),
	}
}

// StateGet retrieves a value from StateFabric
// The path format is: "tenant_id/fabric_id/key" or just "key" (uses handler's tenant/fabric)
func (h *StateFabricHostHandler) StateGet(path string) (string, error) {
	// Parse the path to extract tenant, fabric, and key
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Get the fabric to verify access
	fabric, err := h.repo.GetFabric(h.ctx, tenantID, fabricID)
	if err != nil {
		return "", fmt.Errorf("fabric not found: %w", err)
	}

	// Check if fabric is active
	if fabric.Status != "active" && fabric.Status != "pending" {
		return "", fmt.Errorf("fabric is not active: %s", fabric.Status)
	}

	// Get the value from the fabric store
	// For edge functions, we use a simplified key-value interface
	value, err := h.getValueFromFabric(fabricID, key)
	if err != nil {
		return "", err
	}

	// Return as JSON
	result := map[string]interface{}{
		"value":     value,
		"path":      path,
		"fabric_id": fabricID.String(),
		"tenant_id": tenantID.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
}

// StateSet stores a value in StateFabric
func (h *StateFabricHostHandler) StateSet(path string, value string) error {
	// Parse the path
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Verify fabric exists and is active
	fabric, err := h.repo.GetFabric(h.ctx, tenantID, fabricID)
	if err != nil {
		return fmt.Errorf("fabric not found: %w", err)
	}

	if fabric.Status != "active" && fabric.Status != "pending" {
		return fmt.Errorf("fabric is not active: %s", fabric.Status)
	}

	// Validate the value is valid JSON
	var valueData interface{}
	if err := json.Unmarshal([]byte(value), &valueData); err != nil {
		return fmt.Errorf("invalid JSON value: %w", err)
	}

	// Store the value
	if err := h.setValueInFabric(fabricID, key, value); err != nil {
		return err
	}

	return nil
}

// StateDelete removes a value from StateFabric
func (h *StateFabricHostHandler) StateDelete(path string) error {
	// Parse the path
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Verify fabric exists and is active
	fabric, err := h.repo.GetFabric(h.ctx, tenantID, fabricID)
	if err != nil {
		return fmt.Errorf("fabric not found: %w", err)
	}

	if fabric.Status != "active" && fabric.Status != "pending" {
		return fmt.Errorf("fabric is not active: %s", fabric.Status)
	}

	// Delete the value
	if err := h.deleteValueFromFabric(fabricID, key); err != nil {
		return err
	}

	return nil
}

// StateGetFabric retrieves fabric metadata
func (h *StateFabricHostHandler) StateGetFabric(fabricIDStr string) (string, error) {
	// Parse fabric ID
	var fabricID uuid.UUID
	if fabricIDStr == "" {
		fabricID = h.fabricID
	} else {
		var err error
		fabricID, err = uuid.Parse(fabricIDStr)
		if err != nil {
			return "", fmt.Errorf("invalid fabric ID: %w", err)
		}
	}

	// Get fabric info
	fabric, err := h.repo.GetFabric(h.ctx, h.tenantID, fabricID)
	if err != nil {
		return "", fmt.Errorf("fabric not found: %w", err)
	}

	// Get stores
	stores, err := h.repo.ListStores(h.ctx, h.tenantID, fabricID)
	if err != nil {
		// Non-fatal, continue without stores
		stores = []statefabric.FabricStore{}
	}

	// Build response
	result := map[string]interface{}{
		"fabric": map[string]interface{}{
			"id":          fabric.ID.String(),
			"name":        fabric.Name,
			"description": fabric.Description,
			"status":      fabric.Status,
			"type":        fabric.Type,
			"settings":    fabric.Settings,
			"throughput":  fabric.Throughput,
			"latency_ms":  fabric.Latency,
			"created_at":  fabric.CreatedAt.Format(time.RFC3339),
		},
		"stores": make([]map[string]interface{}, len(stores)),
	}

	for i, store := range stores {
		result["stores"].([]map[string]interface{})[i] = map[string]interface{}{
			"id":       store.ID,
			"name":     store.Name,
			"type":     store.Type,
			"status":   store.Status,
			"size":     store.Size,
			"max_size": store.MaxSize,
			"region":   store.Region,
		}
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal fabric info: %w", err)
	}

	return string(jsonResult), nil
}

// StateCreateSnapshot creates a snapshot of state
func (h *StateFabricHostHandler) StateCreateSnapshot(path string, label string) (string, error) {
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	snapshotName := label
	if snapshotName == "" {
		snapshotName = fmt.Sprintf("snapshot-%s", uuid.New().String()[:8])
	}

	snapshot, err := h.repo.CreateSnapshot(h.ctx, tenantID, fabricID, snapshotName)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	result := map[string]interface{}{
		"id":                snapshot.ID,
		"fabric_id":         fabricID.String(),
		"tenant_id":         tenantID.String(),
		"key":               key,
		"label":             label,
		"path":              path,
		"created_at":        snapshot.CreatedAt.Format(time.RFC3339),
		"event_count":       snapshot.EventCount,
		"size_bytes":        snapshot.SizeBytes,
		"status":            "created",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(jsonResult), nil
}

// parsePath parses a state path into tenant, fabric, and key components
// Supported formats:
// - "key" -> uses handler's default tenant and fabric
// - "fabric_id/key" -> uses handler's default tenant
// - "tenant_id/fabric_id/key" -> full path
func (h *StateFabricHostHandler) parsePath(path string) (uuid.UUID, uuid.UUID, string, error) {
	parts := strings.Split(path, "/")

	switch len(parts) {
	case 1:
		// Just a key, use defaults
		return h.tenantID, h.fabricID, parts[0], nil
	case 2:
		// fabric/key
		fabricID, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid fabric ID: %w", err)
		}
		return h.tenantID, fabricID, parts[1], nil
	case 3:
		// tenant/fabric/key
		tenantID, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid tenant ID: %w", err)
		}
		fabricID, err := uuid.Parse(parts[1])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid fabric ID: %w", err)
		}
		return tenantID, fabricID, parts[2], nil
	default:
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid path format, expected 'tenant/fabric/key', 'fabric/key', or 'key'")
	}
}

// EdgeStateEntry represents a single edge state entry with TTL support
type EdgeStateEntry struct {
	Value      interface{} `json:"value"`
	ExpiresAt  *time.Time  `json:"expiresAt,omitempty"`
	CreatedAt  time.Time   `json:"createdAt"`
	ModifiedAt time.Time   `json:"modifiedAt"`
}

// getValueFromFabric retrieves a value from the fabric with TTL enforcement
// This is a simplified implementation using the fabric's settings as a key-value store
func (h *StateFabricHostHandler) getValueFromFabric(fabricID uuid.UUID, key string) (interface{}, error) {
	// Get fabric
	fabric, err := h.repo.GetFabric(h.ctx, h.tenantID, fabricID)
	if err != nil {
		return nil, err
	}

	// For edge state, we store values in the fabric settings under a "_edge_state" key
	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		return nil, nil // Key not found, return nil
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid edge state format")
	}

	value, exists := stateMap[key]
	if !exists {
		return nil, nil
	}

	// Check if value is an EdgeStateEntry with TTL
	if entry, ok := value.(map[string]interface{}); ok {
		if expiresAt, ok := entry["expiresAt"].(string); ok && expiresAt != "" {
			expTime, err := time.Parse(time.RFC3339, expiresAt)
			if err == nil && time.Now().After(expTime) {
				// TTL expired, delete the key and return nil
				h.deleteValueFromFabric(fabricID, key)
				return nil, nil
			}
		}
		// Return the actual value from the entry
		if v, ok := entry["value"]; ok {
			return v, nil
		}
	}

	return value, nil
}

// setValueInFabric stores a value in the fabric with optional TTL
func (h *StateFabricHostHandler) setValueInFabric(fabricID uuid.UUID, key string, value string) error {
	// Get fabric
	fabric, err := h.repo.GetFabric(h.ctx, h.tenantID, fabricID)
	if err != nil {
		return err
	}

	// Parse the JSON value
	var valueData interface{}
	if err := json.Unmarshal([]byte(value), &valueData); err != nil {
		return err
	}

	// Get or create edge state map
	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		edgeState = make(map[string]interface{})
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		stateMap = make(map[string]interface{})
	}

	// Wrap value in EdgeStateEntry with TTL support
	now := time.Now().UTC()
	entry := map[string]interface{}{
		"value":      valueData,
		"createdAt":  now.Format(time.RFC3339),
		"modifiedAt": now.Format(time.RFC3339),
	}

	// Check for TTL in the value data if it's a map with ttl field
	if valueMap, ok := valueData.(map[string]interface{}); ok {
		if ttl, ok := valueMap["_ttl"].(float64); ok && ttl > 0 {
			expiresAt := now.Add(time.Duration(int64(ttl)) * time.Second)
			entry["expiresAt"] = expiresAt.Format(time.RFC3339)
			// Remove the _ttl field from the actual value
			delete(valueMap, "_ttl")
			entry["value"] = valueMap
		}
	}

	stateMap[key] = entry
	fabric.Settings["_edge_state"] = stateMap

	// Update fabric
	updates := map[string]interface{}{
		"settings": fabric.Settings,
	}

	_, err = h.repo.UpdateFabric(h.ctx, h.tenantID, fabricID, updates)
	if err != nil {
		return fmt.Errorf("failed to update fabric: %w", err)
	}

	return nil
}

// deleteValueFromFabric removes a value from the fabric
func (h *StateFabricHostHandler) deleteValueFromFabric(fabricID uuid.UUID, key string) error {
	// Get fabric
	fabric, err := h.repo.GetFabric(h.ctx, h.tenantID, fabricID)
	if err != nil {
		return err
	}

	// Get edge state map
	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		return nil // Nothing to delete
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		return nil
	}

	// Delete the key
	delete(stateMap, key)
	fabric.Settings["_edge_state"] = stateMap

	// Update fabric
	updates := map[string]interface{}{
		"settings": fabric.Settings,
	}

	_, err = h.repo.UpdateFabric(h.ctx, h.tenantID, fabricID, updates)
	if err != nil {
		return fmt.Errorf("failed to update fabric: %w", err)
	}

	return nil
}

// WithContext returns a new handler with the given context
func (h *StateFabricHostHandler) WithContext(ctx context.Context) *StateFabricHostHandler {
	return &StateFabricHostHandler{
		DefaultHostHandler: h.DefaultHostHandler,
		repo:               h.repo,
		tenantID:           h.tenantID,
		fabricID:           h.fabricID,
		ctx:                ctx,
	}
}

// Ensure StateFabricHostHandler implements HostFunctionHandler
var _ HostFunctionHandler = (*StateFabricHostHandler)(nil)
