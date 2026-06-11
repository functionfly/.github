package statefabric

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockR2StorageBackend is a mock implementation of R2StorageBackend for testing
type MockR2StorageBackend struct {
	mu          sync.RWMutex
	snapshots   map[string][]byte
	replays     map[string][]byte
	events      map[string][]byte
	memoryBlobs map[string][]byte
	shouldFail  bool
	failError   error
}

func NewMockR2StorageBackend() *MockR2StorageBackend {
	return &MockR2StorageBackend{
		snapshots:   make(map[string][]byte),
		replays:     make(map[string][]byte),
		events:      make(map[string][]byte),
		memoryBlobs: make(map[string][]byte),
	}
}

func (m *MockR2StorageBackend) StoreSnapshotData(ctx context.Context, tenantID, fabricID, snapshotID uuid.UUID, data JSONMap, metadata JSONMap) (*R2StorageObject, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := snapshotID.String()
	m.snapshots[key] = []byte{}
	return &R2StorageObject{
		Key:         key,
		Bucket:      "test-bucket",
		ContentHash: "test-hash",
		CreatedAt:   time.Now(),
	}, nil
}

func (m *MockR2StorageBackend) GetSnapshotData(ctx context.Context, key string) (JSONMap, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, ok := m.snapshots[key]; ok {
		return JSONMap{"data": string(data)}, nil
	}
	return nil, nil
}

func (m *MockR2StorageBackend) DeleteSnapshotData(ctx context.Context, key string) error {
	if m.shouldFail {
		return m.failError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.snapshots, key)
	return nil
}

func (m *MockR2StorageBackend) StoreReplayData(ctx context.Context, tenantID, replayID uuid.UUID, events interface{}, metadata JSONMap) (*R2StorageObject, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := replayID.String()
	m.replays[key] = []byte{}
	return &R2StorageObject{
		Key:         key,
		Bucket:      "test-bucket",
		ContentHash: "test-hash",
		CreatedAt:   time.Now(),
	}, nil
}

func (m *MockR2StorageBackend) GetReplayData(ctx context.Context, key string) (*ReplayData, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.replays[key]; ok {
		return &ReplayData{}, nil
	}
	return nil, nil
}

func (m *MockR2StorageBackend) StoreEventLogs(ctx context.Context, tenantID, fabricID uuid.UUID, events interface{}) (*R2StorageObject, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := uuid.New().String()
	m.events[key] = []byte{}
	return &R2StorageObject{
		Key:         key,
		Bucket:      "test-bucket",
		ContentHash: "test-hash",
		CreatedAt:   time.Now(),
	}, nil
}

func (m *MockR2StorageBackend) StoreMemoryBlob(ctx context.Context, tenantID, memoryID uuid.UUID, content []byte, memoryType string, metadata JSONMap) (*R2StorageObject, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := memoryID.String()
	m.memoryBlobs[key] = content
	return &R2StorageObject{
		Key:         key,
		Bucket:      "test-bucket",
		ContentHash: "test-hash",
		CreatedAt:   time.Now(),
		Size:        int64(len(content)),
	}, nil
}

func (m *MockR2StorageBackend) GetMemoryBlob(ctx context.Context, key string) ([]byte, error) {
	if m.shouldFail {
		return nil, m.failError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, ok := m.memoryBlobs[key]; ok {
		return data, nil
	}
	return nil, nil
}

func (m *MockR2StorageBackend) SetFailure(err error) {
	m.shouldFail = true
	m.failError = err
}

func (m *MockR2StorageBackend) ClearFailure() {
	m.shouldFail = false
	m.failError = nil
}

// MockRepository is a mock implementation of Repository for testing without database
type MockRepository struct {
	mu      sync.RWMutex
	fabrics map[uuid.UUID]*Fabric
	stores  map[uuid.UUID][]FabricStore
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		fabrics: make(map[uuid.UUID]*Fabric),
		stores:  make(map[uuid.UUID][]FabricStore),
	}
}

