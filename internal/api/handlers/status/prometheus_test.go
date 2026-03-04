package status

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPrometheusServer creates a test HTTP server that mimics Prometheus API responses
type MockPrometheusServer struct {
	Server      *httptest.Server
	Responses   map[string]*PrometheusResponse
	StatusCodes map[string]int
	CallCount   map[string]int
}

// NewMockPrometheusServer creates a new mock Prometheus server
func NewMockPrometheusServer() *MockPrometheusServer {
	mock := &MockPrometheusServer{
		Responses:   make(map[string]*PrometheusResponse),
		StatusCodes: make(map[string]int),
		CallCount:   make(map[string]int),
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		mock.CallCount[query]++

		// Check if there's a forced status code
		if statusCode, ok := mock.StatusCodes[query]; ok {
			w.WriteHeader(statusCode)
			if statusCode != http.StatusOK {
				json.NewEncoder(w).Encode(map[string]string{"error": "test error"})
				return
			}
		}

		// Return mock response if configured
		if response, ok := mock.Responses[query]; ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Default success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&PrometheusResponse{
			Status: "success",
			Data: &struct {
				ResultType string                  `json:"resultType"`
				Result     []PrometheusQueryResult `json:"result"`
			}{
				ResultType: "vector",
				Result:     []PrometheusQueryResult{},
			},
		})
	}))

	return mock
}

func (m *MockPrometheusServer) Close() {
	m.Server.Close()
}

func (m *MockPrometheusServer) SetResponse(query string, response *PrometheusResponse) {
	m.Responses[query] = response
}

func (m *MockPrometheusServer) SetStatusCode(query string, code int) {
	m.StatusCodes[query] = code
}

// createPrometheusResponse creates a PrometheusResponse with the given result data
func createPrometheusResponse(results []PrometheusQueryResult) *PrometheusResponse {
	return &PrometheusResponse{
		Status: "success",
		Data: &struct {
			ResultType string                  `json:"resultType"`
			Result     []PrometheusQueryResult `json:"result"`
		}{
			ResultType: "vector",
			Result:     results,
		},
	}
}

// createMatrixResponse creates a PrometheusResponse with matrix result data
func createMatrixResponse(results []PrometheusQueryResult) *PrometheusResponse {
	return &PrometheusResponse{
		Status: "success",
		Data: &struct {
			ResultType string                  `json:"resultType"`
			Result     []PrometheusQueryResult `json:"result"`
		}{
			ResultType: "matrix",
			Result:     results,
		},
	}
}

func TestNewPrometheusClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "with custom URL",
			baseURL:     "http://custom-prometheus:9090",
			expectedURL: "http://custom-prometheus:9090",
		},
		{
			name:        "with empty URL leaves URL unset (Prometheus disabled)",
			baseURL:     "",
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewPrometheusClient(tt.baseURL)
			assert.NotNil(t, client)
			assert.Equal(t, tt.expectedURL, client.baseURL)
			assert.NotNil(t, client.httpClient)
			assert.NotNil(t, client.cache)
			assert.Equal(t, 30*time.Second, client.cacheDuration)
		})
	}
}

func TestPrometheusClient_Query_WhenNotConfigured(t *testing.T) {
	client := NewPrometheusClient("")
	ctx := context.Background()
	resp, err := client.query(ctx, "up")
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrPrometheusNotConfigured)
}

func TestPrometheusCache_GetSet(t *testing.T) {
	cache := &PrometheusCache{
		entries: make(map[string]cacheEntry),
	}

	response := &PrometheusResponse{
		Status: "success",
		Data: &struct {
			ResultType string                  `json:"resultType"`
			Result     []PrometheusQueryResult `json:"result"`
		}{
			ResultType: "vector",
			Result: []PrometheusQueryResult{
				{
					Metric: map[string]string{"job": "test"},
					Value:  []interface{}{float64(1234567890), "42"},
				},
			},
		},
	}

	t.Run("set and get cached value", func(t *testing.T) {
		cache.set("test-key", response, 1*time.Minute)
		result := cache.get("test-key")

		assert.NotNil(t, result)
		assert.Equal(t, "success", result.Status)
		assert.Len(t, result.Data.Result, 1)
	})

	t.Run("get non-existent key returns nil", func(t *testing.T) {
		result := cache.get("non-existent")
		assert.Nil(t, result)
	})

	t.Run("expired entry returns nil", func(t *testing.T) {
		cache.set("expired-key", response, 1*time.Nanosecond)
		time.Sleep(5 * time.Millisecond) // Wait for expiration
		result := cache.get("expired-key")
		assert.Nil(t, result)
	})

	t.Run("multiple keys can be cached", func(t *testing.T) {
		cache.set("key1", response, 1*time.Minute)
		cache.set("key2", response, 1*time.Minute)

		assert.NotNil(t, cache.get("key1"))
		assert.NotNil(t, cache.get("key2"))
	})
}

