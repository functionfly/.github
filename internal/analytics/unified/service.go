package unified

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceConfig configures the unified analytics service.
type ServiceConfig struct {
	UseRollups  bool        // when true, read from analytics_rollups first (fallback to live)
	EventStore  *EventStore // optional; when set and UseRollups, rollups are used
}

// Service is the unified analytics service that aggregates data from all sources.
type Service struct {
	db        *gorm.DB
	usageAgg  *services.AggregationService
	config    ServiceConfig
}

// NewService creates a new unified analytics service.
func NewService(db *gorm.DB, usageAgg *services.AggregationService, config ...ServiceConfig) *Service {
	cfg := ServiceConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Service{
		db:       db,
		usageAgg: usageAgg,
		config:   cfg,
	}
}

// TenantSummary returns the unified summary for a tenant over [start, end].
// When config.UseRollups and config.EventStore are set, reads from analytics_rollups first.
func (s *Service) TenantSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*TenantSummary, error) {
	now := time.Now().UTC()
	out := &TenantSummary{
		TenantID:    tenantID,
		Start:       start,
		End:         end,
		GeneratedAt: now,
	}

	if s.config.UseRollups && s.config.EventStore != nil {
		if fromRollups := s.tenantSummaryFromRollups(ctx, tenantID, start, end); fromRollups != nil {
			fromRollups.TenantID = tenantID
			fromRollups.Start = start
			fromRollups.End = end
			fromRollups.GeneratedAt = now
			return fromRollups, nil
		}
		// fallback to live below
	}

	// Function executions
	funcRes, err := GetFunctionExecutions(ctx, s.db, tenantID, start, end)
	if err != nil {
		return out, err
	}
	out.FunctionExecutions = funcRes.Total

	// State usage (from aggregation service; may not align exactly with start/end)
	stateRes, err := GetStateUsage(ctx, s.usageAgg, tenantID)
	if err == nil {
		out.StateStorageBytes = stateRes.StorageBytes
		out.StateReadOps = stateRes.ReadOps
		out.StateWriteOps = stateRes.WriteOps
		out.StateActiveStates = stateRes.ActiveStates
	}

	// Billing
	billingRes, err := GetBillingUsage(ctx, s.db, tenantID, start, end)
	if err == nil {
		out.BillingQuantity = billingRes.TotalQuantity
	}

	// Agent
	agentRes, err := GetAgentUsage(ctx, s.db, tenantID, start, end)
	if err == nil {
		out.AgentCalls = agentRes.Calls
		out.AgentCostUSD = agentRes.CostUSD
		out.AgentSuccessCount = agentRes.SuccessCount
		out.AgentErrorCount = agentRes.ErrorCount
	}

	// Registry
	regRes, err := GetRegistryExecutions(ctx, s.db, tenantID, start, end)
	if err == nil {
		out.RegistryExecutions = regRes.Total
	}

	return out, nil
}

// tenantSummaryFromRollups reads from analytics_rollups for the tenant and range; returns nil on error or no data.
func (s *Service) tenantSummaryFromRollups(ctx context.Context, tenantID uuid.UUID, start, end time.Time) *TenantSummary {
	var rows []struct {
		MetricName string
		Value     float64
	}
	err := s.db.WithContext(ctx).Model(&AnalyticsRollup{}).
		Select("metric_name, SUM(value) AS value").
		Where("tenant_id = ? AND period = ? AND period_start >= ? AND period_start < ?", tenantID, "day", start, end).
		Group("metric_name").
		Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil
	}
	out := &TenantSummary{}
	for _, r := range rows {
		switch r.MetricName {
		case RollupMetricFunctionExecutions:
			out.FunctionExecutions = int64(r.Value)
		case RollupMetricStateReadOps:
			out.StateReadOps = int64(r.Value)
		case RollupMetricStateWriteOps:
			out.StateWriteOps = int64(r.Value)
		case RollupMetricStateStorageBytes:
			out.StateStorageBytes = int64(r.Value)
		case RollupMetricBillingQuantity:
			out.BillingQuantity = int(r.Value)
		case RollupMetricAgentCalls:
			out.AgentCalls = int64(r.Value)
		case RollupMetricAgentCostUSD:
			out.AgentCostUSD = r.Value
		case RollupMetricRegistryExecutions:
			out.RegistryExecutions = int64(r.Value)
		}
	}
	return out
}

