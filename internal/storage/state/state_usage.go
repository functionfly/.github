package state

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Usage Metrics Operations
// ============================================

// RecordUsage records a usage metric
func (r *StateRepository) RecordUsage(ctx context.Context, metric *StateUsageMetric) error {
	if metric.ID == uuid.Nil {
		metric.ID = uuid.New()
	}
	metric.CreatedAt = time.Now()

	return r.db.WithContext(ctx).Create(metric).Error
}

// GetUsage retrieves usage metrics for a state
func (r *StateRepository) GetUsage(ctx context.Context, stateID uuid.UUID, metricType string, start, end time.Time) ([]*StateUsageMetric, error) {
	query := r.db.WithContext(ctx).Where("state_id = ? AND period_start >= ? AND period_end <= ?", stateID, start, end)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}

	var metrics []*StateUsageMetric
	err := query.Order("period_start DESC").Find(&metrics).Error
	return metrics, err
}