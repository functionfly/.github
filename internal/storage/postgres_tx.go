package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ============================================================================
// Transaction Management (delegates to TransactionManager)
// ============================================================================

// TransactionManager returns the transaction manager for advanced transaction handling
func (db *PostgresDB) TransactionManager() *TransactionManager {
	return db.transactionManager
}

// ExecuteInTransaction executes a function within a transaction with timeout
func (db *PostgresDB) ExecuteInTransaction(ctx context.Context, opts *TransactionOptions, fn func(*gorm.DB) error) error {
	return db.transactionManager.ExecuteInTransaction(ctx, opts, fn)
}

// ExecuteInReadTransaction executes read-only operations with snapshot isolation
func (db *PostgresDB) ExecuteInReadTransaction(ctx context.Context, timeout time.Duration, fn func(*gorm.DB) error) error {
	return db.transactionManager.ExecuteInReadTransaction(ctx, timeout, fn)
}

// ExecuteSaga executes a saga pattern transaction
func (db *PostgresDB) ExecuteSaga(ctx context.Context, opts *TransactionOptions, steps []SagaStep) error {
	return db.transactionManager.ExecuteSaga(ctx, opts, steps)
}

// ExecuteWithRetry executes a transaction with retry logic
func (db *PostgresDB) ExecuteWithRetry(ctx context.Context, opts *TransactionOptions, maxRetries int, fn func(*gorm.DB) error) error {
	return db.transactionManager.ExecuteWithRetry(ctx, opts, maxRetries, fn)
}

// NewTransactionScope creates a transaction scope builder
func (db *PostgresDB) NewTransactionScope(ctx context.Context) *TransactionScope {
	return db.transactionManager.NewTransactionScope(ctx)
}

// ============================================================================
// Query Performance Monitoring
// ============================================================================

// QueryMonitor returns the query performance monitor
func (db *PostgresDB) QueryMonitor() *QueryMonitor {
	return db.queryMonitor
}

// EnableSlowQueryLogging enables slow query logging
func (db *PostgresDB) EnableSlowQueryLogging(threshold time.Duration) {
	db.queryMonitor.EnableSlowQueryLogging(threshold)
}

// GetQueryStats returns current query performance statistics
func (db *PostgresDB) GetQueryStats() map[string]*QueryStats {
	return db.queryMonitor.GetQueryStats()
}

// GetSlowQueries returns queries exceeding the slow query threshold
func (db *PostgresDB) GetSlowQueries() []*QueryStats {
	return db.queryMonitor.GetSlowQueries()
}

// ============================================================================
// Incident Operations
// ============================================================================

// CreateIncident creates a new incident
func (db *PostgresDB) CreateIncident(ctx context.Context, incident *Incident) (*Incident, error) {
	return db.incidentRepository.CreateIncident(ctx, incident)
}

// GetIncidentByID retrieves an incident by ID
func (db *PostgresDB) GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*Incident, error) {
	return db.incidentRepository.GetIncidentByID(ctx, incidentID)
}

// ListIncidents retrieves incidents with pagination
func (db *PostgresDB) ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*Incident, error) {
	return db.incidentRepository.ListIncidents(ctx, limit, offset, status)
}

// ListIncidentsSince retrieves incidents since a given time
func (db *PostgresDB) ListIncidentsSince(ctx context.Context, since time.Time, limit int) ([]*Incident, error) {
	return db.incidentRepository.ListIncidentsSince(ctx, since, limit)
}

// CountIncidentsSince counts incidents since a given time
func (db *PostgresDB) CountIncidentsSince(ctx context.Context, since time.Time) (int, error) {
	return db.incidentRepository.CountIncidentsSince(ctx, since)
}

// GetTotalDowntimeMinutesSince returns total downtime minutes since a given time
func (db *PostgresDB) GetTotalDowntimeMinutesSince(ctx context.Context, since time.Time) (int, error) {
	return db.incidentRepository.GetTotalDowntimeMinutesSince(ctx, since)
}

// CountIncidentsGroupedByDay groups incidents by day since a given time
func (db *PostgresDB) CountIncidentsGroupedByDay(ctx context.Context, since time.Time) ([]DailyIncidentCount, error) {
	return db.incidentRepository.CountIncidentsGroupedByDay(ctx, since)
}

// UpdateIncident updates an incident
func (db *PostgresDB) UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*Incident, error) {
	return db.incidentRepository.UpdateIncident(ctx, incidentID, updates)
}

// ResolveIncident marks an incident as resolved
func (db *PostgresDB) ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*Incident, error) {
	return db.incidentRepository.ResolveIncident(ctx, incidentID)
}

// ============================================================================
// Graceful Shutdown
// ============================================================================

// Close gracefully shuts down the database connections and health monitoring
func (db *PostgresDB) Close() error {
	// Stop health monitoring
	if db.healthCheckDone != nil {
		db.healthCheckDone <- true
		close(db.healthCheckDone)
	}

	// Close prepared statements
	db.closePreparedStatements()

	// Close read replica connections
	for _, replica := range db.readReplicas {
		if err := replica.DB.Close(); err != nil {
			logrus.WithFields(logrus.Fields{
				"host": replica.Host,
				"port": replica.Port,
			}).WithError(err).Warn("Failed to close read replica connection")
		}
	}

	return db.DB.Close()
}
