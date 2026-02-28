package fly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

const flyAPIBase = "https://api.fly.io/v1"

// FlyDeploymentClient handles deployment operations for Fly.io
type FlyDeploymentClient struct {
	httpClient *http.Client
	apiToken   string
}

// NewFlyDeploymentClient creates a new Fly.io deployment client
func NewFlyDeploymentClient(apiToken string) *FlyDeploymentClient {
	return &FlyDeploymentClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		apiToken:   apiToken,
	}
}

// FlyDeployResult is the result of a Fly.io deployment operation
type FlyDeployResult struct {
	DeploymentID string
	Status       common.DeploymentStatus
	Message      string
	Metadata     map[string]interface{}
}

func (c *FlyDeploymentClient) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
}

// EnsureApp creates a Fly.io app if it doesn't exist
func (c *FlyDeploymentClient) EnsureApp(ctx context.Context, appName, orgSlug string) error {
	getURL := fmt.Sprintf("%s/apps/%s", flyAPIBase, appName)
	req, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create get app request: %w", err)
	}
	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check app existence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // App already exists
	}
	if resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status checking app: %d - %s", resp.StatusCode, string(body))
	}

	// Create the app
	createData := map[string]interface{}{"app_name": appName}
	if orgSlug != "" {
		createData["org_slug"] = orgSlug
	}
	jsonData, err := json.Marshal(createData)
	if err != nil {
		return fmt.Errorf("failed to marshal create app data: %w", err)
	}

	createReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/apps", flyAPIBase), bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create app request: %w", err)
	}
	c.setAuthHeaders(createReq)

	createResp, err := c.httpClient.Do(createReq)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated && createResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("create app failed with status %d: %s", createResp.StatusCode, string(body))
	}
	return nil
}

// Deploy deploys an artifact to a Fly.io app using the Machines API
func (c *FlyDeploymentClient) Deploy(ctx context.Context, artifact []byte, appName, version string) (*FlyDeployResult, error) {
	machinesURL := fmt.Sprintf("%s/apps/%s/machines", flyAPIBase, appName)

	machineConfig := map[string]interface{}{
		"config": map[string]interface{}{
			"image": fmt.Sprintf("registry.fly.io/%s:%s", appName, version),
			"services": []map[string]interface{}{
				{
					"ports": []map[string]interface{}{
						{"port": 443, "handlers": []string{"tls", "http"}},
						{"port": 80, "handlers": []string{"http"}},
					},
					"protocol":      "tcp",
					"internal_port": 8080,
				},
			},
			"checks": map[string]interface{}{
				"alive": map[string]interface{}{
					"type":     "http",
					"port":     8080,
					"path":     "/healthz",
					"interval": "15s",
					"timeout":  "2s",
				},
			},
		},
	}

	jsonData, err := json.Marshal(machineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal machine config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", machinesURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy request: %w", err)
	}
	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("deploy failed with status %d: %s", resp.StatusCode, string(body))
	}

	var machine struct {
		ID     string `json:"id"`
		State  string `json:"state"`
		Region string `json:"region"`
	}
	if err := json.Unmarshal(body, &machine); err != nil {
		return nil, fmt.Errorf("failed to decode deploy response: %w", err)
	}

	status := common.DeploymentStatusDeploying
	if machine.State == "started" {
		status = common.DeploymentStatusSuccess
	}

	return &FlyDeployResult{
		DeploymentID: machine.ID,
		Status:       status,
		Message:      fmt.Sprintf("Machine %s deployed to region %s", machine.ID, machine.Region),
		Metadata: map[string]interface{}{
			"machine_id":    machine.ID,
			"machine_state": machine.State,
			"region":        machine.Region,
			"app_name":      appName,
			"version":       version,
			"deployed_at":   time.Now().Format(time.RFC3339),
		},
	}, nil
}

