package vercel

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
	ProviderName   = "vercel"
	RequestTimeout = 25 * time.Second // Vercel can be slower than Workers
)

// VercelAdapter implements the ProviderAdapter interface for Vercel
type VercelAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewVercelAdapter creates a new Vercel adapter
func NewVercelAdapter() *VercelAdapter {
	return &VercelAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetName returns the provider name
func (a *VercelAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates Vercel specific configuration
func (a *VercelAdapter) ValidateConfig(region, urlStr string) error {
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

	// Validate URL format - should be vercel.app domain or custom domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a vercel.app subdomain or custom domain
	if !strings.HasSuffix(parsedURL.Host, ".vercel.app") &&
		!strings.Contains(parsedURL.Host, ".") {
		return fmt.Errorf("URL must be a vercel.app subdomain or custom domain, got: %s", parsedURL.Host)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}

	return nil
}

// GetRegions returns available Vercel regions
// Vercel Edge Functions run in multiple regions, Serverless Functions in specific regions
func (a *VercelAdapter) GetRegions() []string {
	return []string{
		"arn1", // Stockholm (ARN)
		"bom1", // Mumbai (BOM)
		"cdg1", // Paris (CDG)
		"cle1", // Cleveland (CLE)
		"cpt1", // Cape Town (CPT)
		"dub1", // Dublin (DUB)
		"fra1", // Frankfurt (FRA)
		"gru1", // São Paulo (GRU)
		"hkg1", // Hong Kong (HKG)
		"hnd1", // Tokyo (HND)
		"iad1", // Washington DC (IAD)
		"icn1", // Seoul (ICN)
		"jnb1", // Johannesburg (JNB)
		"lax1", // Los Angeles (LAX)
		"lhr1", // London (LHR)
		"pdx1", // Portland (PDX)
		"sfo1", // San Francisco (SFO)
		"sin1", // Singapore (SIN)
		"syd1", // Sydney (SYD)
	}
}

// HealthCheck performs health checks for Vercel deployments
func (a *VercelAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
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

	// Add Vercel-specific headers
	req.Header.Set("User-Agent", "FunctionFly-HealthCheck/1.0")

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

	// Vercel should return 200 for healthy
	result := &common.HealthCheckResult{
		OK:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}

	if !result.OK {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to extract Vercel-specific headers
	if region := resp.Header.Get("x-vercel-region"); region != "" {
		result.Region = region
	}
	if version := resp.Header.Get("x-vercel-deployment-id"); version != "" {
		result.Version = version
	}

	return result, nil
}

// SignRequest adds Vercel-specific headers and request signing
func (a *VercelAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	// Add Vercel-specific headers that may be expected
	req.Header.Set("x-forwarded-host", req.Host)
	req.Header.Set("x-forwarded-proto", "https")

	// Add FunctionFly request signing
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for Vercel requests
func (a *VercelAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}

// Deploy implements the DeploymentAdapter interface
func (a *VercelAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	// Extract standardized and Vercel-specific config
	var apiToken, teamID, projectName string
	if spec.ProviderConfig != nil {
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := spec.ProviderConfig["team_id"].(string); ok {
			teamID = team
		}
		if project, ok := spec.ProviderConfig["project_name"].(string); ok {
			projectName = project
		}
	}

	// Use standardized app name as fallback for project name if not specified
	if projectName == "" && spec.AppName != "" {
		projectName = spec.AppName
	}

	if apiToken == "" {
		return nil, fmt.Errorf("missing required Vercel config: api_token")
	}

	if projectName == "" {
		projectName = "functionfly-deployment" // Default project name
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
	client := NewVercelDeploymentClient(apiToken, teamID)
	return client.Deploy(ctx, spec.Artifact, projectName, env)
}

// SetEnv implements the DeploymentAdapter interface
func (a *VercelAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	// Extract Vercel-specific config
	var apiToken, teamID, projectID string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := providerConfig["team_id"].(string); ok {
			teamID = team
		}
		if project, ok := providerConfig["project_id"].(string); ok {
			projectID = project
		}
	}

	if apiToken == "" || projectID == "" {
		return fmt.Errorf("missing required Vercel config: api_token, project_id")
	}

	// Create deployment client and set environment variables
	client := NewVercelDeploymentClient(apiToken, teamID)
	return client.SetEnvironmentVariables(ctx, projectID, envVars, secrets)
}

// BindRoutes implements the DeploymentAdapter interface
func (a *VercelAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	// Extract Vercel-specific config
	var apiToken, teamID, projectID string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := providerConfig["team_id"].(string); ok {
			teamID = team
		}
		if project, ok := providerConfig["project_id"].(string); ok {
			projectID = project
		}
	}

	if apiToken == "" || projectID == "" {
		return fmt.Errorf("missing required Vercel config: api_token, project_id")
	}

	// Create deployment client and bind routes
	client := NewVercelDeploymentClient(apiToken, teamID)
	return client.BindRoutes(ctx, projectID, routes)
}

// GetDeploymentStatus implements the DeploymentAdapter interface
func (a *VercelAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	// Extract Vercel-specific config
	var apiToken, teamID string
	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := providerConfig["team_id"].(string); ok {
			teamID = team
		}
	}

	if apiToken == "" {
		return common.DeploymentStatusFailed, fmt.Errorf("missing required Vercel config: api_token")
	}

	// Create deployment client and get status
	client := NewVercelDeploymentClient(apiToken, teamID)
	return client.GetDeploymentStatus(ctx, deploymentID)
}

// Rollback implements the DeploymentAdapter interface
func (a *VercelAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	// Extract Vercel-specific config from provider config
	var apiToken, teamID, projectName string
	if spec.ProviderConfig != nil {
		if token, ok := spec.ProviderConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := spec.ProviderConfig["team_id"].(string); ok {
			teamID = team
		}
		if project, ok := spec.ProviderConfig["project_name"].(string); ok {
			projectName = project
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("missing required Vercel config: api_token")
	}

	if projectName == "" {
		projectName = "functionfly-deployment" // Default project name
	}

	// Combine env vars and secrets for rollback deployment
	env := make(map[string]string)
	for k, v := range spec.EnvVars {
		env[k] = v
	}
	for k, v := range spec.Secrets {
		env[k] = v
	}

	// Create deployment client and rollback (redeploy)
	client := NewVercelDeploymentClient(apiToken, teamID)
	return client.Rollback(ctx, spec.Artifact, projectName, env)
}

// LinkProject links a FunctionFly app to a Vercel project
func (a *VercelAdapter) LinkProject(ctx context.Context, providerConfig map[string]interface{}, functionFlyAppID, environment string) (*common.DeploymentResult, error) {
	var apiToken, teamID, projectName string

	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := providerConfig["team_id"].(string); ok {
			teamID = team
		}
		if project, ok := providerConfig["project_name"].(string); ok {
			projectName = project
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("missing required Vercel config: api_token")
	}

	if projectName == "" {
		return nil, fmt.Errorf("missing required Vercel config: project_name")
	}

	client := NewVercelDeploymentClient(apiToken, teamID)

	result, err := client.LinkProject(ctx, projectName, functionFlyAppID, environment)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("project linking failed: %v", err),
		}, nil
	}

	return &common.DeploymentResult{
		DeploymentID: result.VercelProjectID,
		Status:       common.DeploymentStatusSuccess,
		Message:      fmt.Sprintf("Successfully linked to Vercel project: %s", result.VercelProjectName),
		Metadata: map[string]interface{}{
			"vercel_project_id":   result.VercelProjectID,
			"vercel_project_name": result.VercelProjectName,
			"functionfly_app_id":  result.FunctionFlyAppID,
			"linked_at":           result.LinkedAt.Format(time.RFC3339),
			"environment":         result.Environment,
		},
	}, nil
}

// GetLinkedProject gets the linked Vercel project info
func (a *VercelAdapter) GetLinkedProject(ctx context.Context, providerConfig map[string]interface{}) (*common.DeploymentResult, error) {
	var apiToken, teamID, projectName string

	if providerConfig != nil {
		if token, ok := providerConfig["api_token"].(string); ok {
			apiToken = token
		}
		if team, ok := providerConfig["team_id"].(string); ok {
			teamID = team
		}
		if project, ok := providerConfig["project_name"].(string); ok {
			projectName = project
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("missing required Vercel config: api_token")
	}

	if projectName == "" {
		return nil, fmt.Errorf("missing required Vercel config: project_name")
	}

	client := NewVercelDeploymentClient(apiToken, teamID)

	project, err := client.GetLinkedProject(ctx, projectName)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to get linked project: %v", err),
		}, nil
	}

	if project == nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "No linked Vercel project found",
		}, nil
	}

	return &common.DeploymentResult{
		DeploymentID: project.ID,
		Status:       common.DeploymentStatusSuccess,
		Message:      fmt.Sprintf("Vercel project: %s", project.Name),
		Metadata: map[string]interface{}{
			"vercel_project_id":   project.ID,
			"vercel_project_name": project.Name,
			"framework":           project.Framework,
			"created_at":          project.CreatedAt,
			"updated_at":          project.UpdatedAt,
		},
	}, nil
}
