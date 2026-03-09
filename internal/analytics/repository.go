package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Repository provides data access for analytics
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new analytics repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// AutoMigrate runs database migrations for analytics tables
func (r *Repository) AutoMigrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&FactoryMetric{}, &AggregatedMetric{})
}

// RecordMetric records a single metric data point
func (r *Repository) RecordMetric(ctx context.Context, record MetricRecord) error {
	metric := &FactoryMetric{
		ID:          uuid.New(),
		RunID:       record.RunID,
		AgentID:     record.AgentID,
		MetricType:  record.MetricType,
		MetricValue: record.MetricValue,
		Labels:      record.Labels,
	}
	if record.Labels == nil {
		metric.Labels = make(map[string]any)
	}
	if record.Timestamp != nil {
		metric.CreatedAt = *record.Timestamp
	} else {
		metric.CreatedAt = time.Now().UTC()
	}

	return r.db.WithContext(ctx).Create(metric).Error
}

// RecordMetrics records multiple metrics in a batch
func (r *Repository) RecordMetrics(ctx context.Context, records []MetricRecord) error {
	if len(records) == 0 {
		return nil
	}

	metrics := make([]FactoryMetric, len(records))
	now := time.Now().UTC()
	for i, record := range records {
		labels := record.Labels
		if labels == nil {
			labels = make(map[string]any)
		}
		ts := now
		if record.Timestamp != nil {
			ts = *record.Timestamp
		}
		metrics[i] = FactoryMetric{
			ID:          uuid.New(),
			RunID:       record.RunID,
			AgentID:     record.AgentID,
			MetricType:  record.MetricType,
			MetricValue: record.MetricValue,
			Labels:      labels,
			CreatedAt:   ts,
		}
	}

	return r.db.WithContext(ctx).CreateInBatches(metrics, 100).Error
}

