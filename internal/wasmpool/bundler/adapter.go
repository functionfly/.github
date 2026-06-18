// Package bundleradapter: adapter that wraps the orchestrator's
// bundler.BundleService to satisfy the wasm.Bundler interface. This
// keeps the wasm module decoupled from the orchestrator's bundler
// package.
package bundleradapter

import (
	wasmpool "github.com/functionfly/functionfly/internal/wasm"
	"github.com/functionfly/functionfly/internal/bundler"
)

// Adapter wraps bundler.BundleService to satisfy wasmpool.Bundler.
type Adapter struct {
	Svc *bundler.BundleService
}

// New returns a wasmpool.Bundler backed by the given bundle service.
func New(svc *bundler.BundleService) wasmpool.Bundler { return &Adapter{Svc: svc} }

// GetWASMBinary delegates to the bundle service.
func (a *Adapter) GetWASMBinary(input []byte) ([]byte, error) {
	return bundler.GetWASMBinary(input)
}

// ExtractMetadata maps the orchestrator's WASMMetadata to the wasm
// module's minimal WASMMetadata struct.
func (a *Adapter) ExtractMetadata(wasmBinary []byte) (*wasmpool.WASMMetadata, error) {
	md, err := bundler.ExtractMetadata(wasmBinary)
	if err != nil || md == nil {
		return nil, err
	}
	return &wasmpool.WASMMetadata{
		HandlerName:       md.HandlerName,
		MemoryPages:       md.MemoryPages,
		ExportedFunctions: md.ExportedFunctions,
		WASITarget:        md.WASITarget,
	}, nil
}
