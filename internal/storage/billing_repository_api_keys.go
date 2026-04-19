package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== Per-API-Key Cost Tracking ====================

// RecordAPIKeyCostEntry records a cost allocation entry with API key attribution
func (r *BillingRepository) RecordAPIKeyCostEntry(ctx context.Context, entry *CostAllocationEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now().UTC()

	query := `
		INSERT INTO cost_allocation_entries (
			id, tenant_id, api_key_id, function_id, function_name, function_author,
			execution_id, execution_outcome, cached,
			duration_ms, cpu_time_ms, memory_used_mb, wall_time_ms,
			execution_cost_cents, compute_cost_cents, platform_fee_cents,
			data_transfer_cents, total_cost_cents,
			region, timestamp, period_start, period_end, tags, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
		)`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.TenantID, entry.APIKeyID, entry.FunctionID, entry.FunctionName, entry.FunctionAuthor,
		entry.ExecutionID, entry.ExecutionOutcome, entry.Cached,
		entry.DurationMs, entry.CPUTimeMs, entry.MemoryUsedMB, entry.WallTimeMs,
		entry.ExecutionCostCents, entry.ComputeCostCents, entry.PlatformFeeCents,
		entry.DataTransferCents, entry.TotalCostCents,
		entry.Region, entry.Timestamp, entry.PeriodStart, entry.PeriodEnd,
		entry.Tags, entry.Metadata, entry.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record API key cost entry: %w", err)
	}

	return nil
}

// GetAPIKeyCostSummary returns aggregated cost for a specific API key over a period
func (r *BillingRepository) GetAPIKeyCostSummary(
	ctx context.Context,
	apiKeyID uuid.UUID,
	start, end time.Time,
) (*APIKeyCostSummary, error) {
	query := `
		SELECT
			api_key_id,
			COUNT(*) as total_executions,
			COALESCE(SUM(execution_cost_cents), 0) as execution_cost_cents,
			COALESCE(SUM(compute_cost_cents), 0) as compute_cost_cents,
			COALESCE(SUM(platform_fee_cents), 0) as platform_fee_cents,
			COALESCE(SUM(data_transfer_cents), 0) as data_transfer_cents,
			COALESCE(SUM(total_cost_cents), 0) as total_cost_cents
		FROM cost_allocation_entries
		WHERE api_key_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY api_key_id
	`

	summary := &APIKeyCostSummary{
		APIKeyID:    apiKeyID,
		PeriodStart: start,
		PeriodEnd:   end,
	}

	err := r.db.QueryRowContext(ctx, query, apiKeyID, start, end).Scan(
		&summary.APIKeyID,
		&summary.TotalExecutions,
		&summary.ExecutionCostCents,
		&summary.ComputeCostCents,
		&summary.PlatformFeeCents,
		&summary.DataTransferCents,
		&summary.TotalCostCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key cost summary: %w", err)
	}

	// Get API key name
	var keyName string
	_ = r.db.QueryRowContext(ctx, `SELECT name FROM api_keys WHERE id = $1`, apiKeyID).Scan(&keyName)
	summary.APIKeyName = keyName

	return summary, nil
}

