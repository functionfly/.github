package vercel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

const defaultVercelAPIBase = "https://api.vercel.com"

// Vercel API rate limit: avoid bursts (e.g. 100/min); throttle between calls
const rateLimitMinInterval = 200 * time.Millisecond

func getVercelAPIBase() string {
	if v := os.Getenv("VERCEL_API_BASE"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return defaultVercelAPIBase
}

// VercelDeploymentClient handles deployment operations for Vercel
type VercelDeploymentClient struct {
	httpClient *http.Client
	apiToken   string
	teamID     string // Optional team ID for team deployments
	baseURL    string
	lastReq    time.Time
	reqMu      sync.Mutex
}

// VercelProject represents a Vercel project
type VercelProject struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// ProjectLinkResult represents the result of linking a project
type ProjectLinkResult struct {
	VercelProjectID   string
	VercelProjectName string
	FunctionFlyAppID  string
	LinkedAt          time.Time
	Environment       string
}

// NewVercelDeploymentClient creates a new Vercel deployment client (uses VERCEL_API_BASE or default).
func NewVercelDeploymentClient(apiToken, teamID string) *VercelDeploymentClient {
	return NewVercelDeploymentClientWithBase(apiToken, teamID, getVercelAPIBase())
}

// NewVercelDeploymentClientWithBase creates a client with an explicit API base URL (for tests).
func NewVercelDeploymentClientWithBase(apiToken, teamID, baseURL string) *VercelDeploymentClient {
	return &VercelDeploymentClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		apiToken:   apiToken,
		teamID:     teamID,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		lastReq:    time.Time{},
	}
}

func (c *VercelDeploymentClient) rateLimit() {
	c.reqMu.Lock()
	elapsed := time.Since(c.lastReq)
	if elapsed < rateLimitMinInterval {
		time.Sleep(rateLimitMinInterval - elapsed)
	}
	c.lastReq = time.Now()
	c.reqMu.Unlock()
}

// Deploy creates a new deployment on Vercel
func (c *VercelDeploymentClient) Deploy(ctx context.Context, functionContent []byte, projectName string, env map[string]string) (*common.DeploymentResult, error) {
	// Create multipart form data for file upload
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the function file
	filename := "api/index.js" // Vercel expects API routes in api/ directory
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(functionContent); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	// Add deployment metadata
	if err := writer.WriteField("name", projectName); err != nil {
		return nil, fmt.Errorf("failed to write name field: %w", err)
	}

	// Set to production deployment
	if err := writer.WriteField("target", "production"); err != nil {
		return nil, fmt.Errorf("failed to write target field: %w", err)
	}

	// Add environment variables
	for key, value := range env {
		envField := fmt.Sprintf("env_%s", key)
		if err := writer.WriteField(envField, value); err != nil {
			return nil, fmt.Errorf("failed to write env field %s: %w", key, err)
		}
	}

	writer.Close()

	c.rateLimit()
	deployURL := c.baseURL + "/v13/deployments"
	if c.teamID != "" {
		deployURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", deployURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var deployResponse struct {
		UID       string `json:"uid"`
		URL       string `json:"url"`
		State     string `json:"state"`
		CreatedAt int64  `json:"createdAt"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deployResponse); err != nil {
		return nil, fmt.Errorf("failed to parse deployment response: %w", err)
	}

	status := common.DeploymentStatusPending
	switch deployResponse.State {
	case "READY":
		status = common.DeploymentStatusSuccess
	case "ERROR":
		status = common.DeploymentStatusFailed
	case "BUILDING", "DEPLOYING":
		status = common.DeploymentStatusDeploying
	}

	result := &common.DeploymentResult{
		DeploymentID:  deployResponse.UID,
		Status:        status,
		Message:       fmt.Sprintf("Deployment created: %s", deployResponse.URL),
		DeploymentURL: deployResponse.URL,
		Metadata: map[string]interface{}{
			"url":       deployResponse.URL,
			"project":   projectName,
			"state":     deployResponse.State,
			"createdAt": deployResponse.CreatedAt,
		},
	}
	return result, nil
}

// SetEnvironmentVariables updates environment variables for a Vercel project
func (c *VercelDeploymentClient) SetEnvironmentVariables(ctx context.Context, projectID string, vars, secrets map[string]string) error {
	c.rateLimit()
	envURL := fmt.Sprintf("%s/v10/projects/%s/env", c.baseURL, projectID)
	if c.teamID != "" {
		envURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	// Combine vars and secrets (Vercel treats them similarly)
	allEnv := make(map[string]string)
	for k, v := range vars {
		allEnv[k] = v
	}
	for k, v := range secrets {
		allEnv[k] = v
	}

	for key, value := range allEnv {
		c.rateLimit()
		envData := map[string]interface{}{
			"key":    key,
			"value":  value,
			"type":   "plain", // Could be "secret" for encrypted values
			"target": []string{"production", "preview"},
		}

		jsonData, err := json.Marshal(envData)
		if err != nil {
			return fmt.Errorf("failed to marshal env data for %s: %w", key, err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", envURL, bytes.NewReader(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create env request for %s: %w", key, err)
		}

		c.setAuthHeaders(req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to set env var %s: %w", key, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("failed to set env var %s, status: %d", key, resp.StatusCode)
		}
	}

	return nil
}

// GetDeploymentStatus retrieves the status of a Vercel deployment
func (c *VercelDeploymentClient) GetDeploymentStatus(ctx context.Context, deploymentID string) (common.DeploymentStatus, error) {
	c.rateLimit()
	statusURL := fmt.Sprintf("%s/v13/deployments/%s", c.baseURL, deploymentID)
	if c.teamID != "" {
		statusURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to create status request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to get deployment status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.DeploymentStatusFailed, fmt.Errorf("status request failed with status: %d", resp.StatusCode)
	}

	var statusResponse struct {
		State string `json:"state"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to parse status response: %w", err)
	}

	switch statusResponse.State {
	case "READY":
		return common.DeploymentStatusSuccess, nil
	case "ERROR":
		return common.DeploymentStatusFailed, nil
	case "BUILDING", "DEPLOYING":
		return common.DeploymentStatusDeploying, nil
	default:
		return common.DeploymentStatusPending, nil
	}
}

