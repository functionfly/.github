package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// FunctionRepository handles function and deployment operations
type FunctionRepository interface {
	CreateFunction(ctx context.Context, function *storage.FunctionConfig) (*storage.FunctionConfig, error)
	GetFunctionByID(ctx context.Context, functionID uuid.UUID) (*storage.FunctionConfig, error)
	ListFunctionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.FunctionConfig, error)
	ListAllFunctions(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status *string) ([]*storage.FunctionConfig, int, error)
	UpdateFunction(ctx context.Context, functionID uuid.UUID, updates map[string]interface{}) (*storage.FunctionConfig, error)
	DeleteFunction(ctx context.Context, functionID uuid.UUID) error
	GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*storage.FunctionConfig, error)
	GetActiveDeploymentForFunction(ctx context.Context, functionID uuid.UUID) (*storage.FunctionDeployment, error)

	// Function deployment operations
	CreateFunctionDeployment(ctx context.Context, deployment *storage.FunctionDeployment) (*storage.FunctionDeployment, error)
	GetFunctionDeploymentByID(ctx context.Context, deploymentID uuid.UUID) (*storage.FunctionDeployment, error)
	ListFunctionDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*storage.FunctionDeployment, error)
	UpdateFunctionDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL, errorMessage *string) error

	// Function log operations
	CreateFunctionLog(ctx context.Context, log *storage.FunctionLog) error
	GetFunctionLogs(ctx context.Context, functionID *uuid.UUID, deploymentID *uuid.UUID, limit int, since *time.Time, level *string) ([]*storage.FunctionLog, error)
	DeleteFunctionLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// AppRepository handles application operations
type AppRepository interface {
	CreateApp(name, slug string, tenantID uuid.UUID) (*storage.App, error)
	GetAppByID(id uuid.UUID) (*storage.App, error)
	GetAppBySlug(slug string) (*storage.App, error)
	GetAppBySlugAndTenant(slug string, tenantID uuid.UUID) (*storage.App, error)
	ListAppsByTenant(tenantID uuid.UUID) ([]*storage.App, error)
}

// BackendRepository handles backend routing and operations
type BackendRepository interface {
	CreateBackend(appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*storage.Backend, error)
	ListBackendsByAppID(appID uuid.UUID) ([]*storage.Backend, error)
	GetBackendByID(id uuid.UUID) (*storage.Backend, error)
	GetAllEnabledBackends() ([]*storage.Backend, error)
	ListAllBackends(ctx context.Context) ([]*storage.Backend, error)
	UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error
	DeleteBackend(ctx context.Context, backendID uuid.UUID) error
}

// DeploymentRepository handles deployment lifecycle
type DeploymentRepository interface {
	CreateDeployment(appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*storage.Deployment, error)
	UpdateDeploymentStatus(id uuid.UUID, status, message string, metadata map[string]interface{}) error
	GetDeploymentByID(id uuid.UUID) (*storage.Deployment, error)
	ListDeploymentsByAppID(appID uuid.UUID, limit int) ([]*storage.Deployment, error)
	GetLatestSuccessfulDeployment(appID uuid.UUID, provider string) (*storage.Deployment, error)
}

// RoutingRepository handles routing and load balancing
type RoutingRepository interface {
	InsertRoutingEvent(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error
	GetRecentRoutingEvents(limit int, since time.Time) ([]*storage.RoutingEvent, error)
	CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error)
}

// CircuitRepository handles circuit breaker state
type CircuitRepository interface {
	GetCircuitState(backendID uuid.UUID) (*storage.CircuitState, error)
	UpdateCircuitState(state *storage.CircuitState) error
	UpsertCircuitState(state *storage.CircuitState) error
}

// HealthRepository handles health checks
type HealthRepository interface {
	InsertHealthCheck(backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error
	GetRecentHealthChecks(backendID uuid.UUID, limit int) ([]*storage.HealthCheck, error)
	GetBackendStatusByAppID(appID uuid.UUID) ([]*storage.BackendStatus, error)
}

// ArtifactRepository handles deployment artifacts
type ArtifactRepository interface {
	StoreDeploymentArtifact(appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*storage.DeploymentArtifact, error)
	GetDeploymentArtifact(key string) (*storage.DeploymentArtifact, error)
}

// ProviderRepository handles cloud provider integrations
type ProviderRepository interface {
	CreateProvider(provider *storage.Provider) error
	GetProviderByID(providerID string) (*storage.Provider, error)
	GetProviderByUserAndType(userID uuid.UUID, providerType string) (*storage.Provider, error)
	GetProvidersByUser(userID uuid.UUID) ([]*storage.Provider, error)
	ListAllProviders(ctx context.Context) ([]*storage.Provider, error)
	UpdateProviderStatus(providerID string, status string) error
	UpdateProvider(ctx context.Context, providerID string, updates map[string]interface{}) (*storage.Provider, error)
	UpdateProviderLastUsed(ctx context.Context, providerID string) error
	GetStaleProviders(ctx context.Context, since time.Time) ([]*storage.Provider, error)
	ShareProviderWithTeam(providerID string, teamID string) error
	DeleteProvider(ctx context.Context, providerID string, userID uuid.UUID) error
}

// LocalRuntimeRepository handles local runtime registrations
type LocalRuntimeRepository interface {
	RegisterLocalRuntime(ctx context.Context, instance *storage.LocalRuntimeInstance) (*storage.LocalRuntimeInstance, error)
	UpdateLocalRuntimeHeartbeat(ctx context.Context, runtimeID string) error
	GetLocalRuntimeByID(ctx context.Context, instanceID uuid.UUID) (*storage.LocalRuntimeInstance, error)
	GetLocalRuntimeByRuntimeID(ctx context.Context, runtimeID string) (*storage.LocalRuntimeInstance, error)
	ListActiveLocalRuntimes(ctx context.Context) ([]*storage.LocalRuntimeInstance, error)
	DeregisterLocalRuntime(ctx context.Context, runtimeID string) error
	CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error)

	// Metrics
	RecordLocalRuntimeMetrics(ctx context.Context, metrics *storage.LocalRuntimeMetric) error
	GetLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time, limit int) ([]*storage.LocalRuntimeMetric, error)
	GetLatestLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID) (*storage.LocalRuntimeMetric, error)
	GetAggregatedLocalRuntimeMetrics(ctx context.Context, since time.Time) (map[string]interface{}, error)

	// Health
	RecordLocalRuntimeHealth(ctx context.Context, health *storage.LocalRuntimeHealth) error
	GetLocalRuntimeHealth(ctx context.Context, instanceID uuid.UUID) (*storage.LocalRuntimeHealth, error)
}
