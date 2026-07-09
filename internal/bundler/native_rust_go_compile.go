// native_rust_go_compile.go implements Rust and Go → WASM for fly bundle / wasm bundle.
// Do not name this file *_wasm.go: the Go toolchain treats that suffix as GOARCH=wasm-only sources.
package bundler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/functionfly/functionfly/internal/manifest"
)

// Rust WASM targets tried in order (wasip1 is current; wasi is legacy alias on some toolchains).
var cargoWasmTargets = []string{"wasm32-wasip1", "wasm32-wasi"}

const rustTempCrate = "functionfly_fn"

// bundleRustToWasm builds a Rust project to WebAssembly using cargo.
// If Cargo.toml exists in the working directory, it builds that crate; otherwise it creates a
// temporary cdylib project and uses the manifest entry file as src/lib.rs.
func bundleRustToWasm(m *manifest.Manifest) ([]byte, error) {
	if _, err := exec.LookPath("cargo"); err != nil {
		return nil, NewBundlerErrorWithCause("rust bundle", "cargo not on PATH; install Rust from https://rustup.rs", err)
	}

	if _, err := os.Stat("Cargo.toml"); err == nil {
		return cargoBuildReleaseWasmInDir(".")
	}

	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("rust bundle", "failed to read entry file", err)
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-rust-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("rust bundle", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cargoToml := fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[profile.release]
opt-level = "s"
lto = true
panic = "abort"
`, rustTempCrate)
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644); err != nil {
		return nil, NewBundlerErrorWithCause("rust bundle", "failed to write Cargo.toml", err)
	}
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return nil, NewBundlerErrorWithCause("rust bundle", "failed to create src", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "lib.rs"), src, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("rust bundle", "failed to write lib.rs", err)
	}

	return cargoBuildReleaseWasmInDir(tmpDir)
}

func cargoBuildReleaseWasmInDir(dir string) ([]byte, error) {
	var lastOut string
	var lastErr error
	for _, target := range cargoWasmTargets {
		cmd := exec.Command("cargo", "build", "--release", "--target", target)
		cmd.Dir = dir
		cmd.Env = os.Environ()
		out, err := cmd.CombinedOutput()
		if err != nil {
			lastErr = err
			lastOut = string(out)
			continue
		}
		wasm, pickErr := pickWasmFromCargoTarget(dir, target)
		if pickErr != nil {
			lastErr = pickErr
			lastOut = pickErr.Error()
			continue
		}
		return wasm, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("cargo wasm build failed")
	}
	return nil, NewCompilationErrorWithOutput("cargo", "wasm32-wasip1/wasm32-wasi", lastOut, lastErr)
}

func pickWasmFromCargoTarget(dir, target string) ([]byte, error) {
	relDir := filepath.Join(dir, "target", target, "release")
	matches, err := filepath.Glob(filepath.Join(relDir, "*.wasm"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no .wasm under %s", relDir)
	}
	sort.Strings(matches)
	return os.ReadFile(matches[0])
}

// bundleGoToWasm builds a Go main package to WASI WebAssembly (Go 1.26.5+).
// If go.mod exists in the working directory, it runs go build in place; otherwise
// it creates a temporary module and prepends a generated wrapper that provides
// `main()` and the host I/O contract.
func bundleGoToWasm(m *manifest.Manifest) ([]byte, error) {
	if err := requireGoVersion(); err != nil {
		return nil, NewBundlerErrorWithCause("go bundle", err.Error(), err)
	}

	outName := ".functionfly_gobuild.wasm"
	defer func() { _ = os.Remove(outName) }()

	if _, err := os.Stat("go.mod"); err == nil {
		wd, err := os.Getwd()
		if err != nil {
			return nil, NewBundlerErrorWithCause("go bundle", "getwd", err)
		}
		cmd := exec.Command("go", "build", "-o", outName, ".")
		cmd.Dir = wd
		cmd.Env = goWasip1Env()
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, NewCompilationErrorWithOutput("go", ".", string(out), err)
		}
		return os.ReadFile(outName)
	}

	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("go bundle", "failed to read entry file", err)
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-go-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("go bundle", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(generatedGoMod), 0644); err != nil {
		return nil, NewBundlerErrorWithCause("go bundle", "failed to write go.mod", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), src, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("go bundle", "failed to write main.go", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "zz_functionfly.go"), []byte(goWrapperSource), 0644); err != nil {
		return nil, NewBundlerErrorWithCause("go bundle", "failed to write wrapper", err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmpDir
	tidy.Env = goWasip1Env()
	if out, err := tidy.CombinedOutput(); err != nil {
		return nil, NewCompilationErrorWithOutput("go", "mod tidy", string(out), err)
	}

	outPath := filepath.Join(tmpDir, "out.wasm")
	build := exec.Command("go", "build", "-o", outPath, ".")
	build.Dir = tmpDir
	build.Env = goWasip1Env()
	if out, err := build.CombinedOutput(); err != nil {
		return nil, NewCompilationErrorWithOutput("go", "build wasip1/wasm", string(out), err)
	}
	return os.ReadFile(outPath)
}
