package unified

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EventStore writes to the canonical analytics_events table and runs sync/aggregation jobs.
type EventStore struct {
	db *gorm.DB
}

// NewEventStore creates a new event store.
func NewEventStore(db *gorm.DB) *EventStore {
	return &EventStore{db: db}
}

// RecordEvent inserts a single analytics event.
func (s *EventStore) RecordEvent(ctx context.Context, e *AnalyticsEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	return s.db.WithContext(ctx).Create(e).Error
}

// RecordEvents inserts multiple events in a batch.
func (s *EventStore) RecordEvents(ctx context.Context, events []*AnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, e := range events {
		if e.ID == uuid.Nil {
			e.ID = uuid.New()
		}
		if e.CreatedAt.IsZero() {
			e.CreatedAt = now
		}
	}
	return s.db.WithContext(ctx).CreateInBatches(events, 200).Error
}

// SyncFromSources reads from existing tables (function_logs, state_usage_metrics, agent_execution_records,
// registry_function_executions, usage_rollups) for the given time range and writes aggregated facts into
// analytics_rollups (daily). Idempotent: upserts by (tenant_id, period, period_start, metric_name).
func (s *EventStore) SyncFromSources(ctx context.Context, start, end time.Time) error {
	log := logrus.WithField("start", start).WithField("end", end)
	log.Info("Unified analytics: syncing from sources to rollups")

	// Iterate day by day
	for d := start.Truncate(24 * time.Hour); d.Before(end) || d.Equal(end); d = d.AddDate(0, 0, 1) {
		dayEnd := d.AddDate(0, 0, 1)

		// Function executions per tenant for this day
		var funcRows []struct {
			TenantID uuid.UUID
			Count    int64
		}
		if err := s.db.WithContext(ctx).Table("function_logs fl").
			Select("f.tenant_id AS tenant_id, COUNT(*)::bigint AS count").
			Joins("INNER JOIN functions f ON f.id = fl.function_id").
			Where("fl.timestamp >= ? AND fl.timestamp < ?", d, dayEnd).
			Group("f.tenant_id").
			Scan(&funcRows).Error; err != nil {
			log.WithError(err).Warn("sync function_logs failed")
		} else {
			for _, r := range funcRows {
				_ = s.upsertRollup(ctx, r.TenantID, "day", d, RollupMetricFunctionExecutions, float64(r.Count))
			}
		}

		// State ops: from state_usage_metrics (daily_read_ops, daily_write_ops or by date)
		var stateReadRows []struct {
			TenantID uuid.UUID
			Sum      int64
		}
		s.db.WithContext(ctx).Table("state_usage_metrics").
			Select("tenant_id, COALESCE(SUM(value), 0)::bigint AS sum").
			Where("metric_type IN (?, ?) AND period_start >= ? AND period_start < ?", "read_ops", "daily_read_ops", d, dayEnd).
			Group("tenant_id").
			Scan(&stateReadRows)
		for _, r := range stateReadRows {
			_ = s.upsertRollup(ctx, r.TenantID, "day", d, RollupMetricStateReadOps, float64(r.Sum))
		}
		var stateWriteRows []struct {
			TenantID uuid.UUID
			Sum      int64
		}
		s.db.WithContext(ctx).Table("state_usage_metrics").
			Select("tenant_id, COALESCE(SUM(value), 0)::bigint AS sum").
			Where("metric_type IN (?, ?) AND period_start >= ? AND period_start < ?", "write_ops", "daily_write_ops", d, dayEnd).
			Group("tenant_id").
			Scan(&stateWriteRows)
		for _, r := range stateWriteRows {
			_ = s.upsertRollup(ctx, r.TenantID, "day", d, RollupMetricStateWriteOps, float64(r.Sum))
		}

		// Agent executions per tenant
		var agentRows []struct {
			TenantID uuid.UUID
			Calls    int64
			Cost     float64
		}
		s.db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
			Select("tenant_id, COUNT(*) AS calls, COALESCE(SUM(cost_usd), 0) AS cost").
			Where("timestamp >= ? AND timestamp < ?", d, dayEnd).
			Group("tenant_id").
			Scan(&agentRows)
		for _, r := range agentRows {
			_ = s.upsertRollup(ctx, r.TenantID, "day", d, RollupMetricAgentCalls, float64(r.Calls))
			_ = s.upsertRollup(ctx, r.TenantID, "day", d, RollupMetricAgentCostUSD, r.Cost)
		}

		// Registry executions per tenant (registry_function_executions.tenant_id)
		var regRows []struct {
			TenantID *uuid.UUID
			Count    int64
		}
		s.db.WithContext(ctx).Table("registry_function_executions").
			Select("tenant_id, COUNT(*)::bigint AS count").
			Where("timestamp >= ? AND timestamp < ?", d, dayEnd).
			Group("tenant_id").
			Scan(&regRows)
		for _, r := range regRows {
			if r.TenantID != nil {
				_ = s.upsertRollup(ctx, *r.TenantID, "day", d, RollupMetricRegistryExecutions, float64(r.Count))
			}
		}

		// Billing: usage_rollups by period_date
		var billingRows []struct {
			TenantID uuid.UUID
			Sum      int
		}
		s.db.WithContext(ctx).Table("usage_rollups").
			Select("tenant_id, COALESCE(SUM(total_quantity), 0) AS sum").
			Where("period_date >= ? AND period_date < ?", d, dayEnd).
			Group("tenant_id").
			Scan(&billingRows)
		for _, r := range billingRows {
			_ = s.upsertRollup(ctx, r.TenantID, "day", d, RollupMetricBillingQuantity, float64(r.Sum))
		}
	}

	log.Info("Unified analytics: sync from sources completed")
	return nil
}

