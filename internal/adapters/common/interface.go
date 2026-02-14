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

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	OK           bool
	StatusCode   int
	LatencyMs    int
	ErrorMessage string
	Region       string
	Version      string
}
