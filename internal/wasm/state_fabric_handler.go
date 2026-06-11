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

type StateFabricHostHandler struct {
	*DefaultHostHandler
	repo     *statefabric.Repository
	tenantID uuid.UUID
	fabricID uuid.UUID
	ctx      context.Context
}

func NewStateFabricHostHandler(
	ctx context.Context,
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
		ctx:                ctx,
	}
}

func (h *StateFabricHostHandler) StateGet(path string) (string, error) {
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	fabric, err := h.repo.GetFabric(h.ctx, tenantID, fabricID)
	if err != nil {
		return "", fmt.Errorf("fabric not found")
	}

	if fabric.Status != statefabric.FabricStatusActive && fabric.Status != statefabric.FabricStatusPending {
		return "", fmt.Errorf("fabric is not active")
	}

	value, err := h.getValueFromFabricWithFabric(fabric, key)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"value":     value,
		"path":      path,
		"fabric_id": fabricID.String(),
		"tenant_id": tenantID.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result")
	}

	return string(jsonResult), nil
}

func (h *StateFabricHostHandler) StateSet(path string, value string) error {
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path")
	}

	fabric, err := h.repo.GetFabric(h.ctx, tenantID, fabricID)
	if err != nil {
		return fmt.Errorf("fabric not found")
	}

	if fabric.Status != statefabric.FabricStatusActive && fabric.Status != statefabric.FabricStatusPending {
		return fmt.Errorf("fabric is not active")
	}

	var valueData interface{}
	if err := json.Unmarshal([]byte(value), &valueData); err != nil {
		return fmt.Errorf("invalid JSON value")
	}

	if err := h.setValueInFabricWithFabric(fabric, key, valueData); err != nil {
		return err
	}

	return nil
}

func (h *StateFabricHostHandler) StateDelete(path string) error {
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path")
	}

	fabric, err := h.repo.GetFabric(h.ctx, tenantID, fabricID)
	if err != nil {
		return fmt.Errorf("fabric not found")
	}

	if fabric.Status != statefabric.FabricStatusActive && fabric.Status != statefabric.FabricStatusPending {
		return fmt.Errorf("fabric is not active")
	}

	if err := h.deleteValueFromFabricWithFabric(fabric, key); err != nil {
		return err
	}

	return nil
}

func (h *StateFabricHostHandler) StateGetFabric(fabricIDStr string) (string, error) {
	var fabricID uuid.UUID
	var err error

	if fabricIDStr == "" {
		fabricID = h.fabricID
	} else {
		fabricID, err = uuid.Parse(fabricIDStr)
		if err != nil {
			return "", fmt.Errorf("invalid fabric ID")
		}
	}

	fabric, err := h.repo.GetFabric(h.ctx, h.tenantID, fabricID)
	if err != nil {
		return "", fmt.Errorf("fabric not found")
	}

	if fabric.TenantID != h.tenantID {
		return "", fmt.Errorf("access denied: fabric belongs to different tenant")
	}

	stores, err := h.repo.ListStores(h.ctx, h.tenantID, fabricID)
	if err != nil {
		stores = []statefabric.FabricStore{}
	}

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
		return "", fmt.Errorf("failed to marshal fabric info")
	}

	return string(jsonResult), nil
}

func (h *StateFabricHostHandler) StateCreateSnapshot(path string, label string) (string, error) {
	tenantID, fabricID, key, err := h.parsePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	snapshotName := label
	if snapshotName == "" {
		snapshotName = fmt.Sprintf("snapshot-%s", uuid.New().String()[:8])
	}

	snapshot, err := h.repo.CreateSnapshot(h.ctx, tenantID, fabricID, snapshotName)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot")
	}

	result := map[string]interface{}{
		"id":          snapshot.ID,
		"fabric_id":   fabricID.String(),
		"tenant_id":   tenantID.String(),
		"key":         key,
		"label":       label,
		"path":        path,
		"created_at":  snapshot.CreatedAt.Format(time.RFC3339),
		"event_count": snapshot.EventCount,
		"size_bytes":  snapshot.SizeBytes,
		"status":      "created",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot")
	}

	return string(jsonResult), nil
}

func (h *StateFabricHostHandler) parsePath(path string) (uuid.UUID, uuid.UUID, string, error) {
	parts := strings.Split(path, "/")

	switch len(parts) {
	case 1:
		return h.tenantID, h.fabricID, parts[0], nil
	case 2:
		fabricID, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid fabric ID")
		}
		return h.tenantID, fabricID, parts[1], nil
	case 3:
		tenantID, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid tenant ID")
		}
		fabricID, err := uuid.Parse(parts[1])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid fabric ID")
		}
		return tenantID, fabricID, parts[2], nil
	default:
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid path format")
	}
}

func (h *StateFabricHostHandler) getValueFromFabricWithFabric(fabric *statefabric.Fabric, key string) (interface{}, error) {
	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		return nil, nil
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid edge state format")
	}

	value, exists := stateMap[key]
	if !exists {
		return nil, nil
	}

	return value, nil
}

func (h *StateFabricHostHandler) setValueInFabricWithFabric(fabric *statefabric.Fabric, key string, valueData interface{}) error {
	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		edgeState = make(map[string]interface{})
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid edge state format: expected map, got %T", edgeState)
	}

	stateMap[key] = valueData
	fabric.Settings["_edge_state"] = stateMap

	updates := map[string]interface{}{
		"settings": fabric.Settings,
	}

	_, err := h.repo.UpdateFabric(h.ctx, fabric.TenantID, fabric.ID, updates)
	if err != nil {
		return fmt.Errorf("failed to update fabric")
	}

	return nil
}

func (h *StateFabricHostHandler) deleteValueFromFabricWithFabric(fabric *statefabric.Fabric, key string) error {
	edgeState, ok := fabric.Settings["_edge_state"]
	if !ok {
		return nil
	}

	stateMap, ok := edgeState.(map[string]interface{})
	if !ok {
		return nil
	}

	delete(stateMap, key)
	fabric.Settings["_edge_state"] = stateMap

	updates := map[string]interface{}{
		"settings": fabric.Settings,
	}

	_, err := h.repo.UpdateFabric(h.ctx, fabric.TenantID, fabric.ID, updates)
	if err != nil {
		return fmt.Errorf("failed to update fabric")
	}

	return nil
}

func (h *StateFabricHostHandler) WithContext(ctx context.Context) *StateFabricHostHandler {
	return &StateFabricHostHandler{
		DefaultHostHandler: h.DefaultHostHandler,
		repo:               h.repo,
		tenantID:           h.tenantID,
		fabricID:           h.fabricID,
		ctx:                ctx,
	}
}

var _ HostFunctionHandler = (*StateFabricHostHandler)(nil)
