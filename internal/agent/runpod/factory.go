package runpod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/generation"
	"gorm.io/gorm"
)

// CodeGeneratorFactory creates the appropriate CodeGenerator based on configuration
type CodeGeneratorFactory struct {
	config *Config
}

// NewCodeGeneratorFactory creates a new factory
func NewCodeGeneratorFactory() *CodeGeneratorFactory {
	return &CodeGeneratorFactory{
		config: LoadConfig(),
	}
}

// NewCodeGeneratorFactoryWithConfig creates a factory with custom config
func NewCodeGeneratorFactoryWithConfig(config *Config) *CodeGeneratorFactory {
	if config == nil {
		config = LoadConfig()
	}
	return &CodeGeneratorFactory{
		config: config,
	}
}

// CreateGenerator creates the appropriate CodeGenerator based on configuration
func (f *CodeGeneratorFactory) CreateGenerator() (generation.CodeGenerator, error) {
	switch f.config.Mode {
	case InferenceModeSelfHosted:
		return NewSelfHostedGenerator(f.config)
	case InferenceModeCluster:
		return NewClusterGenerator(f.config)
	case InferenceModeAPI:
		// Return nil to indicate use of OpenRouter (handled elsewhere)
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown inference mode: %s", f.config.Mode)
	}
}

// CreateOpenRouterGenerator creates an OpenRouter-based generator
// This is used when Mode is InferenceModeAPI
func (f *CodeGeneratorFactory) CreateOpenRouterGenerator(apiKey string) *generation.OpenRouterClient {
	return generation.NewOpenRouterClient(apiKey, nil, nil, nil)
}

// GetConfig returns the current configuration
func (f *CodeGeneratorFactory) GetConfig() *Config {
	return f.config
}

// SetConfig sets a new configuration
func (f *CodeGeneratorFactory) SetConfig(config *Config) {
	f.config = config
}

// CreateGenerationService creates a generation.Service with the appropriate CodeGenerator.
// db may be nil for API-only mode; when using self-hosted inference (gen != nil), db is required for persistence.
func (f *CodeGeneratorFactory) CreateGenerationService(db *gorm.DB) (*generation.Service, error) {
	gen, err := f.CreateGenerator()
	if err != nil {
		return nil, err
	}

	if gen != nil {
		if db == nil {
			return nil, fmt.Errorf("db is required when using self-hosted inference")
		}
		return generation.NewServiceWithGenerator(db, gen), nil
	}

	// API mode: service without generator (use CreateOpenRouterGenerator for OpenRouter client)
	return generation.NewService(db), nil
}

// HybridGenerator combines API-based and self-hosted inference with fallback
type HybridGenerator struct {
	selfHosted *SelfHostedGenerator
	apiClient  *generation.OpenRouterClient
}

// NewHybridGenerator creates a hybrid generator that tries self-hosted first, falls back to API
func NewHybridGenerator(selfHostedConfig *Config, apiKey string) (*HybridGenerator, error) {
	var selfHosted *SelfHostedGenerator
	var err error

	if selfHostedConfig != nil && selfHostedConfig.IsSelfHosted() {
		selfHosted, err = NewSelfHostedGenerator(selfHostedConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create self-hosted generator: %w", err)
		}
	}

	apiClient := generation.NewOpenRouterClient(apiKey, nil, nil, nil)

	return &HybridGenerator{
		selfHosted: selfHosted,
		apiClient:  apiClient,
	}, nil
}

// GenerateCode tries self-hosted first, falls back to API on failure
func (h *HybridGenerator) GenerateCode(ctx context.Context, req *generation.GenerationRequest) (string, error) {
	// Try self-hosted first if available
	if h.selfHosted != nil {
		code, err := h.selfHosted.GenerateCode(ctx, req)
		if err == nil {
			return code, nil
		}
		// Log fallback and continue to API
	}

	// Fall back to API
	if h.apiClient != nil {
		return h.apiClient.GenerateCode(ctx, req)
	}

	return "", fmt.Errorf("no generator available")
}

// Terminate terminates self-hosted resources
func (h *HybridGenerator) Terminate(ctx context.Context) error {
	if h.selfHosted != nil {
		return h.selfHosted.Terminate(ctx)
	}
	return nil
}

