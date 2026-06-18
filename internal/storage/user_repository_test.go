package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	test.TestSuite
	db *PostgresDB
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupTest() {
	s.TestSuite.SetupTest()

	db, err := s.setupTestDatabase()
	require.NoError(s.T(), err)
	s.db = db
	s.SetDB(db)

	err = RunMigrations(db)
	require.NoError(s.T(), err)
}

func (s *UserRepositoryTestSuite) setupTestDatabase() (*PostgresDB, error) {
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

func (s *UserRepositoryTestSuite) TearDownTest() {
	if s.db != nil {
		s.CleanupTestData()
	}
}

func (s *UserRepositoryTestSuite) CleanupTestData() {
	tables := []string{
		"users",
		"tenants",
	}
	for _, table := range tables {
		_, err := s.db.DB.Exec("TRUNCATE TABLE " + table + " CASCADE")
		if err != nil {
			s.T().Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}

func (s *UserRepositoryTestSuite) createTestTenant() (*Tenant, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	return repo.CreateTenant(ctx, "test-tenant-"+uuid.New().String())
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	user, err := repo.CreateUser(ctx, "test@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), user.ID)
	assert.Equal(s.T(), "test@example.com", user.Email)
	assert.Equal(s.T(), tenant.ID, user.TenantID)
	assert.Equal(s.T(), "user", user.Role)
	assert.False(s.T(), user.EmailVerified)
	assert.False(s.T(), user.MFAEnabled)
}

func (s *UserRepositoryTestSuite) TestCreateUserWithSocialAuth() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	provider := "github"
	providerID := "123456"
	providerData := map[string]interface{}{
		"access_token": "abc123",
		"scope":        "user:email",
	}

	user, err := repo.CreateUserWithSocialAuth(ctx, "social@example.com", tenant.ID, provider, providerID, providerData)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), user.ID)
	assert.Equal(s.T(), "social@example.com", user.Email)
	assert.True(s.T(), user.EmailVerified)
	assert.Equal(s.T(), &provider, user.Provider)
	assert.Equal(s.T(), &providerID, user.ProviderID)
}

func (s *UserRepositoryTestSuite) TestCreateUserWithRole() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	user := &User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		Email:        "admin@example.com",
		PasswordHash: "hashedpassword123",
		Role:         "admin",
		EmailVerified: true,
	}
	created, err := repo.CreateUserWithRole(ctx, user)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), created.ID)
	assert.Equal(s.T(), "admin", created.Role)
	assert.True(s.T(), created.EmailVerified)
}

func (s *UserRepositoryTestSuite) TestGetUserByEmail() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "getbyemail@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByEmail(ctx, "getbyemail@example.com")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), "getbyemail@example.com", retrieved.Email)
}

func (s *UserRepositoryTestSuite) TestGetUserByEmailNotFound() {
	repo := NewUserRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *UserRepositoryTestSuite) TestGetUserByUsername() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	username := "testuser"
	created, err := repo.CreateUser(ctx, "username@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	_, err = repo.UpdateUser(ctx, created.ID, map[string]interface{}{"username": username})
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByUsername(ctx, username)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), &username, retrieved.Username)
}

func (s *UserRepositoryTestSuite) TestGetUserByUsernameNotFound() {
	repo := NewUserRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetUserByUsername(ctx, "nonexistentuser")
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *UserRepositoryTestSuite) TestGetUserByID() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "getbyid@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), "getbyid@example.com", retrieved.Email)
}

func (s *UserRepositoryTestSuite) TestGetUserByIDNotFound() {
	repo := NewUserRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetUserByID(ctx, uuid.New())
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *UserRepositoryTestSuite) TestUpdateUser() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "update@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	updates := map[string]interface{}{
		"name":  "Test User",
		"bio":   "This is a test bio",
		"role":  "admin",
	}

	updated, err := repo.UpdateUser(ctx, created.ID, updates)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test User", updated.Name)
	assert.Equal(s.T(), "This is a test bio", *updated.Bio)
	assert.Equal(s.T(), "admin", updated.Role)
}

func (s *UserRepositoryTestSuite) TestUpdateUserNotFound() {
	repo := NewUserRepository(s.db)
	ctx := context.Background()

	updates := map[string]interface{}{
		"name": "Should Fail",
	}

	_, err := repo.UpdateUser(ctx, uuid.New(), updates)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "user not found")
}

