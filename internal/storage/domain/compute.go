package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// FunctionRepository handles function and deployment operations
type FunctionRepository interface {
	CreateFunction(ctx context.Context, function *types.FunctionConfig) (*types.FunctionConfig, error)
	GetFunctionByID(ctx context.Context, functionID uuid.UUID) (*types.FunctionConfig, error)
	ListFunctionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.FunctionConfig, error)
	ListAllFunctions(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status *string) ([]*types.FunctionConfig, int, error)
	UpdateFunction(ctx context.Context, functionID uuid.UUID, updates map[string]interface{}) (*types.FunctionConfig, error)
	DeleteFunction(ctx context.Context, functionID uuid.UUID) error
	GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*types.FunctionConfig, error)
	GetActiveDeploymentForFunction(ctx context.Context, functionID uuid.UUID) (*types.FunctionDeployment, error)

	// Function deployment operations
	CreateFunctionDeployment(ctx context.Context, deployment *types.FunctionDeployment) (*types.FunctionDeployment, error)
	GetFunctionDeploymentByID(ctx context.Context, deploymentID uuid.UUID) (*types.FunctionDeployment, error)
	ListFunctionDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*types.FunctionDeployment, error)
	UpdateFunctionDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL, errorMessage *string) error

	// Function log operations
	CreateFunctionLog(ctx context.Context, log *types.FunctionLog) error
	GetFunctionLogs(ctx context.Context, functionID *uuid.UUID, deploymentID *uuid.UUID, limit int, since *time.Time, level *string) ([]*types.FunctionLog, error)
	DeleteFunctionLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// AppRepository handles application operations
type AppRepository interface {
	CreateApp(ctx context.Context, name, slug string, tenantID uuid.UUID) (*types.App, error)
	GetAppByID(ctx context.Context, id uuid.UUID) (*types.App, error)
	GetAppBySlug(ctx context.Context, slug string) (*types.App, error)
	GetAppBySlugAndTenant(ctx context.Context, slug string, tenantID uuid.UUID) (*types.App, error)
	ListAppsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.App, error)
}

// BackendRepository handles backend routing and operations
type BackendRepository interface {
	CreateBackend(ctx context.Context, appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*types.Backend, error)
	ListBackendsByAppID(ctx context.Context, appID uuid.UUID) ([]*types.Backend, error)
	GetBackendByID(ctx context.Context, id uuid.UUID) (*types.Backend, error)
	GetAllEnabledBackends(ctx context.Context) ([]*types.Backend, error)
	ListAllBackends(ctx context.Context) ([]*types.Backend, error)
	UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error
	DeleteBackend(ctx context.Context, backendID uuid.UUID) error
}

// DeploymentRepository handles deployment lifecycle
type DeploymentRepository interface {
	CreateDeployment(ctx context.Context, appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*types.Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status, message string, metadata map[string]interface{}) error
	GetDeploymentByID(ctx context.Context, id uuid.UUID) (*types.Deployment, error)
	ListDeploymentsByAppID(ctx context.Context, appID uuid.UUID, limit int) ([]*types.Deployment, error)
	GetLatestSuccessfulDeployment(ctx context.Context, appID uuid.UUID, provider string) (*types.Deployment, error)
}

// RoutingRepository handles routing and load balancing
type RoutingRepository interface {
	InsertRoutingEvent(ctx context.Context, appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error
	GetRecentRoutingEvents(ctx context.Context, limit int, since time.Time) ([]*types.RoutingEvent, error)
	CountRoutingEventsForTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)
}

// CircuitRepository handles circuit breaker state
type CircuitRepository interface {
	GetCircuitState(ctx context.Context, backendID uuid.UUID) (*types.CircuitState, error)
	UpdateCircuitState(ctx context.Context, state *types.CircuitState) error
	UpsertCircuitState(ctx context.Context, state *types.CircuitState) error
}

// HealthRepository handles health checks
type HealthRepository interface {
	InsertHealthCheck(ctx context.Context, backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error
	GetRecentHealthChecks(ctx context.Context, backendID uuid.UUID, limit int) ([]*types.HealthCheck, error)
	GetBackendStatusByAppID(ctx context.Context, appID uuid.UUID) ([]*types.BackendStatus, error)
}

// ArtifactRepository handles deployment artifacts
type ArtifactRepository interface {
	StoreDeploymentArtifact(ctx context.Context, appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*types.DeploymentArtifact, error)
	GetDeploymentArtifact(ctx context.Context, key string) (*types.DeploymentArtifact, error)
}

// ProviderRepository handles cloud provider integrations
type ProviderRepository interface {
	CreateProvider(ctx context.Context, provider *types.Provider) error
	GetProviderByID(ctx context.Context, providerID string) (*types.Provider, error)
	GetProviderByUserAndType(ctx context.Context, userID uuid.UUID, providerType string) (*types.Provider, error)
	GetProvidersByUser(ctx context.Context, userID uuid.UUID) ([]*types.Provider, error)
	ListAllProviders(ctx context.Context) ([]*types.Provider, error)
	ListProviderSettings(ctx context.Context) ([]*types.ProviderSettings, error)
	GetProviderSettings(ctx context.Context, provider string) (*types.ProviderSettings, error)
	SetProviderDisabled(ctx context.Context, provider string, disabled bool, reason, disabledBy string) error
	UpdateProviderStatus(ctx context.Context, providerID string, status string) error
	UpdateProvider(ctx context.Context, providerID string, updates map[string]interface{}) (*types.Provider, error)
	UpdateProviderLastUsed(ctx context.Context, providerID string) error
	GetStaleProviders(ctx context.Context, since time.Time) ([]*types.Provider, error)
	ShareProviderWithTeam(ctx context.Context, providerID string, teamID string) error
	DeleteProvider(ctx context.Context, providerID string, userID uuid.UUID) error
}

// LocalRuntimeRepository handles local runtime registrations
type LocalRuntimeRepository interface {
	RegisterLocalRuntime(ctx context.Context, instance *types.LocalRuntimeInstance) (*types.LocalRuntimeInstance, error)
	UpdateLocalRuntimeHeartbeat(ctx context.Context, runtimeID string) error
	GetLocalRuntimeByID(ctx context.Context, instanceID uuid.UUID) (*types.LocalRuntimeInstance, error)
	GetLocalRuntimeByRuntimeID(ctx context.Context, runtimeID string) (*types.LocalRuntimeInstance, error)
	ListActiveLocalRuntimes(ctx context.Context) ([]*types.LocalRuntimeInstance, error)
	DeregisterLocalRuntime(ctx context.Context, runtimeID string) error
	CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error)

	// Metrics
	RecordLocalRuntimeMetrics(ctx context.Context, metrics *types.LocalRuntimeMetric) error
	GetLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time, limit int) ([]*types.LocalRuntimeMetric, error)
	GetLatestLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID) (*types.LocalRuntimeMetric, error)
	GetAggregatedLocalRuntimeMetrics(ctx context.Context, since time.Time) (map[string]interface{}, error)

	// Health
	RecordLocalRuntimeHealth(ctx context.Context, health *types.LocalRuntimeHealth) error
	GetLocalRuntimeHealth(ctx context.Context, instanceID uuid.UUID) (*types.LocalRuntimeHealth, error)
}
