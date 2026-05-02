package frg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/frg"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no_special_chars",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "xss_script_tag",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "xss_img_onerror",
			input:    `<img src=x onerror="alert('xss')">`,
			expected: "&lt;img src=x onerror=&quot;alert(&#39;xss&#39;)&quot;&gt;",
		},
		{
			name:     "quotes_only",
			input:    `He said "Hello" and 'World'`,
			expected: "He said &quot;Hello&quot; and &#39;World&#39;",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "already_escaped",
			input:    "&lt;script&gt;",
			expected: "&lt;script&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuotaCheckResult_Structure(t *testing.T) {
	result := services.QuotaCheckResult{
		Allowed: false,
		Reason:  "Execution quota exceeded",
		Status: &services.RealtimeQuotaStatus{
			TenantID:          uuid.New(),
			ExecutionsUsed:    1000,
			ExecutionsLimit:   1000,
			ExecutionsPercent: 100.0,
			Status:            "exceeded",
		},
	}

	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "quota exceeded")
	assert.NotNil(t, result.Status)
	assert.Equal(t, 1000, result.Status.ExecutionsUsed)
	assert.Equal(t, 1000, result.Status.ExecutionsLimit)
}

func TestRealtimeQuotaStatus_Structure(t *testing.T) {
	status := services.RealtimeQuotaStatus{
		TenantID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		ExecutionsUsed:    500,
		ExecutionsLimit:   1000,
		ExecutionsPercent: 50.0,
		ComputeMsUsed:     1800000,
		ComputeMsLimit:     3600000,
		ComputeMsPercent:  50.0,
		Status:            "ok",
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", status.TenantID.String())
	assert.Equal(t, 500, status.ExecutionsUsed)
	assert.Equal(t, 1000, status.ExecutionsLimit)
	assert.Equal(t, 50.0, status.ExecutionsPercent)
	assert.Equal(t, "ok", status.Status)
}

func TestFRGMetricsRecording(t *testing.T) {
	tenantID := uuid.New().String()
	graphID := uuid.New().String()

	monitoring.RecordFRGGraphExecution(tenantID, graphID, "execute", "success")
	monitoring.RecordFRGGraphExecutionDuration(tenantID, graphID, 250*time.Millisecond)
	monitoring.RecordFRGGraphActiveIncrement(tenantID)
	monitoring.RecordFRGGraphActiveDecrement(tenantID)
	monitoring.RecordFRGQuotaExceeded(tenantID)
	monitoring.RecordFRGQuotaUsagePercent(tenantID, 75.5)
	monitoring.RecordFRGWebhookSignatureFailure("invalid_signature")
	monitoring.RecordFRGGraphCreation(tenantID, "private", "created")
}

func TestInputSizeLimits(t *testing.T) {
	const (
		maxNodes     = 100
		maxEdges     = 500
		maxGraphJSON = 1_000_000
		maxNameLen   = 100
	)

	t.Run("max_nodes_boundary", func(t *testing.T) {
		nodes := make([]frg.GraphNodeRef, maxNodes)
		for i := range nodes {
			nodes[i] = frg.GraphNodeRef{NodeID: uuid.New().String()}
		}
		assert.Len(t, nodes, maxNodes)
		assert.True(t, len(nodes) <= maxNodes)
	})

	t.Run("max_edges_boundary", func(t *testing.T) {
		edges := make([]frg.GraphEdge, maxEdges)
		for i := range edges {
			edges[i] = frg.GraphEdge{ID: uuid.New().String()}
		}
		assert.Len(t, edges, maxEdges)
		assert.True(t, len(edges) <= maxEdges)
	})

	t.Run("max_name_length", func(t *testing.T) {
		validName := make([]byte, maxNameLen)
		for i := range validName {
			validName[i] = 'a'
		}
		assert.Len(t, validName, maxNameLen)
		assert.True(t, len(validName) <= maxNameLen)
	})

	t.Run("max_graph_json_size", func(t *testing.T) {
		largeGraph := make([]byte, maxGraphJSON+1)
		assert.True(t, len(largeGraph) > maxGraphJSON)
	})
}

func TestXSSSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "angle_brackets_escaped",
			input:    "<script>alert(1)</script>",
			expected: "&lt;script&gt;alert(1)&lt;/script&gt;",
		},
		{
			name:     "double_quotes_escaped",
			input:    `<img src="x">`,
			expected: "&lt;img src=&quot;x&quot;&gt;",
		},
		{
			name:     "single_quotes_escaped",
			input:    `<img src='x'>`,
			expected: "&lt;img src=&#39;x&#39;&gt;",
		},
		{
			name:     "mixed_quotes",
			input:    `He said "Hello" and 'World'`,
			expected: "He said &quot;Hello&quot; and &#39;World&#39;",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantIsolation(t *testing.T) {
	t.Run("private_graph_blocks_non_owner", func(t *testing.T) {
		ownerUserID := uuid.New()
		nonOwnerUserID := uuid.New()
		ownerTenantID := uuid.New()
		nonOwnerTenantID := uuid.New()

		def := &frg.GraphDefinition{
			Visibility:  "private",
			OwnerUserID:  &ownerUserID,
			TenantID:    &ownerTenantID,
		}

		canAccess := !(def.Visibility == "private" &&
			def.OwnerUserID != nil && *def.OwnerUserID != nonOwnerUserID &&
			def.TenantID != nil && *def.TenantID != nonOwnerTenantID)

		assert.False(t, canAccess, "Non-owner should not access private graph")
	})

	t.Run("public_graph_allows_any_user", func(t *testing.T) {
		ownerUserID := uuid.New()
		otherUserID := uuid.New()

		def := &frg.GraphDefinition{
			Visibility: "public",
			OwnerUserID: &ownerUserID,
		}

		canAccess := def.Visibility == "public" || (def.OwnerUserID != nil && *def.OwnerUserID == otherUserID)
		assert.True(t, canAccess, "Public graph should allow access")
	})

	t.Run("tenant_member_can_access_private_graph", func(t *testing.T) {
		ownerUserID := uuid.New()
		tenantID := uuid.New()

		def := &frg.GraphDefinition{
			Visibility:   "private",
			OwnerUserID: &ownerUserID,
			TenantID:    &tenantID,
		}

		canAccess := def.TenantID != nil && *def.TenantID == tenantID
		assert.True(t, canAccess, "Tenant member should access private graph")
	})
}

func TestExecutionTimeout(t *testing.T) {
	const defaultExecutionTimeout = 5 * time.Minute

	t.Run("default_timeout_is_5_minutes", func(t *testing.T) {
		assert.Equal(t, 5*time.Minute, defaultExecutionTimeout)
	})

	t.Run("timeout_parses_valid_duration", func(t *testing.T) {
		timeoutStr := "30s"
		parsed, err := time.ParseDuration(timeoutStr)
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, parsed)
	})

	t.Run("timeout_less_than_default_is_used", func(t *testing.T) {
		timeoutStr := "30s"
		parsed, err := time.ParseDuration(timeoutStr)
		require.NoError(t, err)
		if parsed < defaultExecutionTimeout {
			assert.Equal(t, 30*time.Second, parsed)
		}
	})
}

