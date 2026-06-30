package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/functionfly/functionfly/internal/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PostgresTestSuite struct {
	test.TestSuite
	db *PostgresDB
}

func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

// SetupTest sets up the test database
func (s *PostgresTestSuite) SetupTest() {
	// Call parent setup
	s.TestSuite.SetupTest()

	// Set up test database
	db, err := s.setupTestDatabase()
	require.NoError(s.T(), err)
	s.db = db
	s.SetDB(db) // Set the parent TestSuite db field

	// Run migrations
	err = RunMigrations(db)
	require.NoError(s.T(), err)
}

// setupTestDatabase creates a test database connection
func (s *PostgresTestSuite) setupTestDatabase() (*PostgresDB, error) {
	// Use test database configuration
	os.Setenv("DB_HOST", getEnvOrDefault("TEST_DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnvOrDefault("TEST_DB_PORT", "5434"))
	os.Setenv("DB_USER", getEnvOrDefault("TEST_DB_USER", "postgres"))
	os.Setenv("DB_PASSWORD", getEnvOrDefault("TEST_DB_PASSWORD", "postgres"))
	os.Setenv("DB_NAME", getEnvOrDefault("TEST_DB_NAME", "functionfly_test"))
	os.Setenv("DB_SSLMODE", "disable")

	// Smaller connection pool for tests
	os.Setenv("DB_MAX_OPEN_CONNS", "5")
	os.Setenv("DB_MAX_IDLE_CONNS", "2")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1m")

	return NewPostgresDB()
}

// CleanupTestData removes test data between tests
func (s *PostgresTestSuite) CleanupTestData() {
	db := s.DB().(*PostgresDB)
	// Truncate all tables to clean up test data
	tables := []string{
		"audit_events",
		"content_changes",
		"changelog_entries",
		"blog_posts",
		"feedback",
		"invoices",
		"subscriptions",
		"coupons",
		"billing_tiers",
		"users",
		"tenants",
		"backends",
		"apps",
		"health_checks",
		"circuit_states",
		"routing_events",
	}

	for _, table := range tables {
		_, err := db.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			s.T().Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}


// Convenience methods for test repository operations
func (s *PostgresTestSuite) createTenant(name, domain, adminEmail string) (string, error) {
	repo := s.db.Repository()
	ctx := context.Background()

	// Create tenant
	tenant, err := repo.CreateTenant(ctx, name)
	if err != nil {
		return "", err
	}

	// Update tenant with domain (if supported)
	// Note: This might need adjustment based on actual interface
	return tenant.ID.String(), nil
}

func (s *PostgresTestSuite) getTenant(id string) (*Tenant, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	tenantID, err := parseUUIDLocal(id)
	if err != nil {
		return nil, err
	}
	return repo.GetTenantByID(ctx, tenantID)
}

func (s *PostgresTestSuite) listTenants() ([]*Tenant, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	return repo.ListTenants(ctx)
}

func (s *PostgresTestSuite) updateTenant(id, name, domain string) error {
	repo := s.db.Repository()
	ctx := context.Background()
	tenantID, err := parseUUIDLocal(id)
	if err != nil {
		return err
	}
	_, err = repo.UpdateTenant(ctx, tenantID, map[string]interface{}{
		"name":   name,
		"domain": domain,
	})
	return err
}

func (s *PostgresTestSuite) deleteTenant(id string) error {
	repo := s.db.Repository()
	ctx := context.Background()
	tenantID, err := parseUUIDLocal(id)
	if err != nil {
		return err
	}
	return repo.DeleteTenant(ctx, tenantID)
}

func (s *PostgresTestSuite) createUser(tenantID, username, email, hashedPassword string, isAdmin bool) (string, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	user, err := repo.CreateUser(ctx, email, hashedPassword, parseUUIDMust(tenantID))
	if err != nil {
		return "", err
	}
	return user.ID.String(), nil
}

func (s *PostgresTestSuite) getUser(id string) (*User, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	userID, err := parseUUIDLocal(id)
	if err != nil {
		return nil, err
	}
	return repo.GetUserByID(ctx, userID)
}

func (s *PostgresTestSuite) authenticateUser(username, passwordHash string) (*User, error) {
	// Simple implementation - find user by email and check password
	repo := s.db.Repository()
	ctx := context.Background()
	user, err := repo.GetUserByEmail(ctx, username) // Assuming username is email for now
	if err != nil {
		return nil, err
	}
	// In a real implementation, we'd check password hash here
	if user.PasswordHash != passwordHash {
		return nil, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

func (s *PostgresTestSuite) createApp(tenantID, name, description string) (string, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	app, err := repo.CreateApp(ctx, name, name, parseUUIDMust(tenantID)) // Using name as slug
	if err != nil {
		return "", err
	}
	return app.ID.String(), nil
}

func (s *PostgresTestSuite) getApp(id string) (*App, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	appID, err := parseUUIDLocal(id)
	if err != nil {
		return nil, err
	}
	return repo.GetAppByID(ctx, appID)
}

func (s *PostgresTestSuite) listAppsByTenant(tenantID string) ([]*App, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	tID, err := parseUUIDLocal(tenantID)
	if err != nil {
		return nil, err
	}
	return repo.ListAppsByTenant(ctx, tID)
}

func (s *PostgresTestSuite) createBackend(appID, name, provider, config string, enabled bool) (string, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	appUUID := parseUUIDMust(appID)
	backend, err := repo.CreateBackend(ctx, appUUID, provider, provider, "", "", nil) // Simplified
	if err != nil {
		return "", err
	}
	return backend.ID.String(), nil
}

func (s *PostgresTestSuite) getBackend(id string) (*Backend, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	backendID, err := parseUUIDLocal(id)
	if err != nil {
		return nil, err
	}
	return repo.GetBackendByID(ctx, backendID)
}

func (s *PostgresTestSuite) getAllEnabledBackends() ([]*Backend, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	return repo.GetAllEnabledBackends(ctx)
}

func (s *PostgresTestSuite) recordHealthCheck(backendID string, ok bool, latencyMs int, message string) error {
	repo := s.db.Repository()
	ctx := context.Background()
	backendUUID := parseUUIDMust(backendID)
	return repo.InsertHealthCheck(ctx, backendUUID, ok, 200, latencyMs, message)
}

func (s *PostgresTestSuite) getRecentHealthChecks(backendID string, limit int) ([]*HealthCheck, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	backendUUID := parseUUIDMust(backendID)
	return repo.GetRecentHealthChecks(ctx, backendUUID, limit)
}

func (s *PostgresTestSuite) updateCircuitState(backendID string, success bool) error {
	repo := s.db.Repository()
	ctx := context.Background()
	backendUUID := parseUUIDMust(backendID)
	state := &CircuitState{
		BackendID: backendUUID,
	}
	if success {
		state.SuccessCount = 1
	} else {
		state.FailCount = 1
	}
	return repo.UpsertCircuitState(ctx, state)
}

func (s *PostgresTestSuite) getCircuitState(backendID string) (*CircuitState, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	backendUUID := parseUUIDMust(backendID)
	return repo.GetCircuitState(ctx, backendUUID)
}

// Transactional versions
func (s *PostgresTestSuite) createTenantTx(tx *sql.Tx, name, domain, adminEmail string) (string, error) {
	// For now, implement without transaction support
	return s.createTenant(name, domain, adminEmail)
}

func (s *PostgresTestSuite) createAppTx(tx *sql.Tx, tenantID, name, description string) (string, error) {
	// For now, implement without transaction support
	return s.createApp(tenantID, name, description)
}

// Helper functions
func parseUUIDLocal(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parseUUIDMust(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *PostgresTestSuite) TestNewPostgresDB() {
	db := s.DB().(*PostgresDB)
	assert.NotNil(s.T(), db)
	assert.NotNil(s.T(), db.DB)

	// Test connection
	err := db.DB.Ping()
	assert.NoError(s.T(), err)
}

func (s *PostgresTestSuite) TestConnectionPooling() {
	db := s.DB().(*PostgresDB)

	// Check connection pool settings
	stats := db.DB.Stats()
	assert.Greater(s.T(), stats.MaxOpenConnections, 0)
	assert.GreaterOrEqual(s.T(), stats.Idle, 0)
}

func (s *PostgresTestSuite) TestRepository() {
	repo := s.DB().(*PostgresDB).Repository()
	assert.NotNil(s.T(), repo)
}

func (s *PostgresTestSuite) TestCreateAndGetTenant() {
	// Create tenant
	tenantID, err := s.createTenant("Test Tenant", "test.example.com", "admin@test.com")
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), tenantID)

	// Get tenant
	tenant, err := s.getTenant(tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test Tenant", tenant.Name)
}

func (s *PostgresTestSuite) TestListTenants() {
	// Create multiple tenants
	tenantID1, err := s.createTenant("Tenant 1", "tenant1.com", "admin1@test.com")
	require.NoError(s.T(), err)

	tenantID2, err := s.createTenant("Tenant 2", "tenant2.com", "admin2@test.com")
	require.NoError(s.T(), err)

	// List all tenants
	tenants, err := s.listTenants()
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(tenants), 2)

	// Verify our tenants are in the list
	tenantIDs := make(map[string]bool)
	for _, tenant := range tenants {
		tenantIDs[tenant.ID.String()] = true
	}

	assert.True(s.T(), tenantIDs[tenantID1])
	assert.True(s.T(), tenantIDs[tenantID2])
}

func (s *PostgresTestSuite) TestUpdateTenant() {
	// Create tenant
	tenantID, err := s.createTenant("Original Name", "original.com", "admin@test.com")
	require.NoError(s.T(), err)

	// Update tenant
	err = s.updateTenant(tenantID, "Updated Name", "updated.com")
	require.NoError(s.T(), err)

	// Verify update
	tenant, err := s.getTenant(tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Name", tenant.Name)
}

func (s *PostgresTestSuite) TestDeleteTenant() {
	// Create tenant
	tenantID, err := s.createTenant("Delete Me", "delete.com", "admin@test.com")
	require.NoError(s.T(), err)

	// Verify it exists
	_, err = s.getTenant(tenantID)
	require.NoError(s.T(), err)

	// Delete tenant
	err = s.deleteTenant(tenantID)
	require.NoError(s.T(), err)

	// Verify it's gone
	_, err = s.getTenant(tenantID)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), sql.ErrNoRows, err)
}

func (s *PostgresTestSuite) TestCreateAndGetUser() {
	// Create tenant first
	tenantID, err := s.createTenant("User Test", "user.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	// Create user
	userID, err := s.createUser(tenantID, "testuser", "test@example.com", "hashedpass", false)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), userID)

	// Get user
	user, err := s.getUser(userID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test@example.com", user.Email)
	assert.Equal(s.T(), parseUUIDMust(tenantID), user.TenantID)
}

func (s *PostgresTestSuite) TestAuthenticateUser() {
	// Create tenant and user
	tenantID, err := s.createTenant("Auth Test", "auth.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	hashedPassword := "hashedpassword123"
	_, err = s.createUser(tenantID, "authuser", "auth@example.com", hashedPassword, false)
	require.NoError(s.T(), err)

	// Test successful authentication
	user, err := s.authenticateUser("auth@example.com", hashedPassword)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "auth@example.com", user.Email)

	// Test failed authentication
	_, err = s.authenticateUser("auth@example.com", "wrongpassword")
	assert.Error(s.T(), err)

	// Test non-existent user
	_, err = s.authenticateUser("nonexistent", hashedPassword)
	assert.Error(s.T(), err)
}

func (s *PostgresTestSuite) TestCreateAndGetApp() {
	// Create tenant and app
	tenantID, err := s.createTenant("App Test", "app.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	appID, err := s.createApp(tenantID, "Test App", "A test application")
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), appID)

	// Get app
	app, err := s.getApp(appID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test App", app.Name)
	assert.Equal(s.T(), parseUUIDMust(tenantID), app.TenantID)
}

func (s *PostgresTestSuite) TestListAppsByTenant() {
	// Create tenant
	tenantID, err := s.createTenant("Apps Test", "apps.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	// Create multiple apps
	appID1, err := s.createApp(tenantID, "App 1", "First app")
	require.NoError(s.T(), err)

	appID2, err := s.createApp(tenantID, "App 2", "Second app")
	require.NoError(s.T(), err)

	// List apps
	apps, err := s.listAppsByTenant(tenantID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(apps), 2)

	// Verify our apps are in the list
	appIDs := make(map[string]bool)
	for _, app := range apps {
		appIDs[app.ID.String()] = true
	}

	assert.True(s.T(), appIDs[appID1])
	assert.True(s.T(), appIDs[appID2])
}

func (s *PostgresTestSuite) TestCreateAndGetBackend() {
	// Create tenant, app, and backend
	tenantID, err := s.createTenant("Backend Test", "backend.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	appID, err := s.createApp(tenantID, "Test App", "App for backend testing")
	require.NoError(s.T(), err)

	config := `{"url": "https://api.example.com", "token": "secret"}`
	backendID, err := s.createBackend(appID, "test-backend", "vercel", config, true)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), backendID)

	// Get backend
	backend, err := s.getBackend(backendID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "vercel", backend.Provider)
	assert.Equal(s.T(), parseUUIDMust(appID), backend.AppID)
}

func (s *PostgresTestSuite) TestGetAllEnabledBackends() {
	// Create tenant and app
	tenantID, err := s.createTenant("Backends Test", "backends.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	appID, err := s.createApp(tenantID, "Test App", "App for backends testing")
	require.NoError(s.T(), err)

	// Create enabled backend
	_, err = s.createBackend(appID, "enabled-backend", "vercel", "{}", true)
	require.NoError(s.T(), err)

	// Create disabled backend
	_, err = s.createBackend(appID, "disabled-backend", "fly", "{}", false)
	require.NoError(s.T(), err)

	// Get all enabled backends
	backends, err := s.getAllEnabledBackends()
	require.NoError(s.T(), err)

	// Should have at least our enabled backend
	found := false
	for _, backend := range backends {
		if backend.Provider == "vercel" { // We use provider as name for simplicity
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Enabled backend should be found")
}

func (s *PostgresTestSuite) TestHealthChecks() {
	// Create backend
	tenantID, err := s.createTenant("Health Test", "health.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	appID, err := s.createApp(tenantID, "Test App", "App for health testing")
	require.NoError(s.T(), err)

	backendID, err := s.createBackend(appID, "health-backend", "vercel", "{}", true)
	require.NoError(s.T(), err)

	// Record health checks
	err = s.recordHealthCheck(backendID, true, 150, "OK")
	require.NoError(s.T(), err)

	err = s.recordHealthCheck(backendID, false, 0, "Connection timeout")
	require.NoError(s.T(), err)

	// Get recent health checks
	checks, err := s.getRecentHealthChecks(backendID, 10)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(checks), 2)

	// Check the most recent check
	recent := checks[0]
	assert.Equal(s.T(), parseUUIDMust(backendID), recent.BackendID)
	assert.True(s.T(), recent.OK)
	assert.Equal(s.T(), 150, recent.LatencyMs)
	assert.Empty(s.T(), recent.ErrorMessage)
}

func (s *PostgresTestSuite) TestCircuitBreaker() {
	// Create backend
	tenantID, err := s.createTenant("Circuit Test", "circuit.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	appID, err := s.createApp(tenantID, "Test App", "App for circuit testing")
	require.NoError(s.T(), err)

	backendID, err := s.createBackend(appID, "circuit-backend", "vercel", "{}", true)
	require.NoError(s.T(), err)

	// Update circuit state - simulate failures
	for i := 0; i < 5; i++ {
		err = s.updateCircuitState(backendID, false)
		require.NoError(s.T(), err)
	}

	// Update circuit state - simulate success
	err = s.updateCircuitState(backendID, true)
	require.NoError(s.T(), err)

	// Get circuit state
	state, err := s.getCircuitState(backendID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), parseUUIDMust(backendID), state.BackendID)
	assert.Equal(s.T(), 1, state.SuccessCount)
	assert.Equal(s.T(), 5, state.FailCount)
}

func (s *PostgresTestSuite) TestConcurrentAccess() {
	// Create tenant
	tenantID, err := s.createTenant("Concurrent Test", "concurrent.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	// Test concurrent app creation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			appName := fmt.Sprintf("Concurrent App %d", index)
			_, err := s.createApp(tenantID, appName, "Concurrent test app")
			assert.NoError(s.T(), err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all apps were created
	apps, err := s.listAppsByTenant(tenantID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(apps), 10)
}

func (s *PostgresTestSuite) TestTransactionRollback() {
	// Start a transaction
	tx, err := s.db.DB.BeginTx(context.Background(), nil)
	require.NoError(s.T(), err)

	// Create tenant in transaction
	tenantID, err := s.createTenantTx(tx, "Tx Test", "tx.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	// Create app in transaction
	appID, err := s.createAppTx(tx, tenantID, "Tx App", "Transactional app")
	require.NoError(s.T(), err)

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(s.T(), err)

	// Verify tenant doesn't exist
	_, err = s.getTenant(tenantID)
	assert.Error(s.T(), err)

	// Verify app doesn't exist
	_, err = s.getApp(appID)
	assert.Error(s.T(), err)
}
