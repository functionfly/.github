package monitoring

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// GetDatabaseHealth returns comprehensive database health metrics
func (s *Service) GetDatabaseHealth(ctx context.Context) (map[string]interface{}, error) {
	// Query the database for real metrics
	health, err := s.db.GetDatabaseHealthMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database health metrics: %w", err)
	}

	// Add timestamp
	health["lastUpdated"] = time.Now()

	return health, nil
}

// GetDatabaseMetrics returns time-series database performance metrics
func (s *Service) GetDatabaseMetrics(ctx context.Context, timeRange time.Duration) ([]map[string]interface{}, error) {
	now := time.Now()
	since := now.Add(-timeRange)

	// Query different metric types
	metricTypes := []string{"connections", "size_gb", "query_time", "cache_hit_ratio", "throughput"}
	allMetrics := make(map[time.Time]map[string]interface{})

	// Collect metrics for each type
	for _, metricType := range metricTypes {
		dbMetrics, err := s.db.QueryDatabaseMetrics(ctx, metricType, since, 100) // Get up to 100 data points per type
		if err != nil {
			// Log error but continue with other metrics
			continue
		}

		for _, metric := range dbMetrics {
			// Round timestamp to nearest 5 minutes for aggregation
			roundedTime := metric.RecordedAt.Truncate(5 * time.Minute)

			if allMetrics[roundedTime] == nil {
				allMetrics[roundedTime] = map[string]interface{}{
					"timestamp": roundedTime,
				}
			}

			// Store the metric value
			switch metric.MetricType {
			case "connections":
				allMetrics[roundedTime]["connections"] = metric.Value
			case "size_gb":
				allMetrics[roundedTime]["databaseSizeGB"] = metric.Value
			case "query_time":
				allMetrics[roundedTime]["avgResponseTime"] = metric.Value
			case "cache_hit_ratio":
				allMetrics[roundedTime]["cacheHitRatio"] = metric.Value
			case "throughput":
				allMetrics[roundedTime]["queryCount"] = metric.Value
			}
		}
	}

	// Convert to slice and sort by timestamp (most recent first)
	metrics := make([]map[string]interface{}, 0, len(allMetrics))
	for _, metric := range allMetrics {
		metrics = append(metrics, metric)
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(metrics, func(i, j int) bool {
		timeI := metrics[i]["timestamp"].(time.Time)
		timeJ := metrics[j]["timestamp"].(time.Time)
		return timeI.After(timeJ)
	})

	// If no historical data, fall back to current health metrics
	if len(metrics) == 0 {
		currentHealth, err := s.GetDatabaseHealth(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get database health metrics: %w", err)
		}

		// Create a single data point with current metrics
		currentMetric := map[string]interface{}{
			"timestamp":       now,
			"connections":     currentHealth["connections"].(map[string]interface{})["total"],
			"queryCount":      currentHealth["performance"].(map[string]interface{})["throughput"],
			"avgResponseTime": currentHealth["performance"].(map[string]interface{})["avgQueryTime"],
			"cacheHitRatio":   currentHealth["performance"].(map[string]interface{})["cacheHitRatio"],
			"databaseSizeGB":  currentHealth["storage"].(map[string]interface{})["usedGB"],
		}
		metrics = append(metrics, currentMetric)
	}

	return metrics, nil
}

// StartDatabaseMetricsCollection starts periodic collection and storage of database metrics
func (s *Service) StartDatabaseMetricsCollection(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Collect every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.collectAndStoreDatabaseMetrics(ctx)
			}
		}
	}()
}

// collectAndStoreDatabaseMetrics collects current database metrics and stores them for historical analysis
func (s *Service) collectAndStoreDatabaseMetrics(ctx context.Context) {
	// Get current health metrics
	healthMetrics, err := s.GetDatabaseHealth(ctx)
	if err != nil {
		// Log error but don't fail - metrics collection should be resilient
		return
	}

	// Store the metrics
	err = s.db.StoreDatabaseMetrics(ctx, healthMetrics)
	if err != nil {
		// Log error but continue - storage failures shouldn't stop collection
		return
	}
}

