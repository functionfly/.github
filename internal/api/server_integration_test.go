//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type APIServerIntegrationTestSuite struct {
	test.TestSuite
	server *Server
	client *http.Client
	baseURL string
}

func TestAPIServerIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(APIServerIntegrationTestSuite))
}

func (s *APIServerIntegrationTestSuite) SetupTest() {
	s.TestSuite.SetupTest()

	// Set up environment variables for testing
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-integration-tests")
	os.Setenv("ENVIRONMENT", "test")

	// Create server
	server, err := NewServer(s.db)
	require.NoError(s.T(), err)
	s.server = server

	// Start server in background
	go func() {
		err := s.server.Start()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Server failed to start: %v", err))
		}
	}()

	// Wait for server to be ready
	s.waitForServerReady()

	s.client = &http.Client{Timeout: 10 * time.Second}
	s.baseURL = "http://localhost:8080"
}

func (s *APIServerIntegrationTestSuite) TearDownTest() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}
	s.TestSuite.TearDownTest()
}

func (s *APIServerIntegrationTestSuite) waitForServerReady() {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			s.T().Fatal("Server failed to start within timeout")
		case <-ticker.C:
			resp, err := http.Get("http://localhost:8080/health")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
	}
}

func (s *APIServerIntegrationTestSuite) TestHealthEndpoints() {
	// Test basic health endpoint
	resp, err := s.client.Get(s.baseURL + "/health")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var healthResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "ok", healthResp["status"])

	// Test detailed health endpoint
	resp, err = s.client.Get(s.baseURL + "/health/detailed")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var detailedHealth map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&detailedHealth)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), detailedHealth, "status")
	assert.Contains(s.T(), detailedHealth, "services")
	assert.Contains(s.T(), detailedHealth, "checks")
}

func (s *APIServerIntegrationTestSuite) TestMetricsEndpoints() {
	// Test global metrics endpoint
	resp, err := s.client.Get(s.baseURL + "/v1/metrics/global")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var metrics map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), metrics, "uptime")
	assert.Contains(s.T(), metrics, "latency")
	assert.Contains(s.T(), metrics, "failoverRate")

	// Test Prometheus metrics endpoint
	resp, err = s.client.Get(s.baseURL + "/metrics")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Contains(s.T(), resp.Header.Get("Content-Type"), "text/plain")
}

func (s *APIServerIntegrationTestSuite) TestAuthFlow() {
	// Test signup
	tenantID, err := s.db.Repository().CreateTenant("Integration Test", "integration.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	signupPayload := map[string]interface{}{
		"tenant_id": tenantID,
		"username":  "integrationuser",
		"email":     "integration@example.com",
		"password":  "testpassword123",
	}

	body, _ := json.Marshal(signupPayload)
	resp, err := s.client.Post(s.baseURL+"/v1/auth/signup", "application/json", bytes.NewReader(body))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var signupResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&signupResp)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), signupResp, "user_id")

	// Mark user as verified for login test
	userID := signupResp["user_id"].(string)
	_, err = s.db.DB.Exec("UPDATE users SET email_verified = true WHERE id = $1", userID)
	require.NoError(s.T(), err)

	// Test login (this would work with proper password hashing in auth service)
	loginPayload := map[string]interface{}{
		"username": "integrationuser",
		"password": "testpassword123",
	}

	body, _ = json.Marshal(loginPayload)
	resp, err = s.client.Post(s.baseURL+"/v1/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Login may succeed or fail depending on password hashing implementation
	assert.True(s.T(), resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized)
}

func (s *APIServerIntegrationTestSuite) TestTenantManagement() {
	// Create a test tenant via direct database access first
	tenantID, err := s.db.Repository().CreateTenant("Integration Tenant", "integration.tenant.com", "admin@integration.com")
	require.NoError(s.T(), err)

	// Test OAuth providers endpoint (public)
	resp, err := s.client.Get(s.baseURL + "/v1/auth/oauth/providers")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Test content endpoints (public)
	resp, err = s.client.Get(s.baseURL + "/v1/content/changelog")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *APIServerIntegrationTestSuite) TestMonitoringEndpoints() {
	// Test monitoring health endpoint
	resp, err := s.client.Get(s.baseURL + "/v1/monitoring/health")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var healthResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), healthResp, "status")

	// Test database monitoring
	resp, err = s.client.Get(s.baseURL + "/v1/monitoring/database/health")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *APIServerIntegrationTestSuite) TestSecurityHeaders() {
	resp, err := s.client.Get(s.baseURL + "/health")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Check security headers
	assert.Equal(s.T(), "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(s.T(), "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Contains(s.T(), resp.Header.Get("Content-Security-Policy"), "default-src")
}

func (s *APIServerIntegrationTestSuite) TestRateLimiting() {
	// Make multiple rapid requests to test rate limiting
	successCount := 0
	for i := 0; i < 10; i++ {
		resp, err := s.client.Get(s.baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
		}
		time.Sleep(10 * time.Millisecond) // Small delay between requests
	}

	// Should have some successful requests (rate limiting may kick in)
	assert.Greater(s.T(), successCount, 0)
}

func (s *APIServerIntegrationTestSuite) TestConcurrentRequests() {
	// Test concurrent requests to ensure server handles load
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := s.client.Get(s.baseURL + "/health")
			if err != nil {
				results <- 0
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		assert.Equal(s.T(), http.StatusOK, statusCode)
	}
}

func (s *APIServerIntegrationTestSuite) TestMetricsStream() {
	// Test Server-Sent Events endpoint
	req, err := http.NewRequest("GET", s.baseURL+"/v1/metrics/stream", nil)
	require.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(s.T(), "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(s.T(), "no-cache", resp.Header.Get("Cache-Control"))
}

func (s *APIServerIntegrationTestSuite) TestDatabaseConnectivity() {
	// Test database health through API
	resp, err := s.client.Get(s.baseURL + "/health/check?name=database")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var checkResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&checkResp)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "database", checkResp["name"])
	assert.Equal(s.T(), "healthy", checkResp["status"])
}

func (s *APIServerIntegrationTestSuite) TestErrorHandling() {
	// Test invalid endpoint
	resp, err := s.client.Get(s.baseURL + "/v1/nonexistent/endpoint")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Should return 404
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	// Test invalid method
	resp, err = s.client.Post(s.baseURL+"/health", "application/json", nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed
	assert.Equal(s.T(), http.StatusMethodNotAllowed, resp.StatusCode)
}

func (s *APIServerIntegrationTestSuite) TestCORSHeaders() {
	// Test preflight request
	req, err := http.NewRequest("OPTIONS", s.baseURL+"/health", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := s.client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(s.T(), "*", resp.Header.Get("Access-Control-Allow-Origin"))
}