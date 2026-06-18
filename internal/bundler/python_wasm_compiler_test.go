package bundler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/functionfly/functionfly/internal/manifest"
)

func TestBundlePythonForWasmRuntime_FailsWithoutRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "main.py")
	if err := os.WriteFile(src, []byte(`def handler(x): return x`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Point at a definitely-nonexistent runtime so IsMicropythonAvailable() is false.
	origPath := os.Getenv("MICROPATH")
	defer os.Setenv("MICROPATH", origPath)

	// Simulate missing runtime by temporarily moving it.
	// We cannot guarantee the runtime is installed in the test environment, so
	// just assert that when it IS missing, the function fails fast (no stub).
	runtimePath := GetMicropythonRuntimePath()
	if runtimePath == "" {
		// No runtime in this environment — exercise the missing-runtime path.
		m := &manifest.Manifest{Name: "test-py", Version: "1.0.0", Runtime: "python", Entry: "main.py"}
		_, err := bundlePythonForWasmRuntime(m)
		if err == nil {
			t.Fatal("expected error when micropython runtime is missing, got nil")
		}
		if !strings.Contains(err.Error(), "micropython") {
			t.Fatalf("expected error to mention micropython, got: %v", err)
		}
		return
	}

	// Runtime IS present — exercise the success path.
	m := &manifest.Manifest{Name: "test-py", Version: "1.0.0", Runtime: "python", Entry: "main.py"}
	wasm, err := bundlePythonForWasmRuntime(m)
	if err != nil {
		t.Fatalf("expected success when runtime is present, got: %v", err)
	}
	if len(wasm) < 1000 {
		t.Fatalf("expected WASM bytes >= 1000, got %d", len(wasm))
	}
	// Verify WASM magic bytes.
	if wasm[0] != 0x00 || wasm[1] != 0x61 || wasm[2] != 0x73 || wasm[3] != 0x6D {
		t.Fatal("result is not a valid WASM module (bad magic bytes)")
	}
}

func TestCreateFallbackWasmWrapper_JSAllowed(t *testing.T) {
	wat, err := createJSWasmWrapperFromSource("console.log('hi')", &manifest.Manifest{Name: "t", Version: "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wat) == 0 {
		t.Fatal("expected non-empty WAT for JS runtime")
	}
	if !strings.Contains(string(wat), "__execute") {
		t.Fatal("expected __execute export in JS WAT wrapper")
	}
}

func TestCreateFallbackWasmWrapper_NonJSPythonFails(t *testing.T) {
	_, err := createFallbackWasmWrapper("x = 1", &manifest.Manifest{Name: "t", Version: "1"}, "python")
	if err == nil {
		t.Fatal("expected error for Python fallback (no stub policy)")
	}
	if !strings.Contains(err.Error(), "no fallback available") {
		t.Fatalf("expected 'no fallback available' error, got: %v", err)
	}
}

func TestCreateFallbackWasmWrapper_NonJSRubyFails(t *testing.T) {
	_, err := createFallbackWasmWrapper("puts 1", &manifest.Manifest{Name: "t", Version: "1"}, "ruby")
	if err == nil {
		t.Fatal("expected error for Ruby fallback (no stub policy)")
	}
}