// GetDatabaseAlerts returns current database-specific alerts
func (s *Service) GetDatabaseAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	// Query active alerts from the database
	activeAlerts, err := s.queryActiveAlerts(nil) // nil for tenantID to get all alerts
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}

	// Transform Alert structs to map format for API response
	alerts := make([]map[string]interface{}, 0, len(activeAlerts))
	for _, alert := range activeAlerts {
		alertMap := map[string]interface{}{
			"id":        alert.ID.String(),
			"type":      alert.AlertType,
			"severity":  alert.Severity,
			"title":     alert.Title,
			"message":   alert.Message,
			"timestamp": alert.CreatedAt,
			"resolved":  alert.Status == "resolved",
		}

		// Add optional fields if they exist
		if alert.TenantID != nil {
			alertMap["tenant_id"] = alert.TenantID.String()
		}
		if alert.AppID != nil {
			alertMap["app_id"] = alert.AppID.String()
		}
		if alert.BackendID != nil {
			alertMap["backend_id"] = alert.BackendID.String()
		}
		if alert.ResolvedAt != nil {
			alertMap["resolved_at"] = alert.ResolvedAt
		}
		if alert.ResolvedBy != nil {
			alertMap["resolved_by"] = alert.ResolvedBy.String()
		}
		if alert.Metadata != nil {
			alertMap["metadata"] = alert.Metadata
		}

		alerts = append(alerts, alertMap)
	}

	return alerts, nil
}

// CheckDatabaseConnectionPool monitors database connection pool health
func (s *Service) CheckDatabaseConnectionPool(ctx context.Context) error {
	// Get current database health metrics to check connection pool status
	healthMetrics, err := s.GetDatabaseHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database health metrics: %w", err)
	}

	// Extract connection pool information
	connections, ok := healthMetrics["connections"].(map[string]interface{})
	if !ok {
		// If connection data is not available, skip the check
		return nil
	}

	activeConnections, _ := connections["active"].(float64)
	maxConnections, _ := connections["max"].(float64)

	if maxConnections == 0 {
		// Cannot calculate usage if max connections is not available
		return nil
	}

	usagePercent := activeConnections / maxConnections

	if usagePercent > 0.9 {
		// Critical usage alert
		alert := &storage.Alert{
			AlertType: "connection_pool_exhausted",
			Severity:  "high",
			Title:     "Connection Pool Near Capacity",
			Message:   fmt.Sprintf("Database connection pool utilization is at %.1f%% (%d/%d connections)", usagePercent*100, int(activeConnections), int(maxConnections)),
			Status:    "active",
			Metadata: map[string]interface{}{
				"active_connections": int(activeConnections),
				"max_connections":    int(maxConnections),
				"usage_percent":      usagePercent,
				"managed_by":         connections["managed_by"],
			},
		}

		return s.RecordAlert(ctx, alert)
	} else if usagePercent > 0.75 {
		// Warning level
		alert := &storage.Alert{
			AlertType: "connection_pool_warning",
			Severity:  "warning",
			Title:     "High Connection Pool Usage",
			Message:   fmt.Sprintf("Database connection pool utilization is at %.1f%% (%d/%d connections)", usagePercent*100, int(activeConnections), int(maxConnections)),
			Status:    "active",
			Metadata: map[string]interface{}{
				"active_connections": int(activeConnections),
				"max_connections":    int(maxConnections),
				"usage_percent":      usagePercent,
				"managed_by":         connections["managed_by"],
			},
		}

		return s.RecordAlert(ctx, alert)
	}

	return nil
}

