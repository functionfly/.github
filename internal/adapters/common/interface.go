package common

import (
	"context"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// ProviderAdapter defines the interface that all provider adapters must implement
type ProviderAdapter interface {
	// GetName returns the provider name (e.g., "workers", "vercel", "fly")
	GetName() string

	// ValidateConfig validates provider-specific configuration
	ValidateConfig(region, url string) error

	// GetRegions returns available regions for this provider
	GetRegions() []string

	// HealthCheck performs a provider-specific health check
	HealthCheck(ctx context.Context, backend *storage.Backend) (*HealthCheckResult, error)

	// SignRequest adds provider-specific headers/signatures to requests
	SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error

	// GetRequestTimeout returns the recommended timeout for requests to this provider
	GetRequestTimeout() time.Duration
}

// DeploymentAdapter extends ProviderAdapter with deployment capabilities
// Providers implement this interface to support deployment operations
type DeploymentAdapter interface {
	ProviderAdapter

	// Deploy creates a new deployment using the provided spec
	Deploy(ctx context.Context, spec *DeploymentSpec) (*DeploymentResult, error)

	// SetEnv updates environment variables for an existing deployment
	SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error

	// BindRoutes binds routes to an existing deployment
	BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []RouteBinding) error

	// GetDeploymentStatus returns the current status of a deployment
	GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (DeploymentStatus, error)

	// Rollback reverts to a previous deployment (optional operation)
	Rollback(ctx context.Context, spec *DeploymentSpec) (*DeploymentResult, error)
}

// ExtendedDeploymentAdapter provides additional deployment capabilities
// These are optional methods that providers can implement for advanced features
type ExtendedDeploymentAdapter interface {
	DeploymentAdapter

	// DeployBlueGreen performs a blue/green deployment with DNS switching (Cloudflare)
	DeployBlueGreen(ctx context.Context, spec *DeploymentSpec, zoneID, domain string, enableProxied bool) (*DeploymentResult, error)

	// LinkProject links FunctionFly project to provider project (Vercel)
	LinkProject(ctx context.Context, providerConfig map[string]interface{}, functionFlyAppID, environment string) (*DeploymentResult, error)

	// GetLinkedProject gets the linked provider project info (Vercel)
	GetLinkedProject(ctx context.Context, providerConfig map[string]interface{}) (*DeploymentResult, error)

	// SetSecrets sets secrets for the deployment (Fly.io)
	SetSecrets(ctx context.Context, providerConfig map[string]interface{}, secrets map[string]string) (*DeploymentResult, error)

	// UnsetSecret removes a secret from the deployment (Fly.io)
	UnsetSecret(ctx context.Context, providerConfig map[string]interface{}, secretName string) (*DeploymentResult, error)

	// ListSecrets lists all secrets for the deployment (Fly.io)
	ListSecrets(ctx context.Context, providerConfig map[string]interface{}) (*DeploymentResult, error)
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	OK           bool
	StatusCode   int
	LatencyMs    int
	ErrorMessage string
	Region       string
	Version      string
}

// DeploymentSpec defines the standardized specification for deploying to any provider
type DeploymentSpec struct {
	// Artifact contains the deployment artifact (bundle bytes or reference)
	Artifact []byte
	// ArtifactKey is an optional reference to stored artifact (e.g., S3 key)
	ArtifactKey string
	// AppName is the standardized application/project name across providers
	AppName string
	// Environment specifies the deployment environment (dev, staging, prod)
	Environment string
	// Version is an optional version identifier for the deployment
	Version string
	// Routes specifies the route patterns to bind to this deployment
	Routes []string
	// EnvVars contains non-secret environment variables
	EnvVars map[string]string
	// Secrets contains secret environment variables
	Secrets map[string]string
	// ProviderConfig contains provider-specific configuration
	ProviderConfig map[string]interface{}
	// Timeout specifies the deployment timeout duration
	Timeout *time.Duration
}

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	DeploymentID  string
	Status        DeploymentStatus
	Message       string
	DeploymentURL string
	Metadata      map[string]interface{}
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusSuccess   DeploymentStatus = "success"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusRollback  DeploymentStatus = "rollback"
)

// RouteBinding represents a route binding configuration
type RouteBinding struct {
	Pattern string
	ZoneID  string // For Cloudflare zones
	Domain  string // For custom domains
}

// DNSRecord represents a DNS record for switching
type DNSRecord struct {
	Type     string // A, AAAA, CNAME
	Name     string
	Value    string
	TTL      int
	Priority int // For MX records
}
