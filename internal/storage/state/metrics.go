package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

func (m *MetricsCollector) collectStorageMetrics(ctx context.Context) {
	var states []State
	err := m.db.WithContext(ctx).Find(&states).Error
	if err != nil {
		m.logger.WithError(err).Error("Failed to collect storage metrics")
		return
	}

	now := time.Now()
	periodStart := now.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, state := range states {
		wg.Add(1)
		go func(s State) {
			defer wg.Done()

			var totalSize int64
			err := m.db.WithContext(ctx).
				Model(&StateValue{}).
				Where("state_id = ?", s.ID).
				Select("COALESCE(SUM(jsonb_approx_size(value::jsonb)), 0)").
				Scan(&totalSize).Error

			if err != nil {
				m.logger.WithError(err).Errorf("Failed to calculate size for state %s", s.ID)
				return
			}

			metric := &StateUsageMetric{
				ID:          uuid.New(),
				TenantID:    s.TenantID,
				StateID:     &s.ID,
				MetricType:  "storage",
				Value:       totalSize,
				Unit:        "bytes",
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				CreatedAt:   now,
			}

			mu.Lock()
			err = m.db.WithContext(ctx).Create(metric).Error
			mu.Unlock()

			if err != nil {
				m.logger.WithError(err).Errorf("Failed to store storage metric for state %s", s.ID)
			}
		}(state)
	}

	wg.Wait()
}

func (m *MetricsCollector) collectOperationMetrics(ctx context.Context) {
	var states []State
	err := m.db.WithContext(ctx).Find(&states).Error
	if err != nil {
		m.logger.WithError(err).Error("Failed to collect operation metrics")
		return
	}

	now := time.Now()
	periodStart := now.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, state := range states {
		wg.Add(1)
		go func(s State) {
			defer wg.Done()

			var writeOps int64
			err := m.db.WithContext(ctx).
				Model(&StateEvent{}).
				Where("state_id = ? AND timestamp >= ?", s.ID, periodStart).
				Count(&writeOps).Error

			if err != nil {
				m.logger.WithError(err).Errorf("Failed to count writes for state %s", s.ID)
				return
			}

			writeMetric := &StateUsageMetric{
				ID:          uuid.New(),
				TenantID:    s.TenantID,
				StateID:     &s.ID,
				MetricType:  "write_ops",
				Value:       writeOps,
				Unit:        "ops",
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				CreatedAt:   now,
			}

			mu.Lock()
			err = m.db.WithContext(ctx).Create(writeMetric).Error
			mu.Unlock()

			if err != nil {
				m.logger.WithError(err).Errorf("Failed to store write metric for state %s", s.ID)
			}
		}(state)
	}

	wg.Wait()
}

func (m *MetricsCollector) collectTenantMetrics(ctx context.Context) {
	var tenants []uuid.UUID
	err := m.db.WithContext(ctx).
		Model(&State{}).
		Distinct("tenant_id").
		Pluck("tenant_id", &tenants).Error
	if err != nil {
		m.logger.WithError(err).Error("Failed to collect tenant metrics")
		return
	}

	now := time.Now()
	periodStart := now.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, tenantID := range tenants {
		wg.Add(1)
		go func(tid uuid.UUID) {
			defer wg.Done()

			var stateCount int64
			m.db.WithContext(ctx).
				Model(&State{}).
				Where("tenant_id = ?", tid).
				Count(&stateCount)

			var totalStorage int64
			m.db.WithContext(ctx).
				Raw(`
					SELECT COALESCE(SUM(jsonb_approx_size(value::jsonb)), 0)
					FROM state_values sv
					JOIN states s ON sv.state_id = s.id
					WHERE s.tenant_id = ?
				`, tid).Scan(&totalStorage)

			metric := &StateUsageMetric{
				ID:          uuid.New(),
				TenantID:    tid,
				MetricType:  "tenant_aggregate",
				Value:       totalStorage,
				Unit:        "bytes",
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				CreatedAt:   now,
			}

			mu.Lock()
			err = m.db.WithContext(ctx).Create(metric).Error
			mu.Unlock()

			if err != nil {
				m.logger.WithError(err).Errorf("Failed to store tenant metric for %s", tid)
			}
		}(tenantID)
	}

	wg.Wait()
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
			SELECT COALESCE(SUM(jsonb_approx_size(value::jsonb)), 0)
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
