package status

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrPrometheusNotConfigured is returned when PROMETHEUS_URL is not set and no queries are made.
var ErrPrometheusNotConfigured = errors.New("prometheus not configured")

// PrometheusClient provides methods to query Prometheus metrics
type PrometheusClient struct {
	baseURL       string
	httpClient    *http.Client
	cacheDuration time.Duration
	cache         *PrometheusCache
}

// PrometheusCache provides simple caching for Prometheus queries
type PrometheusCache struct {
	entries map[string]cacheEntry
}

type cacheEntry struct {
	value      *PrometheusResponse
	expiration time.Time
}

// NewPrometheusClient creates a new Prometheus client.
// When baseURL is empty, all queries return ErrPrometheusNotConfigured (no network calls, no DNS lookup).
func NewPrometheusClient(baseURL string) *PrometheusClient {
	timeout := 3 * time.Second // fail fast on DNS/connect when Prometheus is unreachable
	return &PrometheusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cacheDuration: 30 * time.Second,
		cache: &PrometheusCache{
			entries: make(map[string]cacheEntry),
		},
	}
}

// SetCacheDuration sets the cache duration
func (c *PrometheusClient) SetCacheDuration(duration time.Duration) {
	c.cacheDuration = duration
}

// query performs a Prometheus instant query with caching
func (c *PrometheusClient) query(ctx context.Context, query string) (*PrometheusResponse, error) {
	if c.baseURL == "" {
		return nil, ErrPrometheusNotConfigured
	}
	// Check cache first
	if cached := c.cache.get(query); cached != nil {
		return cached, nil
	}

	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse prometheus URL: %w", err)
	}

	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	// Execute request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var result PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode prometheus response: %w", err)
	}

	// Cache result
	c.cache.set(query, &result, c.cacheDuration)

	return &result, nil
}

// queryRange performs a Prometheus range query with caching
func (c *PrometheusClient) queryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*PrometheusResponse, error) {
	if c.baseURL == "" {
		return nil, ErrPrometheusNotConfigured
	}
	cacheKey := fmt.Sprintf("%s|%d|%d|%d", query, start.Unix(), end.Unix(), int(step.Seconds()))

	// Check cache first
	if cached := c.cache.get(cacheKey); cached != nil {
		return cached, nil
	}

	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query_range", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse prometheus URL: %w", err)
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("start", formatPrometheusTime(start))
	q.Set("end", formatPrometheusTime(end))
	q.Set("step", strconv.Itoa(int(step.Seconds())))
	u.RawQuery = q.Encode()

	// Execute request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var result PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode prometheus response: %w", err)
	}

	// Cache result
	c.cache.set(cacheKey, &result, c.cacheDuration)

	return &result, nil
}

// formatPrometheusTime formats time for Prometheus API
func formatPrometheusTime(t time.Time) string {
	return strconv.FormatFloat(float64(t.Unix()), 'f', -1, 64)
}

// get retrieves a cached value if it exists and is not expired
func (c *PrometheusCache) get(key string) *PrometheusResponse {
	entry, exists := c.entries[key]
	if !exists {
		return nil
	}
	if time.Now().After(entry.expiration) {
		delete(c.entries, key)
		return nil
	}
	return entry.value
}

// set stores a value in the cache with expiration
func (c *PrometheusCache) set(key string, value *PrometheusResponse, duration time.Duration) {
	c.entries[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(duration),
	}
}

// GetServiceHealth returns health status of core services
func (c *PrometheusClient) GetServiceHealth(ctx context.Context) (map[string]bool, error) {
	query := `up{job=~"orchestrator-api|health-monitor|postgres|redis"}`
	resp, err := c.query(ctx, query)
	if err != nil {
		return nil, err
	}

	health := make(map[string]bool)
	if resp.Data != nil {
		for _, result := range resp.Data.Result {
			job := result.Metric["job"]
			if len(result.Value) >= 2 {
				valueStr, ok := result.Value[1].(string)
				if ok {
					value, _ := strconv.ParseFloat(valueStr, 64)
					// If multiple targets per job (e.g. container + host.docker.internal), consider up if any is up
					health[job] = health[job] || value > 0
				}
			}
		}
	}

	return health, nil
}

