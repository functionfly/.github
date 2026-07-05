package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: backends, routing, health, deployments, feature measures.

// Backend operations
func (db *PostgresDB) CreateBackend(ctx context.Context, appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error) {
	return db.backendRepository.CreateBackend(ctx, appID, provider, region, url, sharedSecret, priority)
}

func (db *PostgresDB) ListBackendsByAppID(ctx context.Context, appID uuid.UUID) ([]*Backend, error) {
	return db.backendRepository.ListBackendsByAppID(ctx, appID)
}

func (db *PostgresDB) CountBackendsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return db.backendRepository.CountBackendsByTenant(ctx, tenantID)
}

func (db *PostgresDB) GetBackendByID(ctx context.Context, id uuid.UUID) (*Backend, error) {
	return db.backendRepository.GetBackendByID(ctx, id)
}

func (db *PostgresDB) GetAllEnabledBackends(ctx context.Context) ([]*Backend, error) {
	return db.backendRepository.GetAllEnabledBackends(ctx)
}

func (db *PostgresDB) ListAllBackends(ctx context.Context) ([]*Backend, error) {
	return db.backendRepository.ListAllBackends(ctx)
}

func (db *PostgresDB) UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error {
	return db.backendRepository.UpdateBackendEnabled(ctx, backendID, enabled)
}

func (db *PostgresDB) DeleteBackend(ctx context.Context, backendID uuid.UUID) error {
	return db.backendRepository.DeleteBackend(ctx, backendID)
}

func (db *PostgresDB) ListFeatureMeasures(ctx context.Context) ([]*FeatureMeasure, error) {
	return db.featureMeasureRepository.ListFeatureMeasures(ctx)
}

func (db *PostgresDB) UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return db.featureMeasureRepository.UpdateFeatureMeasureEnabled(ctx, id, enabled)
}

func (db *PostgresDB) InsertHealthCheck(ctx context.Context, backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error {
	return db.backendRepository.InsertHealthCheck(ctx, backendID, ok, statusCode, latencyMs, errorMessage)
}

func (db *PostgresDB) GetRecentHealthChecks(ctx context.Context, backendID uuid.UUID, limit int) ([]*HealthCheck, error) {
	return db.backendRepository.GetRecentHealthChecks(ctx, backendID, limit)
}

func (db *PostgresDB) DeleteHealthChecksBefore(ctx context.Context, before time.Time) (int64, error) {
	return db.backendRepository.DeleteHealthChecksBefore(ctx, before)
}

func (db *PostgresDB) GetCircuitState(ctx context.Context, backendID uuid.UUID) (*CircuitState, error) {
	return db.backendRepository.GetCircuitState(ctx, backendID)
}

func (db *PostgresDB) UpdateCircuitState(ctx context.Context, state *CircuitState) error {
	return db.backendRepository.UpdateCircuitState(ctx, state)
}

func (db *PostgresDB) UpsertCircuitState(ctx context.Context, state *CircuitState) error {
	return db.backendRepository.UpsertCircuitState(ctx, state)
}

func (db *PostgresDB) InsertRoutingEvent(ctx context.Context, appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	return db.backendRepository.InsertRoutingEvent(ctx, appID, backendID, latencyMs, outcome, requestID)
}

func (db *PostgresDB) GetRecentRoutingEvents(ctx context.Context, limit int, since time.Time) ([]*RoutingEvent, error) {
	return db.backendRepository.GetRecentRoutingEvents(ctx, limit, since)
}

func (db *PostgresDB) GetRecentRoutingEventsByBackend(ctx context.Context, backendID uuid.UUID, limit int) ([]*RoutingEvent, error) {
	return db.backendRepository.GetRecentRoutingEventsByBackend(ctx, backendID, limit)
}

// App analytics operations

func (db *PostgresDB) GetAppAnalyticsSummary(ctx context.Context, appID uuid.UUID, since time.Time) (*AppAnalyticsSummary, error) {
	return db.appAnalyticsRepository.GetAppAnalyticsSummary(ctx, appID, since)
}

func (db *PostgresDB) GetAppRequestTimeseries(ctx context.Context, appID uuid.UUID, since time.Time, interval string) ([]*AppRequestTimeseriesPoint, error) {
	return db.appAnalyticsRepository.GetAppRequestTimeseries(ctx, appID, since, interval)
}

func (db *PostgresDB) GetAppLatencyTimeseries(ctx context.Context, appID uuid.UUID, since time.Time, interval string) ([]*AppLatencyTimeseriesPoint, error) {
	return db.appAnalyticsRepository.GetAppLatencyTimeseries(ctx, appID, since, interval)
}

func (db *PostgresDB) GetAppTopErrors(ctx context.Context, appID uuid.UUID, since time.Time) ([]*AppErrorBreakdown, error) {
	return db.appAnalyticsRepository.GetAppTopErrors(ctx, appID, since)
}

func (db *PostgresDB) GetAppBackendBreakdown(ctx context.Context, appID uuid.UUID, since time.Time) ([]*AppBackendBreakdown, error) {
	return db.appAnalyticsRepository.GetAppBackendBreakdown(ctx, appID, since)
}

func (db *PostgresDB) GetBackendStatusByAppID(ctx context.Context, appID uuid.UUID) ([]*BackendStatus, error) {
	return db.backendRepository.GetBackendStatusByAppID(ctx, appID)
}

// Deployment operations
func (db *PostgresDB) CreateDeployment(ctx context.Context, appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*Deployment, error) {
	return db.deploymentRepository.CreateDeployment(ctx, appID, provider, region, deploymentID, artifactKey, routes)
}

func (db *PostgresDB) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status, message string, metadata map[string]interface{}) error {
	return db.deploymentRepository.UpdateDeploymentStatus(ctx, id, status, message, metadata)
}

func (db *PostgresDB) GetDeploymentByID(ctx context.Context, id uuid.UUID) (*Deployment, error) {
	return db.deploymentRepository.GetDeploymentByID(ctx, id)
}

func (db *PostgresDB) ListDeploymentsByAppID(ctx context.Context, appID uuid.UUID, limit int) ([]*Deployment, error) {
	return db.deploymentRepository.ListDeploymentsByAppID(ctx, appID, limit)
}

func (db *PostgresDB) GetLatestSuccessfulDeployment(ctx context.Context, appID uuid.UUID, provider string) (*Deployment, error) {
	return db.deploymentRepository.GetLatestSuccessfulDeployment(ctx, appID, provider)
}

func (db *PostgresDB) StoreDeploymentArtifact(ctx context.Context, appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*DeploymentArtifact, error) {
	return db.deploymentRepository.StoreDeploymentArtifact(ctx, appID, provider, key, contentType, checksum, size)
}

func (db *PostgresDB) GetDeploymentArtifact(ctx context.Context, key string) (*DeploymentArtifact, error) {
	return db.deploymentRepository.GetDeploymentArtifact(ctx, key)
}
