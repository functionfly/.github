package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RecordCostAllocationEntry records a detailed cost allocation entry for an execution
func (r *BillingRepository) RecordCostAllocationEntry(ctx context.Context, entry *CostAllocationEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now().UTC()

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

	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.TenantID, entry.FunctionID, entry.FunctionName, entry.FunctionAuthor,
		entry.ExecutionID, entry.ExecutionOutcome, entry.Cached,
		entry.DurationMs, entry.CPUTimeMs, entry.MemoryUsedMB, entry.WallTimeMs,
		entry.ExecutionCostCents, entry.ComputeCostCents, entry.PlatformFeeCents,
		entry.DataTransferCents, entry.TotalCostCents,
		entry.Region, entry.Timestamp, entry.PeriodStart, entry.PeriodEnd,
		entry.Tags, entry.Metadata, entry.CreatedAt,
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
			SUM(execution_cost_cents) as execution_cost_cents,
			SUM(compute_cost_cents) as compute_cost_cents,
			SUM(platform_fee_cents) as platform_fee_cents,
			SUM(data_transfer_cents) as data_transfer_cents,
			SUM(total_cost_cents) as total_cost_cents
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
		whereClause += fmt.Sprintf(" AND function_name ILIKE $%d", argIdx)
		args = append(args, fmt.Sprintf("%%%s%%", *filter.FunctionName))
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
		ReportID:       uuid.New(),
		GeneratedAt:    time.Now().UTC(),
		PeriodStart:    start,
		PeriodEnd:      end,
		TenantCount:    len(summaries),
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
