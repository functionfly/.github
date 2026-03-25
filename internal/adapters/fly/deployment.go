package fly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

// Default Fly Machines API base URL (https://fly.io/docs/machines/api/working-with-machines-api/)
const defaultFlyAPIBase = "https://api.machines.dev"

// Machines API rate limit: 1 req/s per action (burst 3). We throttle to avoid 408.
const rateLimitMinInterval = 400 * time.Millisecond

func getFlyAPIBase() string {
	if v := os.Getenv("FLY_API_HOSTNAME"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return defaultFlyAPIBase
}

// FlyDeploymentClient handles deployment operations for Fly.io Machines API
type FlyDeploymentClient struct {
	httpClient *http.Client
	apiToken   string
	baseURL    string
	lastReq    time.Time
	reqMu      sync.Mutex
	storeMu    sync.RWMutex
	artifacts  map[string]map[string]*DeploymentArtifact
}

// NewFlyDeploymentClient creates a new Fly.io deployment client using FLY_API_HOSTNAME or default.
func NewFlyDeploymentClient(apiToken string) *FlyDeploymentClient {
	return NewFlyDeploymentClientWithBase(apiToken, getFlyAPIBase())
}

// NewFlyDeploymentClientWithBase creates a client with an explicit API base URL (for tests).
func NewFlyDeploymentClientWithBase(apiToken, baseURL string) *FlyDeploymentClient {
	return &FlyDeploymentClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		apiToken:   apiToken,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		lastReq:    time.Time{},
		artifacts:  make(map[string]map[string]*DeploymentArtifact),
	}
}

