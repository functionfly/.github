package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AppAnalyticsRepository handles app analytics queries against routing_events
// and related tables.
type AppAnalyticsRepository struct {
	db *PostgresDB
}

// NewAppAnalyticsRepository creates a new app analytics repository.
func NewAppAnalyticsRepository(db *PostgresDB) *AppAnalyticsRepository {
	return &AppAnalyticsRepository{db: db}
}

// GetAppAnalyticsSummary returns aggregated stats for an app since the given time.
func (r *AppAnalyticsRepository) GetAppAnalyticsSummary(ctx context.Context, appID uuid.UUID, since time.Time) (*AppAnalyticsSummary, error) {
	var s AppAnalyticsSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(COUNT(*), 0) AS total_requests,
			COALESCE(AVG(latency_ms), 0) AS avg_latency_ms,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms)::INTEGER, 0) AS p95_latency_ms,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::INTEGER, 0) AS p99_latency_ms,
			COALESCE(SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END)::FLOAT / NULLIF(COUNT(*), 0), 0) AS success_rate,
			COALESCE(SUM(CASE WHEN outcome != 'success' THEN 1 ELSE 0 END)::FLOAT / NULLIF(COUNT(*), 0), 0) AS error_rate
		FROM routing_events
		WHERE app_id = $1 AND timestamp >= $2
	`, appID, since).Scan(
		&s.TotalRequests, &s.AvgLatencyMs, &s.P95LatencyMs, &s.P99LatencyMs,
		&s.SuccessRate, &s.ErrorRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get app analytics summary: %w", err)
	}

	// Get function execution totals (via functions.app_id → registry_function_executions)
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(COUNT(*), 0)
		FROM registry_function_executions rfe
		JOIN registry_functions rf ON rf.id = rfe.function_id
		WHERE rf.app_id = $1 AND rfe.timestamp >= $2
	`, appID, since).Scan(&s.TotalExecutions)
	if err != nil {
		s.TotalExecutions = 0
	}

	// Get cost totals (via functions.app_id → cost_allocation_entries)
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries cae
		JOIN registry_functions rf ON rf.id = cae.function_id
		WHERE rf.app_id = $1 AND cae.timestamp >= $2
	`, appID, since).Scan(&s.TotalCostCents)
	if err != nil {
		s.TotalCostCents = 0
	}

	return &s, nil
}

// GetAppRequestTimeseries returns time-bucketed request counts.
// interval should be "1h", "1d", etc.
func (r *AppAnalyticsRepository) GetAppRequestTimeseries(ctx context.Context, appID uuid.UUID, since time.Time, interval string) ([]*AppRequestTimeseriesPoint, error) {
	bucket := "1 hour"
	switch interval {
	case "1d":
		bucket = "1 day"
	case "1h":
		bucket = "1 hour"
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			date_trunc($3, timestamp) AS bucket,
			COUNT(*) AS total,
			SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) AS success,
			SUM(CASE WHEN outcome != 'success' THEN 1 ELSE 0 END) AS errors
		FROM routing_events
		WHERE app_id = $1 AND timestamp >= $2
		GROUP BY bucket
		ORDER BY bucket ASC
	`), appID, since, bucketToTrunc(bucket))
	if err != nil {
		return nil, fmt.Errorf("failed to get request timeseries: %w", err)
	}
	defer rows.Close()

	var points []*AppRequestTimeseriesPoint
	for rows.Next() {
		p := &AppRequestTimeseriesPoint{}
		if err := rows.Scan(&p.Timestamp, &p.Total, &p.Success, &p.Errors); err != nil {
			return nil, fmt.Errorf("failed to scan request timeseries: %w", err)
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetAppLatencyTimeseries returns time-bucketed latency percentiles.
func (r *AppAnalyticsRepository) GetAppLatencyTimeseries(ctx context.Context, appID uuid.UUID, since time.Time, interval string) ([]*AppLatencyTimeseriesPoint, error) {
	bucket := "1 hour"
	switch interval {
	case "1d":
		bucket = "1 day"
	case "1h":
		bucket = "1 hour"
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			date_trunc($3, timestamp) AS bucket,
			COALESCE(AVG(latency_ms), 0) AS avg_ms,
			COALESCE(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY latency_ms)::INTEGER, 0) AS p50_ms,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms)::INTEGER, 0) AS p95_ms,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::INTEGER, 0) AS p99_ms
		FROM routing_events
		WHERE app_id = $1 AND timestamp >= $2
		GROUP BY bucket
		ORDER BY bucket ASC
	`), appID, since, bucketToTrunc(bucket))
	if err != nil {
		return nil, fmt.Errorf("failed to get latency timeseries: %w", err)
	}
	defer rows.Close()

	var points []*AppLatencyTimeseriesPoint
	for rows.Next() {
		p := &AppLatencyTimeseriesPoint{}
		if err := rows.Scan(&p.Timestamp, &p.AvgMs, &p.P50Ms, &p.P95Ms, &p.P99Ms); err != nil {
			return nil, fmt.Errorf("failed to scan latency timeseries: %w", err)
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetAppTopErrors returns error breakdown by HTTP status code.
// Since routing_events doesn't store status codes directly, we use outcome as a proxy.
func (r *AppAnalyticsRepository) GetAppTopErrors(ctx context.Context, appID uuid.UUID, since time.Time) ([]*AppErrorBreakdown, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			CASE outcome
				WHEN 'error' THEN 500
				WHEN 'timeout' THEN 504
				WHEN 'circuit_open' THEN 503
				ELSE 400
			END AS status_code,
			COUNT(*) AS cnt
		FROM routing_events
		WHERE app_id = $1 AND timestamp >= $2 AND outcome != 'success'
		GROUP BY status_code
		ORDER BY cnt DESC
		LIMIT 10
	`, appID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get top errors: %w", err)
	}
	defer rows.Close()

	var errors []*AppErrorBreakdown
	for rows.Next() {
		e := &AppErrorBreakdown{}
		if err := rows.Scan(&e.StatusCode, &e.Count); err != nil {
			return nil, fmt.Errorf("failed to scan top errors: %w", err)
		}
		errors = append(errors, e)
	}
	return errors, rows.Err()
}

