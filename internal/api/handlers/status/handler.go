package status

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains status page handlers
type Handler struct {
	repo       RepositoryInterface
	prometheus *PrometheusClient
	metrics    *PrometheusMetrics
	authSvc    *auth.AuthService
	statusHub  *StatusWebSocketHub // optional; when set, HandleWebSocketStatus uses this shared hub
}

// NewHandler creates a new status page handler
func NewHandler(repo RepositoryInterface, prometheusURL string, authSvc *auth.AuthService) *Handler {
	promClient := NewPrometheusClient(prometheusURL)
	return &Handler{
		repo:       repo,
		prometheus: promClient,
		metrics:    NewPrometheusMetrics(promClient),
		authSvc:    authSvc,
	}
}

// SetPrometheusClient allows setting a custom prometheus client (useful for testing)
func (h *Handler) SetPrometheusClient(client *PrometheusClient) {
	h.prometheus = client
	h.metrics = NewPrometheusMetrics(client)
}

// SetStatusHub sets the shared WebSocket hub used by HandleWebSocketStatus when present.
// Routes wire a single hub at startup (see routes.go: statusWSHub) so all /ws/v1/status connections
// use one hub; if unset (e.g. in tests), HandleWebSocketStatus creates a temporary hub per request.
func (h *Handler) SetStatusHub(hub *StatusWebSocketHub) {
	h.statusHub = hub
}

// --- Status Endpoints ---

