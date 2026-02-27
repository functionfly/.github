package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocalRuntimeRepository handles local runtime registry operations
type LocalRuntimeRepository struct {
	db *gorm.DB
}

// NewLocalRuntimeRepository creates a new local runtime repository
func NewLocalRuntimeRepository(db *PostgresDB) *LocalRuntimeRepository {
	return &LocalRuntimeRepository{db: db.GORM}
}

// RegisterLocalRuntime registers a new local runtime instance
func (r *LocalRuntimeRepository) RegisterLocalRuntime(ctx context.Context, instance *LocalRuntimeInstance) (*LocalRuntimeInstance, error) {
	if instance.ID == uuid.Nil {
		instance.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(instance).Error; err != nil {
		return nil, fmt.Errorf("failed to register local runtime: %w", err)
	}

	return instance, nil
}

// UpdateLocalRuntimeHeartbeat updates the heartbeat timestamp for a runtime instance
func (r *LocalRuntimeRepository) UpdateLocalRuntimeHeartbeat(ctx context.Context, runtimeID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&LocalRuntimeInstance{}).
		Where("runtime_id = ?", runtimeID).
		Updates(map[string]interface{}{
			"last_heartbeat": now,
			"updated_at":     now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update heartbeat: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("runtime instance not found: %s", runtimeID)
	}

	return nil
}

// GetLocalRuntimeByID retrieves a runtime instance by its UUID
func (r *LocalRuntimeRepository) GetLocalRuntimeByID(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeInstance, error) {
	var instance LocalRuntimeInstance
	err := r.db.WithContext(ctx).First(&instance, instanceID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("runtime instance not found: %s", instanceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime instance: %w", err)
	}

	return &instance, nil
}

// GetLocalRuntimeByRuntimeID retrieves a runtime instance by its runtime ID
func (r *LocalRuntimeRepository) GetLocalRuntimeByRuntimeID(ctx context.Context, runtimeID string) (*LocalRuntimeInstance, error) {
	var instance LocalRuntimeInstance
	err := r.db.WithContext(ctx).Where("runtime_id = ?", runtimeID).First(&instance).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("runtime instance not found: %s", runtimeID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime instance: %w", err)
	}

	return &instance, nil
}

// ListActiveLocalRuntimes returns all active runtime instances (updated within last 5 minutes)
func (r *LocalRuntimeRepository) ListActiveLocalRuntimes(ctx context.Context) ([]*LocalRuntimeInstance, error) {
	var instances []*LocalRuntimeInstance
	cutoff := time.Now().Add(-5 * time.Minute)

	err := r.db.WithContext(ctx).
		Where("last_heartbeat > ?", cutoff).
		Order("last_heartbeat DESC").
		Find(&instances).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list active runtimes: %w", err)
	}

	return instances, nil
}

// DeregisterLocalRuntime removes a runtime instance from the registry
func (r *LocalRuntimeRepository) DeregisterLocalRuntime(ctx context.Context, runtimeID string) error {
	result := r.db.WithContext(ctx).Where("runtime_id = ?", runtimeID).Delete(&LocalRuntimeInstance{})
	if result.Error != nil {
		return fmt.Errorf("failed to deregister runtime: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("runtime instance not found: %s", runtimeID)
	}

	return nil
}

// CleanupStaleLocalRuntimes removes runtime instances that haven't sent heartbeats recently
func (r *LocalRuntimeRepository) CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)
	result := r.db.WithContext(ctx).Where("last_heartbeat < ?", cutoff).Delete(&LocalRuntimeInstance{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup stale runtimes: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// RecordLocalRuntimeMetrics records metrics for a runtime instance
func (r *LocalRuntimeRepository) RecordLocalRuntimeMetrics(ctx context.Context, metrics *LocalRuntimeMetric) error {
	if metrics.ID == uuid.Nil {
		metrics.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(metrics).Error; err != nil {
		return fmt.Errorf("failed to record runtime metrics: %w", err)
	}

	return nil
}

// GetLocalRuntimeMetrics retrieves metrics for a runtime instance
func (r *LocalRuntimeRepository) GetLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time, limit int) ([]*LocalRuntimeMetric, error) {
	var metrics []*LocalRuntimeMetric
	err := r.db.WithContext(ctx).
		Where("runtime_instance_id = ? AND timestamp >= ?", instanceID, since).
		Order("timestamp DESC").
		Limit(limit).
		Find(&metrics).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get runtime metrics: %w", err)
	}

	return metrics, nil
}

// GetLatestLocalRuntimeMetrics gets the most recent metrics for a runtime instance
func (r *LocalRuntimeRepository) GetLatestLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeMetric, error) {
	var metric LocalRuntimeMetric
	err := r.db.WithContext(ctx).
		Where("runtime_instance_id = ?", instanceID).
		Order("timestamp DESC").
		First(&metric).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("no metrics found for runtime instance: %s", instanceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}

	return &metric, nil
}

// GetAggregatedLocalRuntimeMetrics returns aggregated metrics across all active runtimes
func (r *LocalRuntimeRepository) GetAggregatedLocalRuntimeMetrics(ctx context.Context, since time.Time) (map[string]interface{}, error) {
	// Only include metrics from runtimes that have sent heartbeats in the last 5 minutes
	heartbeatCutoff := time.Now().Add(-5 * time.Minute)

	var result struct {
		TotalInstances      int     `json:"total_instances"`
		AvgCPUUsage         *float64 `json:"avg_cpu_usage"`
		TotalConnections    *int64   `json:"total_connections"`
		AvgThroughput       *float64 `json:"avg_throughput"`
		TotalRequests       *int64   `json:"total_requests"`
		AvgErrorRate        *float64 `json:"avg_error_rate"`
		TotalExecutions     *int64   `json:"total_executions"`
		AvgLatencyNS        *float64 `json:"avg_latency_ns"`
		TotalErrors         *int64   `json:"total_errors"`
	}

	query := `
		SELECT
			COUNT(DISTINCT runtime_instance_id) as total_instances,
			AVG(cpu_usage) as avg_cpu_usage,
			SUM(active_connections) as total_connections,
			AVG(request_throughput) as avg_throughput,
			SUM(total_requests) as total_requests,
			AVG(error_rate) as avg_error_rate,
			SUM(execution_count) as total_executions,
			AVG(EXTRACT(epoch FROM average_latency)) * 1000000000 as avg_latency_ns,
			SUM(error_count) as total_errors
		FROM local_runtime_metrics m
		JOIN local_runtime_instances i ON m.runtime_instance_id = i.id
		WHERE m.timestamp >= ? AND i.last_heartbeat > ?`

	err := r.db.WithContext(ctx).Raw(query, since, heartbeatCutoff).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated metrics: %w", err)
	}

	response := map[string]interface{}{
		"total_instances":     result.TotalInstances,
		"timestamp":           time.Now(),
		"time_range_start":    since,
	}

	if result.AvgCPUUsage != nil {
		response["average_cpu_usage_percent"] = *result.AvgCPUUsage
	}
	if result.TotalConnections != nil {
		response["total_active_connections"] = *result.TotalConnections
	}
	if result.AvgThroughput != nil {
		response["average_request_throughput"] = *result.AvgThroughput
	}
	if result.TotalRequests != nil {
		response["total_requests"] = *result.TotalRequests
	}
	if result.AvgErrorRate != nil {
		response["average_error_rate_percent"] = *result.AvgErrorRate
	}
	if result.TotalExecutions != nil {
		response["total_function_executions"] = *result.TotalExecutions
	}
	if result.AvgLatencyNS != nil {
		response["average_execution_latency_ms"] = *result.AvgLatencyNS / 1000000 // Convert ns to ms
	}
	if result.TotalErrors != nil {
		response["total_execution_errors"] = *result.TotalErrors
	}

	return response, nil
}

// RecordLocalRuntimeHealth records health status for a runtime instance
func (r *LocalRuntimeRepository) RecordLocalRuntimeHealth(ctx context.Context, health *LocalRuntimeHealth) error {
	if err := r.db.WithContext(ctx).Save(health).Error; err != nil {
		return fmt.Errorf("failed to record runtime health: %w", err)
	}

	return nil
}

// GetLocalRuntimeHealth gets the latest health status for a runtime instance
func (r *LocalRuntimeRepository) GetLocalRuntimeHealth(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeHealth, error) {
	var health LocalRuntimeHealth
	err := r.db.WithContext(ctx).
		Where("runtime_instance_id = ?", instanceID).
		Order("timestamp DESC").
		First(&health).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("no health data found for runtime instance: %s", instanceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime health: %w", err)
	}

	return &health, nil
}