// GetMetrics retrieves metrics based on filter criteria
func (r *Repository) GetMetrics(ctx context.Context, filter MetricFilter) ([]FactoryMetric, int64, error) {
	query := r.db.WithContext(ctx).Model(&FactoryMetric{})

	if filter.AgentID != "" {
		query = query.Where("agent_id = ?", filter.AgentID)
	}
	if filter.MetricType != "" {
		query = query.Where("metric_type = ?", filter.MetricType)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count metrics: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	var metrics []FactoryMetric
	if err := query.Order("created_at DESC").Limit(limit).Offset(filter.Offset).Find(&metrics).Error; err != nil {
		return nil, 0, fmt.Errorf("find metrics: %w", err)
	}

	return metrics, total, nil
}

// GetAggregatedMetrics retrieves pre-computed aggregated metrics
func (r *Repository) GetAggregatedMetrics(ctx context.Context, filter MetricFilter) ([]AggregatedMetric, error) {
	query := r.db.WithContext(ctx).Model(&AggregatedMetric{})

	if filter.AgentID != "" {
		query = query.Where("agent_id = ?", filter.AgentID)
	}
	if filter.MetricType != "" {
		query = query.Where("metric_type = ?", filter.MetricType)
	}
	if filter.Period != "" {
		query = query.Where("period = ?", filter.Period)
	}
	if filter.StartTime != nil {
		query = query.Where("period_start >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("period_start <= ?", *filter.EndTime)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	var metrics []AggregatedMetric
	if err := query.Order("period_start DESC").Limit(limit).Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("find aggregated metrics: %w", err)
	}

	return metrics, nil
}

// GetDashboardStats retrieves real-time dashboard statistics
func (r *Repository) GetDashboardStats(ctx context.Context, agentID string, period time.Duration) (*DashboardStats, error) {
	now := time.Now().UTC()
	startTime := now.Add(-period)

	stats := &DashboardStats{
		PeriodStart: startTime,
		PeriodEnd:   now,
		LastUpdated: now,
	}

	// Get run statistics
	var runStats struct {
		Total      int64
		Successful int64
		Failed     int64
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'succeeded') as successful,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM factory_runs
		WHERE agent_id = ? AND created_at >= ?
	`, agentID, startTime).Scan(&runStats).Error; err != nil {
		logrus.WithError(err).Warn("failed to get run stats")
	}

	stats.TotalRuns = runStats.Total
	stats.SuccessfulRuns = runStats.Successful
	stats.FailedRuns = runStats.Failed
	if runStats.Total > 0 {
		stats.SuccessRate = float64(runStats.Successful) / float64(runStats.Total) * 100
	}

	// Get quality and test scores
	var scoreStats struct {
		AvgQuality float64
		AvgTest    float64
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(quality_score), 0) as avg_quality,
			COALESCE(AVG(test_score), 0) as avg_test
		FROM factory_versions
		WHERE created_at >= ?
	`, startTime).Scan(&scoreStats).Error; err != nil {
		logrus.WithError(err).Warn("failed to get score stats")
	}
	stats.AvgQualityScore = scoreStats.AvgQuality
	stats.AvgTestScore = scoreStats.AvgTest

	// Get throughput statistics
	var throughputStats struct {
		Generated int64
		Published int64
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as generated,
			COUNT(*) FILTER (WHERE NOT review_required) as published
		FROM factory_versions
		WHERE created_at >= ?
	`, startTime).Scan(&throughputStats).Error; err != nil {
		logrus.WithError(err).Warn("failed to get throughput stats")
	}
	stats.FunctionsGenerated = throughputStats.Generated
	stats.FunctionsPublished = throughputStats.Published

	hours := period.Hours()
	if hours > 0 {
		stats.ThroughputPerHour = float64(throughputStats.Generated) / hours
	}

	// Get latency metrics from recorded metrics
	var latencyStats struct {
		AvgGeneration float64
		AvgTesting    float64
		AvgPublishing float64
		AvgTotal      float64
		P95Total      float64
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(metric_value) FILTER (WHERE metric_type = 'latency_generation'), 0) as avg_generation,
			COALESCE(AVG(metric_value) FILTER (WHERE metric_type = 'latency_testing'), 0) as avg_testing,
			COALESCE(AVG(metric_value) FILTER (WHERE metric_type = 'latency_publishing'), 0) as avg_publishing,
			COALESCE(AVG(metric_value) FILTER (WHERE metric_type = 'latency_total'), 0) as avg_total,
			COALESCE((
				SELECT percentile_cont(0.95) WITHIN GROUP (ORDER BY metric_value)
				FROM factory_analytics_metrics
				WHERE metric_type = 'latency_total' AND agent_id = ? AND created_at >= ?
			), 0) as p95_total
		FROM factory_analytics_metrics
		WHERE agent_id = ? AND created_at >= ?
	`, agentID, startTime, agentID, startTime).Scan(&latencyStats).Error; err != nil {
		logrus.WithError(err).Warn("failed to get latency stats")
	}
	stats.AvgGenerationLatency = latencyStats.AvgGeneration
	stats.AvgTestingLatency = latencyStats.AvgTesting
	stats.AvgPublishingLatency = latencyStats.AvgPublishing
	stats.AvgTotalLatency = latencyStats.AvgTotal
	stats.P95Latency = latencyStats.P95Total

	// Get error rate
	var errorStats struct {
		Total    int64
		Errors   int64
		LastTime *time.Time
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE metric_type = 'generation_failure') as errors,
			MAX(created_at) FILTER (WHERE metric_type = 'generation_failure') as last_time
		FROM factory_analytics_metrics
		WHERE agent_id = ? AND created_at >= ?
		AND metric_type IN ('generation_success', 'generation_failure')
	`, agentID, startTime).Scan(&errorStats).Error; err != nil {
		logrus.WithError(err).Warn("failed to get error stats")
	}
	if errorStats.Total > 0 {
		stats.ErrorRate = float64(errorStats.Errors) / float64(errorStats.Total) * 100
	}
	stats.LastErrorTime = errorStats.LastTime

	// Get review statistics
	var reviewStats struct {
		Pending int64
		Total   int64
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE review_required) as pending,
			COUNT(*) as total
		FROM factory_versions
		WHERE created_at >= ?
	`, startTime).Scan(&reviewStats).Error; err != nil {
		logrus.WithError(err).Warn("failed to get review stats")
	}
	stats.PendingReviews = reviewStats.Pending
	if reviewStats.Total > 0 {
		stats.ReviewRate = float64(reviewStats.Pending) / float64(reviewStats.Total) * 100
	}

	return stats, nil
}

