package monitoring

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// Private methods for database operations - these delegate to the repository layer

// insertPerformanceMetric inserts a performance metric into the database
func (s *Service) insertPerformanceMetric(metric *storage.PerformanceMetric) error {
	return s.db.InsertPerformanceMetric(metric)
}

// insertAlert inserts an alert into the database
func (s *Service) insertAlert(alert *storage.Alert) error {
	return s.db.InsertAlert(alert)
}

// insertSystemHealthCheck inserts a system health check into the database
func (s *Service) insertSystemHealthCheck(check *storage.SystemHealthCheck) error {
	return s.db.InsertSystemHealthCheck(check)
}

// insertMonitoringEvent inserts a monitoring event into the database
func (s *Service) insertMonitoringEvent(event *storage.MonitoringEvent) error {
	return s.db.InsertMonitoringEvent(event)
}

// updateAlertStatus updates an alert's status
func (s *Service) updateAlertStatus(alert *storage.Alert) error {
	return s.db.UpdateAlertStatus(alert)
}

// queryPerformanceMetrics retrieves performance metrics
func (s *Service) queryPerformanceMetrics(metricType string, tenantID *uuid.UUID, since time.Time) ([]*storage.PerformanceMetric, error) {
	const defaultLimit = 1000
	return s.db.QueryPerformanceMetrics(metricType, tenantID, since, defaultLimit)
}

// queryActiveAlerts retrieves active alerts
func (s *Service) queryActiveAlerts(tenantID *uuid.UUID) ([]*storage.Alert, error) {
	return s.db.QueryActiveAlerts(tenantID)
}

// queryLatestSystemHealthChecks retrieves the latest system health checks
func (s *Service) queryLatestSystemHealthChecks() (map[string]*storage.SystemHealthCheck, error) {
	return s.db.QueryLatestSystemHealthChecks()
}