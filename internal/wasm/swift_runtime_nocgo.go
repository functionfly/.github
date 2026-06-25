//go:build !cgo

// Package wasm — Swift WASI runtime stub for non-CGO builds.
package wasm

import (
	"context"
	"fmt"
)

// SwiftWASIRuntime executes Swift WASM modules via WASI.
// In non-CGO builds, this is a stub that returns an error.
type SwiftWASIRuntime struct{}

func NewSwiftWASIRuntime(wasmBinary []byte, handler HostFunctionHandler) (*SwiftWASIRuntime, error) {
	return nil, fmt.Errorf("Swift WASM runtime requires CGO (wasmtime)")
}

func NewSwiftWASIRuntimeWithConfig(wasmBinary []byte, handler HostFunctionHandler, config *WASMSecurityConfig) (*SwiftWASIRuntime, error) {
	return nil, fmt.Errorf("Swift WASM runtime requires CGO (wasmtime)")
}

func (r *SwiftWASIRuntime) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return nil, fmt.Errorf("Swift WASM runtime requires CGO (wasmtime)")
}

func (r *SwiftWASIRuntime) Close() error { return nil }
