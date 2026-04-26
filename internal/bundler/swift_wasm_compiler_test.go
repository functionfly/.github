package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/functionfly/functionfly/internal/manifest"
)

// ─── Swift Compiler Tests ──────────────────────────────────────────────────────

func TestBundleSwift_MissingEntry(t *testing.T) {
	t.Parallel()
	m := &manifest.Manifest{Name: "test-swift", Version: "1.0.0", Runtime: "swift", Entry: "nonexistent.swift"}
	_, err := bundleSwiftForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error for missing entry file")
	}
}

func TestBundleSwift_EmptySource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.swift")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{Name: "test-swift", Version: "1.0.0", Runtime: "swift", Entry: "empty.swift"}
	_, err := bundleSwiftForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestBundleSwift_NoCompiler(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "main.swift")
	if err := os.WriteFile(src, []byte(`print("hello")`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	m := &manifest.Manifest{Name: "test-swift", Version: "1.0.0", Runtime: "swift", Entry: "main.swift"}
	_, err := bundleSwiftForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error when swift compiler not available")
	}
}

// ─── Swift Package Scaffolding ─────────────────────────────────────────────────

func TestCreateSwiftPackage(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := []byte(`print("hello")`)
	m := &manifest.Manifest{Name: "test-swift", Version: "1.0.0", Runtime: "swift"}

	err := createSwiftPackage(tmpDir, src, m)
	if err != nil {
		t.Fatalf("createSwiftPackage failed: %v", err)
	}

	// Verify package structure
	files := []string{"Package.swift", "Sources/main.swift"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// Verify Package.swift contains function name
	pkgContent, err := os.ReadFile(filepath.Join(tmpDir, "Package.swift"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(string(pkgContent), "test-swift") {
		t.Error("Package.swift should contain function name")
	}
}

// ─── Swift WASM Output Finder ──────────────────────────────────────────────────

func TestFindSwiftWasmOutput_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, err := findSwiftWasmOutput(tmpDir)
	if err == nil {
		t.Fatal("expected error when no WASM files found")
	}
}

func TestFindSwiftWasmOutput_ValidFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	wasmDir := filepath.Join(tmpDir, ".build", "wasm32-unknown-wasi", "release")
	os.MkdirAll(wasmDir, 0755)
	wasmContent := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(filepath.Join(wasmDir, "test.wasm"), wasmContent, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := findSwiftWasmOutput(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(wasmContent) {
		t.Errorf("expected %d bytes, got %d", len(wasmContent), len(result))
	}
}

// ─── Swift Compiler Detection ──────────────────────────────────────────────────

func TestFindSwiftWasmCompiler_NotFound(t *testing.T) {
	t.Parallel()
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	_, _, err := findSwiftWasmCompiler()
	if err == nil {
		t.Fatal("expected error when swift compiler not available")
	}
}

func TestFindSwiftWasmCompiler_Available(t *testing.T) {
	t.Parallel()
	compiler, toolchain, err := findSwiftWasmCompiler()
	if err != nil {
		t.Skip("Swift WASM compiler not available")
	}
	t.Logf("Found Swift compiler: %s (%s)", compiler, toolchain)
	if toolchain != "carton" && toolchain != "swiftwasm" {
		t.Errorf("unexpected toolchain: %s", toolchain)
	}
}

// ─── String Utilities ──────────────────────────────────────────────────────────

func TestContains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello world", "world", true},
		{"hello world", "xyz", false},
		{"", "", true},
		{"a", "ab", false},
		{"swiftwasm", "wasm", true},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestSearchSubstring(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello", "ell", true},
		{"hello", "xyz", false},
		{"", "a", false},
		{"a", "", true},
	}

	for _, tt := range tests {
		result := searchSubstring(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("searchSubstring(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

// helper
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}
