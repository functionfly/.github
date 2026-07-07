package status

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
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

	// Full product catalog + Prometheus/DB-derived metrics; overlay latest rows from system_health_checks
	// so new catalog entries always appear even when only a subset is actively probed.
	components := h.getComponentSummaries(ctx)
	if dbChecks, err := h.repo.GetSystemHealthChecks(ctx); err != nil {
		logrus.WithError(err).Warn("Failed to load system_health_checks; using catalog summaries without DB overlay")
	} else {
		components = h.overlayDBHealthChecks(components, dbChecks)
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
		apierror.WriteError(w, apierror.NewInternal("Failed to get provider status"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to list incidents"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	incident, err := h.repo.GetIncidentByID(ctx, incidentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Incident not found"))
			return
		}
		logrus.WithError(err).Error("Failed to get incident")
		apierror.WriteError(w, apierror.NewInternal("Failed to get incident"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.isAdmin(user) {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	var req CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate required fields
	if req.Title == "" || req.Severity == "" || req.Description == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Title, severity, and description are required"))
		return
	}

	// Validate severity
	validSeverities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	if !validSeverities[req.Severity] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid severity value"))
		return
	}

	incident, err := h.repo.CreateIncident(ctx, &req, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create incident")
		apierror.WriteError(w, apierror.NewInternal("Failed to create incident"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.isAdmin(user) {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	vars := mux.Vars(r)
	incidentID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	var req UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate severity if provided
	if req.Severity != "" {
		validSeverities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		if !validSeverities[req.Severity] {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid severity value"))
			return
		}
	}

	// Validate status if provided
	if req.Status != "" {
		validStatuses := map[string]bool{"investigating": true, "identified": true, "monitoring": true, "resolved": true}
		if !validStatuses[req.Status] {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid status value"))
			return
		}
	}

	incident, err := h.repo.UpdateIncident(ctx, incidentID, &req, user.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Incident not found"))
			return
		}
		logrus.WithError(err).Error("Failed to update incident")
		apierror.WriteError(w, apierror.NewInternal("Failed to update incident"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to list maintenance"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.isAdmin(user) {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	var req CreateMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate required fields
	if req.Title == "" || req.ScheduledStart.IsZero() || req.ScheduledEnd.IsZero() {
		apierror.WriteError(w, apierror.NewBadRequest("Title, scheduled_start, and scheduled_end are required"))
		return
	}

	if req.ScheduledEnd.Before(req.ScheduledStart) {
		apierror.WriteError(w, apierror.NewBadRequest("scheduled_end must be after scheduled_start"))
		return
	}

	maintenance, err := h.repo.CreateMaintenance(ctx, &req, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create maintenance")
		apierror.WriteError(w, apierror.NewInternal("Failed to create maintenance"))
		return
	}

	respondJSON(w, http.StatusCreated, maintenance)
}

// HandleGetRSSFeed returns an RSS feed of recent incidents and maintenance
func (h *Handler) HandleGetRSSFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get recent incidents (last 30 days)
	incidentsQuery := ListIncidentsQuery{
		Limit:  50,
		Offset: 0,
	}
	incidentsResp, err := h.repo.ListIncidents(ctx, incidentsQuery)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get incidents for RSS feed")
		incidentsResp = &IncidentsListResponse{Incidents: []Incident{}}
	}

	// Get recent maintenance windows
	maintenanceQuery := ListMaintenanceQuery{
		Upcoming: false,
		Limit:    20,
	}
	maintenanceResp, err := h.repo.ListMaintenance(ctx, maintenanceQuery)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get maintenance for RSS feed")
		maintenanceResp = &MaintenanceListResponse{MaintenanceWindows: []MaintenanceWindow{}}
	}

	// Build RSS feed
	rss := h.generateRSSFeed(incidentsResp.Incidents, maintenanceResp.MaintenanceWindows)

	// Set headers
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rss))
}

// generateRSSFeed generates an RSS 2.0 feed from incidents and maintenance
func (h *Handler) generateRSSFeed(incidents []Incident, maintenance []MaintenanceWindow) string {
	now := time.Now().UTC()
	baseURL := "https://status.functionfly.com"

	// Start RSS document
	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
  <title>FunctionFly Status - Incident Feed</title>
  <link>` + baseURL + `</link>
  <description>Real-time updates on FunctionFly platform status, incidents, and maintenance</description>
  <language>en</language>
  <lastBuildDate>` + now.Format(time.RFC1123) + `</lastBuildDate>
  <atom:link href="` + baseURL + `/api/v1/status/rss" rel="self" type="application/rss+xml" />
  <generator>FunctionFly Status</generator>
`

	// Add incidents as items
	for _, incident := range incidents {
		item := h.incidentToRSSItem(incident, baseURL)
		rss += item
	}

	// Add completed maintenance as items
	for _, m := range maintenance {
		if m.Status == "completed" || m.Status == "in_progress" {
			item := h.maintenanceToRSSItem(m, baseURL)
			rss += item
		}
	}

	// Close RSS document
	rss += `</channel>
</rss>`

	return rss
}

// incidentToRSSItem converts an incident to an RSS item
func (h *Handler) incidentToRSSItem(incident Incident, baseURL string) string {
	status := incident.Status
	severity := incident.Severity
	title := fmt.Sprintf("[%s] %s - %s", strings.ToUpper(severity), incident.Title, status)

	var description strings.Builder
	description.WriteString(fmt.Sprintf("<p><strong>Severity:</strong> %s</p>", severity))
	description.WriteString(fmt.Sprintf("<p><strong>Status:</strong> %s</p>", status))
	description.WriteString(fmt.Sprintf("<p>%s</p>", html.EscapeString(incident.Description)))

	if len(incident.AffectedComponents) > 0 {
		description.WriteString(fmt.Sprintf("<p><strong>Affected:</strong> %s</p>", strings.Join(incident.AffectedComponents, ", ")))
	}

	if len(incident.Updates) > 0 {
		description.WriteString("<h3>Updates:</h3><ul>")
		for _, update := range incident.Updates {
			description.WriteString(fmt.Sprintf("<li><strong>%s</strong> (%s): %s</li>",
				update.Status,
				update.CreatedAt.Format(time.RFC1123),
				html.EscapeString(update.Message)))
		}
		description.WriteString("</ul>")
	}

	pubDate := incident.CreatedAt.Format(time.RFC1123)
	guid := fmt.Sprintf("%s/incidents/%s", baseURL, incident.ID)
	link := guid

	return fmt.Sprintf(`
  <item>
    <title>%s</title>
    <link>%s</link>
    <guid isPermaLink="true">%s</guid>
    <pubDate>%s</pubDate>
    <description><![CDATA[%s]]></description>
    <category>%s</category>
  </item>
`, xmlEscape(title), link, guid, pubDate, description.String(), status)
}

// maintenanceToRSSItem converts a maintenance window to an RSS item
func (h *Handler) maintenanceToRSSItem(m MaintenanceWindow, baseURL string) string {
	status := m.Status
	title := fmt.Sprintf("[Maintenance] %s - %s", m.Title, status)

	var description strings.Builder
	description.WriteString(fmt.Sprintf("<p><strong>Status:</strong> %s</p>", status))
	description.WriteString(fmt.Sprintf("<p>%s</p>", html.EscapeString(m.Description)))
	description.WriteString(fmt.Sprintf("<p><strong>Scheduled:</strong> %s to %s</p>",
		m.ScheduledStart.Format(time.RFC1123),
		m.ScheduledEnd.Format(time.RFC1123)))

	if len(m.AffectedComponents) > 0 {
		description.WriteString(fmt.Sprintf("<p><strong>Affected:</strong> %s</p>", strings.Join(m.AffectedComponents, ", ")))
	}

	pubDate := m.CreatedAt.Format(time.RFC1123)
	guid := fmt.Sprintf("%s/maintenance/%s", baseURL, m.ID)
	link := guid

	return fmt.Sprintf(`
  <item>
    <title>%s</title>
    <link>%s</link>
    <guid isPermaLink="true">%s</guid>
    <pubDate>%s</pubDate>
    <description><![CDATA[%s]]></description>
    <category>maintenance</category>
  </item>
`, xmlEscape(title), link, guid, pubDate, description.String())
}

// xmlEscape escapes XML special characters
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
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

// overlayDBHealthChecks merges the latest DB health row per component name onto the catalog.
// DB response times of 0 are ignored so Prometheus/type defaults from getComponentSummaries stay visible.
func (h *Handler) overlayDBHealthChecks(catalog []Component, db []Component) []Component {
	if len(db) == 0 {
		return catalog
	}
	byName := make(map[string]Component, len(db))
	for _, row := range db {
		byName[row.Name] = row
	}
	catalogNames := make(map[string]struct{}, len(catalog))
	for i := range catalog {
		catalogNames[catalog[i].Name] = struct{}{}
		if row, ok := byName[catalog[i].Name]; ok {
			catalog[i].Status = row.Status
			if row.ResponseTime > 0 {
				catalog[i].ResponseTime = row.ResponseTime
			}
			if !row.LastChecked.IsZero() {
				catalog[i].LastChecked = row.LastChecked
			}
		}
	}
	out := catalog
	for _, row := range db {
		if _, ok := catalogNames[row.Name]; ok {
			continue
		}
		out = append(out, row)
	}
	return out
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
	// Components are ordered by criticality (most critical first)
	components := []Component{
		{
			ID:          "api",
			Name:        "API",
			Type:        "api",
			Status:      statusOrDefault("orchestrator-api"),
			Description: "Main API serving all requests",
			Uptime24h:   0, // Will be calculated from health history
		},
		{
			ID:          "database",
			Name:        "Database",
			Type:        "database",
			Status:      func() string { healthy, _ := checkDedicatedDBHealth(ctx); return mapBoolStatus(healthy) }(),
			Description: "Primary PostgreSQL database",
			Uptime24h:   0, // Will be calculated from health history
		},
		{
			ID:          "cache",
			Name:        "Cache",
			Type:        "cache",
			Status:      statusOrDefault("redis"),
			Description: "Redis cache layer",
			Uptime24h:   0, // Will be calculated from health history
		},
		{
			ID:          "ai-service",
			Name:        "AI Service",
			Type:        "ai",
			Status:      statusOrDefault("ai-service"),
			Description: "FlyMind AI assistant and code generation",
			Uptime24h:   0,
		},
		{
			ID:          "embeddings",
			Name:        "Embeddings",
			Type:        "ai",
			Status:      statusOrDefault("fly-embed"),
			Description: "Vector search and function embeddings",
			Uptime24h:   0,
		},
		{
			ID:          "state-fabric",
			Name:        "State Fabric",
			Type:        "storage",
			Status:      statusOrDefault("state-fabric"),
			Description: "Distributed state management and triggers",
			Uptime24h:   0,
		},
		{
			ID:          "microvm",
			Name:        "MicroVM Runtime",
			Type:        "runtime",
			Status:      statusOrDefault("microvm"),
			Description: "Secure function execution environment",
			Uptime24h:   0,
		},
		{
			ID:          "queue",
			Name:        "Queue Worker",
			Type:        "worker",
			Status:      statusOrDefault("queue"),
			Description: "Background job processing",
			Uptime24h:   0,
		},
		{
			ID:          "function-backup",
			Name:        "Function Backup",
			Type:        "backup",
			Status:      statusOrDefault("function-backup"),
			Description: "Automated function backup to R2",
			Uptime24h:   0,
		},
		{
			ID:          "email",
			Name:        "Email Delivery",
			Type:        "email",
			Status:      statusOrDefault("email"),
			Description: "Transactional and notification emails",
			Uptime24h:   0,
		},
		{
			ID:          "billing",
			Name:        "Billing",
			Type:        "billing",
			Status:      statusOrDefault("billing"),
			Description: "Stripe payment processing and invoicing",
			Uptime24h:   0,
		},
		{
			ID:          "storage",
			Name:        "Object Storage",
			Type:        "storage",
			Status:      statusOrDefault("storage"),
			Description: "R2 object storage for artifacts and backups",
			Uptime24h:   0,
		},
		{
			ID:          "cdn",
			Name:        "CDN",
			Type:        "cdn",
			Status:      func() string { healthy, _, _ := checkCloudflareHealth(ctx); return mapBoolStatus(healthy) }(),
			Description: "Cloudflare edge caching and delivery",
			Uptime24h:   0,
		},
		{
			ID:          "pgbouncer",
			Name:        "Connection Pool",
			Type:        "infrastructure",
			Status:      statusOrDefault("pgbouncer"),
			Description: "PostgreSQL connection pooling",
			Uptime24h:   0,
		},
		{
			ID:          "recommendations",
			Name:        "Recommendations",
			Type:        "ai",
			Status:      statusOrDefault("recommendations"),
			Description: "Function recommendation engine",
			Uptime24h:   0,
		},
		{
			ID:          "verification",
			Name:        "Verification Pipeline",
			Type:        "security",
			Status:      statusOrDefault("verification"),
			Description: "Function verification and security scanning",
			Uptime24h:   0,
		},
		{
			ID:          "trust-api",
			Name:        "Trust API",
			Type:        "security",
			Status:      statusOrDefault("trust-api"),
			Description: "Trust scoring and reputation system",
			Uptime24h:   0,
		},
		{
			ID:          "support",
			Name:        "Support System",
			Type:        "service",
			Status:      statusOrDefault("support"),
			Description: "Customer support and ticket management",
			Uptime24h:   0,
		},
		{
			ID:          "registry",
			Name:        "Function Registry",
			Type:        "service",
			Status:      statusOrDefault("registry"),
			Description: "Function discovery and metadata",
			Uptime24h:   0,
		},
		{
			ID:          "health-monitor",
			Name:        "Health Monitor",
			Type:        "monitoring",
			Status:      statusOrDefault("health-monitor"),
			Description: "Internal health monitoring service",
			Uptime24h:   0, // Will be calculated from health history
		},
	}

	// Calculate real uptime and response time for each component
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

		// Get response time - first try database health checks
		dbLatency, err := h.repo.GetLatestComponentResponseTime(ctx, components[i].Name)
		if err == nil && dbLatency > 0 {
			components[i].ResponseTime = dbLatency
		} else {
			// Fallback to Prometheus HTTP metrics
			promLatency, err := h.prometheus.GetComponentHTTPLatency(ctx, components[i].Name, "5m")
			if err == nil && promLatency > 0 {
				components[i].ResponseTime = int(promLatency)
			} else {
				// Use reasonable defaults based on component type
				components[i].ResponseTime = getDefaultResponseTime(components[i].Type)
			}
		}
	}

	return components
}

// getDefaultResponseTime returns a reasonable default response time for a component type
func getDefaultResponseTime(componentType string) int {
	switch componentType {
	case "api":
		return 45
	case "database":
		return 12
	case "cache":
		return 5
	case "ai":
		return 250 // AI operations are typically slower
	case "email":
		return 150
	case "billing":
		return 80
	case "storage":
		return 60
	case "cdn":
		return 25
	case "monitoring":
		return 30
	case "runtime":
		return 100 // Function execution runtime
	case "worker":
		return 200 // Background workers
	case "backup":
		return 500 // Backup operations are slower
	case "infrastructure":
		return 20 // Infrastructure components like connection pools
	case "security":
		return 120 // Security scanning operations
	case "service":
		return 75 // General services
	default:
		return 50
	}
}

// mapBoolStatus converts a boolean health status to component status
func mapBoolStatus(healthy bool) string {
	if healthy {
		return "operational"
	}
	return "major_outage"
}

// checkDedicatedDBHealth checks the health of the local/primary database server
func checkDedicatedDBHealth(ctx context.Context) (bool, int) {
	connString := storage.GetConnectionString()
	if connString == "" {
		logrus.Warn("checkDedicatedDBHealth: connection string empty")
		return true, 0 // Return healthy if not configured (fallback)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logrus.WithError(err).Warn("checkDedicatedDBHealth: failed to parse connection string")
		return false, 0
	}

	poolConfig.MaxConns = 1
	poolConfig.MinConns = 0

	pool, err := pgxpool.NewWithConfig(dbCtx, poolConfig)
	if err != nil {
		logrus.WithError(err).Warn("checkDedicatedDBHealth: failed to create pool")
		return false, 0
	}
	defer pool.Close()

	start := time.Now()
	err = pool.Ping(dbCtx)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		logrus.WithError(err).Warn("checkDedicatedDBHealth: ping failed")
		return false, latencyMs
	}

	return true, latencyMs
}

// checkCloudflareHealth checks CDN health via Cloudflare API
func checkCloudflareHealth(ctx context.Context) (bool, int, string) {
	apiToken := os.Getenv("CF_API_TOKEN")
	zoneID := os.Getenv("CF_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		return true, 0, "cloudflare not configured"
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := fmt.Sprintf(`{
  viewer {
    zones(filter: { zoneTag: "%s" }) {
      zoneStatus
      healthCheck {
        status
      }
    }
  }
}`, zoneID)

	body := map[string]interface{}{"query": query}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(reqCtx, "POST", "https://api.cloudflare.com/graphql", bytes.NewReader(bodyBytes))
	if err != nil {
		return false, 0, fmt.Sprintf("request creation failed: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return false, latencyMs, fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, latencyMs, fmt.Sprintf("cloudflare returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Viewer struct {
				Zones []struct {
					ZoneStatus   string `json:"zoneStatus"`
					HealthCheck  struct {
						Status string `json:"status"`
					} `json:"healthCheck"`
				} `json:"zones"`
			} `json:"viewer"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, latencyMs, fmt.Sprintf("failed to parse response: %v", err)
	}

	if len(result.Data.Viewer.Zones) == 0 {
		return false, latencyMs, "zone not found"
	}

	zone := result.Data.Viewer.Zones[0]
	if zone.ZoneStatus != "active" {
		return false, latencyMs, fmt.Sprintf("zone status: %s", zone.ZoneStatus)
	}

	return true, latencyMs, "operational"
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
