package swarm

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/agent/discovery"
	factorysvc "github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MetricsCollector struct {
	db             *gorm.DB
	factoryService *factorysvc.Service
	discoverySvc   *discovery.Service
}

func NewMetricsCollector(db *gorm.DB, factoryService *factorysvc.Service, discoverySvc *discovery.Service) *MetricsCollector {
	return &MetricsCollector{
		db:             db,
		factoryService: factoryService,
		discoverySvc:   discoverySvc,
	}
}

func (mc *MetricsCollector) CollectDailyMetrics(ctx context.Context) (*DailySwarmMetrics, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.Add(-24 * time.Hour)

	metrics := &DailySwarmMetrics{
		Date: today,
	}

	var runs []factorysvc.FactoryRun
	if err := mc.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Find(&runs).Error; err != nil {
		return nil, err
	}

	metrics.TotalRuns = len(runs)
	for _, run := range runs {
		if run.Status == factorysvc.RunStatusSucceeded {
			metrics.SuccessfulRuns++
			metrics.FunctionsGenerated += run.FunctionsGenerated
			metrics.FunctionsPublished += run.FunctionsPublished
			metrics.TotalQualityScore += run.AverageQualityScore
		} else if run.Status == factorysvc.RunStatusFailed {
			metrics.FailedRuns++
		}
	}

	if metrics.SuccessfulRuns > 0 {
		metrics.AverageQualityScore = metrics.TotalQualityScore / float64(metrics.SuccessfulRuns)
	}

	var opps []discovery.Opportunity
	if err := mc.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Find(&opps).Error; err != nil {
		return nil, err
	}

	for _, opp := range opps {
		switch opp.Status {
		case discovery.OpportunityStatusQualified:
			metrics.OpportunitiesDiscovered++
		case discovery.OpportunityStatusGenerated:
			metrics.OpportunitiesGenerated++
		case discovery.OpportunityStatusPublished:
			metrics.OpportunitiesPublished++
		case discovery.OpportunityStatusRejected:
			metrics.OpportunitiesRejected++
		}
	}

	if err := mc.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Model(&discovery.Opportunity{}).
		Count(&metrics.NewOpportunities).Error; err != nil {
		return nil, err
	}

	metrics.ConversionRate = calculateConversionRate(metrics.OpportunitiesDiscovered, metrics.OpportunitiesPublished)

	return metrics, nil
}

func (mc *MetricsCollector) GetWeeklyMetrics(ctx context.Context) (*WeeklySwarmMetrics, error) {
	now := time.Now()
	weekAgo := now.Add(-7 * 24 * time.Hour)

	metrics := &WeeklySwarmMetrics{
		StartDate: weekAgo,
		EndDate:   now,
	}

	var runs []factorysvc.FactoryRun
	if err := mc.db.WithContext(ctx).
		Where("created_at >= ?", weekAgo).
		Find(&runs).Error; err != nil {
		return nil, err
	}

	dailyBreakdown := make([]DailySwarmMetrics, 0, 7)
	for i := 0; i < 7; i++ {
		day := weekAgo.Add(time.Duration(i) * 24 * time.Hour)
		dayMetrics := &DailySwarmMetrics{Date: day}

		for _, run := range runs {
			runDay := time.Date(run.CreatedAt.Year(), run.CreatedAt.Month(), run.CreatedAt.Day(), 0, 0, 0, 0, time.UTC)
			if runDay.Equal(day) {
				dayMetrics.TotalRuns++
				if run.Status == factorysvc.RunStatusSucceeded {
					dayMetrics.SuccessfulRuns++
					dayMetrics.FunctionsGenerated += run.FunctionsGenerated
					dayMetrics.FunctionsPublished += run.FunctionsPublished
				} else if run.Status == factorysvc.RunStatusFailed {
					dayMetrics.FailedRuns++
				}
			}
		}

		dailyBreakdown = append(dailyBreakdown, *dayMetrics)
	}

	metrics.DailyBreakdown = dailyBreakdown

	for _, day := range dailyBreakdown {
		metrics.TotalRuns += day.TotalRuns
		metrics.SuccessfulRuns += day.SuccessfulRuns
		metrics.FailedRuns += day.FailedRuns
		metrics.FunctionsGenerated += day.FunctionsGenerated
		metrics.FunctionsPublished += day.FunctionsPublished
		metrics.NewOpportunities += day.NewOpportunities
	}

	if metrics.SuccessfulRuns > 0 {
		metrics.AverageQualityScore = metrics.TotalQualityScore / float64(metrics.SuccessfulRuns)
	}

	return metrics, nil
}

func (mc *MetricsCollector) GetProductivityScore(ctx context.Context) (float64, error) {
	daily, err := mc.CollectDailyMetrics(ctx)
	if err != nil {
		return 0, err
	}

	qualityWeight := 0.4
	volumeWeight := 0.3
	conversionWeight := 0.3

	qualityScore := daily.AverageQualityScore
	volumeScore := float64(daily.FunctionsPublished) / 10.0
	if volumeScore > 100 {
		volumeScore = 100
	}
	conversionScore := daily.ConversionRate

	return (qualityScore * qualityWeight) + (volumeScore * volumeWeight) + (conversionScore * conversionWeight), nil
}