func (c *FlyDeploymentClient) rateLimit() {
	c.reqMu.Lock()
	elapsed := time.Since(c.lastReq)
	if elapsed < rateLimitMinInterval {
		time.Sleep(rateLimitMinInterval - elapsed)
	}
	c.lastReq = time.Now()
	c.reqMu.Unlock()
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

// EnsureApp creates a Fly.io app if it doesn't exist. org_slug is required when creating a new app (Machines API).
func (c *FlyDeploymentClient) EnsureApp(ctx context.Context, appName, orgSlug string) error {
	c.rateLimit()
	getURL := fmt.Sprintf("%s/v1/apps/%s", c.baseURL, appName)
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

	// Create the app (Machines API requires org_slug)
	if orgSlug == "" {
		return fmt.Errorf("org_slug is required when creating a new Fly app; set provider_config.org_slug (e.g. \"personal\")")
	}
	createData := map[string]interface{}{
		"app_name": appName,
		"org_slug": orgSlug,
	}
	jsonData, err := json.Marshal(createData)
	if err != nil {
		return fmt.Errorf("failed to marshal create app data: %w", err)
	}

	c.rateLimit()
	createReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/apps", c.baseURL), bytes.NewReader(jsonData))
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

// Deploy creates a Machine for the app using the given image reference. The image must already exist
// (e.g. built and pushed via flyctl or CI). imageRef defaults to registry.fly.io/<appName>:<version>.
func (c *FlyDeploymentClient) Deploy(ctx context.Context, appName, imageRef string) (*FlyDeployResult, error) {
	if imageRef == "" {
		return nil, fmt.Errorf("image ref is required for Fly deploy (image must be pre-pushed to registry.fly.io or set provider_config.image)")
	}

	c.rateLimit()
	machinesURL := fmt.Sprintf("%s/v1/apps/%s/machines", c.baseURL, appName)

	machineConfig := map[string]interface{}{
		"config": map[string]interface{}{
			"image": imageRef,
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
			"image":         imageRef,
			"deployed_at":   time.Now().Format(time.RFC3339),
		},
	}, nil
}

// SetEnvVars sets environment variables for a Fly.io app (legacy API; may require FLY_API_HOSTNAME=https://api.fly.io/v1)
func (c *FlyDeploymentClient) SetEnvVars(ctx context.Context, appName string, envVars map[string]string) error {
	c.rateLimit()
	envURL := fmt.Sprintf("%s/v1/apps/%s/env_vars", c.baseURL, appName)
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

// SetSecrets sets secrets for a Fly.io app (legacy API; may require FLY_API_HOSTNAME=https://api.fly.io/v1)
func (c *FlyDeploymentClient) SetSecrets(ctx context.Context, appName string, secrets map[string]string) error {
	c.rateLimit()
	secretsURL := fmt.Sprintf("%s/v1/apps/%s/secrets", c.baseURL, appName)
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

// UnsetSecret removes a secret from a Fly.io app
func (c *FlyDeploymentClient) UnsetSecret(ctx context.Context, appName, secretName string) error {
	c.rateLimit()
	secretsURL := fmt.Sprintf("%s/v1/apps/%s/secrets/%s", c.baseURL, appName, secretName)
	req, err := http.NewRequestWithContext(ctx, "DELETE", secretsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create unset secret request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unset secret: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unset secret failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListSecrets lists all secrets for a Fly.io app (returns only secret names, not values)
func (c *FlyDeploymentClient) ListSecrets(ctx context.Context, appName string) (map[string]string, error) {
	c.rateLimit()
	secretsURL := fmt.Sprintf("%s/v1/apps/%s/secrets", c.baseURL, appName)
	req, err := http.NewRequestWithContext(ctx, "GET", secretsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list secrets request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list secrets failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Secrets []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"secrets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode secrets response: %w", err)
	}

	secrets := make(map[string]string)
	for _, s := range result.Secrets {
		secrets[s.Name] = s.Type
	}
	return secrets, nil
}

// AddCertificate adds a custom domain certificate to a Fly.io app
func (c *FlyDeploymentClient) AddCertificate(ctx context.Context, appName, hostname string) error {
	c.rateLimit()
	certURL := fmt.Sprintf("%s/v1/apps/%s/certificates", c.baseURL, appName)
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
	c.rateLimit()
	machineURL := fmt.Sprintf("%s/v1/apps/%s/machines/%s", c.baseURL, appName, machineID)
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

// WaitForDeployment polls deployment status until it reaches a terminal state
func (c *FlyDeploymentClient) WaitForDeployment(ctx context.Context, appName, machineID string, timeout time.Duration) (common.DeploymentStatus, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second

	for time.Now().Before(deadline) {
		status, err := c.GetDeploymentStatus(ctx, appName, machineID)
		if err != nil {
			return common.DeploymentStatusFailed, err
		}

		// Check if we've reached a terminal state
		if status == common.DeploymentStatusSuccess || status == common.DeploymentStatusFailed {
			return status, nil
		}

		// Wait before next poll
		select {
		case <-ctx.Done():
			return common.DeploymentStatusFailed, ctx.Err()
		case <-time.After(pollInterval):
			continue
		}
	}

	return common.DeploymentStatusFailed, fmt.Errorf("deployment timed out after %v", timeout)
}

// flyMachine is a subset of the Machines API response for list/get
type flyMachine struct {
	ID     string                 `json:"id"`
	State  string                 `json:"state"`
	Region string                 `json:"region"`
	Config map[string]interface{} `json:"config"`
}

// ListMachines returns all machines for an app (Machines API)
func (c *FlyDeploymentClient) ListMachines(ctx context.Context, appName string) ([]flyMachine, error) {
	c.rateLimit()
	listURL := fmt.Sprintf("%s/v1/apps/%s/machines", c.baseURL, appName)
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list machines request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list machines failed with status %d: %s", resp.StatusCode, string(body))
	}
	var machines []flyMachine
	if err := json.NewDecoder(resp.Body).Decode(&machines); err != nil {
		return nil, fmt.Errorf("failed to decode list machines response: %w", err)
	}
	return machines, nil
}

// UpdateMachine updates a machine's config (full config required). Used for rollback by setting a new image.
func (c *FlyDeploymentClient) UpdateMachine(ctx context.Context, appName, machineID string, config map[string]interface{}) error {
	c.rateLimit()
	updateURL := fmt.Sprintf("%s/v1/apps/%s/machines/%s", c.baseURL, appName, machineID)
	body := map[string]interface{}{"config": config}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal update body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", updateURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update machine: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update machine failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// Rollback reverts all machines in the app to the given image version (Machines API: update each machine's image).
// version is the image tag to roll back to (e.g. "previous" or a semantic version). Image ref used: registry.fly.io/<appName>:<version>
func (c *FlyDeploymentClient) Rollback(ctx context.Context, appName, version string) (*common.DeploymentResult, error) {
	machines, err := c.ListMachines(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("list machines for rollback: %w", err)
	}
	if len(machines) == 0 {
		return nil, fmt.Errorf("no machines found in app %s", appName)
	}

	imageRef := fmt.Sprintf("registry.fly.io/%s:%s", appName, version)
	updated := 0
	for _, m := range machines {
		if m.Config == nil {
			continue
		}
		// Clone config and set new image
		newConfig := make(map[string]interface{})
		for k, v := range m.Config {
			newConfig[k] = v
		}
		newConfig["image"] = imageRef
		if err := c.UpdateMachine(ctx, appName, m.ID, newConfig); err != nil {
			return nil, fmt.Errorf("update machine %s: %w", m.ID, err)
		}
		updated++
	}

	return &common.DeploymentResult{
		DeploymentID: appName,
		Status:       common.DeploymentStatusSuccess,
		Message:      fmt.Sprintf("Rolled back app %s to image %s (%d machine(s) updated)", appName, imageRef, updated),
		Metadata: map[string]interface{}{
			"app_name":           appName,
			"image":              imageRef,
			"machines_updated":   updated,
			"rollback_initiated": time.Now().Format(time.RFC3339),
		},
	}, nil
}

// GetAppInfo retrieves app information from Fly.io
func (c *FlyDeploymentClient) GetAppInfo(ctx context.Context, appName string) (map[string]interface{}, error) {
	c.rateLimit()
	appURL := fmt.Sprintf("%s/v1/apps/%s", c.baseURL, appName)
	req, err := http.NewRequestWithContext(ctx, "GET", appURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get app request: %w", err)
	}
	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get app info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get app info failed with status %d: %s", resp.StatusCode, string(body))
	}
	var appInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&appInfo); err != nil {
		return nil, fmt.Errorf("failed to decode app info: %w", err)
	}
	return appInfo, nil
}

// ListAppRegions lists all regions where an app has machines
func (c *FlyDeploymentClient) ListAppRegions(ctx context.Context, appName string) ([]string, error) {
	machines, err := c.ListMachines(ctx, appName)
	if err != nil {
		return nil, err
	}

	regionSet := make(map[string]bool)
	for _, m := range machines {
		if m.Region != "" {
			regionSet[m.Region] = true
		}
	}

	regions := make([]string, 0, len(regionSet))
	for region := range regionSet {
		regions = append(regions, region)
	}
	return regions, nil
}

// ScaleApp scales an app to the specified number of machines in a region
func (c *FlyDeploymentClient) ScaleApp(ctx context.Context, appName, region string, count int) error {
	machines, err := c.ListMachines(ctx, appName)
	if err != nil {
		return err
	}

	// Count machines in the target region
	regionCount := 0
	for _, m := range machines {
		if m.Region == region {
			regionCount++
		}
	}

	// If we need more machines, create them
	if regionCount < count {
		// Get the first machine's config as a template
		if len(machines) == 0 {
			return fmt.Errorf("no machines found to use as template")
		}

		templateConfig := machines[0].Config
		for i := regionCount; i < count; i++ {
			c.rateLimit()
			machinesURL := fmt.Sprintf("%s/v1/apps/%s/machines", c.baseURL, appName)
			machineConfig := map[string]interface{}{
				"config": templateConfig,
				"region": region,
			}
			jsonData, err := json.Marshal(machineConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal machine config: %w", err)
			}
			req, err := http.NewRequestWithContext(ctx, "POST", machinesURL, bytes.NewReader(jsonData))
			if err != nil {
				return fmt.Errorf("failed to create scale request: %w", err)
			}
			c.setAuthHeaders(req)
			resp, err := c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to scale app: %w", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("scale app failed with status %d: %s", resp.StatusCode, string(body))
			}
		}
	}

	return nil
}

// DeploymentArtifact represents a stored deployment artifact
type DeploymentArtifact struct {
	AppName     string    `json:"app_name"`
	ImageRef    string    `json:"image_ref"`
	Version     string    `json:"version"`
	DeployedAt  time.Time `json:"deployed_at"`
	DeployedBy  string    `json:"deployed_by"`
	MachineID   string    `json:"machine_id"`
	Region      string    `json:"region"`
	Status      string    `json:"status"`
}

// StoreDeploymentArtifact stores a deployment artifact for rollback history
func (c *FlyDeploymentClient) StoreDeploymentArtifact(ctx context.Context, artifact *DeploymentArtifact) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if artifact == nil {
		return fmt.Errorf("artifact is required")
	}
	if strings.TrimSpace(artifact.AppName) == "" {
		return fmt.Errorf("artifact app name is required")
	}
	if strings.TrimSpace(artifact.Version) == "" {
		return fmt.Errorf("artifact version is required")
	}
	if artifact.DeployedAt.IsZero() {
		artifact.DeployedAt = time.Now().UTC()
	}
	copied := *artifact
	c.storeMu.Lock()
	defer c.storeMu.Unlock()
	if c.artifacts[copied.AppName] == nil {
		c.artifacts[copied.AppName] = make(map[string]*DeploymentArtifact)
	}
	c.artifacts[copied.AppName][copied.Version] = &copied
	return nil
}

// ListDeploymentArtifacts lists deployment artifacts for rollback history
func (c *FlyDeploymentClient) ListDeploymentArtifacts(ctx context.Context, appName string, limit int) ([]*DeploymentArtifact, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if strings.TrimSpace(appName) == "" {
		return nil, fmt.Errorf("app name is required")
	}
	c.storeMu.RLock()
	defer c.storeMu.RUnlock()
	byVersion := c.artifacts[appName]
	if len(byVersion) == 0 {
		return []*DeploymentArtifact{}, nil
	}
	artifacts := make([]*DeploymentArtifact, 0, len(byVersion))
	for _, a := range byVersion {
		copied := *a
		artifacts = append(artifacts, &copied)
	}
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].DeployedAt.After(artifacts[j].DeployedAt)
	})
	if limit > 0 && len(artifacts) > limit {
		artifacts = artifacts[:limit]
	}
	return artifacts, nil
}

