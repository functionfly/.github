//go:build !cgo

package wasm

import (
	"sync"
	"time"
)

type SwiftPoolEntry struct{}

type SwiftRuntimePool struct {
	mu      sync.RWMutex
	maxSize int
}

func NewSwiftRuntimePool(maxSize int, maxAge time.Duration) *SwiftRuntimePool {
	return &SwiftRuntimePool{maxSize: maxSize}
}

func (p *SwiftRuntimePool) Get(wasmBinary []byte, handler HostFunctionHandler, config *WASMSecurityConfig) (*SwiftWASIRuntime, error) {
	return NewSwiftWASIRuntimeWithConfig(wasmBinary, handler, config)
}

func (p *SwiftRuntimePool) Stats() map[string]interface{} {
	return map[string]interface{}{"entries": 0, "cgo": false}
}

func (p *SwiftRuntimePool) Close() {}

var GlobalSwiftPool = NewSwiftRuntimePool(0, 0)