// GetBackendHealthStatus returns health status for FunctionFly backends
func (c *PrometheusClient) GetBackendHealthStatus(ctx context.Context) (*PrometheusResponse, error) {
	query := `functionfly_backend_health_status`
	return c.query(ctx, query)
}

// GetProbeSuccessRate returns probe success rates by backend
func (c *PrometheusClient) GetProbeSuccessRate(ctx context.Context, provider string) (*PrometheusResponse, error) {
	query := `functionfly_probe_success_rate`
	if provider != "" && provider != "all" {
		query = fmt.Sprintf(`functionfly_probe_success_rate{provider="%s"}`, provider)
	}
	return c.query(ctx, query)
}

// GetProbeLatency returns probe latency metrics
func (c *PrometheusClient) GetProbeLatency(ctx context.Context, provider, region string, percentile float64) (*PrometheusResponse, error) {
	query := fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(functionfly_probe_latency_ms_bucket[5m])) by (provider, region, le))`, percentile)

	if provider != "" && provider != "all" {
		if region != "" && region != "all" {
			query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(functionfly_probe_latency_ms_bucket{provider="%s",region="%s"}[5m])) by (le))`, percentile, provider, region)
		} else {
			query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(functionfly_probe_latency_ms_bucket{provider="%s"}[5m])) by (region, le))`, percentile, provider)
		}
	}

	return c.query(ctx, query)
}

// GetLatencyRange returns latency over a time range
func (c *PrometheusClient) GetLatencyRange(ctx context.Context, provider string, start, end time.Time, step time.Duration, percentile float64) (*PrometheusResponse, error) {
	query := fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(functionfly_probe_latency_ms_bucket[5m])) by (provider, le))`, percentile)

	if provider != "" && provider != "all" {
		query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(functionfly_probe_latency_ms_bucket{provider="%s"}[5m])) by (le))`, percentile, provider)
	}

	return c.queryRange(ctx, query, start, end, step)
}

// GetComponentHTTPLatency returns the average HTTP request latency for a component in milliseconds
func (c *PrometheusClient) GetComponentHTTPLatency(ctx context.Context, component string, duration string) (float64, error) {
	// Map component names to Prometheus job names or use a generic query
	var query string
	switch component {
	case "api", "API":
		query = fmt.Sprintf(`avg(rate(functionfly_http_request_duration_seconds_sum[%s]) / rate(functionfly_http_request_duration_seconds_count[%s])) * 1000`, duration, duration)
	case "database", "Database":
		query = fmt.Sprintf(`avg(rate(functionfly_db_query_duration_seconds_sum[%s]) / rate(functionfly_db_query_duration_seconds_count[%s])) * 1000`, duration, duration)
	case "cache", "Cache":
		query = fmt.Sprintf(`avg(rate(functionfly_cache_operation_duration_seconds_sum[%s]) / rate(functionfly_cache_operation_duration_seconds_count[%s])) * 1000`, duration, duration)
	default:
		// Generic HTTP latency
		query = fmt.Sprintf(`avg(rate(functionfly_http_request_duration_seconds_sum[%s]) / rate(functionfly_http_request_duration_seconds_count[%s])) * 1000`, duration, duration)
	}

	resp, err := c.query(ctx, query)
	if err != nil {
		return 0, err
	}

	if resp.Data != nil && len(resp.Data.Result) > 0 {
		if len(resp.Data.Result[0].Value) >= 2 {
			return parseValue(resp.Data.Result[0].Value[1]), nil
		}
	}

	return 0, nil // No data available
}

// GetErrorRate returns error rates by backend/provider
func (c *PrometheusClient) GetErrorRate(ctx context.Context, provider string) (*PrometheusResponse, error) {
	query := `sum(rate(functionfly_request_error_rate[10m])) by (backend, provider)`
	if provider != "" && provider != "all" {
		query = fmt.Sprintf(`sum(rate(functionfly_request_error_rate{provider="%s"}[10m])) by (backend)`, provider)
	}
	return c.query(ctx, query)
}

// GetUptimeRatio returns uptime ratio for a time period
func (c *PrometheusClient) GetUptimeRatio(ctx context.Context, component, provider string, duration string) (*PrometheusResponse, error) {
	query := fmt.Sprintf(`avg_over_time(functionfly_uptime_ratio[%s]) * 100`, duration)

	if component != "" && component != "all" {
		query = fmt.Sprintf(`avg_over_time(functionfly_uptime_ratio{component="%s"}[%s]) * 100`, component, duration)
	} else if provider != "" && provider != "all" {
		query = fmt.Sprintf(`avg_over_time(functionfly_uptime_ratio{provider="%s"}[%s]) * 100`, provider, duration)
	}

	return c.query(ctx, query)
}

// GetUptimeRange returns uptime data over a time range
func (c *PrometheusClient) GetUptimeRange(ctx context.Context, component string, start, end time.Time, step time.Duration) (*PrometheusResponse, error) {
	query := `functionfly_uptime_ratio * 100`
	if component != "" && component != "all" {
		query = fmt.Sprintf(`functionfly_uptime_ratio{component="%s"} * 100`, component)
	}
	return c.queryRange(ctx, query, start, end, step)
}

// GetCircuitState returns circuit breaker states
func (c *PrometheusClient) GetCircuitState(ctx context.Context) (*PrometheusResponse, error) {
	query := `functionfly_circuit_state`
	return c.query(ctx, query)
}

// GetRequestVolume returns request volume by provider
func (c *PrometheusClient) GetRequestVolume(ctx context.Context, provider string) (*PrometheusResponse, error) {
	query := `sum(rate(functionfly_requests_total[5m])) by (provider)`
	if provider != "" && provider != "all" {
		query = fmt.Sprintf(`sum(rate(functionfly_requests_total{provider="%s"}[5m]))`, provider)
	}
	return c.query(ctx, query)
}

// GetActiveAlertCount returns the count of active alerts
func (c *PrometheusClient) GetActiveAlertCount(ctx context.Context) (int, error) {
	query := `sum(functionfly_active_alerts)`
	resp, err := c.query(ctx, query)
	if err != nil {
		return 0, err
	}

	if resp.Data != nil && len(resp.Data.Result) > 0 {
		if len(resp.Data.Result[0].Value) >= 2 {
			valueStr, ok := resp.Data.Result[0].Value[1].(string)
			if ok {
				value, _ := strconv.ParseFloat(valueStr, 64)
				return int(value), nil
			}
		}
	}

	return 0, nil
}

// GetPlatformHealthPercentage returns overall platform health percentage
func (c *PrometheusClient) GetPlatformHealthPercentage(ctx context.Context) (float64, error) {
	query := `(sum(functionfly_probe_success_rate > 0.95) / count(functionfly_probe_success_rate)) * 100`
	resp, err := c.query(ctx, query)
	if err != nil {
		return 0, err
	}

	if resp.Data != nil && len(resp.Data.Result) > 0 {
		if len(resp.Data.Result[0].Value) >= 2 {
			valueStr, ok := resp.Data.Result[0].Value[1].(string)
			if ok {
				value, _ := strconv.ParseFloat(valueStr, 64)
				return value, nil
			}
		}
	}

	return 100.0, nil // Default to 100% if no data
}

// parseValue parses a Prometheus value string to float64
func parseValue(v interface{}) float64 {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case float64:
		return val
	default:
		return 0
	}
}

// PrometheusMetrics provides helper methods to extract metrics from Prometheus responses
type PrometheusMetrics struct {
	client *PrometheusClient
}

// NewPrometheusMetrics creates a new Prometheus metrics helper
func NewPrometheusMetrics(client *PrometheusClient) *PrometheusMetrics {
	return &PrometheusMetrics{client: client}
}

// GetProviderLatencies returns latencies for all providers
func (m *PrometheusMetrics) GetProviderLatencies(ctx context.Context, percentile float64) (map[string]float64, error) {
	resp, err := m.client.GetProbeLatency(ctx, "", "", percentile)
	if err != nil {
		return nil, err
	}

	latencies := make(map[string]float64)
	if resp.Data != nil {
		for _, result := range resp.Data.Result {
			provider := result.Metric["provider"]
			if provider != "" && len(result.Value) >= 2 {
				latencies[provider] = parseValue(result.Value[1])
			}
		}
	}

	return latencies, nil
}

// GetProviderErrorRates returns error rates for all providers
func (m *PrometheusMetrics) GetProviderErrorRates(ctx context.Context) (map[string]float64, error) {
	resp, err := m.client.GetErrorRate(ctx, "")
	if err != nil {
		return nil, err
	}

	rates := make(map[string]float64)
	if resp.Data != nil {
		for _, result := range resp.Data.Result {
			provider := result.Metric["provider"]
			if provider != "" && len(result.Value) >= 2 {
				rates[provider] = parseValue(result.Value[1])
			}
		}
	}

	return rates, nil
}

// GetProviderHealthSummary returns a summary of provider health
func (m *PrometheusMetrics) GetProviderHealthSummary(ctx context.Context) (map[string]map[string]interface{}, error) {
	// Get success rates
	successResp, err := m.client.GetProbeSuccessRate(ctx, "")
	if err != nil {
		logrus.WithError(err).Warn("Failed to get probe success rates")
	}

	// Get latencies
	latencyResp, err := m.client.GetProbeLatency(ctx, "", "", 0.95)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get probe latencies")
	}

	// Get circuit states
	circuitResp, err := m.client.GetCircuitState(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get circuit states")
	}

	summary := make(map[string]map[string]interface{})

	// Process success rates
	if successResp != nil && successResp.Data != nil {
		for _, result := range successResp.Data.Result {
			provider := result.Metric["provider"]
			if provider != "" {
				if _, exists := summary[provider]; !exists {
					summary[provider] = make(map[string]interface{})
				}
				if len(result.Value) >= 2 {
					summary[provider]["success_rate"] = parseValue(result.Value[1])
				}
			}
		}
	}

	// Process latencies
	if latencyResp != nil && latencyResp.Data != nil {
		for _, result := range latencyResp.Data.Result {
			provider := result.Metric["provider"]
			if provider != "" {
				if _, exists := summary[provider]; !exists {
					summary[provider] = make(map[string]interface{})
				}
				if len(result.Value) >= 2 {
					summary[provider]["latency_p95_ms"] = parseValue(result.Value[1])
				}
			}
		}
	}

	// Process circuit states
	if circuitResp != nil && circuitResp.Data != nil {
		for _, result := range circuitResp.Data.Result {
			backend := result.Metric["backend"]
			if backend != "" && len(result.Value) >= 2 {
				state := parseValue(result.Value[1])
				// circuit_state: 0=closed, 1=half-open, 2=open
				stateStr := "closed"
				if state == 1 {
					stateStr = "half-open"
				} else if state == 2 {
					stateStr = "open"
				}

				// Group by provider if available in metric
				provider := result.Metric["provider"]
				if provider != "" {
					if _, exists := summary[provider]; !exists {
						summary[provider] = make(map[string]interface{})
					}
					// Count circuits per state
					key := "circuits_" + stateStr
					if count, ok := summary[provider][key].(int); ok {
						summary[provider][key] = count + 1
					} else {
						summary[provider][key] = 1
					}
				}
			}
		}
	}

	return summary, nil
}
