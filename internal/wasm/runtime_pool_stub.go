//go:build !cgo

package wasm

import (
	"context"
	"io"
)

// PythonRuntimePool stub when built without CGO.
type PythonRuntimePool struct{}

// pathPool stub.
type pathPool struct{}

// PythonRuntimeFactory stub type.
type PythonRuntimeFactory func(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*PythonRuntime, error)

// NewPythonRuntimePool returns a stub pool when CGO is disabled.
func NewPythonRuntimePool(factory PythonRuntimeFactory, maxSize int) *PythonRuntimePool {
	return &PythonRuntimePool{}
}

// Prewarm is a no-op when CGO is disabled.
func (p *PythonRuntimePool) Prewarm(ctx context.Context, wasmPath string, count int) error {
	return nil
}

// Get returns an error when CGO is disabled.
func (p *PythonRuntimePool) Get(ctx context.Context, wasmPath string) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

// Put is a no-op when CGO is disabled.
func (p *PythonRuntimePool) Put(rt *PythonRuntime, wasmPath string) error {
	return nil
}

// Close is a no-op when CGO is disabled.
func (p *PythonRuntimePool) Close() error {
	return nil
}

// Stats returns zero values when CGO is disabled.
func (p *PythonRuntimePool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"disabled": true,
	}
}

// InitPythonRuntimePool is a no-op when CGO is disabled.
func InitPythonRuntimePool(factory PythonRuntimeFactory, poolSize int) {}

// InitPythonRuntimePoolWithPrewarm is a no-op when CGO is disabled.
func InitPythonRuntimePoolWithPrewarm(ctx context.Context, factory PythonRuntimeFactory, poolSize int, wasmPath string, prewarmCount int) error {
	return errWasmNotAvailable
}
