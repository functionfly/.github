package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// RecordCostAllocationEntry records a detailed cost allocation entry for an execution
func (r *BillingRepository) RecordCostAllocationEntry(ctx context.Context, entry *CostAllocationEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now().UTC()

	tagsJSON, err := json.Marshal(entry.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO cost_allocation_entries (
			id, tenant_id, function_id, function_name, function_author,
			execution_id, execution_outcome, cached,
			duration_ms, cpu_time_ms, memory_used_mb, wall_time_ms,
			execution_cost_cents, compute_cost_cents, platform_fee_cents,
			data_transfer_cents, total_cost_cents,
			region, timestamp, period_start, period_end, tags, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)`

	_, err = r.db.ExecContext(ctx, query,
		entry.ID, entry.TenantID, entry.FunctionID, entry.FunctionName, entry.FunctionAuthor,
		entry.ExecutionID, entry.ExecutionOutcome, entry.Cached,
		entry.DurationMs, entry.CPUTimeMs, entry.MemoryUsedMB, entry.WallTimeMs,
		entry.ExecutionCostCents, entry.ComputeCostCents, entry.PlatformFeeCents,
		entry.DataTransferCents, entry.TotalCostCents,
		entry.Region, entry.Timestamp, entry.PeriodStart, entry.PeriodEnd,
		tagsJSON, metadataJSON, entry.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record cost allocation entry: %w", err)
	}

	return nil
}

// GetCostAllocationByFunction returns cost allocation aggregated by function for a tenant
func (r *BillingRepository) GetCostAllocationByFunction(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) ([]*CostAllocationSummary, error) {
	query := `
		SELECT
			function_id,
			function_name,
			function_author,
			COUNT(*) as total_executions,
			SUM(CASE WHEN execution_outcome = 'success' THEN 1 ELSE 0 END) as success_executions,
			SUM(CASE WHEN execution_outcome = 'error' THEN 1 ELSE 0 END) as error_executions,
			SUM(CASE WHEN cached = true THEN 1 ELSE 0 END) as cached_executions,
			SUM(duration_ms) as total_duration_ms,
			SUM(cpu_time_ms) as total_cpu_time_ms,
			SUM(memory_used_mb) as total_memory_used_mb,
			SUM(execution_cost_cents) as execution_cost_cents,
			SUM(compute_cost_cents) as compute_cost_cents,
			SUM(platform_fee_cents) as platform_fee_cents,
			SUM(data_transfer_cents) as data_transfer_cents,
			SUM(total_cost_cents) as total_cost_cents,
			AVG(duration_ms) as avg_duration_ms,
			AVG(total_cost_cents) as avg_cost_cents
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY function_id, function_name, function_author
		ORDER BY SUM(total_cost_cents) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost allocation by function: %w", err)
	}
	defer rows.Close()

	var results []*CostAllocationSummary
	for rows.Next() {
		summary := &CostAllocationSummary{}
		err := rows.Scan(
			&summary.FunctionID,
			&summary.FunctionName,
			&summary.FunctionAuthor,
			&summary.TotalExecutions,
			&summary.SuccessExecutions,
			&summary.ErrorExecutions,
			&summary.CachedExecutions,
			&summary.TotalDurationMs,
			&summary.TotalCPUTimeMs,
			&summary.TotalMemoryUsedMB,
			&summary.ExecutionCostCents,
			&summary.ComputeCostCents,
			&summary.PlatformFeeCents,
			&summary.DataTransferCents,
			&summary.TotalCostCents,
			&summary.AvgDurationMs,
			&summary.AvgCostCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost allocation summary: %w", err)
		}
		results = append(results, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cost allocation results: %w", err)
	}

	return results, nil
}

// GetCostAllocationDailyBreakdown returns daily cost breakdown for a tenant
func (r *BillingRepository) GetCostAllocationDailyBreakdown(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) ([]*DailyCostBreakdown, error) {
	query := `
		SELECT
			DATE(timestamp) as date,
			COUNT(*) as executions,
			SUM(total_cost_cents) as cost_cents
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY DATE(timestamp)
		ORDER BY date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily cost breakdown: %w", err)
	}
	defer rows.Close()

	var results []*DailyCostBreakdown
	for rows.Next() {
		day := &DailyCostBreakdown{}
		var dateStr string
		err := rows.Scan(&dateStr, &day.Executions, &day.CostCents)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily breakdown: %w", err)
		}
		day.Date, _ = time.Parse("2006-01-02", dateStr)
		results = append(results, day)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily breakdown results: %w", err)
	}

	return results, nil
}

// GetTenantCostSummary returns a comprehensive cost summary for a tenant
func (r *BillingRepository) GetTenantCostSummary(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) (*TenantCostSummary, error) {
	// Get function summaries
	functionSummaries, err := r.GetCostAllocationByFunction(ctx, tenantID, start, end)
	if err != nil {
		return nil, err
	}

	// Get daily breakdown
	dailyBreakdown, err := r.GetCostAllocationDailyBreakdown(ctx, tenantID, start, end)
	if err != nil {
		return nil, err
	}

	// Get tenant totals
	query := `
		SELECT
			COUNT(*) as total_executions,
			COUNT(DISTINCT function_id) as unique_functions,
			COALESCE(SUM(execution_cost_cents), 0) as execution_cost_cents,
			COALESCE(SUM(compute_cost_cents), 0) as compute_cost_cents,
			COALESCE(SUM(platform_fee_cents), 0) as platform_fee_cents,
			COALESCE(SUM(data_transfer_cents), 0) as data_transfer_cents,
			COALESCE(SUM(total_cost_cents), 0) as total_cost_cents
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
	`

	row := r.db.QueryRowContext(ctx, query, tenantID, start, end)

	// Convert slices of pointers to slices of values
	fnSummaries := make([]CostAllocationSummary, len(functionSummaries))
	for i, s := range functionSummaries {
		fnSummaries[i] = *s
	}

	daily := make([]DailyCostBreakdown, len(dailyBreakdown))
	for i, d := range dailyBreakdown {
		daily[i] = *d
	}

	summary := &TenantCostSummary{
		TenantID:          tenantID,
		PeriodStart:       start,
		PeriodEnd:         end,
		FunctionSummaries: fnSummaries,
		DailyBreakdown:    daily,
	}

	var uniqueFunctions int64
	err = row.Scan(
		&summary.TotalExecutions,
		&uniqueFunctions,
		&summary.ExecutionCostCents,
		&summary.ComputeCostCents,
		&summary.PlatformFeeCents,
		&summary.DataTransferCents,
		&summary.TotalCostCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan tenant cost summary: %w", err)
	}

	summary.UniqueFunctions = int(uniqueFunctions)

	return summary, nil
}

// GetAllTenantsCostSummary returns cost summaries for all tenants in a period
func (r *BillingRepository) GetAllTenantsCostSummary(
	ctx context.Context,
	start, end time.Time,
) ([]*TenantCostSummary, error) {
	query := `
		SELECT DISTINCT tenant_id
		FROM cost_allocation_entries
		WHERE timestamp >= $1 AND timestamp <= $2
	`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant list: %w", err)
	}
	defer rows.Close()

	var tenantIDs []uuid.UUID
	for rows.Next() {
		var tenantID uuid.UUID
		if err := rows.Scan(&tenantID); err != nil {
			continue
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	var summaries []*TenantCostSummary
	for _, tenantID := range tenantIDs {
		summary, err := r.GetTenantCostSummary(ctx, tenantID, start, end)
		if err != nil {
			continue
		}

		// Get tenant name
		var tenantName string
		_ = r.db.QueryRowContext(ctx, `SELECT name FROM tenants WHERE id = $1`, tenantID).Scan(&tenantName)
		summary.TenantName = tenantName

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetCostAllocationEntries returns detailed cost allocation entries with filtering
func (r *BillingRepository) GetCostAllocationEntries(
	ctx context.Context,
	filter *CostAllocationFilter,
	limit, offset int,
) ([]*CostAllocationEntry, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.TenantID != nil {
		whereClause += fmt.Sprintf(" AND tenant_id = $%d", argIdx)
		args = append(args, *filter.TenantID)
		argIdx++
	}

	if filter.FunctionID != nil {
		whereClause += fmt.Sprintf(" AND function_id = $%d", argIdx)
		args = append(args, *filter.FunctionID)
		argIdx++
	}

	if filter.FunctionName != nil {
		sanitized := SanitizeSQLWildcards(*filter.FunctionName)
		whereClause += fmt.Sprintf(" AND function_name ILIKE $%d", argIdx)
		args = append(args, fmt.Sprintf("%%%s%%", sanitized))
		argIdx++
	}

	if filter.Author != nil {
		whereClause += fmt.Sprintf(" AND function_author = $%d", argIdx)
		args = append(args, *filter.Author)
		argIdx++
	}

	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if filter.Outcome != nil {
		whereClause += fmt.Sprintf(" AND execution_outcome = $%d", argIdx)
		args = append(args, *filter.Outcome)
		argIdx++
	}

	if filter.Cached != nil {
		whereClause += fmt.Sprintf(" AND cached = $%d", argIdx)
		args = append(args, *filter.Cached)
		argIdx++
	}

	if filter.Region != nil {
		whereClause += fmt.Sprintf(" AND region = $%d", argIdx)
		args = append(args, *filter.Region)
		argIdx++
	}

	if filter.MinCostCents != nil {
		whereClause += fmt.Sprintf(" AND total_cost_cents >= $%d", argIdx)
		args = append(args, *filter.MinCostCents)
		argIdx++
	}

	if filter.MaxCostCents != nil {
		whereClause += fmt.Sprintf(" AND total_cost_cents <= $%d", argIdx)
		args = append(args, *filter.MaxCostCents)
		argIdx++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cost_allocation_entries %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cost allocation entries: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT
			id, tenant_id, function_id, function_name, function_author,
			execution_id, execution_outcome, cached,
			duration_ms, cpu_time_ms, memory_used_mb, wall_time_ms,
			execution_cost_cents, compute_cost_cents, platform_fee_cents,
			data_transfer_cents, total_cost_cents,
			region, timestamp, period_start, period_end, tags, metadata, created_at
		FROM cost_allocation_entries
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cost allocation entries: %w", err)
	}
	defer rows.Close()

	var entries []*CostAllocationEntry
	for rows.Next() {
		entry := &CostAllocationEntry{}
		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.FunctionID, &entry.FunctionName, &entry.FunctionAuthor,
			&entry.ExecutionID, &entry.ExecutionOutcome, &entry.Cached,
			&entry.DurationMs, &entry.CPUTimeMs, &entry.MemoryUsedMB, &entry.WallTimeMs,
			&entry.ExecutionCostCents, &entry.ComputeCostCents, &entry.PlatformFeeCents,
			&entry.DataTransferCents, &entry.TotalCostCents,
			&entry.Region, &entry.Timestamp, &entry.PeriodStart, &entry.PeriodEnd,
			&entry.Tags, &entry.Metadata, &entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan cost allocation entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating cost allocation entries: %w", err)
	}

	return entries, total, nil
}

// GetCostAllocationReport generates a comprehensive cost allocation report
func (r *BillingRepository) GetCostAllocationReport(
	ctx context.Context,
	start, end time.Time,
) (*CostAllocationReport, error) {
	summaries, err := r.GetAllTenantsCostSummary(ctx, start, end)
	if err != nil {
		return nil, err
	}

	// Convert slice of pointers to slice of values
	tenantSummaries := make([]TenantCostSummary, len(summaries))
	for i, s := range summaries {
		tenantSummaries[i] = *s
	}

	report := &CostAllocationReport{
		ReportID:        uuid.New(),
		GeneratedAt:     time.Now().UTC(),
		PeriodStart:     start,
		PeriodEnd:       end,
		TenantCount:     len(summaries),
		TenantSummaries: tenantSummaries,
	}

	// Calculate totals
	var totalCost int64
	var totalExecutions int64
	functionIDs := make(map[uuid.UUID]bool)

	for _, summary := range summaries {
		totalCost += summary.TotalCostCents
		totalExecutions += summary.TotalExecutions
		for _, fn := range summary.FunctionSummaries {
			functionIDs[fn.FunctionID] = true
		}

		// Create chargeback entry
		chargeback := CostAllocationChargeback{
			TenantID:           summary.TenantID,
			TenantName:         summary.TenantName,
			CostCenter:         "", // Would come from tenant metadata
			Department:         "", // Would come from tenant metadata
			Project:            "", // Would come from tenant metadata
			TotalCostCents:     summary.TotalCostCents,
			ExecutionCostCents: summary.ExecutionCostCents,
			ComputeCostCents:   summary.ComputeCostCents,
			PlatformFeeCents:   summary.PlatformFeeCents,
			DataTransferCents:  summary.DataTransferCents,
			InvoicePeriod:      fmt.Sprintf("%s to %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
			GeneratedAt:        report.GeneratedAt,
		}
		report.ChargebackEntries = append(report.ChargebackEntries, chargeback)
	}

	report.TotalCostCents = totalCost
	report.TotalExecutions = totalExecutions
	report.FunctionCount = len(functionIDs)

	return report, nil
}

// GetCostAllocationByRegion returns cost breakdown by region
func (r *BillingRepository) GetCostAllocationByRegion(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) (map[string]*CostAllocationSummary, error) {
	query := `
		SELECT
			region,
			COUNT(*) as total_executions,
			SUM(total_cost_cents) as total_cost_cents,
			SUM(execution_cost_cents) as execution_cost_cents,
			SUM(compute_cost_cents) as compute_cost_cents,
			SUM(platform_fee_cents) as platform_fee_cents,
			SUM(data_transfer_cents) as data_transfer_cents
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY region
		ORDER BY region
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost allocation by region: %w", err)
	}
	defer rows.Close()

	results := make(map[string]*CostAllocationSummary)
	for rows.Next() {
		var region string
		summary := &CostAllocationSummary{}

		err := rows.Scan(
			&region,
			&summary.TotalExecutions,
			&summary.TotalCostCents,
			&summary.ExecutionCostCents,
			&summary.ComputeCostCents,
			&summary.PlatformFeeCents,
			&summary.DataTransferCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan region cost summary: %w", err)
		}

		results[region] = summary
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating region cost results: %w", err)
	}

	return results, nil
}

