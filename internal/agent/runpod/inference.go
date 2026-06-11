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

	"github.com/functionfly/functionfly/internal/agent/generation"
)

const (
	defaultInferenceTimeout = 45 * time.Second
	maxRetries              = 3
	retryDelay              = 2 * time.Second
)

// SelfHostedClient implements the CodeGenerator interface for self-hosted GPU inference
type SelfHostedClient struct {
	pool       *InstancePool
	config     *Config
	httpClient *http.Client
}

// NewSelfHostedClient creates a new self-hosted inference client
func NewSelfHostedClient(config *Config) (*SelfHostedClient, error) {
	if config == nil {
		config = LoadConfig()
	}

	if !config.IsSelfHosted() {
		return nil, fmt.Errorf("config must be set to self-hosted mode")
	}

	client := &SelfHostedClient{
		config: config,
		httpClient: &http.Client{
			Timeout: defaultInferenceTimeout,
		},
	}

	// Initialize the instance pool
	runpodClient := NewRunPodClient(config.RunPodAPIKey, config.RunPodAPIBaseURL)
	pool := NewInstancePool(config, runpodClient)
	client.pool = pool

	// Start idle monitoring if configured
	if config.IdleTimeout > 0 {
		ctx := context.Background()
		pool.StartIdleMonitor(ctx, 30*time.Second)
	}

	return client, nil
}

// GenerateCode generates code using self-hosted GPU inference
// Implements the CodeGenerator interface from generation package
func (c *SelfHostedClient) GenerateCode(ctx context.Context, req *generation.GenerationRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("generation request is required")
	}

	// Build the prompt
	selector := generation.HeuristicModelSelector{}
	selection := selector.SelectModel(req)

	if req.Model == "" {
		req.Model = selection.Model
	}
	if req.Prompt == "" {
		req.Prompt = generation.BuildPrompt(req, selection)
	}

	// Provision a GPU instance
	instance, err := c.pool.Provision(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to provision GPU instance: %w", err)
	}

	// Wait for instance to be ready
	waitCtx, cancel := context.WithTimeout(ctx, c.config.ProvisioningTimeout)
	defer cancel()

	for {
		inst, ok := c.pool.GetInstance(instance.ID)
		if !ok {
			c.pool.Release(instance.ID)
			return "", fmt.Errorf("instance not found")
		}

		inst.mu.RLock()
		state := inst.State
		endpoint := inst.Endpoint
		inst.mu.RUnlock()

		if state == InstanceStateRunning && endpoint != "" {
			break
		}

		if state == InstanceStateFailed {
			c.pool.Release(instance.ID)
			return "", fmt.Errorf("instance failed to start")
		}

		select {
		case <-waitCtx.Done():
			c.pool.Release(instance.ID)
			return "", fmt.Errorf("timeout waiting for instance to be ready")
		case <-time.After(2 * time.Second):
			// Continue waiting
		}
	}

	// Make inference request with retries
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		code, err := c.makeInferenceRequest(ctx, instance, req)
		if err == nil {
			// Release instance back to pool
			c.pool.Release(instance.ID)
			return code, nil
		}

		lastErr = err
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				c.pool.Release(instance.ID)
				return "", ctx.Err()
			case <-time.After(retryDelay):
				// Retry
			}
		}
	}

	// Release instance on failure
	c.pool.Release(instance.ID)
	return "", fmt.Errorf("inference failed after %d retries: %w", maxRetries, lastErr)
}

// makeInferenceRequest makes a request to the self-hosted inference endpoint
func (c *SelfHostedClient) makeInferenceRequest(ctx context.Context, instance *GPUInstance, req *generation.GenerationRequest) (string, error) {
	instance.mu.RLock()
	endpoint := instance.Endpoint
	instance.mu.RUnlock()

	if endpoint == "" {
		return "", fmt.Errorf("instance endpoint not available")
	}

	// Format the request for the LLM
	prompt := c.formatPrompt(req)

	payload := inferenceRequest{
		Inputs: prompt,
		Parameters: inferenceParameters{
			Temperature:  0.4,
			MaxNewTokens: 2048,
			DoSample:     true,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("inference error: %s", strings.TrimSpace(string(raw)))
	}

	var result inferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.GeneratedText == "" {
		return "", fmt.Errorf("no generated text returned")
	}

	// Strip code fences if present
	code := strings.TrimSpace(stripCodeFences(result.GeneratedText))
	return code, nil
}

// formatPrompt formats the request into a prompt for the LLM
func (c *SelfHostedClient) formatPrompt(req *generation.GenerationRequest) string {
	inputSchema, _ := json.Marshal(req.InputSchema)
	outputSchema, _ := json.Marshal(req.OutputSchema)

	return fmt.Sprintf(`You are a serverless function generator for FunctionFly platform.
Generate production-ready code for the following function:

Function Name: %s
Description: %s
Category: %s
Runtime: %s
Input Schema: %s
Output Schema: %s

Requirements:
1. Production-ready code only
2. Include validation and error handling
3. Avoid unsafe dynamic execution
4. Keep dependencies minimal
5. Optimize for cold start and latency

Prompt: %s

Return ONLY the source code, no explanations or markdown formatting.`,
		req.Name,
		req.Description,
		req.Category,
		req.Runtime,
		string(inputSchema),
		string(outputSchema),
		req.Prompt,
	)
}

// stripCodeFences removes code fences from generated text
func stripCodeFences(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```python")
	content = strings.TrimPrefix(content, "```javascript")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

// inferenceRequest represents a request to the inference endpoint
type inferenceRequest struct {
	Inputs     string              `json:"inputs"`
	Parameters inferenceParameters `json:"parameters"`
}

// inferenceParameters represents parameters for the inference request
type inferenceParameters struct {
	Temperature  float64 `json:"temperature"`
	MaxNewTokens int     `json:"max_new_tokens"`
	DoSample     bool    `json:"do_sample"`
}

// inferenceResponse represents the response from the inference endpoint
type inferenceResponse struct {
	GeneratedText string `json:"generated_text"`
}

// Terminate terminates all GPU instances
func (c *SelfHostedClient) Terminate(ctx context.Context) error {
	instances := c.pool.ListInstances()
	for _, inst := range instances {
		if err := c.pool.Terminate(ctx, inst.ID); err != nil {
			return err
		}
	}
	return nil
}

// GetStats returns pool statistics
func (c *SelfHostedClient) GetStats() (total, running, idle, failed int) {
	return c.pool.GetStats()
}

// SelfHostedGenerator is a wrapper that makes SelfHostedClient compatible with CodeGenerator
type SelfHostedGenerator struct {
	client *SelfHostedClient
}

// NewSelfHostedGenerator creates a new self-hosted generator
func NewSelfHostedGenerator(config *Config) (*SelfHostedGenerator, error) {
	client, err := NewSelfHostedClient(config)
	if err != nil {
		return nil, err
	}
	return &SelfHostedGenerator{client: client}, nil
}

// GenerateCode generates code using self-hosted GPU inference
// Implements the CodeGenerator interface
func (g *SelfHostedGenerator) GenerateCode(ctx context.Context, req *generation.GenerationRequest) (string, error) {
	return g.client.GenerateCode(ctx, req)
}

// Terminate terminates all GPU instances
func (g *SelfHostedGenerator) Terminate(ctx context.Context) error {
	return g.client.Terminate(ctx)
}

// GetStats returns pool statistics
func (g *SelfHostedGenerator) GetStats() (total, running, idle, failed int) {
	return g.client.GetStats()
}
