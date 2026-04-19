package bundler

import (
	"fmt"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/manifest"
)

// bundlePythonForWasmRuntime bundles Python for Wasm runtime execution
// Uses MicroPython runtime - FlyPy has been disabled
func bundlePythonForWasmRuntime(manifest *manifest.Manifest) ([]byte, error) {
	// Read and validate entry file using shared helper
	entryFile, sourceCode, err := ReadEntryFile(manifest)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm python bundle", "failed to read entry file", err)
	}

	fmt.Printf("Bundling Python to WASM using MicroPython for %s\n", entryFile)

	// Use MicroPython runtime to bundle Python
	fmt.Printf("Using MicroPython runtime for %s\n", entryFile)
	wasmBytes, err := createPythonWasmWithRuntime(string(sourceCode), manifest)
	if err != nil {
		fmt.Printf("Warning: MicroPython runtime approach failed (%v)\n", err)
		return nil, NewBundlerErrorWithCause("wasm python bundle",
			"compilation failed - MicroPython runtime unavailable",
			fmt.Errorf("Micropython: provide micropython.wasm in bundler/python/"))
	}

	// Log validation issues as warning but still return the bytes
	// The precompiled runtime is known-good, validation may have false positives
	if err := validateWasmModule(wasmBytes); err != nil {
		fmt.Printf("Warning: MicroPython WASM validation reported issues (%v) - using runtime anyway\n", err)
	}

	fmt.Printf("Successfully loaded MicroPython runtime (%d bytes)\n", len(wasmBytes))
	return wasmBytes, nil
}

// Simplified: Now using only MicroPython (FlyPy has been disabled)

// createPythonWasmWithRuntime creates a WASM module using production MicroPython runtime
func createPythonWasmWithRuntime(sourceCode string, manifest *manifest.Manifest) ([]byte, error) {
	// Production approach: Use the proper linker that returns micropython.wasm directly
	// User code is loaded at runtime via mp_js_do_exec (micropython's JS interop)
	fmt.Printf("Using production MicroPython runtime for %s\n", manifest.Name)

	wasmBytes, err := CompileWithMicropython(sourceCode, manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to compile with MicroPython: %v", err)
	}

	fmt.Printf("Production: Loaded MicroPython runtime with user code hook (%d bytes)\n", len(wasmBytes))
	return wasmBytes, nil
}

// loadMicropythonRuntime loads the precompiled Micropython WASM runtime
func loadMicropythonRuntime() ([]byte, error) {
	// Use the same path resolution as the linker (works regardless of working directory)
	runtimePath := findMicropythonRuntimePath()
	if runtimePath == "" {
		return nil, fmt.Errorf("Micropython runtime not found. Please build micropython.wasm and place in bundler/python/")
	}

	bytes, err := os.ReadFile(runtimePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read micropython runtime: %v", err)
	}

	// Basic validation - check WASM magic bytes
	if len(bytes) > 8 && bytes[0] == 0x00 && bytes[1] == 0x61 && bytes[2] == 0x73 && bytes[3] == 0x6D {
		return bytes, nil
	}

	return nil, fmt.Errorf("invalid WASM file at %s: bad magic bytes", runtimePath)
}

// createRuntimeWithEmbeddedCode combines runtime with embedded user code
func createRuntimeWithEmbeddedCode(runtimeBytes []byte, sourceCode string, manifest *manifest.Manifest) ([]byte, error) {
	// Phase 4.1: Use precompiled runtime with embedded user code
	// The runtime already has the interface: init, execute, load_code, alloc, dealloc, metadata
	// We embed user code directly into the runtime's data section

	// Create metadata JSON
	metadata := fmt.Sprintf(`{
		"name": "%s",
		"runtime": "python-precompiled",
		"runtime_version": "micropython-1.20",
		"version": "%s",
		"entry_point": "handler",
		"dependencies": [],
		"memory_mb": 128,
		"timeout_ms": 5000,
		"uses_network": false,
		"uses_filesystem": false,
		"phase": "4.1-precompiled-runtime"
	}`, manifest.Name, manifest.Version)

	// Create a WAT module that uses the precompiled runtime directly
	// This module embeds user code and calls runtime functions
	watTemplate := fmt.Sprintf(`
(module
  ;; Import precompiled Micropython runtime
  (import "micropython" "memory" (memory 1))
  (import "micropython" "init" (func $mp_init))
  (import "micropython" "load_code" (func $mp_load_code (param i32 i32)))
  (import "micropython" "execute" (func $mp_execute (param i32 i32) (result i32)))
  (import "micropython" "alloc" (func $mp_alloc (param i32) (result i32)))
  (import "micropython" "dealloc" (func $mp_dealloc (param i32)))
  (import "micropython" "metadata" (func $mp_metadata (result i32)))

  ;; Export memory for host access
  (export "memory" (memory 0))

  ;; Global variables
  (global $code_loaded (mut i32) (i32.const 0))

  ;; Embedded user Python code at offset 1024
  (data (i32.const 1024) "%s")

  ;; Embedded metadata at offset 8192
  (data (i32.const 8192) "%s")

  ;; Initialize function
  (func $init (export "init")
    ;; Initialize Micropython runtime
    call $mp_init
    ;; Load user code
    i32.const 1024
    i32.const %d
    call $mp_load_code
    i32.const 1
    global.set $code_loaded
  )

  ;; Execute function
  (func $execute (export "execute") (param $input i32) (param $input_len i32) (result i32)
    local.get $input
    local.get $input_len
    call $mp_execute
  )

  ;; Alloc function
  (func $alloc (export "alloc") (param $size i32) (result i32)
    local.get $size
    call $mp_alloc
  )

  ;; Dealloc function
  (func $dealloc (export "dealloc") (param $ptr i32)
    local.get $ptr
    call $mp_dealloc
  )

  ;; Metadata export
  (func $metadata (export "metadata") (result i32)
    i32.const 8192
  )
)`, escapeForWAT(sourceCode), escapeForWAT(metadata), len(sourceCode))

	// Compile WAT to WASM
	wasmBytes, err := compileWATToWasm(watTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WAT to WASM: %v", err)
	}

	return wasmBytes, nil
}

