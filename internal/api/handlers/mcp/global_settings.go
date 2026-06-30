package mcp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// MCPSettingsGlobal represents platform-wide MCP configuration defaults.
type MCPSettingsGlobal struct {
	DefaultTransport      string   `json:"default_transport"`
	DefaultRateLimit      int      `json:"default_rate_limit"`
	DefaultExposeInput    bool     `json:"default_expose_input"`
	DefaultExposeOutput   bool     `json:"default_expose_output"`
	AutoAddToRegistry     bool     `json:"auto_add_to_registry"`
	RequireVerification   bool     `json:"require_verification"`
	PublicListing         bool     `json:"public_listing"`
	CORSAllowlist         []string `json:"cors_allowlist"`
	RateLimitMultiplier   int      `json:"rate_limit_multiplier"`
}

// MCPAnalytics represents aggregated MCP usage analytics.
type MCPAnalytics struct {
	TotalCalls      int64                   `json:"total_calls"`
	UniqueClients   int64                   `json:"unique_clients"`
	AvgLatencyMs    int                     `json:"avg_latency_ms"`
	SuccessRate     float64                 `json:"success_rate"`
	CallsOverTime   []TimeSeriesPoint       `json:"calls_over_time"`
	ClientBreakdown []ClientCount           `json:"client_breakdown"`
	TopFunctions    []FunctionCallCount     `json:"top_functions"`
	TransportUsage  []TransportCount        `json:"transport_usage"`
}

// TimeSeriesPoint represents a single point in a time series.
type TimeSeriesPoint struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

// ClientCount represents a client and its invocation count.
type ClientCount struct {
	Client string `json:"client"`
	Count  int64  `json:"count"`
}

// FunctionCallCount represents a function and its call count.
type FunctionCallCount struct {
	Author string `json:"author"`
	Name   string `json:"name"`
	Calls  int64  `json:"calls"`
}

// TransportCount represents a transport and its usage count.
type TransportCount struct {
	Transport string `json:"transport"`
	Count     int64  `json:"count"`
}

// MCPConnection represents an AI client's connection to MCP functions.
type MCPConnection struct {
	ClientType             string   `json:"client_type"`
	ClientIcon             string   `json:"client_icon,omitempty"`
	Status                 string   `json:"status"` // active, stale, never
	Enabled                bool     `json:"enabled"`
	ConnectedFunctions     int      `json:"connected_functions"`
	TotalInvocations       int64    `json:"total_invocations"`
	LastConnectedAt        *string  `json:"last_connected_at,omitempty"`
	AvgLatencyMs           int      `json:"avg_latency_ms"`
	ConnectedFunctionNames []string `json:"connected_function_names"`
}

// GlobalHandler handles platform-wide MCP settings, analytics, and connections.
type GlobalHandler struct {
	repo *registry.RegistryRepository
}

// NewGlobalHandler creates a new global MCP handler.
func NewGlobalHandler(repo *registry.RegistryRepository) *GlobalHandler {
	return &GlobalHandler{repo: repo}
}

