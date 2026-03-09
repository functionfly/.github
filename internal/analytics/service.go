package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides analytics business logic
type Service struct {
	repo          *Repository
	db            *gorm.DB
	agentID       string
	metricsBuffer []MetricRecord
	bufferMutex   sync.Mutex
	bufferSize    int
	flushInterval time.Duration
	stopCh        chan struct{}
}

// ServiceConfig holds configuration for the analytics service
type ServiceConfig struct {
	AgentID       string
	BufferSize    int
	FlushInterval time.Duration
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig(agentID string) ServiceConfig {
	return ServiceConfig{
		AgentID:       agentID,
		BufferSize:    100,
		FlushInterval: 5 * time.Second,
	}
}

// NewService creates a new analytics service
func NewService(db *gorm.DB, config ServiceConfig) *Service {
	s := &Service{
		repo:          NewRepository(db),
		db:            db,
		agentID:       config.AgentID,
		metricsBuffer: make([]MetricRecord, 0, config.BufferSize),
		bufferSize:    config.BufferSize,
		flushInterval: config.FlushInterval,
		stopCh:        make(chan struct{}),
	}

	// Start background flush goroutine
	go s.backgroundFlush()

	return s
}

// AutoMigrate runs database migrations
func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.repo.AutoMigrate(ctx)
}

// Stop stops the background flush goroutine
func (s *Service) Stop() {
	close(s.stopCh)
	// Final flush
	s.Flush(context.Background())
}

// RecordMetric records a single metric
func (s *Service) RecordMetric(ctx context.Context, record MetricRecord) error {
	if record.AgentID == "" {
		record.AgentID = s.agentID
	}
	return s.repo.RecordMetric(ctx, record)
}

// RecordMetricBuffered adds a metric to the buffer for batch processing
func (s *Service) RecordMetricBuffered(record MetricRecord) {
	if record.AgentID == "" {
		record.AgentID = s.agentID
	}

	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	s.metricsBuffer = append(s.metricsBuffer, record)

	// Flush if buffer is full
	if len(s.metricsBuffer) >= s.bufferSize {
		go s.Flush(context.Background())
	}
}

// Flush writes all buffered metrics to the database
func (s *Service) Flush(ctx context.Context) error {
	s.bufferMutex.Lock()
	if len(s.metricsBuffer) == 0 {
		s.bufferMutex.Unlock()
		return nil
	}

	// Swap buffer
	metrics := s.metricsBuffer
	s.metricsBuffer = make([]MetricRecord, 0, s.bufferSize)
	s.bufferMutex.Unlock()

	return s.repo.RecordMetrics(ctx, metrics)
}

// backgroundFlush periodically flushes the metrics buffer
func (s *Service) backgroundFlush() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Flush(context.Background()); err != nil {
				logrus.WithError(err).Warn("failed to flush metrics buffer")
			}
		case <-s.stopCh:
			return
		}
	}
}

// RecordFactoryRun records metrics for a completed factory run
func (s *Service) RecordFactoryRun(ctx context.Context, runID uuid.UUID, agentID string, metrics RunMetrics) error {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeLatencyTotal,
			MetricValue: metrics.TotalLatencyMs,
			Labels:      map[string]any{"unit": "ms"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeLatencyGeneration,
			MetricValue: metrics.GenerationLatencyMs,
			Labels:      map[string]any{"unit": "ms"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeLatencyTesting,
			MetricValue: metrics.TestingLatencyMs,
			Labels:      map[string]any{"unit": "ms"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeLatencyPublishing,
			MetricValue: metrics.PublishingLatencyMs,
			Labels:      map[string]any{"unit": "ms"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeQualityScore,
			MetricValue: metrics.AvgQualityScore,
			Labels:      map[string]any{"unit": "score"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeTestScore,
			MetricValue: metrics.AvgTestScore,
			Labels:      map[string]any{"unit": "score"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeOpportunityScanned,
			MetricValue: float64(metrics.OpportunitiesScanned),
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeFunctionPublished,
			MetricValue: float64(metrics.FunctionsPublished),
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		},
	}

	// Record success or failure
	if metrics.Success {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeGenerationSuccess,
			MetricValue: 1,
			Labels:      map[string]any{"status": "success"},
			Timestamp:   &now,
		})
	} else {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeGenerationFailure,
			MetricValue: 1,
			Labels:      map[string]any{"status": "failed", "error": metrics.ErrorMessage},
			Timestamp:   &now,
		})
	}

	// Record review required if applicable
	if metrics.ReviewRequired {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     agentID,
			MetricType:  MetricTypeReviewRequired,
			MetricValue: 1,
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		})
	}

	return s.repo.RecordMetrics(ctx, records)
}