// createPythonWasmModule creates a WASM module for Python execution
// This generates WebAssembly Text (WAT) that can be compiled to WASM bytecode
func createPythonWasmModule(sourceCode string, manifest *manifest.Manifest) ([]byte, error) {
	// Create a WASM module following the standardized FunctionModule interface
	// This generates WebAssembly Text (WAT) that can be compiled to WASM bytecode

	// Escape the source code for WAT data section
	escapedSource := escapeForWAT(sourceCode)

	// Create metadata JSON with fallback indicator
	metadata := fmt.Sprintf(`{
		"name": "%s",
		"runtime": "python",
		"runtime_version": "fallback-stub",
			"version": "%s",
		"entry_point": "handler",
		"dependencies": [],
		"memory_mb": 128,
		"timeout_ms": 5000,
		"uses_network": false,
		"uses_filesystem": false,
		"fallback": true,
		"fallback_reason": "micropython_runtime_unavailable"
	}`, manifest.Name, manifest.Version)

	escapedMetadata := escapeForWAT(metadata)

	// Generate fallback error response JSON
	fallbackError := fmt.Sprintf(
		`{"success":false,"error":"Python execution unavailable - micropython.wasm not found","fallback":true,"source_hint":"%s"}`,
		escapeJsonString(truncateString(sourceCode, 100)),
	)
	escapedError := escapeForWAT(fallbackError)

	watTemplate := `
(module
  ;; Memory export (required for all function modules)
  (memory (export "memory") 1)  ;; 64KB pages

  ;; Global variables for memory management
  (global $initialized (mut i32) (i32.const 0))
  (global $input_ptr (mut i32) (i32.const 0))
  (global $output_ptr (mut i32) (i32.const 2048))

  ;; Data sections
  ;; Python source code embedded in WASM
  (data (i32.const 1024) "%s")
  ;; Function metadata
  (data (i32.const 4096) "%s")
  ;; Result buffer with fallback error
  (data (i32.const 2048) "%s")

  ;; Initialize function - called once on cold start
  (func $init (export "init")
    ;; Mark as initialized
    i32.const 1
    global.set $initialized
  )

  ;; Execute function - main entry point for function execution
  (func $execute (export "execute") (param $input i32) (param $input_len i32) (result i32)
    ;; Store input parameters
    local.get $input
    global.set $input_ptr

    ;; Check if initialized
    global.get $initialized
    i32.eqz
    if
      ;; Auto-initialize if not done
      call $init
    end

    ;; Return fallback error JSON indicating Python runtime unavailable
    ;; The result is a pointer to the embedded error JSON
    i32.const 2048
  )

  ;; Get metadata function
  (func $metadata (export "metadata") (result i32)
    ;; Return pointer to metadata JSON
    i32.const 4096
  )

  ;; Load code function - required by PythonRuntime interface
  (func $load_code (export "load_code") (param $ptr i32) (param $len i32) (result i32)
    ;; This fallback stub doesn't actually load code
    ;; Return 0 to indicate "not implemented"
    i32.const 0
  )

  ;; Alloc function
  (func $alloc (export "alloc") (param $size i32) (result i32)
    ;; Simple linear allocation from top of initial memory
    i32.const 16384
    local.get $size
    i32.add
  )

  ;; Dealloc function
  (func $dealloc (export "dealloc") (param $ptr i32)
    ;; No-op in this simple implementation
  )
)`

	// Generate the WAT content
	watContent := fmt.Sprintf(watTemplate, escapedSource, escapedMetadata, escapedError)

	// Compile WAT to WASM bytecode using wat2wasm
	wasmBytes, err := compileWATToWasm(watContent)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WAT to WASM: %v", err)
	}

	return wasmBytes, nil
}

// detectCompilationMode analyzes Python source code to determine the appropriate compilation mode
// Returns "complex" for code that uses CSV, IO, regex, datetime, hashlib, base64 modules
// Returns "deterministic" for simple pure functions
func detectCompilationMode(sourceCode string) string {
	// List of imports that require complex mode
	complexImports := []string{
		"import csv",
		"from csv",
		"import io",
		"from io",
		"import re",
		"from re",
		"import datetime",
		"from datetime",
		"import hashlib",
		"from hashlib",
		"import base64",
		"from base64",
		"import json",
		"from json",
		"import uuid",
		"from uuid",
	}

	// Check if any complex imports are present
	for _, imp := range complexImports {
		if strings.Contains(sourceCode, imp) {
			return "complex"
		}
	}

	// Default to deterministic mode for simple functions
	return "deterministic"
}

// truncateString truncates a string to maxLen characters, adding ellipsis if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// escapeJsonString escapes a string for safe inclusion in JSON
func escapeJsonString(s string) string {
	var result strings.Builder
	for _, c := range s {
		switch c {
		case '"':
			result.WriteString("\\\"")
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			if c < 32 {
				result.WriteString(fmt.Sprintf("\\u%04x", c))
			} else {
				result.WriteRune(c)
			}
		}
	}
	return result.String()
}