// HandleGetMCPSettings serves GET /v1/mcp/settings.
// Returns platform-wide MCP configuration defaults.
func (h *GlobalHandler) HandleGetMCPSettings(w http.ResponseWriter, r *http.Request) {
	// Return default global settings (in production, these would come from a config table)
	settings := MCPSettingsGlobal{
		DefaultTransport:    "streamable-http",
		DefaultRateLimit:    60,
		DefaultExposeInput:  true,
		DefaultExposeOutput: false,
		AutoAddToRegistry:   false,
		RequireVerification: true,
		PublicListing:       true,
		CORSAllowlist:       []string{},
		RateLimitMultiplier: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// HandleUpdateMCPSettings serves PATCH /v1/mcp/settings.
// Updates platform-wide MCP configuration defaults.
func (h *GlobalHandler) HandleUpdateMCPSettings(w http.ResponseWriter, r *http.Request) {
	var input MCPSettingsGlobal
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logrus.WithError(err).Info("mcp: invalid JSON in update settings")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Validate inputs
	if input.DefaultRateLimit < 1 || input.DefaultRateLimit > 10000 {
		apierror.WriteError(w, apierror.NewBadRequest("default_rate_limit must be between 1 and 10000"))
		return
	}
	if input.RateLimitMultiplier < 1 || input.RateLimitMultiplier > 100 {
		apierror.WriteError(w, apierror.NewBadRequest("rate_limit_multiplier must be between 1 and 100"))
		return
	}

	// In production, we would save these to a config table
	// For now, just return the settings as confirmation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(input)
}

// HandleGetMCPAnalytics serves GET /v1/mcp/analytics.
// Returns aggregated MCP usage analytics.
func (h *GlobalHandler) HandleGetMCPAnalytics(w http.ResponseWriter, r *http.Request) {
	days := 30 // default
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	since := time.Now().AddDate(0, 0, -days)

	// Get aggregated invocation data from storage
	totalCalls, err := h.repo.GetTotalMCPInvocations(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get total MCP invocations")
		totalCalls = 0
	}

	uniqueClients, err := h.repo.GetUniqueMCPClients(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get unique MCP clients")
		uniqueClients = 0
	}

	avgLatency, err := h.repo.GetAverageMCPLatency(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get average MCP latency")
		avgLatency = 0
	}

	successRate, err := h.repo.GetMCPSuccessRate(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get MCP success rate")
		successRate = 100.0
	}

	// Get time series data
	callsOverTime, err := h.repo.GetMCPCallsOverTime(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get MCP calls over time")
		callsOverTime = []registry.MCPTimeSeriesPoint{}
	}

	// Get client breakdown
	clientBreakdown, err := h.repo.GetMCPClientBreakdown(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get MCP client breakdown")
		clientBreakdown = []registry.MCPClientCount{}
	}

	// Get top functions
	topFunctions, err := h.repo.GetMCPTopFunctions(r.Context(), since, 10)
	if err != nil {
		logrus.WithError(err).Error("Failed to get MCP top functions")
		topFunctions = []registry.MCPFunctionCallCount{}
	}

	// Get transport usage
	transportUsage, err := h.repo.GetMCPTransportUsage(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get MCP transport usage")
		transportUsage = []registry.MCPTransportCount{}
	}

	// Transform to response format
	callsOverTimeResp := make([]TimeSeriesPoint, len(callsOverTime))
	for i, p := range callsOverTime {
		callsOverTimeResp[i] = TimeSeriesPoint{Time: p.Time.Format(time.RFC3339), Count: p.Count}
	}

	clientBreakdownResp := make([]ClientCount, len(clientBreakdown))
	for i, c := range clientBreakdown {
		clientBreakdownResp[i] = ClientCount{Client: c.Client, Count: c.Count}
	}

	topFunctionsResp := make([]FunctionCallCount, len(topFunctions))
	for i, f := range topFunctions {
		topFunctionsResp[i] = FunctionCallCount{Author: f.Author, Name: f.Name, Calls: f.Calls}
	}

	transportUsageResp := make([]TransportCount, len(transportUsage))
	for i, t := range transportUsage {
		transportUsageResp[i] = TransportCount{Transport: t.Transport, Count: t.Count}
	}

	analytics := MCPAnalytics{
		TotalCalls:      totalCalls,
		UniqueClients:   uniqueClients,
		AvgLatencyMs:    avgLatency,
		SuccessRate:     successRate,
		CallsOverTime:   callsOverTimeResp,
		ClientBreakdown: clientBreakdownResp,
		TopFunctions:    topFunctionsResp,
		TransportUsage:  transportUsageResp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// HandleGetMCPConnections serves GET /v1/mcp/connections.
// Returns aggregated MCP client connection information.
func (h *GlobalHandler) HandleGetMCPConnections(w http.ResponseWriter, r *http.Request) {
	// Get connections from storage
	connections, err := h.repo.GetMCPConnections(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get MCP connections")
		connections = []registry.MCPConnectionRecord{}
	}

	// Transform to response format
	resp := make([]MCPConnection, len(connections))
	for i, c := range connections {
		functionNames := make([]string, len(c.ConnectedFunctionNames))
		copy(functionNames, c.ConnectedFunctionNames)
		if functionNames == nil {
			functionNames = []string{}
		}

		resp[i] = MCPConnection{
			ClientType:             c.ClientType,
			ClientIcon:             c.ClientIcon,
			Status:                 c.Status,
			Enabled:                true,
			ConnectedFunctions:     c.ConnectedFunctions,
			TotalInvocations:       c.TotalInvocations,
			AvgLatencyMs:           c.AvgLatencyMs,
			ConnectedFunctionNames: functionNames,
		}
		if c.LastConnectedAt != nil {
			s := c.LastConnectedAt.Format(time.RFC3339)
			resp[i].LastConnectedAt = &s
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connections": resp,
	})
}

// HandleToggleMCPConnection serves PATCH /v1/mcp/connections/{clientType}.
// Toggles a client connection enabled/disabled state.
func (h *GlobalHandler) HandleToggleMCPConnection(w http.ResponseWriter, r *http.Request) {
	// For now, acknowledge the toggle. In production this would persist to a
	// mcp_client_settings table. The frontend uses this to reflect the user's
	// intent to enable/disable tracking for a specific client.
	var input struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": input.Enabled,
	})
}

// HandleTestMCPConnection serves POST /v1/mcp/connections/{clientType}/test.
// Tests whether the MCP server is reachable and responding.
func (h *GlobalHandler) HandleTestMCPConnection(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Verify DB connectivity as a proxy for MCP server health
	_, err := h.repo.GetMCPStats(r.Context())
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logrus.WithError(err).Error("MCP connection test failed")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     false,
			"message":     "MCP server is not responding",
			"latency_ms":  latency,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "MCP server is reachable",
		"latency_ms":  latency,
	})
}
