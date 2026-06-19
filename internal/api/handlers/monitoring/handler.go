package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// Handler handles monitoring API requests
type Handler struct {
	repo            storage.Repository
	monitoringSvc   *monitoring.Service
	realtimeMonitor *monitoring.RealtimeMonitor
	authSvc         *auth.AuthService
	upgrader        websocket.Upgrader
}

// NewHandler creates a new monitoring handler
func NewHandler(repo storage.Repository, monitoringSvc *monitoring.Service, realtimeMonitor *monitoring.RealtimeMonitor, authSvc *auth.AuthService) *Handler {
	return &Handler{
		repo:            repo,
		monitoringSvc:   monitoringSvc,
		realtimeMonitor: realtimeMonitor,
		authSvc:         authSvc,
		upgrader: websocket.Upgrader{
			CheckOrigin: middleware.IsOriginAllowedForRequest,
		},
	}
}

// HandleGetMetrics returns performance metrics
func (h *Handler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	metricType := r.URL.Query().Get("type")
	tenantIDStr := r.URL.Query().Get("tenant_id")
	sinceStr := r.URL.Query().Get("since")

	var tenantID *uuid.UUID
	if tenantIDStr != "" {
		if id, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &id
		}
	}

	since := time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	if sinceStr != "" {
		if parsed, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsed
		}
	}

	metrics, err := h.monitoringSvc.GetMetrics(r.Context(), metricType, tenantID, since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to get metrics"))
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

// HandleGetAlerts returns active alerts
func (h *Handler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID *uuid.UUID
	if tenantIDStr != "" {
		if id, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &id
		}
	}

	alerts, err := h.monitoringSvc.GetActiveAlerts(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get alerts")
		apierror.WriteError(w, apierror.NewInternal("Failed to get alerts"))
		return
	}

	json.NewEncoder(w).Encode(alerts)
}

// HandleGetSystemHealth returns system health status
func (h *Handler) HandleGetSystemHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health, err := h.monitoringSvc.GetSystemHealthStatus(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get system health")
		apierror.WriteError(w, apierror.NewInternal("Failed to get system health"))
		return
	}

	json.NewEncoder(w).Encode(health)
}

// HandleResolveAlert resolves an alert
func (h *Handler) HandleResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertIDStr := vars["alertId"]

	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	// Get user ID from authentication context
	userClaims := middleware.GetUserFromContext(r)
	if userClaims == nil {
		logrus.Warn("No authenticated user found in context for alert resolution")
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	resolvedBy := userClaims.UserID

	if err := h.monitoringSvc.ResolveAlert(r.Context(), alertID, resolvedBy); err != nil {
		logrus.WithError(err).WithField("alert_id", alertID).Error("Failed to resolve alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to resolve alert"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

// HandleRealtimeConnection establishes a WebSocket connection for real-time monitoring
func (h *Handler) HandleRealtimeConnection(w http.ResponseWriter, r *http.Request) {
	// Extract user and tenant ID from authentication context
	userClaims := middleware.GetUserFromContext(r)

	// If no user in context, try to authenticate via token query parameter (for WebSocket connections)
	if userClaims == nil {
		token := r.URL.Query().Get("token")
		logrus.WithFields(logrus.Fields{
			"token_length": len(token),
			"has_token":    token != "",
		}).Info("WebSocket authentication attempt")
		if token != "" {
			claims, err := h.authSvc.ValidateToken(r.Context(), token)
			if err != nil {
				tokenPrefix := token
				if len(token) > 50 {
					tokenPrefix = token[:50] + "..."
				}
				logrus.WithError(err).WithField("token_prefix", tokenPrefix).Warn("WebSocket authentication failed via token")
				// For WebSocket, send a proper close frame with policy violation code
				if r.Header.Get("Upgrade") == "websocket" {
					conn, upgradeErr := h.upgrader.Upgrade(w, r, nil)
					if upgradeErr == nil {
						conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "Authentication required"), time.Now().Add(time.Second))
						conn.Close()
					}
				} else {
					apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
				}
				return
			}
			userClaims = claims
			logrus.WithField("user_id", userClaims.UserID).Info("WebSocket authentication successful")
		} else {
			logrus.Warn("No authenticated user found in context and no token provided for realtime connection")
			// For WebSocket, send a proper close frame
			if r.Header.Get("Upgrade") == "websocket" {
				conn, upgradeErr := h.upgrader.Upgrade(w, r, nil)
				if upgradeErr == nil {
					conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "Authentication required"), time.Now().Add(time.Second))
					conn.Close()
				}
			} else {
				apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
			}
			return
		}
	}

	userID := &userClaims.UserID
	tenantID := &userClaims.TenantID

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade connection")
		return
	}

	h.realtimeMonitor.HandleRealtimeConnection(r.Context(), conn, userID, tenantID)
}

