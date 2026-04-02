package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ============================================================================
// Read Replica Management
// ============================================================================

// initReadReplicas initializes connections to read replica databases
func (db *PostgresDB) initReadReplicas(config *DatabaseConfig) error {
	for _, replicaConfig := range config.ReadReplicas {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			replicaConfig.Host, replicaConfig.Port, config.User, config.Password, config.Database, config.SSLMode)

		replicaDB, err := sql.Open("postgres", connStr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"host": replicaConfig.Host,
				"port": replicaConfig.Port,
			}).WithError(err).Error("Failed to connect to read replica")
			continue
		}

		// Configure connection pool for replica
		if err := configureConnectionPool(replicaDB, config); err != nil {
			replicaDB.Close()
			logrus.WithFields(logrus.Fields{
				"host": replicaConfig.Host,
				"port": replicaConfig.Port,
			}).WithError(err).Error("Failed to configure read replica connection pool")
			continue
		}

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := replicaDB.PingContext(ctx); err != nil {
			cancel()
			replicaDB.Close()
			logrus.WithFields(logrus.Fields{
				"host": replicaConfig.Host,
				"port": replicaConfig.Port,
			}).WithError(err).Error("Read replica health check failed")
			continue
		}
		cancel()

		replicaConn := ReadReplicaConnection{
			DB:              replicaDB,
			Host:            replicaConfig.Host,
			Port:            replicaConfig.Port,
			Weight:          replicaConfig.Weight,
			Priority:        replicaConfig.Priority,
			Region:          replicaConfig.Region,
			Healthy:         true,
			Failures:        0,
			LastHealthCheck: time.Now(),
		}

		db.readReplicas = append(db.readReplicas, replicaConn)

		logrus.WithFields(logrus.Fields{
			"host":     replicaConfig.Host,
			"port":     replicaConfig.Port,
			"weight":   replicaConfig.Weight,
			"priority": replicaConfig.Priority,
			"region":   replicaConfig.Region,
		}).Info("Connected to read replica")
	}

	if len(db.readReplicas) == 0 {
		return fmt.Errorf("no read replicas could be initialized")
	}

	logrus.WithField("count", len(db.readReplicas)).Info("Successfully initialized read replicas")
	return nil
}

// getReadReplica returns a healthy read replica connection using weighted load balancing
func (db *PostgresDB) getReadReplica() *sql.DB {
	if !db.readReplicaEnabled || len(db.readReplicas) == 0 {
		return db.DB // Fall back to primary
	}

	// Filter healthy replicas
	var healthyReplicas []ReadReplicaConnection
	for _, replica := range db.readReplicas {
		if replica.Healthy {
			healthyReplicas = append(healthyReplicas, replica)
		}
	}

	if len(healthyReplicas) == 0 {
		logrus.Warn("No healthy read replicas available, falling back to primary")
		return db.DB
	}

	// Simple weighted random selection (could be enhanced with more sophisticated load balancing)
	totalWeight := 0
	for _, replica := range healthyReplicas {
		totalWeight += replica.Weight
	}

	if totalWeight == 0 {
		// If no weights specified, use round-robin
		return healthyReplicas[0].DB
	}

	// Simple random weighted selection
	randomWeight := time.Now().UnixNano() % int64(totalWeight)
	currentWeight := 0

	for _, replica := range healthyReplicas {
		currentWeight += replica.Weight
		if int64(currentWeight) > randomWeight {
			return replica.DB
		}
	}

	// Fallback to first healthy replica
	return healthyReplicas[0].DB
}

// GetReadDB returns a read-optimized database connection (replica if available, primary otherwise)
func (db *PostgresDB) GetReadDB() *sql.DB {
	if db.readReplicaEnabled && len(db.readReplicas) > 0 {
		return db.getReadReplica()
	}
	return db.DB
}

// GetStats returns connection pool statistics for monitoring
func (db *PostgresDB) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Primary connection stats
	primaryStats := db.DB.Stats()
	stats["primary"] = map[string]interface{}{
		"open_connections":    primaryStats.OpenConnections,
		"in_use":              primaryStats.InUse,
		"idle":                primaryStats.Idle,
		"wait_count":          primaryStats.WaitCount,
		"wait_duration":       primaryStats.WaitDuration,
		"max_idle_closed":     primaryStats.MaxIdleClosed,
		"max_lifetime_closed": primaryStats.MaxLifetimeClosed,
	}

	// Read replica stats
	replicas := make([]map[string]interface{}, len(db.readReplicas))
	for i, replica := range db.readReplicas {
		replicaStats := replica.DB.Stats()
		replicas[i] = map[string]interface{}{
			"host":              replica.Host,
			"port":              replica.Port,
			"region":            replica.Region,
			"weight":            replica.Weight,
			"priority":          replica.Priority,
			"healthy":           replica.Healthy,
			"failures":          replica.Failures,
			"last_health_check": replica.LastHealthCheck,
			"open_connections":  replicaStats.OpenConnections,
			"in_use":            replicaStats.InUse,
			"idle":              replicaStats.Idle,
			"wait_count":        replicaStats.WaitCount,
			"wait_duration":     replicaStats.WaitDuration,
		}
	}
	stats["read_replicas"] = replicas

	return stats
}
