package bundler

import (
	"fmt"
)

// validateWasmModule performs basic validation on compiled WebAssembly bytes
func validateWasmModule(wasmBytes []byte) error {
	if len(wasmBytes) < 8 {
		return fmt.Errorf("WebAssembly module too small")
	}

	// Check magic bytes (\0asm)
	if string(wasmBytes[0:4]) != "\x00asm" {
		return fmt.Errorf("invalid WebAssembly magic bytes")
	}

	// Check version (should be 1)
	if wasmBytes[4] != 1 || wasmBytes[5] != 0 || wasmBytes[6] != 0 || wasmBytes[7] != 0 {
		return fmt.Errorf("unsupported WebAssembly version")
	}

	return nil
}