// GetDeploymentArtifact gets a specific deployment artifact
func (c *FlyDeploymentClient) GetDeploymentArtifact(ctx context.Context, appName, version string) (*DeploymentArtifact, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if strings.TrimSpace(appName) == "" {
		return nil, fmt.Errorf("app name is required")
	}
	if strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("version is required")
	}
	c.storeMu.RLock()
	defer c.storeMu.RUnlock()
	artifact, ok := c.artifacts[appName][version]
	if !ok {
		return nil, fmt.Errorf("deployment artifact not found")
	}
	copied := *artifact
	return &copied, nil
}

// DeleteDeploymentArtifact deletes a deployment artifact
func (c *FlyDeploymentClient) DeleteDeploymentArtifact(ctx context.Context, appName, version string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if strings.TrimSpace(appName) == "" {
		return fmt.Errorf("app name is required")
	}
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("version is required")
	}
	c.storeMu.Lock()
	defer c.storeMu.Unlock()
	versions := c.artifacts[appName]
	if len(versions) == 0 {
		return nil
	}
	delete(versions, version)
	if len(versions) == 0 {
		delete(c.artifacts, appName)
	}
	return nil
}

// GetRollbackHistory gets rollback history for an app
func (c *FlyDeploymentClient) GetRollbackHistory(ctx context.Context, appName string, limit int) ([]*DeploymentArtifact, error) {
	artifacts, err := c.ListDeploymentArtifacts(ctx, appName, 0)
	if err != nil {
		return nil, err
	}
	rollbackArtifacts := make([]*DeploymentArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Status == "rollback" || artifact.Status == "rollback_failed" {
			rollbackArtifacts = append(rollbackArtifacts, artifact)
		}
	}
	if limit > 0 && len(rollbackArtifacts) > limit {
		rollbackArtifacts = rollbackArtifacts[:limit]
	}
	return rollbackArtifacts, nil
}

// RecordRollback records a rollback event
func (c *FlyDeploymentClient) RecordRollback(ctx context.Context, appName, fromVersion, toVersion string, success bool) error {
	status := "rollback"
	if !success {
		status = "rollback_failed"
	}
	return c.StoreDeploymentArtifact(ctx, &DeploymentArtifact{
		AppName:    appName,
		ImageRef:   fmt.Sprintf("registry.fly.io/%s:%s", appName, toVersion),
		Version:    fmt.Sprintf("rollback:%s->%s@%d", fromVersion, toVersion, time.Now().UTC().UnixNano()),
		DeployedAt: time.Now().UTC(),
		Status:     status,
	})
}