// GetTenantAPIKeyCosts returns cost summaries for all API keys of a tenant
func (r *BillingRepository) GetTenantAPIKeyCosts(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) ([]*APIKeyCostSummary, error) {
	query := `
		SELECT
			api_key_id,
			COUNT(*) as total_executions,
			COALESCE(SUM(execution_cost_cents), 0) as execution_cost_cents,
			COALESCE(SUM(compute_cost_cents), 0) as compute_cost_cents,
			COALESCE(SUM(platform_fee_cents), 0) as platform_fee_cents,
			COALESCE(SUM(data_transfer_cents), 0) as data_transfer_cents,
			COALESCE(SUM(total_cost_cents), 0) as total_cost_cents
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3 AND api_key_id IS NOT NULL
		GROUP BY api_key_id
		ORDER BY SUM(total_cost_cents) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant API key costs: %w", err)
	}
	defer rows.Close()

	var results []*APIKeyCostSummary
	for rows.Next() {
		summary := &APIKeyCostSummary{
			PeriodStart: start,
			PeriodEnd:   end,
		}
		err := rows.Scan(
			&summary.APIKeyID,
			&summary.TotalExecutions,
			&summary.ExecutionCostCents,
			&summary.ComputeCostCents,
			&summary.PlatformFeeCents,
			&summary.DataTransferCents,
			&summary.TotalCostCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key cost summary: %w", err)
		}

		// Get API key name
		var keyName string
		_ = r.db.QueryRowContext(ctx, `SELECT name FROM api_keys WHERE id = $1`, summary.APIKeyID).Scan(&keyName)
		summary.APIKeyName = keyName

		results = append(results, summary)
	}

	return results, rows.Err()
}

// GetHighValueKeyCosts returns costs for API keys marked as high-value
func (r *BillingRepository) GetHighValueKeyCosts(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) ([]*APIKeyCostSummary, error) {
	query := `
		SELECT
			cae.api_key_id,
			COUNT(*) as total_executions,
			COALESCE(SUM(cae.execution_cost_cents), 0) as execution_cost_cents,
			COALESCE(SUM(cae.compute_cost_cents), 0) as compute_cost_cents,
			COALESCE(SUM(cae.platform_fee_cents), 0) as platform_fee_cents,
			COALESCE(SUM(cae.data_transfer_cents), 0) as data_transfer_cents,
			COALESCE(SUM(cae.total_cost_cents), 0) as total_cost_cents
		FROM cost_allocation_entries cae
		INNER JOIN api_keys ak ON ak.id = cae.api_key_id
		WHERE cae.tenant_id = $1 AND cae.timestamp >= $2 AND cae.timestamp <= $3
		  AND ak.is_high_value = true
		GROUP BY cae.api_key_id
		ORDER BY SUM(cae.total_cost_cents) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get high-value key costs: %w", err)
	}
	defer rows.Close()

	var results []*APIKeyCostSummary
	for rows.Next() {
		summary := &APIKeyCostSummary{
			PeriodStart: start,
			PeriodEnd:   end,
		}
		err := rows.Scan(
			&summary.APIKeyID,
			&summary.TotalExecutions,
			&summary.ExecutionCostCents,
			&summary.ComputeCostCents,
			&summary.PlatformFeeCents,
			&summary.DataTransferCents,
			&summary.TotalCostCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan high-value key cost: %w", err)
		}

		// Get API key name
		var keyName string
		_ = r.db.QueryRowContext(ctx, `SELECT name FROM api_keys WHERE id = $1`, summary.APIKeyID).Scan(&keyName)
		summary.APIKeyName = keyName

		results = append(results, summary)
	}

	return results, rows.Err()
}

// ==================== API Key Budgets ====================

// UpsertAPIKeyBudget creates or updates an API key budget
func (r *BillingRepository) UpsertAPIKeyBudget(ctx context.Context, budget *APIKeyBudget) error {
	query := `
		INSERT INTO api_key_budgets (
			id, api_key_id, tenant_id, budget_cents,
			warning_threshold_pct, critical_threshold_pct,
			period_start, period_end, alert_email, disable_at_limit,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (api_key_id) DO UPDATE SET
			budget_cents = EXCLUDED.budget_cents,
			warning_threshold_pct = EXCLUDED.warning_threshold_pct,
			critical_threshold_pct = EXCLUDED.critical_threshold_pct,
			period_start = EXCLUDED.period_start,
			period_end = EXCLUDED.period_end,
			alert_email = EXCLUDED.alert_email,
			disable_at_limit = EXCLUDED.disable_at_limit,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`

	if budget.ID == uuid.Nil {
		budget.ID = uuid.New()
	}
	budget.UpdatedAt = time.Now().UTC()
	if budget.CreatedAt.IsZero() {
		budget.CreatedAt = budget.UpdatedAt
	}

	_, err := r.db.ExecContext(ctx, query,
		budget.ID, budget.APIKeyID, budget.TenantID, budget.BudgetCents,
		budget.WarningThresholdPct, budget.CriticalThresholdPct,
		budget.PeriodStart, budget.PeriodEnd, budget.AlertEmail, budget.DisableAtLimit,
		budget.IsActive, budget.CreatedAt, budget.UpdatedAt,
	)

	return err
}

