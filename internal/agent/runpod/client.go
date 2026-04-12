package runpod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// PodStatus represents the status of a RunPod pod
type PodStatus string

const (
	PodStatusUnknown    PodStatus = "UNKNOWN"
	PodStatusStarting   PodStatus = "STARTING"
	PodStatusRunning    PodStatus = "RUNNING"
	PodStatusReady      PodStatus = "READY"
	PodStatusFailed     PodStatus = "FAILED"
	PodStatusTerminated PodStatus = "TERMINATED"
	PodStatusRemoved    PodStatus = "REMOVED"
)

// GPUPod represents a GPU pod in RunPod
type GPUPod struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Status         PodStatus `json:"runtime.status"`
	GPUCount       int       `json:"runtime.gpuCount"`
	GPUType        string    `json:"runtime.gpuTypeId"`
	ContainerURL   string    `json:"runtime.container.httpUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	RuntimeMinutes float64   `json:"runtime.runtimeInSeconds"` // in seconds
}

// PodSpec defines the specification for a GPU pod
type PodSpec struct {
	Name            string
	ContainerImage  string
	GPUType         string
	GPUCount        int
	EnvVars         map[string]string
	HTTPHost        string
	HTTPPort        int
	HealthCheckPath string
	ModelName       string
}

// RunPodClient provides methods to interact with the RunPod API
type RunPodClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewRunPodClient creates a new RunPod API client
func NewRunPodClient(apiKey string, baseURL string) *RunPodClient {
	if baseURL == "" {
		baseURL = "https://api.runpod.io/graphql"
	}
	return &RunPodClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// PodResponse represents the response from pod operations
type PodResponse struct {
	Data struct {
		Pod struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Status       string `json:"status"`
			ContainerURL string `json:"containerHttpUrl"`
		} `json:"pod"`
	} `json:"data"`
	Errors []GraphQLError `json:"errors"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// PodStatusResponse represents the response for pod status queries
type PodStatusResponse struct {
	Data struct {
		Pod struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Runtime struct {
				ContainerHTTPURL string `json:"containerHttpUrl"`
				GPUCount         int    `json:"gpuCount"`
				GPUTypeID        string `json:"gpuTypeId"`
				UptimeInSeconds  int    `json:"uptimeInSeconds"`
			} `json:"runtime"`
		} `json:"pod"`
	} `json:"data"`
	Errors []GraphQLError `json:"errors"`
}

// CreatePod creates a new GPU pod
func (c *RunPodClient) CreatePod(ctx context.Context, spec *PodSpec) (*GPUPod, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("RunPod API key is required")
	}

	query := `
		mutation($input: PodCreateInput!) {
			pod(input: $input) {
				id
				name
				status
				containerHttpUrl
			}
		}
	`

	envVars := []map[string]string{
		{"key": "MODEL_NAME", "value": spec.ModelName},
		{"key": "HF_HOME", "value": "/model"},
		{"key": "TRANSFORMERS_CACHE", "value": "/model"},
	}
	for k, v := range spec.EnvVars {
		envVars = append(envVars, map[string]string{"key": k, "value": v})
	}

	input := map[string]interface{}{
		"name":            spec.Name,
		"image":           spec.ContainerImage,
		"env":             envVars,
		"gpuTypeId":       spec.GPUType,
		"gpuCount":        spec.GPUCount,
		"cloudType":       "SECURE",
		"networkVolumeId": nil,
		"portConfigs": []map[string]interface{}{
			{"containerPort": spec.HTTPPort, "hostPort": spec.HTTPPort, "protocol": "http"},
		},
	}

	variables := map[string]interface{}{
		"input": input,
	}

	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RunPod API error: %s", strings.TrimSpace(string(raw)))
	}

	var result PodResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", result.Errors)
	}

	pod := &GPUPod{
		ID:           result.Data.Pod.ID,
		Name:         result.Data.Pod.Name,
		Status:       PodStatus(result.Data.Pod.Status),
		ContainerURL: result.Data.Pod.ContainerURL,
		GPUType:      spec.GPUType,
		GPUCount:     spec.GPUCount,
		CreatedAt:    time.Now(),
	}

	return pod, nil
}

// GetPodStatus retrieves the status of a pod
func (c *RunPodClient) GetPodStatus(ctx context.Context, podID string) (*GPUPod, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("RunPod API key is required")
	}

	query := `
		query($podId: String!) {
			pod(input: {podId: $podId}) {
				id
				status
				runtime {
					containerHttpUrl
					gpuCount
					gpuTypeId
					uptimeInSeconds
				}
			}
		}
	`

	variables := map[string]interface{}{
		"podId": podID,
	}

	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RunPod API error: %s", strings.TrimSpace(string(raw)))
	}

	var result PodStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", result.Errors)
	}

	pod := &GPUPod{
		ID:             result.Data.Pod.ID,
		Status:         PodStatus(result.Data.Pod.Status),
		ContainerURL:   result.Data.Pod.Runtime.ContainerHTTPURL,
		GPUCount:       result.Data.Pod.Runtime.GPUCount,
		GPUType:        result.Data.Pod.Runtime.GPUTypeID,
		RuntimeMinutes: float64(result.Data.Pod.Runtime.UptimeInSeconds) / 60.0,
	}

	return pod, nil
}

// TerminatePod terminates a pod
func (c *RunPodClient) TerminatePod(ctx context.Context, podID string) error {
	if c.apiKey == "" {
		return fmt.Errorf("RunPod API key is required")
	}

	query := `
		mutation($input: TerminatePodInput!) {
			terminatePod(input: $input) {
				success
			}
		}
	`

	input := map[string]interface{}{
		"podId": podID,
	}

	body, err := json.Marshal(map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"input": input,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RunPod API error: %s", strings.TrimSpace(string(raw)))
	}

	// Parse response to check for errors
	var result struct {
		Data struct {
			TerminatePod struct {
				Success bool `json:"success"`
			} `json:"terminatePod"`
		} `json:"data"`
		Errors []GraphQLError `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", result.Errors)
	}

	if !result.Data.TerminatePod.Success {
		return fmt.Errorf("failed to terminate pod")
	}

	return nil
}

// WaitForPodReady waits for a pod to become ready
func (c *RunPodClient) WaitForPodReady(ctx context.Context, podID string, timeout time.Duration) (*GPUPod, error) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for pod to be ready: %w", timeoutCtx.Err())
		case <-ticker.C:
			pod, err := c.GetPodStatus(ctx, podID)
			if err != nil {
				return nil, err
			}

			if pod.Status == PodStatusRunning || pod.Status == PodStatusReady {
				return pod, nil
			}

			if pod.Status == PodStatusFailed || pod.Status == PodStatusTerminated {
				return nil, fmt.Errorf("pod failed with status: %s", pod.Status)
			}
		}
	}
}

// PingHealthEndpoint pings a pod's health endpoint and returns latency
func (c *RunPodClient) PingHealthEndpoint(ctx context.Context, podURL string, healthPath string) (latencyMs float64, healthy bool, err error) {
	if podURL == "" {
		return 0, false, fmt.Errorf("pod URL is empty")
	}

	url := fmt.Sprintf("%s%s", podURL, healthPath)
	if healthPath == "" {
		url = fmt.Sprintf("%s/health", podURL)
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	latencyMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		return latencyMs, false, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return latencyMs, true, nil
	}

	return latencyMs, false, fmt.Errorf("health check returned status %d", resp.StatusCode)
}
