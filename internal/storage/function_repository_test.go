package storage

import (
	"context"
	"os"
	"testing"

	"github.com/functionfly/functionfly/internal/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FunctionRepositoryTestSuite struct {
	test.TestSuite
	db *PostgresDB
}

func TestFunctionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(FunctionRepositoryTestSuite))
}

func (s *FunctionRepositoryTestSuite) SetupTest() {
	s.TestSuite.SetupTest()

	db, err := s.setupTestDatabase()
	require.NoError(s.T(), err)
	s.db = db
	s.SetDB(db)

	err = RunMigrations(db)
	require.NoError(s.T(), err)
}

func (s *FunctionRepositoryTestSuite) setupTestDatabase() (*PostgresDB, error) {
	os.Setenv("DB_HOST", getEnvOrDefault("TEST_DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnvOrDefault("TEST_DB_PORT", "5434"))
	os.Setenv("DB_USER", getEnvOrDefault("TEST_DB_USER", "postgres"))
	os.Setenv("DB_PASSWORD", getEnvOrDefault("TEST_DB_PASSWORD", "postgres"))
	os.Setenv("DB_NAME", getEnvOrDefault("TEST_DB_NAME", "functionfly_test"))
	os.Setenv("DB_SSLMODE", "disable")

	os.Setenv("DB_MAX_OPEN_CONNS", "5")
	os.Setenv("DB_MAX_IDLE_CONNS", "2")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1m")

	return NewPostgresDB()
}

func (s *FunctionRepositoryTestSuite) TearDownTest() {
	if s.db != nil {
		s.CleanupTestData()
	}
}

func (s *FunctionRepositoryTestSuite) CleanupTestData() {
	tables := []string{
		"function_deployments",
		"functions",
		"apps",
		"tenants",
	}
	for _, table := range tables {
		_, err := s.db.DB.Exec("TRUNCATE TABLE " + table + " CASCADE")
		if err != nil {
			s.T().Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}

func (s *FunctionRepositoryTestSuite) createTestTenant() (*Tenant, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	return repo.CreateTenant(ctx, "test-tenant-"+uuid.New().String())
}

func (s *FunctionRepositoryTestSuite) createTestApp(tenantID uuid.UUID) (*App, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	return repo.CreateApp(ctx, "test-app-"+uuid.New().String(), "test-app", tenantID)
}

func (s *FunctionRepositoryTestSuite) TestCreateFunction() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "test-function",
		Providers: []string{"aws", "cloudflare"},
		Region:    "us-east-1",
		Code:      "package main func main() {}",
		Version:   "1.0.0",
		Status:    "draft",
	}

	created, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), created.ID)
	assert.Equal(s.T(), tenant.ID, created.TenantID)
	assert.Equal(s.T(), "test-function", created.Name)
	assert.Equal(s.T(), "1.0.0", created.Version)
	assert.Equal(s.T(), "draft", created.Status)
}

func (s *FunctionRepositoryTestSuite) TestCreateFunctionWithEnvVars() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "test-function-env",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "package main func main() {}",
		EnvVars: []EnvironmentVariable{
			{Key: "API_KEY", Value: "secret123", IsSecret: true},
			{Key: "DEBUG", Value: "false", IsSecret: false},
		},
	}

	created, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)
	assert.Len(s.T(), created.EnvVars, 2)
	assert.Equal(s.T(), "API_KEY", created.EnvVars[0].Key)
	assert.Equal(s.T(), "secret123", created.EnvVars[0].Value)
	assert.True(s.T(), created.EnvVars[0].IsSecret)
}

func (s *FunctionRepositoryTestSuite) TestGetFunctionByID() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "get-test-function",
		Providers: []string{"aws"},
		Region:    "us-west-2",
		Code:      "package main",
	}

	created, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetFunctionByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), "get-test-function", retrieved.Name)
}

func (s *FunctionRepositoryTestSuite) TestGetFunctionByIDNotFound() {
	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := repo.GetFunctionByID(ctx, nonExistentID)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "function not found")
}

func (s *FunctionRepositoryTestSuite) TestListFunctionsByTenant() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function1 := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "list-function-1",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "code1",
	}
	function2 := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "list-function-2",
		Providers: []string{"cloudflare"},
		Region:    "eu-west-1",
		Code:      "code2",
	}

	_, err = repo.CreateFunction(ctx, function1)
	require.NoError(s.T(), err)
	_, err = repo.CreateFunction(ctx, function2)
	require.NoError(s.T(), err)

	functions, err := repo.ListFunctionsByTenant(ctx, tenant.ID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(functions), 2)
}

