package fly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/signing"
	"github.com/functionfly/functionfly/internal/storage"
)

const (
	ProviderName   = "fly"
	RequestTimeout = 30 * time.Second
)

// FlyAdapter implements the ProviderAdapter interface for Fly.io
type FlyAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewFlyAdapter creates a new Fly.io adapter
func NewFlyAdapter() *FlyAdapter {
	return &FlyAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *FlyAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Fly.io specific configuration
func (a *FlyAdapter) ValidateConfig(region, urlStr string) error {
	// Validate region
	validRegions := a.GetRegions()
	regionValid := false
	for _, r := range validRegions {
		if r == region {
			regionValid = true
			break
		}
	}
	if !regionValid {
		return fmt.Errorf("invalid region '%s', valid regions: %v", region, validRegions)
	}

	// Validate URL format - should be a fly.dev domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a fly.dev subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".fly.dev") &&
		!strings.Contains(parsedURL.Host, ".") &&
		!strings.HasSuffix(parsedURL.Host, ".internal") {
		return fmt.Errorf("URL must be a fly.dev subdomain, .internal domain, or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Fly.io regions
func (a *FlyAdapter) GetRegions() []string {
	return []string{
		"ams", // Amsterdam
		"arn", // Stockholm
		"atl", // Atlanta
		"bog", // Bogotá
		"bos", // Boston
		"bru", // Brussels
		"cdg", // Paris
		"den", // Denver
		"dfw", // Dallas
		"ewr", // New Jersey
		"eze", // Buenos Aires
		"fra", // Frankfurt
		"gig", // Rio de Janeiro
		"gru", // São Paulo
		"hkg", // Hong Kong
		"iad", // Ashburn
		"jnb", // Johannesburg
		"lax", // Los Angeles
		"lhr", // London
		"mad", // Madrid
		"mia", // Miami
		"nrt", // Tokyo
		"ord", // Chicago
		"phx", // Phoenix
		"pma", // Palm Beach
		"sea", // Seattle
		"sfo", // San Francisco
		"sin", // Singapore
		"syd", // Sydney
		"tsn", // Toronto
		"waw", // Warsaw
		"yyz", // Toronto
	}
}

// HealthCheck performs health checks for Fly.io deployments
func (a *FlyAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
	startTime := time.Now()

	// Check /healthz endpoint
	healthURL := strings.TrimSuffix(backend.URL, "/") + "/healthz"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Add Fly.io specific headers
	req.Header.Set("User-Agent", "FunctionFly-HealthCheck/1.0")
	req.Header.Set("fly-client-ip", "0.0.0.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("health check request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latencyMs := int(time.Since(startTime).Milliseconds())

	// Fly.io should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("health check returned status %d", resp.StatusCode)
	}

	return result, nil
}

// SignRequest adds Fly.io specific headers/signatures to requests
func (a *FlyAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Fly.io specific headers
	req.Header.Set("fly-client-ip", "0.0.0.0")
	req.Header.Set("fly-forwarded-proto", "https")

	// Use HMAC signing if configured
	if backend.SharedSecret != "" {
		return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
	}

	return nil
}

// GetRequestTimeout returns the recommended timeout for requests to Fly.io
func (a *FlyAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}

// NewFlyDeploymentAdapter creates a new Fly.io adapter (alias for server.go compatibility)
func NewFlyDeploymentAdapter() *FlyAdapter {
	return NewFlyAdapter()
}

// Deploy implements the DeploymentAdapter interface
func (a *FlyAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	var apiToken, appName, orgSlug string
	if spec.ProviderConfig != nil {
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := spec.ProviderConfig["app_name"].(string); ok {
			appName = name
		}
		if org, ok := spec.ProviderConfig["org_slug"].(string); ok {
			orgSlug = org
		}
	}
	if appName == "" && spec.AppName != "" {
		appName = spec.AppName
	}
	if apiToken == "" || appName == "" {
		return nil, fmt.Errorf("missing required Fly.io config: api_token, app_name")
	}

	client := NewFlyDeploymentClient(apiToken)
	if err := client.EnsureApp(ctx, appName, orgSlug); err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to ensure app exists: %v", err),
		}, nil
	}

	result, err := client.Deploy(ctx, spec.Artifact, appName, spec.Version)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("deployment failed: %v", err),
		}, nil
	}

	if len(spec.EnvVars) > 0 {
		if err := client.SetEnvVars(ctx, appName, spec.EnvVars); err != nil {
			return &common.DeploymentResult{
				Status:  common.DeploymentStatusFailed,
				Message: fmt.Sprintf("failed to set env vars: %v", err),
			}, nil
		}
	}
	if len(spec.Secrets) > 0 {
		if err := client.SetSecrets(ctx, appName, spec.Secrets); err != nil {
			return &common.DeploymentResult{
				Status:  common.DeploymentStatusFailed,
				Message: fmt.Sprintf("failed to set secrets: %v", err),
			}, nil
		}
	}

	return &common.DeploymentResult{
		DeploymentID:  result.DeploymentID,
		Status:        result.Status,
		Message:       result.Message,
		DeploymentURL: fmt.Sprintf("https://%s.fly.dev", appName),
		Metadata:      result.Metadata,
	}, nil
}

// SetEnv implements the DeploymentAdapter interface
func (a *FlyAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	var apiToken, appName string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := providerConfig["app_name"].(string); ok {
			appName = name
		}
	}
	if apiToken == "" || appName == "" {
		return fmt.Errorf("missing required Fly.io config: api_token, app_name")
	}
	client := NewFlyDeploymentClient(apiToken)
	if len(envVars) > 0 {
		if err := client.SetEnvVars(ctx, appName, envVars); err != nil {
			return fmt.Errorf("failed to set env vars: %w", err)
		}
	}
	if len(secrets) > 0 {
		if err := client.SetSecrets(ctx, appName, secrets); err != nil {
			return fmt.Errorf("failed to set secrets: %w", err)
		}
	}
	return nil
}

// BindRoutes implements the DeploymentAdapter interface
func (a *FlyAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	var apiToken, appName string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := providerConfig["app_name"].(string); ok {
			appName = name
		}
	}
	if apiToken == "" || appName == "" {
		return fmt.Errorf("missing required Fly.io config: api_token, app_name")
	}
	client := NewFlyDeploymentClient(apiToken)
	for _, route := range routes {
		if route.Domain != "" {
			if err := client.AddCertificate(ctx, appName, route.Domain); err != nil {
				return fmt.Errorf("failed to add certificate for domain %s: %w", route.Domain, err)
			}
		}
	}
	return nil
}

// GetDeploymentStatus implements the DeploymentAdapter interface
func (a *FlyAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	var apiToken, appName string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := providerConfig["app_name"].(string); ok {
			appName = name
		}
	}
	if apiToken == "" || appName == "" {
		return common.DeploymentStatusFailed, fmt.Errorf("missing required Fly.io config: api_token, app_name")
	}
	client := NewFlyDeploymentClient(apiToken)
	return client.GetDeploymentStatus(ctx, appName, deploymentID)
}

// Rollback implements the DeploymentAdapter interface
func (a *FlyAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	var apiToken, appName string
	if spec.ProviderConfig != nil {
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := spec.ProviderConfig["app_name"].(string); ok {
			appName = name
		}
	}
	if appName == "" && spec.AppName != "" {
		appName = spec.AppName
	}
	if apiToken == "" || appName == "" {
		return nil, fmt.Errorf("missing required Fly.io config: api_token, app_name")
	}
	client := NewFlyDeploymentClient(apiToken)
	return client.Rollback(ctx, appName, spec.Version)
}
