// Package statefabric: implements wasmpool.StateFabricRepo by
// delegating to the orchestrator's statefabric.Repository and mapping
// between the full `statefabric.Fabric` struct and the minimal
// `wasmpool.Fabric` struct.
package statefabric

import (
	"context"

	wasmpool "github.com/functionfly/functionfly/internal/wasm"
	"github.com/functionfly/functionfly/internal/storage/statefabric"
	"github.com/google/uuid"
)

// Adapter wraps statefabric.Repository to satisfy wasmpool.StateFabricRepo.
type Adapter struct {
	Repo *statefabric.Repository
}

// NewAdapter returns a wasmpool.StateFabricRepo backed by the given repository.
func NewAdapter(repo *statefabric.Repository) wasmpool.StateFabricRepo {
	return &Adapter{Repo: repo}
}

// GetFabric fetches a fabric and maps it to the wasm module's minimal
// Fabric struct.
func (a *Adapter) GetFabric(ctx context.Context, tenantID, fabricID uuid.UUID) (*wasmpool.Fabric, error) {
	f, err := a.Repo.GetFabric(ctx, tenantID, fabricID)
	if err != nil {
		return nil, err
	}
	return toWasmFabric(f), nil
}

// ListStores fetches a fabric's stores and maps them.
func (a *Adapter) ListStores(ctx context.Context, tenantID, fabricID uuid.UUID) ([]wasmpool.FabricStore, error) {
	stores, err := a.Repo.ListStores(ctx, tenantID, fabricID)
	if err != nil {
		return nil, err
	}
	out := make([]wasmpool.FabricStore, len(stores))
	for i, s := range stores {
		out[i] = toWasmStore(s)
	}
	return out, nil
}

// CreateSnapshot creates a snapshot and maps it.
func (a *Adapter) CreateSnapshot(ctx context.Context, tenantID, fabricID uuid.UUID, name string) (*wasmpool.Snapshot, error) {
	s, err := a.Repo.CreateSnapshot(ctx, tenantID, fabricID, name)
	if err != nil {
		return nil, err
	}
	return toWasmSnapshot(s), nil
}

// UpdateFabric updates a fabric with the given patch and returns the
// updated fabric mapped to the wasm module's minimal struct.
func (a *Adapter) UpdateFabric(ctx context.Context, tenantID, fabricID uuid.UUID, updates map[string]interface{}) (*wasmpool.Fabric, error) {
	f, err := a.Repo.UpdateFabric(ctx, tenantID, fabricID, updates)
	if err != nil {
		return nil, err
	}
	return toWasmFabric(f), nil
}

// toWasmFabric converts a statefabric.Fabric to wasmpool.Fabric.
// Performs type conversions: float64 latency → int64 (rounded) and filters
// out fields the wasm module doesn't need. IDs are already uuid.UUID.
func toWasmFabric(f *statefabric.Fabric) *wasmpool.Fabric {
	return &wasmpool.Fabric{
		ID:          f.ID,
		TenantID:    f.TenantID,
		Name:        f.Name,
		Description: f.Description,
		Type:        f.Type,
		Status:      f.Status,
		Settings:    f.Settings,
		Throughput:  f.Throughput,
		Latency:     int64(f.Latency),
		CreatedAt:   f.CreatedAt,
	}
}

// toWasmStore converts a statefabric.FabricStore to wasmpool.FabricStore.
func toWasmStore(s statefabric.FabricStore) wasmpool.FabricStore {
	return wasmpool.FabricStore{
		ID:      parseUUID(s.ID),
		Name:    s.Name,
		Type:    s.Type,
		Status:  s.Status,
		Size:    s.Size,
		MaxSize: s.MaxSize,
		Region:  s.Region,
	}
}

// toWasmSnapshot converts a statefabric.Snapshot to wasmpool.Snapshot.
func toWasmSnapshot(s *statefabric.Snapshot) *wasmpool.Snapshot {
	return &wasmpool.Snapshot{
		ID:         parseUUID(s.ID),
		FabricID:   parseUUID(s.FabricID),
		Name:       s.Name,
		CreatedAt:  s.CreatedAt,
		EventCount: int64(s.EventCount),
		SizeBytes:  s.SizeBytes,
	}
}

// parseUUID converts a string UUID to uuid.UUID. Returns uuid.Nil if the
// input is empty or invalid; the caller should treat nil as a missing
// reference rather than a fatal error.
func parseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
