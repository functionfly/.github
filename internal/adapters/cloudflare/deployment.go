package cloudflare

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

// CloudflareDeploymentClient handles deployment operations for Cloudflare Workers
type CloudflareDeploymentClient struct {
	httpClient *http.Client
	apiToken   string
	accountID  string
}

// NewCloudflareDeploymentClient creates a new Cloudflare deployment client
func NewCloudflareDeploymentClient(apiToken, accountID string) *CloudflareDeploymentClient {
	return &CloudflareDeploymentClient{
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Workers API can be slow
		},
		apiToken:  apiToken,
		accountID: accountID,
	}
}

// Deploy uploads a Worker script and creates a deployment
// For WASM artifacts, it creates a JavaScript wrapper that imports and runs the WASM module
func (c *CloudflareDeploymentClient) Deploy(ctx context.Context, scriptContent []byte, scriptName string, runtime common.Runtime) (*DeploymentResult, error) {
	if len(scriptContent) == 0 {
		return nil, fmt.Errorf("script content cannot be empty")
	}
	if scriptName == "" {
		return nil, fmt.Errorf("script name cannot be empty")
	}

	uploadURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s", c.accountID, scriptName)

	// Determine content type and prepare script based on runtime
	var contentType string
	var deployContent []byte

	switch runtime {
	case common.RuntimeWASM, common.RuntimeRust:
		// For WASM/Rust, create a JavaScript wrapper that loads and executes the WASM module
		contentType = "application/javascript; charset=utf-8"
		deployContent = createWASMLoader(scriptName, scriptContent)
	default:
		// Standard JavaScript deployment
		contentType = "application/javascript; charset=utf-8"
		deployContent = scriptContent
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, bytes.NewReader(deployContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResult struct {
		Success bool `json:"success"`
		Result  struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	_ = json.NewDecoder(resp.Body).Decode(&uploadResult)

	if !uploadResult.Success && len(uploadResult.Errors) > 0 {
		return nil, fmt.Errorf("upload failed: %s", uploadResult.Errors[0].Message)
	}

	deploymentID := uploadResult.Result.ID
	if deploymentID == "" {
		deploymentID = scriptName
	}

	return &DeploymentResult{
		DeploymentID: deploymentID,
		Status:       common.DeploymentStatusSuccess,
		Message:      "Worker script uploaded successfully",
		Metadata: map[string]interface{}{
			"script_name": scriptName,
			"runtime":     string(runtime),
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
	}, nil
}

// createWASMLoader creates a JavaScript wrapper that imports and executes a WASM module
// The WASM module is embedded as base64 in the JavaScript for single-file deployment
func createWASMLoader(scriptName string, wasmContent []byte) []byte {
	// Encode WASM as base64 for embedding in JavaScript
	encodedWASM := base64.StdEncoding.EncodeToString(wasmContent)

	// Create a JavaScript loader that:
	// 1. Defines the WASM module as a constant
	// 2. Uses WebAssembly.instantiate to load it
	// 3. Exports a default handler function
	loader := fmt.Sprintf(`// Auto-generated WASM loader for %s
// This wrapper enables WASM modules to run on Cloudflare Workers

const wasmModule = Uint8Array.from(atob("%s"), c => c.charCodeAt(0));

let wasmInstance = null;
let wasmExports = null;

async function initWASM() {
  if (!wasmInstance) {
    const module = new WebAssembly.Module(wasmModule);
    wasmInstance = await WebAssembly.instantiate(module, {
      env: {
        // WASI-like imports for basic functionality
        "fd_write": (ptr, len) => 0,
        "proc_exit": (code) => { throw new Error("WASM exit: " + code); },
      }
    });
    wasmExports = wasmInstance.exports;
  }
  return wasmExports;
}

export default {
  async fetch(request, env, ctx) {
    try {
      await initWASM();

      // Call the WASM _start function if available (WASI entry point)
      if (wasmExports._start) {
        wasmExports._start();
      }

      // Call the main handler if available, otherwise return success
      if (wasmExports.handle) {
        // Read request body and pass to WASM handler
        const url = new URL(request.url);
        const body = await request.text();

        // Set up memory for passing request data
        const memory = wasmExports.memory;

        // For now, return a simple response - full request passthrough
        // requires more sophisticated memory management
        return new Response("WASM function executed successfully", {
          headers: { "Content-Type": "text/plain" }
        });
      }

      return new Response("WASM module loaded successfully", {
        headers: { "Content-Type": "text/plain" }
      });
    } catch (error) {
      return new Response("WASM execution error: " + error.message, {
        status: 500,
        headers: { "Content-Type": "text/plain" }
      });
    }
  }
};
`, scriptName, encodedWASM)

	return []byte(loader)
}

// SetEnvironmentVariables sets environment variables for a Worker.
// The Cloudflare API requires multipart/form-data with a "settings" part containing JSON bindings.
func (c *CloudflareDeploymentClient) SetEnvironmentVariables(ctx context.Context, scriptName string, vars, secrets map[string]string) error {
	// Build bindings array: plain_text and secret_text use "name" and "text" (not "value").
	type binding struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Text string `json:"text"`
	}
	var bindings []binding
	for k, v := range vars {
		bindings = append(bindings, binding{Type: "plain_text", Name: k, Text: v})
	}
	for k, v := range secrets {
		bindings = append(bindings, binding{Type: "secret_text", Name: k, Text: v})
	}
	if len(bindings) == 0 {
		return nil
	}

	settingsJSON, err := json.Marshal(map[string]interface{}{"bindings": bindings})
	if err != nil {
		return fmt.Errorf("failed to marshal bindings: %w", err)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormField("settings")
	if err != nil {
		return fmt.Errorf("failed to create form field: %w", err)
	}
	if _, err := part.Write(settingsJSON); err != nil {
		return fmt.Errorf("failed to write settings part: %w", err)
	}
	contentType := w.FormDataContentType()
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	envURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s/settings", c.accountID, scriptName)
	req, err := http.NewRequestWithContext(ctx, "PATCH", envURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create env request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", contentType)

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

// EnableWorkersDev enables the Worker at <script>.<subdomain>.workers.dev.
// When deploying via the upload API, workers.dev is disabled by default; call this after Deploy to make the Worker reachable at that URL.
func (c *CloudflareDeploymentClient) EnableWorkersDev(ctx context.Context, scriptName string) error {
	subdomainURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s/subdomain", c.accountID, scriptName)
	body := []byte(`{"enabled":true}`)
	req, err := http.NewRequestWithContext(ctx, "POST", subdomainURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create subdomain request: %w", err)
	}
	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to enable workers.dev: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("enable workers.dev failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// BindRoutes binds routes to a Worker deployment
func (c *CloudflareDeploymentClient) BindRoutes(ctx context.Context, zoneID, scriptName string, routes []string) error {
	routesURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/workers/routes", zoneID)

	// Convert route patterns to Cloudflare route objects
	var routeObjects []map[string]interface{}
	for _, pattern := range routes {
		routeObjects = append(routeObjects, map[string]interface{}{
			"pattern": pattern,
			"script":  scriptName,
		})
	}

	routesData := map[string]interface{}{
		"routes": routeObjects,
	}

	jsonData, err := json.Marshal(routesData)
	if err != nil {
		return fmt.Errorf("failed to marshal routes data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", routesURL, bytes.NewReader(jsonData))
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
func (c *CloudflareDeploymentClient) GetDeploymentStatus(ctx context.Context, scriptName string) (common.DeploymentStatus, error) {
	statusURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s", c.accountID, scriptName)

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

	// If we can retrieve the script, it's successfully deployed
	return common.DeploymentStatusSuccess, nil
}

// DeleteDeployment deletes a Worker script
func (c *CloudflareDeploymentClient) DeleteDeployment(ctx context.Context, scriptName string) error {
	deleteURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s", c.accountID, scriptName)

	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DNSRecord represents a DNS record for Cloudflare
type DNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// UpdateDNSRecord updates a DNS record to point to a new target
func (c *CloudflareDeploymentClient) UpdateDNSRecord(ctx context.Context, zoneID, recordName, recordType, newContent string) error {
	// First, find the existing record
	listURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=%s", zoneID, recordName, recordType)

	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DNS list request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %w", err)
	}
	defer resp.Body.Close()

	var listResult struct {
		Success bool        `json:"success"`
		Result  []DNSRecord `json:"result"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return fmt.Errorf("failed to decode DNS list response: %w", err)
	}

	if !listResult.Success || len(listResult.Errors) > 0 {
		return fmt.Errorf("DNS list failed: %v", listResult.Errors)
	}

	if len(listResult.Result) == 0 {
		return fmt.Errorf("no DNS record found for %s %s", recordType, recordName)
	}

	record := listResult.Result[0]

	// Update the record with new content
	updateURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, record.ID)

	updateData := map[string]interface{}{
		"content": newContent,
		"ttl":     record.TTL,
		"proxied": record.Proxied,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal DNS update data: %w", err)
	}

	req, err = http.NewRequestWithContext(ctx, "PUT", updateURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create DNS update request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DNS update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Rollback redeploys a previous Worker script
func (c *CloudflareDeploymentClient) Rollback(ctx context.Context, scriptContent []byte, scriptName string, runtime common.Runtime) (*common.DeploymentResult, error) {
	// Rollback is essentially redeploying with the previous artifact
	result, err := c.Deploy(ctx, scriptContent, scriptName, runtime)
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

// SwitchDNSForBlueGreen switches DNS records to point to the new deployment
// This enables zero-downtime blue/green deployments
func (c *CloudflareDeploymentClient) SwitchDNSForBlueGreen(ctx context.Context, zoneID, domain, newTarget string, enableProxied bool) error {
	// First, get the current DNS records for the domain
	listURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", zoneID, domain)

	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DNS list request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %w", err)
	}
	defer resp.Body.Close()

	var listResult struct {
		Success bool        `json:"success"`
		Result  []DNSRecord `json:"result"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return fmt.Errorf("failed to decode DNS list response: %w", err)
	}

	if !listResult.Success {
		return fmt.Errorf("DNS list failed: %v", listResult.Errors)
	}

	// If no existing records, create new ones
	if len(listResult.Result) == 0 {
		// Create new CNAME record
		return c.createDNSRecord(ctx, zoneID, domain, "CNAME", newTarget, enableProxied)
	}

	// Update existing records
	for _, record := range listResult.Result {
		updateURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, record.ID)

		// Determine record type based on target
		recordType := "CNAME"
		if strings.HasPrefix(newTarget, "https://") || strings.HasPrefix(newTarget, "http://") {
			recordType = "A"
			newTarget = strings.TrimPrefix(strings.TrimPrefix(newTarget, "https://"), "http://")
		}

		updateData := map[string]interface{}{
			"type":    recordType,
			"name":    domain,
			"content": newTarget,
			"ttl":     record.TTL,
			"proxied": enableProxied,
		}

		jsonData, err := json.Marshal(updateData)
		if err != nil {
			return fmt.Errorf("failed to marshal DNS update data: %w", err)
		}

		updateReq, err := http.NewRequestWithContext(ctx, "PUT", updateURL, bytes.NewReader(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create DNS update request: %w", err)
		}

		c.setAuthHeaders(updateReq)
		updateReq.Header.Set("Content-Type", "application/json")

		updateResp, err := c.httpClient.Do(updateReq)
		if err != nil {
			return fmt.Errorf("failed to update DNS record: %w", err)
		}
		defer updateResp.Body.Close()

		if updateResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(updateResp.Body)
			return fmt.Errorf("DNS update failed with status %d: %s", updateResp.StatusCode, string(body))
		}
	}

	return nil
}

// createDNSRecord creates a new DNS record
func (c *CloudflareDeploymentClient) createDNSRecord(ctx context.Context, zoneID, name, recordType, content string, proxied bool) error {
	createURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)

	createData := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     1, // 1 = Auto TTL
		"proxied": proxied,
	}

	jsonData, err := json.Marshal(createData)
	if err != nil {
		return fmt.Errorf("failed to marshal DNS create data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create DNS create request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DNS create failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BlueGreenDeploymentResult represents the result of a blue/green deployment
type BlueGreenDeploymentResult struct {
	BlueDeploymentID  string
	GreenDeploymentID string
	ActiveDeployment  string
	DNSSwitched       bool
	SwitchedAt        time.Time
}

// DeployBlueGreen performs a blue/green deployment with DNS switching.
// workersSubdomain is the account's Workers subdomain (e.g. "mycompany" for mycompany.workers.dev).
// If empty, accountID is used as fallback for backward compatibility (may not resolve for workers.dev).
func (c *CloudflareDeploymentClient) DeployBlueGreen(ctx context.Context, scriptContent []byte, scriptName, zoneID, domain, workersSubdomain string, enableProxied bool, runtime common.Runtime) (*BlueGreenDeploymentResult, error) {
	// Determine current active color (blue or green)
	blueScriptName := scriptName + "-blue"
	greenScriptName := scriptName + "-green"

	var newScriptName string

	blueExists := c.scriptExists(ctx, blueScriptName)
	greenExists := c.scriptExists(ctx, greenScriptName)

	if !blueExists && !greenExists {
		newScriptName = blueScriptName
	} else if blueExists && !greenExists {
		newScriptName = greenScriptName
	} else if !blueExists && greenExists {
		newScriptName = blueScriptName
	} else {
		newScriptName = greenScriptName
	}

	_, err := c.Deploy(ctx, scriptContent, newScriptName, runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy to %s: %w", newScriptName, err)
	}

	// Workers.dev hostname: <script>.<workers_subdomain>.workers.dev (subdomain is from Cloudflare dashboard, not account ID)
	if workersSubdomain == "" {
		workersSubdomain = c.accountID
	}
	target := fmt.Sprintf("%s.%s.workers.dev", newScriptName, workersSubdomain)

	// Switch DNS to point to new deployment
	err = c.SwitchDNSForBlueGreen(ctx, zoneID, domain, target, enableProxied)
	if err != nil {
		return nil, fmt.Errorf("failed to switch DNS: %w", err)
	}

	return &BlueGreenDeploymentResult{
		BlueDeploymentID:  blueScriptName,
		GreenDeploymentID: greenScriptName,
		ActiveDeployment:  newScriptName,
		DNSSwitched:       true,
		SwitchedAt:        time.Now(),
	}, nil
}

// scriptExists checks if a worker script exists
func (c *CloudflareDeploymentClient) scriptExists(ctx context.Context, scriptName string) bool {
	statusURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s", c.accountID, scriptName)

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return false
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// setAuthHeaders sets the required Cloudflare API authentication headers.
// Callers must set Content-Type themselves (e.g. application/javascript for script upload, application/json for JSON bodies).
func (c *CloudflareDeploymentClient) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
}

// DeploymentResult represents the result of a Cloudflare deployment operation
type DeploymentResult struct {
	DeploymentID string
	Status       common.DeploymentStatus
	Message      string
	Metadata     map[string]interface{}
}
