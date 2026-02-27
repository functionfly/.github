package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LocalRuntimeTestSuite struct {
	test.TestSuite
	repo *LocalRuntimeRepository
}

func TestLocalRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(LocalRuntimeTestSuite))
}

// SetupTest sets up the test database
func (s *LocalRuntimeTestSuite) SetupTest() {
	// Call parent setup
	s.TestSuite.SetupTest()

	// Set up test database
	db, err := s.setupTestDatabase()
	require.NoError(s.T(), err)
	s.SetDB(db) // Set the parent TestSuite db field

	// For testing local runtime repository, we only need the local runtime tables
	// Create them directly instead of running all migrations
	err = s.createLocalRuntimeTables(db)
	require.NoError(s.T(), err)

	// Create repository
	s.repo = NewLocalRuntimeRepository(db)
}

// setupTestDatabase creates a test database connection
func (s *LocalRuntimeTestSuite) setupTestDatabase() (*PostgresDB, error) {
	// Use test database configuration
	os.Setenv("DB_HOST", getEnvOrDefault("TEST_DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnvOrDefault("TEST_DB_PORT", "5434")) // Use port 5434 for docker-compose postgres
	os.Setenv("DB_USER", getEnvOrDefault("TEST_DB_USER", "postgres"))
	os.Setenv("DB_PASSWORD", getEnvOrDefault("TEST_DB_PASSWORD", "postgres"))
	os.Setenv("DB_NAME", getEnvOrDefault("TEST_DB_NAME", "functionfly_test")) // Use a separate test database
	os.Setenv("DB_SSLMODE", "disable")

	// Smaller connection pool for tests
	os.Setenv("DB_MAX_OPEN_CONNS", "5")
	os.Setenv("DB_MAX_IDLE_CONNS", "2")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1m")

	return NewPostgresDB()
}

// CleanupTestData removes test data between tests
func (s *LocalRuntimeTestSuite) CleanupTestData() {
	db := s.DB().(*PostgresDB)
	// Truncate all local runtime tables to clean up test data
	tables := []string{
		"local_runtime_health",
		"local_runtime_metrics",
		"local_runtime_instances",
	}

	for _, table := range tables {
		_, err := db.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			s.T().Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// createLocalRuntimeTables creates only the tables needed for local runtime testing
func (s *LocalRuntimeTestSuite) createLocalRuntimeTables(db *PostgresDB) error {
	// Drop tables if they exist to ensure clean schema
	tables := []string{"local_runtime_health", "local_runtime_metrics", "local_runtime_instances"}
	for _, table := range tables {
		_, err := db.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Create local runtime instances table
	_, err := db.DB.Exec(`
		CREATE TABLE local_runtime_instances (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			runtime_id VARCHAR(255) UNIQUE NOT NULL,
			runtime_type VARCHAR(50) NOT NULL,
			function_name VARCHAR(255) NOT NULL,
			manifest_path TEXT NOT NULL,
			host VARCHAR(255) NOT NULL,
			port INTEGER NOT NULL,
			pid INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'running',
			last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			uptime BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return err
	}

	// Create indexes
	_, err = db.DB.Exec(`
		CREATE INDEX idx_local_runtime_instances_last_heartbeat ON local_runtime_instances(last_heartbeat);
		CREATE INDEX idx_local_runtime_instances_status ON local_runtime_instances(status);
		CREATE INDEX idx_local_runtime_instances_runtime_type ON local_runtime_instances(runtime_type);
	`)
	if err != nil {
		return err
	}

	// Create local runtime metrics table
	_, err = db.DB.Exec(`
		CREATE TABLE local_runtime_metrics (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			runtime_instance_id UUID NOT NULL REFERENCES local_runtime_instances(id) ON DELETE CASCADE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			memory_usage JSONB NOT NULL,
			cpu_usage DOUBLE PRECISION NOT NULL,
			active_connections INTEGER NOT NULL DEFAULT 0,
			request_throughput DOUBLE PRECISION NOT NULL DEFAULT 0,
			total_requests BIGINT NOT NULL DEFAULT 0,
			error_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
			execution_count BIGINT NOT NULL DEFAULT 0,
			average_latency BIGINT NOT NULL DEFAULT 0,
			error_count BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return err
	}

	// Create indexes for metrics
	_, err = db.DB.Exec(`
		CREATE INDEX idx_local_runtime_metrics_instance_timestamp ON local_runtime_metrics(runtime_instance_id, timestamp DESC);
		CREATE INDEX idx_local_runtime_metrics_timestamp ON local_runtime_metrics(timestamp DESC);
	`)
	if err != nil {
		return err
	}

	// Create local runtime health table
	_, err = db.DB.Exec(`
		CREATE TABLE local_runtime_health (
			runtime_instance_id UUID PRIMARY KEY REFERENCES local_runtime_instances(id) ON DELETE CASCADE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			status VARCHAR(50) NOT NULL,
			response_time BIGINT NOT NULL DEFAULT 0,
			checks JSONB NOT NULL DEFAULT '{}',
			error TEXT
		)`)
	if err != nil {
		return err
	}

	// Create index for health
	_, err = db.DB.Exec(`
		CREATE INDEX idx_local_runtime_health_timestamp ON local_runtime_health(timestamp DESC);
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalRuntimeTestSuite) TestRegisterLocalRuntime() {
	ctx := context.Background()

	runtimeID := "test-register-" + uuid.New().String()[:8]
	instance := &LocalRuntimeInstance{
		RuntimeID:     runtimeID,
		RuntimeType:   "node18",
		FunctionName:  "test-function",
		ManifestPath:  "/path/to/manifest",
		Host:          "localhost",
		Port:          8080,
		PID:           12345,
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        300,
	}

	// Register the runtime instance
	created, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), created)
	assert.NotEqual(s.T(), uuid.Nil, created.ID)
	assert.Equal(s.T(), runtimeID, created.RuntimeID)

	// Verify it was created by retrieving it
	retrieved, err := s.repo.GetLocalRuntimeByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), runtimeID, retrieved.RuntimeID)
}

func (s *LocalRuntimeTestSuite) TestGetLocalRuntimeByRuntimeID() {
	ctx := context.Background()

	runtimeID := "test-by-id-" + uuid.New().String()[:8]
	instance := &LocalRuntimeInstance{
		RuntimeID:     runtimeID,
		RuntimeType:   "node20",
		FunctionName:  "test-function-2",
		ManifestPath:  "/path/to/manifest2",
		Host:          "localhost",
		Port:          8081,
		PID:           12346,
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        150,
	}

	// Register the runtime instance
	created, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)

	// Retrieve by runtime ID
	retrieved, err := s.repo.GetLocalRuntimeByRuntimeID(ctx, runtimeID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), runtimeID, retrieved.RuntimeID)
}

func (s *LocalRuntimeTestSuite) TestUpdateLocalRuntimeHeartbeat() {
	ctx := context.Background()

	runtimeID := "test-heartbeat-" + uuid.New().String()[:8]
	instance := &LocalRuntimeInstance{
		RuntimeID:     runtimeID,
		RuntimeType:   "python3.11",
		FunctionName:  "test-function-3",
		ManifestPath:  "/path/to/manifest3",
		Host:          "localhost",
		Port:          8082,
		PID:           12347,
		Status:        "running",
		LastHeartbeat: time.Now().Add(-time.Hour), // Old heartbeat
		Uptime:        3600,
	}

	// Register the runtime instance
	created, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)

	// Update heartbeat
	err = s.repo.UpdateLocalRuntimeHeartbeat(ctx, runtimeID)
	require.NoError(s.T(), err)

	// Verify heartbeat was updated
	retrieved, err := s.repo.GetLocalRuntimeByRuntimeID(ctx, runtimeID)
	require.NoError(s.T(), err)
	assert.True(s.T(), retrieved.LastHeartbeat.After(created.LastHeartbeat))
	assert.True(s.T(), retrieved.UpdatedAt.After(created.UpdatedAt))
}

func (s *LocalRuntimeTestSuite) TestListActiveLocalRuntimes() {
	ctx := context.Background()

	activeRuntimeID := "active-runtime-" + uuid.New().String()[:8]
	inactiveRuntimeID := "inactive-runtime-" + uuid.New().String()[:8]

	// Create active runtime (recent heartbeat)
	activeInstance := &LocalRuntimeInstance{
		RuntimeID:     activeRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "active-function",
		ManifestPath:  "/path/to/active",
		Host:          "localhost",
		Port:          8083,
		PID:           12348,
		Status:        "running",
		LastHeartbeat: time.Now(), // Recent heartbeat
		Uptime:        600,
	}

	// Create inactive runtime (old heartbeat)
	inactiveInstance := &LocalRuntimeInstance{
		RuntimeID:     inactiveRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "inactive-function",
		ManifestPath:  "/path/to/inactive",
		Host:          "localhost",
		Port:          8084,
		PID:           12349,
		Status:        "running",
		LastHeartbeat: time.Now().Add(-10 * time.Minute), // Old heartbeat (more than 5 minutes)
		Uptime:        600,
	}

	// Register both instances
	_, err := s.repo.RegisterLocalRuntime(ctx, activeInstance)
	require.NoError(s.T(), err)
	_, err = s.repo.RegisterLocalRuntime(ctx, inactiveInstance)
	require.NoError(s.T(), err)

	// List active runtimes
	activeRuntimes, err := s.repo.ListActiveLocalRuntimes(ctx)
	require.NoError(s.T(), err)

	// Should only return the active runtime
	assert.Len(s.T(), activeRuntimes, 1)
	assert.Equal(s.T(), activeRuntimeID, activeRuntimes[0].RuntimeID)
}

func (s *LocalRuntimeTestSuite) TestCleanupStaleLocalRuntimes() {
	ctx := context.Background()

	staleRuntimeID := "stale-runtime-" + uuid.New().String()[:8]

	// Create stale runtime (very old heartbeat)
	staleInstance := &LocalRuntimeInstance{
		RuntimeID:     staleRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "stale-function",
		ManifestPath:  "/path/to/stale",
		Host:          "localhost",
		Port:          8085,
		PID:           12350,
		Status:        "running",
		LastHeartbeat: time.Now().Add(-time.Hour), // Very old heartbeat
		Uptime:        600,
	}

	// Register the stale instance
	_, err := s.repo.RegisterLocalRuntime(ctx, staleInstance)
	require.NoError(s.T(), err)

	// Clean up runtimes older than 30 minutes
	deletedCount, err := s.repo.CleanupStaleLocalRuntimes(ctx, 30*time.Minute)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), deletedCount)

	// Verify it was deleted
	retrieved, err := s.repo.GetLocalRuntimeByRuntimeID(ctx, staleRuntimeID)
	assert.Error(s.T(), err) // Should return error (not found)
	assert.Nil(s.T(), retrieved)
}

func (s *LocalRuntimeTestSuite) TestRecordLocalRuntimeMetrics() {
	ctx := context.Background()

	metricsRuntimeID := "metrics-runtime-" + uuid.New().String()[:8]

	// First create a runtime instance
	instance := &LocalRuntimeInstance{
		RuntimeID:     metricsRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "metrics-function",
		ManifestPath:  "/path/to/metrics",
		Host:          "localhost",
		Port:          8086,
		PID:           12351,
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        300,
	}

	createdInstance, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)

	// Record metrics
	metrics := &LocalRuntimeMetric{
		RuntimeInstanceID: createdInstance.ID,
		Timestamp:         time.Now(),
		MemoryUsage: MemoryStats{
			Heap:   1024 * 1024, // 1MB
			Stack:  512 * 1024,  // 512KB
			System: 2048 * 1024, // 2MB
		},
		CPUUsage:          45.5,
		ActiveConnections: 10,
		RequestThroughput: 100.5,
		TotalRequests:     1000,
		ErrorRate:         2.5,
		ExecutionCount:    950,
		AverageLatency:    50 * time.Millisecond,
		ErrorCount:        25,
	}

	err = s.repo.RecordLocalRuntimeMetrics(ctx, metrics)
	require.NoError(s.T(), err)

	// Verify metrics were recorded
	retrievedMetrics, err := s.repo.GetLocalRuntimeMetrics(ctx, createdInstance.ID, time.Now().Add(-time.Hour), 10)
	require.NoError(s.T(), err)
	assert.Len(s.T(), retrievedMetrics, 1)
	assert.Equal(s.T(), createdInstance.ID, retrievedMetrics[0].RuntimeInstanceID)
	assert.Equal(s.T(), float64(45.5), retrievedMetrics[0].CPUUsage)
}

func (s *LocalRuntimeTestSuite) TestGetLatestLocalRuntimeMetrics() {
	ctx := context.Background()

	latestMetricsRuntimeID := "latest-metrics-runtime-" + uuid.New().String()[:8]

	// First create a runtime instance
	instance := &LocalRuntimeInstance{
		RuntimeID:     latestMetricsRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "latest-metrics-function",
		ManifestPath:  "/path/to/latest",
		Host:          "localhost",
		Port:          8087,
		PID:           12352,
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        300,
	}

	createdInstance, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)

	// Record multiple metrics
	for i := 0; i < 3; i++ {
		metrics := &LocalRuntimeMetric{
			RuntimeInstanceID: createdInstance.ID,
			Timestamp:         time.Now().Add(time.Duration(i) * time.Minute),
			MemoryUsage: MemoryStats{
				Heap:   uint64(1000 + i*100),
				Stack:  uint64(500 + i*50),
				System: uint64(2000 + i*200),
			},
			CPUUsage:          float64(40 + i),
			ActiveConnections: 5 + i,
			RequestThroughput: 90 + float64(i),
			TotalRequests:     int64(800 + i*100),
			ErrorRate:         float64(i),
			ExecutionCount:    int64(750 + i*50),
			AverageLatency:    time.Duration(40+i*5) * time.Millisecond,
			ErrorCount:        int64(i * 10),
		}

		err = s.repo.RecordLocalRuntimeMetrics(ctx, metrics)
		require.NoError(s.T(), err)
	}

	// Get latest metrics
	latestMetrics, err := s.repo.GetLatestLocalRuntimeMetrics(ctx, createdInstance.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), latestMetrics)
	assert.Equal(s.T(), createdInstance.ID, latestMetrics.RuntimeInstanceID)
	assert.Equal(s.T(), float64(42), latestMetrics.CPUUsage) // Should be the highest CPU usage (40+2)
}

func (s *LocalRuntimeTestSuite) TestRecordLocalRuntimeHealth() {
	ctx := context.Background()

	healthRuntimeID := "health-runtime-" + uuid.New().String()[:8]

	// First create a runtime instance
	instance := &LocalRuntimeInstance{
		RuntimeID:     healthRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "health-function",
		ManifestPath:  "/path/to/health",
		Host:          "localhost",
		Port:          8088,
		PID:           12353,
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        300,
	}

	createdInstance, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)

	// Record health
	health := &LocalRuntimeHealth{
		RuntimeInstanceID: createdInstance.ID,
		Timestamp:         time.Now(),
		Status:            "healthy",
		ResponseTime:      25 * time.Millisecond,
		Checks: JSONMap{
			"function_loaded": map[string]interface{}{
				"status": "healthy",
				"detail": "Function code is loaded",
			},
			"memory_usage": map[string]interface{}{
				"status": "healthy",
				"detail": "Memory usage within limits",
			},
		},
	}

	err = s.repo.RecordLocalRuntimeHealth(ctx, health)
	require.NoError(s.T(), err)

	// Verify health was recorded
	retrievedHealth, err := s.repo.GetLocalRuntimeHealth(ctx, createdInstance.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "healthy", retrievedHealth.Status)
	assert.Equal(s.T(), 25*time.Millisecond, retrievedHealth.ResponseTime)
	assert.NotNil(s.T(), retrievedHealth.Checks)
}

func (s *LocalRuntimeTestSuite) TestDeregisterLocalRuntime() {
	ctx := context.Background()

	deregisterRuntimeID := "deregister-test-" + uuid.New().String()[:8]

	instance := &LocalRuntimeInstance{
		RuntimeID:     deregisterRuntimeID,
		RuntimeType:   "node18",
		FunctionName:  "deregister-function",
		ManifestPath:  "/path/to/deregister",
		Host:          "localhost",
		Port:          8089,
		PID:           12354,
		Status:        "running",
		LastHeartbeat: time.Now(),
		Uptime:        300,
	}

	// Register the runtime instance
	created, err := s.repo.RegisterLocalRuntime(ctx, instance)
	require.NoError(s.T(), err)

	// Verify it exists
	retrieved, err := s.repo.GetLocalRuntimeByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), retrieved)

	// Deregister it
	err = s.repo.DeregisterLocalRuntime(ctx, deregisterRuntimeID)
	require.NoError(s.T(), err)

	// Verify it was deleted
	retrieved, err = s.repo.GetLocalRuntimeByID(ctx, created.ID)
	assert.Error(s.T(), err) // Should return error (not found)
	assert.Nil(s.T(), retrieved)
}

func TestMemoryStatsJSON(t *testing.T) {
	// Test that MemoryStats can be marshaled to JSON
	stats := MemoryStats{
		Heap:   1024,
		Stack:  512,
		System: 2048,
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var decoded MemoryStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, stats, decoded)
}

func TestLocalRuntimeHealth(t *testing.T) {
	// Test LocalRuntimeHealth structure
	now := time.Now()
	health := LocalRuntimeHealth{
		RuntimeInstanceID: uuid.New(),
		Timestamp:         now,
		Status:            "healthy",
		ResponseTime:      50 * time.Millisecond,
		Checks: JSONMap{
			"function_loaded": map[string]interface{}{
				"status": "healthy",
				"detail": "Function code is loaded",
			},
		},
	}

	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, 50*time.Millisecond, health.ResponseTime)
	assert.NotNil(t, health.Checks)
}
