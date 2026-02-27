package bundler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/functionfly/functionfly/internal/manifest"
)

// bundleJSForWasmRuntime bundles JavaScript/TypeScript for Wasm runtime execution
// Attempts actual WebAssembly compilation using Javy (QuickJS-based), falls back to wrapper
func bundleJSForWasmRuntime(manifest *manifest.Manifest) ([]byte, error) {
	// Read and validate entry file using shared helper
	entryFile, sourceCode, err := ReadEntryFile(manifest)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm js bundle", "failed to read entry file", err)
	}

	// Try actual WebAssembly compilation using Javy
	if wasmBytes, err := compileJSToWasm(entryFile, manifest); err == nil {
		// Validate the compiled WebAssembly
		if err := validateWasmModule(wasmBytes); err != nil {
			fmt.Printf("Warning: Compiled WebAssembly validation failed (%v), using fallback\n", err)
			return createFallbackWasmWrapper(string(sourceCode), manifest, "javascript")
		}
		fmt.Printf("Successfully compiled %s to WebAssembly using Javy\n", entryFile)
		return wasmBytes, nil
	} else {
		fmt.Printf("Warning: WebAssembly compilation failed (%v), using fallback wrapper\n", err)
	}

	// Fallback: Create a JavaScript wrapper for Wasm runtime
	return createFallbackWasmWrapper(string(sourceCode), manifest, "javascript")
}

// compileJSToWasm attempts to compile JavaScript to WebAssembly using Javy
func compileJSToWasm(entryFile string, manifest *manifest.Manifest) ([]byte, error) {
	// Check if Javy is available
	if _, err := exec.LookPath("javy"); err != nil {
		return nil, NewBundlerError("wasm js compile", "javy not found in PATH. Install with: npm install -g @shopify/javy")
	}

	// Create temporary output file with unique name to avoid conflicts
	tempDir := os.TempDir()
	tempOut := filepath.Join(tempDir, fmt.Sprintf("functionfly-js-%d.wasm", os.Getpid()))
	defer os.Remove(tempOut)

	// Source content is already validated by ReadEntryFile, but read again for compilation
	sourceContent, err := os.ReadFile(entryFile)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm js compile", "failed to read source file for compilation", err)
	}

	// Basic validation - check for minimal function export
	sourceStr := string(sourceContent)
	hasExport := strings.Contains(sourceStr, "export") ||
		strings.Contains(sourceStr, "module.exports") ||
		strings.Contains(sourceStr, "exports.")

	if !hasExport {
		return nil, NewBundlerError("wasm js compile", "JavaScript file must export a function (use 'export default' or 'module.exports')")
	}

	// Build Javy command with optimized settings
	args := []string{
		"compile",
		entryFile,
		"-o", tempOut,
		"--dynamic", // Enable dynamic linking for better compatibility
	}

	// Add TypeScript support if needed
	if strings.HasSuffix(entryFile, ".ts") || strings.HasSuffix(entryFile, ".tsx") {
		args = append(args, "--typescript")
	}

	// Set working directory to the source file directory for relative imports
	workDir := filepath.Dir(entryFile)
	cmd := exec.Command("javy", args...)
	cmd.Dir = workDir

	// Execute compilation
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("javy", entryFile, string(output), err)
	}

	// Verify the output file was created and has content
	if _, err := os.Stat(tempOut); os.IsNotExist(err) {
		return nil, NewBundlerError("wasm js compile", "javy compilation succeeded but output file was not created")
	}

	// Read the compiled Wasm output
	wasmBytes, err := os.ReadFile(tempOut)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm js compile", "failed to read compiled Wasm", err)
	}

	if len(wasmBytes) == 0 {
		return nil, NewBundlerError("wasm js compile", "compiled Wasm file is empty")
	}

	// Validate the compiled WebAssembly
	if err := validateWasmModule(wasmBytes); err != nil {
		return nil, fmt.Errorf("compiled output validation failed: %v", err)
	}

	return wasmBytes, nil
}

// createJSWasmWrapper creates a JavaScript wrapper for WASM runtime execution
func createJSWasmWrapper(entryFile string, manifest *manifest.Manifest) ([]byte, error) {
	// Read the source code
	sourceCode, err := os.ReadFile(entryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry file: %v", err)
	}

	// Create a Wasm-compatible wrapper
	// Note: This is a JavaScript wrapper, not actual WebAssembly
	wasmWrapper := fmt.Sprintf(`
// FunctionFly Wasm Wrapper for %s
const sourceCode = %q;

// Simple execution environment
globalThis.console = {
  log: (...args) => {
    // Send to stdout
    const message = args.join(' ');
    // Wasm host will capture this
  },
  error: (...args) => {
    const message = args.join(' ');
    // Wasm host will capture this
  }
};

// Execute the source code
try {
  const exports = {};
  const module = { exports };

  // Simple CommonJS-style require (mock)
  const require = (name) => {
    throw new Error('Module ' + name + ' not available in Wasm runtime');
  };

  // Execute the code
  const func = new Function('exports', 'require', 'module', 'globalThis', sourceCode);
  func(exports, require, module, globalThis);

  // Export the default export or main function
  if (module.exports.default) {
    globalThis.main = module.exports.default;
  } else if (typeof module.exports === 'function') {
    globalThis.main = module.exports;
  } else {
    globalThis.main = () => {
      return JSON.stringify(module.exports);
    };
  }
} catch (error) {
  globalThis.main = () => {
    throw error;
  };
}

// Wasm entry point
export function _start() {
  // Wasm initialization
}

export function execute(input) {
  try {
    const result = globalThis.main(input);
    return result || input; // Fallback to input if no result
  } catch (error) {
    throw new Error('Function execution failed: ' + error.message);
  }
}
`, entryFile, string(sourceCode))

	// For now, return the JavaScript code as bytes
	// In a real implementation, this would be compiled to Wasm
	return []byte(wasmWrapper), nil
}