// CheckDatabasePerformance monitors query performance and latency
func (s *Service) CheckDatabasePerformance(ctx context.Context) error {
	// Get current database health metrics to check performance
	healthMetrics, err := s.GetDatabaseHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database health metrics: %w", err)
	}

	// Extract performance information
	performance, ok := healthMetrics["performance"].(map[string]interface{})
	if !ok {
		// If performance data is not available, skip the check
		return nil
	}

	avgQueryTime, _ := performance["avgQueryTime"].(float64)
	slowQueryThreshold := 100.0 // milliseconds - configurable threshold

	if avgQueryTime > slowQueryThreshold {
		// Count slow queries from recent metrics (this is a simplified approach)
		slowQueryCount := 0
		recentMetrics, err := s.queryPerformanceMetrics("database_query_time", nil, time.Now().Add(-1*time.Hour))
		if err == nil {
			for _, metric := range recentMetrics {
				if metric.Value > slowQueryThreshold {
					slowQueryCount++
				}
			}
		}

		alert := &storage.Alert{
			AlertType: "high_query_latency",
			Severity:  "warning",
			Title:     "High Query Response Time",
			Message:   fmt.Sprintf("Average query response time (%.1fms) exceeded threshold (%.1fms)", avgQueryTime, slowQueryThreshold),
			Status:    "active",
			Metadata: map[string]interface{}{
				"avg_response_time":    avgQueryTime,
				"slow_query_threshold": slowQueryThreshold,
				"slow_query_count":     slowQueryCount,
				"throughput":           performance["throughput"],
				"cache_hit_ratio":      performance["cacheHitRatio"],
				"managed_by":           performance["managed_by"],
			},
		}

		return s.RecordAlert(ctx, alert)
	}

	return nil
}

// CheckDatabaseStorage monitors database storage usage and growth
func (s *Service) CheckDatabaseStorage(ctx context.Context) error {
	// Get current database health metrics to check storage
	healthMetrics, err := s.GetDatabaseHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database health metrics: %w", err)
	}

	// Extract storage information
	storageInfo, ok := healthMetrics["storage"].(map[string]interface{})
	if !ok {
		// If storage data is not available, skip the check
		return nil
	}

	usedGB, _ := storageInfo["usedGB"].(float64)
	availablePercent, _ := storageInfo["availablePercent"].(float64)
	growthRate, _ := storageInfo["growthRate"].(float64)

	// Calculate usage percentage (100 - available = used)
	usagePercent := 100.0 - availablePercent

	if usagePercent > 90 {
		alert := &storage.Alert{
			AlertType: "storage_critical",
			Severity:  "critical",
			Title:     "Critical Database Storage Usage",
			Message:   fmt.Sprintf("Database storage usage is at %.1f%% (%.1fGB used, %.1f%% available)", usagePercent, usedGB, availablePercent),
			Status:    "active",
			Metadata: map[string]interface{}{
				"used_gb":           usedGB,
				"available_percent": availablePercent,
				"usage_percent":     usagePercent,
				"growth_rate":       growthRate,
				"managed_by":        storageInfo["managed_by"],
			},
		}

		return s.RecordAlert(ctx, alert)
	} else if usagePercent > 80 {
		alert := &storage.Alert{
			AlertType: "storage_warning",
			Severity:  "warning",
			Title:     "High Database Storage Usage",
			Message:   fmt.Sprintf("Database storage usage is at %.1f%% (%.1fGB used, %.1f%% available)", usagePercent, usedGB, availablePercent),
			Status:    "active",
			Metadata: map[string]interface{}{
				"used_gb":           usedGB,
				"available_percent": availablePercent,
				"usage_percent":     usagePercent,
				"growth_rate":       growthRate,
				"managed_by":        storageInfo["managed_by"],
			},
		}

		return s.RecordAlert(ctx, alert)
	}

	return nil
}

// CheckDatabaseReplication monitors replication lag and status
func (s *Service) CheckDatabaseReplication(ctx context.Context) error {
	// Get current database health metrics to check replication
	healthMetrics, err := s.GetDatabaseHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database health metrics: %w", err)
	}

	// Extract replication information
	replication, ok := healthMetrics["replication"].(map[string]interface{})
	if !ok {
		// If replication data is not available, skip the check
		return nil
	}

	lag, _ := replication["lag"].(float64)
	maxAcceptableLag := 100.0 // milliseconds - configurable threshold

	if lag > maxAcceptableLag {
		alert := &storage.Alert{
			AlertType: "replication_lag",
			Severity:  "error",
			Title:     "High Database Replication Lag",
			Message:   fmt.Sprintf("Database replication lag is %.1fms (threshold: %.1fms)", lag, maxAcceptableLag),
			Status:    "active",
			Metadata: map[string]interface{}{
				"replication_lag_ms":    lag,
				"max_acceptable_lag_ms": maxAcceptableLag,
				"replication_status":    replication["status"],
				"replicas":              replication["replicas"],
				"provider":              replication["provider"],
			},
		}

		return s.RecordAlert(ctx, alert)
	}

	return nil
}