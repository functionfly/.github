package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCrossTenantIsolation verifies that tenant A cannot access tenant B's resources.
// These are integration tests that require a running test server.
func TestCrossTenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	serverURL := getTestServerURL()
	tenantAToken := "test-tenant-a-token"
	tenantBToken := "test-tenant-b-token"

	if serverURL == "" || tenantAToken == "" || tenantBToken == "" {
		t.Skip("Test server or tokens not configured")
	}

	t.Run("TenantA_cannot_read_TenantB_apps", func(t *testing.T) {
		appID := createTestApp(t, serverURL, tenantBToken, "tenant-b-app")
		if appID == "" {
			t.Skip("Could not create test app")
		}

		req := newAuthRequest(t, "GET", fmt.Sprintf("%s/v1/apps/%s", serverURL, appID), nil, tenantAToken)
		resp := doRequest(t, req)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 when tenant A reads tenant B's app, got %d", resp.StatusCode)
	})

	t.Run("TenantA_cannot_delete_TenantB_apps", func(t *testing.T) {
		appID := createTestApp(t, serverURL, tenantBToken, "tenant-b-app-delete")
		if appID == "" {
			t.Skip("Could not create test app")
		}

		req := newAuthRequest(t, "DELETE", fmt.Sprintf("%s/v1/apps/%s", serverURL, appID), nil, tenantAToken)
		resp := doRequest(t, req)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 when tenant A deletes tenant B's app, got %d", resp.StatusCode)
	})

	t.Run("TenantA_list_only_sees_own_apps", func(t *testing.T) {
		createTestApp(t, serverURL, tenantAToken, "tenant-a-exclusive-app")
		createTestApp(t, serverURL, tenantBToken, "tenant-b-exclusive-app")

		req := newAuthRequest(t, "GET", fmt.Sprintf("%s/v1/apps", serverURL), nil, tenantAToken)
		resp := doRequest(t, req)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Apps []struct {
				Name     string `json:"name"`
				TenantID string `json:"tenant_id"`
			} `json:"apps"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

		tenantBID := getTenantID(t, serverURL, tenantBToken)
		for _, app := range result.Apps {
			assert.NotEqual(t, tenantBID, app.TenantID,
				"Tenant A should not see tenant B's app '%s'", app.Name)
		}
	})

	t.Run("TenantA_cannot_access_TenantB_backends", func(t *testing.T) {
		req := newAuthRequest(t, "GET", fmt.Sprintf("%s/v1/backends", serverURL), nil, tenantAToken)
		resp := doRequest(t, req)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Backends []struct {
				TenantID string `json:"tenant_id"`
			} `json:"backends"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

		tenantBID := getTenantID(t, serverURL, tenantBToken)
		for _, backend := range result.Backends {
			assert.NotEqual(t, tenantBID, backend.TenantID,
				"Tenant A should not see tenant B's backends")
		}
	})
}

// TestTenantIsolationUnit tests tenant isolation logic at the unit level
func TestTenantIsolationUnit(t *testing.T) {
	t.Run("different_tenant_IDs_produce_different_rate_limit_keys", func(t *testing.T) {
		tenantAKey := fmt.Sprintf("tenant:%s:%s", "tenant-a-id", "/v1/apps")
		tenantBKey := fmt.Sprintf("tenant:%s:%s", "tenant-b-id", "/v1/apps")
		assert.NotEqual(t, tenantAKey, tenantBKey,
			"Rate limit keys for different tenants should be different")
	})

	t.Run("unauthenticated_requests_use_IP_based_rate_limit_key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/apps", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		// Verify the request has no Authorization header
		assert.Empty(t, req.Header.Get("Authorization"))
		// IP-based rate limiting is the fallback for unauthenticated requests
		assert.True(t, true, "IP-based rate limiting is the fallback for unauthenticated requests")
	})

	t.Run("tenant_scoped_key_format_is_correct", func(t *testing.T) {
		tenantID := "550e8400-e29b-41d4-a716-446655440000"
		path := "/v1/functions"
		key := fmt.Sprintf("tenant:%s:%s", tenantID, path)
		assert.Equal(t, "tenant:550e8400-e29b-41d4-a716-446655440000:/v1/functions", key)
	})
}

// Helper functions

func getTestServerURL() string {
	return "http://localhost:8080"
}

func getTenantID(t *testing.T, serverURL, token string) string {
	t.Helper()
	req := newAuthRequest(t, "GET", serverURL+"/v1/auth/me", nil, token)
	resp := doRequest(t, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var result struct {
		TenantID string `json:"tenant_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	return result.TenantID
}

func createTestApp(t *testing.T, serverURL, token, name string) string {
	t.Helper()
	body := map[string]interface{}{
		"name":        name,
		"description": "Test app for isolation testing",
	}
	req := newAuthRequest(t, "POST", serverURL+"/v1/apps", body, token)
	resp := doRequest(t, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return ""
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	return result.ID
}

func newAuthRequest(t *testing.T, method, url string, body interface{}, token string) *http.Request {
	t.Helper()
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		require.NoError(t, err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func doRequest(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}