func (s *FunctionRepositoryTestSuite) TestUpdateFunction() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "update-test-function",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "original code",
		Status:    "draft",
	}

	created, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)

	updates := map[string]interface{}{
		"name":   "updated-function-name",
		"status": "published",
		"code":   "updated code",
	}

	updated, err := repo.UpdateFunction(ctx, created.ID, updates)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "updated-function-name", updated.Name)
	assert.Equal(s.T(), "published", updated.Status)
	assert.Equal(s.T(), "updated code", updated.Code)
}

func (s *FunctionRepositoryTestSuite) TestUpdateFunctionNotFound() {
	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	nonExistentID := uuid.New()
	updates := map[string]interface{}{
		"name": "should-fail",
	}

	_, err := repo.UpdateFunction(ctx, nonExistentID, updates)
	assert.Error(s.T(), err)
}

func (s *FunctionRepositoryTestSuite) TestDeleteFunction() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "delete-test-function",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "to be deleted",
	}

	created, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)

	err = repo.DeleteFunction(ctx, created.ID)
	require.NoError(s.T(), err)

	_, err = repo.GetFunctionByID(ctx, created.ID)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "function not found")
}

func (s *FunctionRepositoryTestSuite) TestDeleteFunctionNotFound() {
	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	nonExistentID := uuid.New()
	err := repo.DeleteFunction(ctx, nonExistentID)
	assert.Error(s.T(), err)
}

func (s *FunctionRepositoryTestSuite) TestCreateFunctionDeployment() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "deployment-test-function",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "package main",
	}

	createdFunc, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)

	deployment := &FunctionDeployment{
		FunctionID: createdFunc.ID,
		Version:    "1.0.0",
		Status:     "success",
		Provider:   "aws",
		Region:     "us-east-1",
		DeployedURL: stringPtr("https://example.com/func"),
	}

	created, err := repo.CreateFunctionDeployment(ctx, deployment)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), created.ID)
	assert.Equal(s.T(), createdFunc.ID, created.FunctionID)
	assert.Equal(s.T(), "success", created.Status)
}

func (s *FunctionRepositoryTestSuite) TestGetActiveDeploymentForFunction() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "active-deployment-test",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "package main",
	}

	createdFunc, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)

	deployment := &FunctionDeployment{
		FunctionID: createdFunc.ID,
		Version:    "1.0.0",
		Status:     "success",
		Provider:   "aws",
		Region:     "us-east-1",
		DeployedURL: stringPtr("https://active.example.com"),
	}

	_, err = repo.CreateFunctionDeployment(ctx, deployment)
	require.NoError(s.T(), err)

	active, err := repo.GetActiveDeploymentForFunction(ctx, createdFunc.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "success", active.Status)
	assert.Equal(s.T(), "https://active.example.com", *active.DeployedURL)
}

func (s *FunctionRepositoryTestSuite) TestGetActiveDeploymentForFunctionNotFound() {
	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := repo.GetActiveDeploymentForFunction(ctx, nonExistentID)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "no active deployment found")
}

func (s *FunctionRepositoryTestSuite) TestListFunctionDeployments() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	app, err := s.createTestApp(tenant.ID)
	require.NoError(s.T(), err)

	repo := NewFunctionRepository(s.db.DB)
	ctx := context.Background()

	function := &FunctionConfig{
		TenantID:  tenant.ID,
		AppID:     &app.ID,
		Name:      "list-deployments-test",
		Providers: []string{"aws"},
		Region:    "us-east-1",
		Code:      "package main",
	}

	createdFunc, err := repo.CreateFunction(ctx, function)
	require.NoError(s.T(), err)

	for i := 0; i < 3; i++ {
		deployment := &FunctionDeployment{
			FunctionID: createdFunc.ID,
			Version:    "1.0." + string(rune('0'+i)),
			Status:     "success",
			Provider:   "aws",
			Region:     "us-east-1",
		}
		_, err = repo.CreateFunctionDeployment(ctx, deployment)
		require.NoError(s.T(), err)
	}

	deployments, err := repo.ListFunctionDeployments(ctx, createdFunc.ID, 0)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(deployments), 3)
}
