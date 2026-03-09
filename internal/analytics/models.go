// Package analytics provides models and services for tracking factory performance metrics
package analytics

import (
	"time"

	"github.com/google/uuid"
)

// MetricType represents the type of metric being recorded
type MetricType string

const (
	MetricTypeGenerationSuccess  MetricType = "generation_success"
	MetricTypeGenerationFailure  MetricType = "generation_failure"
	MetricTypeQualityScore       MetricType = "quality_score"
	MetricTypeTestScore          MetricType = "test_score"
	MetricTypeLatencyGeneration  MetricType = "latency_generation"
	MetricTypeLatencyTesting     MetricType = "latency_testing"
	MetricTypeLatencyPublishing  MetricType = "latency_publishing"
	MetricTypeLatencyTotal       MetricType = "latency_total"
	MetricTypeErrorRate          MetricType = "error_rate"
	MetricTypeThroughput         MetricType = "throughput"
	MetricTypeOpportunityScanned MetricType = "opportunity_scanned"
	MetricTypeFunctionPublished  MetricType = "function_published"
	MetricTypeReviewRequired     MetricType = "review_required"
)

// AggregationPeriod represents the time period for aggregated statistics
type AggregationPeriod string

const (
	AggregationPeriodHourly  AggregationPeriod = "hourly"
	AggregationPeriodDaily   AggregationPeriod = "daily"
	AggregationPeriodWeekly  AggregationPeriod = "weekly"
	AggregationPeriodMonthly AggregationPeriod = "monthly"
)

// FactoryMetric represents a single metric data point
type FactoryMetric struct {
	ID          uuid.UUID      `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RunID       *uuid.UUID     `json:"run_id,omitempty" db:"run_id" gorm:"type:uuid;index"`
	AgentID     string         `json:"agent_id" db:"agent_id" gorm:"not null;index"`
	MetricType  MetricType     `json:"metric_type" db:"metric_type" gorm:"not null;index"`
	MetricValue float64        `json:"metric_value" db:"metric_value" gorm:"not null"`
	Labels      map[string]any `json:"labels" db:"labels" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at" gorm:"not null;default:now();index"`
}

// TableName specifies the table name for FactoryMetric
func (FactoryMetric) TableName() string {
	return "factory_analytics_metrics"
}

// AggregatedMetric represents pre-computed aggregated statistics
type AggregatedMetric struct {
	ID           uuid.UUID         `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AgentID      string            `json:"agent_id" db:"agent_id" gorm:"not null;index"`
	Period       AggregationPeriod `json:"period" db:"period" gorm:"not null;index"`
	PeriodStart  time.Time         `json:"period_start" db:"period_start" gorm:"not null;index"`
	MetricType   MetricType        `json:"metric_type" db:"metric_type" gorm:"not null;index"`
	Count        int64             `json:"count" db:"count" gorm:"not null;default:0"`
	Sum          float64           `json:"sum" db:"sum" gorm:"type:decimal(20,6);default:0"`
	Avg          float64           `json:"avg" db:"avg" gorm:"type:decimal(20,6);default:0"`
	Min          float64           `json:"min" db:"min" gorm:"type:decimal(20,6);default:0"`
	Max          float64           `json:"max" db:"max" gorm:"type:decimal(20,6);default:0"`
	P50          float64           `json:"p50" db:"p50" gorm:"type:decimal(20,6);default:0"`
	P95          float64           `json:"p95" db:"p95" gorm:"type:decimal(20,6);default:0"`
	P99          float64           `json:"p99" db:"p99" gorm:"type:decimal(20,6);default:0"`
	SuccessCount int64             `json:"success_count" db:"success_count" gorm:"default:0"`
	FailureCount int64             `json:"failure_count" db:"failure_count" gorm:"default:0"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for AggregatedMetric
func (AggregatedMetric) TableName() string {
	return "factory_analytics_aggregated"
}

// DashboardStats represents real-time dashboard statistics
type DashboardStats struct {
	// Generation metrics
	TotalRuns      int64   `json:"total_runs"`
	SuccessfulRuns int64   `json:"successful_runs"`
	FailedRuns     int64   `json:"failed_runs"`
	SuccessRate    float64 `json:"success_rate"`

	// Quality metrics
	AvgQualityScore float64 `json:"avg_quality_score"`
	AvgTestScore    float64 `json:"avg_test_score"`
	QualityTrend    float64 `json:"quality_trend"` // Percentage change from previous period

	// Throughput metrics
	FunctionsGenerated int64   `json:"functions_generated"`
	FunctionsPublished int64   `json:"functions_published"`
	ThroughputPerHour  float64 `json:"throughput_per_hour"`

	// Latency metrics (in milliseconds)
	AvgGenerationLatency float64 `json:"avg_generation_latency"`
	AvgTestingLatency    float64 `json:"avg_testing_latency"`
	AvgPublishingLatency float64 `json:"avg_publishing_latency"`
	AvgTotalLatency      float64 `json:"avg_total_latency"`
	P95Latency           float64 `json:"p95_latency"`

	// Error metrics
	ErrorRate     float64    `json:"error_rate"`
	LastErrorTime *time.Time `json:"last_error_time,omitempty"`

	// Review metrics
	PendingReviews int64   `json:"pending_reviews"`
	ReviewRate     float64 `json:"review_rate"` // Percentage of functions requiring review

	// Period info
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	LastUpdated time.Time `json:"last_updated"`
}

// TimeSeriesPoint represents a single point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Count     int64     `json:"count,omitempty"`
}

// TimeSeriesData represents time series data for charts
type TimeSeriesData struct {
	MetricType MetricType        `json:"metric_type"`
	Points     []TimeSeriesPoint `json:"points"`
	Period     AggregationPeriod `json:"period"`
}

// MetricFilter represents filters for querying metrics
type MetricFilter struct {
	AgentID    string            `json:"agent_id,omitempty"`
	MetricType MetricType        `json:"metric_type,omitempty"`
	StartTime  *time.Time        `json:"start_time,omitempty"`
	EndTime    *time.Time        `json:"end_time,omitempty"`
	Period     AggregationPeriod `json:"period,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Offset     int               `json:"offset,omitempty"`
}

// MetricRecord represents a metric to be recorded
type MetricRecord struct {
	RunID       *uuid.UUID     `json:"run_id,omitempty"`
	AgentID     string         `json:"agent_id"`
	MetricType  MetricType     `json:"metric_type"`
	MetricValue float64        `json:"metric_value"`
	Labels      map[string]any `json:"labels,omitempty"`
	Timestamp   *time.Time     `json:"timestamp,omitempty"`
}

// RunMetricsSummary represents metrics summary for a factory run
type RunMetricsSummary struct {
	RunID                uuid.UUID  `json:"run_id"`
	AgentID              string     `json:"agent_id"`
	Status               string     `json:"status"`
	Duration             float64    `json:"duration_ms"`
	OpportunitiesScanned int        `json:"opportunities_scanned"`
	FunctionsGenerated   int        `json:"functions_generated"`
	FunctionsPublished   int        `json:"functions_published"`
	AvgQualityScore      float64    `json:"avg_quality_score"`
	AvgTestScore         float64    `json:"avg_test_score"`
	GenerationLatency    float64    `json:"generation_latency_ms"`
	TestingLatency       float64    `json:"testing_latency_ms"`
	PublishingLatency    float64    `json:"publishing_latency_ms"`
	TotalLatency         float64    `json:"total_latency_ms"`
	ErrorCount           int        `json:"error_count"`
	ReviewRequired       int        `json:"review_required"`
	CreatedAt            time.Time  `json:"created_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
}
