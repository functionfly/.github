// swift_wasm_compiler.go implements Swift → WASM via SwiftWasm toolchain.
package bundler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
)

const swiftBuildTimeout = 120 * time.Second

// bundleSwiftForWasmRuntime compiles Swift source to WASM using SwiftWasm.
func bundleSwiftForWasmRuntime(m *manifest.Manifest) ([]byte, error) {
	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("swift bundle", "failed to read entry file", err)
	}

	// Find Swift WASM compiler
	swiftc, toolchain, err := findSwiftWasmCompiler()
	if err != nil {
		return nil, NewBundlerErrorWithCause("swift bundle", "no Swift WASM compiler available", err)
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-swift-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("swift bundle", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	switch toolchain {
	case "carton":
		return compileWithCarton(tmpDir, src, m)
	case "swiftwasm":
		return compileWithSwiftWasm(swiftc, tmpDir, src, m)
	default:
		return nil, NewBundlerError("swift bundle", fmt.Sprintf("unsupported toolchain: %s", toolchain))
	}
}

// findSwiftWasmCompiler locates a Swift WASM compiler.
func findSwiftWasmCompiler() (compilerPath, toolchain string, err error) {
	// Try carton (SwiftWasm build tool)
	if carton, lookErr := exec.LookPath("carton"); lookErr == nil {
		return carton, "carton", nil
	}

	// Try swiftc with WASM target support
	if swiftc, lookErr := exec.LookPath("swiftc"); lookErr == nil {
		// Check if swiftc supports wasm32 target
		cmd := exec.Command(swiftc, "-version")
		out, err := cmd.CombinedOutput()
		if err == nil {
			output := string(out)
			// SwiftWasm toolchain reports version with wasm support
			if contains(output, "wasm") || contains(output, "SwiftWasm") {
				return swiftc, "swiftwasm", nil
			}
			// Standard Swift might still support wasm32-unknown-wasi
			return swiftc, "swiftwasm", nil
		}
	}

	return "", "", fmt.Errorf("Swift WASM compiler not found; install SwiftWasm (https://swiftwasm.org) or carton (brew install swiftwasm/swiftwasm/carton)")
}

// compileWithCarton uses carton to build a Swift package to WASM.
func compileWithCarton(tmpDir string, src []byte, m *manifest.Manifest) ([]byte, error) {
	if err := createSwiftPackage(tmpDir, src, m); err != nil {
		return nil, NewBundlerErrorWithCause("carton", "failed to create Swift package", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), swiftBuildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "carton", "bundle", "--product", m.Name)
	cmd.Dir = tmpDir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("carton", "swift wasm bundle", string(out), err)
	}

	return findSwiftWasmOutput(tmpDir)
}

// compileWithSwiftWasm uses swiftc directly with WASM target.
func compileWithSwiftWasm(swiftc, tmpDir string, src []byte, m *manifest.Manifest) ([]byte, error) {
	srcFile := filepath.Join(tmpDir, "main.swift")
	if err := os.WriteFile(srcFile, src, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("swift bundle", "failed to write source", err)
	}

	outFile := filepath.Join(tmpDir, "function.wasm")

	ctx, cancel := context.WithTimeout(context.Background(), swiftBuildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, swiftc,
		srcFile,
		"-target", "wasm32-unknown-wasi",
		"-o", outFile,
		"-Osize",
		"-wmo",
		"-Xlinker", "--export-all",
	)
	cmd.Dir = tmpDir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("swiftc", srcFile, string(out), err)
	}

	wasm, err := os.ReadFile(outFile)
	if err != nil {
		return nil, NewBundlerErrorWithCause("swift bundle", "failed to read output WASM", err)
	}

	if valErr := validateWasmModule(wasm); valErr != nil {
		return nil, NewBundlerErrorWithCause("swift bundle", "output WASM validation failed", valErr)
	}

	return wasm, nil
}

// createSwiftPackage scaffolds a Swift package for WASM compilation.
func createSwiftPackage(dir string, src []byte, m *manifest.Manifest) error {
	// Package.swift
	pkgSwift := fmt.Sprintf(`// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "%s",
    targets: [
        .executableTarget(
            name: "%s",
            path: "Sources",
            swiftSettings: [
                .unsafeFlags(["-Xfrontend", "-disable-reflection-metadata"])
            ]
        )
    ]
)
`, m.Name, m.Name)

	if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte(pkgSwift), 0644); err != nil {
		return err
	}

	srcDir := filepath.Join(dir, "Sources")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(srcDir, "main.swift"), src, 0644)
}

// findSwiftWasmOutput locates the compiled .wasm file.
func findSwiftWasmOutput(projectDir string) ([]byte, error) {
	var found []byte
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".wasm" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if len(data) >= 8 && data[0] == 0x00 && data[1] == 0x61 && data[2] == 0x73 && data[3] == 0x6D {
			found = data
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found != nil {
		return found, nil
	}
	return nil, fmt.Errorf("no .wasm file found in Swift build output")
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
