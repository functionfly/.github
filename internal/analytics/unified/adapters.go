package unified

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Adapters read from existing stores and return raw data for the unified service.
// They use storage.Repository, gorm.DB, AggregationService, and attribution tables.

// FunctionExecutionsResult holds function execution stats for a tenant in a range.
type FunctionExecutionsResult struct {
	Total  int64
	ByDay  []TimeSeriesPoint
	ByHour []TimeSeriesPoint
}

// StateUsageResult holds state usage for a tenant (from AggregationService).
type StateUsageResult struct {
	StorageBytes int64
	ReadOps      int64
	WriteOps     int64
	ActiveStates int64
	PeriodStart  time.Time
	PeriodEnd    time.Time
}

// BillingUsageResult holds billing usage for a tenant in a range.
type BillingUsageResult struct {
	TotalQuantity int
	ByDay         []TimeSeriesPoint
}

// AgentUsageResult holds agent execution stats for a tenant in a range.
type AgentUsageResult struct {
	Calls        int64
	CostUSD      float64
	SuccessCount int64
	ErrorCount   int64
	ByDay        []TimeSeriesPoint
}

// RegistryExecutionsResult holds registry execution count for a tenant in a range.
type RegistryExecutionsResult struct {
	Total int64
	ByDay []TimeSeriesPoint
}

// GetFunctionExecutions returns function execution counts for a tenant in [start, end].
func GetFunctionExecutions(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, start, end time.Time) (FunctionExecutionsResult, error) {
	var total int64
	err := db.WithContext(ctx).Table("function_logs fl").
		Joins("INNER JOIN functions f ON f.id = fl.function_id").
		Where("f.tenant_id = ? AND fl.timestamp >= ? AND fl.timestamp <= ?", tenantID, start, end).
		Count(&total).Error
	if err != nil {
		return FunctionExecutionsResult{}, err
	}

	// Daily buckets
	var dayRows []struct {
		Bucket time.Time
		Count  int64
	}
	err = db.WithContext(ctx).Table("function_logs fl").
		Select("date_trunc('day', fl.timestamp) AS bucket, COUNT(*)::bigint AS count").
		Joins("INNER JOIN functions f ON f.id = fl.function_id").
		Where("f.tenant_id = ? AND fl.timestamp >= ? AND fl.timestamp <= ?", tenantID, start, end).
		Group("date_trunc('day', fl.timestamp)").
		Order("bucket").
		Scan(&dayRows).Error
	if err != nil {
		return FunctionExecutionsResult{Total: total}, err
	}
	byDay := make([]TimeSeriesPoint, len(dayRows))
	for i, r := range dayRows {
		byDay[i] = TimeSeriesPoint{Bucket: r.Bucket, Value: float64(r.Count), Count: r.Count}
	}

	return FunctionExecutionsResult{Total: total, ByDay: byDay}, nil
}

// GetStateUsage returns state usage for a tenant (from AggregationService).
func GetStateUsage(ctx context.Context, agg *services.AggregationService, tenantID uuid.UUID) (StateUsageResult, error) {
	sum, err := agg.GetTenantUsageSummary(ctx, tenantID)
	if err != nil {
		return StateUsageResult{}, err
	}
	return StateUsageResult{
		StorageBytes: sum.TotalStorageBytes,
		ReadOps:      sum.TotalReadOps,
		WriteOps:     sum.TotalWriteOps,
		ActiveStates: sum.ActiveStates,
		PeriodStart:  sum.PeriodStart,
		PeriodEnd:    sum.PeriodEnd,
	}, nil
}

// GetBillingUsage returns billing usage for a tenant in [start, end] from usage_rollups.
func GetBillingUsage(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, start, end time.Time) (BillingUsageResult, error) {
	var total int
	err := db.WithContext(ctx).Table("usage_rollups").
		Where("tenant_id = ? AND period_date >= ? AND period_date <= ?", tenantID, start, end).
		Select("COALESCE(SUM(total_quantity), 0)").
		Scan(&total).Error
	if err != nil {
		return BillingUsageResult{}, err
	}

	var dayRows []struct {
		Bucket time.Time
		Sum    int
	}
	err = db.WithContext(ctx).Table("usage_rollups").
		Select("period_date AS bucket, COALESCE(SUM(total_quantity), 0) AS sum").
		Where("tenant_id = ? AND period_date >= ? AND period_date <= ?", tenantID, start, end).
		Group("period_date").
		Order("bucket").
		Scan(&dayRows).Error
	if err != nil {
		return BillingUsageResult{TotalQuantity: total}, err
	}
	byDay := make([]TimeSeriesPoint, len(dayRows))
	for i, r := range dayRows {
		byDay[i] = TimeSeriesPoint{Bucket: r.Bucket, Value: float64(r.Sum), Count: int64(r.Sum)}
	}

	return BillingUsageResult{TotalQuantity: total, ByDay: byDay}, nil
}

