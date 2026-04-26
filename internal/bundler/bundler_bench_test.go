package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/functionfly/functionfly/internal/manifest"
)

// ─── Benchmarks ────────────────────────────────────────────────────────────────

func BenchmarkDetectToolchains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectToolchains()
	}
}

func BenchmarkDetectAvailableRuntimes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectAvailableRuntimes()
	}
}

func BenchmarkManifestValidation_NewRuntime(b *testing.B) {
	runtimes := []string{"rust", "go", "c", "cpp", "ruby", "kotlin", "swift"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt := runtimes[i%len(runtimes)]
		m := &manifest.Manifest{
			Name:    "bench-fn",
			Version: "1.0.0",
			Runtime: rt,
		}
		m.Validate()
	}
}

func BenchmarkManifestValidation_LegacyRuntime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := &manifest.Manifest{
			Name:    "bench-fn",
			Version: "1.0.0",
			Runtime: "node20",
		}
		m.Validate()
	}
}

func BenchmarkValidateWasmModule(b *testing.B) {
	wasm := createMinimalValidWASM()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateWasmModule(wasm)
	}
}

func BenchmarkEscapeWatString_Short(b *testing.B) {
	input := "hello world"
	for i := 0; i < b.N; i++ {
		escapeWatString(input)
	}
}

func BenchmarkEscapeWatString_Long(b *testing.B) {
	input := make([]byte, 4096)
	for i := range input {
		input[i] = 'a'
	}
	input[100] = '\n'
	input[200] = '"'
	input[300] = '\\'
	s := string(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		escapeWatString(s)
	}
}

func BenchmarkEncodeSourceLength(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encodeSourceLength(4096)
	}
}

func BenchmarkFormatCExports(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatCExports()
	}
}

// ─── Entry File Detection Benchmarks ───────────────────────────────────────────

func BenchmarkEntryFinder_AllRuntimes(b *testing.B) {
	runtimes := []string{"node18", "python3.11", "rust", "go", "c", "cpp", "ruby", "kotlin", "swift"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt := runtimes[i%len(runtimes)]
		getEntryFileCandidates(rt)
	}
}

// ─── Ruby Fallback Wrapper Benchmark ──────────────────────────────────────────

func BenchmarkCreateRubyFallbackWasmWrapper(b *testing.B) {
	if _, err := findWat2Wasm(); err != nil {
		b.Skip("wat2wasm not available")
	}

	m := &manifest.Manifest{Name: "bench-ruby", Version: "1.0.0", Runtime: "ruby"}
	source := `def handler(input); '{"ok":true,"data":"hello"}'; end`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createRubyFallbackWasmWrapper(source, m)
	}
}

// ─── File I/O Benchmarks ──────────────────────────────────────────────────────

func BenchmarkReadEntryFile(b *testing.B) {
	tmpDir := b.TempDir()
	src := []byte(`// test source code`)
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), src, 0644); err != nil {
		b.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{Name: "bench", Version: "1.0.0", Runtime: "go", Entry: "main.go"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadEntryFile(m)
	}
}

// ─── WASM Validation Config Benchmarks ────────────────────────────────────────

func BenchmarkDefaultWASMValidationConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultWASMValidationConfig()
	}
}

func BenchmarkValidateWASM_WithConfig(b *testing.B) {
	wasm := createMinimalValidWASM()
	config := DefaultWASMValidationConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateWASM(wasm, config)
	}
}

// ─── Runtime Type Parsing Benchmark ───────────────────────────────────────────

func BenchmarkManifestApplyDefaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := &manifest.Manifest{Name: "bench", Version: "1.0.0", Runtime: "rust"}
		// Validate triggers defaults via Load path; just measure Validate cost
		m.Validate()
	}
}

// ─── Memory Allocation Benchmarks ──────────────────────────────────────────────

func BenchmarkCreateMinimalValidWASM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createMinimalValidWASM()
	}
}

func BenchmarkCreateWASMWithImports(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createWASMWithImports(10)
	}
}