func (s *UserRepositoryTestSuite) TestUpdateUserEmailVerification() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "verify@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)
	assert.False(s.T(), created.EmailVerified)

	token := "test-verification-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	err = repo.UpdateUserEmailVerification(ctx, created.ID, true, &token, &expiresAt)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), retrieved.EmailVerified)
}

func (s *UserRepositoryTestSuite) TestGetUserByVerificationToken() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	token := "test-token-123"
	expiresAt := time.Now().Add(24 * time.Hour)

	created, err := repo.CreateUser(ctx, "token@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	err = repo.UpdateUserEmailVerification(ctx, created.ID, false, &token, &expiresAt)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByVerificationToken(ctx, token)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
}

func (s *UserRepositoryTestSuite) TestGetUserByVerificationTokenNotFound() {
	repo := NewUserRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetUserByVerificationToken(ctx, "nonexistent-token")
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *UserRepositoryTestSuite) TestGetUserBySocialProvider() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	provider := "github"
	providerID := "789012"
	providerData := map[string]interface{}{
		"access_token": "token123",
	}

	created, err := repo.CreateUserWithSocialAuth(ctx, "social2@example.com", tenant.ID, provider, providerID, providerData)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserBySocialProvider(ctx, provider, providerID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
}

func (s *UserRepositoryTestSuite) TestGetUserBySocialProviderNotFound() {
	repo := NewUserRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetUserBySocialProvider(ctx, "github", "nonexistent")
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *UserRepositoryTestSuite) TestListActiveUsersByTenant() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	_, err = repo.CreateUser(ctx, "list1@example.com", "pass1", tenant.ID)
	require.NoError(s.T(), err)
	_, err = repo.CreateUser(ctx, "list2@example.com", "pass2", tenant.ID)
	require.NoError(s.T(), err)

	users, err := repo.ListActiveUsersByTenant(ctx, tenant.ID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(users), 2)
}

func (s *UserRepositoryTestSuite) TestCountActiveUsersByTenant() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	_, err = repo.CreateUser(ctx, "count1@example.com", "pass1", tenant.ID)
	require.NoError(s.T(), err)
	_, err = repo.CreateUser(ctx, "count2@example.com", "pass2", tenant.ID)
	require.NoError(s.T(), err)

	count, err := repo.CountActiveUsersByTenant(ctx, tenant.ID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), count, 2)
}

func (s *UserRepositoryTestSuite) TestDeactivateUser() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "deactivate@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	deactivatedBy := uuid.New()
	err = repo.DeactivateUser(ctx, created.ID, deactivatedBy)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), retrieved.DeactivatedAt)
	assert.Equal(s.T(), &deactivatedBy, retrieved.DeactivatedBy)
}

func (s *UserRepositoryTestSuite) TestReactivateUser() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "reactivate@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	deactivatedBy := uuid.New()
	err = repo.DeactivateUser(ctx, created.ID, deactivatedBy)
	require.NoError(s.T(), err)

	err = repo.ReactivateUser(ctx, created.ID)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetUserByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved.DeactivatedAt)
	assert.Nil(s.T(), retrieved.DeactivatedBy)
}

func (s *UserRepositoryTestSuite) TestUpdateUserLastActive() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "lastactive@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)

	err = repo.UpdateUserLastActive(ctx, created.ID)
	require.NoError(s.T(), err)

	lastActive, err := repo.GetUserLastActive(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), lastActive)
}

func (s *UserRepositoryTestSuite) TestIncrementUserTokenVersion() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "tokenversion@example.com", "hashedpassword123", tenant.ID)
	require.NoError(s.T(), err)
	initialVersion := created.TokenVersion

	newVersion, err := repo.IncrementUserTokenVersion(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), initialVersion+1, newVersion)
}

func (s *UserRepositoryTestSuite) TestListUserIDsByTenant() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	repo := NewUserRepository(s.db)
	ctx := context.Background()

	user1, err := repo.CreateUser(ctx, "ids1@example.com", "pass1", tenant.ID)
	require.NoError(s.T(), err)
	user2, err := repo.CreateUser(ctx, "ids2@example.com", "pass2", tenant.ID)
	require.NoError(s.T(), err)

	ids, err := repo.ListUserIDsByTenant(ctx, tenant.ID)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), ids, user1.ID)
	assert.Contains(s.T(), ids, user2.ID)
}
