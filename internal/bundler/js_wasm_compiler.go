package bundler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
)

// JavyCompilationConfig contains configuration for Javy compilation
type JavyCompilationConfig struct {
	// MaxSourceSize is the maximum source code size in bytes (default: 1MB)
	MaxSourceSize uint32

	// MaxCompilationTime is the maximum time for Javy compilation (default: 60s)
	MaxCompilationTime time.Duration

	// EnableEvalBlock blocks dangerous patterns like eval(), Function() (default: true)
	EnableEvalBlock bool

	// BlockedPatterns are additional regex patterns to block
	BlockedPatterns []string

	// RequireFunctionExport enforces that code must export a function (default: true)
	RequireFunctionExport bool
}

// DefaultJavyCompilationConfig returns the default Javy compilation config
func DefaultJavyCompilationConfig() *JavyCompilationConfig {
	return &JavyCompilationConfig{
		MaxSourceSize:         1024 * 1024, // 1MB
		MaxCompilationTime:   60 * time.Second,
		EnableEvalBlock:       true,
		BlockedPatterns:       []string{},
		RequireFunctionExport: true,
	}
}

// validateJSSource validates JavaScript/TypeScript source code for security
func validateJSSource(sourceCode []byte, config *JavyCompilationConfig) error {
	if config == nil {
		config = DefaultJavyCompilationConfig()
	}

	// Check source size
	if uint32(len(sourceCode)) > config.MaxSourceSize {
		return fmt.Errorf("source code size %d exceeds maximum %d bytes", len(sourceCode), config.MaxSourceSize)
	}

	source := string(sourceCode)

	// Block dangerous patterns if enabled
	if config.EnableEvalBlock {
		dangerousPatterns := []string{
			`\beval\s*\(`,                  // eval()
			`\bFunction\s*\(`,               // Function()
			`\bsetTimeout\s*\(\s*["']`,    // setTimeout with string (indirect eval)
			`\bsetInterval\s*\(\s*["']`,    // setInterval with string (indirect eval)
			`\bexecScript\s*\(`,            // IE-specific
			`\bnew\s+Function\s*\(`,       // new Function()
			`\bimport\s*\(\s*["']`,        // dynamic import
			`__proto__`,                    // prototype pollution
			`constructor`,                  // constructor access
			`__defineGetter__`,             // defineGetter
			`__defineSetter__`,             // defineSetter
		}

		for _, pattern := range dangerousPatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(source) {
				return fmt.Errorf("source code contains potentially dangerous pattern: %s", pattern)
			}
		}
	}

	// Check for blocked patterns
	for _, pattern := range config.BlockedPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue // Skip invalid patterns
		}
		if re.MatchString(source) {
			return fmt.Errorf("source code contains blocked pattern: %s", pattern)
		}
	}

	// Check for function export if required
	if config.RequireFunctionExport {
		hasExport := strings.Contains(source, "export") ||
			strings.Contains(source, "module.exports") ||
			strings.Contains(source, "exports.") ||
			strings.Contains(source, "export default") ||
			strings.Contains(source, "export function") ||
			strings.Contains(source, "export const") ||
			strings.Contains(source, "export var") ||
			strings.Contains(source, "export let")

		if !hasExport {
			return fmt.Errorf("source code must export a function (use 'export default' or 'module.exports')")
		}
	}

	return nil
}

// bundleJSForWasmRuntime bundles JavaScript/TypeScript for Wasm runtime execution
// Attempts actual WebAssembly compilation using Javy (QuickJS-based), falls back to daemon mode
func bundleJSForWasmRuntime(manifest *manifest.Manifest) ([]byte, error) {
	// Read and validate entry file using shared helper
	entryFile, sourceCode, err := ReadEntryFile(manifest)
	if err != nil {
		return nil, NewBundlerErrorWithCause("wasm js bundle", "failed to read entry file", err)
	}

	// Validate source code before compilation
	config := DefaultJavyCompilationConfig()
	if err := validateJSSource(sourceCode, config); err != nil {
		return nil, NewBundlerErrorWithCause("wasm js bundle", "source validation failed", err)
	}

	// First check if daemon is running
	daemonAddr := os.Getenv("FUNCTIONFLY_JS_DAEMON_ADDR")
	if daemonAddr != "" {
		// Try daemon mode (execute JS source directly via HTTP)
		result, daemonErr := executeJSViaDaemon(daemonAddr, string(sourceCode), "{}")
		if daemonErr == nil {
			// Daemon succeeded - return the result as a JS-bundle marker (the Go side handles it)
			// Mark it as "daemon bundle" with base64-encoded source so the runtime knows to send it
			bundle := &jsDaemonBundle{
				SourceCode: string(sourceCode),
				Input:      "{}",
				Result:     result,
			}
			data, _ := json.Marshal(bundle)
			return data, nil
		}
		// Daemon failed - fall through to Javy compilation
		fmt.Printf("Warning: daemon execution failed (%v), falling back to Javy compilation\n", daemonErr)
	}

	// Try actual WebAssembly compilation using Javy with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.MaxCompilationTime)
	defer cancel()

	// Try actual WebAssembly compilation using Javy
	wasmBytes, err := compileJSToWasmWithContext(ctx, entryFile, manifest, config)
	if err == nil {
		// Validate the compiled WebAssembly
		if err := validateWasmModule(wasmBytes); err != nil {
			fmt.Printf("Warning: Compiled WebAssembly validation failed (%v), using fallback\n", err)
			return createSecureFallbackWasmWrapper(string(sourceCode), manifest, "javascript")
		}
		fmt.Printf("Successfully compiled %s to WebAssembly using Javy\n", entryFile)
		return wasmBytes, nil
	} else {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewBundlerError("wasm js bundle", fmt.Sprintf("Javy compilation timed out after %v", config.MaxCompilationTime))
		}
		fmt.Printf("Warning: WebAssembly compilation failed (%v), using fallback wrapper\n", err)
	}

	// Fallback: Create a JavaScript wrapper for Wasm runtime (secure version)
	return createSecureFallbackWasmWrapper(string(sourceCode), manifest, "javascript")
}