// SetEnvVars sets environment variables for a Fly.io app
func (c *FlyDeploymentClient) SetEnvVars(ctx context.Context, appName string, envVars map[string]string) error {
	envURL := fmt.Sprintf("%s/apps/%s/env_vars", flyAPIBase, appName)
	envData := map[string]interface{}{"env": envVars}
	jsonData, err := json.Marshal(envData)
	if err != nil {
		return fmt.Errorf("failed to marshal env vars: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", envURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create env vars request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set env vars: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set env vars failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// SetSecrets sets secrets for a Fly.io app
func (c *FlyDeploymentClient) SetSecrets(ctx context.Context, appName string, secrets map[string]string) error {
	secretsURL := fmt.Sprintf("%s/apps/%s/secrets", flyAPIBase, appName)
	secretsData := map[string]interface{}{"secrets": secrets}
	jsonData, err := json.Marshal(secretsData)
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", secretsURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create secrets request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set secrets: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set secrets failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// AddCertificate adds a custom domain certificate to a Fly.io app
func (c *FlyDeploymentClient) AddCertificate(ctx context.Context, appName, hostname string) error {
	certURL := fmt.Sprintf("%s/apps/%s/certificates", flyAPIBase, appName)
	certData := map[string]interface{}{"hostname": hostname}
	jsonData, err := json.Marshal(certData)
	if err != nil {
		return fmt.Errorf("failed to marshal certificate data: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", certURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create certificate request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add certificate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add certificate failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetDeploymentStatus gets the current status of a Fly.io machine
func (c *FlyDeploymentClient) GetDeploymentStatus(ctx context.Context, appName, machineID string) (common.DeploymentStatus, error) {
	machineURL := fmt.Sprintf("%s/apps/%s/machines/%s", flyAPIBase, appName, machineID)
	req, err := http.NewRequestWithContext(ctx, "GET", machineURL, nil)
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
		return common.DeploymentStatusFailed, fmt.Errorf("machine %s not found in app %s", machineID, appName)
	}
	if resp.StatusCode != http.StatusOK {
		return common.DeploymentStatusFailed, fmt.Errorf("status check failed with status %d", resp.StatusCode)
	}

	var machine struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to decode status response: %w", err)
	}

	switch machine.State {
	case "started":
		return common.DeploymentStatusSuccess, nil
	case "starting":
		return common.DeploymentStatusDeploying, nil
	case "stopping", "stopped", "destroying", "destroyed":
		return common.DeploymentStatusFailed, nil
	default:
		return common.DeploymentStatusPending, nil
	}
}

// Rollback reverts a Fly.io app to a previous release
func (c *FlyDeploymentClient) Rollback(ctx context.Context, appName, version string) (*common.DeploymentResult, error) {
	releasesURL := fmt.Sprintf("%s/apps/%s/releases", flyAPIBase, appName)
	req, err := http.NewRequestWithContext(ctx, "GET", releasesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create releases request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list releases failed with status %d: %s", resp.StatusCode, string(body))
	}

	var releases struct {
		Releases []struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
			Status  string `json:"status"`
		} `json:"releases"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases response: %w", err)
	}

	if len(releases.Releases) < 2 {
		return nil, fmt.Errorf("no previous release available for rollback")
	}

	var previousReleaseID string
	for i, release := range releases.Releases {
		if i == 0 {
			continue
		}
		if release.Status == "complete" {
			previousReleaseID = release.ID
			break
		}
	}
	if previousReleaseID == "" {
		return nil, fmt.Errorf("no successful previous release found for rollback")
	}

	rollbackURL := fmt.Sprintf("%s/apps/%s/releases/%s/rollback", flyAPIBase, appName, previousReleaseID)
	rollbackReq, err := http.NewRequestWithContext(ctx, "POST", rollbackURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback request: %w", err)
	}
	c.setAuthHeaders(rollbackReq)

	rollbackResp, err := c.httpClient.Do(rollbackReq)
	if err != nil {
		return nil, fmt.Errorf("failed to rollback: %w", err)
	}
	defer rollbackResp.Body.Close()

	if rollbackResp.StatusCode != http.StatusOK && rollbackResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(rollbackResp.Body)
		return nil, fmt.Errorf("rollback failed with status %d: %s", rollbackResp.StatusCode, string(body))
	}

	return &common.DeploymentResult{
		DeploymentID: previousReleaseID,
		Status:       common.DeploymentStatusSuccess,
		Message:      fmt.Sprintf("Rolled back app %s to release %s", appName, previousReleaseID),
		Metadata: map[string]interface{}{
			"app_name":           appName,
			"rolled_back_to":     previousReleaseID,
			"rollback_initiated": time.Now().Format(time.RFC3339),
		},
	}, nil
}