// RunMetrics holds metrics for a factory run
type RunMetrics struct {
	Success              bool
	TotalLatencyMs       float64
	GenerationLatencyMs  float64
	TestingLatencyMs     float64
	PublishingLatencyMs  float64
	AvgQualityScore      float64
	AvgTestScore         float64
	OpportunitiesScanned int
	FunctionsPublished   int
	ReviewRequired       bool
	ErrorMessage         string
}

// GetDashboardStats retrieves dashboard statistics
func (s *Service) GetDashboardStats(ctx context.Context, period time.Duration) (*DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx, s.agentID, period)
}

// GetDashboardStatsForAgent retrieves dashboard statistics for a specific agent
func (s *Service) GetDashboardStatsForAgent(ctx context.Context, agentID string, period time.Duration) (*DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx, agentID, period)
}

// GetTimeSeries retrieves time series data for a metric
func (s *Service) GetTimeSeries(ctx context.Context, metricType MetricType, period AggregationPeriod, startTime, endTime time.Time) (*TimeSeriesData, error) {
	return s.repo.GetTimeSeries(ctx, s.agentID, metricType, period, startTime, endTime)
}

// GetTimeSeriesForAgent retrieves time series data for a specific agent
func (s *Service) GetTimeSeriesForAgent(ctx context.Context, agentID string, metricType MetricType, period AggregationPeriod, startTime, endTime time.Time) (*TimeSeriesData, error) {
	return s.repo.GetTimeSeries(ctx, agentID, metricType, period, startTime, endTime)
}

// GetRunMetricsSummary retrieves metrics summary for a specific run
func (s *Service) GetRunMetricsSummary(ctx context.Context, runID uuid.UUID) (*RunMetricsSummary, error) {
	return s.repo.GetRunMetricsSummary(ctx, runID)
}

// GetMetrics retrieves metrics based on filter
func (s *Service) GetMetrics(ctx context.Context, filter MetricFilter) ([]FactoryMetric, int64, error) {
	if filter.AgentID == "" {
		filter.AgentID = s.agentID
	}
	return s.repo.GetMetrics(ctx, filter)
}

// GetAggregatedMetrics retrieves pre-computed aggregated metrics
func (s *Service) GetAggregatedMetrics(ctx context.Context, filter MetricFilter) ([]AggregatedMetric, error) {
	if filter.AgentID == "" {
		filter.AgentID = s.agentID
	}
	return s.repo.GetAggregatedMetrics(ctx, filter)
}

// AggregateMetrics computes and stores aggregated metrics
func (s *Service) AggregateMetrics(ctx context.Context, period AggregationPeriod, periodStart time.Time) error {
	return s.repo.AggregateMetrics(ctx, s.agentID, period, periodStart)
}

// AggregateAllAgents aggregates metrics for all agents
func (s *Service) AggregateAllAgents(ctx context.Context, period AggregationPeriod, periodStart time.Time) error {
	// Get all unique agent IDs
	var agentIDs []string
	if err := s.db.WithContext(ctx).Model(&FactoryMetric{}).
		Distinct("agent_id").
		Pluck("agent_id", &agentIDs).Error; err != nil {
		return fmt.Errorf("get agent IDs: %w", err)
	}

	var lastErr error
	for _, agentID := range agentIDs {
		if err := s.repo.AggregateMetrics(ctx, agentID, period, periodStart); err != nil {
			logrus.WithError(err).WithField("agent_id", agentID).Warn("failed to aggregate metrics for agent")
			lastErr = err
		}
	}
	return lastErr
}

