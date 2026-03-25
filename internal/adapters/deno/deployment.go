package deno

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

// DenoDeploymentClient handles deployment operations for Deno Deploy
type DenoDeploymentClient struct {
	httpClient *http.Client
	apiToken   string
	projectID  string
}

// NewDenoDeploymentClient creates a new Deno deployment client
func NewDenoDeploymentClient(apiToken, projectID string) *DenoDeploymentClient {
	return &DenoDeploymentClient{
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Deno Deploy API can be slow
		},
		apiToken:  apiToken,
		projectID: projectID,
	}
}

// Deploy uploads a deployment to Deno Deploy
// For WASM artifacts, Deno can run them natively
func (c *DenoDeploymentClient) Deploy(ctx context.Context, scriptContent []byte, domain string, env map[string]string, runtime common.Runtime) (*common.DeploymentResult, error) {
	// Determine how to deploy based on runtime type
	var payload map[string]interface{}

	switch runtime {
	case common.RuntimeWASM, common.RuntimeRust:
		// Deno can natively run WASM - create a loader script
		wasmLoader := createDenoWASMLoader(scriptContent)
		encodedScript := base64.StdEncoding.EncodeToString([]byte(wasmLoader))
		payload = map[string]interface{}{
			"url":  "data:text/typescript;base64," + encodedScript,
			"envs": env,
		}
	default:
		// Standard JavaScript/TypeScript deployment
		encodedScript := base64.StdEncoding.EncodeToString(scriptContent)
		payload = map[string]interface{}{
			"url":  "data:text/typescript;base64," + encodedScript,
			"envs": env,
		}
	}

	if domain != "" {
		payload["domains"] = []string{domain}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment data: %w", err)
	}

	deployURL := fmt.Sprintf("https://dash.deno.com/api/projects/%s/deployments", c.projectID)

	req, err := http.NewRequestWithContext(ctx, "POST", deployURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployResult struct {
		ID      string `json:"id"`
		URL     string `json:"url"`
		Domain  string `json:"domain"`
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deployResult); err != nil {
		return nil, fmt.Errorf("failed to decode deployment response: %w", err)
	}

	var status common.DeploymentStatus
	switch deployResult.Status {
	case "success", "succeeded":
		status = common.DeploymentStatusSuccess
	case "pending", "deploying":
		status = common.DeploymentStatusDeploying
	case "failed":
		status = common.DeploymentStatusFailed
	default:
		status = common.DeploymentStatusPending
	}

	return &common.DeploymentResult{
		DeploymentID: deployResult.ID,
		Status:       status,
		Message:      deployResult.Message,
		Metadata: map[string]interface{}{
			"url":         deployResult.URL,
			"domain":      deployResult.Domain,
			"project":     c.projectID,
			"runtime":     string(runtime),
			"deployed_at": time.Now().Format(time.RFC3339),
		},
	}, nil
}

// createDenoWASMLoader creates a Deno TypeScript loader for WASM modules
func createDenoWASMLoader(wasmContent []byte) string {
	encodedWASM := base64.StdEncoding.EncodeToString(wasmContent)
	return fmt.Sprintf(`const wasmModule = Uint8Array.from(atob("%s"), c => c.charCodeAt(0));
let wasmInstance = null;
async function initWASM() {
  if (!wasmInstance) {
    const module = new WebAssembly.Module(wasmModule);
    wasmInstance = await WebAssembly.instantiate(module, {wasi_snapshot_preview1:{fd_write:()=>0,proc_exit:()=>{throw new Error("exit")}}});
  }
  return wasmInstance.exports;
}
Deno.serve(async (req) => {
  try {
    const exports = await initWASM();
    if (exports._start) exports._start();
    return new Response("WASM on Deno: OK", {headers:{"Content-Type":"text/plain"}});
  } catch (e) {
    return new Response("Error: "+e.message, {status:500});
  }
});`, encodedWASM)
}

// SetEnvironmentVariables sets environment variables for a deployment
func (c *DenoDeploymentClient) SetEnvironmentVariables(ctx context.Context, deploymentID string, vars, secrets map[string]string) error {
	// Combine vars and secrets (Deno Deploy treats them similarly)
	env := make(map[string]string)
	for k, v := range vars {
		env[k] = v
	}
	for k, v := range secrets {
		env[k] = v
	}

	envData := map[string]interface{}{
		"envs": env,
	}

	jsonData, err := json.Marshal(envData)
	if err != nil {
		return fmt.Errorf("failed to marshal environment data: %w", err)
	}

	envURL := fmt.Sprintf("https://dash.deno.com/api/projects/%s/deployments/%s", c.projectID, deploymentID)

	req, err := http.NewRequestWithContext(ctx, "PATCH", envURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create env request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set environment variables: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set env failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BindRoutes binds routes to a deployment
func (c *DenoDeploymentClient) BindRoutes(ctx context.Context, deploymentID string, routes []common.RouteBinding) error {
	// Extract domains from routes
	domains := make([]string, 0, len(routes))
	for _, route := range routes {
		if route.Domain != "" {
			domains = append(domains, route.Domain)
		}
	}

	if len(domains) == 0 {
		return nil // No domains to bind
	}

	routesData := map[string]interface{}{
		"domains": domains,
	}

	jsonData, err := json.Marshal(routesData)
	if err != nil {
		return fmt.Errorf("failed to marshal routes data: %w", err)
	}

	routesURL := fmt.Sprintf("https://dash.deno.com/api/projects/%s/deployments/%s", c.projectID, deploymentID)

	req, err := http.NewRequestWithContext(ctx, "PATCH", routesURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create routes request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to bind routes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bind routes failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetDeploymentStatus gets the current status of a deployment
func (c *DenoDeploymentClient) GetDeploymentStatus(ctx context.Context, deploymentID string) (common.DeploymentStatus, error) {
	statusURL := fmt.Sprintf("https://dash.deno.com/api/projects/%s/deployments/%s", c.projectID, deploymentID)

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

	if resp.StatusCode == http.StatusNotFound {
		return common.DeploymentStatusFailed, fmt.Errorf("deployment not found")
	}

	if resp.StatusCode != http.StatusOK {
		return common.DeploymentStatusFailed, fmt.Errorf("status check failed with status %d", resp.StatusCode)
	}

	var statusResult struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&statusResult); err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to decode status response: %w", err)
	}

	// Map status to common status
	switch statusResult.Status {
	case "success", "succeeded":
		return common.DeploymentStatusSuccess, nil
	case "pending", "deploying":
		return common.DeploymentStatusDeploying, nil
	case "failed":
		return common.DeploymentStatusFailed, nil
	default:
		return common.DeploymentStatusPending, nil
	}
}

// Rollback redeploys a previous version
func (c *DenoDeploymentClient) Rollback(ctx context.Context, scriptContent []byte, domain string, env map[string]string, runtime common.Runtime) (*common.DeploymentResult, error) {
	// Rollback is essentially redeploying with the previous artifact
	result, err := c.Deploy(ctx, scriptContent, domain, env, runtime)
	if err != nil {
		return nil, err
	}

	// Convert to common.DeploymentResult
	return &common.DeploymentResult{
		DeploymentID: result.DeploymentID,
		Status:       result.Status,
		Message:      result.Message,
		Metadata:     result.Metadata,
	}, nil
}

// setAuthHeaders sets the required Deno Deploy API authentication headers
func (c *DenoDeploymentClient) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
}
