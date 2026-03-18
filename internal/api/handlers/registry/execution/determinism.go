package execution

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RegistryDeterminismExecutor implements generation.DeterminismExecutor by resolving
// the function via the registry and executing it with the sandbox (WASM only).
type RegistryDeterminismExecutor struct {
	registryRepo *registry.RegistryRepository
	sandbox      *SandboxExecutor
}

// NewRegistryDeterminismExecutor creates an executor that runs functions via registry + sandbox.
// Only functions with a WASM binary are executable; backend/deployment execution is not supported.
func NewRegistryDeterminismExecutor(registryRepo *registry.RegistryRepository, sandbox *SandboxExecutor) generation.DeterminismExecutor {
	return &RegistryDeterminismExecutor{registryRepo: registryRepo, sandbox: sandbox}
}

// NewRegistryDeterminismExecutorWithSandbox creates an executor with a new SandboxExecutor.
// Returns nil if the sandbox cannot be created (e.g. runtime binary not found); callers can skip setting the executor.
func NewRegistryDeterminismExecutorWithSandbox(registryRepo *registry.RegistryRepository) generation.DeterminismExecutor {
	sandbox, err := NewSandboxExecutor()
	if err != nil {
		logrus.WithError(err).Debug("determinism executor: sandbox not available, runtime checks disabled")
		return nil
	}
	return NewRegistryDeterminismExecutor(registryRepo, sandbox)
}

// Execute runs the function with the given input and returns the output.
func (e *RegistryDeterminismExecutor) Execute(ctx context.Context, functionID uuid.UUID, input []byte) ([]byte, error) {
	fnVersion, err := e.registryRepo.GetLatestFunctionVersion(functionID)
	if err != nil {
		return nil, fmt.Errorf("get latest version: %w", err)
	}
	if len(fnVersion.WasmBinary) == 0 {
		return nil, fmt.Errorf("function has no WASM binary (not executable for determinism check)")
	}
	timeoutMs := fnVersion.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	return e.sandbox.ExecuteFunction(fnVersion, input, timeoutMs)
}