// GetTimeSeries retrieves time series data for a specific metric
func (r *Repository) GetTimeSeries(ctx context.Context, agentID string, metricType MetricType, period AggregationPeriod, startTime, endTime time.Time) (*TimeSeriesData, error) {
	// Try to get pre-aggregated data first
	var aggregated []AggregatedMetric
	if err := r.db.WithContext(ctx).Model(&AggregatedMetric{}).
		Where("agent_id = ? AND metric_type = ? AND period = ? AND period_start >= ? AND period_start <= ?",
			agentID, metricType, period, startTime, endTime).
		Order("period_start ASC").
		Find(&aggregated).Error; err != nil {
		return nil, fmt.Errorf("find aggregated time series: %w", err)
	}

	// If we have aggregated data, use it
	if len(aggregated) > 0 {
		points := make([]TimeSeriesPoint, len(aggregated))
		for i, agg := range aggregated {
			points[i] = TimeSeriesPoint{
				Timestamp: agg.PeriodStart,
				Value:     agg.Avg,
				Count:     agg.Count,
			}
		}
		return &TimeSeriesData{
			MetricType: metricType,
			Points:     points,
			Period:     period,
		}, nil
	}

	// Fall back to raw metrics
	var metrics []FactoryMetric
	if err := r.db.WithContext(ctx).Model(&FactoryMetric{}).
		Where("agent_id = ? AND metric_type = ? AND created_at >= ? AND created_at <= ?",
			agentID, metricType, startTime, endTime).
		Order("created_at ASC").
		Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("find raw time series: %w", err)
	}

	// Aggregate raw metrics based on period
	points := r.aggregateRawMetrics(metrics, period, startTime, endTime)

	return &TimeSeriesData{
		MetricType: metricType,
		Points:     points,
		Period:     period,
	}, nil
}

// aggregateRawMetrics aggregates raw metrics into time buckets
func (r *Repository) aggregateRawMetrics(metrics []FactoryMetric, period AggregationPeriod, startTime, endTime time.Time) []TimeSeriesPoint {
	if len(metrics) == 0 {
		return []TimeSeriesPoint{}
	}

	// Determine bucket size based on period
	var bucketSize time.Duration
	switch period {
	case AggregationPeriodHourly:
		bucketSize = time.Hour
	case AggregationPeriodDaily:
		bucketSize = 24 * time.Hour
	case AggregationPeriodWeekly:
		bucketSize = 7 * 24 * time.Hour
	case AggregationPeriodMonthly:
		bucketSize = 30 * 24 * time.Hour
	default:
		bucketSize = time.Hour
	}

	// Create buckets
	buckets := make(map[int64][]float64)
	for _, m := range metrics {
		bucket := m.CreatedAt.Truncate(bucketSize).Unix()
		buckets[bucket] = append(buckets[bucket], m.MetricValue)
	}

	// Convert to points
	points := make([]TimeSeriesPoint, 0, len(buckets))
	for bucket, values := range buckets {
		var sum float64
		for _, v := range values {
			sum += v
		}
		avg := sum / float64(len(values))
		points = append(points, TimeSeriesPoint{
			Timestamp: time.Unix(bucket, 0),
			Value:     avg,
			Count:     int64(len(values)),
		})
	}

	// Sort by timestamp
	for i := 0; i < len(points)-1; i++ {
		for j := i + 1; j < len(points); j++ {
			if points[i].Timestamp.After(points[j].Timestamp) {
				points[i], points[j] = points[j], points[i]
			}
		}
	}

	return points
}

