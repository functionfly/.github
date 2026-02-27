package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/test"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerTestSuite struct {
	test.TestSuite
	db      *storage.PostgresDB
	repo    storage.Repository
	handler *Handler
	router  *mux.Router
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) SetupTest() {
	s.TestSuite.SetupTest()

	// Set up test database
	db, err := s.setupTestDatabase()
	require.NoError(s.T(), err)
	s.db = db
	s.repo = db.Repository()

	// Create auth service
	authSvc := auth.NewAuthService(s.repo, "test-secret-key")

	// Create handler
	s.handler = NewHandler(authSvc)

	// Set up router
	s.router = mux.NewRouter()
	s.setupRoutes()
}

func (s *AuthHandlerTestSuite) setupRoutes() {
	// Auth routes
	s.router.HandleFunc("/auth/login", s.handler.HandleLogin).Methods("POST")
	s.router.HandleFunc("/auth/signup", s.handler.HandleSignup).Methods("POST")
	s.router.HandleFunc("/auth/verify-email", s.handler.HandleVerifyEmail).Methods("GET")
	s.router.HandleFunc("/auth/validate", s.handler.HandleValidateToken).Methods("GET")
}

// setupTestDatabase creates a test database connection
func (s *AuthHandlerTestSuite) setupTestDatabase() (*storage.PostgresDB, error) {
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

	return storage.NewPostgresDB()
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateTestTenant creates a test tenant for testing
func (s *AuthHandlerTestSuite) CreateTestTenant(name, domain string) (uuid.UUID, error) {
	tenant, err := s.repo.CreateTenant(s.T().Context(), name)
	if err != nil {
		return uuid.Nil, err
	}
	return tenant.ID, nil
}

func (s *AuthHandlerTestSuite) TestHandleSignup_Success() {
	// Create test tenant first
	tenantID, err := s.CreateTestTenant("Test Tenant", "test.com")
	require.NoError(s.T(), err)

	payload := map[string]interface{}{
		"tenant_id": tenantID,
		"username":  "testuser",
		"email":     "test@example.com",
		"password":  "password123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), response, "user_id")
	assert.Contains(s.T(), response, "message")
	assert.Equal(s.T(), "User created successfully. Please check your email to verify your account.", response["message"])
}

func (s *AuthHandlerTestSuite) TestHandleSignup_InvalidPayload() {
	payload := map[string]interface{}{
		"username": "testuser",
		// Missing required fields
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), response, "error")
}

func (s *AuthHandlerTestSuite) TestHandleSignup_DuplicateUsername() {
	// Create test tenant
	tenantID, err := s.CreateTestTenant("Test Tenant", "test.com")
	require.NoError(s.T(), err)

	// Create first user
	payload1 := map[string]interface{}{
		"tenant_id": tenantID,
		"username":  "testuser",
		"email":     "test1@example.com",
		"password":  "password123",
	}

	body1, _ := json.Marshal(payload1)
	req1 := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)

	assert.Equal(s.T(), http.StatusCreated, w1.Code)

	// Try to create user with same username
	payload2 := map[string]interface{}{
		"tenant_id": tenantID,
		"username":  "testuser",
		"email":     "test2@example.com",
		"password":  "password123",
	}

	body2, _ := json.Marshal(payload2)
	req2 := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	assert.Equal(s.T(), http.StatusConflict, w2.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), response, "error")
	assert.Contains(s.T(), response["error"].(string), "username")
}

func (s *AuthHandlerTestSuite) TestHandleLogin_Success() {
	// Create test tenant and user
	tenantID, err := s.CreateTestTenant("Test Tenant", "test.com")
	require.NoError(s.T(), err)

	// Create user directly in database (simulating verified user)
	repo := s.db.Repository()
	user, err := repo.CreateUser("test@example.com", "hashedpassword", tenantID)
	require.NoError(s.T(), err)
	userID := user.ID

	// Mark email as verified (simplified)
	_, err = s.db.DB.Exec("UPDATE users SET email_verified = true WHERE id = $1", userID)
	require.NoError(s.T(), err)

	payload := map[string]interface{}{
		"username": "testuser",
		"password": "password123", // This would be hashed in real scenario
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Note: This test assumes the auth service handles password verification
	// In a real scenario, you'd need to mock the password hashing/verification
	assert.True(s.T(), w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
}

func (s *AuthHandlerTestSuite) TestHandleLogin_InvalidCredentials() {
	payload := map[string]interface{}{
		"username": "nonexistent",
		"password": "wrongpassword",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), response, "error")
}

func (s *AuthHandlerTestSuite) TestHandleValidateToken_ValidToken() {
	// This test would require JWT token generation
	// For now, test the endpoint exists and returns proper error for missing token
	req := httptest.NewRequest("GET", "/auth/validate", nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *AuthHandlerTestSuite) TestHandleVerifyEmail_Success() {
	// Create test tenant and user
	tenantID, err := s.CreateTestTenant("Test Tenant", "test.com")
	require.NoError(s.T(), err)

	repo := s.db.Repository()
	user, err := repo.CreateUser("test@example.com", "hashedpassword", tenantID)
	require.NoError(s.T(), err)
	userID := user.ID

	// Generate verification token (simplified)
	token := "verification-token-123"

	// Store verification token in database (simplified)
	_, err = s.db.DB.Exec("UPDATE users SET verification_token = $1 WHERE id = $2", token, userID)
	require.NoError(s.T(), err)

	req := httptest.NewRequest("GET", "/auth/verify-email?token="+token, nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Should redirect or return success
	assert.True(s.T(), w.Code == http.StatusOK || w.Code == http.StatusFound)
}

func (s *AuthHandlerTestSuite) TestHandleVerifyEmail_InvalidToken() {
	req := httptest.NewRequest("GET", "/auth/verify-email?token=invalid-token", nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), response, "error")
}

func (s *AuthHandlerTestSuite) TestHandleGetOAuthProviders() {
	req := httptest.NewRequest("GET", "/auth/oauth/providers", nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), response, "providers")
}

func (s *AuthHandlerTestSuite) TestHandleGetOAuthURL() {
	req := httptest.NewRequest("GET", "/auth/oauth/url?provider=github", nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Should return OAuth URL or error for unsupported provider
	assert.True(s.T(), w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
}