func TestPrometheusClient_Query(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)
	client.SetCacheDuration(100 * time.Millisecond)

	t.Run("successful query", func(t *testing.T) {
		mock.SetResponse(`up{job="test"}`, createPrometheusResponse([]PrometheusQueryResult{
			{
				Metric: map[string]string{"job": "test", "instance": "localhost:9090"},
				Value:  []interface{}{float64(1234567890), "1"},
			},
		}))

		ctx := context.Background()
		result, err := client.query(ctx, `up{job="test"}`)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "success", result.Status)
		assert.Len(t, result.Data.Result, 1)
	})

	t.Run("cached query avoids second request", func(t *testing.T) {
		query := `up{job="cached"}`
		mock.CallCount[query] = 0

		mock.SetResponse(query, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"job": "cached"}, Value: []interface{}{float64(1234567890), "1"}},
		}))

		ctx := context.Background()

		// First query - should hit the server
		_, err := client.query(ctx, query)
		require.NoError(t, err)
		firstCount := mock.CallCount[query]
		assert.Equal(t, 1, firstCount)

		// Second query - should use cache
		_, err = client.query(ctx, query)
		require.NoError(t, err)
		secondCount := mock.CallCount[query]
		assert.Equal(t, 1, secondCount, "Second query should use cache")
	})

	t.Run("expired cache triggers new request", func(t *testing.T) {
		query := `up{job="expiring"}`
		client.SetCacheDuration(50 * time.Millisecond)

		mock.SetResponse(query, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"job": "expiring"}, Value: []interface{}{float64(1234567890), "1"}},
		}))

		ctx := context.Background()

		// First query
		_, err := client.query(ctx, query)
		require.NoError(t, err)
		firstCount := mock.CallCount[query]

		// Wait for cache to expire
		time.Sleep(100 * time.Millisecond)

		// Second query - cache expired, should hit server again
		_, err = client.query(ctx, query)
		require.NoError(t, err)
		secondCount := mock.CallCount[query]

		assert.Equal(t, 1, firstCount)
		assert.Equal(t, 2, secondCount, "Should make new request after cache expires")
	})

	t.Run("prometheus returns error status", func(t *testing.T) {
		mock.SetStatusCode(`up{job="error"}`, http.StatusInternalServerError)

		ctx := context.Background()
		_, err := client.query(ctx, `up{job="error"}`)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prometheus returned status 500")
	})

	t.Run("network error handling", func(t *testing.T) {
		// Create client with invalid URL
		badClient := NewPrometheusClient("http://invalid-url-that-does-not-exist:9090")
		badClient.httpClient.Timeout = 100 * time.Millisecond

		ctx := context.Background()
		_, err := badClient.query(ctx, "up")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query prometheus")
	})
}

func TestPrometheusClient_QueryRange(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)

	t.Run("successful range query", func(t *testing.T) {
		start := time.Now().Add(-1 * time.Hour)
		end := time.Now()
		step := 5 * time.Minute

		mock.SetResponse("test_range_query", createMatrixResponse([]PrometheusQueryResult{
			{
				Metric: map[string]string{"job": "test"},
				Values: [][]interface{}{
					{float64(start.Unix()), "10"},
					{float64(start.Add(step).Unix()), "20"},
					{float64(start.Add(2 * step).Unix()), "30"},
				},
			},
		}))

		ctx := context.Background()
		result, err := client.queryRange(ctx, "test_range_query", start, end, step)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "success", result.Status)
	})

	t.Run("range query with caching", func(t *testing.T) {
		start := time.Now().Add(-1 * time.Hour)
		end := time.Now()
		step := 5 * time.Minute

		cacheKey := fmt.Sprintf("%s|%d|%d|%d", "cached_range", start.Unix(), end.Unix(), int(step.Seconds()))
		mock.Responses[cacheKey] = createMatrixResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"job": "cached"}, Values: [][]interface{}{{float64(1234567890), "1"}}},
		})

		ctx := context.Background()

		// First query
		_, err := client.queryRange(ctx, "cached_range", start, end, step)
		require.NoError(t, err)
		firstCount := mock.CallCount[cacheKey]

		// Second query - should use cache
		_, err = client.queryRange(ctx, "cached_range", start, end, step)
		require.NoError(t, err)
		secondCount := mock.CallCount[cacheKey]

		assert.Equal(t, 1, firstCount)
		assert.Equal(t, 1, secondCount, "Second query should use cache")
	})
}

