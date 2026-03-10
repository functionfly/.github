package services

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AggregationConfig holds configuration for the usage metrics aggregation service
type AggregationConfig struct {
	// Enabled indicates whether the aggregation service is enabled
	Enabled bool

	// Interval is the time between aggregation runs
	Interval time.Duration

	// DailyAggregationEnabled enables daily aggregation
	DailyAggregationEnabled bool

	// WeeklyAggregationEnabled enables weekly aggregation
	WeeklyAggregationEnabled bool

	// MonthlyAggregationEnabled enables monthly aggregation
	MonthlyAggregationEnabled bool
}

// DefaultAggregationConfig returns the default configuration
func DefaultAggregationConfig() *AggregationConfig {
	return &AggregationConfig{
		Enabled:                   true,
		Interval:                  1 * time.Hour,
		DailyAggregationEnabled:   true,
		WeeklyAggregationEnabled:  true,
		MonthlyAggregationEnabled: true,
	}
}

// LoadAggregationConfig loads configuration from environment variables
func LoadAggregationConfig() *AggregationConfig {
	config := DefaultAggregationConfig()

	// Check if service is enabled
	if enabledStr := os.Getenv("USAGE_METRICS_ENABLED"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			config.Enabled = enabled
		}
	}

	// Load interval
	if intervalStr := os.Getenv("USAGE_METRICS_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil && interval > 0 {
			config.Interval = interval
		}
	}

	return config
}

// TenantUsageSummary represents aggregated usage for a tenant
type TenantUsageSummary struct {
	TenantID          uuid.UUID
	PeriodStart       time.Time
	PeriodEnd         time.Time
	TotalStorageBytes int64
	TotalWriteOps     int64
	TotalReadOps      int64
	ActiveStates      int64
}

// AggregationService handles periodic aggregation of state usage metrics
type AggregationService struct {
	db       *gorm.DB
	config   *AggregationConfig
	logger   *logrus.Logger
	stopChan chan struct{}
}

// NewAggregationService creates a new usage metrics aggregation service
func NewAggregationService(db *gorm.DB, config *AggregationConfig) *AggregationService {
	return &AggregationService{
		db:       db,
		config:   config,
		logger:   logrus.New(),
		stopChan: make(chan struct{}),
	}
}

// IsEnabled returns whether the service is enabled
func (s *AggregationService) IsEnabled() bool {
	return s.config.Enabled
}

// StartAggregationRoutine starts the background aggregation routine
func (s *AggregationService) StartAggregationRoutine(ctx context.Context) {
	if !s.config.Enabled {
		s.logger.Info("Usage metrics aggregation service is disabled")
		return
	}

	s.logger.WithField("interval", s.config.Interval).Info("Starting usage metrics aggregation routine")

	// Run initial aggregation
	if err := s.AggregateUsage(ctx); err != nil {
		s.logger.WithError(err).Error("Initial usage aggregation failed")
	}

	// Start background ticker
	go func() {
		ticker := time.NewTicker(s.config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Usage metrics aggregation service stopping due to context cancellation")
				return
			case <-s.stopChan:
				s.logger.Info("Usage metrics aggregation service stopped")
				return
			case <-ticker.C:
				if err := s.AggregateUsage(ctx); err != nil {
					s.logger.WithError(err).Error("Periodic usage aggregation failed")
				}
			}
		}
	}()
}

// Stop stops the aggregation service
func (s *AggregationService) Stop() {
	close(s.stopChan)
}