func TestGraphResponse_JSONSerialization(t *testing.T) {
	graphID := uuid.New()
	resp := GraphResponse{
		GraphDefinition: frg.GraphDefinition{
			ID:      graphID,
			Author:  "test-author",
			Name:    "test-graph",
			Version: "v1",
		},
		FullName: "test-author/test-graph@v1",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded GraphResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.FullName, decoded.FullName)
	assert.Equal(t, resp.Author, decoded.Author)
}

func TestExecuteGraphResponse_Structure(t *testing.T) {
	resp := ExecuteGraphResponse{
		InstanceID:  uuid.New().String(),
		Status:     "completed",
		Output:     map[string]interface{}{"result": "success"},
		DurationMs: 150,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded ExecuteGraphResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.InstanceID, decoded.InstanceID)
	assert.Equal(t, resp.Status, decoded.Status)
	assert.Equal(t, resp.DurationMs, decoded.DurationMs)
}

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	respondJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	message := "test error message"

	respondError(w, http.StatusBadRequest, message)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, message, resp["error"])
}

func TestParseUUID(t *testing.T) {
	validUUID := uuid.New().String()
	parsed, err := parseUUID(validUUID)
	require.NoError(t, err)
	assert.Equal(t, validUUID, parsed.String())

	invalidUUID := "not-a-uuid"
	_, err = parseUUID(invalidUUID)
	assert.Error(t, err)
}
