package fly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultFlyMachinesBaseURL = "https://api.machines.dev"
	rateLimitInterval         = 400 * time.Millisecond
)

func getFlyMachinesBaseURL() string {
	if v := os.Getenv("FLY_API_HOSTNAME"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return defaultFlyMachinesBaseURL
}

func getEnterpriseAppName() string {
	if v := os.Getenv("FLY_ENTERPRISE_APP_NAME"); v != "" {
		return v
	}
	return "functionfly-enterprise"
}

type FlyMachinesClient struct {
	apiToken   string
	baseURL    string
	appName    string
	httpClient *http.Client
	lastReq    time.Time
	reqMu      sync.Mutex
}

func NewFlyMachinesClient(apiToken string) *FlyMachinesClient {
	return NewFlyMachinesClientWithBase(apiToken, getFlyMachinesBaseURL(), getEnterpriseAppName())
}

func NewFlyMachinesClientWithBase(apiToken, baseURL, appName string) *FlyMachinesClient {
	return &FlyMachinesClient{
		apiToken:   apiToken,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		appName:    appName,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		lastReq:    time.Time{},
	}
}

func (c *FlyMachinesClient) rateLimit() {
	c.reqMu.Lock()
	elapsed := time.Since(c.lastReq)
	if elapsed < rateLimitInterval {
		time.Sleep(rateLimitInterval - elapsed)
	}
	c.lastReq = time.Now()
	c.reqMu.Unlock()
}

func (c *FlyMachinesClient) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
}

type GuestConfig struct {
	CPUKind  string `json:"cpu_kind,omitempty"`
	CPUs     int    `json:"cpus"`
	MemoryMB int    `json:"memory_mb"`
}

type MachineConfig struct {
	Image    string            `json:"image"`
	Guest    GuestConfig      `json:"guest"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
}

type CreateMachineRequest struct {
	Config MachineConfig `json:"config"`
	Region string        `json:"region,omitempty"`
}

type Machine struct {
	ID     string         `json:"id"`
	State  string         `json:"state"`
	Region string         `json:"region"`
	Config MachineConfig  `json:"config"`
}

type MachineEvent struct {
	ID        string `json:"id"`
	MachineID string `json:"machine_id"`
	State     string `json:"state"`
	OldState  string `json:"old_state"`
}

func (c *FlyMachinesClient) CreateMachine(ctx context.Context, req *CreateMachineRequest) (*Machine, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/v1/apps/%s/machines", c.baseURL, c.appName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal machine config: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create machine failed with status %d: %s", resp.StatusCode, string(body))
	}

	var machine Machine
	if err := json.Unmarshal(body, &machine); err != nil {
		return nil, fmt.Errorf("failed to decode machine response: %w", err)
	}

	return &machine, nil
}

func (c *FlyMachinesClient) GetMachine(ctx context.Context, machineID string) (*Machine, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s", c.baseURL, c.appName, machineID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get machine failed with status %d: %s", resp.StatusCode, string(body))
	}

	var machine Machine
	if err := json.Unmarshal(body, &machine); err != nil {
		return nil, fmt.Errorf("failed to decode machine response: %w", err)
	}

	return &machine, nil
}

func (c *FlyMachinesClient) StopMachine(ctx context.Context, machineID string) error {
	c.rateLimit()

	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s/stop", c.baseURL, c.appName, machineID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to stop machine: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stop machine failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *FlyMachinesClient) DeleteMachine(ctx context.Context, machineID string) error {
	c.rateLimit()

	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s", c.baseURL, c.appName, machineID)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete machine failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *FlyMachinesClient) WaitForMachine(ctx context.Context, machineID, desiredState string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		machine, err := c.GetMachine(ctx, machineID)
		if err != nil {
			return fmt.Errorf("failed to get machine state: %w", err)
		}
		if machine == nil {
			return fmt.Errorf("machine %s not found", machineID)
		}

		if machine.State == desiredState {
			return nil
		}

		if machine.State == "destroyed" || machine.State == "failed" {
			return fmt.Errorf("machine entered terminal state: %s", machine.State)
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout waiting for machine %s to reach state %s", machineID, desiredState)
}

func (c *FlyMachinesClient) ListMachines(ctx context.Context) ([]Machine, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/v1/apps/%s/machines", c.baseURL, c.appName)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list machines failed with status %d: %s", resp.StatusCode, string(body))
	}

	var machines []Machine
	if err := json.Unmarshal(body, &machines); err != nil {
		return nil, fmt.Errorf("failed to decode machines response: %w", err)
	}

	return machines, nil
}

func (c *FlyMachinesClient) ExecInMachine(ctx context.Context, machineID, path string, body io.Reader) ([]byte, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s/exec/%s", c.baseURL, c.appName, machineID, path)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec request: %w", err)
	}
	c.setAuthHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to exec in machine: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exec in machine failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *FlyMachinesClient) GetAppName() string {
	return c.appName
}
