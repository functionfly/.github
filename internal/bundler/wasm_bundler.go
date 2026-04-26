package bundler

import (
	"github.com/functionfly/functionfly/internal/manifest"
)

// BundleForWasmRuntime bundles code for WebAssembly runtime execution
// Implements actual WebAssembly compilation using QuickJS (via Javy) for JavaScript
// and Pyodide/micropython for Python, with fallback to WAT templates
func BundleForWasmRuntime(manifest *manifest.Manifest) ([]byte, error) {
	return BundleForWasmRuntimeWithWorkingDirectory(manifest, "")
}

// BundleForWasmRuntimeWithWorkingDirectory bundles code for WebAssembly runtime execution
// with explicit working directory support for consistent path resolution.
// Dependencies are automatically installed if specified in the manifest.
func BundleForWasmRuntimeWithWorkingDirectory(manifest *manifest.Manifest, workingDir string) ([]byte, error) {
	if manifest == nil {
		return nil, NewBundlerError("wasm bundle", "manifest cannot be nil")
	}

	// Resolve and validate working directory
	resolvedDir, err := ResolveWorkingDirectory(workingDir)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm bundle", "failed to resolve working directory", err)
	}

	// Execute bundling within the specified working directory
	var result []byte
	err = WithWorkingDirectory(resolvedDir, func() error {
		// Install dependencies first
		if depErr := InstallDependencies(manifest); depErr != nil {
			return NewBundlerErrorWithCause("wasm bundle", "dependency installation failed", depErr)
		}

		var bundleErr error
		switch manifest.Runtime {
		case "node18", "node20", "deno", "bun":
			result, bundleErr = bundleJSForWasmRuntime(manifest)
		case "python3.11", "python3.12", "python":
			result, bundleErr = bundlePythonForWasmRuntime(manifest)
		case "rust":
			result, bundleErr = bundleRustToWasm(manifest)
		case "go", "go1.21":
			result, bundleErr = bundleGoToWasm(manifest)
		case "c", "c11":
			result, bundleErr = bundleCToWasm(manifest)
		case "cpp", "cpp17", "c++":
			result, bundleErr = bundleCppToWasm(manifest)
		case "ruby", "ruby3.3":
			result, bundleErr = bundleRubyForWasmRuntime(manifest)
		case "kotlin", "kotlin1.9":
			result, bundleErr = bundleKotlinForWasmRuntime(manifest)
		case "swift", "swift5.9":
			result, bundleErr = bundleSwiftForWasmRuntime(manifest)
		default:
			result, bundleErr = bundleJSForWasmRuntime(manifest) // Default to JS
		}
		return bundleErr
	})

	return result, err
}
