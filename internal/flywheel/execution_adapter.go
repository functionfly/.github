package flywheel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/sirupsen/logrus"
)

// RegistryFunctionExecutor runs a registry function version with the given input.
// Implementations typically use the sandbox/WASM execution path (e.g. execution.LocalExecutor).
type RegistryFunctionExecutor interface {
	Execute(fnVersion *registry.RegistryFunctionVersion, input []byte, timeoutMs int) ([]byte, error)
}

// ExecutionAdapter integrates Flywheel with the existing execution infrastructure
type ExecutionAdapter struct {
	registryRepo *registry.RegistryRepository
	cacheService *cache.CacheService
	executor     RegistryFunctionExecutor
	logger       *logrus.Logger
}

// NewExecutionAdapter creates a new execution adapter. executor must be non-nil; use
// execution.NewLocalExecutor() (e.g. as in internal/api/routes.go) so ExecuteFunction runs functions.
func NewExecutionAdapter(registryRepo *registry.RegistryRepository, cacheService *cache.CacheService, executor RegistryFunctionExecutor, logger *logrus.Logger) *ExecutionAdapter {
	if executor == nil {
		panic("flywheel: executor is required; use execution.NewLocalExecutor() for production")
	}
	return &ExecutionAdapter{
		registryRepo: registryRepo,
		cacheService: cacheService,
		executor:     executor,
		logger:       logger,
	}
}

// ExecuteCode executes code using the existing execution infrastructure.
// This is currently limited to deterministic validation only - actual WASM sandbox
// execution for arbitrary code is not yet implemented for production.
func (a *ExecutionAdapter) ExecuteCode(ctx context.Context, code string, language string, input json.RawMessage) (*ExecutionResult, error) {
	startTime := time.Now()

	// Validate code is not empty and language is supported
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if language != "javascript" && language != "typescript" && language != "python" {
		return nil, fmt.Errorf("unsupported language: %s (supported: javascript, typescript, python)", language)
	}

	// Check for non-deterministic patterns
	isDeterministic := a.checkDeterminism(code)
	if !isDeterministic {
		a.logger.WithField("language", language).Warn("Code contains non-deterministic patterns")
	}

	// Estimate execution metrics (production: replace with actual sandbox execution)
	executionTime := a.estimateExecutionTime(code)
	computeCost := a.calculateComputeCost(code, executionTime)

	// For production readiness: return not implemented error
	// TODO: Implement actual WASM sandbox execution for arbitrary code
	a.logger.WithFields(logrus.Fields{
		"language":    language,
		"code_length": len(code),
		"has_input":   len(input) > 0,
	}).Warn("Code execution not fully implemented - returning simulated result for validation only")

	// Create validation output (not actual execution)
	var output json.RawMessage
	if len(input) > 0 {
		output = json.RawMessage(`{"result": "validated", "status": "simulated", "input_preview": "` + fmt.Sprintf("%d bytes", len(input)) + `"}`)
	} else {
		output = json.RawMessage(`{"result": "validated", "status": "simulated"}`)
	}

	result := &ExecutionResult{
		Output:          output,
		RuntimeMS:       int(executionTime.Milliseconds()),
		MemoryMB:        a.estimateMemoryUsage(code),
		ComputeCost:     computeCost,
		IsDeterministic: isDeterministic,
	}

	a.logger.WithFields(logrus.Fields{
		"runtime_ms":    result.RuntimeMS,
		"memory_mb":     result.MemoryMB,
		"compute_cost":  result.ComputeCost,
		"deterministic": result.IsDeterministic,
		"duration":      time.Since(startTime),
	}).Debug("Code validation completed (simulated execution)")

	return result, nil
}

// VerifyOutput verifies if actual output matches expected output
func (a *ExecutionAdapter) VerifyOutput(ctx context.Context, actual, expected json.RawMessage) (bool, string, error) {
	var actualVal, expectedVal interface{}

	if err := json.Unmarshal(actual, &actualVal); err != nil {
		return false, "Failed to parse actual output", fmt.Errorf("failed to parse actual output: %w", err)
	}

	if err := json.Unmarshal(expected, &expectedVal); err != nil {
		return false, "Failed to parse expected output", fmt.Errorf("failed to parse expected output: %w", err)
	}

	match := deepEqual(actualVal, expectedVal)

	if match {
		return true, "", nil
	}

	return false, "Output does not match expected value", nil
}

// ExecuteFunction executes a registered function via the configured executor (sandbox/WASM or backend).
func (a *ExecutionAdapter) ExecuteFunction(ctx context.Context, author, name, version string, input json.RawMessage) (*functionregistry.ExecutionResponse, error) {
	fn, err := a.registryRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		return nil, fmt.Errorf("function not found: %w", err)
	}

	var fnVersion *registry.RegistryFunctionVersion
	if version != "" {
		fnVersion, err = a.registryRepo.GetFunctionVersion(fn.ID, version)
	} else {
		fnVersion, err = a.registryRepo.GetLatestFunctionVersion(fn.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("function version not found: %w", err)
	}

	if a.executor == nil {
		return nil, fmt.Errorf("executor not configured: pass a RegistryFunctionExecutor to NewExecutionAdapter for production")
	}

	timeoutMs := fnVersion.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	output, err := a.executor.Execute(fnVersion, []byte(input), timeoutMs)
	if err != nil {
		a.logger.WithError(err).WithFields(logrus.Fields{
			"author":  author,
			"name":    name,
			"version": fnVersion.Version,
		}).Warn("Function execution failed")
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return &functionregistry.ExecutionResponse{
		OK:   true,
		Data: json.RawMessage(output),
	}, nil
}

// Helper methods

func (a *ExecutionAdapter) estimateExecutionTime(code string) time.Duration {
	baseTime := 50 * time.Millisecond
	complexity := len(code) / 100
	if complexity > 1000 {
		complexity = 1000
	}
	return baseTime + time.Duration(complexity)*time.Millisecond
}

func (a *ExecutionAdapter) calculateComputeCost(code string, executionTime time.Duration) float64 {
	baseCost := 0.001
	timeCost := float64(executionTime.Milliseconds()) * 0.00001
	memoryCost := float64(a.estimateMemoryUsage(code)) * 0.000001
	return baseCost + timeCost + memoryCost
}

func (a *ExecutionAdapter) estimateMemoryUsage(code string) int {
	baseMemory := 64
	codeMemory := len(code) / 10000
	if codeMemory > 512 {
		codeMemory = 512
	}
	return baseMemory + codeMemory
}

func (a *ExecutionAdapter) checkDeterminism(code string) bool {
	nonDeterministicPatterns := []string{
		"Math.random()",
		"Date.now()",
		"new Date()",
		"setTimeout",
		"setInterval",
		"fetch",
		"XMLHttpRequest",
	}

	for _, pattern := range nonDeterministicPatterns {
		if contains(code, pattern) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// deepEqual performs a deep comparison of two values
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for key, aVal := range av {
			bVal, ok := bv[key]
			if !ok || !deepEqual(aVal, bVal) {
				return false
			}
		}
		return true

	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true

	case float64:
		bv, ok := b.(float64)
		if !ok {
			return false
		}
		diff := av - bv
		if diff < 0 {
			diff = -diff
		}
		return diff < 0.0001

	case string:
		bv, ok := b.(string)
		return ok && av == bv

	case bool:
		bv, ok := b.(bool)
		return ok && av == bv

	default:
		return a == b
	}
}