// GetRunMetricsSummary retrieves metrics summary for a specific factory run
func (r *Repository) GetRunMetricsSummary(ctx context.Context, runID uuid.UUID) (*RunMetricsSummary, error) {
	var summary RunMetricsSummary

	// Get basic run info
	var run struct {
		ID                   uuid.UUID
		AgentID              string
		Status               string
		OpportunitiesScanned int
		FunctionsGenerated   int
		FunctionsPublished   int
		AverageQualityScore  float64
		CreatedAt            time.Time
		CompletedAt          *time.Time
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT id, agent_id, status, opportunities_scanned, functions_generated,
		       functions_published, average_quality_score, created_at, completed_at
		FROM factory_runs
		WHERE id = ?
	`, runID).Scan(&run).Error; err != nil {
		return nil, fmt.Errorf("get run info: %w", err)
	}

	summary.RunID = run.ID
	summary.AgentID = run.AgentID
	summary.Status = run.Status
	summary.OpportunitiesScanned = run.OpportunitiesScanned
	summary.FunctionsGenerated = run.FunctionsGenerated
	summary.FunctionsPublished = run.FunctionsPublished
	summary.AvgQualityScore = run.AverageQualityScore
	summary.CreatedAt = run.CreatedAt
	summary.CompletedAt = run.CompletedAt

	// Calculate duration
	if run.CompletedAt != nil {
		summary.Duration = float64(run.CompletedAt.Sub(run.CreatedAt).Milliseconds())
	}

	// Get metrics for this run
	var metrics []FactoryMetric
	if err := r.db.WithContext(ctx).Model(&FactoryMetric{}).
		Where("run_id = ?", runID).
		Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("get run metrics: %w", err)
	}

	// Aggregate metrics
	for _, m := range metrics {
		switch m.MetricType {
		case MetricTypeTestScore:
			summary.AvgTestScore = m.MetricValue
		case MetricTypeLatencyGeneration:
			summary.GenerationLatency = m.MetricValue
		case MetricTypeLatencyTesting:
			summary.TestingLatency = m.MetricValue
		case MetricTypeLatencyPublishing:
			summary.PublishingLatency = m.MetricValue
		case MetricTypeLatencyTotal:
			summary.TotalLatency = m.MetricValue
		case MetricTypeGenerationFailure:
			summary.ErrorCount++
		case MetricTypeReviewRequired:
			summary.ReviewRequired++
		}
	}

	return &summary, nil
}

// AggregateMetrics computes and stores aggregated metrics for a time period
func (r *Repository) AggregateMetrics(ctx context.Context, agentID string, period AggregationPeriod, periodStart time.Time) error {
	var periodEnd time.Time
	switch period {
	case AggregationPeriodHourly:
		periodEnd = periodStart.Add(time.Hour)
	case AggregationPeriodDaily:
		periodEnd = periodStart.Add(24 * time.Hour)
	case AggregationPeriodWeekly:
		periodEnd = periodStart.Add(7 * 24 * time.Hour)
	case AggregationPeriodMonthly:
		periodEnd = periodStart.AddDate(0, 1, 0)
	}

	// Get all metric types for this period
	var metricTypes []MetricType
	if err := r.db.WithContext(ctx).Model(&FactoryMetric{}).
		Distinct("metric_type").
		Where("agent_id = ? AND created_at >= ? AND created_at < ?", agentID, periodStart, periodEnd).
		Pluck("metric_type", &metricTypes).Error; err != nil {
		return fmt.Errorf("get metric types: %w", err)
	}

	for _, metricType := range metricTypes {
		var stats struct {
			Count        int64
			Sum          float64
			Avg          float64
			Min          float64
			Max          float64
			P50          float64
			P95          float64
			P99          float64
			SuccessCount int64
			FailureCount int64
		}

		// Calculate basic stats
		if err := r.db.WithContext(ctx).Raw(`
			SELECT
				COUNT(*) as count,
				COALESCE(SUM(metric_value), 0) as sum,
				COALESCE(AVG(metric_value), 0) as avg,
				COALESCE(MIN(metric_value), 0) as min,
				COALESCE(MAX(metric_value), 0) as max,
				COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY metric_value), 0) as p50,
				COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY metric_value), 0) as p95,
				COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY metric_value), 0) as p99
			FROM factory_analytics_metrics
			WHERE agent_id = ? AND metric_type = ? AND created_at >= ? AND created_at < ?
		`, agentID, metricType, periodStart, periodEnd).Scan(&stats).Error; err != nil {
			logrus.WithError(err).Warnf("failed to calculate stats for metric type %s", metricType)
			continue
		}

		// Get success/failure counts for relevant metrics
		if metricType == MetricTypeGenerationSuccess || metricType == MetricTypeGenerationFailure {
			if err := r.db.WithContext(ctx).Raw(`
				SELECT
					COUNT(*) FILTER (WHERE metric_type = 'generation_success') as success_count,
					COUNT(*) FILTER (WHERE metric_type = 'generation_failure') as failure_count
				FROM factory_analytics_metrics
				WHERE agent_id = ? AND created_at >= ? AND created_at < ?
				AND metric_type IN ('generation_success', 'generation_failure')
			`, agentID, periodStart, periodEnd).Scan(&stats).Error; err != nil {
				logrus.WithError(err).Warn("failed to get success/failure counts")
			}
		}

		// Upsert aggregated metric
		aggregated := &AggregatedMetric{
			ID:           uuid.New(),
			AgentID:      agentID,
			Period:       period,
			PeriodStart:  periodStart,
			MetricType:   metricType,
			Count:        stats.Count,
			Sum:          stats.Sum,
			Avg:          stats.Avg,
			Min:          stats.Min,
			Max:          stats.Max,
			P50:          stats.P50,
			P95:          stats.P95,
			P99:          stats.P99,
			SuccessCount: stats.SuccessCount,
			FailureCount: stats.FailureCount,
		}

		// Use upsert to handle duplicates
		if err := r.db.WithContext(ctx).
			Where("agent_id = ? AND period = ? AND period_start = ? AND metric_type = ?",
				agentID, period, periodStart, metricType).
			Assign(map[string]any{
				"count":         stats.Count,
				"sum":           stats.Sum,
				"avg":           stats.Avg,
				"min":           stats.Min,
				"max":           stats.Max,
				"p50":           stats.P50,
				"p95":           stats.P95,
				"p99":           stats.P99,
				"success_count": stats.SuccessCount,
				"failure_count": stats.FailureCount,
				"updated_at":    time.Now().UTC(),
			}).
			FirstOrCreate(aggregated).Error; err != nil {
			logrus.WithError(err).Warnf("failed to upsert aggregated metric for %s", metricType)
		}
	}

	return nil
}

// GetQualityTrend calculates the quality trend compared to the previous period
func (r *Repository) GetQualityTrend(ctx context.Context, agentID string, currentStart, previousStart time.Time, period time.Duration) (float64, error) {
	var currentAvg, previousAvg float64

	// Get current period average
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(quality_score), 0)
		FROM factory_versions fv
		JOIN factory_runs fr ON fv.run_id = fr.id
		WHERE fr.agent_id = ? AND fv.created_at >= ?
	`, agentID, currentStart).Scan(&currentAvg).Error; err != nil {
		return 0, err
	}

	// Get previous period average
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(quality_score), 0)
		FROM factory_versions fv
		JOIN factory_runs fr ON fv.run_id = fr.id
		WHERE fr.agent_id = ? AND fv.created_at >= ? AND fv.created_at < ?
	`, agentID, previousStart, currentStart).Scan(&previousAvg).Error; err != nil {
		return 0, err
	}

	// Calculate percentage change
	if previousAvg == 0 {
		return 0, nil
	}
	return ((currentAvg - previousAvg) / previousAvg) * 100, nil
}

// CleanupOldMetrics removes metrics older than the retention period
func (r *Repository) CleanupOldMetrics(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)

	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&FactoryMetric{})

	return result.RowsAffected, result.Error
}

// GetMetricPercentiles calculates percentiles for a metric over a time period
func (r *Repository) GetMetricPercentiles(ctx context.Context, agentID string, metricType MetricType, startTime, endTime time.Time) (p50, p95, p99 float64, err error) {
	var stats struct {
		P50 float64
		P95 float64
		P99 float64
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY metric_value), 0) as p50,
			COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY metric_value), 0) as p95,
			COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY metric_value), 0) as p99
		FROM factory_analytics_metrics
		WHERE agent_id = ? AND metric_type = ? AND created_at >= ? AND created_at <= ?
	`, agentID, metricType, startTime, endTime).Scan(&stats).Error

	return stats.P50, stats.P95, stats.P99, err
}

// GetDB returns the underlying database connection for raw queries
func (r *Repository) GetDB() *sql.DB {
	db, err := r.db.DB()
	if err != nil {
		logrus.WithError(err).Error("failed to get underlying DB connection")
		return nil
	}
	return db
}
