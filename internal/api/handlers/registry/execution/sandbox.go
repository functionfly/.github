package execution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// SandboxExecutor handles execution of WASM modules in a sandboxed environment
// It communicates with the local runtime via HTTP
type SandboxExecutor struct {
	runtimePath  string
	tempDir      string
	runtimePort  int
	runtimeCmd   *exec.Cmd
	httpClient   *http.Client
	runtimeMu    sync.Mutex
	isRunning    bool
	wasmPath     string
	wasmHash     string // SHA-256 hash of the loaded WASM binary
	fnVersion    *storage.RegistryFunctionVersion
}

// hashWasmBinary computes a hex-encoded SHA-256 hash of the given WASM bytes.
func hashWasmBinary(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// NewSandboxExecutor creates a new sandbox executor
func NewSandboxExecutor() (*SandboxExecutor, error) {
	// Find the local runtime binary
	runtimePath, err := findLocalRuntime()
	if err != nil {
		return nil, fmt.Errorf("sandbox runtime not found: %w", err)
	}

	// Create temporary directory for WASM files
	tempDir, err := os.MkdirTemp("", "functionfly-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create HTTP client with reasonable defaults
	// Note: We need to disable keep-alives because the runtime server closes
	// connections after each request, which can cause EOF errors
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   true,
			MaxConnsPerHost:     1,
			MaxIdleConnsPerHost: 1,
		},
	}

	return &SandboxExecutor{
		runtimePath: runtimePath,
		tempDir:     tempDir,
		httpClient:  httpClient,
	}, nil
}

// Close closes the sandbox executor and stops the runtime
func (se *SandboxExecutor) Close() {
	se.stopRuntime()
	if se.tempDir != "" {
		os.RemoveAll(se.tempDir)
	}
}

// stopRuntime stops the running runtime process
func (se *SandboxExecutor) stopRuntime() {
	se.runtimeMu.Lock()
	defer se.runtimeMu.Unlock()

	if se.runtimeCmd != nil && se.runtimeCmd.Process != nil {
		logrus.Debug("Stopping runtime process...")

		// Try graceful shutdown first
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Send interrupt signal
		se.runtimeCmd.Process.Signal(os.Interrupt)

		// Wait for process to exit or kill after timeout
		done := make(chan error, 1)
		go func() {
			done <- se.runtimeCmd.Wait()
		}()

		select {
		case <-done:
			logrus.Debug("Runtime stopped gracefully")
		case <-ctx.Done():
			logrus.Warn("Runtime did not stop gracefully, killing...")
			se.runtimeCmd.Process.Kill()
			<-done
		}

		se.isRunning = false
		se.runtimeCmd = nil
	}
}

// ExecuteFunction executes a function version with the given input
func (se *SandboxExecutor) ExecuteFunction(fnVersion *storage.RegistryFunctionVersion, input []byte, timeoutMs int) ([]byte, error) {
	return se.ExecuteFunctionWithLimits(fnVersion, input, timeoutMs, fnVersion.MemoryMB, timeoutMs)
}

