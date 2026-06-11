package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MonitoringRepository handles monitoring-related database operations
type MonitoringRepository struct {
	db *PostgresDB
}

// NewMonitoringRepository creates a new monitoring repository
func NewMonitoringRepository(db *PostgresDB) *MonitoringRepository {
	return &MonitoringRepository{db: db}
}

// InsertPerformanceMetric inserts a performance metric into the database
func (r *MonitoringRepository) InsertPerformanceMetric(ctx context.Context, metric *PerformanceMetric) error {
	labelsJSON, err := json.Marshal(metric.Labels)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO performance_metrics (
			metric_type, tenant_id, app_id, backend_id, value, unit, labels, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.db.ExecContext(ctx, query,
		metric.MetricType,
		metric.TenantID,
		metric.AppID,
		metric.BackendID,
		metric.Value,
		metric.Unit,
		labelsJSON,
		metric.Timestamp,
	)

	return err
}

// InsertAlert inserts an alert into the database
func (r *MonitoringRepository) InsertAlert(ctx context.Context, alert *Alert) error {
	metadataJSON, err := json.Marshal(alert.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO alerts (
			alert_type, severity, tenant_id, app_id, backend_id, title, message, status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = r.db.ExecContext(ctx, query,
		alert.AlertType,
		alert.Severity,
		alert.TenantID,
		alert.AppID,
		alert.BackendID,
		alert.Title,
		alert.Message,
		alert.Status,
		metadataJSON,
	)

	return err
}

// InsertSystemHealthCheck inserts a system health check into the database
func (r *MonitoringRepository) InsertSystemHealthCheck(ctx context.Context, check *SystemHealthCheck) error {
	metadataJSON, err := json.Marshal(check.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO system_health_checks (
			check_type, component_name, status, response_time_ms, message, metadata, checked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = r.db.ExecContext(ctx, query,
		check.CheckType,
		check.ComponentName,
		check.Status,
		check.ResponseTimeMs,
		check.Message,
		metadataJSON,
		check.CheckedAt,
	)

	return err
}

// InsertMonitoringEvent inserts a monitoring event into the database
func (r *MonitoringRepository) InsertMonitoringEvent(ctx context.Context, event *MonitoringEvent) error {
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO monitoring_events (
			event_type, tenant_id, app_id, backend_id, request_id, user_id, data, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.db.ExecContext(ctx, query,
		event.EventType,
		event.TenantID,
		event.AppID,
		event.BackendID,
		event.RequestID,
		event.UserID,
		dataJSON,
		event.Timestamp,
	)

	return err
}

// QueryMonitoringEvents retrieves monitoring events with filtering
func (r *MonitoringRepository) QueryMonitoringEvents(ctx context.Context, eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*MonitoringEvent, error) {
	var rows *sql.Rows
	var err error

	// Build query based on filters
	if tenantID != nil && eventType != "" {
		query := `
			SELECT id, event_type, tenant_id, app_id, backend_id, request_id, user_id, data, timestamp, created_at
			FROM monitoring_events
			WHERE event_type = $1 AND tenant_id = $2 AND timestamp >= $3
			ORDER BY timestamp DESC
			LIMIT $4`
		rows, err = r.db.QueryContext(ctx, query, eventType, tenantID, since, limit)
	} else if tenantID != nil {
		query := `
			SELECT id, event_type, tenant_id, app_id, backend_id, request_id, user_id, data, timestamp, created_at
			FROM monitoring_events
			WHERE tenant_id = $1 AND timestamp >= $2
			ORDER BY timestamp DESC
			LIMIT $3`
		rows, err = r.db.QueryContext(ctx, query, tenantID, since, limit)
	} else if eventType != "" {
		query := `
			SELECT id, event_type, tenant_id, app_id, backend_id, request_id, user_id, data, timestamp, created_at
			FROM monitoring_events
			WHERE event_type = $1 AND timestamp >= $2
			ORDER BY timestamp DESC
			LIMIT $3`
		rows, err = r.db.QueryContext(ctx, query, eventType, since, limit)
	} else {
		query := `
			SELECT id, event_type, tenant_id, app_id, backend_id, request_id, user_id, data, timestamp, created_at
			FROM monitoring_events
			WHERE timestamp >= $1
			ORDER BY timestamp DESC
			LIMIT $2`
		rows, err = r.db.QueryContext(ctx, query, since, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*MonitoringEvent
	for rows.Next() {
		event := &MonitoringEvent{}
		var dataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.TenantID,
			&event.AppID,
			&event.BackendID,
			&event.RequestID,
			&event.UserID,
			&dataJSON,
			&event.Timestamp,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal data JSON
		if len(dataJSON) > 0 {
			err = json.Unmarshal(dataJSON, &event.Data)
			if err != nil {
				return nil, err
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// UpdateAlertStatus updates an alert's status
func (r *MonitoringRepository) UpdateAlertStatus(ctx context.Context, alert *Alert) error {
	query := `
		UPDATE alerts
		SET status = $1, resolved_at = $2, resolved_by = $3, updated_at = NOW()
		WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, alert.Status, alert.ResolvedAt, alert.ResolvedBy, alert.ID)
	return err
}

// QueryPerformanceMetrics retrieves performance metrics with filtering
func (r *MonitoringRepository) QueryPerformanceMetrics(ctx context.Context, metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*PerformanceMetric, error) {
	var rows *sql.Rows
	var err error

	if tenantID != nil {
		query := `
			SELECT id, metric_type, tenant_id, app_id, backend_id, value, unit, labels, timestamp, created_at
			FROM performance_metrics
			WHERE metric_type = $1 AND tenant_id = $2 AND timestamp >= $3
			ORDER BY timestamp DESC
			LIMIT $4`
		rows, err = r.db.QueryContext(ctx, query, metricType, tenantID, since, limit)
	} else {
		query := `
			SELECT id, metric_type, tenant_id, app_id, backend_id, value, unit, labels, timestamp, created_at
			FROM performance_metrics
			WHERE metric_type = $1 AND timestamp >= $2
			ORDER BY timestamp DESC
			LIMIT $3`
		rows, err = r.db.QueryContext(ctx, query, metricType, since, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*PerformanceMetric
	for rows.Next() {
		metric := &PerformanceMetric{}
		var labelsJSON []byte

		err := rows.Scan(
			&metric.ID,
			&metric.MetricType,
			&metric.TenantID,
			&metric.AppID,
			&metric.BackendID,
			&metric.Value,
			&metric.Unit,
			&labelsJSON,
			&metric.Timestamp,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal labels JSON
		if len(labelsJSON) > 0 {
			err = json.Unmarshal(labelsJSON, &metric.Labels)
			if err != nil {
				return nil, err
			}
		}

		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// QueryActiveAlerts retrieves currently active alerts
func (r *MonitoringRepository) QueryActiveAlerts(tenantID *uuid.UUID) ([]*Alert, error) {
	var rows *sql.Rows
	var err error

	if tenantID != nil {
		query := `
			SELECT id, alert_type, severity, tenant_id, app_id, backend_id, title, message,
				   status, resolved_at, resolved_by, metadata, created_at, updated_at
			FROM alerts
			WHERE status = 'active' AND tenant_id = $1
			ORDER BY created_at DESC`
		rows, err = r.db.Query(query, tenantID)
	} else {
		query := `
			SELECT id, alert_type, severity, tenant_id, app_id, backend_id, title, message,
				   status, resolved_at, resolved_by, metadata, created_at, updated_at
			FROM alerts
			WHERE status = 'active'
			ORDER BY created_at DESC`
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*Alert
	for rows.Next() {
		alert := &Alert{}
		var metadataJSON []byte

		err := rows.Scan(
			&alert.ID,
			&alert.AlertType,
			&alert.Severity,
			&alert.TenantID,
			&alert.AppID,
			&alert.BackendID,
			&alert.Title,
			&alert.Message,
			&alert.Status,
			&alert.ResolvedAt,
			&alert.ResolvedBy,
			&metadataJSON,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &alert.Metadata)
			if err != nil {
				return nil, err
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// QueryLatestSystemHealthChecks retrieves the latest system health checks for each component
func (r *MonitoringRepository) QueryLatestSystemHealthChecks() (map[string]*SystemHealthCheck, error) {
	query := `
		SELECT DISTINCT ON (component_name) id, check_type, component_name, status,
			   response_time_ms, message, metadata, checked_at, created_at
		FROM system_health_checks
		ORDER BY component_name, checked_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	healthChecks := make(map[string]*SystemHealthCheck)
	for rows.Next() {
		check := &SystemHealthCheck{}
		var metadataJSON []byte

		err := rows.Scan(
			&check.ID,
			&check.CheckType,
			&check.ComponentName,
			&check.Status,
			&check.ResponseTimeMs,
			&check.Message,
			&metadataJSON,
			&check.CheckedAt,
			&check.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal metadata JSON
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &check.Metadata)
			if err != nil {
				return nil, err
			}
		}

		healthChecks[check.ComponentName] = check
	}

	return healthChecks, rows.Err()
}

// PgNotify sends a PostgreSQL notification using pg_notify
func (r *MonitoringRepository) PgNotify(channel, payload string) error {
	query := "SELECT pg_notify($1, $2)"
	_, err := r.db.Exec(query, channel, payload)
	return err
}

// PgListen starts listening on a PostgreSQL channel for notifications
func (r *MonitoringRepository) PgListen(ctx context.Context, channel string) error {
	query := "LISTEN " + channel
	_, err := r.db.Exec(query)
	return err
}

// PgWaitForNotification waits for a notification on listened channels
// Note: This is a simplified implementation. For production use, consider using pgx driver
// which has better support for asynchronous notifications.
func (r *MonitoringRepository) PgWaitForNotification(ctx context.Context) (*PgNotification, error) {
	// This implementation uses polling since the standard database/sql driver
	// doesn't support asynchronous notifications well.
	// For production, consider switching to pgx.

	// Check if there are pending notifications
	row := r.db.QueryRowContext(ctx, "SELECT pg_notification_queue_usage()")
	var queueUsage float64
	if err := row.Scan(&queueUsage); err != nil {
		return nil, fmt.Errorf("failed to check notification queue: %w", err)
	}

	// If queue has notifications, get the next one
	if queueUsage > 0 {
		row := r.db.QueryRowContext(ctx, `
			SELECT pg_notify_pid, pg_notify_channel, pg_notify_payload
			FROM pg_notification_queue_get()
			LIMIT 1
		`)
		var pid int
		var channel, payload string
		if err := row.Scan(&pid, &channel, &payload); err != nil {
			return nil, fmt.Errorf("failed to get notification: %w", err)
		}

		return &PgNotification{
			PID:     pid,
			Channel: channel,
			Payload: payload,
		}, nil
	}

	return nil, nil // No notifications available
}


// StoreDatabaseMetrics stores database metrics for historical analysis
func (r *MonitoringRepository) StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error {
	now := time.Now()

	// Store connection metrics
	if connData, ok := metrics["connections"].(map[string]interface{}); ok {
		if total, ok := connData["total"].(float64); ok {
			_, err := r.db.ExecContext(ctx, `
				INSERT INTO database_metrics (metric_type, value, unit, metadata, recorded_at)
				VALUES ($1, $2, $3, $4, $5)
			`, "connections", total, "count", map[string]interface{}{
				"active": connData["active"],
				"idle":   connData["idle"],
			}, now)
			if err != nil {
				return fmt.Errorf("failed to store connection metrics: %w", err)
			}
		}
	}

	// Store database size
	if storageData, ok := metrics["storage"].(map[string]interface{}); ok {
		if sizeGB, ok := storageData["usedGB"].(float64); ok {
			_, err := r.db.ExecContext(ctx, `
				INSERT INTO database_metrics (metric_type, value, unit, recorded_at)
				VALUES ($1, $2, $3, $4)
			`, "size_gb", sizeGB, "gb", now)
			if err != nil {
				return fmt.Errorf("failed to store database size metric: %w", err)
			}
		}
	}

	// Store performance metrics
	if perfData, ok := metrics["performance"].(map[string]interface{}); ok {
		if cacheRatio, ok := perfData["cacheHitRatio"].(float64); ok {
			_, err := r.db.ExecContext(ctx, `
				INSERT INTO database_metrics (metric_type, value, unit, recorded_at)
				VALUES ($1, $2, $3, $4)
			`, "cache_hit_ratio", cacheRatio, "ratio", now)
			if err != nil {
				return fmt.Errorf("failed to store cache hit ratio metric: %w", err)
			}
		}

		if avgTime, ok := perfData["avgQueryTime"].(float64); ok {
			_, err := r.db.ExecContext(ctx, `
				INSERT INTO database_metrics (metric_type, value, unit, recorded_at)
				VALUES ($1, $2, $3, $4)
			`, "query_time", avgTime, "ms", now)
			if err != nil {
				return fmt.Errorf("failed to store query time metric: %w", err)
			}
		}

		if throughput, ok := perfData["throughput"].(float64); ok {
			_, err := r.db.ExecContext(ctx, `
				INSERT INTO database_metrics (metric_type, value, unit, recorded_at)
				VALUES ($1, $2, $3, $4)
			`, "throughput", throughput, "qps", now)
			if err != nil {
				return fmt.Errorf("failed to store throughput metric: %w", err)
			}
		}
	}

	return nil
}

// QueryDatabaseMetrics retrieves historical database metrics
func (r *MonitoringRepository) QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*DatabaseMetric, error) {
	query := `
		SELECT id, metric_type, value, unit, metadata, recorded_at, created_at
		FROM database_metrics
		WHERE metric_type = $1 AND recorded_at >= $2
		ORDER BY recorded_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, metricType, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query database metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*DatabaseMetric
	for rows.Next() {
		var metric DatabaseMetric
		var metadataBytes []byte

		err := rows.Scan(
			&metric.ID,
			&metric.MetricType,
			&metric.Value,
			&metric.Unit,
			&metadataBytes,
			&metric.RecordedAt,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database metric: %w", err)
		}

		// Parse metadata JSON
		if len(metadataBytes) > 0 {
			err = json.Unmarshal(metadataBytes, &metric.Metadata)
			if err != nil {
				// Log error but continue - metadata is optional
				metric.Metadata = make(map[string]interface{})
			}
		} else {
			metric.Metadata = make(map[string]interface{})
		}

		metrics = append(metrics, &metric)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating database metrics: %w", err)
	}

	return metrics, nil
}

// GetDatabaseHealthMetrics returns comprehensive database health and performance metrics
func (r *MonitoringRepository) GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get connection stats
	stats := r.db.Stats()
	metrics["connections"] = map[string]interface{}{
		"active": stats.InUse,
		"idle":   stats.Idle,
		"total":  stats.OpenConnections,
		"max":    stats.MaxOpenConnections,
	}

	// Query for database size and growth
	var dbSizeGB float64
	var dbSizeGrowth float64
	err := r.db.QueryRowContext(ctx, `
		SELECT
			pg_size_pretty(pg_database_size(current_database())) as size,
			CASE
				WHEN pg_database_size(current_database()) > 0
				THEN pg_database_size(current_database())::float / (1024*1024*1024)
				ELSE 0
			END as size_gb
	`).Scan(&dbSizeGB, &dbSizeGB)
	if err != nil {
		// Failed to get database size, use 0
		dbSizeGB = 0
	}

	// Calculate growth rate based on historical data
	dbSizeGrowth = 0
	if dbSizeGB > 0 {
		// Get database size from 7 days ago
		weekAgo := time.Now().AddDate(0, 0, -7)
		historicalSizes, err := r.QueryDatabaseMetrics(ctx, "size_gb", weekAgo, 1)
		if err == nil && len(historicalSizes) > 0 {
			oldSize := historicalSizes[0].Value
			if oldSize > 0 {
				dbSizeGrowth = ((dbSizeGB - oldSize) / oldSize) * 100 // Growth rate as percentage
			}
		}
	}

	// Get active connections count
	var activeConnCount int
	err = r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM pg_stat_activity
		WHERE state = 'active' AND pid <> pg_backend_pid()
	`).Scan(&activeConnCount)
	if err != nil {
		// Failed to get active connection count, use 0
		activeConnCount = 0
	}

	// Get slow queries count (queries taking more than 1 second in the last hour)
	var slowQueriesCount int
	err = r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM pg_stat_activity
		WHERE state = 'active'
		AND now() - query_start > interval '1 second'
		AND pid <> pg_backend_pid()
	`).Scan(&slowQueriesCount)
	if err != nil {
		// Failed to get slow queries count, use 0
		slowQueriesCount = 0
	}

	// Get cache hit ratio
	var cacheHitRatio float64
	err = r.db.QueryRowContext(ctx, `
		SELECT
			CASE
				WHEN blks_hit + blks_read > 0
				THEN (blks_hit::float / (blks_hit + blks_read))
				ELSE 0
			END as cache_hit_ratio
		FROM pg_stat_database
		WHERE datname = current_database()
	`).Scan(&cacheHitRatio)
	if err != nil {
		// Failed to get cache hit ratio, use 0
		cacheHitRatio = 0
	}

	// Estimate throughput (queries per second in last minute)
	var queriesPerSecond float64
	err = r.db.QueryRowContext(ctx, `
		SELECT
			CASE
				WHEN extract(epoch from (now() - stats_reset)) > 0
				THEN (xact_commit + xact_rollback)::float / extract(epoch from (now() - stats_reset))
				ELSE 0
			END as qps
		FROM pg_stat_database
		WHERE datname = current_database()
	`).Scan(&queriesPerSecond)
	if err != nil {
		// Failed to get queries per second, use 0
		queriesPerSecond = 0
	}

	// Get average query time from pg_stat_statements if available
	var avgQueryTime float64
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(avg(total_time / calls), 0) as avg_query_time_ms
		FROM pg_stat_statements
		WHERE calls > 0
		LIMIT 100
	`).Scan(&avgQueryTime)
	if err != nil {
		// pg_stat_statements might not be available, use a fallback
		avgQueryTime = 0
	}

	// Check replication status and lag (for primary-replica setups)
	var replicationLag int
	var replicationStatus string
	var isReplica bool

	// Check if this is a replica instance
	err = r.db.QueryRowContext(ctx, `SELECT pg_is_in_recovery()`).Scan(&isReplica)
	if err != nil {
		// If we can't determine, assume it's not a replica
		isReplica = false
	}

	if isReplica {
		// This is a replica, check replication lag
		err = r.db.QueryRowContext(ctx, `
			SELECT COALESCE(extract(epoch from now() - pg_last_xact_replay_timestamp())::int, 0) as lag_seconds
		`).Scan(&replicationLag)
		if err != nil {
			// Failed to get replication lag
			replicationLag = -1
			replicationStatus = "error"
		} else {
			// Determine status based on lag
			if replicationLag == -1 {
				replicationStatus = "error"
			} else if replicationLag > 300 { // More than 5 minutes lag
				replicationStatus = "critical"
			} else if replicationLag > 60 { // More than 1 minute lag
				replicationStatus = "warning"
			} else {
				replicationStatus = "healthy"
			}
		}
	} else {
		// This is a primary instance
		replicationLag = 0
		replicationStatus = "primary"
	}

	// Calculate available disk space percentage
	var availablePercent float64
	err = r.db.QueryRowContext(ctx, `
		SELECT
			CASE
				WHEN total_bytes > 0
				THEN ((total_bytes - used_bytes)::float / total_bytes) * 100
				ELSE 100.0
			END as available_percent
		FROM (
			SELECT
				pg_tablespace_size('pg_default') as used_bytes,
				CASE
					WHEN pg_tablespace_location('pg_default') != ''
					THEN (SELECT (pg_stat_file(pg_tablespace_location('pg_default') || '/PG_VERSION')).size)
					ELSE NULL
				END as total_bytes
		) disk_stats
	`).Scan(&availablePercent)
	if err != nil {
		// Fallback: try PostgreSQL system statistics approach
		// For production deployments, external monitoring is recommended:
		// - Prometheus node_exporter: Direct filesystem metrics
		// - CloudWatch/GCP Monitoring/Azure Monitor: Cloud-native monitoring
		// - Nagios/Icinga: Traditional infrastructure monitoring
		err = r.db.QueryRowContext(ctx, `
			-- Use PostgreSQL system statistics and tablespace analysis for better estimation
			WITH system_disk_stats AS (
				-- Get tablespace locations and sizes for more accurate filesystem analysis
				SELECT
					spcname as tablespace_name,
					pg_tablespace_location(oid) as location,
					pg_tablespace_size(oid) as size_bytes
				FROM pg_tablespace
				WHERE pg_tablespace_location(oid) IS NOT NULL
				AND pg_tablespace_location(oid) != ''
			),
			database_metrics AS (
				-- Calculate database size and growth patterns
				SELECT
					current_database() as db_name,
					pg_database_size(current_database()) as current_size,
					-- Use more sophisticated space estimation based on PostgreSQL best practices
					-- Large databases (>10GB) typically need 3-5x current size for operations
					-- Medium databases (1-10GB) typically need 5-10x current size
					-- Small databases (<1GB) typically need 10-20x current size
					CASE
						WHEN pg_database_size(current_database()) > 10737418240 THEN
							-- >10GB databases: more conservative estimate
							pg_database_size(current_database()) * 4
						WHEN pg_database_size(current_database()) > 1073741824 THEN
							-- 1-10GB databases: moderate estimate
							pg_database_size(current_database()) * 7
						WHEN pg_database_size(current_database()) > 104857600 THEN
							-- 100MB-1GB databases: higher estimate
							pg_database_size(current_database()) * 12
						ELSE
							-- <100MB databases: generous estimate for growth
							pg_database_size(current_database()) * 25
					END as estimated_capacity
			)
			-- Calculate available percentage with safeguards
			SELECT
				CASE
					WHEN estimated_capacity > 0 AND current_size <= estimated_capacity
					THEN ((estimated_capacity - current_size)::float / estimated_capacity) * 100
					WHEN estimated_capacity > 0 AND current_size > estimated_capacity
					THEN 0.0  -- Database is over capacity estimate
					ELSE 85.0  -- Default fallback percentage
				END as available_percent
			FROM database_metrics
		`).Scan(&availablePercent)
		if err != nil {
			// If all methods fail, use default
			availablePercent = 100.0
		}
	}

	// Determine overall status
	status := "healthy"
	if activeConnCount > stats.MaxOpenConnections/2 {
		status = "warning"
	}
	if stats.OpenConnections >= stats.MaxOpenConnections {
		status = "critical"
	}

	metrics["status"] = status
	metrics["performance"] = map[string]interface{}{
		"avgQueryTime":  avgQueryTime,
		"slowQueries":   slowQueriesCount,
		"throughput":    queriesPerSecond,
		"cacheHitRatio": cacheHitRatio,
	}
	metrics["storage"] = map[string]interface{}{
		"usedGB":           dbSizeGB,
		"growthRate":       dbSizeGrowth,
		"availablePercent": availablePercent,
	}
	metrics["replication"] = map[string]interface{}{
		"lag":    replicationLag,
		"status": replicationStatus,
	}

	return metrics, nil
}
