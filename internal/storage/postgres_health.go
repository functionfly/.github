package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"
)

// ============================================================================
// Health Monitoring
// ============================================================================

// startHealthMonitoring starts background health checking for all database connections
func (db *PostgresDB) startHealthMonitoring() {
	db.healthCheckDone = make(chan bool)
	db.healthCheckTicker = time.NewTicker(db.config.HealthCheckInterval)

	go func() {
		defer db.healthCheckTicker.Stop()
		for {
			select {
			case <-db.healthCheckDone:
				return
			case <-db.healthCheckTicker.C:
				db.performHealthChecks()
			}
		}
	}()

	logrus.WithField("interval", db.config.HealthCheckInterval).Info("Started database health monitoring")
}

// performHealthChecks performs health checks on primary and all read replicas
func (db *PostgresDB) performHealthChecks() {
	// Check primary connection
	db.checkConnectionHealth(db.DB, "primary", db.config.Host, db.config.Port)

	// Check read replicas
	for i := range db.readReplicas {
		replica := &db.readReplicas[i]
		healthy := db.checkConnectionHealth(replica.DB, "replica", replica.Host, replica.Port)
		replica.LastHealthCheck = time.Now()

		if healthy {
			replica.Healthy = true
			replica.Failures = 0
		} else {
			replica.Failures++
			if replica.Failures >= db.config.MaxHealthFailures {
				replica.Healthy = false
				logrus.WithFields(logrus.Fields{
					"host":     replica.Host,
					"port":     replica.Port,
					"failures": replica.Failures,
				}).Warn("Read replica marked as unhealthy")
			}
		}
	}
}

// checkConnectionHealth performs a health check on a single database connection
func (db *PostgresDB) checkConnectionHealth(conn *sql.DB, connType, host string, port int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), db.config.HealthCheckTimeout)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"type": connType,
			"host": host,
			"port": port,
		}).WithError(err).Warn("Database health check failed")
		return false
	}

	return true
}
