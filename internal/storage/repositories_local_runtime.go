package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: local runtime registry, metrics, health.

// Local runtime registry operations
func (db *PostgresDB) RegisterLocalRuntime(ctx context.Context, instance *LocalRuntimeInstance) (*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.RegisterLocalRuntime(ctx, instance)
}

func (db *PostgresDB) UpdateLocalRuntimeHeartbeat(ctx context.Context, runtimeID string) error {
	return db.localRuntimeRepository.UpdateLocalRuntimeHeartbeat(ctx, runtimeID)
}

func (db *PostgresDB) GetLocalRuntimeByID(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.GetLocalRuntimeByID(ctx, instanceID)
}

func (db *PostgresDB) GetLocalRuntimeByRuntimeID(ctx context.Context, runtimeID string) (*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.GetLocalRuntimeByRuntimeID(ctx, runtimeID)
}

func (db *PostgresDB) ListActiveLocalRuntimes(ctx context.Context) ([]*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.ListActiveLocalRuntimes(ctx)
}

func (db *PostgresDB) DeregisterLocalRuntime(ctx context.Context, runtimeID string) error {
	return db.localRuntimeRepository.DeregisterLocalRuntime(ctx, runtimeID)
}

func (db *PostgresDB) CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error) {
	return db.localRuntimeRepository.CleanupStaleLocalRuntimes(ctx, maxAge)
}

// Local runtime metrics operations
func (db *PostgresDB) RecordLocalRuntimeMetrics(ctx context.Context, metrics *LocalRuntimeMetric) error {
	return db.localRuntimeRepository.RecordLocalRuntimeMetrics(ctx, metrics)
}

func (db *PostgresDB) GetLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time, limit int) ([]*LocalRuntimeMetric, error) {
	return db.localRuntimeRepository.GetLocalRuntimeMetrics(ctx, instanceID, since, limit)
}

func (db *PostgresDB) GetLatestLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeMetric, error) {
	return db.localRuntimeRepository.GetLatestLocalRuntimeMetrics(ctx, instanceID)
}

func (db *PostgresDB) GetAggregatedLocalRuntimeMetrics(ctx context.Context, since time.Time) (map[string]interface{}, error) {
	return db.localRuntimeRepository.GetAggregatedLocalRuntimeMetrics(ctx, since)
}

// Local runtime health operations
func (db *PostgresDB) RecordLocalRuntimeHealth(ctx context.Context, health *LocalRuntimeHealth) error {
	return db.localRuntimeRepository.RecordLocalRuntimeHealth(ctx, health)
}

func (db *PostgresDB) GetLocalRuntimeHealth(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeHealth, error) {
	return db.localRuntimeRepository.GetLocalRuntimeHealth(ctx, instanceID)
}
