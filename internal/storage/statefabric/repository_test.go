package statefabric

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StateFabricRepositoryTestSuite struct {
	suite.Suite
	repo *Repository
	db   *storage.PostgresDB
}

func TestStateFabricRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(StateFabricRepositoryTestSuite))
}

// SetupTest sets up the test database
func (s *StateFabricRepositoryTestSuite) SetupTest() {
	// Use test database configuration - use local Postgres on 5432 per AGENTS.md
	os.Setenv("DB_HOST", getEnvOrDefault("TEST_DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnvOrDefault("TEST_DB_PORT", "5432"))
	os.Setenv("DB_USER", getEnvOrDefault("TEST_DB_USER", "postgres"))
	os.Setenv("DB_PASSWORD", getEnvOrDefault("TEST_DB_PASSWORD", "postgres"))
	os.Setenv("DB_NAME", getEnvOrDefault("TEST_DB_NAME", "functionfly_test"))
	os.Setenv("DB_SSLMODE", "disable")

	// Smaller connection pool for tests
	os.Setenv("DB_MAX_OPEN_CONNS", "5")
	os.Setenv("DB_MAX_IDLE_CONNS", "2")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1m")

	db, err := storage.NewPostgresDB()
	require.NoError(s.T(), err)
	s.db = db

	// Create tables needed for state fabric testing
	err = s.createStateFabricTables(db)
	require.NoError(s.T(), err)

	// Create repository
	s.repo = NewRepository(db.GORM)
}

// TearDownTest cleans up after each test
func (s *StateFabricRepositoryTestSuite) TearDownTest() {
	if s.db != nil {
		s.CleanupTestData()
	}
}

// CleanupTestData removes test data between tests
func (s *StateFabricRepositoryTestSuite) CleanupTestData() {
	tables := []string{
		"state_fabric_events",
		"state_fabric_snapshots",
		"state_fabric_replays",
		"state_fabric_pipelines",
		"state_fabric_stores",
		"state_triggers",
		"state_snapshots",
		"state_events",
		"states",
	}

	for _, table := range tables {
		_, err := s.db.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			s.T().Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// createStateFabricTables creates the tables needed for state fabric testing
func (s *StateFabricRepositoryTestSuite) createStateFabricTables(db *storage.PostgresDB) error {
	// Drop tables if they exist to ensure clean schema
	// Note: state has CASCADE dependencies so we drop it last
	tables := []string{
		"state_fabric_events",
		"state_fabric_snapshots",
		"state_fabric_replays",
		"state_fabric_pipelines",
		"state_fabric_stores",
	}
	for _, table := range tables {
		db.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
	}
	// Drop state_triggers first (depends on state but not CASCADE)
	db.DB.Exec("DROP TABLE IF EXISTS state_triggers CASCADE")
	// Drop state_snapshots and state_events (depend on state via CASCADE)
	db.DB.Exec("DROP TABLE IF EXISTS state_snapshots CASCADE")
	db.DB.Exec("DROP TABLE IF EXISTS state_events CASCADE")
	// Finally drop state (this should be able to proceed now)
	db.DB.Exec("DROP TABLE IF EXISTS states CASCADE")

	// Create state table (matches State model TableName = "states")
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS states (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			full_path VARCHAR(512),
			function_id UUID,
			storage_type VARCHAR(50) NOT NULL DEFAULT 'keyvalue',
			ttl_days INTEGER NOT NULL DEFAULT 0,
			max_size_mb INTEGER NOT NULL DEFAULT 100,
			current_version INTEGER NOT NULL DEFAULT 1,
			is_versioned BOOLEAN NOT NULL DEFAULT true,
			is_encrypted BOOLEAN NOT NULL DEFAULT false,
			is_public BOOLEAN NOT NULL DEFAULT false,
			allow_cross_tenant BOOLEAN NOT NULL DEFAULT false,
			description TEXT,
			storage_used_mb BIGINT NOT NULL DEFAULT 0,
			write_ops_month BIGINT NOT NULL DEFAULT 0,
			read_ops_month BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			last_accessed_at TIMESTAMP NOT NULL DEFAULT NOW(),
			tags JSONB NOT NULL DEFAULT '{}',
			UNIQUE(tenant_id, full_path)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create state table: %w", err)
	}

	// Create state_events table
	_, err = db.DB.Exec(`
		CREATE TABLE state_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
			event_type VARCHAR(50) NOT NULL,
			key VARCHAR(512),
			value JSONB,
			new_value JSONB,
			previous_value JSONB,
			causation_id UUID,
			correlation_id VARCHAR(255) NOT NULL,
			source_type VARCHAR(50) NOT NULL,
			source_id VARCHAR(255) NOT NULL,
			input_hash VARCHAR(128),
			output_hash VARCHAR(128),
			deterministic BOOLEAN NOT NULL DEFAULT false,
			sequence_num BIGINT NOT NULL DEFAULT 0,
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			r2_object_key TEXT,
			r2_bucket VARCHAR(255),
			batch_id VARCHAR(255),
			is_archived BOOLEAN NOT NULL DEFAULT false,
			archived_at TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create state_events table: %w", err)
	}

	// Create state_snapshots table
	_, err = db.DB.Exec(`
		CREATE TABLE state_snapshots (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
			label VARCHAR(255),
			state_data JSONB NOT NULL DEFAULT '{}',
			state_size_bytes BIGINT NOT NULL DEFAULT 0,
			snapshot_version INTEGER NOT NULL DEFAULT 0,
			key_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create state_snapshots table: %w", err)
	}

	// Create state_triggers table
	_, err = db.DB.Exec(`
		CREATE TABLE state_triggers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			source_state_id UUID REFERENCES states(id) ON DELETE CASCADE,
			trigger_type VARCHAR(50) NOT NULL,
			target_function VARCHAR(255),
			condition JSONB NOT NULL DEFAULT '{}',
			key_pattern VARCHAR(512),
			include_new BOOLEAN NOT NULL DEFAULT true,
			max_invocations_per_minute INTEGER NOT NULL DEFAULT 60,
			is_active BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create state_triggers table: %w", err)
	}

	// Create indexes
	_, err = db.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_states_tenant_id ON states(tenant_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	_, err = db.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_state_events_state_id ON state_events(state_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	_, err = db.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_state_snapshots_state_id ON state_snapshots(state_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	_, err = db.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_state_triggers_source_state_id ON state_triggers(source_state_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestCreateFabric tests creating a new state fabric
func (s *StateFabricRepositoryTestSuite) TestCreateFabric() {
	ctx := context.Background()
	tenantID := uuid.New()

	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Test description", "custom", nil)

	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), uuid.Nil, fabric.ID)
	assert.Equal(s.T(), "test-fabric", fabric.Name)
	assert.Equal(s.T(), "Test description", fabric.Description)
	assert.Equal(s.T(), tenantID, fabric.TenantID)
}

// TestGetFabric tests getting a state fabric by ID
func (s *StateFabricRepositoryTestSuite) TestGetFabric() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	created, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Test description", "custom", nil)
	require.NoError(s.T(), err)

	// Get the fabric
	fabric, err := s.repo.GetFabric(ctx, tenantID, created.ID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, fabric.ID)
	assert.Equal(s.T(), "test-fabric", fabric.Name)
}

// TestGetFabric_NotFound tests getting a non-existent fabric
func (s *StateFabricRepositoryTestSuite) TestGetFabric_NotFound() {
	ctx := context.Background()
	tenantID := uuid.New()

	_, err := s.repo.GetFabric(ctx, tenantID, uuid.New())

	assert.Error(s.T(), err)
}

// TestGetFabric_WrongTenant tests getting a fabric with wrong tenant
func (s *StateFabricRepositoryTestSuite) TestGetFabric_WrongTenant() {
	ctx := context.Background()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// Create a fabric with one tenant
	created, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Test description", "custom", nil)
	require.NoError(s.T(), err)

	// Try to get with different tenant
	_, err = s.repo.GetFabric(ctx, otherTenantID, created.ID)

	assert.Error(s.T(), err)
}

// TestUpdateFabric tests updating a state fabric
func (s *StateFabricRepositoryTestSuite) TestUpdateFabric() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	created, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Test description", "custom", nil)
	require.NoError(s.T(), err)

	// Update the fabric
	updates := map[string]interface{}{
		"name":        "updated-fabric",
		"description": "Updated description",
	}
	updated, err := s.repo.UpdateFabric(ctx, tenantID, created.ID, updates)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "updated-fabric", updated.Name)
	assert.Equal(s.T(), "Updated description", updated.Description)
}

// TestDeleteFabric tests deleting a state fabric
func (s *StateFabricRepositoryTestSuite) TestDeleteFabric() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	created, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Test description", "custom", nil)
	require.NoError(s.T(), err)

	// Delete the fabric
	err = s.repo.DeleteFabric(ctx, tenantID, created.ID)
	require.NoError(s.T(), err)

	// Verify it's deleted
	_, err = s.repo.GetFabric(ctx, tenantID, created.ID)
	assert.Error(s.T(), err)
}

// TestListFabrics tests listing state fabrics
func (s *StateFabricRepositoryTestSuite) TestListFabrics() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create multiple fabrics
	_, err := s.repo.CreateFabric(ctx, tenantID, "fabric-1", "Description 1", "custom", nil)
	require.NoError(s.T(), err)
	_, err = s.repo.CreateFabric(ctx, tenantID, "fabric-2", "Description 2", "catalog", nil)
	require.NoError(s.T(), err)
	_, err = s.repo.CreateFabric(ctx, tenantID, "fabric-3", "Description 3", "workflow", nil)
	require.NoError(s.T(), err)

	// List fabrics
	fabrics, total, err := s.repo.ListFabrics(ctx, ListOptions{
		TenantID: tenantID,
		Limit:    10,
		Offset:   0,
	})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), fabrics, 3)
}

// TestListFabrics_WithStatusFilter tests listing with status filter
func (s *StateFabricRepositoryTestSuite) TestListFabrics_WithStatusFilter() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric
	_, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// List fabrics with status filter
	fabrics, _, err := s.repo.ListFabrics(ctx, ListOptions{
		TenantID: tenantID,
		Status:   "online",
		Limit:    10,
		Offset:   0,
	})

	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(fabrics), 0)
}

// TestListFabrics_WithSearch tests listing with search filter
func (s *StateFabricRepositoryTestSuite) TestListFabrics_WithSearch() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create fabrics with different names
	_, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric-1", "Description 1", "custom", nil)
	require.NoError(s.T(), err)
	_, err = s.repo.CreateFabric(ctx, tenantID, "other-fabric", "Description 2", "catalog", nil)
	require.NoError(s.T(), err)

	// Search for "test"
	fabrics, _, err := s.repo.ListFabrics(ctx, ListOptions{
		TenantID: tenantID,
		Search:   "test",
		Limit:    10,
		Offset:   0,
	})

	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(fabrics), 1)
	for _, f := range fabrics {
		assert.Contains(s.T(), f.Name, "test")
	}
}

// TestCreateStore tests creating a store
func (s *StateFabricRepositoryTestSuite) TestCreateStore() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create a store
	store, err := s.repo.CreateStore(ctx, tenantID, fabric.ID, "test-store", "persistent", 1024*1024*1024, "us-east-1")

	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), store.ID)
	assert.Equal(s.T(), "test-store", store.Name)
}

// TestDeleteStore tests deleting a store
func (s *StateFabricRepositoryTestSuite) TestDeleteStore() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Delete the store (which validates the fabric exists)
	err = s.repo.DeleteStore(ctx, tenantID, fabric.ID, "test-store")

	require.NoError(s.T(), err)
}

// TestCreatePipeline tests creating a pipeline
func (s *StateFabricRepositoryTestSuite) TestCreatePipeline() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create a pipeline
	steps := []map[string]interface{}{
		{
			"name": "step1",
			"type": "transform",
		},
	}
	pipeline, err := s.repo.CreatePipeline(ctx, tenantID, fabric.ID, "test-pipeline", "Pipeline description", steps)

	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), pipeline.ID)
	assert.Equal(s.T(), "test-pipeline", pipeline.Name)
	assert.Equal(s.T(), "draft", pipeline.Status)
}

// TestUpdatePipeline tests updating a pipeline
func (s *StateFabricRepositoryTestSuite) TestUpdatePipeline() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric and pipeline
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	pipeline, err := s.repo.CreatePipeline(ctx, tenantID, fabric.ID, "test-pipeline", "Pipeline description", nil)
	require.NoError(s.T(), err)

	pipelineID, err := uuid.Parse(pipeline.ID)
	require.NoError(s.T(), err)

	// Update the pipeline
	updates := map[string]interface{}{
		"name":   "updated-pipeline",
		"status": "active",
	}
	updated, err := s.repo.UpdatePipeline(ctx, tenantID, fabric.ID, pipelineID, updates)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "updated-pipeline", updated.Name)
}

// TestDeletePipeline tests deleting a pipeline
func (s *StateFabricRepositoryTestSuite) TestDeletePipeline() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric and pipeline
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	pipeline, err := s.repo.CreatePipeline(ctx, tenantID, fabric.ID, "test-pipeline", "Pipeline description", nil)
	require.NoError(s.T(), err)

	pipelineID, err := uuid.Parse(pipeline.ID)
	require.NoError(s.T(), err)

	// Delete the pipeline
	err = s.repo.DeletePipeline(ctx, tenantID, fabric.ID, pipelineID)

	require.NoError(s.T(), err)
}

// TestCreateSnapshot tests creating a snapshot
func (s *StateFabricRepositoryTestSuite) TestCreateSnapshot() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create a snapshot
	snapshot, err := s.repo.CreateSnapshot(ctx, tenantID, fabric.ID, "test-snapshot")

	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), snapshot.ID)
	assert.Equal(s.T(), "test-snapshot", snapshot.Name)
	assert.Equal(s.T(), fabric.ID.String(), snapshot.FabricID)
}

// TestDeleteSnapshot tests deleting a snapshot
func (s *StateFabricRepositoryTestSuite) TestDeleteSnapshot() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create a snapshot
	snapshot, err := s.repo.CreateSnapshot(ctx, tenantID, fabric.ID, "test-snapshot")
	require.NoError(s.T(), err)

	snapshotID, err := uuid.Parse(snapshot.ID)
	require.NoError(s.T(), err)

	// Delete the snapshot
	err = s.repo.DeleteSnapshot(ctx, tenantID, fabric.ID, snapshotID)

	require.NoError(s.T(), err)
}

// TestListSnapshots tests listing snapshots
func (s *StateFabricRepositoryTestSuite) TestListSnapshots() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create snapshots
	_, err = s.repo.CreateSnapshot(ctx, tenantID, fabric.ID, "snapshot-1")
	require.NoError(s.T(), err)
	_, err = s.repo.CreateSnapshot(ctx, tenantID, fabric.ID, "snapshot-2")
	require.NoError(s.T(), err)

	// List snapshots
	snapshots, err := s.repo.ListSnapshots(ctx, tenantID, fabric.ID)

	require.NoError(s.T(), err)
	assert.Len(s.T(), snapshots, 2)
}

// TestCreateReplay tests creating a replay
func (s *StateFabricRepositoryTestSuite) TestCreateReplay() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create a snapshot
	snapshot, err := s.repo.CreateSnapshot(ctx, tenantID, fabric.ID, "test-snapshot")
	require.NoError(s.T(), err)

	// Create a replay
	req := ReplayCreateRequest{
		SnapshotID:   snapshot.ID,
		StartEventID: "",
		EndEventID:   "",
	}
	replay, err := s.repo.CreateReplay(ctx, tenantID, fabric.ID, req)

	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), replay.ID)
	assert.Equal(s.T(), "completed", replay.Status)
	assert.Equal(s.T(), fabric.ID.String(), replay.FabricID)
}

// TestGetReplay tests getting a replay by ID
func (s *StateFabricRepositoryTestSuite) TestGetReplay() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Create a snapshot
	snapshot, err := s.repo.CreateSnapshot(ctx, tenantID, fabric.ID, "test-snapshot")
	require.NoError(s.T(), err)

	// Create a replay
	req := ReplayCreateRequest{
		SnapshotID: snapshot.ID,
	}
	created, err := s.repo.CreateReplay(ctx, tenantID, fabric.ID, req)
	require.NoError(s.T(), err)

	// Get the replay
	replay, err := s.repo.GetReplay(ctx, tenantID, fabric.ID, created.ID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, replay.ID)
}

// TestGetReplay_NotFound tests getting a non-existent replay
func (s *StateFabricRepositoryTestSuite) TestGetReplay_NotFound() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Try to get non-existent replay
	_, err = s.repo.GetReplay(ctx, tenantID, fabric.ID, "non-existent-id")

	assert.Error(s.T(), err)
}

// TestListEvents tests listing events
func (s *StateFabricRepositoryTestSuite) TestListEvents() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// List events (should be empty initially)
	events, total, err := s.repo.ListEvents(ctx, tenantID, fabric.ID, EventListOptions{
		Limit: 10,
	})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), total)
	assert.Len(s.T(), events, 0)
}

// TestListEvents_WithFilters tests listing events with filters
func (s *StateFabricRepositoryTestSuite) TestListEvents_WithFilters() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)

	// List events with time filters
	events, _, err := s.repo.ListEvents(ctx, tenantID, fabric.ID, EventListOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     10,
	})

	require.NoError(s.T(), err)
	// Events list may be empty but should not error
	assert.NotNil(s.T(), events)
}

// TestGetMetrics tests getting fabric metrics
func (s *StateFabricRepositoryTestSuite) TestGetMetrics() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	fabric, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Get metrics
	metrics, err := s.repo.GetMetrics(ctx, fabric.ID, "")

	require.NoError(s.T(), err)
	assert.NotNil(s.T(), metrics)
	assert.GreaterOrEqual(s.T(), metrics.TotalOperations, int64(0))
}

// TestStats tests getting stats
func (s *StateFabricRepositoryTestSuite) TestStats() {
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a fabric first
	_, err := s.repo.CreateFabric(ctx, tenantID, "test-fabric", "Description", "custom", nil)
	require.NoError(s.T(), err)

	// Get stats
	stats, err := s.repo.Stats(ctx)

	require.NoError(s.T(), err)
	assert.NotNil(s.T(), stats)
	assert.GreaterOrEqual(s.T(), stats["totalFabrics"], int64(1))
}

// Helper function to print JSON for debugging
func printJSON(t *testing.T, name string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("Error marshaling %s: %v", name, err)
		return
	}
	t.Logf("%s: %s", name, string(data))
}