// AggregateUsage performs the aggregation of usage metrics
func (s *AggregationService) AggregateUsage(ctx context.Context) error {
	start := time.Now()
	s.logger.Info("Starting usage metrics aggregation")

	// Aggregate daily metrics if enabled
	if s.config.DailyAggregationEnabled {
		if err := s.aggregateDailyMetrics(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to aggregate daily metrics")
			return err
		}
	}

	// Aggregate weekly metrics if enabled
	if s.config.WeeklyAggregationEnabled {
		if err := s.aggregateWeeklyMetrics(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to aggregate weekly metrics")
			return err
		}
	}

	// Aggregate monthly metrics if enabled
	if s.config.MonthlyAggregationEnabled {
		if err := s.aggregateMonthlyMetrics(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to aggregate monthly metrics")
			return err
		}
	}

	// Generate tenant summaries
	if err := s.generateTenantSummaries(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to generate tenant summaries")
		return err
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Usage metrics aggregation completed")

	return nil
}

// aggregateDailyMetrics aggregates metrics for the previous day
func (s *AggregationService) aggregateDailyMetrics(ctx context.Context) error {
	// Get yesterday's date range
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	periodStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	s.logger.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
	}).Debug("Aggregating daily metrics")

	// Check if we already have aggregated data for this period
	var count int64
	s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Where("tenant_id IS NOT NULL AND state_id IS NULL AND metric_type = ?", "daily_storage").
		Where("period_start = ? AND period_end = ?", periodStart, periodEnd).
		Count(&count)

	if count > 0 {
		s.logger.Debug("Daily metrics already aggregated for this period")
		return nil
	}

	// Aggregate per tenant for storage
	if err := s.aggregateTenantStorage(ctx, periodStart, periodEnd, "daily_storage"); err != nil {
		return err
	}

	// Aggregate per tenant for write operations
	if err := s.aggregateTenantOperations(ctx, periodStart, periodEnd, "write_ops", "daily_write_ops"); err != nil {
		return err
	}

	// Aggregate per tenant for read operations
	if err := s.aggregateTenantOperations(ctx, periodStart, periodEnd, "read_ops", "daily_read_ops"); err != nil {
		return err
	}

	return nil
}

// aggregateWeeklyMetrics aggregates metrics for the previous week
func (s *AggregationService) aggregateWeeklyMetrics(ctx context.Context) error {
	// Get last week's date range
	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -7)
	periodStart := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 0, 7)

	s.logger.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
	}).Debug("Aggregating weekly metrics")

	// Check if we already have aggregated data for this period
	var count int64
	s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Where("tenant_id IS NOT NULL AND state_id IS NULL AND metric_type = ?", "weekly_storage").
		Where("period_start = ? AND period_end = ?", periodStart, periodEnd).
		Count(&count)

	if count > 0 {
		s.logger.Debug("Weekly metrics already aggregated for this period")
		return nil
	}

	// Aggregate per tenant for storage
	if err := s.aggregateTenantStorage(ctx, periodStart, periodEnd, "weekly_storage"); err != nil {
		return err
	}

	// Aggregate per tenant for write operations
	if err := s.aggregateTenantOperations(ctx, periodStart, periodEnd, "write_ops", "weekly_write_ops"); err != nil {
		return err
	}

	// Aggregate per tenant for read operations
	if err := s.aggregateTenantOperations(ctx, periodStart, periodEnd, "read_ops", "weekly_read_ops"); err != nil {
		return err
	}

	return nil
}

// aggregateMonthlyMetrics aggregates metrics for the previous month
func (s *AggregationService) aggregateMonthlyMetrics(ctx context.Context) error {
	// Get last month's date range
	now := time.Now().UTC()
	monthStart := now.AddDate(0, -1, 0)
	periodStart := time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	// End of last month
	periodEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	s.logger.WithFields(logrus.Fields{
		"period_start": periodStart,
		"period_end":   periodEnd,
	}).Debug("Aggregating monthly metrics")

	// Check if we already have aggregated data for this period
	var count int64
	s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Where("tenant_id IS NOT NULL AND state_id IS NULL AND metric_type = ?", "monthly_storage").
		Where("period_start = ? AND period_end = ?", periodStart, periodEnd).
		Count(&count)

	if count > 0 {
		s.logger.Debug("Monthly metrics already aggregated for this period")
		return nil
	}

	// Aggregate per tenant for storage
	if err := s.aggregateTenantStorage(ctx, periodStart, periodEnd, "monthly_storage"); err != nil {
		return err
	}

	// Aggregate per tenant for write operations
	if err := s.aggregateTenantOperations(ctx, periodStart, periodEnd, "write_ops", "monthly_write_ops"); err != nil {
		return err
	}

	// Aggregate per tenant for read operations
	if err := s.aggregateTenantOperations(ctx, periodStart, periodEnd, "read_ops", "monthly_read_ops"); err != nil {
		return err
	}

	return nil
}

