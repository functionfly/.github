package bundler

import (
	"os"
	"testing"
)

// ─── Toolchain Detector Tests ──────────────────────────────────────────────────

func TestDetectToolchains_ReturnsAll(t *testing.T) {
	t.Parallel()
	toolchains := DetectToolchains()

	expectedLangs := []string{"rust", "go", "c", "cpp", "ruby", "kotlin", "swift", "javascript", "python"}
	for _, lang := range expectedLangs {
		if _, ok := toolchains[lang]; !ok {
			t.Errorf("missing toolchain entry for %q", lang)
		}
	}
}

func TestDetectToolchains_HasNameAndLanguage(t *testing.T) {
	t.Parallel()
	toolchains := DetectToolchains()

	for lang, info := range toolchains {
		if info.Name == "" {
			t.Errorf("toolchain %q has empty Name", lang)
		}
		if info.Language == "" {
			t.Errorf("toolchain %q has empty Language", lang)
		}
	}
}

func TestDetectAvailableRuntimes(t *testing.T) {
	t.Parallel()
	available := DetectAvailableRuntimes()
	t.Logf("Available runtimes: %v", available)

	// At minimum, Go or Rust should be available on a dev machine
	hasGo := false
	hasRust := false
	for _, r := range available {
		if r == "go" {
			hasGo = true
		}
		if r == "rust" {
			hasRust = true
		}
	}

	if !hasGo && !hasRust {
		t.Log("Warning: neither Go nor Rust toolchains detected — this is unusual for a dev machine")
	}
}

// ─── Individual Detector Tests ─────────────────────────────────────────────────

func TestDetectRustToolchain(t *testing.T) {
	t.Parallel()
	info := detectRustToolchain()
	if info.Language != "rust" {
		t.Errorf("expected language 'rust', got %q", info.Language)
	}
	if info.Available {
		if info.Path == "" {
			t.Error("available toolchain should have a path")
		}
		t.Logf("Rust: %s at %s", info.Version, info.Path)
	} else {
		t.Log("Rust not available (cargo not on PATH)")
	}
}

func TestDetectGoToolchain(t *testing.T) {
	t.Parallel()
	info := detectGoToolchain()
	if info.Language != "go" {
		t.Errorf("expected language 'go', got %q", info.Language)
	}
	if info.Available {
		if info.Path == "" {
			t.Error("available toolchain should have a path")
		}
		t.Logf("Go: %s at %s", info.Version, info.Path)
	} else {
		t.Log("Go not available")
	}
}

func TestDetectCToolchain(t *testing.T) {
	t.Parallel()
	info := detectCToolchain()
	if info.Language != "c" {
		t.Errorf("expected language 'c', got %q", info.Language)
	}
	if info.Available {
		t.Logf("C: %s (%s) at %s", info.Toolchain, info.Version, info.Path)
		if info.Toolchain != "wasi-sdk" && info.Toolchain != "emscripten" {
			t.Errorf("unexpected toolchain: %s", info.Toolchain)
		}
	} else {
		t.Log("C WASM compiler not available")
	}
}

func TestDetectCppToolchain(t *testing.T) {
	t.Parallel()
	info := detectCppToolchain()
	if info.Language != "cpp" {
		t.Errorf("expected language 'cpp', got %q", info.Language)
	}
	t.Logf("C++ available: %v, toolchain: %s", info.Available, info.Toolchain)
}

func TestDetectRubyToolchain(t *testing.T) {
	t.Parallel()
	info := detectRubyToolchain()
	if info.Language != "ruby" {
		t.Errorf("expected language 'ruby', got %q", info.Language)
	}
	t.Logf("Ruby available: %v", info.Available)
}

func TestDetectKotlinToolchain(t *testing.T) {
	t.Parallel()
	info := detectKotlinToolchain()
	if info.Language != "kotlin" {
		t.Errorf("expected language 'kotlin', got %q", info.Language)
	}
	if info.Available {
		t.Logf("Kotlin: %s at %s", info.Version, info.Path)
	}
}

func TestDetectSwiftToolchain(t *testing.T) {
	t.Parallel()
	info := detectSwiftToolchain()
	if info.Language != "swift" {
		t.Errorf("expected language 'swift', got %q", info.Language)
	}
	if info.Available {
		t.Logf("Swift: %s (%s) at %s", info.Toolchain, info.Version, info.Path)
	}
}

func TestDetectJavaScriptToolchain(t *testing.T) {
	t.Parallel()
	info := detectJavaScriptToolchain()
	if info.Language != "javascript" {
		t.Errorf("expected language 'javascript', got %q", info.Language)
	}
	// JavaScript should always be available (bundled Javy)
	if !info.Available {
		t.Error("JavaScript toolchain should always be available (bundled)")
	}
	t.Logf("JavaScript: %s (%s)", info.Toolchain, info.Version)
}

func TestDetectPythonToolchain(t *testing.T) {
	t.Parallel()
	info := detectPythonToolchain()
	if info.Language != "python" {
		t.Errorf("expected language 'python', got %q", info.Language)
	}
	t.Logf("Python available: %v", info.Available)
}

// ─── Environment Variable Tests ────────────────────────────────────────────────

func TestDetectCToolchain_WithWASISDKEnv(t *testing.T) {
	origWasi := os.Getenv("WASI_SDK_PATH")
	defer os.Setenv("WASI_SDK_PATH", origWasi)

	os.Setenv("WASI_SDK_PATH", "/nonexistent/wasi-sdk")
	info := detectCToolchain()
	// Should not crash even with bad path
	t.Logf("C with bad WASI_SDK_PATH: available=%v", info.Available)
}

func TestDetectRubyToolchain_WithEnvPath(t *testing.T) {
	origMRuby := os.Getenv("MRUBY_WASM_PATH")
	defer os.Setenv("MRUBY_WASM_PATH", origMRuby)

	os.Setenv("MRUBY_WASM_PATH", "/nonexistent/mruby.wasm")
	info := detectRubyToolchain()
	if info.Available {
		t.Error("should not be available with non-existent path")
	}
}
