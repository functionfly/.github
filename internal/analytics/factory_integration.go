// Package analytics provides integration with the factory pipeline for metrics recording
package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FactoryMetricsRecorder records metrics for factory pipeline operations
type FactoryMetricsRecorder struct {
	svc     *Service
	agentID string
}

// NewFactoryMetricsRecorder creates a new metrics recorder for the factory
func NewFactoryMetricsRecorder(svc *Service, agentID string) *FactoryMetricsRecorder {
	return &FactoryMetricsRecorder{
		svc:     svc,
		agentID: agentID,
	}
}

// RecordRunStart records the start of a factory run
func (r *FactoryMetricsRecorder) RecordRunStart(ctx context.Context, runID uuid.UUID) {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeOpportunityScanned,
			MetricValue: 0,
			Labels:      map[string]any{"phase": "start"},
			Timestamp:   &now,
		},
	}

	if err := r.svc.repo.RecordMetrics(ctx, records); err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("failed to record run start metrics")
	}
}

// RecordRunComplete records metrics for a completed factory run
func (r *FactoryMetricsRecorder) RecordRunComplete(ctx context.Context, runID uuid.UUID, metrics RunMetricsComplete) {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeLatencyTotal,
			MetricValue: float64(metrics.DurationMs),
			Labels:      map[string]any{"unit": "ms", "status": metrics.Status},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeQualityScore,
			MetricValue: metrics.AvgQualityScore,
			Labels:      map[string]any{"unit": "score"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeTestScore,
			MetricValue: metrics.AvgTestScore,
			Labels:      map[string]any{"unit": "score"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeOpportunityScanned,
			MetricValue: float64(metrics.OpportunitiesScanned),
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeFunctionPublished,
			MetricValue: float64(metrics.FunctionsPublished),
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeThroughput,
			MetricValue: float64(metrics.FunctionsGenerated),
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		},
	}

	// Record success or failure
	if metrics.Status == "succeeded" {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeGenerationSuccess,
			MetricValue: 1,
			Labels:      map[string]any{"status": "success"},
			Timestamp:   &now,
		})
	} else {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeGenerationFailure,
			MetricValue: 1,
			Labels:      map[string]any{"status": "failed", "error": metrics.ErrorMessage},
			Timestamp:   &now,
		})
	}

	// Record review required if applicable
	if metrics.ReviewRequiredCount > 0 {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeReviewRequired,
			MetricValue: float64(metrics.ReviewRequiredCount),
			Labels:      map[string]any{"unit": "count"},
			Timestamp:   &now,
		})
	}

	if err := r.svc.repo.RecordMetrics(ctx, records); err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("failed to record run complete metrics")
	}
}

// RecordGenerationPhase records metrics for the generation phase
func (r *FactoryMetricsRecorder) RecordGenerationPhase(ctx context.Context, runID uuid.UUID, metrics GenerationMetrics) {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeLatencyGeneration,
			MetricValue: float64(metrics.DurationMs),
			Labels: map[string]any{
				"unit":    "ms",
				"model":   metrics.ModelUsed,
				"success": metrics.Success,
			},
			Timestamp: &now,
		},
	}

	if !metrics.Success {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeGenerationFailure,
			MetricValue: 1,
			Labels:      map[string]any{"phase": "generation", "error": metrics.ErrorMessage},
			Timestamp:   &now,
		})
	}

	if err := r.svc.repo.RecordMetrics(ctx, records); err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("failed to record generation phase metrics")
	}
}

// RecordTestingPhase records metrics for the testing phase
func (r *FactoryMetricsRecorder) RecordTestingPhase(ctx context.Context, runID uuid.UUID, metrics TestingMetrics) {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeLatencyTesting,
			MetricValue: float64(metrics.DurationMs),
			Labels: map[string]any{
				"unit":         "ms",
				"tests_run":    metrics.TestsRun,
				"tests_passed": metrics.TestsPassed,
				"tests_failed": metrics.TestsFailed,
			},
			Timestamp: &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeTestScore,
			MetricValue: metrics.Score,
			Labels:      map[string]any{"unit": "score"},
			Timestamp:   &now,
		},
	}

	if err := r.svc.repo.RecordMetrics(ctx, records); err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("failed to record testing phase metrics")
	}
}

// RecordPublishingPhase records metrics for the publishing phase
func (r *FactoryMetricsRecorder) RecordPublishingPhase(ctx context.Context, runID uuid.UUID, metrics PublishingMetrics) {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeLatencyPublishing,
			MetricValue: float64(metrics.DurationMs),
			Labels: map[string]any{
				"unit":     "ms",
				"success":  metrics.Success,
				"platform": metrics.Platform,
			},
			Timestamp: &now,
		},
	}

	if metrics.Success {
		records = append(records, MetricRecord{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeFunctionPublished,
			MetricValue: 1,
			Labels:      map[string]any{"platform": metrics.Platform},
			Timestamp:   &now,
		})
	}

	if err := r.svc.repo.RecordMetrics(ctx, records); err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("failed to record publishing phase metrics")
	}
}

// RecordError records an error that occurred during factory execution
func (r *FactoryMetricsRecorder) RecordError(ctx context.Context, runID uuid.UUID, phase, errorMessage string) {
	now := time.Now().UTC()
	records := []MetricRecord{
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeGenerationFailure,
			MetricValue: 1,
			Labels:      map[string]any{"phase": phase, "error": errorMessage},
			Timestamp:   &now,
		},
		{
			RunID:       &runID,
			AgentID:     r.agentID,
			MetricType:  MetricTypeErrorRate,
			MetricValue: 1,
			Labels:      map[string]any{"phase": phase},
			Timestamp:   &now,
		},
	}

	if err := r.svc.repo.RecordMetrics(ctx, records); err != nil {
		logrus.WithError(err).WithField("run_id", runID).Warn("failed to record error metrics")
	}
}

// Metric data structures

// RunMetricsComplete contains metrics for a completed factory run
type RunMetricsComplete struct {
	Status               string
	DurationMs           int64
	OpportunitiesScanned int
	FunctionsGenerated   int
	FunctionsPublished   int
	AvgQualityScore      float64
	AvgTestScore         float64
	ReviewRequiredCount  int
	ErrorMessage         string
}

// GenerationMetrics contains metrics for the generation phase
type GenerationMetrics struct {
	DurationMs   int64
	ModelUsed    string
	Success      bool
	ErrorMessage string
}

// TestingMetrics contains metrics for the testing phase
type TestingMetrics struct {
	DurationMs  int64
	TestsRun    int
	TestsPassed int
	TestsFailed int
	Score       float64
}

// PublishingMetrics contains metrics for the publishing phase
type PublishingMetrics struct {
	DurationMs   int64
	Success      bool
	Platform     string
	ErrorMessage string
}