func (m *MockRepository) CreateFabric(ctx context.Context, tenantID uuid.UUID, name, description, fabricType string, settings map[string]interface{}) (*Fabric, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fabric := &Fabric{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Type:        fabricType,
		TenantID:    tenantID,
		Status:      "online",
		Settings:    settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.fabrics[fabric.ID] = fabric
	return fabric, nil
}

func (m *MockRepository) GetFabric(ctx context.Context, tenantID, fabricID uuid.UUID) (*Fabric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fabric, ok := m.fabrics[fabricID]
	if !ok {
		return nil, fmt.Errorf("state fabric not found")
	}
	if fabric.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	return fabric, nil
}

func (m *MockRepository) ListFabrics(ctx context.Context, opts ListOptions) ([]Fabric, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var fabrics []Fabric
	for _, fabric := range m.fabrics {
		if fabric.TenantID == opts.TenantID {
			fabrics = append(fabrics, *fabric)
		}
	}
	return fabrics, int64(len(fabrics)), nil
}

func (m *MockRepository) DeleteFabric(ctx context.Context, tenantID, fabricID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fabric, ok := m.fabrics[fabricID]
	if !ok {
		return fmt.Errorf("state fabric not found")
	}
	if fabric.TenantID != tenantID {
		return fmt.Errorf("state fabric not found")
	}
	delete(m.fabrics, fabricID)
	return nil
}

func (m *MockRepository) CreateStore(ctx context.Context, tenantID, fabricID uuid.UUID, name, storeType string, maxSize int64, region string) (*FabricStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	store := &FabricStore{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      storeType,
		Status:    "active",
		MaxSize:   maxSize,
		Region:    region,
		Provider:  "functionfly",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.stores[fabricID] = append(m.stores[fabricID], *store)
	return store, nil
}

func (m *MockRepository) ListStores(ctx context.Context, tenantID, fabricID uuid.UUID) ([]FabricStore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stores, ok := m.stores[fabricID]; ok {
		return stores, nil
	}
	return []FabricStore{}, nil
}

func (m *MockRepository) DeleteStore(ctx context.Context, tenantID, fabricID uuid.UUID, storeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stores := m.stores[fabricID]
	for i, store := range stores {
		if store.ID == storeID {
			m.stores[fabricID] = append(stores[:i], stores[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("store not found")
}

func (m *MockRepository) CreateSnapshot(ctx context.Context, tenantID, fabricID uuid.UUID, name string) (*Snapshot, error) {
	return &Snapshot{
		ID:         uuid.New().String(),
		FabricID:   fabricID.String(),
		Name:       name,
		EventCount: 0,
		SizeBytes:  0,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockRepository) ListSnapshots(ctx context.Context, tenantID, fabricID uuid.UUID) ([]Snapshot, error) {
	return []Snapshot{}, nil
}

func (m *MockRepository) CreateReplay(ctx context.Context, tenantID, fabricID uuid.UUID, req ReplayCreateRequest) (*ReplaySession, error) {
	return &ReplaySession{
		ID:        uuid.New().String(),
		FabricID:  fabricID.String(),
		Status:    "completed",
		Progress:  100,
		StartedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ListReplays(ctx context.Context, tenantID, fabricID uuid.UUID) ([]ReplaySession, error) {
	return []ReplaySession{}, nil
}

func (m *MockRepository) GetReplay(ctx context.Context, tenantID, fabricID uuid.UUID, replayID string) (*ReplaySession, error) {
	return &ReplaySession{
		ID:        replayID,
		FabricID:  fabricID.String(),
		Status:    "completed",
		Progress:  100,
		StartedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ExecutePipeline(ctx context.Context, tenantID, fabricID, pipelineID uuid.UUID, input map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"executionId": uuid.New().String(),
		"status":      "completed",
		"output":      map[string]interface{}{"result": "success"},
	}, nil
}

func (m *MockRepository) LogStateFabricAudit(ctx context.Context, event *StateFabricAuditEvent) error {
	return nil
}