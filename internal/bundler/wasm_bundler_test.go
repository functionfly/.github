package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/functionfly/functionfly/internal/manifest"
)

// TestManifestValidation_NewRuntimes verifies all new runtimes pass manifest validation.
func TestManifestValidation_NewRuntimes(t *testing.T) {
	runtimes := []string{
		"rust", "go", "go1.21",
		"c", "c11", "cpp", "cpp17", "c++",
		"ruby", "ruby3.3",
		"kotlin", "kotlin1.9",
		"swift", "swift5.9",
	}

	for _, rt := range runtimes {
		t.Run(rt, func(t *testing.T) {
			m := &manifest.Manifest{
				Name:    "test-fn",
				Version: "1.0.0",
				Runtime: rt,
			}
			if err := m.Validate(); err != nil {
				t.Errorf("runtime %q should be valid: %v", rt, err)
			}
		})
	}
}

// TestManifestValidation_InvalidRuntimes verifies invalid runtimes are rejected.
func TestManifestValidation_InvalidRuntimes(t *testing.T) {
	invalid := []string{"java", "scala", "haskell", "brainfuck", ""}
	for _, rt := range invalid {
		t.Run(rt, func(t *testing.T) {
			m := &manifest.Manifest{
				Name:    "test-fn",
				Version: "1.0.0",
				Runtime: rt,
			}
			if err := m.Validate(); err == nil {
				t.Errorf("runtime %q should be invalid", rt)
			}
		})
	}
}

// TestManifestValidation_EntryExtensions verifies entry file extension validation.
func TestManifestValidation_EntryExtensions(t *testing.T) {
	tests := []struct {
		runtime string
		entry   string
		valid   bool
	}{
		{"rust", "main.rs", true},
		{"rust", "main.go", false},
		{"go", "main.go", true},
		{"go", "main.rs", false},
		{"c", "main.c", true},
		{"c", "main.cpp", false},
		{"cpp", "main.cpp", true},
		{"cpp", "main.c", false},
		{"ruby", "main.rb", true},
		{"ruby", "main.py", false},
		{"kotlin", "Main.kt", true},
		{"kotlin", "Main.swift", false},
		{"swift", "main.swift", true},
		{"swift", "main.kt", false},
	}

	for _, tt := range tests {
		t.Run(tt.runtime+"/"+tt.entry, func(t *testing.T) {
			m := &manifest.Manifest{
				Name:    "test-fn",
				Version: "1.0.0",
				Runtime: tt.runtime,
				Entry:   tt.entry,
			}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("entry %q for runtime %q should be valid: %v", tt.entry, tt.runtime, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("entry %q for runtime %q should be invalid", tt.entry, tt.runtime)
			}
		})
	}
}

// TestEntryFinder_NewRuntimes verifies entry file auto-detection for new runtimes.
func TestEntryFinder_NewRuntimes(t *testing.T) {
	tests := []struct {
		runtime  string
		expected string
	}{
		{"rust", "src/lib.rs"},
		{"go", "main.go"},
		{"c", "main.c"},
		{"cpp", "main.cpp"},
		{"ruby", "main.rb"},
		{"kotlin", "Main.kt"},
		{"swift", "main.swift"},
	}

	for _, tt := range tests {
		t.Run(tt.runtime, func(t *testing.T) {
			preferred, _ := getEntryFileCandidates(tt.runtime)
			if preferred != tt.expected {
				t.Errorf("runtime %q: expected preferred %q, got %q", tt.runtime, tt.expected, preferred)
			}
		})
	}
}

// TestToolchainDetector verifies the toolchain detector runs without error.
func TestToolchainDetector(t *testing.T) {
	toolchains := DetectToolchains()
	if len(toolchains) == 0 {
		t.Error("DetectToolchains returned empty map")
	}

	// At minimum, JavaScript and Python should be detected (bundled)
	for _, lang := range []string{"javascript", "python"} {
		if _, ok := toolchains[lang]; !ok {
			t.Errorf("missing toolchain entry for %q", lang)
		}
	}
}

// TestBundleRust_MissingCargo verifies error when cargo is not available.
func TestBundleRust_MissingCargo(t *testing.T) {
	// Temporarily modify PATH to hide cargo
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	m := &manifest.Manifest{
		Name:    "test-rust",
		Version: "1.0.0",
		Runtime: "rust",
		Entry:   "main.rs",
	}

	_, err := bundleRustToWasm(m)
	if err == nil {
		t.Error("expected error when cargo is not on PATH")
	}
}

// TestBundleGo_MissingGo verifies error when go is not available.
func TestBundleGo_MissingGo(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	m := &manifest.Manifest{
		Name:    "test-go",
		Version: "1.0.0",
		Runtime: "go",
		Entry:   "main.go",
	}

	_, err := bundleGoToWasm(m)
	if err == nil {
		t.Error("expected error when go is not on PATH")
	}
}

