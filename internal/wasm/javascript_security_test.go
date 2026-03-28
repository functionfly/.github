//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// Tests for JavaScript/TypeScript security validation
package wasm

import (
	"testing"
)

func TestJavaScriptSecurityConfigDefaults(t *testing.T) {
	config := DefaultJavaScriptSecurityConfig()

	if config.MaxSourceSize != 1024*1024 {
		t.Errorf("Expected MaxSourceSize=1048576, got %d", config.MaxSourceSize)
	}

	if config.MaxCompilationTime != 60*second {
		t.Errorf("Expected MaxCompilationTime=60s, got %v", config.MaxCompilationTime)
	}

	if !config.EnableEvalBlock {
		t.Error("Expected EnableEvalBlock=true by default")
	}

	if !config.RequireFunctionExport {
		t.Error("Expected RequireFunctionExport=true by default")
	}
}

func TestValidateSourceCode_EmptySource(t *testing.T) {
	config := DefaultJavaScriptSecurityConfig()
	config.RequireFunctionExport = false

	err := ValidateSourceCode([]byte(""), config)
	if err != nil {
		t.Errorf("Expected empty source to pass validation, got: %v", err)
	}
}

func TestValidateSourceCode_SizeLimit(t *testing.T) {
	config := DefaultJavaScriptSecurityConfig()
	config.MaxSourceSize = 100
	config.RequireFunctionExport = false

	// Create source that exceeds limit
	longSource := make([]byte, 150)
	for i := range longSource {
		longSource[i] = 'x'
	}

	err := ValidateSourceCode(longSource, config)
	if err == nil {
		t.Error("Expected error for source exceeding size limit")
	}
}

func TestValidateSourceCode_EvalBlocking(t *testing.T) {
	config := DefaultJavaScriptSecurityConfig()
	config.RequireFunctionExport = false

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{"safe code", "const x = 1;", false},
		{"safe function", "function hello() { return 'world'; }; export default hello;", false},
		{"eval blocked", "eval('dangerous');", true},
		{"Function() blocked", "new Function('x');", true},
		{"setTimeout string blocked", "setTimeout('alert(1)', 1000);", true},
		{"prototype pollution", "obj.__proto__.x = 1;", true},
		{"constructor access", "obj.constructor;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourceCode([]byte(tt.source), config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourceCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSourceCode_RequireExport(t *testing.T) {
	config := DefaultJavaScriptSecurityConfig()
	config.RequireFunctionExport = true

	// Source without export
	err := ValidateSourceCode([]byte("const x = 1;"), config)
	if err == nil {
		t.Error("Expected error for source without export")
	}

	// Source with export default
	err = ValidateSourceCode([]byte("export default function() { return 1; };"), config)
	if err != nil {
		t.Errorf("Expected source with export default to pass, got: %v", err)
	}

	// Source with module.exports
	err = ValidateSourceCode([]byte("module.exports = function() { return 1; };"), config)
	if err != nil {
		t.Errorf("Expected source with module.exports to pass, got: %v", err)
	}

	// Source with exports.foo
	err = ValidateSourceCode([]byte("exports.foo = function() { return 1; };"), config)
	if err != nil {
		t.Errorf("Expected source with exports.foo to pass, got: %v", err)
	}
}

func TestValidateSourceCode_BlockedPatterns(t *testing.T) {
	config := DefaultJavaScriptSecurityConfig()
	config.RequireFunctionExport = false
	config.BlockedPatterns = []string{
		`console\.log`,
		`debugger`,
	}

	// Should pass - safe code
	err := ValidateSourceCode([]byte("const x = 1 + 2;"), config)
	if err != nil {
		t.Errorf("Expected safe code to pass, got: %v", err)
	}

	// Should fail - blocked pattern
	err = ValidateSourceCode([]byte("console.log('test');"), config)
	if err == nil {
		t.Error("Expected error for blocked pattern")
	}

	// Should fail - debugger
	err = ValidateSourceCode([]byte("debugger;"), config)
	if err == nil {
		t.Error("Expected error for debugger statement")
	}
}

func TestSourceCodeHash(t *testing.T) {
	source1 := []byte("const x = 1;")
	source2 := []byte("const x = 2;")

	hash1 := SourceCodeHash(source1)
	hash2 := SourceCodeHash(source2)

	if hash1 == hash2 {
		t.Error("Different sources should have different hashes")
	}

	if len(hash1) != 64 { // SHA-256 produces 64 hex characters
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}

	// Same source should produce same hash
	hash1Again := SourceCodeHash(source1)
	if hash1 != hash1Again {
		t.Error("Same source should produce same hash")
	}
}

func TestReadUint32LE(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
	}{
		{"zero", []byte{0, 0, 0, 0}, 0},
		{"one", []byte{1, 0, 0, 0}, 1},
		{"256", []byte{0, 1, 0, 0}, 256},
		{"large", []byte{0, 0, 1, 0}, 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readUint32LE(tt.data)
			if result != tt.expected {
				t.Errorf("readUint32LE() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestExtractWasmBinary_RawWasm(t *testing.T) {
	raw := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	extracted, err := extractWasmBinary(raw)
	if err != nil {
		t.Fatalf("extractWasmBinary() returned error: %v", err)
	}

	if string(extracted) != string(raw) {
		t.Fatalf("extractWasmBinary() returned different binary")
	}
}

func TestExtractWasmBinary_MetadataWrapped(t *testing.T) {
	raw := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	metadata := []byte(`{"handler":"main"}`)
	metadataLen := len(metadata)

	bundled := make([]byte, 0, 4+4+metadataLen+len(raw))
	bundled = append(bundled, []byte("FFWB")...)
	bundled = append(bundled,
		byte(metadataLen>>24),
		byte(metadataLen>>16),
		byte(metadataLen>>8),
		byte(metadataLen),
	)
	bundled = append(bundled, metadata...)
	bundled = append(bundled, raw...)

	extracted, err := extractWasmBinary(bundled)
	if err != nil {
		t.Fatalf("extractWasmBinary() returned error: %v", err)
	}

	if string(extracted) != string(raw) {
		t.Fatalf("extractWasmBinary() did not extract raw WASM payload")
	}
}

func TestExtractWasmBinary_MetadataWrappedInvalidLength(t *testing.T) {
	raw := []byte{0x00, 0x61, 0x73, 0x6d}
	bundled := make([]byte, 0, 4+4+len(raw))
	bundled = append(bundled, []byte("FFWB")...)
	// Deliberately oversized metadata length (255 bytes) with no metadata payload.
	bundled = append(bundled, 0x00, 0x00, 0x00, 0xFF)
	bundled = append(bundled, raw...)

	_, err := extractWasmBinary(bundled)
	if err == nil {
		t.Fatalf("expected error for malformed metadata-wrapped binary")
	}
}

func TestExtractWasmBinary_MetadataWrappedInvalidPayload(t *testing.T) {
	metadata := []byte(`{"handler":"main"}`)
	metadataLen := len(metadata)

	bundled := make([]byte, 0, 4+4+metadataLen+4)
	bundled = append(bundled, []byte("FFWB")...)
	bundled = append(bundled,
		byte(metadataLen>>24),
		byte(metadataLen>>16),
		byte(metadataLen>>8),
		byte(metadataLen),
	)
	bundled = append(bundled, metadata...)
	// Not a WASM payload (missing \x00asm magic bytes).
	bundled = append(bundled, []byte("nope")...)

	_, err := extractWasmBinary(bundled)
	if err == nil {
		t.Fatalf("expected error for non-WASM payload in wrapped bundle")
	}
}

const second = 1000000000 // 1 second in nanoseconds