// Rollback redeploys a previous deployment by creating a new deployment with the same code
func (c *VercelDeploymentClient) Rollback(ctx context.Context, functionContent []byte, projectName string, env map[string]string) (*common.DeploymentResult, error) {
	// For Vercel rollback, we redeploy with the previous artifact
	return c.Deploy(ctx, functionContent, projectName, env)
}

// BindDomain adds a custom domain to a Vercel project
func (c *VercelDeploymentClient) BindDomain(ctx context.Context, projectID, domain string) error {
	c.rateLimit()
	domainURL := fmt.Sprintf("%s/v10/projects/%s/domains", c.baseURL, projectID)
	if c.teamID != "" {
		domainURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	domainData := map[string]string{
		"name": domain,
	}

	jsonData, err := json.Marshal(domainData)
	if err != nil {
		return fmt.Errorf("failed to marshal domain data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", domainURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create domain binding request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to bind domain: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("domain binding failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateRoutingConfig updates the vercel.json configuration for routing
func (c *VercelDeploymentClient) UpdateRoutingConfig(ctx context.Context, projectID string, routes []common.RouteBinding) error {
	c.rateLimit()
	configURL := fmt.Sprintf("%s/v1/projects/%s", c.baseURL, projectID)
	if c.teamID != "" {
		configURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	// Convert RouteBinding to Vercel routing configuration
	var vercelRoutes []map[string]interface{}
	for _, route := range routes {
		vercelRoute := map[string]interface{}{
			"source":      route.Pattern,
			"destination": "/", // Default destination, could be made configurable
		}
		vercelRoutes = append(vercelRoutes, vercelRoute)
	}

	configData := map[string]interface{}{
		"functions": map[string]interface{}{
			"api/**/*.js": map[string]interface{}{
				"runtime": "nodejs18.x",
			},
		},
		"rewrites": vercelRoutes,
	}

	jsonData, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal routing config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", configURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create routing config request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update routing config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("routing config update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BindRoutes handles route binding for Vercel deployments
func (c *VercelDeploymentClient) BindRoutes(ctx context.Context, projectID string, routes []common.RouteBinding) error {
	// Handle custom domains
	for _, route := range routes {
		if route.Domain != "" {
			if err := c.BindDomain(ctx, projectID, route.Domain); err != nil {
				return fmt.Errorf("failed to bind domain %s: %w", route.Domain, err)
			}
		}
	}

	// Update routing configuration
	if err := c.UpdateRoutingConfig(ctx, projectID, routes); err != nil {
		return fmt.Errorf("failed to update routing config: %w", err)
	}

	return nil
}

// LinkProject links a FunctionFly app to a Vercel project
func (c *VercelDeploymentClient) LinkProject(ctx context.Context, projectName, functionFlyAppID, environment string) (*ProjectLinkResult, error) {
	// First, check if the Vercel project exists
	project, err := c.GetProject(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Vercel project: %w", err)
	}

	if project == nil {
		// Create new project if it doesn't exist
		newProject, err := c.CreateProject(ctx, projectName)
		if err != nil {
			return nil, fmt.Errorf("failed to create Vercel project: %w", err)
		}
		project = newProject
	}

	return &ProjectLinkResult{
		VercelProjectID:   project.ID,
		VercelProjectName: project.Name,
		FunctionFlyAppID:  functionFlyAppID,
		LinkedAt:          time.Now(),
		Environment:       environment,
	}, nil
}

// GetProject retrieves a Vercel project by name
func (c *VercelDeploymentClient) GetProject(ctx context.Context, projectName string) (*VercelProject, error) {
	c.rateLimit()
	projectURL := fmt.Sprintf("%s/v6/projects/%s", c.baseURL, projectName)
	if c.teamID != "" {
		projectURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", projectURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create project request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get project failed with status: %d", resp.StatusCode)
	}

	var project VercelProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to parse project response: %w", err)
	}

	return &project, nil
}

// CreateProject creates a new Vercel project
func (c *VercelDeploymentClient) CreateProject(ctx context.Context, projectName string) (*VercelProject, error) {
	c.rateLimit()
	createURL := c.baseURL + "/v6/projects"
	if c.teamID != "" {
		createURL += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	projectData := map[string]interface{}{
		"name":      projectName,
		"framework": "nextjs", // Default framework, can be overridden
	}

	jsonData, err := json.Marshal(projectData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create project request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create project failed with status %d: %s", resp.StatusCode, string(body))
	}

	var project VercelProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to parse create project response: %w", err)
	}

	return &project, nil
}

// UnlinkProject removes the link between a FunctionFly app and Vercel project
func (c *VercelDeploymentClient) UnlinkProject(ctx context.Context, projectID string) error {
	// Vercel doesn't have a direct unlink API
	// We just delete the project association from our side
	// The Vercel project remains but is no longer linked
	return nil
}

// GetLinkedProject returns the linked Vercel project for a FunctionFly app
func (c *VercelDeploymentClient) GetLinkedProject(ctx context.Context, projectName string) (*VercelProject, error) {
	return c.GetProject(ctx, projectName)
}

// setAuthHeaders sets the required Vercel API authentication headers
func (c *VercelDeploymentClient) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
}
