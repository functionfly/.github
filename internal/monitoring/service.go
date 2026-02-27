package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service provides comprehensive monitoring capabilities using Supabase's built-in tools
type Service struct {
	db            storage.Repository
	alertChans    map[string]chan *storage.Alert // Alert channels for real-time notifications
	alertEngine   *AlertEngine                   // Automatic alerting engine
	mu            sync.RWMutex
}

// NewService creates a new monitoring service
func NewService(db storage.Repository) *Service {
	service := &Service{
		db:         db,
		alertChans: make(map[string]chan *storage.Alert),
	}

	// Initialize alert engine with default rules
	service.alertEngine = NewAlertEngine(service)

	return service
}

// RecordPerformanceMetric records a performance metric
func (s *Service) RecordPerformanceMetric(ctx context.Context, metric *storage.PerformanceMetric) error {
	// Set ID if not provided
	if metric.ID == uuid.Nil {
		metric.ID = uuid.New()
	}

	// Insert the metric
	err := s.insertPerformanceMetric(metric)
	if err != nil {
		return err
	}

	// Check alert rules against this metric
	go s.alertEngine.ProcessMetric(metric)

	return nil
}

// RecordAlert creates and records an alert
func (s *Service) RecordAlert(ctx context.Context, alert *storage.Alert) error {
	// Set ID if not provided
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = now
	}
	alert.UpdatedAt = now

	// Insert alert into database
	if err := s.insertAlert(alert); err != nil {
		return fmt.Errorf("failed to insert alert: %w", err)
	}

	// Broadcast alert to real-time subscribers using Supabase's LISTEN/NOTIFY
	if err := s.broadcastAlert(alert); err != nil {
		logrus.WithError(err).WithField("alert_id", alert.ID).Warn("Failed to broadcast alert")
	}

	// Send alert to registered channels for real-time processing
	s.broadcastToChannels(alert)

	logrus.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"alert_type": alert.AlertType,
		"severity":   alert.Severity,
	}).Info("Alert recorded and broadcasted")

	return nil
}

// RecordSystemHealthCheck records a system health check result
func (s *Service) RecordSystemHealthCheck(ctx context.Context, check *storage.SystemHealthCheck) error {
	// Set ID if not provided
	if check.ID == uuid.Nil {
		check.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	check.CheckedAt = now
	if check.CreatedAt.IsZero() {
		check.CreatedAt = now
	}

	return s.insertSystemHealthCheck(check)
}

// RecordMonitoringEvent records a monitoring event for real-time tracking
func (s *Service) RecordMonitoringEvent(ctx context.Context, event *storage.MonitoringEvent) error {
	// Set ID if not provided
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	event.Timestamp = now
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}

	// Insert event
	if err := s.insertMonitoringEvent(event); err != nil {
		return fmt.Errorf("failed to insert monitoring event: %w", err)
	}

	// Broadcast event using Supabase real-time
	if err := s.broadcastMonitoringEvent(event); err != nil {
		logrus.WithError(err).WithField("event_id", event.ID).Warn("Failed to broadcast monitoring event")
	}

	return nil
}

// SubscribeToAlerts creates a channel for receiving real-time alerts
func (s *Service) SubscribeToAlerts(subscriberID string) <-chan *storage.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *storage.Alert, 100) // Buffered channel to prevent blocking
	s.alertChans[subscriberID] = ch

	logrus.WithField("subscriber_id", subscriberID).Debug("Alert subscriber added")

	return ch
}

// UnsubscribeFromAlerts removes an alert subscription
func (s *Service) UnsubscribeFromAlerts(subscriberID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, exists := s.alertChans[subscriberID]; exists {
		close(ch)
		delete(s.alertChans, subscriberID)
		logrus.WithField("subscriber_id", subscriberID).Debug("Alert subscriber removed")
	}
}

// GetMetrics retrieves performance metrics with optional filtering
func (s *Service) GetMetrics(ctx context.Context, metricType string, tenantID *uuid.UUID, since time.Time) ([]*storage.PerformanceMetric, error) {
	return s.queryPerformanceMetrics(metricType, tenantID, since)
}

// GetActiveAlerts retrieves currently active alerts
func (s *Service) GetActiveAlerts(ctx context.Context, tenantID *uuid.UUID) ([]*storage.Alert, error) {
	return s.queryActiveAlerts(tenantID)
}

// GetSystemHealthStatus returns the current system health status
func (s *Service) GetSystemHealthStatus(ctx context.Context) (map[string]*storage.SystemHealthCheck, error) {
	return s.queryLatestSystemHealthChecks()
}

// ResolveAlert marks an alert as resolved
func (s *Service) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolvedBy uuid.UUID) error {
	now := time.Now()

	alert := &storage.Alert{
		ID:         alertID,
		Status:     "resolved",
		ResolvedAt: &now,
		ResolvedBy: &resolvedBy,
		UpdatedAt:  now,
	}

	if err := s.updateAlertStatus(alert); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	// Broadcast resolution
	alert.Status = "resolved"
	s.broadcastToChannels(alert)

	return nil
}

// Dashboard configuration methods

