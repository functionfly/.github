//go:build !cgo

// Package wasm: stub implementation when CGO is disabled (e.g. Docker Alpine build).
// Python/WASM execution is unavailable; callers get a clear error.
package wasm

import (
	"errors"
	"io"
)

var errWasmNotAvailable = errors.New("WASM runtime not available in this build (requires CGO); rebuild with CGO enabled for Python/WASM execution")

// PythonRuntime stub when built without CGO
type PythonRuntime struct{}

// NewPythonRuntime returns an error when CGO is disabled
func NewPythonRuntime(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

// NewPythonRuntimeWithDebug returns an error when CGO is disabled
func NewPythonRuntimeWithDebug(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, debug bool) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

// Init returns an error when CGO is disabled
func (r *PythonRuntime) Init() error {
	return errWasmNotAvailable
}

// LoadCode returns an error when CGO is disabled
func (r *PythonRuntime) LoadCode(code string) error {
	return errWasmNotAvailable
}

// Execute returns an error when CGO is disabled
func (r *PythonRuntime) Execute(input []byte) ([]byte, error) {
	return nil, errWasmNotAvailable
}

// Close is a no-op for the stub
func (r *PythonRuntime) Close() error {
	return nil
}
