package runpod

import (
	"context"

	"github.com/functionfly/functionfly/internal/agent/generation"
)

// RunPodAdapter wraps the RunPod self-hosted generator to integrate with the generation service
// This adapter implements the CodeGenerator interface from the generation package
type RunPodAdapter struct {
	generator *SelfHostedGenerator
}

// NewRunPodAdapter creates a new RunPod adapter for the generation service
func NewRunPodAdapter(config *Config) (*RunPodAdapter, error) {
	generator, err := NewSelfHostedGenerator(config)
	if err != nil {
		return nil, err
	}

	return &RunPodAdapter{
		generator: generator,
	}, nil
}

// GenerateCode generates code using self-hosted GPU inference
// Implements the CodeGenerator interface from generation package
func (a *RunPodAdapter) GenerateCode(ctx context.Context, req *generation.GenerationRequest) (string, error) {
	return a.generator.GenerateCode(ctx, req)
}

// Terminate terminates all GPU instances
func (a *RunPodAdapter) Terminate(ctx context.Context) error {
	return a.generator.Terminate(ctx)
}

// GetStats returns pool statistics
func (a *RunPodAdapter) GetStats() (total, running, idle, failed int) {
	return a.generator.GetStats()
}

// AdapterGenerator is an interface that the generation.Service can use
// This allows the runpod adapter to be used wherever CodeGenerator is expected
type AdapterGenerator interface {
	GenerateCode(ctx context.Context, req *generation.GenerationRequest) (string, error)
	Terminate(ctx context.Context) error
	GetStats() (total, running, idle, failed int)
}

// Ensure RunPodAdapter implements generation.CodeGenerator
var _ generation.CodeGenerator = (*RunPodAdapter)(nil)

// Factory helpers for creating generators based on config

// CreateCodeGenerator creates the appropriate CodeGenerator based on environment
// This is the main entry point for the generation service
func CreateCodeGenerator() generation.CodeGenerator {
	config := LoadConfig()

	switch config.Mode {
	case InferenceModeSelfHosted:
		adapter, err := NewRunPodAdapter(config)
		if err != nil {
			// Log error and fall back to API
			return nil
		}
		return adapter

	case InferenceModeAPI:
		// Return nil - the caller should use OpenRouter client instead
		return nil

	default:
		return nil
	}
}

// CreateCodeGeneratorWithConfig creates a CodeGenerator with explicit config
func CreateCodeGeneratorWithConfig(config *Config) generation.CodeGenerator {
	if config == nil {
		return CreateCodeGenerator()
	}

	switch config.Mode {
	case InferenceModeSelfHosted:
		adapter, err := NewRunPodAdapter(config)
		if err != nil {
			return nil
		}
		return adapter

	case InferenceModeAPI:
		return nil

	default:
		return nil
	}
}
