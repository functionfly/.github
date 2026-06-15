package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/dna"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/sirupsen/logrus"
)

// getVersion returns the application version from build info
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "1.0.0" // Fallback version
}

// handleHealth returns a simple health check response
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

// handleEdgeStatus returns edge (edge.functionfly.com) health, uptime, and request stats.
// GET /v1/status/edge
func (s *Server) handleEdgeStatus(w http.ResponseWriter, r *http.Request) {
	stats := monitoring.GetEdgeStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "failed to encode edge stats", http.StatusInternalServerError)
		return
	}
}

// handleAdminEdgeStatus returns edge stats in admin API response shape.
// GET /v1/admin/status/edge
func (s *Server) handleAdminEdgeStatus(w http.ResponseWriter, r *http.Request) {
	stats := monitoring.GetEdgeStats()
	w.Header().Set("Content-Type", "application/json")
	body := map[string]interface{}{
		"data":      stats,
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "failed to encode edge stats", http.StatusInternalServerError)
		return
	}
}

// handleDetailedHealth returns comprehensive health status
func (s *Server) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetRequestLogger(r)

	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   getVersion(),
		"services":  map[string]interface{}{},
		"checks":    []map[string]interface{}{},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Check database connectivity
	dbHealthy := s.checkDatabaseHealth(ctx)
	servicesMap, ok := health["services"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	servicesMap["database"] = map[string]interface{}{
		"status":    dbHealthy,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Check monitoring service
	monitoringHealthy := s.checkMonitoringHealth(ctx)
	servicesMap["monitoring"] = map[string]interface{}{
		"status":    monitoringHealthy,
		"message":   "monitoring disabled",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Check routing service
	routingHealthy := s.checkRoutingHealth(ctx)
	health["services"].(map[string]interface{})["routing"] = map[string]interface{}{
		"status":    routingHealthy,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Check notification service
	notificationHealth := s.checkNotificationHealth(ctx)
	servicesMap["notification"] = notificationHealth

	// Perform system health checks
	checks := []map[string]interface{}{}

	// Database connection check
	checks = append(checks, map[string]interface{}{
		"name":        "database_connection",
		"status":      s.statusString(dbHealthy),
		"description": "Database connectivity check",
		"timestamp":   time.Now().Format(time.RFC3339),
	})

	// Memory usage check
	memoryCheck := s.checkMemoryUsage()
	checks = append(checks, memoryCheck)

	// Disk space check
	diskCheck := s.checkDiskSpace()
	checks = append(checks, diskCheck)

	// Recent error rate check
	errorRateCheck := s.checkRecentErrorRate(ctx)
	checks = append(checks, errorRateCheck)

	// Backend health summary
	backendHealthCheck := s.checkBackendHealthSummary(ctx)
	checks = append(checks, backendHealthCheck)

	health["checks"] = checks

	// Determine overall health status: 503 only when DB is down; monitoring/routing failures → degraded, 200
	overallStatus := "healthy"
	statusCode := http.StatusOK

	if !dbHealthy {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else if !monitoringHealthy || !routingHealthy {
		overallStatus = "degraded"
		// Keep 200 so load balancers/clients don't treat as down when only monitoring/routing is missing
		statusCode = http.StatusOK
	} else {
		// Check if any critical checks failed
		for _, check := range checks {
			if status, ok := check["status"].(string); ok && status == "unhealthy" {
				if name, ok := check["name"].(string); ok && s.isCriticalCheck(name) {
					overallStatus = "degraded"
					statusCode = http.StatusOK // Still return 200 but indicate degradation
					break
				}
			}
		}
	}

	health["status"] = overallStatus

	logger.WithFields(logrus.Fields{
		"overall_status": overallStatus,
		"db_healthy":     dbHealthy,
		"checks_count":   len(checks),
	}).Info("Detailed health check completed")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

// handleHealthCheck performs a specific health check by name
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetRequestLogger(r)

	checkName := r.URL.Query().Get("name")
	if checkName == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var checkResult map[string]interface{}
	var statusCode int

	switch checkName {
	case "database":
		healthy := s.checkDatabaseHealth(r.Context())
		checkResult = map[string]interface{}{
			"name":      "database",
			"status":    s.statusString(healthy),
			"timestamp": time.Now().Format(time.RFC3339),
			"details":   "Database connectivity and basic query test",
		}
		statusCode = s.statusCodeFromHealth(healthy)

	case "memory":
		checkResult = s.checkMemoryUsage()
		statusCode = s.statusCodeFromCheck(checkResult)

	case "disk":
		checkResult = s.checkDiskSpace()
		statusCode = s.statusCodeFromCheck(checkResult)

	case "backends":
		checkResult = s.checkBackendHealthSummary(r.Context())
		statusCode = s.statusCodeFromCheck(checkResult)

	case "routing":
		healthy := s.checkRoutingHealth(r.Context())
		checkResult = map[string]interface{}{
			"name":      "routing",
			"status":    s.statusString(healthy),
			"timestamp": time.Now().Format(time.RFC3339),
			"details":   "Routing service availability and basic functionality",
		}
		statusCode = s.statusCodeFromHealth(healthy)

	case "notification":
		checkResult = s.checkNotificationHealth(r.Context())
		statusCode = s.statusCodeFromCheck(checkResult)

	default:
		http.Error(w, "Unknown health check: "+checkName, http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"check_name": checkName,
		"status":     checkResult["status"],
	}).Info("Individual health check completed")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(checkResult)
}

// checkDatabaseHealth verifies database connectivity
func (s *Server) checkDatabaseHealth(ctx context.Context) bool {
	// Simple query to test database connectivity
	_, err := s.repo.ListTenants(ctx)
	return err == nil
}

// checkMonitoringHealth verifies monitoring service health
func (s *Server) checkMonitoringHealth(ctx context.Context) bool {
	// Skip monitoring health check if monitoring is disabled
	if s.monitoringSvc == nil {
		return true // Consider monitoring healthy if disabled
	}
	// Check if we can get metrics (basic functionality test)
	_, err := s.monitoringSvc.GetSystemHealthStatus(ctx)
	return err == nil
}

// checkRoutingHealth verifies routing service health
func (s *Server) checkRoutingHealth(ctx context.Context) bool {
	// Check if routing service is responsive
	return s.routingSvc != nil
}

// checkNotificationHealth verifies notification service health
func (s *Server) checkNotificationHealth(ctx context.Context) map[string]interface{} {
	result := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if s.notificationSvc == nil {
		result["status"] = "unhealthy"
		result["message"] = "notification service not initialized"
		return result
	}

	health := s.notificationSvc.HealthCheck()
	if status, ok := health["status"].(string); ok {
		result["status"] = status
	}
	result["queue"] = health["queue"]
	result["channels"] = health["channels"]

	if queueSat, ok := health["queue"].(map[string]interface{})["saturation_pct"].(float64); ok && queueSat > 90 {
		result["message"] = "notification queue saturation critical"
	} else {
		result["message"] = "notification service operational"
	}

	return result
}

// checkMemoryUsage performs memory usage health check
func (s *Server) checkMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory usage metrics
	allocMB := float64(m.Alloc) / 1024 / 1024
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024
	heapAllocMB := float64(m.HeapAlloc) / 1024 / 1024
	heapSysMB := float64(m.HeapSys) / 1024 / 1024

	// Calculate heap utilization percentage
	heapUtilization := float64(0)
	if m.HeapSys > 0 {
		heapUtilization = float64(m.HeapAlloc) / float64(m.HeapSys) * 100
	}

	// Determine status based on heap utilization
	status := "healthy"
	description := "Memory usage within acceptable limits"

	if heapUtilization > 90 {
		status = "critical"
		description = "High memory usage detected"
	} else if heapUtilization > 80 {
		status = "warning"
		description = "Elevated memory usage"
	}

	return map[string]interface{}{
		"name":        "memory_usage",
		"status":      status,
		"description": description,
		"timestamp":   time.Now().Format(time.RFC3339),
		"details": map[string]interface{}{
			"current_alloc_mb":     allocMB,
			"total_alloc_mb":       totalAllocMB,
			"system_memory_mb":     sysMB,
			"heap_alloc_mb":        heapAllocMB,
			"heap_sys_mb":          heapSysMB,
			"heap_utilization_pct": heapUtilization,
			"gc_cycles":            m.NumGC,
			"next_gc_target_mb":    float64(m.NextGC) / 1024 / 1024,
		},
	}
}

// checkDiskSpace performs disk space health check
func (s *Server) checkDiskSpace() map[string]interface{} {
	// Get disk usage statistics for the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return map[string]interface{}{
			"name":        "disk_space",
			"status":      "unknown",
			"description": "Unable to determine current working directory",
			"timestamp":   time.Now().Format(time.RFC3339),
			"error":       err.Error(),
		}
	}

	var stat syscall.Statfs_t
	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return map[string]interface{}{
			"name":        "disk_space",
			"status":      "unknown",
			"description": "Unable to get filesystem statistics",
			"timestamp":   time.Now().Format(time.RFC3339),
			"error":       err.Error(),
		}
	}

	// Calculate disk usage metrics
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	usedBytes := totalBytes - availableBytes

	totalGB := float64(totalBytes) / 1024 / 1024 / 1024
	usedGB := float64(usedBytes) / 1024 / 1024 / 1024
	availableGB := float64(availableBytes) / 1024 / 1024 / 1024

	// Calculate usage percentage
	usagePercent := float64(0)
	if totalBytes > 0 {
		usagePercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	// Determine status based on usage percentage
	status := "healthy"
	description := "Disk space usage within acceptable limits"

	if usagePercent > 95 {
		status = "critical"
		description = "Critical disk space usage - immediate action required"
	} else if usagePercent > 90 {
		status = "warning"
		description = "High disk space usage detected"
	} else if usagePercent > 80 {
		status = "degraded"
		description = "Elevated disk space usage"
	}

	return map[string]interface{}{
		"name":        "disk_space",
		"status":      status,
		"description": description,
		"timestamp":   time.Now().Format(time.RFC3339),
		"details": map[string]interface{}{
			"total_gb":      totalGB,
			"used_gb":       usedGB,
			"available_gb":  availableGB,
			"usage_percent": usagePercent,
			"filesystem":    wd,
		},
	}
}

// checkRecentErrorRate checks recent error rates
func (s *Server) checkRecentErrorRate(ctx context.Context) map[string]interface{} {
	now := time.Now()
	since := now.Add(-time.Hour) // Check last hour

	// Get recent routing events
	events, err := s.repo.GetRecentRoutingEvents(ctx, 1000, since)
	if err != nil {
		return map[string]interface{}{
			"name":        "error_rate",
			"status":      "unknown",
			"description": "Unable to calculate error rate",
			"timestamp":   now.Format(time.RFC3339),
			"error":       err.Error(),
		}
	}

	// Calculate error rate
	totalRequests := len(events)
	errorCount := 0

	for _, event := range events {
		if event.Outcome != "success" {
			errorCount++
		}
	}

	errorRate := float64(0)
	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests)
	}

	// Determine status based on error rate
	status := "healthy"
	description := "Error rate within acceptable limits"
	threshold := 0.05 // 5% threshold

	if errorRate > 0.10 { // 10% critical
		status = "critical"
		description = "Critical error rate detected"
	} else if errorRate > threshold { // 5% warning
		status = "warning"
		description = "Elevated error rate detected"
	}

	return map[string]interface{}{
		"name":        "error_rate",
		"status":      status,
		"description": description,
		"timestamp":   now.Format(time.RFC3339),
		"details": map[string]interface{}{
			"period":         "1h",
			"error_rate":     errorRate,
			"threshold":      threshold,
			"total_requests": totalRequests,
			"error_count":    errorCount,
		},
	}
}

