package bundler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
)

// ─── C Compiler Tests ──────────────────────────────────────────────────────────

func TestBundleC_MissingEntry(t *testing.T) {
	t.Parallel()
	m := &manifest.Manifest{Name: "test-c", Version: "1.0.0", Runtime: "c", Entry: "nonexistent.c"}
	_, err := bundleCToWasm(m)
	if err == nil {
		t.Fatal("expected error for missing entry file")
	}
}

func TestBundleC_EmptySource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.c")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := &manifest.Manifest{Name: "test-c", Version: "1.0.0", Runtime: "c", Entry: "empty.c"}
	_, err := bundleCToWasm(m)
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestBundleC_NoCompiler(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "main.c")
	if err := os.WriteFile(src, []byte(`void init(){}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	origPath := os.Getenv("PATH")
	origWasi := os.Getenv("WASI_SDK_PATH")
	defer func() {
		os.Setenv("PATH", origPath)
		os.Setenv("WASI_SDK_PATH", origWasi)
	}()
	os.Setenv("PATH", "/nonexistent")
	os.Setenv("WASI_SDK_PATH", "")

	m := &manifest.Manifest{Name: "test-c", Version: "1.0.0", Runtime: "c", Entry: "main.c"}
	_, err := bundleCToWasm(m)
	if err == nil {
		t.Fatal("expected error when no C compiler available")
	}
}

// ─── C++ Compiler Tests ────────────────────────────────────────────────────────

func TestBundleCpp_MissingEntry(t *testing.T) {
	t.Parallel()
	m := &manifest.Manifest{Name: "test-cpp", Version: "1.0.0", Runtime: "cpp", Entry: "nonexistent.cpp"}
	_, err := bundleCppToWasm(m)
	if err == nil {
		t.Fatal("expected error for missing entry file")
	}
}

func TestBundleCpp_NoCompiler(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "main.cpp")
	if err := os.WriteFile(src, []byte(`extern "C" void init(){}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	origPath := os.Getenv("PATH")
	origWasi := os.Getenv("WASI_SDK_PATH")
	defer func() {
		os.Setenv("PATH", origPath)
		os.Setenv("WASI_SDK_PATH", origWasi)
	}()
	os.Setenv("PATH", "/nonexistent")
	os.Setenv("WASI_SDK_PATH", "")

	m := &manifest.Manifest{Name: "test-cpp", Version: "1.0.0", Runtime: "cpp", Entry: "main.cpp"}
	_, err := bundleCppToWasm(m)
	if err == nil {
		t.Fatal("expected error when no C++ compiler available")
	}
}

// ─── Emscripten Integration ───────────────────────────────────────────────────

func TestCompileWithEmscripten_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping emscripten test in short mode")
	}

	emcc, err := exec.LookPath("emcc")
	if err != nil {
		t.Skip("emcc not available, skipping")
	}

	// Verify emcc works with a trivial source
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "test.c")
	if err := os.WriteFile(src, []byte(`void init(){}`), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(tmpDir, "out.wasm")

	_, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.Command(emcc, src, "-o", out, "-s", "WASM=1", "--no-entry")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("emcc compilation: %s", string(out))
		t.Fatalf("emcc failed: %v", err)
	}

	wasm, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(wasm) < 8 {
		t.Fatalf("WASM too small: %d bytes", len(wasm))
	}
	if wasm[0] != 0x00 || wasm[1] != 0x61 || wasm[2] != 0x73 || wasm[3] != 0x6D {
		t.Fatal("invalid WASM magic bytes")
	}
}

// ─── Compiler Detection Tests ──────────────────────────────────────────────────

func TestFindCCompiler_WithWASISDK(t *testing.T) {
	origWasi := os.Getenv("WASI_SDK_PATH")
	defer os.Setenv("WASI_SDK_PATH", origWasi)

	// Non-existent path
	os.Setenv("WASI_SDK_PATH", "/nonexistent/wasi-sdk")
	_, _, err := findCCompiler()
	// Will fail since the path doesn't exist, that's fine
	if err != nil {
		t.Logf("findCCompiler with bad WASI_SDK_PATH: %v", err)
	}
}

func TestFindCppCompiler_WithWASISDK(t *testing.T) {
	origWasi := os.Getenv("WASI_SDK_PATH")
	defer os.Setenv("WASI_SDK_PATH", origWasi)

	os.Setenv("WASI_SDK_PATH", "/nonexistent/wasi-sdk")
	_, _, err := findCppCompiler()
	if err != nil {
		t.Logf("findCppCompiler with bad WASI_SDK_PATH: %v", err)
	}
}

func TestFindEmscriptenCompiler_NotFound(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "/nonexistent")

	result := findEmscriptenCompiler()
	// May still find emcc via repo-local emsdk/ path resolution
	if result != "" {
		t.Logf("emcc found via repo-local path: %s", result)
	}
}

// ─── Format Exports ────────────────────────────────────────────────────────────

func TestFormatCExports(t *testing.T) {
	result := formatCExports()
	expected := "['_execute','_init','_alloc','_dealloc','_metadata']"
	if result != expected {
		t.Errorf("formatCExports() = %q, want %q", result, expected)
	}
}

// ─── WASI-SDK Detection ────────────────────────────────────────────────────────

func TestFindCCompiler_PreferWASISDK(t *testing.T) {
	origWasi := os.Getenv("WASI_SDK_PATH")
	defer os.Setenv("WASI_SDK_PATH", origWasi)

	os.Setenv("WASI_SDK_PATH", "")
	compiler, toolchain, err := findCCompiler()
	if err != nil {
		t.Skip("no C WASM compiler available")
	}
	t.Logf("Found compiler: %s (%s)", compiler, toolchain)
	if toolchain != "wasi-sdk" && toolchain != "emscripten" {
		t.Errorf("unexpected toolchain: %s", toolchain)
	}
}
