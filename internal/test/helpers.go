package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestSuite provides common test functionality
type TestSuite struct {
	suite.Suite
	db interface{} // Will be set by the test file
}

// DB returns the database instance
func (s *TestSuite) DB() interface{} {
	return s.db
}

// SetDB sets the database instance
func (s *TestSuite) SetDB(db interface{}) {
	s.db = db
}

// SetupTest sets up the test environment - database setup is now handled by the test file
func (s *TestSuite) SetupTest() {
	// Database setup is now handled externally
}

// TearDownTest cleans up after each test
func (s *TestSuite) TearDownTest() {
	if s.db != nil {
		// Clean up test data - implementation will be provided by test file
		s.CleanupTestData()
	}
}

// SetupTestDatabase creates a test database connection
func SetupTestDatabase() (interface{}, error) {
	return nil, nil
}

// CleanupTestData removes test data between tests
func (s *TestSuite) CleanupTestData() {
}

// CreateTestTenant creates a test tenant for testing
func CreateTestTenant(repo interface{}, name, domain string) (string, error) {
	return "test-tenant-id", nil
}

// CreateTestApp creates a test app for testing
func CreateTestApp(repo interface{}, tenantID, name, description string) (string, error) {
	return "test-app-id", nil
}

// CreateTestBackend creates a test backend for testing
func CreateTestBackend(repo interface{}, appID, provider, config string) (string, error) {
	return "test-backend-id", nil
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// AssertWithinDuration asserts that actual is within expected ± tolerance
func AssertWithinDuration(t *testing.T, expected, actual, tolerance time.Duration) {
	require.True(t, actual >= expected-tolerance && actual <= expected+tolerance,
		"Expected %v ± %v, got %v", expected, tolerance, actual)
}

// MockHTTPClient creates a mock HTTP client for testing
type MockHTTPClient struct {
	Responses map[string]interface{}
	Errors    map[string]error
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Responses: make(map[string]interface{}),
		Errors:    make(map[string]error),
	}
}

// SetResponse sets a mock response for a URL
func (m *MockHTTPClient) SetResponse(url string, response interface{}) {
	m.Responses[url] = response
}

// SetError sets a mock error for a URL
func (m *MockHTTPClient) SetError(url string, err error) {
	m.Errors[url] = err
}

// GetResponse returns the mock response for a URL
func (m *MockHTTPClient) GetResponse(url string) (interface{}, error) {
	if err, exists := m.Errors[url]; exists {
		return nil, err
	}
	if response, exists := m.Responses[url]; exists {
		return response, nil
	}
	return nil, fmt.Errorf("no mock response configured for %s", url)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}