// GetAPIKeyBudget retrieves a budget for a specific API key
func (r *BillingRepository) GetAPIKeyBudget(ctx context.Context, apiKeyID uuid.UUID) (*APIKeyBudget, error) {
	query := `
		SELECT id, api_key_id, tenant_id, budget_cents,
			warning_threshold_pct, critical_threshold_pct,
			period_start, period_end, alert_email, disable_at_limit,
			is_active, created_at, updated_at
		FROM api_key_budgets
		WHERE api_key_id = $1 AND is_active = true
	`

	var b APIKeyBudget
	err := r.db.QueryRowContext(ctx, query, apiKeyID).Scan(
		&b.ID, &b.APIKeyID, &b.TenantID, &b.BudgetCents,
		&b.WarningThresholdPct, &b.CriticalThresholdPct,
		&b.PeriodStart, &b.PeriodEnd, &b.AlertEmail, &b.DisableAtLimit,
		&b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// GetAPIKeyBudgetWithSpend calculates current spend against an API key budget
func (r *BillingRepository) GetAPIKeyBudgetWithSpend(ctx context.Context, apiKeyID uuid.UUID) (*APIKeyBudget, error) {
	budget, err := r.GetAPIKeyBudget(ctx, apiKeyID)
	if err != nil {
		return nil, err
	}

	// Calculate spend for the budget period
	var totalSpent int64
	spentQuery := `
		SELECT COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries
		WHERE api_key_id = $1
		  AND timestamp >= $2
		  AND timestamp <= $3
	`
	err = r.db.QueryRowContext(ctx, spentQuery, apiKeyID, budget.PeriodStart, budget.PeriodEnd).Scan(&totalSpent)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate API key budget spend: %w", err)
	}

	budget.SpentCents = totalSpent
	budget.RemainingCents = budget.BudgetCents - totalSpent
	if budget.BudgetCents > 0 {
		budget.SpentPercent = float64(totalSpent) / float64(budget.BudgetCents) * 100
	}

	return budget, nil
}

// ListAPIKeyBudgets lists all active budgets for a tenant
func (r *BillingRepository) ListAPIKeyBudgets(ctx context.Context, tenantID uuid.UUID) ([]*APIKeyBudget, error) {
	query := `
		SELECT akb.id, akb.api_key_id, akb.tenant_id, akb.budget_cents,
			akb.warning_threshold_pct, akb.critical_threshold_pct,
			akb.period_start, akb.period_end, akb.alert_email, akb.disable_at_limit,
			akb.is_active, akb.created_at, akb.updated_at
		FROM api_key_budgets akb
		INNER JOIN api_keys ak ON ak.id = akb.api_key_id
		WHERE akb.tenant_id = $1 AND akb.is_active = true
		ORDER BY akb.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []*APIKeyBudget
	for rows.Next() {
		var b APIKeyBudget
		err := rows.Scan(
			&b.ID, &b.APIKeyID, &b.TenantID, &b.BudgetCents,
			&b.WarningThresholdPct, &b.CriticalThresholdPct,
			&b.PeriodStart, &b.PeriodEnd, &b.AlertEmail, &b.DisableAtLimit,
			&b.IsActive, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			continue
		}
		budgets = append(budgets, &b)
	}

	return budgets, rows.Err()
}

// CheckAPIKeyBudgetAlerts checks and returns any budget alerts that should be triggered
func (r *BillingRepository) CheckAPIKeyBudgetAlerts(ctx context.Context, apiKeyID uuid.UUID) (string, bool, error) {
	budget, err := r.GetAPIKeyBudgetWithSpend(ctx, apiKeyID)
	if err != nil {
		return "", false, err // No budget set or other error
	}

	// Check if we should disable
	if budget.DisableAtLimit && budget.SpentPercent >= 100 {
		// Disable the API key
		_, _ = r.db.ExecContext(ctx, `UPDATE api_keys SET is_active = false WHERE id = $1`, apiKeyID)
		return "critical", true, nil
	}

	// Check for critical threshold
	if budget.SpentPercent >= float64(budget.CriticalThresholdPct) {
		if budget.LastAlertLevel != "critical" || budget.LastAlertAt == nil ||
			time.Since(*budget.LastAlertAt) > 24*time.Hour {
			return "critical", true, nil
		}
	}

	// Check for warning threshold
	if budget.SpentPercent >= float64(budget.WarningThresholdPct) {
		if budget.LastAlertLevel != "warning" || budget.LastAlertAt == nil ||
			time.Since(*budget.LastAlertAt) > 24*time.Hour {
			return "warning", true, nil
		}
	}

	return "", false, nil
}

// UpdateAPIKeyBudgetAlert updates the last alert timestamp after sending
func (r *BillingRepository) UpdateAPIKeyBudgetAlert(ctx context.Context, apiKeyID uuid.UUID, level string) error {
	now := time.Now()
	query := `
		UPDATE api_key_budgets
		SET last_alert_at = $2, last_alert_level = $3, updated_at = $2
		WHERE api_key_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, apiKeyID, now, level)
	return err
}

// GetUnattributedCosts returns costs that don't have API key attribution
func (r *BillingRepository) GetUnattributedCosts(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) (int64, error) {
	query := `
		SELECT COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries
		WHERE tenant_id = $1
		  AND timestamp >= $2
		  AND timestamp <= $3
		  AND api_key_id IS NULL
	`

	var total int64
	err := r.db.QueryRowContext(ctx, query, tenantID, start, end).Scan(&total)
	return total, err
}

// GetAPIKeyAttributionRate returns the percentage of costs that have API key attribution
func (r *BillingRepository) GetAPIKeyAttributionRate(
	ctx context.Context,
	tenantID uuid.UUID,
	start, end time.Time,
) (float64, error) {
	var total, attributed int64

	// Get total costs
	totalQuery := `
		SELECT COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
	`
	err := r.db.QueryRowContext(ctx, totalQuery, tenantID, start, end).Scan(&total)
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 100, nil // All zero, consider it 100% attributed
	}

	// Get attributed costs
	attributedQuery := `
		SELECT COALESCE(SUM(total_cost_cents), 0)
		FROM cost_allocation_entries
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3 AND api_key_id IS NOT NULL
	`
	err = r.db.QueryRowContext(ctx, attributedQuery, tenantID, start, end).Scan(&attributed)
	if err != nil {
		return 0, err
	}

	return float64(attributed) / float64(total) * 100, nil
}
