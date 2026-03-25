package deno

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
	ProviderName   = "deno-deploy"
	RequestTimeout = 25 * time.Second // Deno Deploy has good performance
)

// DenoAdapter implements the ProviderAdapter interface for Deno Deploy
type DenoAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewDenoAdapter creates a new Deno Deploy adapter
func NewDenoAdapter() *DenoAdapter {
	return &DenoAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *DenoAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Deno Deploy specific configuration
func (a *DenoAdapter) ValidateConfig(region, urlStr string) error {
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

	// Validate URL format - should be deno.dev domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a deno.dev subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".deno.dev") &&
	   !strings.Contains(parsedURL.Host, ".") {
		return fmt.Errorf("URL must be a deno.dev subdomain or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Deno Deploy regions
func (a *DenoAdapter) GetRegions() []string {
	return []string{
		"us-east4", // Northern Virginia (US East)
		"europe-west4", // Eemshaven (Europe)
		"asia-southeast1", // Jurong West (Asia Pacific)
		"us-west2", // Los Angeles (US West)
	}
}

// HealthCheck performs health checks for Deno Deploy deployments
func (a *DenoAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
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

	// Add Deno Deploy specific headers
	req.Header.Set("User-Agent", "FunctionFly-HealthCheck/1.0")
	req.Header.Set("X-Forwarded-Proto", "https")

	// Add request signing
	err = a.SignRequest(req, backend, startTime)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to sign request: %v", err),
		}, nil
	}

	resp, err := a.client.Do(req)
	latencyMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("health check failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Deno Deploy should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to extract Deno Deploy specific headers
	if region := resp.Header.Get("x-served-by"); region != "" {
		result.Region = region
	}
	if version := resp.Header.Get("x-deployment-id"); version != "" {
		result.Version = version
	}

	return result, nil
}

// SignRequest adds Deno Deploy specific headers and request signing
func (a *DenoAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Deno Deploy specific headers that may be expected
	req.Header.Set("x-forwarded-host", req.Host)
	req.Header.Set("x-forwarded-proto", "https")

	// Add FunctionFly request signing
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for Deno Deploy requests
func (a *DenoAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}

// Deploy implements the DeploymentAdapter interface
func (a *DenoAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	// Extract Deno Deploy-specific config from provider config
	var apiToken, projectID, domain string
	if spec.ProviderConfig != nil {
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if project, ok := spec.ProviderConfig["project_id"].(string); ok {
			projectID = project
		}
		if dom, ok := spec.ProviderConfig["domain"].(string); ok {
			domain = dom
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("missing required Deno Deploy config: api_token")
	}

	if projectID == "" {
		projectID = "functionfly-deployment" // Default project ID
	}

	// Determine runtime (default to JavaScript if not specified)
	runtime := spec.Runtime
	if runtime == "" {
		runtime = common.RuntimeJavaScript
	}

	// Combine env vars and secrets for deployment
	env := make(map[string]string)
	for k, v := range spec.EnvVars {
		env[k] = v
	}
	for k, v := range spec.Secrets {
		env[k] = v
	}

	// Create deployment client and deploy
	client := NewDenoDeploymentClient(apiToken, projectID)
	return client.Deploy(ctx, spec.Artifact, domain, env, runtime)
}

// SetEnv implements the DeploymentAdapter interface
func (a *DenoAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	// Extract Deno Deploy-specific config
	var apiToken, projectID string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if project, ok := providerConfig["project_id"].(string); ok {
			projectID = project
		}
	}

	if apiToken == "" || projectID == "" {
		return fmt.Errorf("missing required Deno Deploy config: api_token, project_id")
	}

	// Create deployment client and set environment variables
	client := NewDenoDeploymentClient(apiToken, projectID)
	return client.SetEnvironmentVariables(ctx, deploymentID, envVars, secrets)
}

// BindRoutes implements the DeploymentAdapter interface
func (a *DenoAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	// Extract Deno Deploy-specific config
	var apiToken, projectID string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if project, ok := providerConfig["project_id"].(string); ok {
			projectID = project
		}
	}

	if apiToken == "" || projectID == "" {
		return fmt.Errorf("missing required Deno Deploy config: api_token, project_id")
	}

	// Create deployment client and bind routes
	client := NewDenoDeploymentClient(apiToken, projectID)
	return client.BindRoutes(ctx, deploymentID, routes)
}

// GetDeploymentStatus implements the DeploymentAdapter interface
func (a *DenoAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	// Extract Deno Deploy-specific config
	var apiToken, projectID string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if project, ok := providerConfig["project_id"].(string); ok {
			projectID = project
		}
	}

	if apiToken == "" || projectID == "" {
		return common.DeploymentStatusFailed, fmt.Errorf("missing required Deno Deploy config: api_token, project_id")
	}

	// Create deployment client and get status
	client := NewDenoDeploymentClient(apiToken, projectID)
	return client.GetDeploymentStatus(ctx, deploymentID)
}

// Rollback implements the DeploymentAdapter interface
func (a *DenoAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	// Extract Deno Deploy-specific config from provider config
	var apiToken, projectID, domain string
	if spec.ProviderConfig != nil {
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if project, ok := spec.ProviderConfig["project_id"].(string); ok {
			projectID = project
		}
		if dom, ok := spec.ProviderConfig["domain"].(string); ok {
			domain = dom
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("missing required Deno Deploy config: api_token")
	}

	if projectID == "" {
		projectID = "functionfly-deployment" // Default project ID
	}

	// Combine env vars and secrets for rollback deployment
	env := make(map[string]string)
	for k, v := range spec.EnvVars {
		env[k] = v
	}
	for k, v := range spec.Secrets {
		env[k] = v
	}

	// Determine runtime (default to JavaScript if not specified)
	runtime := spec.Runtime
	if runtime == "" {
		runtime = common.RuntimeJavaScript
	}

	// Create deployment client and rollback (redeploy)
	client := NewDenoDeploymentClient(apiToken, projectID)
	return client.Rollback(ctx, spec.Artifact, domain, env, runtime)
}