// GetAppBackendBreakdown returns per-backend aggregated stats.
func (r *AppAnalyticsRepository) GetAppBackendBreakdown(ctx context.Context, appID uuid.UUID, since time.Time) ([]*AppBackendBreakdown, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			re.backend_id::TEXT,
			COALESCE(b.provider, 'unknown') AS provider,
			COUNT(*) AS requests,
			COALESCE(AVG(re.latency_ms), 0) AS avg_latency_ms,
			COALESCE(SUM(CASE WHEN re.outcome != 'success' THEN 1 ELSE 0 END)::FLOAT / NULLIF(COUNT(*), 0), 0) AS error_rate
		FROM routing_events re
		LEFT JOIN backends b ON b.id = re.backend_id
		WHERE re.app_id = $1 AND re.timestamp >= $2
		GROUP BY re.backend_id, b.provider
		ORDER BY requests DESC
	`, appID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []*AppBackendBreakdown
	for rows.Next() {
		b := &AppBackendBreakdown{}
		if err := rows.Scan(&b.BackendID, &b.Provider, &b.Requests, &b.AvgLatencyMs, &b.ErrorRate); err != nil {
			return nil, fmt.Errorf("failed to scan backend breakdown: %w", err)
		}
		breakdown = append(breakdown, b)
	}
	return breakdown, rows.Err()
}

// bucketToTrunc maps a human-readable interval to a PostgreSQL date_trunc precision.
func bucketToTrunc(interval string) string {
	switch interval {
	case "1 day", "1d":
		return "day"
	case "1 hour", "1h":
		return "hour"
	default:
		return "hour"
	}
}