func TestPrometheusClient_GetServiceHealth(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)

	mock.SetResponse(`up{job=~"orchestrator-api|health-monitor|postgres|redis"}`, createPrometheusResponse([]PrometheusQueryResult{
		{Metric: map[string]string{"job": "orchestrator-api"}, Value: []interface{}{float64(1234567890), "1"}},
		{Metric: map[string]string{"job": "health-monitor"}, Value: []interface{}{float64(1234567890), "1"}},
		{Metric: map[string]string{"job": "postgres"}, Value: []interface{}{float64(1234567890), "0"}},
		{Metric: map[string]string{"job": "redis"}, Value: []interface{}{float64(1234567890), "1"}},
	}))

	ctx := context.Background()
	health, err := client.GetServiceHealth(ctx)

	require.NoError(t, err)
	assert.Len(t, health, 4)
	assert.True(t, health["orchestrator-api"])
	assert.True(t, health["health-monitor"])
	assert.False(t, health["postgres"])
	assert.True(t, health["redis"])
}

func TestPrometheusClient_GetPlatformHealthPercentage(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)

	t.Run("returns calculated health percentage", func(t *testing.T) {
		mock.SetResponse(`(sum(functionfly_probe_success_rate > 0.95) / count(functionfly_probe_success_rate)) * 100`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{}, Value: []interface{}{float64(1234567890), "87.5"}},
		}))

		ctx := context.Background()
		percentage, err := client.GetPlatformHealthPercentage(ctx)

		require.NoError(t, err)
		assert.InDelta(t, 87.5, percentage, 0.01)
	})

	t.Run("returns 100 when no data", func(t *testing.T) {
		mock.SetResponse(`(sum(functionfly_probe_success_rate > 0.95) / count(functionfly_probe_success_rate)) * 100`, createPrometheusResponse([]PrometheusQueryResult{}))

		ctx := context.Background()
		percentage, err := client.GetPlatformHealthPercentage(ctx)

		require.NoError(t, err)
		assert.Equal(t, 100.0, percentage)
	})

	t.Run("handles prometheus error", func(t *testing.T) {
		mock.SetStatusCode(`(sum(functionfly_probe_success_rate > 0.95) / count(functionfly_probe_success_rate)) * 100`, http.StatusServiceUnavailable)

		ctx := context.Background()
		_, err := client.GetPlatformHealthPercentage(ctx)

		assert.Error(t, err)
	})
}

func TestPrometheusClient_GetActiveAlertCount(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)

	t.Run("returns alert count", func(t *testing.T) {
		mock.SetResponse(`sum(functionfly_active_alerts)`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{}, Value: []interface{}{float64(1234567890), "5"}},
		}))

		ctx := context.Background()
		count, err := client.GetActiveAlertCount(ctx)

		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("returns zero when no alerts", func(t *testing.T) {
		mock.SetResponse(`sum(functionfly_active_alerts)`, createPrometheusResponse([]PrometheusQueryResult{}))

		ctx := context.Background()
		count, err := client.GetActiveAlertCount(ctx)

		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestPrometheusClient_GetProbeLatency(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)

	t.Run("all providers query", func(t *testing.T) {
		mock.SetResponse(`histogram_quantile(0.95, sum(rate(functionfly_probe_latency_ms_bucket[5m])) by (provider, region, le))`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"provider": "fly", "region": "iad"}, Value: []interface{}{float64(1234567890), "45.2"}},
			{Metric: map[string]string{"provider": "vercel", "region": "iad"}, Value: []interface{}{float64(1234567890), "32.1"}},
		}))

		ctx := context.Background()
		result, err := client.GetProbeLatency(ctx, "", "", 0.95)

		require.NoError(t, err)
		assert.NotNil(t, result.Data)
		assert.Len(t, result.Data.Result, 2)
	})

	t.Run("specific provider query", func(t *testing.T) {
		mock.SetResponse(`histogram_quantile(0.95, sum(rate(functionfly_probe_latency_ms_bucket{provider="fly"}[5m])) by (region, le))`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"region": "iad"}, Value: []interface{}{float64(1234567890), "45.2"}},
		}))

		ctx := context.Background()
		result, err := client.GetProbeLatency(ctx, "fly", "", 0.95)

		require.NoError(t, err)
		assert.NotNil(t, result.Data)
	})

	t.Run("specific provider and region query", func(t *testing.T) {
		mock.SetResponse(`histogram_quantile(0.99, sum(rate(functionfly_probe_latency_ms_bucket{provider="fly",region="iad"}[5m])) by (le))`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{}, Value: []interface{}{float64(1234567890), "67.8"}},
		}))

		ctx := context.Background()
		result, err := client.GetProbeLatency(ctx, "fly", "iad", 0.99)

		require.NoError(t, err)
		assert.NotNil(t, result.Data)
		assert.Len(t, result.Data.Result, 1)
	})
}