// checkBackendHealthSummary provides backend health summary
func (s *Server) checkBackendHealthSummary(ctx context.Context) map[string]interface{} {
	// Get all enabled backends for counting
	backends, backendsErr := s.repo.GetAllEnabledBackends(ctx)

	// Get backend health summary
	uptime, uptimeErr := s.calculateUptime(s.repo)

	// Handle errors
	if backendsErr != nil && uptimeErr != nil {
		return map[string]interface{}{
			"name":        "backend_health",
			"status":      "unknown",
			"description": "Unable to determine backend health status",
			"timestamp":   time.Now().Format(time.RFC3339),
			"error":       "Failed to get backend data and uptime",
		}
	}

	// Count total and healthy backends
	totalBackends := len(backends)
	healthyBackends := 0

	if backendsErr == nil {
		for _, backend := range backends {
			// Get recent health checks (last 10)
			checks, err := s.repo.GetRecentHealthChecks(ctx, backend.ID, 10)
			if err != nil {
				continue // Skip this backend if we can't get health data
			}

			if len(checks) > 0 && checks[0].OK {
				healthyBackends++
			}
		}
	}

	// Determine status based on uptime (if available)
	status := "unknown"
	if uptimeErr == nil {
		status = "healthy"
		if uptime < 95.0 {
			status = "degraded"
		}
		if uptime < 90.0 {
			status = "unhealthy"
		}
	}

	return map[string]interface{}{
		"name":        "backend_health",
		"status":      status,
		"description": "Overall backend health and uptime",
		"timestamp":   time.Now().Format(time.RFC3339),
		"details": map[string]interface{}{
			"uptime_percentage": uptime,
			"total_backends":    totalBackends,
			"healthy_backends":  healthyBackends,
		},
	}
}