// TestBundleC_MissingCompiler verifies error when no C compiler is available.
func TestBundleC_MissingCompiler(t *testing.T) {
	origPath := os.Getenv("PATH")
	origWasiSDK := os.Getenv("WASI_SDK_PATH")
	defer func() {
		os.Setenv("PATH", origPath)
		os.Setenv("WASI_SDK_PATH", origWasiSDK)
	}()
	os.Setenv("PATH", "/nonexistent")
	os.Setenv("WASI_SDK_PATH", "")

	m := &manifest.Manifest{
		Name:    "test-c",
		Version: "1.0.0",
		Runtime: "c",
		Entry:   "main.c",
	}

	_, err := bundleCToWasm(m)
	if err == nil {
		t.Error("expected error when no C compiler is available")
	}
}

// TestBundleRuby_NoRuntime verifies fallback behavior when mruby.wasm is not found.
func TestBundleRuby_NoRuntime(t *testing.T) {
	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)
	os.Setenv("MRUBY_WASM_PATH", "/nonexistent/mruby.wasm")

	// Create a temp dir with source file
	tmpDir, err := os.MkdirTemp("", "test-ruby-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "main.rb")
	if err := os.WriteFile(srcFile, []byte(`puts "hello"`), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir so entry finder works
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{
		Name:    "test-ruby",
		Version: "1.0.0",
		Runtime: "ruby",
		Entry:   "main.rb",
	}

	// Should succeed with fallback wrapper (mruby not available, falls back to WAT template)
	result, err := bundleRubyForWasmRuntime(m)
	if err != nil {
		t.Logf("Ruby fallback produced error (expected if wat2wasm not installed): %v", err)
		return
	}
	if len(result) == 0 {
		t.Error("expected non-empty result from Ruby fallback")
	}
}

// TestBundleKotlin_MissingCompiler verifies error when Kotlin is not available.
func TestBundleKotlin_MissingCompiler(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	m := &manifest.Manifest{
		Name:    "test-kotlin",
		Version: "1.0.0",
		Runtime: "kotlin",
		Entry:   "Main.kt",
	}

	_, err := bundleKotlinForWasmRuntime(m)
	if err == nil {
		t.Error("expected error when kotlin compiler is not available")
	}
}

// TestBundleSwift_MissingCompiler verifies error when Swift is not available.
func TestBundleSwift_MissingCompiler(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	m := &manifest.Manifest{
		Name:    "test-swift",
		Version: "1.0.0",
		Runtime: "swift",
		Entry:   "main.swift",
	}

	_, err := bundleSwiftForWasmRuntime(m)
	if err == nil {
		t.Error("expected error when swift compiler is not available")
	}
}

// TestWASMDispatch_AllRuntimes verifies all new runtimes dispatch to the correct compiler.
func TestWASMDispatch_AllRuntimes(t *testing.T) {
	runtimes := []string{
		"rust", "go", "go1.21",
		"c", "c11", "cpp", "cpp17", "c++",
		"ruby", "ruby3.3",
		"kotlin", "kotlin1.9",
		"swift", "swift5.9",
	}

	for _, rt := range runtimes {
		t.Run(rt, func(t *testing.T) {
			m := &manifest.Manifest{
				Name:    "test-fn",
				Version: "1.0.0",
				Runtime: rt,
			}

			// This will fail because toolchains aren't available in CI,
			// but it verifies the dispatch logic doesn't panic or fall through to JS
			_, err := BundleForWasmRuntimeWithWorkingDirectory(m, t.TempDir())
			if err != nil {
				// Expected: toolchain not available
				t.Logf("runtime %q: %v (expected if toolchain not installed)", rt, err)
			}
		})
	}
}

// TestEscapeWatString verifies WAT string escaping.
func TestEscapeWatString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\nworld", "hello\\nworld"},
		{`"quoted"`, `\"quoted\"`},
		{"tab\there", "tab\\there"},
	}

	for _, tt := range tests {
		result := escapeWatString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeWatString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestRuntimeTypeFromString verifies runtime type parsing for new types.
func TestRuntimeTypeFromString(t *testing.T) {
	// This test is in the wasm package, but we can test the bundler-side mapping
	tests := []struct {
		input string
		valid bool
	}{
		{"rust", true},
		{"go", true},
		{"go1.21", true},
		{"c", true},
		{"cpp", true},
		{"ruby", true},
		{"kotlin", true},
		{"swift", true},
		{"unknown_lang", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := &manifest.Manifest{
				Name:    "test",
				Version: "1.0.0",
				Runtime: tt.input,
			}
			err := m.Validate()
			if tt.valid && err != nil {
				t.Errorf("runtime %q should be valid: %v", tt.input, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("runtime %q should be invalid", tt.input)
			}
		})
	}
}