// TenantTimeSeries returns time-bucketed series for the given metric kind and granularity.
// When UseRollups is set, reads from analytics_rollups first.
func (s *Service) TenantTimeSeries(ctx context.Context, tenantID uuid.UUID, kind MetricKind, granularity Granularity, start, end time.Time) (*TimeSeriesResponse, error) {
	out := &TimeSeriesResponse{
		TenantID:    tenantID,
		MetricKind:  kind,
		Granularity: granularity,
		Start:       start,
		End:         end,
		Points:      nil,
	}

	metricName := metricKindToRollup(kind)
	if s.config.UseRollups && s.config.EventStore != nil && metricName != "" {
		points := s.timeSeriesFromRollups(ctx, tenantID, "day", metricName, start, end)
		if len(points) > 0 {
			out.Points = points
			return out, nil
		}
	}

	switch kind {
	case MetricKindExecutions:
		res, err := GetFunctionExecutions(ctx, s.db, tenantID, start, end)
		if err != nil {
			return out, err
		}
		out.Points = res.ByDay
		if granularity == GranularityHour {
			// We could add ByHour in adapters; for now day is supported
			out.Points = res.ByDay
		}
	case MetricKindStateOps:
		points, err := GetStateUsageTimeSeries(ctx, s.db, tenantID, start, end)
		if err != nil {
			return out, err
		}
		out.Points = points
	case MetricKindBilling:
		res, err := GetBillingUsage(ctx, s.db, tenantID, start, end)
		if err != nil {
			return out, err
		}
		out.Points = res.ByDay
	case MetricKindAgentCalls:
		res, err := GetAgentUsage(ctx, s.db, tenantID, start, end)
		if err != nil {
			return out, err
		}
		out.Points = res.ByDay
	case MetricKindRegistryRuns:
		res, err := GetRegistryExecutions(ctx, s.db, tenantID, start, end)
		if err != nil {
			return out, err
		}
		out.Points = res.ByDay
	default:
		out.Points = []TimeSeriesPoint{}
		return out, nil
	}

	return out, nil
}

func metricKindToRollup(kind MetricKind) string {
	switch kind {
	case MetricKindExecutions:
		return RollupMetricFunctionExecutions
	case MetricKindStateOps:
		return "" // rollups store read/write separately; use live adapter for combined state_ops
	case MetricKindBilling:
		return RollupMetricBillingQuantity
	case MetricKindAgentCalls:
		return RollupMetricAgentCalls
	case MetricKindRegistryRuns:
		return RollupMetricRegistryExecutions
	}
	return ""
}

func (s *Service) timeSeriesFromRollups(ctx context.Context, tenantID uuid.UUID, period, metricName string, start, end time.Time) []TimeSeriesPoint {
	var rows []struct {
		Bucket time.Time
		Value  float64
	}
	err := s.db.WithContext(ctx).Model(&AnalyticsRollup{}).
		Select("period_start AS bucket, SUM(value) AS value").
		Where("tenant_id = ? AND period = ? AND metric_name = ? AND period_start >= ? AND period_start < ?", tenantID, period, metricName, start, end).
		Group("period_start").
		Order("bucket").
		Scan(&rows).Error
	if err != nil {
		return nil
	}
	points := make([]TimeSeriesPoint, len(rows))
	for i, r := range rows {
		points[i] = TimeSeriesPoint{Bucket: r.Bucket, Value: r.Value, Count: int64(r.Value)}
	}
	return points
}

// PlatformSummary returns platform-wide rollups for admin (all tenants in [start, end]).
func (s *Service) PlatformSummary(ctx context.Context, start, end time.Time) (*PlatformSummary, error) {
	now := time.Now().UTC()
	out := &PlatformSummary{
		Start:       start,
		End:         end,
		GeneratedAt: now,
	}

	// Count distinct tenants with activity (function_logs + functions)
	var tenantCount int64
	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT f.tenant_id) FROM function_logs fl
		INNER JOIN functions f ON f.id = fl.function_id
		WHERE fl.timestamp >= ? AND fl.timestamp <= ?
	`, start, end).Scan(&tenantCount)
	out.TotalTenantsActive = tenantCount

	// Total function executions
	var totalFunc int64
	s.db.WithContext(ctx).Table("function_logs fl").
		Joins("INNER JOIN functions f ON f.id = fl.function_id").
		Where("fl.timestamp >= ? AND fl.timestamp <= ?", start, end).
		Count(&totalFunc)
	out.TotalFunctionExecs = totalFunc

	// State ops: sum over state_usage_metrics in range
	var stateRead, stateWrite int64
	s.db.WithContext(ctx).Table("state_usage_metrics").
		Where("metric_type IN (?, ?) AND period_start >= ? AND period_start <= ?", "read_ops", "daily_read_ops", start, end).
		Select("COALESCE(SUM(value), 0)").
		Scan(&stateRead)
	s.db.WithContext(ctx).Table("state_usage_metrics").
		Where("metric_type IN (?, ?) AND period_start >= ? AND period_start <= ?", "write_ops", "daily_write_ops", start, end).
		Select("COALESCE(SUM(value), 0)").
		Scan(&stateWrite)
	out.TotalStateReadOps = stateRead
	out.TotalStateWriteOps = stateWrite

	// Agent calls
	var agentCalls int64
	s.db.WithContext(ctx).Table("agent_execution_records").
		Where("timestamp >= ? AND timestamp <= ?", start, end).
		Count(&agentCalls)
	out.TotalAgentCalls = agentCalls

	// Registry executions
	var regExecs int64
	s.db.WithContext(ctx).Table("registry_function_executions").
		Where("timestamp >= ? AND timestamp <= ?", start, end).
		Count(&regExecs)
	out.TotalRegistryExecs = regExecs

	return out, nil
}
