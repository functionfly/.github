package cloudflare

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
	ProviderName   = "workers"
	RequestTimeout = 30 * time.Second
)

// CloudflareAdapter implements the ProviderAdapter interface for Cloudflare Workers
type CloudflareAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewCloudflareAdapter creates a new Cloudflare adapter
func NewCloudflareAdapter() *CloudflareAdapter {
	return &CloudflareAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *CloudflareAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Cloudflare Workers specific configuration
func (a *CloudflareAdapter) ValidateConfig(region, urlStr string) error {
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

	// Validate URL format - should be workers.dev domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a workers.dev subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".workers.dev") &&
		!strings.Contains(parsedURL.Host, ".") {
		return fmt.Errorf("URL must be a workers.dev subdomain or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Cloudflare Workers regions
func (a *CloudflareAdapter) GetRegions() []string {
	return []string{
		"us-east-1",      // Northern Virginia
		"us-west-1",      // Northern California
		"eu-west-1",      // Ireland
		"ap-southeast-1", // Singapore
		"ap-northeast-1", // Tokyo
		"ap-southeast-2", // Sydney
		"sa-east-1",      // São Paulo
		"af-south-1",     // Cape Town
	}
}

// HealthCheck performs health checks for Cloudflare Workers
func (a *CloudflareAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
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

	// Cloudflare Workers should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to extract version info from headers (optional)
	if version := resp.Header.Get("X-Workers-Version"); version != "" {
		result.Version = version
	}

	return result, nil
}

// SignRequest adds Cloudflare-specific headers and request signing
func (a *CloudflareAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Cloudflare-specific headers that Workers expect
	req.Header.Set("CF-Ray", "")           // Will be set by Cloudflare
	req.Header.Set("CF-Connecting-IP", "") // Will be set by Cloudflare
	req.Header.Set("CF-IPCountry", "")     // Will be set by Cloudflare

	// Add FunctionFly request signing
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for Cloudflare Workers requests
func (a *CloudflareAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}

// Deploy implements the DeploymentAdapter interface
func (a *CloudflareAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	// Extract standardized and Cloudflare-specific config
	var accountID, apiToken, scriptName string
	if spec.ProviderConfig != nil {
		if aid, ok := spec.ProviderConfig["account_id"].(string); ok {
			accountID = aid
		}
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := spec.ProviderConfig["script_name"].(string); ok {
			scriptName = name
		}
	}

	// Use standardized app name as fallback for script name if not specified
	if scriptName == "" && spec.AppName != "" {
		scriptName = spec.AppName
	}

	if accountID == "" || apiToken == "" || scriptName == "" {
		return nil, fmt.Errorf("missing required Cloudflare config: account_id, api_token, script_name")
	}

	client := NewCloudflareDeploymentClient(apiToken, accountID)

	// Determine runtime (default to JavaScript if not specified)
	runtime := spec.Runtime
	if runtime == "" {
		runtime = common.RuntimeJavaScript
	}

	// Deploy the script with runtime type
	result, err := client.Deploy(ctx, spec.Artifact, scriptName, runtime)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("deployment failed: %v", err),
		}, nil
	}

	// Set environment variables if provided
	if len(spec.EnvVars) > 0 || len(spec.Secrets) > 0 {
		if err := client.SetEnvironmentVariables(ctx, scriptName, spec.EnvVars, spec.Secrets); err != nil {
			return &common.DeploymentResult{
				Status:  common.DeploymentStatusFailed,
				Message: fmt.Sprintf("failed to set environment variables: %v", err),
			}, nil
		}
	}

	return &common.DeploymentResult{
		DeploymentID: result.DeploymentID,
		Status:       result.Status,
		Message:      result.Message,
		Metadata:     result.Metadata,
	}, nil
}

// SetEnv implements the DeploymentAdapter interface
func (a *CloudflareAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	// Extract Cloudflare-specific config
	var accountID, apiToken, scriptName string
	if providerConfig != nil {
		if aid, ok := providerConfig["account_id"].(string); ok {
			accountID = aid
		}
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := providerConfig["script_name"].(string); ok {
			scriptName = name
		}
	}

	if accountID == "" || apiToken == "" || scriptName == "" {
		return fmt.Errorf("missing required Cloudflare config: account_id, api_token, script_name")
	}

	client := NewCloudflareDeploymentClient(apiToken, accountID)
	return client.SetEnvironmentVariables(ctx, scriptName, envVars, secrets)
}

// BindRoutes implements the DeploymentAdapter interface
func (a *CloudflareAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	// Extract Cloudflare-specific config
	var accountID, apiToken, zoneID string
	if providerConfig != nil {
		if aid, ok := providerConfig["account_id"].(string); ok {
			accountID = aid
		}
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if zid, ok := providerConfig["zone_id"].(string); ok {
			zoneID = zid
		}
	}

	if accountID == "" || apiToken == "" || zoneID == "" {
		return fmt.Errorf("missing required Cloudflare config: account_id, api_token, zone_id")
	}

	client := NewCloudflareDeploymentClient(apiToken, accountID)

	// Convert RouteBinding to string routes
	var routePatterns []string
	for _, route := range routes {
		routePatterns = append(routePatterns, route.Pattern)
	}

	return client.BindRoutes(ctx, zoneID, deploymentID, routePatterns)
}

// GetDeploymentStatus implements the DeploymentAdapter interface
func (a *CloudflareAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	// Extract Cloudflare-specific config
	var accountID, apiToken string
	if providerConfig != nil {
		if aid, ok := providerConfig["account_id"].(string); ok {
			accountID = aid
		}
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
	}

	if accountID == "" || apiToken == "" {
		return common.DeploymentStatusFailed, fmt.Errorf("missing required Cloudflare config: account_id, api_token")
	}

	client := NewCloudflareDeploymentClient(apiToken, accountID)
	return client.GetDeploymentStatus(ctx, deploymentID)
}

// Rollback implements the DeploymentAdapter interface
func (a *CloudflareAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	// Extract Cloudflare-specific config from provider config
	var accountID, apiToken, scriptName string
	if spec.ProviderConfig != nil {
		if aid, ok := spec.ProviderConfig["account_id"].(string); ok {
			accountID = aid
		}
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := spec.ProviderConfig["script_name"].(string); ok {
			scriptName = name
		}
	}

	if accountID == "" || apiToken == "" || scriptName == "" {
		return nil, fmt.Errorf("missing required Cloudflare config: account_id, api_token, script_name")
	}

	// Determine runtime (default to JavaScript if not specified)
	runtime := spec.Runtime
	if runtime == "" {
		runtime = common.RuntimeJavaScript
	}

	// For Cloudflare rollback, we redeploy the previous artifact
	client := NewCloudflareDeploymentClient(apiToken, accountID)
	return client.Rollback(ctx, spec.Artifact, scriptName, runtime)
}

// DeployBlueGreen performs a blue/green deployment with DNS switching
func (a *CloudflareAdapter) DeployBlueGreen(ctx context.Context, spec *common.DeploymentSpec, zoneID, domain string, enableProxied bool) (*common.DeploymentResult, error) {
	// Extract Cloudflare-specific config
	var accountID, apiToken, scriptName string
	if spec.ProviderConfig != nil {
		if aid, ok := spec.ProviderConfig["account_id"].(string); ok {
			accountID = aid
		}
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if name, ok := spec.ProviderConfig["script_name"].(string); ok {
			scriptName = name
		}
	}

	// Use standardized app name as fallback
	if scriptName == "" && spec.AppName != "" {
		scriptName = spec.AppName
	}

	if accountID == "" || apiToken == "" || scriptName == "" {
		return nil, fmt.Errorf("missing required Cloudflare config: account_id, api_token, script_name")
	}

	if zoneID == "" || domain == "" {
		return nil, fmt.Errorf("missing required blue/green config: zone_id, domain")
	}

	// Determine runtime (default to JavaScript if not specified)
	runtime := spec.Runtime
	if runtime == "" {
		runtime = common.RuntimeJavaScript
	}

	// Workers.dev CNAME target: script-name.<workers_subdomain>.workers.dev (set in Cloudflare Workers dashboard)
	var workersSubdomain string
	if spec.ProviderConfig != nil {
		if s, ok := spec.ProviderConfig["workers_subdomain"].(string); ok {
			workersSubdomain = s
		}
	}

	client := NewCloudflareDeploymentClient(apiToken, accountID)

	// Perform blue/green deployment
	result, err := client.DeployBlueGreen(ctx, spec.Artifact, scriptName, zoneID, domain, workersSubdomain, enableProxied, runtime)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("blue/green deployment failed: %v", err),
		}, nil
	}

	return &common.DeploymentResult{
		DeploymentID: result.ActiveDeployment,
		Status:       common.DeploymentStatusSuccess,
		Message:      fmt.Sprintf("Blue/Green deployment complete: %s is now active", result.ActiveDeployment),
		Metadata: map[string]interface{}{
			"blue_deployment":  result.BlueDeploymentID,
			"green_deployment": result.GreenDeploymentID,
			"active_color":     result.ActiveDeployment,
			"dns_switched":     result.DNSSwitched,
			"switched_at":      result.SwitchedAt.Format(time.RFC3339),
		},
	}, nil
}