// HandleGetRealtimeStats returns real-time connection statistics
func (h *Handler) HandleGetRealtimeStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := h.realtimeMonitor.GetConnectionStats()
	json.NewEncoder(w).Encode(stats)
}

// HandleCreateAlert creates a new alert (for testing/admin purposes)
func (h *Handler) HandleCreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert storage.Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if err := h.monitoringSvc.RecordAlert(r.Context(), &alert); err != nil {
		logrus.WithError(err).Error("Failed to create alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to create alert"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// HandleRecordMetric records a performance metric
func (h *Handler) HandleRecordMetric(w http.ResponseWriter, r *http.Request) {
	var metric storage.PerformanceMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if err := h.monitoringSvc.RecordPerformanceMetric(r.Context(), &metric); err != nil {
		logrus.WithError(err).Error("Failed to record metric")
		apierror.WriteError(w, apierror.NewInternal("Failed to record metric"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(metric)
}

// HandleGetDashboardConfig returns dashboard configuration
func (h *Handler) HandleGetDashboardConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user from authentication context
	userClaims := middleware.GetUserFromContext(r)
	if userClaims == nil {
		logrus.Warn("No authenticated user found in context for dashboard config")
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	// Try to get user-specific configs first
	userConfigs, err := h.monitoringSvc.GetDashboardConfigsByUser(userClaims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userClaims.UserID).Error("Failed to get user dashboard configs")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve dashboard configuration"))
		return
	}

	// If no user-specific configs, get tenant-wide configs
	var configs []*storage.DashboardConfig
	if len(userConfigs) == 0 {
		tenantConfigs, err := h.monitoringSvc.GetDashboardConfigsByTenant(userClaims.TenantID)
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", userClaims.TenantID).Error("Failed to get tenant dashboard configs")
			apierror.WriteError(w, apierror.NewInternal("Failed to retrieve dashboard configuration"))
			return
		}
		configs = tenantConfigs
	} else {
		configs = userConfigs
	}

	// If no configs found, return default configuration
	if len(configs) == 0 {
		defaultConfig := map[string]interface{}{
			"panels": []map[string]interface{}{
				{
					"type":        "metric_chart",
					"title":       "Response Time",
					"metric_type": "response_time",
					"time_range":  "1h",
				},
				{
					"type":            "alert_list",
					"title":           "Active Alerts",
					"severity_filter": []string{"warning", "error", "critical"},
				},
			},
		}
		json.NewEncoder(w).Encode(defaultConfig)
		return
	}

	// Group configs by type and build response
	response := map[string]interface{}{
		"panels": []map[string]interface{}{},
	}

	for _, config := range configs {
		if config.ConfigType == "metric_panel" {
			response["panels"] = append(response["panels"].([]map[string]interface{}), config.Config)
		}
	}

	json.NewEncoder(w).Encode(response)
}

// HandleRecordHealthCheck records a system health check
func (h *Handler) HandleRecordHealthCheck(w http.ResponseWriter, r *http.Request) {
	var check storage.SystemHealthCheck
	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if err := h.monitoringSvc.RecordSystemHealthCheck(r.Context(), &check); err != nil {
		logrus.WithError(err).Error("Failed to record health check")
		apierror.WriteError(w, apierror.NewInternal("Failed to record health check"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(check)
}

// HandleGetMonitoringEvents returns recent monitoring events
func (h *Handler) HandleGetMonitoringEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims for tenant filtering
	userClaims := middleware.GetUserFromContext(r)
	if userClaims == nil {
		logrus.Warn("No authenticated user found in context for monitoring events")
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	eventType := r.URL.Query().Get("type")

	// Parse time range parameter (default to 1 hour)
	timeRangeStr := r.URL.Query().Get("range")
	timeRange := time.Hour // default
	if timeRangeStr != "" {
		switch timeRangeStr {
		case "1h":
			timeRange = time.Hour
		case "6h":
			timeRange = 6 * time.Hour
		case "24h":
			timeRange = 24 * time.Hour
		case "7d":
			timeRange = 7 * 24 * time.Hour
		}
	}

	since := time.Now().Add(-timeRange)

	// Query monitoring events from database
	events, err := h.repo.QueryMonitoringEvents(r.Context(), eventType, &userClaims.TenantID, since, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to query monitoring events")
		apierror.WriteError(w, apierror.NewInternal("Failed to query monitoring events"))
		return
	}

	json.NewEncoder(w).Encode(events)
}

// HandleGetDatabaseHealth returns database health metrics
func (h *Handler) HandleGetDatabaseHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health, err := h.monitoringSvc.GetDatabaseHealth(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get database health")
		apierror.WriteError(w, apierror.NewInternal("Failed to get database health"))
		return
	}

	json.NewEncoder(w).Encode(health)
}

// HandleGetDatabaseMetrics returns database performance metrics
func (h *Handler) HandleGetDatabaseMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse time range parameter (default to 1 hour)
	timeRangeStr := r.URL.Query().Get("range")
	timeRange := time.Hour // default
	if timeRangeStr != "" {
		switch timeRangeStr {
		case "1h":
			timeRange = time.Hour
		case "6h":
			timeRange = 6 * time.Hour
		case "24h":
			timeRange = 24 * time.Hour
		case "7d":
			timeRange = 7 * 24 * time.Hour
		}
	}

	metrics, err := h.monitoringSvc.GetDatabaseMetrics(r.Context(), timeRange)
	if err != nil {
		logrus.WithError(err).Error("Failed to get database metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to get database metrics"))
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

// HandleGetDatabaseAlerts returns database-specific alerts
func (h *Handler) HandleGetDatabaseAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts, err := h.monitoringSvc.GetDatabaseAlerts(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get database alerts")
		apierror.WriteError(w, apierror.NewInternal("Failed to get database alerts"))
		return
	}

	json.NewEncoder(w).Encode(alerts)
}

// HandleCheckDatabaseHealth performs comprehensive database health checks
func (h *Handler) HandleCheckDatabaseHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Run all database health checks
	checks := make(map[string]interface{})

	// Check connection pool
	if err := h.monitoringSvc.CheckDatabaseConnectionPool(r.Context()); err != nil {
		checks["connection_pool"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		checks["connection_pool"] = map[string]interface{}{
			"status": "ok",
		}
	}

	// Check performance
	if err := h.monitoringSvc.CheckDatabasePerformance(r.Context()); err != nil {
		checks["performance"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		checks["performance"] = map[string]interface{}{
			"status": "ok",
		}
	}

	// Check storage
	if err := h.monitoringSvc.CheckDatabaseStorage(r.Context()); err != nil {
		checks["storage"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		checks["storage"] = map[string]interface{}{
			"status": "ok",
		}
	}

	// Check replication
	if err := h.monitoringSvc.CheckDatabaseReplication(r.Context()); err != nil {
		checks["replication"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		checks["replication"] = map[string]interface{}{
			"status": "ok",
		}
	}

	result := map[string]interface{}{
		"timestamp": time.Now(),
		"checks":    checks,
	}

	// Determine overall status
	hasErrors := false
	for _, check := range checks {
		if checkMap, ok := check.(map[string]interface{}); ok {
			if checkMap["status"] == "error" {
				hasErrors = true
				break
			}
		}
	}

	if hasErrors {
		result["overall_status"] = "degraded"
	} else {
		result["overall_status"] = "healthy"
	}

	json.NewEncoder(w).Encode(result)
}

// HandleSubscribeToDatabaseChanges allows clients to subscribe to database change notifications
// This endpoint provides information about available database change channels
func (h *Handler) HandleSubscribeToDatabaseChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user claims for tenant filtering
	userClaims := middleware.GetUserFromContext(r)
	if userClaims == nil {
		logrus.Warn("No authenticated user found in context for database change subscription")
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	// Parse query parameters
	table := r.URL.Query().Get("table")
	if table == "" {
		// Return list of available tables for subscription
		response := map[string]interface{}{
			"available_tables": []string{
				"users", "apps", "backends", "alerts",
				"performance_metrics", "monitoring_events", "user_notifications",
			},
			"websocket_channels": map[string]interface{}{
				"description":     "Use WebSocket connection to /monitoring/realtime and subscribe to 'db_changes_{table}' channels",
				"tenant_specific": "For tenant-scoped changes, use 'db_changes_{table}_{tenant_id}' channels",
			},
			"current_tenant_id": userClaims.TenantID.String(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate table name
	validTables := map[string]bool{
		"users": true, "apps": true, "backends": true, "alerts": true,
		"performance_metrics": true, "monitoring_events": true, "user_notifications": true,
	}

	if !validTables[table] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid table name"))
		return
	}

	// Return subscription information for the specific table
	response := map[string]interface{}{
		"table": table,
		"websocket_channels": map[string]interface{}{
			"global": fmt.Sprintf("db_changes_%s", table),
			"tenant": fmt.Sprintf("db_changes_%s_%s", table, userClaims.TenantID.String()),
		},
		"event_types": []string{"INSERT", "UPDATE", "DELETE"},
		"payload_format": map[string]interface{}{
			"type":  "broadcast",
			"event": "db_change",
			"payload": map[string]interface{}{
				"schema":           "public",
				"table":            table,
				"eventType":        "INSERT|UPDATE|DELETE",
				"commit_timestamp": "timestamp",
				"new":              "new row data (null for DELETE)",
				"old":              "old row data (null for INSERT)",
				"ids":              []string{"uuids"},
				"errors":           "error message or null",
			},
		},
		"note": "Database changes are automatically broadcast via WebSocket. Use the useDatabaseChanges hook in the frontend.",
	}

	json.NewEncoder(w).Encode(response)
}

// HandleGetLocalRuntimeMetrics returns local runtime performance metrics
func (h *Handler) HandleGetLocalRuntimeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	timeRangeStr := r.URL.Query().Get("range")
	timeRange := time.Hour // default to 1 hour
	if timeRangeStr != "" {
		switch timeRangeStr {
		case "1h":
			timeRange = time.Hour
		case "6h":
			timeRange = 6 * time.Hour
		case "24h":
			timeRange = 24 * time.Hour
		case "7d":
			timeRange = 7 * 24 * time.Hour
		}
	}

	since := time.Now().Add(-timeRange)

	// Get aggregated metrics from database
	aggregatedMetrics, err := h.repo.GetAggregatedLocalRuntimeMetrics(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get aggregated local runtime metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to get local runtime metrics"))
		return
	}

	// Get list of active runtimes
	activeRuntimes, err := h.repo.ListActiveLocalRuntimes(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("Failed to get active runtimes list")
		// Continue without runtime list
	}

	// Build response
	response := map[string]interface{}{
		"aggregated_metrics": aggregatedMetrics,
		"active_runtimes":    len(activeRuntimes),
		"time_range":         timeRangeStr,
		"since":              since.Format(time.RFC3339),
		"timestamp":          time.Now().Format(time.RFC3339),
		"available_metrics": []string{
			"functionfly_local_runtime_executions_total",
			"functionfly_local_runtime_execution_duration_seconds",
			"functionfly_local_runtime_memory_usage_bytes",
			"functionfly_local_runtime_cpu_usage_percent",
			"functionfly_local_runtime_active_connections",
			"functionfly_local_runtime_request_throughput_per_second",
			"functionfly_local_runtime_errors_total",
		},
	}

	// Add runtime details if available
	if len(activeRuntimes) > 0 {
		runtimeDetails := make([]map[string]interface{}, 0, len(activeRuntimes))
		for _, runtime := range activeRuntimes {
			// Get latest metrics for this runtime
			latestMetrics, err := h.repo.GetLatestLocalRuntimeMetrics(r.Context(), runtime.ID)
			health := "unknown"
			if err == nil && latestMetrics != nil {
				// Simple health determination based on error rate
				if latestMetrics.ErrorRate < 5.0 {
					health = "healthy"
				} else if latestMetrics.ErrorRate < 15.0 {
					health = "degraded"
				} else {
					health = "unhealthy"
				}
			}

			runtimeDetail := map[string]interface{}{
				"id":             runtime.ID.String(),
				"runtime_id":     runtime.RuntimeID,
				"runtime_type":   runtime.RuntimeType,
				"function_name":  runtime.FunctionName,
				"host":           runtime.Host,
				"port":           runtime.Port,
				"status":         runtime.Status,
				"uptime":         runtime.Uptime,
				"last_heartbeat": runtime.LastHeartbeat.Format(time.RFC3339),
				"health":         health,
			}
			runtimeDetails = append(runtimeDetails, runtimeDetail)
		}
		response["runtime_details"] = runtimeDetails
	}

	json.NewEncoder(w).Encode(response)
}