// aggregateTenantStorage aggregates storage metrics per tenant
func (s *AggregationService) aggregateTenantStorage(ctx context.Context, periodStart, periodEnd time.Time, metricType string) error {
	// Query to aggregate storage by tenant
	type StorageAgg struct {
		TenantID uuid.UUID
		Total    int64
	}

	var results []StorageAgg
	err := s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Select("tenant_id, SUM(value) as total").
		Where("metric_type = ?", "storage").
		Where("period_start >= ? AND period_end <= ?", periodStart, periodEnd).
		Group("tenant_id").
		Scan(&results).Error

	if err != nil {
		return err
	}

	// Insert aggregated metrics
	for _, r := range results {
		metric := &state.StateUsageMetric{
			ID:          uuid.New(),
			TenantID:    r.TenantID,
			StateID:     nil,
			MetricType:  metricType,
			Value:       r.Total,
			Unit:        "bytes",
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(metric).Error; err != nil {
			s.logger.WithError(err).WithField("tenant_id", r.TenantID).Error("Failed to insert aggregated storage metric")
		}
	}

	s.logger.WithField("count", len(results)).Debug("Aggregated storage metrics")

	return nil
}

// aggregateTenantOperations aggregates operation metrics per tenant
func (s *AggregationService) aggregateTenantOperations(ctx context.Context, periodStart, periodEnd time.Time, sourceMetricType, targetMetricType string) error {
	type OpsAgg struct {
		TenantID uuid.UUID
		Total    int64
	}

	var results []OpsAgg
	err := s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Select("tenant_id, SUM(value) as total").
		Where("metric_type = ?", sourceMetricType).
		Where("period_start >= ? AND period_end <= ?", periodStart, periodEnd).
		Group("tenant_id").
		Scan(&results).Error

	if err != nil {
		return err
	}

	// Insert aggregated metrics
	for _, r := range results {
		metric := &state.StateUsageMetric{
			ID:          uuid.New(),
			TenantID:    r.TenantID,
			StateID:     nil,
			MetricType:  targetMetricType,
			Value:       r.Total,
			Unit:        "ops",
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(metric).Error; err != nil {
			s.logger.WithError(err).WithField("tenant_id", r.TenantID).Error("Failed to insert aggregated operation metric")
		}
	}

	s.logger.WithField("count", len(results)).Debug("Aggregated operation metrics")

	return nil
}

// generateTenantSummaries generates summary statistics for each tenant
func (s *AggregationService) generateTenantSummaries(ctx context.Context) error {
	// Get all active tenants from the metrics table
	type TenantAgg struct {
		TenantID      uuid.UUID
		TotalStorage  int64
		TotalWriteOps int64
		TotalReadOps  int64
		ActiveStates  int64
	}

	var results []TenantAgg
	err := s.db.WithContext(ctx).Raw(`
		WITH latest_storage AS (
			SELECT tenant_id, value as storage_bytes
			FROM state_usage_metrics
			WHERE metric_type = 'monthly_storage'
			AND period_start = (
				SELECT MAX(period_start)
				FROM state_usage_metrics
				WHERE metric_type = 'monthly_storage'
				AND tenant_id = state_usage_metrics.tenant_id
			)
		),
		write_ops AS (
			SELECT tenant_id, SUM(value) as total_write
			FROM state_usage_metrics
			WHERE metric_type IN ('write_ops', 'daily_write_ops')
			AND period_start >= NOW() - INTERVAL '30 days'
			GROUP BY tenant_id
		),
		read_ops AS (
			SELECT tenant_id, SUM(value) as total_read
			FROM state_usage_metrics
			WHERE metric_type IN ('read_ops', 'daily_read_ops')
			AND period_start >= NOW() - INTERVAL '30 days'
			GROUP BY tenant_id
		),
		active_states AS (
			SELECT DISTINCT tenant_id, COUNT(DISTINCT state_id) as states
			FROM state_usage_metrics
			WHERE state_id IS NOT NULL
			AND period_start >= NOW() - INTERVAL '7 days'
			GROUP BY tenant_id
		)
		SELECT
			COALESCE(ls.tenant_id, w.tenant_id, r.tenant_id, a.tenant_id) as tenant_id,
			COALESCE(ls.storage_bytes, 0) as total_storage,
			COALESCE(w.total_write, 0) as total_write_ops,
			COALESCE(r.total_read, 0) as total_read_ops,
			COALESCE(a.states, 0) as active_states
		FROM latest_storage ls
		FULL OUTER JOIN write_ops w ON ls.tenant_id = w.tenant_id
		FULL OUTER JOIN read_ops r ON COALESCE(ls.tenant_id, w.tenant_id) = r.tenant_id
		FULL OUTER JOIN active_states a ON COALESCE(COALESCE(ls.tenant_id, w.tenant_id), r.tenant_id) = a.tenant_id
	`).Scan(&results).Error

	if err != nil {
		s.logger.WithError(err).Error("Failed to generate tenant summaries")
		return err
	}

	s.logger.WithField("tenant_count", len(results)).Debug("Generated tenant summaries")

	// Log summary for each tenant )
	for _, r := range results {
		s.logger.WithFields(logrus.Fields{
			"tenant_id":       r.TenantID,
			"total_storage":   r.TotalStorage,
			"total_write_ops": r.TotalWriteOps,
			"total_read_ops":  r.TotalReadOps,
			"active_states":   r.ActiveStates,
		}).Debug("Tenant usage summary")
	}

	return nil
}

// GetTenantUsageSummary retrieves the current usage summary for a specific tenant
func (s *AggregationService) GetTenantUsageSummary(ctx context.Context, tenantID uuid.UUID) (*TenantUsageSummary, error) {
	summary := &TenantUsageSummary{
		TenantID: tenantID,
	}

	// Get latest monthly storage
	var storageMetric state.StateUsageMetric
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND metric_type = ?", tenantID, "monthly_storage").
		Order("period_start DESC").First(&storageMetric).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err == nil {
		summary.TotalStorageBytes = storageMetric.Value
		summary.PeriodStart = storageMetric.PeriodStart
		summary.PeriodEnd = storageMetric.PeriodEnd
	}

	// Get total write operations (last 30 days)
	var writeOps int64
	s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Where("tenant_id = ? AND metric_type IN ('write_ops', 'daily_write_ops')", tenantID).
		Where("period_start >= ?", time.Now().UTC().AddDate(0, 0, -30)).
		Select("COALESCE(SUM(value), 0)").
		Scan(&writeOps)
	summary.TotalWriteOps = writeOps

	// Get total read operations (last 30 days)
	var readOps int64
	s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Where("tenant_id = ? AND metric_type IN ('read_ops', 'daily_read_ops')", tenantID).
		Where("period_start >= ?", time.Now().UTC().AddDate(0, 0, -30)).
		Select("COALESCE(SUM(value), 0)").
		Scan(&readOps)
	summary.TotalReadOps = readOps

	// Get active states count (last 7 days)
	s.db.WithContext(ctx).Model(&state.StateUsageMetric{}).
		Where("tenant_id = ? AND state_id IS NOT NULL", tenantID).
		Where("period_start >= ?", time.Now().UTC().AddDate(0, 0, -7)).
		Distinct("state_id").
		Count(&summary.ActiveStates)

	return summary, nil
}