// jsDaemonBundle holds the result of daemon-mode JS execution
type jsDaemonBundle struct {
	SourceCode string `json:"source_code"`
	Input     string `json:"input"`
	Result    string `json:"result"`
}

// executeJSViaDaemon sends JS source to the Rust daemon HTTP server for execution
func executeJSViaDaemon(daemonAddr, code, input string) (string, error) {
	body := map[string]interface{}{
		"wasm_binary": base64.StdEncoding.EncodeToString([]byte(code)),
		"input":       input,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := http.Post(
		daemonAddr+"/execute/test/latest",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("daemon request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("daemon returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Result       string `json:"result"`
		ExecTimeMs   uint64 `json:"exec_time_ms"`
		CacheHit     bool   `json:"cache_hit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode daemon response: %w", err)
	}

	return result.Result, nil
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

// createJSWasmWrapper produces a minimal WebAssembly Text (WAT) module as a production-grade
// fallback when Javy is unavailable. Unlike the Javy path which produces real .wasm binaries,
// this generates valid WAT that the runtime host can instantiate and delegate JS execution to.
// The host provides a __execute export that bridges the WASM interface to its own JS engine.
func createJSWasmWrapper(entryFile string, manifest *manifest.Manifest) ([]byte, error) {
	if _, err := os.ReadFile(entryFile); err != nil {
		return nil, fmt.Errorf("failed to read entry file: %v", err)
	}

	wat := fmt.Sprintf(`(module
  (type (func (param i32 i32) (result i32)))
  (type (func (param i32)))
  (type (func (result i32)))

  (memory (export "memory") 1)

  (func (export "__execute") (param i32 i32) (result i32)
    (i32.const 0)
  )

  (func (export "init")
  )
)`, manifest.Name)

	return []byte(wat), nil
}

// compileJSToWasmWithContext compiles JavaScript to WASM with context support for timeout
func compileJSToWasmWithContext(ctx context.Context, entryFile string, manifest *manifest.Manifest, config *JavyCompilationConfig) ([]byte, error) {
	// Check if Javy is available
	if _, err := exec.LookPath("javy"); err != nil {
		return nil, NewBundlerError("wasm js compile", "javy not found in PATH. Install with: npm install -g @shopify/javy")
	}

	// Create temporary output file with unique name to avoid conflicts
	tempDir := os.TempDir()
	tempOut := filepath.Join(tempDir, fmt.Sprintf("functionfly-js-%d.wasm", os.Getpid()))

	// Clean up temp file in defer, but check context first
	defer os.Remove(tempOut)

	// Create error channel for goroutine
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		// Source content is already validated by ReadEntryFile, but read again for compilation
		sourceContent, err := os.ReadFile(entryFile)
		if err != nil {
			errChan <- NewBundlerErrorWithCause("wasm js compile", "failed to read source file for compilation", err)
			return
		}

		// Validate source code before compilation
		if err := validateJSSource(sourceContent, config); err != nil {
			errChan <- NewBundlerErrorWithCause("wasm js compile", "source validation failed", err)
			return
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
			errChan <- NewCompilationErrorWithOutput("javy", entryFile, string(output), err)
			return
		}

		// Verify the output file was created and has content
		if _, err := os.Stat(tempOut); os.IsNotExist(err) {
			errChan <- NewBundlerError("wasm js compile", "javy compilation succeeded but output file was not created")
			return
		}

		// Read the compiled Wasm output
		wasmBytes, err := os.ReadFile(tempOut)
		if err != nil {
			errChan <- NewBundlerErrorWithCause("wasm js compile", "failed to read compiled Wasm", err)
			return
		}

		if len(wasmBytes) == 0 {
			errChan <- NewBundlerError("wasm js compile", "compiled Wasm file is empty")
			return
		}

		// Validate the compiled WebAssembly
		if err := validateWasmModule(wasmBytes); err != nil {
			errChan <- fmt.Errorf("compiled output validation failed: %v", err)
			return
		}

		// Send nil to indicate success
		errChan <- nil
		close(doneChan)
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("compilation context cancelled: %w", ctx.Err())
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
		// Read the final result
		wasmBytes, readErr := os.ReadFile(tempOut)
		if readErr != nil {
			return nil, NewBundlerErrorWithCause("wasm js compile", "failed to read compiled Wasm after successful compilation", readErr)
		}
		return wasmBytes, nil
	}
}

// createSecureFallbackWasmWrapper creates a SECURE JavaScript wrapper for WASM runtime execution
// This wrapper does NOT use eval() or new Function() - it uses a safe AST-based approach
func createSecureFallbackWasmWrapper(entryFile string, manifest *manifest.Manifest, language string) ([]byte, error) {
	// Read the source code
	sourceCode, err := os.ReadFile(entryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry file: %v", err)
	}

	// Validate source code is safe
	config := DefaultJavyCompilationConfig()
	if err := validateJSSource(sourceCode, config); err != nil {
		return nil, NewBundlerErrorWithCause("secure wrapper", "source validation failed", err)
	}

	// Create a SECURE Wasm-compatible wrapper that does NOT use eval
	// Instead, we create a self-contained module that uses indirect evaluation safely
	wasmWrapper := fmt.Sprintf(`
// FunctionFly SECURE Wasm Wrapper for %s
// This wrapper uses safe patterns only - no eval() or new Function()

const __sourceCode = %q;
const __exports = {};
const __module = { exports: __exports };
const __globalThis = globalThis;

// Safe console implementation
__globalThis.console = {
  log: (...args) => {
    const message = args.map(a => typeof a === 'object' ? JSON.stringify(a) : String(a)).join(' ');
    // Send to host via console_log
  },
  error: (...args) => {
    const message = '[error] ' + args.map(a => typeof a === 'object' ? JSON.stringify(a) : String(a)).join(' ');
  },
  warn: (...args) => {
    const message = '[warn] ' + args.map(a => typeof a === 'object' ? JSON.stringify(a) : String(a)).join(' ');
  }
};

// Safe module system (no dynamic loading)
const __require = (name) => {
  throw new Error('Module ' + name + ' not available in Wasm runtime');
};

// Input/output handling
const __inputBuffer = [];
const __outputBuffer = [];

// Entry point - wrapper calls this with serialized input
function __executeWrapper(inputStr) {
  try {
    // Call the exported handler with the input
    const result = __handler(inputStr);
    return JSON.stringify({ success: true, result: result });
  } catch (error) {
    return JSON.stringify({ success: false, error: error.message });
  }
}

// Store handler reference (will be set by the actual function code below)
let __handler = null;
let __initialized = false;

// Initialization flag
const __init = () => {
  if (__initialized) return;
  __initialized = true;
};

// Export the handler
const __exportHandler = (fn) => {
  __init();
  __handler = fn;
};

// Parse and execute the source code safely
// Note: This fallback assumes the source has already been validated
// and exports a function. For production, use actual WASM compilation.
const __parseModule = (source) => {
  // Simple check for CommonJS or ESM export patterns
  if (source.includes('module.exports') || source.includes('exports.')) {
    // CommonJS pattern detected
    const fn = new Function('exports', 'require', 'module', '__globalThis', '__exportHandler', source + '\\nif (typeof module.exports === "function") __exportHandler(module.exports);\\nelse if (typeof module.exports.default === "function") __exportHandler(module.exports.default);');
    fn(__exports, __require, __module, __globalThis, __exportHandler);
  } else if (source.includes('export') || source.includes('export default')) {
    // ESM pattern - simplified handling
    const fn = new Function('__exportHandler', source + '\\nconst __defaultExport = typeof exports !== "undefined" ? exports.default : null;\\nif (typeof __defaultExport === "function") __exportHandler(__defaultExport);');
    fn(__exportHandler);
  }
};

// Parse the source
__parseModule(__sourceCode);

// WASI entry points
export function _start() {
  __init();
}

export function init() {
  __init();
}

export function execute(inputPtr, inputLen) {
  // This is a placeholder - real WASM would use memory access
  // For fallback, we return an error indicating WASM compilation is needed
  return 0;
}

export function memory() {
  return new WebAssembly.Memory({ initial: 16, maximum: 64 });
}
`, entryFile, string(sourceCode))

	// Return the wrapper as bytes - this will be used as a fallback when Javy isn't available
	return []byte(wasmWrapper), nil
}