// GetQualityTrend calculates quality trend
func (s *Service) GetQualityTrend(ctx context.Context, period time.Duration) (float64, error) {
	now := time.Now().UTC()
	currentStart := now.Add(-period)
	previousStart := currentStart.Add(-period)
	return s.repo.GetQualityTrend(ctx, s.agentID, currentStart, previousStart, period)
}

// CleanupOldMetrics removes old metrics based on retention policy
func (s *Service) CleanupOldMetrics(ctx context.Context, retentionDays int) (int64, error) {
	return s.repo.CleanupOldMetrics(ctx, retentionDays)
}

// GetMetricPercentiles calculates percentiles for a metric
func (s *Service) GetMetricPercentiles(ctx context.Context, metricType MetricType, startTime, endTime time.Time) (p50, p95, p99 float64, err error) {
	return s.repo.GetMetricPercentiles(ctx, s.agentID, metricType, startTime, endTime)
}

// GetSuccessRate calculates the success rate for a time period
func (s *Service) GetSuccessRate(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	var stats struct {
		Total   int64
		Success int64
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE metric_type = 'generation_success') as success
		FROM factory_analytics_metrics
		WHERE agent_id = ? AND created_at >= ? AND created_at <= ?
		AND metric_type IN ('generation_success', 'generation_failure')
	`, s.agentID, startTime, endTime).Scan(&stats).Error

	if err != nil {
		return 0, err
	}

	if stats.Total == 0 {
		return 0, nil
	}

	return float64(stats.Success) / float64(stats.Total) * 100, nil
}

// GetThroughput calculates throughput (functions per hour) for a time period
func (s *Service) GetThroughput(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&FactoryMetric{}).
		Where("agent_id = ? AND metric_type = ? AND created_at >= ? AND created_at <= ?",
			s.agentID, MetricTypeFunctionPublished, startTime, endTime).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	duration := endTime.Sub(startTime).Hours()
	if duration <= 0 {
		return 0, nil
	}

	return float64(count) / duration, nil
}

// GetErrorRate calculates the error rate for a time period
func (s *Service) GetErrorRate(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	var stats struct {
		Total  int64
		Errors int64
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE metric_type = 'generation_failure') as errors
		FROM factory_analytics_metrics
		WHERE agent_id = ? AND created_at >= ? AND created_at <= ?
		AND metric_type IN ('generation_success', 'generation_failure')
	`, s.agentID, startTime, endTime).Scan(&stats).Error

	if err != nil {
		return 0, err
	}

	if stats.Total == 0 {
		return 0, nil
	}

	return float64(stats.Errors) / float64(stats.Total) * 100, nil
}

// GetAverageLatency calculates average latency for a time period
func (s *Service) GetAverageLatency(ctx context.Context, latencyType MetricType, startTime, endTime time.Time) (float64, error) {
	var avg float64
	err := s.db.WithContext(ctx).Model(&FactoryMetric{}).
		Where("agent_id = ? AND metric_type = ? AND created_at >= ? AND created_at <= ?",
			s.agentID, latencyType, startTime, endTime).
		Select("AVG(metric_value)").
		Scan(&avg).Error

	return avg, err
}