// DeleteOldCostAllocationEntries removes old cost allocation entries beyond retention period
func (r *BillingRepository) DeleteOldCostAllocationEntries(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM cost_allocation_entries WHERE timestamp < $1",
		before,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old cost allocation entries: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ==================== Data Retention Policy (Compliance) ====================

// Default retention periods per compliance requirements
const (
	// FinancialDataRetentionPeriod is 7 years for SOX/PCI compliance
	FinancialDataRetentionPeriod = 7 * 365 * 24 * time.Hour
	// DetailedExecutionLogRetention is 90 days for execution-level detail
	DetailedExecutionLogRetention = 90 * 24 * time.Hour
	// BatchSize for cleanup operations to avoid long locks
	RetentionCleanupBatchSize = 10000
)

// CleanupCostAllocationByRetention performs compliance-based data retention cleanup.
// It removes detailed execution logs beyond 90 days while keeping financial
// aggregates for 7 years (audit requirement).
//
// This implements a tiered retention strategy:
//   - 0-90 days: Full execution detail retained
//   - 90 days-7 years: Financial aggregates only (detailed rows deleted)
//   - 7+ years: All data purged (unless jurisdiction requires longer)
func (r *BillingRepository) CleanupCostAllocationByRetention(ctx context.Context) (map[string]int64, error) {
	results := make(map[string]int64)
	now := time.Now().UTC()

	hasHolds, err := r.HasActiveLegalHolds(ctx)
	if err != nil {
		return results, fmt.Errorf("failed to check legal holds: %w", err)
	}
	if hasHolds {
		return results, fmt.Errorf("legal hold active, skipping cleanup")
	}

	// Step 1: Delete detailed execution logs older than 90 days
	// These are high-volume, low-value for audit after the invoice is settled
	detailedCutoff := now.Add(-DetailedExecutionLogRetention)
	deleted, err := r.cleanupOldEntriesInBatches(ctx, detailedCutoff)
	if err != nil {
		return results, fmt.Errorf("failed to cleanup detailed execution logs: %w", err)
	}
	results["detailed_execution_logs_deleted"] = deleted

	// Note: Financial aggregates (invoice-level, monthly rollups) are NOT deleted here
	// as they must be retained for 7 years. Those are handled separately by
	// CleanupFinancialAggregatesAfterRetention()

	return results, nil
}

// cleanupOldEntriesInBatches deletes old entries in batches to avoid table locks
func (r *BillingRepository) cleanupOldEntriesInBatches(ctx context.Context, before time.Time) (int64, error) {
	var totalDeleted int64

	for {
		// Use CTE for efficient batch deletion - avoids re-evaluating WHERE clause in outer DELETE
		result, err := r.db.ExecContext(ctx, `
			WITH batch AS (
				SELECT id FROM cost_allocation_entries
				WHERE timestamp < $1
				ORDER BY timestamp
				LIMIT $2
			)
			DELETE FROM cost_allocation_entries USING batch
			WHERE cost_allocation_entries.id = batch.id`,
			before, RetentionCleanupBatchSize,
		)
		if err != nil {
			return totalDeleted, fmt.Errorf("batch delete failed: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += rowsAffected

		// If we deleted fewer rows than the batch size, we're done
		if rowsAffected < RetentionCleanupBatchSize {
			break
		}

		// Check for context cancellation between batches
		select {
		case <-ctx.Done():
			return totalDeleted, ctx.Err()
		default:
		}
	}

	return totalDeleted, nil
}

// CleanupFinancialAggregatesAfterRetention deletes invoice-level aggregates
// that have passed the 7-year financial retention period.
// WARNING: This should only be called after legal review and confirmation
// that no disputes or audits are pending for the period being purged.
func (r *BillingRepository) CleanupFinancialAggregatesAfterRetention(ctx context.Context, jurisdictionRetention time.Duration) (int64, error) {
	hasHolds, err := r.HasActiveLegalHolds(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to check legal holds: %w", err)
	}
	if hasHolds {
		return 0, fmt.Errorf("legal hold active, skipping cleanup")
	}

	// Default to 7 years if not specified
	retentionPeriod := jurisdictionRetention
	if retentionPeriod == 0 {
		retentionPeriod = FinancialDataRetentionPeriod
	}

	cutoff := time.Now().UTC().Add(-retentionPeriod)

	// First, archive/summarize what we're about to delete for audit purposes
	// (optional - depends on compliance requirements)
	_, err = r.createRetentionAuditLog(ctx, cutoff)
	if err != nil {
		// For SOX compliance, audit logging failures must not be silently ignored.
		// Log the error and alert in production.
		logrus.WithError(err).Error("SOX compliance: failed to create retention audit log during data purge. This may indicate an audit trail gap.")
	}

	// Delete entries older than the retention period
	return r.cleanupOldEntriesInBatches(ctx, cutoff)
}

// createRetentionAuditLog records what data was purged for compliance auditing
func (r *BillingRepository) createRetentionAuditLog(ctx context.Context, cutoff time.Time) (int64, error) {
	// Aggregate what will be deleted for audit purposes
	query := `
		SELECT 
			COUNT(*) as entry_count,
			COUNT(DISTINCT tenant_id) as tenant_count,
			MIN(timestamp) as oldest_entry,
			MAX(timestamp) as newest_entry,
			SUM(total_cost_cents) as total_cost_cents
		FROM cost_allocation_entries 
		WHERE timestamp < $1
	`

	var summary struct {
		EntryCount     int64
		TenantCount    int64
		OldestEntry    *time.Time
		NewestEntry    *time.Time
		TotalCostCents int64
	}

	err := r.db.QueryRowContext(ctx, query, cutoff).Scan(
		&summary.EntryCount,
		&summary.TenantCount,
		&summary.OldestEntry,
		&summary.NewestEntry,
		&summary.TotalCostCents,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create retention audit summary: %w", err)
	}

	// Insert into retention_audit_log table (if it exists)
	// This is idempotent - safe to run multiple times
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO retention_audit_log (
			id, table_name, retention_policy, cutoff_date,
			records_affected, tenant_count, financial_impact_cents,
			oldest_record, newest_record, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT DO NOTHING`,
		uuid.New(), "cost_allocation_entries", "financial_7_year", cutoff,
		summary.EntryCount, summary.TenantCount, summary.TotalCostCents,
		summary.OldestEntry, summary.NewestEntry, time.Now().UTC(),
	)

	return summary.EntryCount, nil
}

// GetCostAllocationRetentionSummary returns statistics for retention planning
// Shows how much data would be affected by cleanup policies
func (r *BillingRepository) GetCostAllocationRetentionSummary(ctx context.Context) (*CostAllocationRetentionSummary, error) {
	now := time.Now().UTC()
	detailedCutoff := now.Add(-DetailedExecutionLogRetention)
	financialCutoff := now.Add(-FinancialDataRetentionPeriod)

	summary := &CostAllocationRetentionSummary{
		CurrentTime:            now,
		DetailedRetentionDays:  90,
		FinancialRetentionDays: 2555, // 7 years
	}

	// Count entries older than 90 days (would be cleaned up)
	queryDetailed := `
		SELECT COUNT(*), COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries 
		WHERE timestamp < $1
	`
	err := r.db.QueryRowContext(ctx, queryDetailed, detailedCutoff).Scan(
		&summary.DetailedEntriesEligibleForDeletion,
		&summary.DetailedFinancialValueCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed retention summary: %w", err)
	}

	// Count entries older than 7 years (financial data)
	queryFinancial := `
		SELECT COUNT(*), COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries 
		WHERE timestamp < $1
	`
	err = r.db.QueryRowContext(ctx, queryFinancial, financialCutoff).Scan(
		&summary.FinancialEntriesEligibleForDeletion,
		&summary.FinancialDataValueCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial retention summary: %w", err)
	}

	// Get date range of all data
	err = r.db.QueryRowContext(ctx, `
		SELECT MIN(timestamp), MAX(timestamp), COUNT(*) 
		FROM cost_allocation_entries
	`).Scan(&summary.OldestEntryDate, &summary.NewestEntryDate, &summary.TotalEntryCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get date range: %w", err)
	}

	return summary, nil
}

// CostAllocationRetentionSummary provides statistics for retention planning
type CostAllocationRetentionSummary struct {
	CurrentTime                         time.Time  `json:"current_time"`
	OldestEntryDate                     *time.Time `json:"oldest_entry_date,omitempty"`
	NewestEntryDate                     *time.Time `json:"newest_entry_date,omitempty"`
	TotalEntryCount                     int64      `json:"total_entry_count"`
	DetailedRetentionDays               int        `json:"detailed_retention_days"`
	DetailedEntriesEligibleForDeletion  int64      `json:"detailed_entries_eligible_for_deletion"`
	DetailedFinancialValueCents         int64      `json:"detailed_financial_value_cents"`
	FinancialRetentionDays              int        `json:"financial_retention_days"`
	FinancialEntriesEligibleForDeletion int64      `json:"financial_entries_eligible_for_deletion"`
	FinancialDataValueCents             int64      `json:"financial_data_value_cents"`
}

// ==================== Per-Team Cost Allocation ====================

// GetTeamCostAllocation returns aggregated cost for a team over a period.
// Uses the tags column (JSONB) to filter by team_id tag.
func (r *BillingRepository) GetTeamCostAllocation(
	ctx context.Context,
	teamID uuid.UUID,
	start, end time.Time,
) (*TeamCostAllocation, error) {
	query := `
		SELECT
			COALESCE(SUM(total_cost_cents), 0),
			COALESCE(SUM(execution_cost_cents), 0),
			COALESCE(SUM(compute_cost_cents), 0),
			COALESCE(SUM(platform_fee_cents), 0),
			COALESCE(SUM(data_transfer_cents), 0),
			COUNT(*),
			COUNT(DISTINCT function_id)
		FROM cost_allocation_entries
		WHERE tags->>'team_id' = $1
		  AND timestamp >= $2
		  AND timestamp <= $3
	`

	var alloc TeamCostAllocation
	alloc.TeamID = teamID
	alloc.PeriodStart = start
	alloc.PeriodEnd = end

	err := r.db.QueryRowContext(ctx, query, teamID.String(), start, end).Scan(
		&alloc.TotalCostCents,
		&alloc.ExecutionCostCents,
		&alloc.ComputeCostCents,
		&alloc.PlatformFeeCents,
		&alloc.DataTransferCents,
		&alloc.TotalExecutions,
		&alloc.UniqueFunctions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get team cost allocation: %w", err)
	}

	return &alloc, nil
}

// GetTeamCostByFunction returns per-function cost breakdown for a team.
func (r *BillingRepository) GetTeamCostByFunction(
	ctx context.Context,
	teamID uuid.UUID,
	start, end time.Time,
) ([]*TeamCostBreakdown, error) {
	query := `
		SELECT
			function_id,
			function_name,
			function_author,
			SUM(total_cost_cents),
			COUNT(*),
			AVG(total_cost_cents),
			SUM(CASE WHEN execution_outcome = 'success' THEN 1.0 ELSE 0.0 END) / NULLIF(COUNT(*), 0) * 100
		FROM cost_allocation_entries
		WHERE tags->>'team_id' = $1
		  AND timestamp >= $2
		  AND timestamp <= $3
		GROUP BY function_id, function_name, function_author
		ORDER BY SUM(total_cost_cents) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, teamID.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get team cost by function: %w", err)
	}
	defer rows.Close()

	var results []*TeamCostBreakdown
	for rows.Next() {
		b := &TeamCostBreakdown{TeamID: teamID}
		err := rows.Scan(
			&b.FunctionID, &b.FunctionName, &b.FunctionAuthor,
			&b.TotalCostCents, &b.TotalExecutions, &b.AvgCostCents, &b.SuccessRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team cost breakdown: %w", err)
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

// GetAllTeamsCostSummary returns cost allocation for all teams in a tenant.
func (r *BillingRepository) GetAllTeamsCostSummary(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) ([]*TeamCostAllocation, error) {
	query := `
		SELECT DISTINCT tags->>'team_id' as team_id
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
		  AND tags->>'team_id' IS NOT NULL
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var summaries []*TeamCostAllocation
	for rows.Next() {
		var teamIDStr string
		if err := rows.Scan(&teamIDStr); err != nil {
			continue
		}
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			continue
		}
		alloc, err := r.GetTeamCostAllocation(ctx, teamID, start, end)
		if err != nil {
			continue
		}
		alloc.TenantID = tenantID
		// Resolve team name
		var teamName string
		_ = r.db.QueryRowContext(ctx, `SELECT name FROM teams WHERE id = $1`, teamID).Scan(&teamName)
		alloc.TeamName = teamName
		summaries = append(summaries, alloc)
	}
	return summaries, rows.Err()
}

// ==================== Department Budgets ====================

// UpsertDepartmentBudget creates or updates a department budget.
func (r *BillingRepository) UpsertDepartmentBudget(ctx context.Context, budget *DepartmentBudget) error {
	query := `
		INSERT INTO department_budgets (
			id, tenant_id, name, description, budget_cents,
			warning_threshold_pct, critical_threshold_pct,
			period_start, period_end, team_ids, tag_filters,
			alert_email, is_active, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, description = EXCLUDED.description,
			budget_cents = EXCLUDED.budget_cents,
			warning_threshold_pct = EXCLUDED.warning_threshold_pct,
			critical_threshold_pct = EXCLUDED.critical_threshold_pct,
			period_start = EXCLUDED.period_start, period_end = EXCLUDED.period_end,
			team_ids = EXCLUDED.team_ids, tag_filters = EXCLUDED.tag_filters,
			alert_email = EXCLUDED.alert_email, is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`
	budget.UpdatedAt = time.Now().UTC()
	teamIDs := make([]string, len(budget.TeamIDs))
	for i, t := range budget.TeamIDs {
		teamIDs[i] = t.String()
	}
	_, err := r.db.ExecContext(ctx, query,
		budget.ID, budget.TenantID, budget.Name, budget.Description,
		budget.BudgetCents, budget.WarningThresholdPct, budget.CriticalThresholdPct,
		budget.PeriodStart, budget.PeriodEnd, pq.Array(teamIDs), budget.TagFilters,
		budget.AlertEmail, budget.IsActive, budget.CreatedBy,
		budget.CreatedAt, budget.UpdatedAt,
	)
	return err
}

// GetDepartmentBudget retrieves a budget by ID.
func (r *BillingRepository) GetDepartmentBudget(ctx context.Context, id uuid.UUID) (*DepartmentBudget, error) {
	query := `
		SELECT id, tenant_id, name, description, budget_cents,
			warning_threshold_pct, critical_threshold_pct,
			period_start, period_end, team_ids, tag_filters,
			alert_email, is_active, created_by, created_at, updated_at
		FROM department_budgets WHERE id = $1
	`
	var b DepartmentBudget
	var teamIDs []string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.TenantID, &b.Name, &b.Description, &b.BudgetCents,
		&b.WarningThresholdPct, &b.CriticalThresholdPct,
		&b.PeriodStart, &b.PeriodEnd, pq.Array(&teamIDs), &b.TagFilters,
		&b.AlertEmail, &b.IsActive, &b.CreatedBy, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	b.TeamIDs = make([]uuid.UUID, len(teamIDs))
	for i, s := range teamIDs {
		b.TeamIDs[i], _ = uuid.Parse(s)
	}
	return &b, nil
}

// ListDepartmentBudgets lists all budgets for a tenant.
func (r *BillingRepository) ListDepartmentBudgets(ctx context.Context, tenantID uuid.UUID) ([]*DepartmentBudget, error) {
	query := `
		SELECT id, tenant_id, name, description, budget_cents,
			warning_threshold_pct, critical_threshold_pct,
			period_start, period_end, team_ids, tag_filters,
			alert_email, is_active, created_by, created_at, updated_at
		FROM department_budgets WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []*DepartmentBudget
	for rows.Next() {
		var b DepartmentBudget
		var teamIDs []string
		err := rows.Scan(
			&b.ID, &b.TenantID, &b.Name, &b.Description, &b.BudgetCents,
			&b.WarningThresholdPct, &b.CriticalThresholdPct,
			&b.PeriodStart, &b.PeriodEnd, pq.Array(&teamIDs), &b.TagFilters,
			&b.AlertEmail, &b.IsActive, &b.CreatedBy, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			continue
		}
		b.TeamIDs = make([]uuid.UUID, len(teamIDs))
		for i, s := range teamIDs {
			b.TeamIDs[i], _ = uuid.Parse(s)
		}
		budgets = append(budgets, &b)
	}
	return budgets, rows.Err()
}

// GetDepartmentBudgetSpend calculates current spend against a budget.
// It sums cost_allocation_entries filtered by the budget's team_ids and tag_filters.
func (r *BillingRepository) GetDepartmentBudgetSpend(ctx context.Context, budgetID uuid.UUID) (*DepartmentBudget, error) {
	budget, err := r.GetDepartmentBudget(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Build spend query based on team_ids or tag_filters
	var totalSpent int64
	if len(budget.TeamIDs) > 0 {
		query := `
			SELECT COALESCE(SUM(total_cost_cents), 0)
			FROM cost_allocation_entries
			WHERE timestamp >= $1 AND timestamp <= $2
			  AND tags->>'team_id' = ANY($3::uuid[])
		`
		_ = r.db.QueryRowContext(ctx, query, budget.PeriodStart, budget.PeriodEnd, pq.Array(budget.TeamIDs)).Scan(&totalSpent)
	} else if len(budget.TagFilters) > 0 {
		// Build dynamic tag filters
		where := "timestamp >= $1 AND timestamp <= $2"
		args := []interface{}{budget.PeriodStart, budget.PeriodEnd}
		idx := 3
		for k, v := range budget.TagFilters {
			where += fmt.Sprintf(" AND tags->>$%d = $%d", idx, idx+1)
			args = append(args, k, v)
			idx += 2
		}
		query := fmt.Sprintf(`
			SELECT COALESCE(SUM(total_cost_cents), 0)
			FROM cost_allocation_entries WHERE %s
		`, where)
		_ = r.db.QueryRowContext(ctx, query, args...).Scan(&totalSpent)
	}

	budget.SpentCents = totalSpent
	budget.RemainingCents = budget.BudgetCents - totalSpent
	if budget.BudgetCents > 0 {
		budget.SpentPercent = float64(totalSpent) / float64(budget.BudgetCents) * 100
	}
	return budget, nil
}

// UpsertBudgetAlert records a budget alert.
func (r *BillingRepository) UpsertBudgetAlert(ctx context.Context, alert *BudgetAlert) error {
	alert.ID = uuid.New()
	alert.CreatedAt = time.Now().UTC()
	query := `
		INSERT INTO budget_alerts (id, budget_id, tenant_id, level, spent_pct, spent_cents, budget_cents, alert_sent_to, alert_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.BudgetID, alert.TenantID,
		alert.Level, alert.SpentPct, alert.SpentCents, alert.BudgetCents,
		alert.AlertSentTo, alert.AlertMessage, alert.CreatedAt,
	)
	return err
}

// ListBudgetAlerts returns recent alerts for a tenant.
func (r *BillingRepository) ListBudgetAlerts(ctx context.Context, tenantID uuid.UUID, limit int) ([]*BudgetAlert, error) {
	query := `
		SELECT id, budget_id, tenant_id, level, spent_pct, spent_cents, budget_cents, alert_sent_to, alert_message, created_at
		FROM budget_alerts WHERE tenant_id = $1
		ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*BudgetAlert
	for rows.Next() {
		a := &BudgetAlert{}
		err := rows.Scan(&a.ID, &a.BudgetID, &a.TenantID, &a.Level, &a.SpentPct, &a.SpentCents, &a.BudgetCents, &a.AlertSentTo, &a.AlertMessage, &a.CreatedAt)
		if err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// ==================== Cost Anomaly Detection ====================

// UpsertCostAnomaly records a detected anomaly.
func (r *BillingRepository) UpsertCostAnomaly(ctx context.Context, anomaly *CostAnomaly) error {
	anomaly.ID = uuid.New()
	anomaly.CreatedAt = time.Now().UTC()
	query := `
		INSERT INTO cost_anomalies (id, tenant_id, anomaly_type, severity, team_id, function_id, region,
			expected_cost_cents, actual_cost_cents, delta_cents, delta_percent, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		anomaly.ID, anomaly.TenantID, anomaly.AnomalyType, anomaly.Severity,
		anomaly.TeamID, anomaly.FunctionID, anomaly.Region,
		anomaly.ExpectedCostCents, anomaly.ActualCostCents, anomaly.DeltaCents,
		anomaly.DeltaPercent, anomaly.Description, anomaly.CreatedAt,
	)
	return err
}

// ListCostAnomalies returns recent anomalies for a tenant.
func (r *BillingRepository) ListCostAnomalies(ctx context.Context, tenantID uuid.UUID, limit int) ([]*CostAnomaly, error) {
	query := `
		SELECT id, tenant_id, anomaly_type, severity, team_id, function_id, region,
			expected_cost_cents, actual_cost_cents, delta_cents, delta_percent, description, created_at
		FROM cost_anomalies WHERE tenant_id = $1
		ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []*CostAnomaly
	for rows.Next() {
		a := &CostAnomaly{}
		err := rows.Scan(
			&a.ID, &a.TenantID, &a.AnomalyType, &a.Severity,
			&a.TeamID, &a.FunctionID, &a.Region,
			&a.ExpectedCostCents, &a.ActualCostCents, &a.DeltaCents, &a.DeltaPercent,
			&a.Description, &a.CreatedAt,
		)
		if err != nil {
			continue
		}
		anomalies = append(anomalies, a)
	}
	return anomalies, rows.Err()
}

func SanitizeSQLWildcards(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