// GetAgentUsage returns agent execution stats for a tenant in [start, end].
func GetAgentUsage(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, start, end time.Time) (AgentUsageResult, error) {
	var totals struct {
		Calls        int64
		CostUSD      float64
		SuccessCount int64
		ErrorCount   int64
	}
	err := db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
		Where("tenant_id = ? AND timestamp >= ? AND timestamp <= ?", tenantID, start, end).
		Select(`
			COUNT(*) AS calls,
			COALESCE(SUM(cost_usd), 0) AS cost_usd,
			SUM(CASE WHEN outcome = ? THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN outcome = ? THEN 1 ELSE 0 END) AS error_count
		`, attribution.OutcomeSuccess, attribution.OutcomeError).
		Scan(&totals).Error
	if err != nil {
		return AgentUsageResult{}, err
	}

	var dayRows []struct {
		Bucket time.Time
		Count  int64
	}
	err = db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
		Select("date_trunc('day', timestamp) AS bucket, COUNT(*)::bigint AS count").
		Where("tenant_id = ? AND timestamp >= ? AND timestamp <= ?", tenantID, start, end).
		Group("date_trunc('day', timestamp)").
		Order("bucket").
		Scan(&dayRows).Error
	if err != nil {
		return AgentUsageResult{
			Calls:        totals.Calls,
			CostUSD:      totals.CostUSD,
			SuccessCount: totals.SuccessCount,
			ErrorCount:   totals.ErrorCount,
		}, err
	}
	byDay := make([]TimeSeriesPoint, len(dayRows))
	for i, r := range dayRows {
		byDay[i] = TimeSeriesPoint{Bucket: r.Bucket, Value: float64(r.Count), Count: r.Count}
	}

	return AgentUsageResult{
		Calls:        totals.Calls,
		CostUSD:      totals.CostUSD,
		SuccessCount: totals.SuccessCount,
		ErrorCount:   totals.ErrorCount,
		ByDay:        byDay,
	}, nil
}

// GetRegistryExecutions returns registry execution count for a tenant in [start, end].
// RegistryFunctionExecution has TenantID; we count by tenant_id.
func GetRegistryExecutions(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, start, end time.Time) (RegistryExecutionsResult, error) {
	// Use registry package model - table name is registry_function_executions
	type regExec struct {
		ID         uuid.UUID
		FunctionID uuid.UUID
		TenantID   *uuid.UUID
		Timestamp  time.Time
	}
	var total int64
	err := db.WithContext(ctx).Table("registry_function_executions").
		Where("tenant_id = ? AND timestamp >= ? AND timestamp <= ?", tenantID, start, end).
		Count(&total).Error
	if err != nil {
		return RegistryExecutionsResult{}, err
	}

	var dayRows []struct {
		Bucket time.Time
		Count  int64
	}
	err = db.WithContext(ctx).Table("registry_function_executions").
		Select("date_trunc('day', timestamp) AS bucket, COUNT(*)::bigint AS count").
		Where("tenant_id = ? AND timestamp >= ? AND timestamp <= ?", tenantID, start, end).
		Group("date_trunc('day', timestamp)").
		Order("bucket").
		Scan(&dayRows).Error
	if err != nil {
		return RegistryExecutionsResult{Total: total}, err
	}
	byDay := make([]TimeSeriesPoint, len(dayRows))
	for i, r := range dayRows {
		byDay[i] = TimeSeriesPoint{Bucket: r.Bucket, Value: float64(r.Count), Count: r.Count}
	}

	return RegistryExecutionsResult{Total: total, ByDay: byDay}, nil
}

// GetStateUsageTimeSeries returns state read+write ops by day for a tenant in [start, end].
// state_usage_metrics has metric_type (read_ops, write_ops, daily_read_ops, daily_write_ops) and period_start.
func GetStateUsageTimeSeries(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, start, end time.Time) ([]TimeSeriesPoint, error) {
	var rows []struct {
		Bucket time.Time
		Value  int64
	}
	err := db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Select("period_start AS bucket, SUM(value) AS value").
		Where("tenant_id = ? AND metric_type IN (?, ?, ?, ?) AND period_start >= ? AND period_start <= ?",
			tenantID, "read_ops", "write_ops", "daily_read_ops", "daily_write_ops", start, end).
		Group("period_start").
		Order("bucket").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	points := make([]TimeSeriesPoint, len(rows))
	for i, r := range rows {
		points[i] = TimeSeriesPoint{Bucket: r.Bucket, Value: float64(r.Value), Count: r.Value}
	}
	return points, nil
}