// HandleGetPlatformStatus returns overall platform status
func (h *Handler) HandleGetPlatformStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get platform health percentage from Prometheus
	var healthPercent float64
	var prometheusErr error

	healthPercent, prometheusErr = h.prometheus.GetPlatformHealthPercentage(ctx)
	if prometheusErr != nil {
		if !errors.Is(prometheusErr, ErrPrometheusNotConfigured) {
			logrus.WithError(prometheusErr).Warn("Failed to get platform health from Prometheus")
		}
		healthPercent = 100.0 // Default to healthy
	}

	// Get active incidents
	incidents, err := h.repo.GetActiveIncidents(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get active incidents")
		incidents = []Incident{}
	}

	// Get upcoming maintenance
	maintenance, err := h.repo.GetUpcomingMaintenance(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get upcoming maintenance")
		maintenance = []MaintenanceWindow{}
	}

	// Determine platform status
	status, indicator, description := h.determinePlatformStatus(healthPercent, incidents)

	// Get component summaries
	components := h.getComponentSummaries(ctx)

	// Convert maintenance to summaries
	maintenanceSummaries := make([]MaintenanceSummary, len(maintenance))
	for i, m := range maintenance {
		maintenanceSummaries[i] = MaintenanceSummary{
			ID:             m.ID,
			Title:          m.Title,
			Status:         m.Status,
			ScheduledStart: m.ScheduledStart,
			ScheduledEnd:   m.ScheduledEnd,
		}
	}

	response := PlatformStatus{
		Status:      status,
		Indicator:   indicator,
		Description: description,
		UpdatedAt:   time.Now(),
		Components:  components,
		Incidents:   incidents,
		Maintenance: maintenanceSummaries,
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleGetComponents returns detailed component status
func (h *Handler) HandleGetComponents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	includeHistory := r.URL.Query().Get("include_history") == "true"
	componentType := r.URL.Query().Get("component_type")

	// Get system health checks from database; if empty, use default summaries so UI always has data
	components, err := h.repo.GetSystemHealthChecks(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get system health checks")
		http.Error(w, "Failed to get component status", http.StatusInternalServerError)
		return
	}
	if len(components) == 0 {
		components = h.getComponentSummaries(ctx)
	}

	// Filter by component type if specified
	if componentType != "" {
		filtered := make([]Component, 0)
		for _, c := range components {
			if c.Type == componentType {
				filtered = append(filtered, c)
			}
		}
		components = filtered
	}

	// Calculate real uptime percentages from health check history
	for i := range components {
		// Calculate uptime for different periods
		uptime24h, err := h.repo.CalculateComponentUptime(ctx, components[i].Name, 24*time.Hour)
		if err == nil {
			components[i].Uptime24h = uptime24h
		} else {
			// Fallback to Prometheus if available
			if uptime, err := h.getUptimeForComponent(ctx, components[i].Name, "24h"); err == nil {
				components[i].Uptime24h = uptime
			} else {
				components[i].Uptime24h = 99.9 // Final fallback
			}
		}

		uptime7d, err := h.repo.CalculateComponentUptime(ctx, components[i].Name, 7*24*time.Hour)
		if err == nil {
			components[i].Uptime7d = uptime7d
		} else {
			components[i].Uptime7d = components[i].Uptime24h // Use 24h as fallback
		}

		uptime30d, err := h.repo.CalculateComponentUptime(ctx, components[i].Name, 30*24*time.Hour)
		if err == nil {
			components[i].Uptime30d = uptime30d
		} else {
			components[i].Uptime30d = components[i].Uptime7d // Use 7d as fallback
		}

		// Get history if requested
		if includeHistory {
			since := time.Now().Add(-24 * time.Hour)
			history, err := h.repo.GetComponentHealthHistory(ctx, components[i].Name, since)
			if err == nil {
				components[i].History = history
			}
		}
	}

	response := ComponentStatusResponse{
		Components:  components,
		GeneratedAt: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleGetProviders returns per-provider status by region
func (h *Handler) HandleGetProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	providerFilter := r.URL.Query().Get("provider")
	regionFilter := r.URL.Query().Get("region")
	detailed := r.URL.Query().Get("detailed") == "true"

	// Get provider status from database
	providers, err := h.repo.GetProviderStatus(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get provider status")
		http.Error(w, "Failed to get provider status", http.StatusInternalServerError)
		return
	}

	// Filter by provider if specified
	if providerFilter != "" && providerFilter != "all" {
		filtered := make([]ProviderStatus, 0)
		for _, p := range providers {
			if p.Name == providerFilter {
				filtered = append(filtered, p)
				break
			}
		}
		providers = filtered
	}

	// Enhance with Prometheus metrics
	for i := range providers {
		// Get latencies
		latency, err := h.prometheus.GetProbeLatency(ctx, providers[i].Name, "", 0.95)
		if err == nil && latency.Data != nil {
			for _, result := range latency.Data.Result {
				if len(result.Value) >= 2 {
					val := parseValue(result.Value[1])
					providers[i].Summary.AvgLatencyMs = val
				}
			}
		}

		// Get error rates
		errorRate, err := h.prometheus.GetErrorRate(ctx, providers[i].Name)
		if err == nil && errorRate.Data != nil {
			totalErrorRate := 0.0
			for _, result := range errorRate.Data.Result {
				if len(result.Value) >= 2 {
					totalErrorRate += parseValue(result.Value[1])
				}
			}
			providers[i].Summary.ErrorRate = totalErrorRate
		}

		// Filter regions if specified
		if regionFilter != "" && regionFilter != "all" {
			filteredRegions := make([]RegionStatus, 0)
			for _, region := range providers[i].Regions {
				if region.Code == regionFilter {
					filteredRegions = append(filteredRegions, region)
				}
			}
			providers[i].Regions = filteredRegions
		}

		// Add backend details if requested
		if detailed {
			for j := range providers[i].Regions {
				backends, err := h.repo.GetProviderBackends(ctx, providers[i].Name, providers[i].Regions[j].Code)
				if err == nil {
					providers[i].Regions[j].Backends = backends
				}
			}
		}
	}

	response := ProvidersStatusResponse{
		Providers:   providers,
		GeneratedAt: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// --- Incident Endpoints ---

// HandleListIncidents lists incidents with filtering
func (h *Handler) HandleListIncidents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := ListIncidentsQuery{}

	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = status
	}

	if severity := r.URL.Query().Get("severity"); severity != "" {
		query.Severity = severity
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			query.StartDate = &t
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			query.EndDate = &t
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil && n > 0 {
			query.Limit = n
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if n, err := strconv.Atoi(offset); err == nil && n >= 0 {
			query.Offset = n
		}
	}

	response, err := h.repo.ListIncidents(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("Failed to list incidents")
		http.Error(w, "Failed to list incidents", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleGetIncident returns a single incident by ID
func (h *Handler) HandleGetIncident(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	incidentID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	incident, err := h.repo.GetIncidentByID(ctx, incidentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Incident not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get incident")
		http.Error(w, "Failed to get incident", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

// HandleCreateIncident creates a new incident (admin only)
func (h *Handler) HandleCreateIncident(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify admin access
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.isAdmin(user) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	var req CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Severity == "" || req.Description == "" {
		http.Error(w, "Title, severity, and description are required", http.StatusBadRequest)
		return
	}

	// Validate severity
	validSeverities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	if !validSeverities[req.Severity] {
		http.Error(w, "Invalid severity value", http.StatusBadRequest)
		return
	}

	incident, err := h.repo.CreateIncident(ctx, &req, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create incident")
		http.Error(w, "Failed to create incident", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, incident)
}

// HandleUpdateIncident updates an incident (admin only)
func (h *Handler) HandleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify admin access
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.isAdmin(user) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	incidentID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	var req UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate severity if provided
	if req.Severity != "" {
		validSeverities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		if !validSeverities[req.Severity] {
			http.Error(w, "Invalid severity value", http.StatusBadRequest)
			return
		}
	}

	// Validate status if provided
	if req.Status != "" {
		validStatuses := map[string]bool{"investigating": true, "identified": true, "monitoring": true, "resolved": true}
		if !validStatuses[req.Status] {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
	}

	incident, err := h.repo.UpdateIncident(ctx, incidentID, &req, user.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Incident not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to update incident")
		http.Error(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

// --- Metrics Endpoints ---

// HandleGetUptimeMetrics returns historical uptime data
func (h *Handler) HandleGetUptimeMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	component := r.URL.Query().Get("component")
	if component == "" {
		component = "all"
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "all"
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	resolution := r.URL.Query().Get("resolution")
	if resolution == "" {
		resolution = "hour"
	}

	// Calculate time range
	end := time.Now()
	var start time.Time
	var step time.Duration

	switch period {
	case "24h":
		start = end.Add(-24 * time.Hour)
		step = time.Hour
	case "7d":
		start = end.Add(-7 * 24 * time.Hour)
		step = 6 * time.Hour
	case "30d":
		start = end.Add(-30 * 24 * time.Hour)
		step = 24 * time.Hour
	case "90d":
		start = end.Add(-90 * 24 * time.Hour)
		step = 24 * time.Hour
	default:
		start = end.Add(-24 * time.Hour)
		step = time.Hour
	}

	if resolution == "day" {
		step = 24 * time.Hour
	}

	// Query Prometheus for uptime data
	promResp, err := h.prometheus.GetUptimeRange(ctx, component, start, end, step)
	if err != nil {
		if !errors.Is(err, ErrPrometheusNotConfigured) {
			logrus.WithError(err).Warn("Failed to get uptime from Prometheus, using fallback")
		}
		// Return empty response as fallback
		promResp = &PrometheusResponse{Status: "success"}
	}

	// Build response
	var dataPoints []UptimeDataPoint
	overallUptime := 100.0

	if promResp.Data != nil {
		for _, result := range promResp.Data.Result {
			if promResp.Data.ResultType == "matrix" {
				for _, values := range result.Values {
					if len(values) >= 2 {
						timestamp := parsePrometheusTimestamp(values[0])
						value := parseValue(values[1])

						dataPoints = append(dataPoints, UptimeDataPoint{
							Timestamp:     timestamp,
							UptimePercent: value,
							TotalChecks:   100, // Placeholder
							FailedChecks:  int((100 - value) * 100 / 100),
						})
					}
				}
			}
		}
	}

	// Calculate overall uptime
	if len(dataPoints) > 0 {
		total := 0.0
		for _, dp := range dataPoints {
			total += dp.UptimePercent
		}
		overallUptime = total / float64(len(dataPoints))
	}

	response := UptimeMetricsResponse{
		Period:        period,
		Resolution:    resolution,
		OverallUptime: overallUptime,
		DataPoints:    dataPoints,
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleGetLatencyMetrics returns latency trends
func (h *Handler) HandleGetLatencyMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "all"
	}

	region := r.URL.Query().Get("region")
	if region == "" {
		region = "all"
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	percentile := r.URL.Query().Get("percentile")
	if percentile == "" {
		percentile = "p95"
	}

	// Map percentile string to value
	percentileValue := 0.95
	switch percentile {
	case "p50":
		percentileValue = 0.50
	case "p95":
		percentileValue = 0.95
	case "p99":
		percentileValue = 0.99
	}

	// Calculate time range
	end := time.Now()
	var start time.Time
	var step time.Duration

	switch period {
	case "1h":
		start = end.Add(-1 * time.Hour)
		step = 5 * time.Minute
	case "24h":
		start = end.Add(-24 * time.Hour)
		step = time.Hour
	case "7d":
		start = end.Add(-7 * 24 * time.Hour)
		step = 6 * time.Hour
	case "30d":
		start = end.Add(-30 * 24 * time.Hour)
		step = 24 * time.Hour
	default:
		start = end.Add(-24 * time.Hour)
		step = time.Hour
	}

	// Query Prometheus
	promResp, err := h.prometheus.GetLatencyRange(ctx, provider, start, end, step, percentileValue)
	if err != nil {
		if !errors.Is(err, ErrPrometheusNotConfigured) {
			logrus.WithError(err).Warn("Failed to get latency from Prometheus, using fallback")
		}
		promResp = &PrometheusResponse{Status: "success"}
	}

	// Build response
	var dataPoints []LatencyDataPoint
	byProvider := make(map[string]LatencyStats)
	overallAvg := 0.0

	if promResp.Data != nil {
		for _, result := range promResp.Data.Result {
			prov := result.Metric["provider"]
			if prov == "" {
				prov = provider
			}

			if promResp.Data.ResultType == "matrix" {
				for _, values := range result.Values {
					if len(values) >= 2 {
						timestamp := parsePrometheusTimestamp(values[0])
						value := parseValue(values[1])

						dataPoints = append(dataPoints, LatencyDataPoint{
							Timestamp: timestamp,
							ValueMs:   value,
							Provider:  prov,
						})

						// Accumulate for provider stats
						stats := byProvider[prov]
						stats.AvgMs += value
						if stats.MinMs == 0 || value < stats.MinMs {
							stats.MinMs = value
						}
						if value > stats.MaxMs {
							stats.MaxMs = value
						}
						byProvider[prov] = stats
					}
				}
			}
		}
	}

	// Calculate averages
	if len(dataPoints) > 0 {
		total := 0.0
		for _, dp := range dataPoints {
			total += dp.ValueMs
		}
		overallAvg = total / float64(len(dataPoints))

		// Finalize provider stats
		for prov, stats := range byProvider {
			count := 0
			for _, dp := range dataPoints {
				if dp.Provider == prov {
					count++
				}
			}
			if count > 0 {
				stats.AvgMs = stats.AvgMs / float64(count)
				// Approximate percentiles from min/max
				stats.P50Ms = stats.MinMs + (stats.MaxMs-stats.MinMs)*0.5
				stats.P95Ms = stats.MinMs + (stats.MaxMs-stats.MinMs)*0.95
				stats.P99Ms = stats.MinMs + (stats.MaxMs-stats.MinMs)*0.99
				byProvider[prov] = stats
			}
		}
	}

	response := LatencyMetricsResponse{
		Period:       period,
		Percentile:   percentile,
		OverallAvgMs: overallAvg,
		DataPoints:   dataPoints,
		ByProvider:   byProvider,
	}

	respondJSON(w, http.StatusOK, response)
}

// --- Maintenance Endpoints ---

// HandleListMaintenance lists scheduled maintenance windows
func (h *Handler) HandleListMaintenance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := ListMaintenanceQuery{}

	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = status
	}

	if upcoming := r.URL.Query().Get("upcoming"); upcoming == "true" {
		query.Upcoming = true
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil && n > 0 {
			query.Limit = n
		}
	}

	response, err := h.repo.ListMaintenance(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("Failed to list maintenance")
		http.Error(w, "Failed to list maintenance", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleCreateMaintenance creates a new maintenance window (admin only)
func (h *Handler) HandleCreateMaintenance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify admin access
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.isAdmin(user) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	var req CreateMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.ScheduledStart.IsZero() || req.ScheduledEnd.IsZero() {
		http.Error(w, "Title, scheduled_start, and scheduled_end are required", http.StatusBadRequest)
		return
	}

	if req.ScheduledEnd.Before(req.ScheduledStart) {
		http.Error(w, "scheduled_end must be after scheduled_start", http.StatusBadRequest)
		return
	}

	maintenance, err := h.repo.CreateMaintenance(ctx, &req, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create maintenance")
		http.Error(w, "Failed to create maintenance", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, maintenance)
}

// --- Helper Methods ---

// determinePlatformStatus calculates platform status based on health and incidents
func (h *Handler) determinePlatformStatus(healthPercent float64, incidents []Incident) (status, indicator, description string) {
	// Check for critical incidents
	hasCritical := false
	hasMajor := false

	for _, incident := range incidents {
		if incident.Status != "resolved" {
			switch incident.Severity {
			case "critical":
				hasCritical = true
			case "high":
				hasMajor = true
			}
		}
	}

	// Determine status based on health and incidents
	switch {
	case hasCritical || healthPercent < 50:
		return "major_outage", "critical", "We are experiencing a major outage. Our team is investigating."
	case hasMajor || healthPercent < 80:
		return "degraded", "major", "We are experiencing degraded performance. Our team is working on it."
	case healthPercent < 95:
		return "degraded", "minor", "Some components may be experiencing issues."
	default:
		return "operational", "none", "All systems operational."
	}
}

// getComponentSummaries returns a summary of component statuses
func (h *Handler) getComponentSummaries(ctx context.Context) []Component {
	// Get service health from Prometheus
	serviceHealth, err := h.prometheus.GetServiceHealth(ctx)
	if err != nil {
		if !errors.Is(err, ErrPrometheusNotConfigured) {
			logrus.WithError(err).Warn("Failed to get service health from Prometheus")
		}
		serviceHealth = make(map[string]bool)
	}

	// When no Prometheus data for a job, treat as operational (avoid showing all Down when Prometheus is unreachable)
	statusOrDefault := func(key string) string {
		if v, ok := serviceHealth[key]; ok {
			return mapBoolStatus(v)
		}
		return "operational"
	}

	// Build component list with real uptime calculations
	components := []Component{
		{
			ID:        "api",
			Name:      "API",
			Type:      "api",
			Status:    statusOrDefault("orchestrator-api"),
			Uptime24h: 0, // Will be calculated from health history
		},
		{
			ID:        "database",
			Name:      "Database",
			Type:      "database",
			Status:    statusOrDefault("postgres"),
			Uptime24h: 0, // Will be calculated from health history
		},
		{
			ID:        "cache",
			Name:      "Cache",
			Type:      "cache",
			Status:    statusOrDefault("redis"),
			Uptime24h: 0, // Will be calculated from health history
		},
		{
			ID:        "health-monitor",
			Name:      "Health Monitor",
			Type:      "monitoring",
			Status:    statusOrDefault("health-monitor"),
			Uptime24h: 0, // Will be calculated from health history
		},
	}

	// Calculate real uptime for each component
	for i := range components {
		// Calculate uptime for different periods
		uptime24h, err := h.repo.CalculateComponentUptime(ctx, components[i].Name, 24*time.Hour)
		if err == nil {
			components[i].Uptime24h = uptime24h
		}

		uptime7d, err := h.repo.CalculateComponentUptime(ctx, components[i].Name, 7*24*time.Hour)
		if err == nil {
			components[i].Uptime7d = uptime7d
		} else {
			components[i].Uptime7d = components[i].Uptime24h
		}

		uptime30d, err := h.repo.CalculateComponentUptime(ctx, components[i].Name, 30*24*time.Hour)
		if err == nil {
			components[i].Uptime30d = uptime30d
		} else {
			components[i].Uptime30d = components[i].Uptime7d
		}
	}

	return components
}

// mapBoolStatus converts a boolean health status to component status
func mapBoolStatus(healthy bool) string {
	if healthy {
		return "operational"
	}
	return "major_outage"
}

// getUptimeForComponent gets uptime percentage for a component from Prometheus
func (h *Handler) getUptimeForComponent(ctx context.Context, component, duration string) (float64, error) {
	resp, err := h.prometheus.GetUptimeRatio(ctx, component, "", duration)
	if err != nil {
		return 0, err
	}

	if resp.Data != nil && len(resp.Data.Result) > 0 {
		if len(resp.Data.Result[0].Value) >= 2 {
			return parseValue(resp.Data.Result[0].Value[1]), nil
		}
	}

	return 99.9, nil // Default fallback
}

// isAdmin checks if the user has admin role
func (h *Handler) isAdmin(claims *auth.Claims) bool {
	return claims.Role == "admin" || claims.Role == "super_admin"
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// parsePrometheusTimestamp parses a Prometheus timestamp
func parsePrometheusTimestamp(ts interface{}) time.Time {
	switch v := ts.(type) {
	case float64:
		return time.Unix(int64(v), 0)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Unix(int64(f), 0)
		}
	}
	return time.Now()
}