// GetRecentRuns retrieves recent factory runs with their metrics
func (s *Service) GetRecentRuns(ctx context.Context, limit int) ([]RunMetricsSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var runIDs []uuid.UUID
	err := s.db.WithContext(ctx).Raw(`
		SELECT id FROM factory_runs
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, s.agentID, limit).Scan(&runIDs).Error

	if err != nil {
		return nil, err
	}

	summaries := make([]RunMetricsSummary, 0, len(runIDs))
	for _, runID := range runIDs {
		summary, err := s.repo.GetRunMetricsSummary(ctx, runID)
		if err != nil {
			logrus.WithError(err).WithField("run_id", runID).Warn("failed to get run summary")
			continue
		}
		summaries = append(summaries, *summary)
	}

	return summaries, nil
}

// GetHourlyStats retrieves hourly statistics for the last N hours
func (s *Service) GetHourlyStats(ctx context.Context, hours int) ([]AggregatedMetric, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 168 {
		hours = 168 // Max 1 week
	}

	startTime := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	filter := MetricFilter{
		AgentID:   s.agentID,
		Period:    AggregationPeriodHourly,
		StartTime: &startTime,
		Limit:     hours,
	}
	return s.repo.GetAggregatedMetrics(ctx, filter)
}

// GetDailyStats retrieves daily statistics for the last N days
func (s *Service) GetDailyStats(ctx context.Context, days int) ([]AggregatedMetric, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90 // Max 3 months
	}

	startTime := time.Now().UTC().AddDate(0, 0, -days)
	filter := MetricFilter{
		AgentID:   s.agentID,
		Period:    AggregationPeriodDaily,
		StartTime: &startTime,
		Limit:     days,
	}
	return s.repo.GetAggregatedMetrics(ctx, filter)
}

// GetWeeklyStats retrieves weekly statistics for the last N weeks
func (s *Service) GetWeeklyStats(ctx context.Context, weeks int) ([]AggregatedMetric, error) {
	if weeks <= 0 {
		weeks = 4
	}
	if weeks > 52 {
		weeks = 52 // Max 1 year
	}

	startTime := time.Now().UTC().AddDate(0, 0, -weeks*7)
	filter := MetricFilter{
		AgentID:   s.agentID,
		Period:    AggregationPeriodWeekly,
		StartTime: &startTime,
		Limit:     weeks,
	}
	return s.repo.GetAggregatedMetrics(ctx, filter)
}

// GetMonthlyStats retrieves monthly statistics for the last N months
func (s *Service) GetMonthlyStats(ctx context.Context, months int) ([]AggregatedMetric, error) {
	if months <= 0 {
		months = 6
	}
	if months > 24 {
		months = 24 // Max 2 years
	}

	startTime := time.Now().UTC().AddDate(0, -months, 0)
	filter := MetricFilter{
		AgentID:   s.agentID,
		Period:    AggregationPeriodMonthly,
		StartTime: &startTime,
		Limit:     months,
	}
	return s.repo.GetAggregatedMetrics(ctx, filter)
}

// RunAggregationJob runs aggregation for all periods
func (s *Service) RunAggregationJob(ctx context.Context) error {
	now := time.Now().UTC()

	// Aggregate hourly (last 24 hours)
	for i := 0; i < 24; i++ {
		hourStart := now.Add(-time.Duration(i+1) * time.Hour).Truncate(time.Hour)
		if err := s.AggregateAllAgents(ctx, AggregationPeriodHourly, hourStart); err != nil {
			logrus.WithError(err).WithField("hour", hourStart).Warn("failed to aggregate hourly metrics")
		}
	}

	// Aggregate daily (last 7 days)
	for i := 0; i < 7; i++ {
		dayStart := now.AddDate(0, 0, -(i + 1)).Truncate(24 * time.Hour)
		if err := s.AggregateAllAgents(ctx, AggregationPeriodDaily, dayStart); err != nil {
			logrus.WithError(err).WithField("day", dayStart).Warn("failed to aggregate daily metrics")
		}
	}

	// Aggregate weekly (last 4 weeks)
	for i := 0; i < 4; i++ {
		weekStart := now.AddDate(0, 0, -(i+1)*7).Truncate(24 * time.Hour)
		if err := s.AggregateAllAgents(ctx, AggregationPeriodWeekly, weekStart); err != nil {
			logrus.WithError(err).WithField("week", weekStart).Warn("failed to aggregate weekly metrics")
		}
	}

	// Aggregate monthly (last 6 months)
	for i := 0; i < 6; i++ {
		monthStart := now.AddDate(0, -(i + 1), 0).Truncate(24 * time.Hour)
		if err := s.AggregateAllAgents(ctx, AggregationPeriodMonthly, monthStart); err != nil {
			logrus.WithError(err).WithField("month", monthStart).Warn("failed to aggregate monthly metrics")
		}
	}

	return nil
}