func (mc *MetricsCollector) RecordMetricsSnapshot(ctx context.Context) error {
	daily, err := mc.CollectDailyMetrics(ctx)
	if err != nil {
		return err
	}

	snapshot := &SwarmMetricsSnapshot{
		ID:                    uuid.New(),
		SnapshotAt:            time.Now(),
		Date:                  daily.Date,
		TotalRuns:             daily.TotalRuns,
		SuccessfulRuns:        daily.SuccessfulRuns,
		FailedRuns:            daily.FailedRuns,
		FunctionsGenerated:    daily.FunctionsGenerated,
		FunctionsPublished:    daily.FunctionsPublished,
		AverageQualityScore:   daily.AverageQualityScore,
		NewOpportunities:      daily.NewOpportunities,
		ConversionRate:        daily.ConversionRate,
	}

	return mc.db.WithContext(ctx).Create(snapshot).Error
}

func (mc *MetricsCollector) GetTopPerformingFunctions(ctx context.Context, limit int) ([]TopFunction, error) {
	type result struct {
		FunctionID       string
		TotalExecutions  int64
		AverageLatencyMs float64
		SuccessRate      float64
	}

	var results []result
	if err := mc.db.WithContext(ctx).
		Model(&factorysvc.FactoryVersion{}).
		Select("function_id, count(*) as total_executions, avg(test_score) as average_latency_ms, sum(case when test_score >= 80 then 1 else 0 end)::float / count(*) * 100 as success_rate").
		Group("function_id").
		Order("total_executions DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	topFunctions := make([]TopFunction, 0, len(results))
	for _, r := range results {
		topFunctions = append(topFunctions, TopFunction{
			FunctionID:       r.FunctionID,
			TotalExecutions:  r.TotalExecutions,
			AverageTestScore: r.AverageLatencyMs,
			SuccessRate:      r.SuccessRate,
		})
	}

	return topFunctions, nil
}

func calculateConversionRate(discovered, published int) float64 {
	if discovered == 0 {
		return 0
	}
	return float64(published) / float64(discovered) * 100
}

type DailySwarmMetrics struct {
	Date                    time.Time `json:"date"`
	TotalRuns               int       `json:"total_runs"`
	SuccessfulRuns          int       `json:"successful_runs"`
	FailedRuns              int       `json:"failed_runs"`
	FunctionsGenerated      int       `json:"functions_generated"`
	FunctionsPublished      int       `json:"functions_published"`
	AverageQualityScore     float64   `json:"average_quality_score"`
	NewOpportunities        int64     `json:"new_opportunities"`
	OpportunitiesDiscovered  int       `json:"opportunities_discovered"`
	OpportunitiesGenerated   int       `json:"opportunities_generated"`
	OpportunitiesPublished   int       `json:"opportunities_published"`
	OpportunitiesRejected    int       `json:"opportunities_rejected"`
	ConversionRate          float64   `json:"conversion_rate"`
	TotalQualityScore       float64   `json:"-"`
}

type WeeklySwarmMetrics struct {
	StartDate          time.Time            `json:"start_date"`
	EndDate            time.Time            `json:"end_date"`
	TotalRuns          int                  `json:"total_runs"`
	SuccessfulRuns     int                  `json:"successful_runs"`
	FailedRuns         int                  `json:"failed_runs"`
	FunctionsGenerated int                  `json:"functions_generated"`
	FunctionsPublished int                  `json:"functions_published"`
	AverageQualityScore float64             `json:"average_quality_score"`
	NewOpportunities   int64                `json:"new_opportunities"`
	DailyBreakdown     []DailySwarmMetrics  `json:"daily_breakdown"`
	TotalQualityScore  float64              `json:"-"`
}

type TopFunction struct {
	FunctionID       string  `json:"function_id"`
	TotalExecutions  int64   `json:"total_executions"`
	AverageTestScore float64 `json:"average_test_score"`
	SuccessRate      float64 `json:"success_rate"`
}

type SwarmMetricsSnapshot struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SnapshotAt          time.Time `json:"snapshot_at" gorm:"autoCreateTime"`
	Date                time.Time `json:"date" gorm:"type:date;index"`
	TotalRuns           int       `json:"total_runs"`
	SuccessfulRuns      int       `json:"successful_runs"`
	FailedRuns          int       `json:"failed_runs"`
	FunctionsGenerated   int       `json:"functions_generated"`
	FunctionsPublished   int       `json:"functions_published"`
	AverageQualityScore float64   `json:"average_quality_score" gorm:"type:decimal(5,2)"`
	NewOpportunities    int64     `json:"new_opportunities"`
	ConversionRate      float64   `json:"conversion_rate" gorm:"type:decimal(5,2)"`
}

func (SwarmMetricsSnapshot) TableName() string { return "swarm_metrics_snapshots" }