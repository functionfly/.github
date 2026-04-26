package bundler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/functionfly/functionfly/internal/manifest"
)

// ─── Kotlin Compiler Tests ─────────────────────────────────────────────────────

func TestBundleKotlin_MissingEntry(t *testing.T) {
	t.Parallel()
	m := &manifest.Manifest{Name: "test-kotlin", Version: "1.0.0", Runtime: "kotlin", Entry: "Nonexistent.kt"}
	_, err := bundleKotlinForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error for missing entry file")
	}
}

func TestBundleKotlin_EmptySource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "Empty.kt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{Name: "test-kotlin", Version: "1.0.0", Runtime: "kotlin", Entry: "Empty.kt"}
	_, err := bundleKotlinForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestBundleKotlin_NoCompiler(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "Main.kt")
	if err := os.WriteFile(src, []byte(`fun main() { println("hello") }`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	m := &manifest.Manifest{Name: "test-kotlin", Version: "1.0.0", Runtime: "kotlin", Entry: "Main.kt"}
	_, err := bundleKotlinForWasmRuntime(m)
	if err == nil {
		t.Fatal("expected error when kotlin compiler not available")
	}
}

// ─── Kotlin Project Scaffolding ────────────────────────────────────────────────

func TestCreateKotlinWasmProject(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := []byte(`fun main() { println("hello") }`)
	m := &manifest.Manifest{Name: "test-kotlin", Version: "1.0.0", Runtime: "kotlin"}

	err := createKotlinWasmProject(tmpDir, src, m)
	if err != nil {
		t.Fatalf("createKotlinWasmProject failed: %v", err)
	}

	// Verify project structure
	files := []string{
		"settings.gradle.kts",
		"build.gradle.kts",
		"src/wasmWasiMain/kotlin/Main.kt",
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// Verify source content
	content, err := os.ReadFile(filepath.Join(tmpDir, "src/wasmWasiMain/kotlin/Main.kt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(src) {
		t.Errorf("source content mismatch: got %q", string(content))
	}
}

// ─── Kotlin JS Output Finder ──────────────────────────────────────────────────

func TestFindKotlinJSOutput_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, err := findKotlinJSOutput(tmpDir)
	if err == nil {
		t.Fatal("expected error when no JS files found")
	}
}

func TestFindKotlinJSOutput_Found(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	jsContent := []byte(`console.log("hello")`)
	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), jsContent, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := findKotlinJSOutput(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != string(jsContent) {
		t.Errorf("content mismatch: got %q", string(result))
	}
}

// ─── Kotlin WASM Output Finder ────────────────────────────────────────────────

func TestFindKotlinWasmOutput_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, err := findKotlinWasmOutput(tmpDir)
	if err == nil {
		t.Fatal("expected error when no WASM files found")
	}
}

func TestFindKotlinWasmOutput_InvalidFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	// Create a file with wrong magic bytes
	badWasm := filepath.Join(tmpDir, "build", "wasmWasi", "test.wasm")
	os.MkdirAll(filepath.Dir(badWasm), 0755)
	if err := os.WriteFile(badWasm, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := findKotlinWasmOutput(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid WASM magic bytes")
	}
}

func TestFindKotlinWasmOutput_ValidFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	wasmDir := filepath.Join(tmpDir, "build", "wasmWasi")
	os.MkdirAll(wasmDir, 0755)
	wasmContent := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(filepath.Join(wasmDir, "test.wasm"), wasmContent, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := findKotlinWasmOutput(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(wasmContent) {
		t.Errorf("expected %d bytes, got %d", len(wasmContent), len(result))
	}
}

// ─── Gradle Wrapper Creation ───────────────────────────────────────────────────

func TestCreateGradleWrapper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	err := createGradleWrapper(tmpDir)
	if err != nil {
		t.Fatalf("createGradleWrapper failed: %v", err)
	}

	// Verify properties file was created
	propsPath := filepath.Join(tmpDir, "gradle", "wrapper", "gradle-wrapper.properties")
	if _, err := os.Stat(propsPath); os.IsNotExist(err) {
		t.Error("expected gradle-wrapper.properties to exist")
	}
}