func TestPrometheusClient_GetErrorRate(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)

	t.Run("all providers", func(t *testing.T) {
		mock.SetResponse(`sum(rate(functionfly_request_error_rate[10m])) by (backend, provider)`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"provider": "fly", "backend": "backend1"}, Value: []interface{}{float64(1234567890), "0.01"}},
			{Metric: map[string]string{"provider": "vercel", "backend": "backend2"}, Value: []interface{}{float64(1234567890), "0.02"}},
		}))

		ctx := context.Background()
		result, err := client.GetErrorRate(ctx, "")

		require.NoError(t, err)
		assert.NotNil(t, result.Data)
		assert.Len(t, result.Data.Result, 2)
	})

	t.Run("specific provider", func(t *testing.T) {
		mock.SetResponse(`sum(rate(functionfly_request_error_rate{provider="fly"}[10m])) by (backend)`, createPrometheusResponse([]PrometheusQueryResult{
			{Metric: map[string]string{"backend": "backend1"}, Value: []interface{}{float64(1234567890), "0.01"}},
		}))

		ctx := context.Background()
		result, err := client.GetErrorRate(ctx, "fly")

		require.NoError(t, err)
		assert.NotNil(t, result.Data)
	})
}

func TestPrometheusMetrics_GetProviderLatencies(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)
	metrics := NewPrometheusMetrics(client)

	mock.SetResponse(`histogram_quantile(0.95, sum(rate(functionfly_probe_latency_ms_bucket[5m])) by (provider, region, le))`, createPrometheusResponse([]PrometheusQueryResult{
		{Metric: map[string]string{"provider": "fly"}, Value: []interface{}{float64(1234567890), "45.2"}},
		{Metric: map[string]string{"provider": "vercel"}, Value: []interface{}{float64(1234567890), "32.1"}},
		{Metric: map[string]string{"provider": "workers"}, Value: []interface{}{float64(1234567890), "28.5"}},
	}))

	ctx := context.Background()
	latencies, err := metrics.GetProviderLatencies(ctx, 0.95)

	require.NoError(t, err)
	assert.Len(t, latencies, 3)
	assert.InDelta(t, 45.2, latencies["fly"], 0.01)
	assert.InDelta(t, 32.1, latencies["vercel"], 0.01)
	assert.InDelta(t, 28.5, latencies["workers"], 0.01)
}

func TestPrometheusMetrics_GetProviderErrorRates(t *testing.T) {
	mock := NewMockPrometheusServer()
	defer mock.Close()

	client := NewPrometheusClient(mock.Server.URL)
	metrics := NewPrometheusMetrics(client)

	mock.SetResponse(`sum(rate(functionfly_request_error_rate[10m])) by (backend, provider)`, createPrometheusResponse([]PrometheusQueryResult{
		{Metric: map[string]string{"provider": "fly"}, Value: []interface{}{float64(1234567890), "0.01"}},
		{Metric: map[string]string{"provider": "vercel"}, Value: []interface{}{float64(1234567890), "0.02"}},
	}))

	ctx := context.Background()
	rates, err := metrics.GetProviderErrorRates(ctx)

	require.NoError(t, err)
	assert.Len(t, rates, 2)
	assert.InDelta(t, 0.01, rates["fly"], 0.001)
	assert.InDelta(t, 0.02, rates["vercel"], 0.001)
}

func TestFormatPrometheusTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Unix epoch",
			input:    time.Unix(0, 0),
			expected: "0",
		},
		{
			name:     "Specific timestamp",
			input:    time.Unix(1609459200, 0),
			expected: "1609459200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPrometheusTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrometheusClient_SetCacheDuration(t *testing.T) {
	client := NewPrometheusClient("http://localhost:9090")

	assert.Equal(t, 30*time.Second, client.cacheDuration)

	client.SetCacheDuration(5 * time.Minute)
	assert.Equal(t, 5*time.Minute, client.cacheDuration)

	client.SetCacheDuration(0)
	assert.Equal(t, time.Duration(0), client.cacheDuration)
}