// Helper methods

func (s *Server) statusString(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}

func (s *Server) statusCodeFromHealth(healthy bool) int {
	if healthy {
		return http.StatusOK
	}
	return http.StatusServiceUnavailable
}

func (s *Server) statusCodeFromCheck(check map[string]interface{}) int {
	if status, ok := check["status"].(string); ok && status == "healthy" {
		return http.StatusOK
	}
	return http.StatusServiceUnavailable
}

func (s *Server) isCriticalCheck(checkName string) bool {
	criticalChecks := []string{"database_connection", "backend_health"}
	for _, critical := range criticalChecks {
		if checkName == critical {
			return true
		}
	}
	return false
}

// handleDNAServiceHealth returns DNA service health status
func (s *Server) handleDNAServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var queueDepth float64
	if s.dnaRepo != nil {
		if depth, err := s.dnaRepo.GetQueueDepth(r.Context()); err == nil {
			queueDepth = float64(depth)
		}
	}
	dna.SetQueueDepth(queueDepth)

	circuitBreakerState := 0
	workerCount := 0
	if s.dnaService != nil {
		circuitBreakerState = s.dnaService.GetCircuitBreakerState()
		workerCount = s.dnaService.GetWorkerCount()
	}
	dna.SetCircuitBreakerState(float64(circuitBreakerState))

	partitionStatus := map[string]interface{}{}
	if s.dnaPartitionScheduler != nil {
		partitionStatus = s.dnaPartitionScheduler.GetStatus(r.Context())
	}

	insightsStatus := map[string]interface{}{}
	if s.dnaInsightsScheduler != nil {
		insightsStatus = s.dnaInsightsScheduler.GetStatus()
	}

	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": map[string]interface{}{
			"dna_service": map[string]interface{}{
				"status":              "ok",
				"queue_depth":         queueDepth,
				"worker_count":        workerCount,
				"circuit_breaker":     circuitBreakerState,
				"partition_scheduler": partitionStatus,
				"insights_scheduler":  insightsStatus,
			},
		},
	}

	statusCode := http.StatusOK
	if circuitBreakerState == 1 {
		health["status"] = "degraded"
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}
