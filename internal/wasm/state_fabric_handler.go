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

	if fabric.Status != string(statefabric.FabricStatusOnline) && fabric.Status != string(statefabric.FabricStatusPending) {
		return "", fmt.Errorf("fabric is not active")
	}

	value, err := h.repo.GetFabricValue(h.ctx, tenantID, fabricID, key)
	if err != nil {
		return "", fmt.Errorf("failed to get value: %v", err)
	}

	result := map[string]interface{}{
		"value":     value,
		"path":      path,
		"key":       key,
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

	if fabric.Status != string(statefabric.FabricStatusOnline) && fabric.Status != string(statefabric.FabricStatusPending) {
		return fmt.Errorf("fabric is not active")
	}

	var valueData map[string]interface{}
	if err := json.Unmarshal([]byte(value), &valueData); err != nil {
		return fmt.Errorf("invalid JSON value: %v", err)
	}

	_, err = h.repo.SetFabricValue(h.ctx, tenantID, fabricID, key, valueData, "edge")
	if err != nil {
		return fmt.Errorf("failed to set value: %v", err)
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

	if fabric.Status != string(statefabric.FabricStatusOnline) && fabric.Status != string(statefabric.FabricStatusPending) {
		return fmt.Errorf("fabric is not active")
	}

	err = h.repo.DeleteFabricValue(h.ctx, tenantID, fabricID, key, "edge")
	if err != nil {
		return fmt.Errorf("failed to delete value: %v", err)
	}

	return nil
}

func (h *StateFabricHostHandler) StateList(prefix string) (string, error) {
	entries, _, _, err := h.repo.ListFabricKeys(h.ctx, h.tenantID, h.fabricID, prefix, 1000, 0)
	if err != nil {
		return "", fmt.Errorf("failed to list keys: %v", err)
	}

	result := map[string]interface{}{
		"keys":      entries,
		"fabric_id": h.fabricID.String(),
		"tenant_id": h.tenantID.String(),
		"prefix":    prefix,
		"count":     len(entries),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result")
	}

	return string(jsonResult), nil
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