// GetDashboardConfigsByTenant retrieves dashboard configurations for a tenant
func (s *Service) GetDashboardConfigsByTenant(tenantID uuid.UUID) ([]*storage.DashboardConfig, error) {
	return s.db.GetDashboardConfigsByTenant(tenantID)
}

// GetDashboardConfigsByUser retrieves dashboard configurations for a specific user
func (s *Service) GetDashboardConfigsByUser(userID uuid.UUID) ([]*storage.DashboardConfig, error) {
	return s.db.GetDashboardConfigsByUser(userID)
}

// CreateDashboardConfig creates a new dashboard configuration
func (s *Service) CreateDashboardConfig(ctx context.Context, config *storage.DashboardConfig) (*storage.DashboardConfig, error) {
	return s.db.CreateDashboardConfig(ctx, config)
}

// UpdateDashboardConfig updates an existing dashboard configuration
func (s *Service) UpdateDashboardConfig(ctx context.Context, configID uuid.UUID, updates map[string]interface{}) (*storage.DashboardConfig, error) {
	return s.db.UpdateDashboardConfig(ctx, configID, updates)
}

// DeleteDashboardConfig deletes a dashboard configuration
func (s *Service) DeleteDashboardConfig(ctx context.Context, configID uuid.UUID) error {
	return s.db.DeleteDashboardConfig(ctx, configID)
}

// Local Runtime Registry Management

// CleanupStaleLocalRuntimes removes runtime instances that haven't sent heartbeats recently
func (s *Service) CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error) {
	return s.db.CleanupStaleLocalRuntimes(ctx, maxAge)
}

// StartLocalRuntimeCleanup starts a periodic cleanup of stale local runtime instances
func (s *Service) StartLocalRuntimeCleanup(ctx context.Context, interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logrus.WithFields(logrus.Fields{
		"interval": interval.String(),
		"max_age":  maxAge.String(),
	}).Info("Started local runtime cleanup service")

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping local runtime cleanup service")
			return
		case <-ticker.C:
			cleaned, err := s.CleanupStaleLocalRuntimes(ctx, maxAge)
			if err != nil {
				logrus.WithError(err).Warn("Failed to cleanup stale local runtimes")
			} else if cleaned > 0 {
				logrus.WithField("count", cleaned).Info("Cleaned up stale local runtime instances")
			}
		}
	}
}

// GetLocalRuntimeHealthStatus returns health status for all active local runtimes
func (s *Service) GetLocalRuntimeHealthStatus(ctx context.Context) (map[string]interface{}, error) {
	activeRuntimes, err := s.db.ListActiveLocalRuntimes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active runtimes: %w", err)
	}

	healthStatus := map[string]interface{}{
		"total_active": len(activeRuntimes),
		"healthy":      0,
		"degraded":     0,
		"unhealthy":    0,
		"unknown":      0,
		"runtimes":     make([]map[string]interface{}, 0, len(activeRuntimes)),
	}

	for _, runtime := range activeRuntimes {
		runtimeHealth := "unknown"

		// Get latest metrics to determine health
		latestMetrics, err := s.db.GetLatestLocalRuntimeMetrics(ctx, runtime.ID)
		if err == nil && latestMetrics != nil {
			// Determine health based on error rate and resource usage
			if latestMetrics.ErrorRate > 20.0 {
				runtimeHealth = "unhealthy"
			} else if latestMetrics.ErrorRate > 5.0 || latestMetrics.CPUUsage > 90.0 {
				runtimeHealth = "degraded"
			} else {
				runtimeHealth = "healthy"
			}
		}

		// Update counters
		switch runtimeHealth {
		case "healthy":
			healthStatus["healthy"] = healthStatus["healthy"].(int) + 1
		case "degraded":
			healthStatus["degraded"] = healthStatus["degraded"].(int) + 1
		case "unhealthy":
			healthStatus["unhealthy"] = healthStatus["unhealthy"].(int) + 1
		default:
			healthStatus["unknown"] = healthStatus["unknown"].(int) + 1
		}

		// Add runtime details
		runtimeInfo := map[string]interface{}{
			"id":             runtime.ID.String(),
			"runtime_id":     runtime.RuntimeID,
			"runtime_type":   runtime.RuntimeType,
			"function_name":  runtime.FunctionName,
			"host":          runtime.Host,
			"port":          runtime.Port,
			"status":        runtime.Status,
			"health":        runtimeHealth,
			"uptime":        runtime.Uptime,
			"last_heartbeat": runtime.LastHeartbeat,
		}

		// Add metrics if available
		if latestMetrics != nil {
			runtimeInfo["metrics"] = map[string]interface{}{
				"cpu_usage":          latestMetrics.CPUUsage,
				"memory_heap":        latestMetrics.MemoryUsage.Heap,
				"memory_stack":       latestMetrics.MemoryUsage.Stack,
				"memory_system":      latestMetrics.MemoryUsage.System,
				"active_connections": latestMetrics.ActiveConnections,
				"request_throughput": latestMetrics.RequestThroughput,
				"total_requests":     latestMetrics.TotalRequests,
				"error_rate":         latestMetrics.ErrorRate,
			}
		}

		healthStatus["runtimes"] = append(healthStatus["runtimes"].([]map[string]interface{}), runtimeInfo)
	}

	return healthStatus, nil
}


