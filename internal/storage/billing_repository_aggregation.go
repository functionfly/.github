package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

)

// AggregatedBillingUsage represents aggregated billing usage data

// AggregateExecutionsForBilling aggregates function executions for billing over a time period
func (r *BillingRepository) AggregateExecutionsForBilling(ctx context.Context, start, end time.Time) ([]*AggregatedBillingUsage, error) {
	query := `
		SELECT
			f.tenant_id,
			e.function_id,
			f.name as function_name,
			f.author,
			COUNT(*) as total_calls,
			SUM(CASE WHEN e.outcome = 'success' THEN 1 ELSE 0 END) as success_calls,
			SUM(CASE WHEN e.outcome = 'error' THEN 1 ELSE 0 END) as error_calls,
			SUM(CASE WHEN e.cached = true THEN 1 ELSE 0 END) as cached_calls,
			SUM(e.duration_ms) as total_duration,
			COALESCE(AVG(e.duration_ms), 0) as avg_duration,
			COALESCE(SUM(r.cpu_time_used_ms), 0) as total_cpu_time_ms,
			COALESCE(SUM(r.memory_used_mb), 0) as total_memory_mb
		FROM registry_function_executions e
		JOIN registry_functions f ON e.function_id = f.id
		LEFT JOIN execution_resource_usage r ON e.id = r.execution_id
		WHERE e.timestamp > $1 AND e.timestamp <= $2
			AND f.tenant_id IS NOT NULL
		GROUP BY f.tenant_id, e.function_id, f.name, f.author
	`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate executions: %w", err)
	}
	defer rows.Close()

	var results []*AggregatedBillingUsage
	for rows.Next() {
		usage := &AggregatedBillingUsage{}
		err := rows.Scan(
			&usage.TenantID,
			&usage.FunctionID,
			&usage.FunctionName,
			&usage.Author,
			&usage.TotalCalls,
			&usage.SuccessCalls,
			&usage.ErrorCalls,
			&usage.CachedCalls,
			&usage.TotalDuration,
			&usage.AvgDuration,
			&usage.TotalCPUTimeMs,
			&usage.TotalMemoryMB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregated usage: %w", err)
		}
		results = append(results, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating aggregated results: %w", err)
	}

	return results, nil
}

// GetLastAggregationTimestamp gets the last timestamp that was aggregated
func (r *BillingRepository) GetLastAggregationTimestamp(ctx context.Context) (time.Time, error) {
	// First try to get from aggregation_state table
	var lastTimestamp sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT last_processed_timestamp FROM aggregation_state WHERE id = 'execution_aggregation'
	`).Scan(&lastTimestamp)

	if err == sql.ErrNoRows {
		// Fallback to max timestamp from usage_events
		err = r.db.QueryRowContext(ctx, `
			SELECT MAX(timestamp) FROM usage_events
		`).Scan(&lastTimestamp)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to get last aggregation timestamp: %w", err)
		}
	} else if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last aggregation timestamp: %w", err)
	}

	if lastTimestamp.Valid {
		return lastTimestamp.Time, nil
	}

	// If no usage events exist, default to 1 hour ago to process recent data
	return time.Now().UTC().Add(-1 * time.Hour), nil
}

// SetLastAggregationTimestamp updates the last aggregation timestamp
func (r *BillingRepository) SetLastAggregationTimestamp(ctx context.Context, timestamp time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO aggregation_state (id, last_processed_timestamp, updated_at)
		VALUES ('execution_aggregation', $1, NOW())
		ON CONFLICT (id) DO UPDATE SET
			last_processed_timestamp = EXCLUDED.last_processed_timestamp,
			updated_at = NOW(),
			processed_count = aggregation_state.processed_count + 1
	`, timestamp)

	if err != nil {
		return fmt.Errorf("failed to set last aggregation timestamp: %w", err)
	}

	return nil
}

// GetLastRollupDate gets the last date that was rolled up
func (r *BillingRepository) GetLastRollupDate(ctx context.Context) (time.Time, error) {
	// First try to get from aggregation_state table
	var lastDate sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT last_processed_timestamp FROM aggregation_state WHERE id = 'rollup_aggregation'
	`).Scan(&lastDate)

	if err == sql.ErrNoRows {
		// Fallback to max period_date from usage_rollups
		err = r.db.QueryRowContext(ctx, `
			SELECT MAX(period_date) FROM usage_rollups
		`).Scan(&lastDate)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to get last rollup date: %w", err)
		}
	} else if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last rollup date: %w", err)
	}

	if lastDate.Valid {
		return lastDate.Time, nil
	}

	// If no rollups exist, default to 2 days ago
	return time.Now().UTC().Add(-48 * time.Hour).Truncate(24 * time.Hour), nil
}

// SetLastRollupDate updates the last rollup date
func (r *BillingRepository) SetLastRollupDate(ctx context.Context, date time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO aggregation_state (id, last_processed_timestamp, updated_at)
		VALUES ('rollup_aggregation', $1, NOW())
		ON CONFLICT (id) DO UPDATE SET
			last_processed_timestamp = EXCLUDED.last_processed_timestamp,
			updated_at = NOW(),
			processed_count = aggregation_state.processed_count + 1
	`, date)

	if err != nil {
		return fmt.Errorf("failed to set last rollup date: %w", err)
	}

	return nil
}
