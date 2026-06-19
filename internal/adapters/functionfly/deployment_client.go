package functionfly

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

const (
	defaultDeployTimeout = 120 * time.Second
	defaultStatusTimeout = 10 * time.Second
)

type DeploymentClient struct {
	httpClient *http.Client
	edgeURL    string
	apiKey     string
}

type EdgeDeploymentClientInterface interface {
	RegisterFunction(ctx context.Context, spec *common.DeploymentSpec) (*DeployResponse, error)
	SetEnvironment(ctx context.Context, appName string, envVars, secrets map[string]string) error
	BindRoutes(ctx context.Context, appName string, routes []common.RouteBinding) error
	GetFunctionStatus(ctx context.Context, appName string) (*StatusResponse, error)
	DeleteFunction(ctx context.Context, appName string) error
}

var _ EdgeDeploymentClientInterface = (*DeploymentClient)(nil)

type DeployResponse struct {
	Success      bool                   `json:"success"`
	DeploymentID string                `json:"deployment_id,omitempty"`
	Status       common.DeploymentStatus `json:"status,omitempty"`
	Message      string                `json:"message,omitempty"`
	DeploymentURL string               `json:"deployment_url,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type StatusResponse struct {
	Exists   bool                     `json:"exists"`
	Status   common.DeploymentStatus `json:"status"`
	Message  string                  `json:"message,omitempty"`
	Deployed bool                     `json:"deployed"`
	Region   string                  `json:"region,omitempty"`
	Version  string                  `json:"version,omitempty"`
}

func NewDeploymentClient(edgeURL, apiKey string) *DeploymentClient {
	return &DeploymentClient{
		httpClient: &http.Client{
			Timeout: defaultDeployTimeout,
		},
		edgeURL: edgeURL,
		apiKey:  apiKey,
	}
}

func (c *DeploymentClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.edgeURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	return c.httpClient.Do(req)
}

func (c *DeploymentClient) RegisterFunction(ctx context.Context, spec *common.DeploymentSpec) (*DeployResponse, error) {
	// Prepare the deployment payload
	payload := map[string]interface{}{
		"app_name":   spec.AppName,
		"runtime":    string(spec.Runtime),
		"version":    spec.Version,
		"environment": spec.Environment,
		"routes":     spec.Routes,
	}

	// Include artifact as base64 if provided (for WASM/binary runtimes)
	if len(spec.Artifact) > 0 {
		payload["artifact_size"] = len(spec.Artifact)
		payload["artifact_checksum"] = fmt.Sprintf("%x", checksum(spec.Artifact))
	}

	// Include environment variables and secrets
	if len(spec.EnvVars) > 0 {
		payload["env_vars"] = spec.EnvVars
	}
	if len(spec.Secrets) > 0 {
		payload["secrets"] = spec.Secrets
	}

	// Include provider config
	if spec.ProviderConfig != nil {
		for k, v := range spec.ProviderConfig {
			if k != "url" && k != "wasm_url" {
				payload[k] = v
			}
		}
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/internal/v1/deployments/register", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to register function: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// Fallback: try to create via legacy endpoint
		return c.registerFunctionLegacy(ctx, spec)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result DeployResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *DeploymentClient) registerFunctionLegacy(ctx context.Context, spec *common.DeploymentSpec) (*DeployResponse, error) {
	// Legacy endpoint for backward compatibility
	payload := map[string]interface{}{
		"name":     spec.AppName,
		"runtime":  string(spec.Runtime),
		"version":  spec.Version,
		"env_vars": spec.EnvVars,
		"secrets":  spec.Secrets,
		"routes":   spec.Routes,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/functions", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create function (legacy): %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("function creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result DeployResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode legacy response: %w", err)
	}

	return &result, nil
}

func (c *DeploymentClient) SetEnvironment(ctx context.Context, appName string, envVars, secrets map[string]string) error {
	payload := map[string]interface{}{
		"env_vars": envVars,
		"secrets":  secrets,
	}

	resp, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/internal/v1/deployments/%s/env", appName), payload)
	if err != nil {
		return fmt.Errorf("failed to set environment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set environment failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *DeploymentClient) BindRoutes(ctx context.Context, appName string, routes []common.RouteBinding) error {
	if len(routes) == 0 {
		return nil
	}

	// Convert RouteBinding to simple route strings
	routePatterns := make([]string, 0, len(routes))
	for _, r := range routes {
		routePatterns = append(routePatterns, r.Pattern)
	}

	payload := map[string]interface{}{
		"routes": routePatterns,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/internal/v1/deployments/%s/routes", appName), payload)
	if err != nil {
		return fmt.Errorf("failed to bind routes: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bind routes failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *DeploymentClient) GetFunctionStatus(ctx context.Context, appName string) (*StatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultStatusTimeout)
	defer cancel()

	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/internal/v1/deployments/%s/status", appName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get function status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return &StatusResponse{
			Exists:   false,
			Status:   common.DeploymentStatusFailed,
			Message:  "function not found",
			Deployed: false,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result StatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	result.Exists = true
	result.Deployed = result.Status == common.DeploymentStatusSuccess

	return &result, nil
}

func (c *DeploymentClient) DeleteFunction(ctx context.Context, appName string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/internal/v1/deployments/%s", appName), nil)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *DeploymentClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.doRequest(ctx, http.MethodGet, "/healthz", nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// checksum computes a SHA-256 hash of the artifact for verification
func checksum(data []byte) [32]byte {
	hash := sha256.Sum256(data)
	return hash
}
