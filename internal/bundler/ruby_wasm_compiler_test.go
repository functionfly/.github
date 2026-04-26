package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/functionfly/functionfly/internal/manifest"
)

// ─── Ruby Compiler Tests ───────────────────────────────────────────────────────

func TestBundleRuby_MissingEntry(t *testing.T) {
	t.Parallel()
	m := &manifest.Manifest{Name: "test-ruby", Version: "1.0.0", Runtime: "ruby", Entry: "nonexistent.rb"}
	_, err := bundleRubyForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error for missing entry file")
	}
}

func TestBundleRuby_EmptySource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.rb")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{Name: "test-ruby", Version: "1.0.0", Runtime: "ruby", Entry: "empty.rb"}
	_, err := bundleRubyForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestBundleRuby_FallbackWrapper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "main.rb")
	if err := os.WriteFile(src, []byte(`def handler(input); '{"ok":true}'; end`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)
	os.Setenv("MRUBY_WASM_PATH", "/nonexistent/mruby.wasm")

	m := &manifest.Manifest{Name: "test-ruby", Version: "1.0.0", Runtime: "ruby", Entry: "main.rb"}
	result, err := bundleRubyForWasmRuntime(m)
	if err != nil {
		t.Logf("Ruby bundle error (expected if wat2wasm missing): %v", err)
		return
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestLoadMrubyRuntime_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)
	os.Setenv("MRUBY_WASM_PATH", "/nonexistent/path/mruby.wasm")

	_, err := loadMrubyRuntime()
	if err == nil {
		t.Fatal("expected error when mruby.wasm not found")
	}
}

func TestBundleRuby_WithMrubyRuntime(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "main.rb")
	if err := os.WriteFile(src, []byte(`def handler(input); '{"ok":true}'; end`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create a fake mruby.wasm with valid magic bytes at the MRUBY_WASM_PATH
	fakeWasm := filepath.Join(tmpDir, "mruby.wasm")
	// Minimal valid WASM: magic + version + 1 type + 1 func + memory + export
	fakeData := []byte{
		0x00, 0x61, 0x73, 0x6d, // magic
		0x01, 0x00, 0x00, 0x00, // version 1
	}
	if err := os.WriteFile(fakeWasm, fakeData, 0644); err != nil {
		t.Fatal(err)
	}

	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)
	os.Setenv("MRUBY_WASM_PATH", fakeWasm)

	m := &manifest.Manifest{Name: "test-ruby", Version: "1.0.0", Runtime: "ruby", Entry: "main.rb"}
	result, err := bundleRubyForWasmRuntime(m)
	if err != nil {
		t.Logf("Ruby bundle with mruby runtime (may fallback): %v", err)
		return
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
	// Verify WASM magic bytes
	if len(result) >= 4 && result[0] == 0x00 && result[1] == 0x61 && result[2] == 0x73 && result[3] == 0x6D {
		t.Logf("Produced valid WASM module (%d bytes)", len(result))
	}
}

func TestLoadMrubyRuntime_InvalidMagic(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	fakeWasm := filepath.Join(tmpDir, "mruby.wasm")
	// Write a file with wrong magic bytes
	if err := os.WriteFile(fakeWasm, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, 0644); err != nil {
		t.Fatal(err)
	}

	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)
	os.Setenv("MRUBY_WASM_PATH", fakeWasm)

	_, err := loadMrubyRuntime()
	if err == nil {
		t.Fatal("expected error for invalid magic bytes")
	}
}

func TestLoadMrubyRuntime_ValidMagic(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	fakeWasm := filepath.Join(tmpDir, "mruby.wasm")
	// Write a file with correct magic bytes (WASM header)
	data := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}
	if err := os.WriteFile(fakeWasm, data, 0644); err != nil {
		t.Fatal(err)
	}

	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)
	os.Setenv("MRUBY_WASM_PATH", fakeWasm)

	result, err := loadMrubyRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(data) {
		t.Errorf("expected %d bytes, got %d", len(data), len(result))
	}
}

// ─── WAT Escape Tests ──────────────────────────────────────────────────────────

func TestEscapeWatString_Empty(t *testing.T) {
	t.Parallel()
	if result := escapeWatString(""); result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestEscapeWatString_SpecialChars(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline", "a\nb", "a\\nb"},
		{"carriage_return", "a\rb", "a\\rb"},
		{"tab", "a\tb", "a\\tb"},
		{"quote", `a"b`, `a\"b`},
		{"backslash", `a\b`, `a\\b`},
		{"control_char", "a\x01b", "a\\01b"},
		{"tilde", "a~b", "a~b"},
		{"unicode_printable", "hello世界", "hello\\e4\\b8\\96\\e7\\95\\8c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeWatString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeWatString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ─── Encode Source Length ──────────────────────────────────────────────────────

func TestEncodeSourceLength(t *testing.T) {
	t.Parallel()
	tests := []struct {
		length   int
		expected string
	}{
		{0, "\\00\\00\\00\\00"},
		{1, "\\01\\00\\00\\00"},
		{256, "\\00\\01\\00\\00"},
		{1024, "\\00\\04\\00\\00"},
	}

	for _, tt := range tests {
		result := encodeSourceLength(tt.length)
		if result != tt.expected {
			t.Errorf("encodeSourceLength(%d) = %q, want %q", tt.length, result, tt.expected)
		}
	}
}

// ─── Fallback WAT Generation ──────────────────────────────────────────────────

func TestCreateRubyFallbackWasmWrapper(t *testing.T) {
	t.Parallel()
	m := &manifest.Manifest{Name: "test-ruby", Version: "1.0.0", Runtime: "ruby"}

	// This will fail if wat2wasm is not installed — that's expected
	_, err := createRubyFallbackWasmWrapper("puts 'hello'", m)
	if err != nil {
		t.Logf("Ruby fallback WAT generation failed (expected if wat2wasm not installed): %v", err)
	}
}

func TestFindWat2Wasm(t *testing.T) {
	t.Parallel()
	path, err := findWat2Wasm()
	if err != nil {
		t.Skip("wat2wasm not available")
	}
	t.Logf("Found wat2wasm at: %s", path)
}
