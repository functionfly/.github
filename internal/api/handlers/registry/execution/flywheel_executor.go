package execution

import (
	"encoding/json"
	"unsafe"

	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
)

// LocalExecutor runs registry functions locally via the sandbox/WASM path.
// It implements flywheel.RegistryFunctionExecutor for use with the Flywheel execution adapter.
type LocalExecutor struct{}

// NewLocalExecutor returns a RegistryFunctionExecutor that executes functions locally.
func NewLocalExecutor() flywheel.RegistryFunctionExecutor {
	return &LocalExecutor{}
}

// Execute runs the given function version with the input and timeout.
func (e *LocalExecutor) Execute(fnVersion *registry.RegistryFunctionVersion, input []byte, timeoutMs int) ([]byte, error) {
	maxMemoryMB := fnVersion.MemoryMB
	if maxMemoryMB <= 0 {
		maxMemoryMB = 128
	}
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	// storage.RegistryFunctionVersion is an alias for registry.RegistryFunctionVersion; same underlying type
	fnVersionStorage := (*storage.RegistryFunctionVersion)(unsafe.Pointer(fnVersion))
	out, err := ExecuteLocally(fnVersionStorage, json.RawMessage(input), maxMemoryMB, timeoutMs)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}
