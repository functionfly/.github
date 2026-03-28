package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: backends, routing, health, deployments, feature measures.

// Backend operations
func (db *PostgresDB) CreateBackend(appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error) {
	return db.backendRepository.CreateBackend(appID, provider, region, url, sharedSecret, priority)
}

func (db *PostgresDB) ListBackendsByAppID(appID uuid.UUID) ([]*Backend, error) {
	return db.backendRepository.ListBackendsByAppID(appID)
}

func (db *PostgresDB) GetBackendByID(id uuid.UUID) (*Backend, error) {
	return db.backendRepository.GetBackendByID(id)
}

func (db *PostgresDB) GetAllEnabledBackends() ([]*Backend, error) {
	return db.backendRepository.GetAllEnabledBackends()
}

func (db *PostgresDB) ListAllBackends(ctx context.Context) ([]*Backend, error) {
	return db.backendRepository.ListAllBackends(ctx)
}

func (db *PostgresDB) UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error {
	return db.backendRepository.UpdateBackendEnabled(ctx, backendID, enabled)
}

func (db *PostgresDB) ListFeatureMeasures(ctx context.Context) ([]*FeatureMeasure, error) {
	return db.featureMeasureRepository.ListFeatureMeasures(ctx)
}

func (db *PostgresDB) UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return db.featureMeasureRepository.UpdateFeatureMeasureEnabled(ctx, id, enabled)
}

func (db *PostgresDB) InsertHealthCheck(backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error {
	return db.backendRepository.InsertHealthCheck(backendID, ok, statusCode, latencyMs, errorMessage)
}

func (db *PostgresDB) GetRecentHealthChecks(backendID uuid.UUID, limit int) ([]*HealthCheck, error) {
	return db.backendRepository.GetRecentHealthChecks(backendID, limit)
}

func (db *PostgresDB) GetCircuitState(backendID uuid.UUID) (*CircuitState, error) {
	return db.backendRepository.GetCircuitState(backendID)
}

func (db *PostgresDB) UpdateCircuitState(state *CircuitState) error {
	return db.backendRepository.UpdateCircuitState(state)
}

func (db *PostgresDB) UpsertCircuitState(state *CircuitState) error {
	return db.backendRepository.UpsertCircuitState(state)
}

func (db *PostgresDB) InsertRoutingEvent(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	return db.backendRepository.InsertRoutingEvent(appID, backendID, latencyMs, outcome, requestID)
}

func (db *PostgresDB) GetRecentRoutingEvents(limit int, since time.Time) ([]*RoutingEvent, error) {
	return db.backendRepository.GetRecentRoutingEvents(limit, since)
}

func (db *PostgresDB) GetBackendStatusByAppID(appID uuid.UUID) ([]*BackendStatus, error) {
	return db.backendRepository.GetBackendStatusByAppID(appID)
}

// Deployment operations
func (db *PostgresDB) CreateDeployment(appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*Deployment, error) {
	return db.deploymentRepository.CreateDeployment(appID, provider, region, deploymentID, artifactKey, routes)
}

func (db *PostgresDB) UpdateDeploymentStatus(id uuid.UUID, status, message string, metadata map[string]interface{}) error {
	return db.deploymentRepository.UpdateDeploymentStatus(id, status, message, metadata)
}

func (db *PostgresDB) GetDeploymentByID(id uuid.UUID) (*Deployment, error) {
	return db.deploymentRepository.GetDeploymentByID(id)
}

func (db *PostgresDB) ListDeploymentsByAppID(appID uuid.UUID, limit int) ([]*Deployment, error) {
	return db.deploymentRepository.ListDeploymentsByAppID(appID, limit)
}

func (db *PostgresDB) GetLatestSuccessfulDeployment(appID uuid.UUID, provider string) (*Deployment, error) {
	return db.deploymentRepository.GetLatestSuccessfulDeployment(appID, provider)
}

func (db *PostgresDB) StoreDeploymentArtifact(appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*DeploymentArtifact, error) {
	return db.deploymentRepository.StoreDeploymentArtifact(appID, provider, key, contentType, checksum, size)
}

func (db *PostgresDB) GetDeploymentArtifact(key string) (*DeploymentArtifact, error) {
	return db.deploymentRepository.GetDeploymentArtifact(key)
}