// ExecuteFunctionWithLimits executes a function version with specific resource limits
func (se *SandboxExecutor) ExecuteFunctionWithLimits(fnVersion *storage.RegistryFunctionVersion, input []byte, timeoutMs, maxMemoryMB, maxCPUTimeMs int) ([]byte, error) {
	if len(fnVersion.WasmBinary) == 0 {
		return nil, fmt.Errorf("function version has no WASM binary")
	}

	// Compute hash of the incoming WASM binary so we can detect whether the
	// runtime already has the correct module loaded without relying on a
	// temp-file path (which changes on every call).
	incomingHash := hashWasmBinary(fnVersion.WasmBinary)

	se.runtimeMu.Lock()
	needsRestart := !se.isRunning || se.wasmHash != incomingHash
	currentWasmPath := se.wasmPath
	se.runtimeMu.Unlock()

	var wasmPath string
	if needsRestart {
		// Write the WASM binary to a stable, deterministic path inside tempDir
		// so the same file is reused across calls for the same binary.
		wasmPath = filepath.Join(se.tempDir, "function-"+incomingHash+".wasm")
		if err := os.WriteFile(wasmPath, fnVersion.WasmBinary, 0600); err != nil {
			return nil, fmt.Errorf("failed to write WASM binary: %w", err)
		}
		se.stopRuntime()
	} else {
		// Reuse the already-loaded WASM file path.
		wasmPath = currentWasmPath
	}

	// Ensure runtime is running with the correct WASM file
	ctx := context.Background()
	if err := se.ensureRuntimeRunning(ctx, wasmPath, fnVersion, maxMemoryMB, maxCPUTimeMs); err != nil {
		return nil, fmt.Errorf("failed to start runtime: %w", err)
	}

	// Execute via HTTP
	result, err := se.executeViaHTTP(input, timeoutMs)
	if err != nil {
		// Check if the process is still running
		if se.runtimeCmd != nil && se.runtimeCmd.Process != nil {
			if err := se.runtimeCmd.Process.Signal(syscall.Signal(0)); err != nil {
				// Process is not running
				return nil, fmt.Errorf("execution failed (runtime process died): %w", err)
			}
		}
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return result, nil
}

// ensureRuntimeRunning starts the runtime HTTP server if not already running
func (se *SandboxExecutor) ensureRuntimeRunning(ctx context.Context, wasmPath string, fnVersion *storage.RegistryFunctionVersion, maxMemoryMB, maxCPUTimeMs int) error {
	se.runtimeMu.Lock()
	defer se.runtimeMu.Unlock()

	// Compute the hash of the WASM file on disk to compare with the currently
	// loaded hash.  This avoids restarting the runtime when the same binary is
	// passed in via a different temp-file path.
	incomingFileHash := ""
	if data, err := os.ReadFile(wasmPath); err == nil {
		incomingFileHash = hashWasmBinary(data)
	}

	if se.isRunning && se.wasmHash != "" && se.wasmHash == incomingFileHash {
		return nil
	}

	// Find available port
	port, err := getAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	se.runtimePort = port

	// Use the runtime from the function version metadata
	// The runtime indicates the language the function was written in (python3.12, nodejs, etc.)
	// Even though the function is compiled to WASM, the runtime determines how to execute it
	// IMPORTANT: When a function has a WASM binary, we must use "wasm" runtime type
	// The original language runtime (python, nodejs) is just metadata about the source
	runtimeType := fnVersion.Runtime

	// If we have a WASM binary, always use "wasm" runtime for execution
	// The local runtime needs to know it's executing WASM, not Python/JS source
	if len(fnVersion.WasmBinary) > 0 {
		runtimeType = "wasm"
	}

	// Log WASM detection for debugging purposes
	wasmContent, err := os.ReadFile(wasmPath)
	if err == nil && len(wasmContent) >= 4 {
		// Check for WASM magic bytes
		if wasmContent[0] == 0x00 && wasmContent[1] == 0x61 && wasmContent[2] == 0x73 && wasmContent[3] == 0x6D {
			logrus.WithField("runtime", runtimeType).Debug("Detected valid WASM binary")
		}
	}

	// Build command arguments
	args := []string{
		"--port", fmt.Sprintf("%d", port),
		"--wasm", wasmPath,
		"--runtime", runtimeType,
		"--timeout-ms", fmt.Sprintf("%d", maxCPUTimeMs),
		"--memory-mb", fmt.Sprintf("%d", maxMemoryMB),
		"--function", fnVersion.FunctionID.String(),
		"--version", fnVersion.Version,
	}

	// Add capabilities if declared in function version
	if len(fnVersion.Capabilities) > 0 {
		var caps []string
		if err := json.Unmarshal(fnVersion.Capabilities, &caps); err == nil {
			args = append(args, "--capabilities", strings.Join(caps, ","))
		}
	}

	// Add deterministic flag if set
	if fnVersion.Deterministic {
		args = append(args, "--deterministic")
	}

	logrus.WithFields(logrus.Fields{
		"port":             port,
		"wasm":             wasmPath,
		"runtime":          runtimeType,
		"original_runtime": fnVersion.Runtime,
	}).Info("Starting local runtime HTTP server")

	// Create the command
	se.runtimeCmd = exec.CommandContext(ctx, se.runtimePath, args...)
	se.runtimeCmd.Dir = se.tempDir

	// Set environment variables for resource limits
	se.runtimeCmd.Env = append(os.Environ(),
		fmt.Sprintf("FUNCTIONFLY_MEMORY_LIMIT_MB=%d", maxMemoryMB),
		fmt.Sprintf("FUNCTIONFLY_CPU_LIMIT_MS=%d", maxCPUTimeMs),
	)

	// Don't capture stdout/stderr to buffers - this can cause the process to block
	// when the buffers fill up. Instead, let the process write to its own stdout/stderr.
	se.runtimeCmd.Stdout = os.Stdout
	se.runtimeCmd.Stderr = os.Stderr

	logrus.WithFields(logrus.Fields{
		"path": se.runtimePath,
		"args": args,
	}).Debug("Starting runtime process")

	// Start the process
	if err := se.runtimeCmd.Start(); err != nil {
		return fmt.Errorf("failed to start runtime: %w", err)
	}

	logrus.WithField("pid", se.runtimeCmd.Process.Pid).Debug("Runtime process started")

	// Wait for server to be ready
	if err := se.waitForServerReady(ctx); err != nil {
		// Kill the process if it didn't start properly
		if se.runtimeCmd.Process != nil {
			se.runtimeCmd.Process.Kill()
		}
		return fmt.Errorf("runtime server did not become ready: %w", err)
	}

	se.isRunning = true
	se.wasmPath = wasmPath
	se.wasmHash = incomingFileHash
	se.fnVersion = fnVersion

	logrus.WithFields(logrus.Fields{
		"port":      port,
		"wasm_hash": incomingFileHash[:8], // log first 8 chars for brevity
	}).Info("Local runtime HTTP server is ready")
	return nil
}

// waitForServerReady polls the health endpoint until the server is ready
func (se *SandboxExecutor) waitForServerReady(ctx context.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", se.runtimePort)

	// Create a client with short timeout for health checks
	healthClient := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	// Try for up to 5 seconds
	maxAttempts := 50
	for i := 0; i < maxAttempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		resp, err := healthClient.Do(req)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("runtime server did not become ready after 5 seconds")
}

// executeViaHTTP sends the execution request to the runtime via HTTP
func (se *SandboxExecutor) executeViaHTTP(input []byte, timeoutMs int) ([]byte, error) {
	se.runtimeMu.Lock()
	port := se.runtimePort
	se.runtimeMu.Unlock()

	// Create request body - the runtime expects {"input": "..."}
	reqBody := map[string]interface{}{
		"input": string(input),
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create URL
	url := fmt.Sprintf("http://127.0.0.1:%d/", port)

	logrus.WithFields(logrus.Fields{
		"url":  url,
		"body": string(jsonBody),
	}).Debug("Sending execution request to runtime")

	// Try the request with retries
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			logrus.WithField("attempt", attempt).Debug("Retrying HTTP request")
			time.Sleep(50 * time.Millisecond)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Connection", "close")

		// Execute request
		resp, err := se.httpClient.Do(req)
		if err != nil {
			cancel()
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("execution timeout after %dms", timeoutMs)
			}
			lastErr = err
			logrus.WithError(err).WithField("attempt", attempt).Debug("HTTP request failed, may retry")
			continue
		}
		defer resp.Body.Close()
		defer cancel()

		// Check status code
		if resp.StatusCode != 200 {
			var errResp struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
				return nil, fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, errResp.Error)
			}
			return nil, fmt.Errorf("runtime returned status %d", resp.StatusCode)
		}

		// Parse response - the runtime returns {"result": "...", "exec_time_ms": ..., ...}
		var result struct {
			Result     string `json:"result"`
			ExecTimeMs uint64 `json:"exec_time_ms"`
			CacheHit   bool   `json:"cache_hit"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"exec_time_ms": result.ExecTimeMs,
			"cache_hit":    result.CacheHit,
		}).Debug("Function executed successfully")

		return []byte(result.Result), nil
	}

	return nil, fmt.Errorf("HTTP request failed after retries: %w", lastErr)
}

// getAvailablePort finds an available port on localhost
func getAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// findLocalRuntime finds the local runtime binary
func findLocalRuntime() (string, error) {
	// Check current directory first
	paths := []string{
		"./runtimes/local/target/release/functionfly-local",
		"./runtimes/local/target/debug/functionfly-local",
		"./setup-bin",
		"/usr/local/bin/functionfly-local",
	}

	for _, path := range paths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				logrus.Infof("Found local runtime at: %s", absPath)
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("local runtime not found, searched: %v", paths)
}

// executeLocallyWithLimits executes a function locally with specific resource limits
func executeLocallyWithLimits(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, maxMemoryMB, maxCPUTimeMs int) (json.RawMessage, error) {
	// Create sandbox executor
	executor, err := NewSandboxExecutor()
	if err != nil {
		return nil, &ExecutionError{
			Err: fmt.Errorf("sandbox initialization failed: %w", err),
			ResourceUsage: &ResourceUsage{
				MaxMemoryMB:  maxMemoryMB,
				MaxCPUTimeMs: maxCPUTimeMs,
			},
			TerminatedBy: "error",
		}
	}
	defer executor.Close()

	// Execute the function with resource limits
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, &ExecutionError{
			Err: fmt.Errorf("failed to marshal input: %w", err),
			ResourceUsage: &ResourceUsage{
				MaxMemoryMB:  maxMemoryMB,
				MaxCPUTimeMs: maxCPUTimeMs,
			},
			TerminatedBy: "error",
		}
	}

	output, err := executor.ExecuteFunctionWithLimits(fnVersion, inputBytes, fnVersion.TimeoutMs, maxMemoryMB, maxCPUTimeMs)
	if err != nil {
		if execErr, ok := err.(*ExecutionError); ok {
			return nil, execErr
		}
		return nil, &ExecutionError{
			Err: err,
			ResourceUsage: &ResourceUsage{
				MaxMemoryMB:  maxMemoryMB,
				MaxCPUTimeMs: maxCPUTimeMs,
			},
			TerminatedBy: "error",
		}
	}

	return json.RawMessage(output), nil
}

// entryFilenameForRuntime returns the default entry filename for a runtime (used for lazy bundling).
func entryFilenameForRuntime(runtime string) string {
	switch {
	case strings.HasPrefix(runtime, "python"):
		return "main.py"
	case runtime == "deno":
		return "main.ts"
	default:
		return "index.js"
	}
}

// executeWithLazyBundling bundles source code to WASM at execution time
// This is used when publish didn't bundle (for faster publish)
func executeWithLazyBundling(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, maxMemoryMB, maxCPUTimeMs int, resourceUsage **ResourceUsage) (json.RawMessage, error) {
	// Get source code from the function version
	sourceCode := fnVersion.SourceCode.String
	if sourceCode == "" {
		return nil, fmt.Errorf("function has no source code to bundle")
	}

	// Parse manifest to get runtime info
	var m manifest.Manifest
	if err := json.Unmarshal(fnVersion.Manifest, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Create a temp dir and write registry source there so the bundler uses it instead of CWD
	tmpDir, err := os.MkdirTemp("", "functionfly-lazy-bundle-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir for lazy bundle: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	entryFile := entryFilenameForRuntime(fnVersion.Runtime)
	entryPath := filepath.Join(tmpDir, entryFile)
	if err := os.WriteFile(entryPath, []byte(sourceCode), 0600); err != nil {
		return nil, fmt.Errorf("failed to write source to temp file: %w", err)
	}
	entryPathAbs, err := filepath.Abs(entryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve temp entry path: %w", err)
	}

	// Create manifest for bundling; set Entry to absolute path so bundler always reads our temp file
	bundleManifest := &manifest.Manifest{
		Name:      m.Name,
		Version:   m.Version,
		Runtime:   fnVersion.Runtime,
		TimeoutMS: &fnVersion.TimeoutMs,
		MemoryMB:  &fnVersion.MemoryMB,
		Entry:     entryPathAbs,
	}

	// Bundle source to WASM (uses MicroPython for Python) from temp dir
	logrus.WithFields(logrus.Fields{
		"function_id": fnVersion.FunctionID,
		"version":     fnVersion.Version,
		"runtime":     fnVersion.Runtime,
	}).Info("Bundling source code to WASM at execution time (lazy bundling)")

	wasmBytes, err := bundler.BundleForWasmRuntimeWithWorkingDirectory(bundleManifest, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to bundle source code: %w", err)
	}

	// Create a temporary fnVersion with the bundled WASM
	bundledFnVersion := *fnVersion
	bundledFnVersion.WasmBinary = wasmBytes

	// Execute the bundled function
	output, execErr := executeLocallyWithLimits(&bundledFnVersion, input, maxMemoryMB, maxCPUTimeMs)
	if execErr != nil {
		if execError, ok := execErr.(*ExecutionError); ok {
			*resourceUsage = execError.ResourceUsage
		}
		return nil, execErr
	}
	return output, nil
}

// executeLocally executes a function locally (used in verification)
func executeLocally(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage) (json.RawMessage, error) {
	return executeLocallyWithLimits(fnVersion, input, fnVersion.MemoryMB, fnVersion.TimeoutMs)
}

// executeOnBackend executes a function on a specific backend
func executeOnBackend(execURL string, input string, timeoutMs int) (json.RawMessage, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"input": input,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request with context for timeout control
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", execURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Registry/1.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute on backend: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	// Read response body
	var response json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode backend response: %w", err)
	}

	return response, nil
}
