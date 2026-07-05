package state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MetricsCollectorConfig struct {
	Interval    time.Duration
	BatchSize   int
	Enabled     bool
}

func DefaultMetricsCollectorConfig() MetricsCollectorConfig {
	return MetricsCollectorConfig{
		Interval:  5 * time.Minute,
		BatchSize: 100,
		Enabled:   true,
	}
}

type MetricsCollector struct {
	db     *gorm.DB
	config MetricsCollectorConfig
	logger *logrus.Logger
}

func NewMetricsCollector(db *gorm.DB, config MetricsCollectorConfig) *MetricsCollector {
	if config.Interval == 0 {
		config = DefaultMetricsCollectorConfig()
	}
	return &MetricsCollector{
		db:     db,
		config: config,
		logger: logrus.New(),
	}
}

func (m *MetricsCollector) Start(ctx context.Context) {
	if !m.config.Enabled {
		m.logger.Info("Metrics collector is disabled")
		return
	}

	m.logger.WithFields(logrus.Fields{
		"interval": m.config.Interval,
	}).Info("Starting metrics collector")

	m.collectMetrics()

	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping metrics collector")
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

func (m *MetricsCollector) collectMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	m.collectStorageMetrics(ctx)
	m.collectOperationMetrics(ctx)
	m.collectTenantMetrics(ctx)

	m.logger.Info("Metrics collection completed")
}

// collectStorageMetrics gathers per-state storage sizes using a single GROUP BY query
// instead of N concurrent goroutines.  This avoids connection pool saturation and
// reduces round-trips from O(states) to 1.
func (m *MetricsCollector) collectStorageMetrics(ctx context.Context) {
	var states []State
	if err := m.db.WithContext(ctx).Find(&states).Error; err != nil {
		m.logger.WithError(err).Error("Failed to collect storage metrics")
		return
	}
	if len(states) == 0 {
		return
	}

	// Single query: total size per state_id
	type stateSize struct {
		StateID   uuid.UUID `gorm:"column:state_id"`
		TotalSize int64     `gorm:"column:total_size"`
	}
	var sizes []stateSize
	if err := m.db.WithContext(ctx).Raw(`
		SELECT state_id, COALESCE(SUM(octet_length(value::text)), 0) AS total_size
		FROM state_values
		GROUP BY state_id
	`).Scan(&sizes).Error; err != nil {
		m.logger.WithError(err).Error("Failed to compute storage sizes")
		return
	}

	sizeMap := make(map[uuid.UUID]int64, len(sizes))
	for _, s := range sizes {
		sizeMap[s.StateID] = s.TotalSize
	}

	now := time.Now()
	periodStart := now.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	metrics := make([]*StateUsageMetric, 0, len(states))
	for _, s := range states {
		metrics = append(metrics, &StateUsageMetric{
			ID:          uuid.New(),
			TenantID:    s.TenantID,
			StateID:     &s.ID,
			MetricType:  "storage",
			Value:       sizeMap[s.ID], // 0 if state has no values
			Unit:        "bytes",
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   now,
		})
	}

	if err := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:     []clause.Column{{Name: "state_id"}, {Name: "metric_type"}, {Name: "period_start"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "state_id IS NOT NULL"}}},
		DoUpdates:   clause.AssignmentColumns([]string{"value", "unit", "period_end", "created_at"}),
	}).CreateInBatches(metrics, m.config.BatchSize).Error; err != nil {
		m.logger.WithError(err).Error("Failed to batch insert storage metrics")
	}
}