func (s *EventStore) upsertRollup(ctx context.Context, tenantID uuid.UUID, period string, periodStart time.Time, metricName string, value float64) error {
	var r AnalyticsRollup
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND period = ? AND period_start = ? AND metric_name = ?", tenantID, period, periodStart, metricName).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.WithContext(ctx).Create(&AnalyticsRollup{
			TenantID:   tenantID,
			Period:     period,
			PeriodStart: periodStart,
			MetricName: metricName,
			Value:      value,
		}).Error
	}
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&r).Updates(map[string]interface{}{"value": value, "updated_at": time.Now().UTC()}).Error
}

// AggregateFromEvents reads analytics_events (if populated) and writes rollups. If events table is sparse,
// use SyncFromSources instead. This aggregates by day from raw events.
func (s *EventStore) AggregateFromEvents(ctx context.Context, start, end time.Time) error {
	var eventRows []struct {
		TenantID     uuid.UUID
		ResourceType string
		EventType    string
		Day          time.Time
		Quantity     int64
		Cost         float64
	}
	err := s.db.WithContext(ctx).Model(&AnalyticsEvent{}).
		Select("tenant_id, resource_type, event_type, date_trunc('day', occurred_at) AS day, COALESCE(SUM(quantity), 0)::bigint AS quantity, COALESCE(SUM(cost_usd), 0) AS cost").
		Where("occurred_at >= ? AND occurred_at < ?", start, end).
		Group("tenant_id, resource_type, event_type, date_trunc('day', occurred_at)").
		Scan(&eventRows).Error
	if err != nil {
		return err
	}
	for _, r := range eventRows {
		metricName := rollupMetricFromEvent(r.ResourceType, r.EventType)
		if metricName == "" {
			continue
		}
		_ = s.upsertRollup(ctx, r.TenantID, "day", r.Day, metricName, float64(r.Quantity))
		if r.Cost != 0 && (r.ResourceType == ResourceTypeAgent) {
			_ = s.upsertRollup(ctx, r.TenantID, "day", r.Day, RollupMetricAgentCostUSD, r.Cost)
		}
	}
	return nil
}

func rollupMetricFromEvent(resourceType, eventType string) string {
	switch resourceType {
	case ResourceTypeFunction:
		if eventType == EventTypeExecution {
			return RollupMetricFunctionExecutions
		}
	case ResourceTypeState:
		if eventType == EventTypeStateRead {
			return RollupMetricStateReadOps
		}
		if eventType == EventTypeStateWrite {
			return RollupMetricStateWriteOps
		}
	case ResourceTypeBilling:
		if eventType == EventTypeUsage {
			return RollupMetricBillingQuantity
		}
	case ResourceTypeAgent:
		if eventType == EventTypeAgentRun {
			return RollupMetricAgentCalls
		}
	case ResourceTypeRegistry:
		if eventType == EventTypeRegistryRun {
			return RollupMetricRegistryExecutions
		}
	}
	return ""
}

// Ensure tables exist (for --skip-migrations setups: call once at startup).
func (s *EventStore) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&AnalyticsEvent{}, &AnalyticsRollup{})
}
