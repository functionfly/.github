package runpod

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/agent/generation"
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
	case InferenceModeAPI:
		// Return nil to indicate use of OpenRouter (handled elsewhere)
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown inference mode: %s", f.config.Mode)
	}
}

// MustCreateGenerator creates the appropriate CodeGenerator and panics on error
func (f *CodeGeneratorFactory) MustCreateGenerator() generation.CodeGenerator {
	gen, err := f.CreateGenerator()
	if err != nil {
		panic(fmt.Sprintf("failed to create generator: %v", err))
	}
	return gen
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

// CreateGenerationService creates a generation.Service with the appropriate CodeGenerator
func (f *CodeGeneratorFactory) CreateGenerationService(db interface{}) (*generation.Service, error) {
	gen, err := f.CreateGenerator()
	if err != nil {
		return nil, err
	}

	// Note: db parameter would be *gorm.DB in real usage
	// This is a simplified signature that matches the generation.Service
	if gen != nil {
		// We need to cast or handle the db properly
		// For now, return a service with the generator
		// The caller would need to pass proper db
		return nil, fmt.Errorf("db must be *gorm.DB")
	}

	// Return service without generator (will use OpenRouter via CreateOpenRouterGenerator)
	return nil, nil
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