// collectOperationMetrics gathers per-state write-operation counts using a single
// GROUP BY query instead of N concurrent goroutines.
func (m *MetricsCollector) collectOperationMetrics(ctx context.Context) {
	var states []State
	if err := m.db.WithContext(ctx).Find(&states).Error; err != nil {
		m.logger.WithError(err).Error("Failed to collect operation metrics")
		return
	}
	if len(states) == 0 {
		return
	}

	now := time.Now()
	periodStart := now.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	// Single query: event count per state_id for the current period
	type stateCount struct {
		StateID   uuid.UUID `gorm:"column:state_id"`
		WriteOps  int64     `gorm:"column:write_ops"`
	}
	var counts []stateCount
	if err := m.db.WithContext(ctx).Raw(`
		SELECT state_id, COUNT(*) AS write_ops
		FROM state_events
		WHERE timestamp >= ?
		GROUP BY state_id
	`, periodStart).Scan(&counts).Error; err != nil {
		m.logger.WithError(err).Error("Failed to compute operation counts")
		return
	}

	countMap := make(map[uuid.UUID]int64, len(counts))
	for _, c := range counts {
		countMap[c.StateID] = c.WriteOps
	}

	metrics := make([]*StateUsageMetric, 0, len(states))
	for _, s := range states {
		metrics = append(metrics, &StateUsageMetric{
			ID:          uuid.New(),
			TenantID:    s.TenantID,
			StateID:     &s.ID,
			MetricType:  "write_ops",
			Value:       countMap[s.ID], // 0 if state has no events
			Unit:        "ops",
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   now,
		})
	}

	if err := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:     []clause.Column{{Name: "state_id"}, {Name: "metric_type"}, {Name: "period_start"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "state_id IS NOT NULL"}}},
		DoUpdates:   clause.AssignmentColumns([]string{"value", "unit", "period_end", "created_at"}),
	}).CreateInBatches(metrics, m.config.BatchSize).Error; err != nil {
		m.logger.WithError(err).Error("Failed to batch insert operation metrics")
	}
}

// collectTenantMetrics gathers per-tenant aggregate storage using a single JOIN + GROUP BY
// query instead of N concurrent goroutines (2 queries per tenant → 1 query total).
func (m *MetricsCollector) collectTenantMetrics(ctx context.Context) {
	now := time.Now()
	periodStart := now.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	type tenantAgg struct {
		TenantID     uuid.UUID `gorm:"column:tenant_id"`
		TotalStorage int64     `gorm:"column:total_storage"`
	}
	var aggs []tenantAgg
	if err := m.db.WithContext(ctx).Raw(`
		SELECT s.tenant_id,
		       COALESCE(SUM(octet_length(sv.value::jsonb::text)), 0) AS total_storage
		FROM states s
		LEFT JOIN state_values sv ON sv.state_id = s.id
		GROUP BY s.tenant_id
	`).Scan(&aggs).Error; err != nil {
		m.logger.WithError(err).Error("Failed to compute tenant aggregates")
		return
	}

	if len(aggs) == 0 {
		return
	}

	metrics := make([]*StateUsageMetric, 0, len(aggs))
	for _, a := range aggs {
		metrics = append(metrics, &StateUsageMetric{
			ID:          uuid.New(),
			TenantID:    a.TenantID,
			MetricType:  "tenant_aggregate",
			Value:       a.TotalStorage,
			Unit:        "bytes",
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   now,
		})
	}

	if err := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:     []clause.Column{{Name: "tenant_id"}, {Name: "metric_type"}, {Name: "period_start"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "state_id IS NULL"}}},
		DoUpdates:   clause.AssignmentColumns([]string{"value", "unit", "period_end", "created_at"}),
	}).CreateInBatches(metrics, m.config.BatchSize).Error; err != nil {
		m.logger.WithError(err).Error("Failed to batch insert tenant metrics")
	}
}

func (m *MetricsCollector) GetMetrics(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*StateUsageMetric, error) {
	var metrics []*StateUsageMetric
	err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND period_start >= ? AND period_end <= ?", tenantID, start, end).
		Order("period_start DESC").
		Find(&metrics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	return metrics, nil
}

func (m *MetricsCollector) GetTenantSummary(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	var totalStates int64
	m.db.WithContext(ctx).Model(&State{}).Where("tenant_id = ?", tenantID).Count(&totalStates)
	summary["total_states"] = totalStates

	var totalStorage int64
	m.db.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(octet_length(value::text)), 0)
			FROM state_values sv
			JOIN states s ON sv.state_id = s.id
			WHERE s.tenant_id = ?
		`, tenantID).Scan(&totalStorage)
	summary["total_storage_bytes"] = totalStorage

	var totalEvents int64
	m.db.WithContext(ctx).
		Model(&StateEvent{}).
		Joins("JOIN states ON state_events.state_id = states.id").
		Where("states.tenant_id = ?", tenantID).
		Count(&totalEvents)
	summary["total_events"] = totalEvents

	var totalMemories int64
	m.db.WithContext(ctx).Model(&AgentMemory{}).Where("tenant_id = ?", tenantID).Count(&totalMemories)
	summary["total_memories"] = totalMemories

	return summary, nil
}
