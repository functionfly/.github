// c_wasm_compiler.go implements C and C++ → WASM compilation via Emscripten or WASI-SDK.
package bundler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
)

// FunctionFlyWASMExports are the standard C exports expected by the WASM runtime.
var functionFlyCExports = []string{
	"_execute",
	"_init",
	"_alloc",
	"_dealloc",
	"_metadata",
}

// bundleCToWasm compiles a C source file to WebAssembly using Emscripten or WASI-SDK.
func bundleCToWasm(m *manifest.Manifest) ([]byte, error) {
	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("c bundle", "failed to read entry file", err)
	}

	compiler, toolchain, err := findCCompiler()
	if err != nil {
		return nil, NewBundlerErrorWithCause("c bundle", "no C WASM compiler available", err)
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-c-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("c bundle", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcFile := filepath.Join(tmpDir, "main.c")
	if err := os.WriteFile(srcFile, src, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("c bundle", "failed to write source", err)
	}

	outFile := filepath.Join(tmpDir, "function.wasm")

	switch toolchain {
	case "emscripten":
		return compileWithEmscripten(compiler, srcFile, outFile, tmpDir, false)
	case "wasi-sdk":
		return compileWithWASISDK(compiler, srcFile, outFile, tmpDir)
	default:
		return nil, NewBundlerError("c bundle", fmt.Sprintf("unsupported toolchain: %s", toolchain))
	}
}

// bundleCppToWasm compiles a C++ source file to WebAssembly.
func bundleCppToWasm(m *manifest.Manifest) ([]byte, error) {
	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("cpp bundle", "failed to read entry file", err)
	}

	compiler, toolchain, err := findCppCompiler()
	if err != nil {
		return nil, NewBundlerErrorWithCause("cpp bundle", "no C++ WASM compiler available", err)
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-cpp-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("cpp bundle", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcFile := filepath.Join(tmpDir, "main.cpp")
	if err := os.WriteFile(srcFile, src, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("cpp bundle", "failed to write source", err)
	}

	outFile := filepath.Join(tmpDir, "function.wasm")

	switch toolchain {
	case "emscripten":
		return compileWithEmscripten(compiler, srcFile, outFile, tmpDir, true)
	case "wasi-sdk":
		return compileWithWASISDK(compiler, srcFile, outFile, tmpDir)
	default:
		return nil, NewBundlerError("cpp bundle", fmt.Sprintf("unsupported toolchain: %s", toolchain))
	}
}

// findCCompiler locates a C-to-WASM compiler. Prefers WASI-SDK (lighter) then falls back to Emscripten.
func findCCompiler() (compilerPath, toolchain string, err error) {
	if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
		clang := filepath.Join(sdkPath, "bin", "clang")
		if _, statErr := os.Stat(clang); statErr == nil {
			return clang, "wasi-sdk", nil
		}
	}

	for _, p := range []string{
		"/opt/wasi-sdk/bin/clang",
		"/usr/local/opt/wasi-sdk/bin/clang",
	} {
		if _, statErr := os.Stat(p); statErr == nil {
			return p, "wasi-sdk", nil
		}
	}

	if emcc := findEmscriptenCompiler(); emcc != "" {
		return emcc, "emscripten", nil
	}

	return "", "", fmt.Errorf("no C WASM compiler found; install WASI-SDK or Emscripten (emsdk/)")
}

// findCppCompiler locates a C++-to-WASM compiler.
func findCppCompiler() (compilerPath, toolchain string, err error) {
	if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
		clangxx := filepath.Join(sdkPath, "bin", "clang++")
		if _, statErr := os.Stat(clangxx); statErr == nil {
			return clangxx, "wasi-sdk", nil
		}
	}

	for _, p := range []string{
		"/opt/wasi-sdk/bin/clang++",
		"/usr/local/opt/wasi-sdk/bin/clang++",
	} {
		if _, statErr := os.Stat(p); statErr == nil {
			return p, "wasi-sdk", nil
		}
	}

	if emcc := findEmscriptenCompiler(); emcc != "" {
		return emcc, "emscripten", nil
	}

	return "", "", fmt.Errorf("no C++ WASM compiler found; install WASI-SDK or Emscripten (emsdk/)")
}

// findEmscriptenCompiler looks for emcc on PATH or in the repo's emsdk/ directory.
func findEmscriptenCompiler() string {
	if p, err := exec.LookPath("emcc"); err == nil {
		return p
	}

	homeDir, _ := os.Getwd()
	emsdkPaths := []string{
		filepath.Join(homeDir, "emsdk"),
		filepath.Join(homeDir, "..", "emsdk"),
		filepath.Join(homeDir, "..", "..", "emsdk"),
	}

	for _, emsdkDir := range emsdkPaths {
		upstream := filepath.Join(emsdkDir, "upstream")
		entries, err := os.ReadDir(upstream)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "emscripten") {
				emcc := filepath.Join(upstream, entry.Name(), "emcc")
				if _, statErr := os.Stat(emcc); statErr == nil {
					return emcc
				}
			}
		}
	}

	return ""
}

// compileWithEmscripten compiles source to WASM using emcc.
func compileWithEmscripten(emcc, srcFile, outFile, tmpDir string, isCpp bool) ([]byte, error) {
	args := []string{
		srcFile,
		"-o", outFile,
		"-O2",
		"-s", "WASM=1",
		"-s", "STANDALONE_WASM=1",
		"-s", fmt.Sprintf("EXPORTED_FUNCTIONS=%s", formatCExports()),
		"-s", "EXPORTED_RUNTIME_METHODS=['ccall','cwrap']",
		"-s", "ALLOW_MEMORY_GROWTH=1",
		"-s", "INITIAL_MEMORY=1048576",
		"--no-entry",
	}

	if isCpp {
		args = append([]string{"-x", "c++"}, args...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, emcc, args...)
	cmd.Dir = tmpDir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("emcc", srcFile, string(out), err)
	}

	wasm, err := os.ReadFile(outFile)
	if err != nil {
		return nil, NewBundlerErrorWithCause("emscripten", "failed to read output WASM", err)
	}

	if valErr := validateWasmModule(wasm); valErr != nil {
		return nil, NewBundlerErrorWithCause("emscripten", "output WASM validation failed", valErr)
	}

	return wasm, nil
}

// compileWithWASISDK compiles source to WASM using WASI-SDK clang.
func compileWithWASISDK(clang, srcFile, outFile, tmpDir string) ([]byte, error) {
	sdkDir := filepath.Dir(filepath.Dir(clang))
	sysroot := filepath.Join(sdkDir, "share", "wasi-sysroot")

	args := []string{
		srcFile,
		"-o", outFile,
		"--target=wasm32-wasi",
		fmt.Sprintf("--sysroot=%s", sysroot),
		"-O2",
		"-nostartfiles",
		"-Wl,--no-entry",
		"-Wl,--export-all",
		"-Wl,--allow-undefined",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, clang, args...)
	cmd.Dir = tmpDir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("wasi-sdk clang", srcFile, string(out), err)
	}

	wasm, err := os.ReadFile(outFile)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasi-sdk", "failed to read output WASM", err)
	}

	if valErr := validateWasmModule(wasm); valErr != nil {
		return nil, NewBundlerErrorWithCause("wasi-sdk", "output WASM validation failed", valErr)
	}

	return wasm, nil
}

// formatCExports formats the C exports list for Emscripten's EXPORTED_FUNCTIONS.
func formatCExports() string {
	parts := make([]string, len(functionFlyCExports))
	for i, e := range functionFlyCExports {
		parts[i] = fmt.Sprintf("'%s'", e)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}