// GetStats returns statistics from both generators
func (h *HybridGenerator) GetStats() (selfHostedTotal, selfHostedRunning, selfHostedIdle, selfHostedFailed int) {
	if h.selfHosted != nil {
		return h.selfHosted.GetStats()
	}
	return 0, 0, 0, 0
}

// ClusterGenerator is a code generator that uses cluster mode with multiple regions
type ClusterGenerator struct {
	manager *ClusterManager
	config  *Config
}

// NewClusterGenerator creates a new cluster-based code generator
func NewClusterGenerator(config *Config) (*ClusterGenerator, error) {
	// Initialize regions from config if not set
	if len(config.Regions) == 0 {
		config.Regions = DefaultRegions()
	}

	manager := NewClusterManager(config)

	return &ClusterGenerator{
		manager: manager,
		config:  config,
	}, nil
}

// GenerateCode routes code generation request to an appropriate cluster
func (cg *ClusterGenerator) GenerateCode(ctx context.Context, req *generation.GenerationRequest) (string, error) {
	// Select the best cluster based on preferred region
	cluster, err := cg.manager.SelectCluster(cg.config.PreferredRegion)
	if err != nil {
		return "", fmt.Errorf("failed to select cluster: %w", err)
	}

	// Get an idle instance from the selected cluster
	instance, ok := cluster.Pool.GetIdleInstance("")
	if !ok {
		// Provision a new instance if none available
		instance, err = cluster.Pool.Provision(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to provision instance: %w", err)
		}
	}

	// Update request count
	instance.mu.Lock()
	instance.RequestCount++
	instance.LastUsed = time.Now()
	instance.mu.Unlock()

	// Generate code using the instance
	// This would call the actual inference endpoint
	code, err := cg.generateWithInstance(ctx, instance, req)
	if err != nil {
		// Release the instance back to the pool on failure
		cluster.Pool.Release(instance.ID)
		return "", err
	}

	// Release the instance back to the pool
	cluster.Pool.Release(instance.ID)

	return code, nil
}

// generateWithInstance generates code using a specific instance
func (cg *ClusterGenerator) generateWithInstance(ctx context.Context, instance *GPUInstance, req *generation.GenerationRequest) (string, error) {
	// Build the inference request
	payload := map[string]interface{}{
		"prompt":      req.Prompt,
		"runtime":     req.Runtime,
		"max_tokens":  2048,
		"temperature": 0.7,
		"top_p":       0.95,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the endpoint URL
	endpoint := instance.Endpoint
	if endpoint == "" {
		return "", fmt.Errorf("instance %s has no endpoint configured", instance.ID)
	}

	// Ensure the endpoint has a scheme
	if !bytes.HasPrefix([]byte(endpoint), []byte("http")) {
		endpoint = "http://" + endpoint
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Create the request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("inference request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("inference request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var result struct {
		GeneratedCode string `json:"generated_code"`
		Code          string `json:"code"`
		Output        string `json:"output"`
		Text          string `json:"text"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the generated code from various possible response formats
	code := result.GeneratedCode
	if code == "" {
		code = result.Code
	}
	if code == "" {
		code = result.Output
	}
	if code == "" {
		code = result.Text
	}

	if code == "" {
		return "", fmt.Errorf("no code generated in response")
	}

	return code, nil
}

// Terminate terminates all clusters and releases resources
func (cg *ClusterGenerator) Terminate(ctx context.Context) error {
	clusters := cg.manager.ListClusters()
	for _, cluster := range clusters {
		instances := cluster.Pool.ListInstances()
		for _, inst := range instances {
			if err := cluster.Pool.Terminate(ctx, inst.ID); err != nil {
				log.Printf("Warning: failed to terminate instance %s: %v", inst.ID, err)
			}
		}
	}
	return nil
}

// GetStats returns aggregated statistics from all clusters
func (cg *ClusterGenerator) GetStats() (total, running, idle, failed int) {
	stats := cg.manager.GetClusterStats()
	return stats.TotalInstances, stats.RunningInstances, stats.IdleInstances, stats.FailedInstances
}

// GetClusterManager returns the cluster manager for external use
func (cg *ClusterGenerator) GetClusterManager() *ClusterManager {
	return cg.manager
}
