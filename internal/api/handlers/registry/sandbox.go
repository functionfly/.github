package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
)

// SandboxExecutor handles execution of WASM modules in a sandboxed environment
type SandboxExecutor struct {
	runtimePath string
	tempDir     string
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

	return &SandboxExecutor{
		runtimePath: runtimePath,
		tempDir:     tempDir,
	}, nil
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

	// Create temporary WASM file
	wasmFile, err := os.CreateTemp(se.tempDir, "function-*.wasm")
	if err != nil {
		return nil, fmt.Errorf("failed to create WASM temp file: %w", err)
	}
	defer os.Remove(wasmFile.Name())
	defer wasmFile.Close()

	// Write WASM binary to file
	if _, err := wasmFile.Write(fnVersion.WasmBinary); err != nil {
		return nil, fmt.Errorf("failed to write WASM binary: %w", err)
	}
	wasmFile.Close()

	// Execute using the Rust runtime with resource limits
	result, err := se.executeWithRuntimeLimits(wasmFile.Name(), fnVersion, input, timeoutMs, maxMemoryMB, maxCPUTimeMs)
	if err != nil {
		return nil, fmt.Errorf("sandbox execution failed: %w", err)
	}

	return result, nil
}

// executeWithRuntimeLimits executes WASM using the Rust runtime subprocess with specific resource limits
// The runtime is an HTTP server, so we start it and make HTTP requests to it
func (se *SandboxExecutor) executeWithRuntimeLimits(wasmPath string, fnVersion *storage.RegistryFunctionVersion, input []byte, timeoutMs, maxMemoryMB, maxCPUTimeMs int) ([]byte, error) {
	// Find an available port for the runtime server
	port, err := getAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	// Prepare command arguments for the HTTP server
	args := []string{
		"--port", fmt.Sprintf("%d", port),
		"--wasm", wasmPath,
		"--function", fnVersion.FunctionID.String(), // Use function ID as function name
		"--runtime", fnVersion.Runtime,
		"--timeout-ms", fmt.Sprintf("%d", maxCPUTimeMs),
		"--memory-mb", fmt.Sprintf("%d", maxMemoryMB),
	}

	// Add capabilities if declared in function version
	if len(fnVersion.Capabilities) > 0 {
		// Parse JSON array to comma-separated string
		var caps []string
		if err := json.Unmarshal(fnVersion.Capabilities, &caps); err == nil {
			args = append(args, "--capabilities", strings.Join(caps, ","))
		}
	}

	// Add deterministic flag if set
	if fnVersion.Deterministic {
		args = append(args, "--deterministic")
	}

	// Create context with timeout for the entire operation (startup + execution + shutdown)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs+10000)*time.Millisecond)
	defer cancel()

	// Create the command to start the HTTP server
	cmd := exec.CommandContext(ctx, se.runtimePath, args...)
	cmd.Dir = se.tempDir

	// Set resource limits using environment variables
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("FUNCTIONFLY_MEMORY_LIMIT_MB=%d", maxMemoryMB),
		fmt.Sprintf("FUNCTIONFLY_CPU_LIMIT_MS=%d", maxCPUTimeMs),
		fmt.Sprintf("FUNCTIONFLY_TIMEOUT_MS=%d", timeoutMs),
	)

	// Capture stderr for debugging
	var stderr safeBuffer
	cmd.Stderr = &stderr

	// Start the server process
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, &execution.ExecutionError{
			Err: fmt.Errorf("failed to start runtime server: %w", err),
			ResourceUsage: &execution.ResourceUsage{
				MaxMemoryMB:    maxMemoryMB,
				MaxCPUTimeMs:   maxCPUTimeMs,
				MemoryUsedMB:   0,
				CPUTimeUsedMs:  int(time.Since(startTime).Milliseconds()),
				WallTimeUsedMs: int(time.Since(startTime).Milliseconds()),
			},
			TerminatedBy: "error",
		}
	}

	// Ensure we kill the process when done
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for the server to be ready
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := waitForServerReady(ctx, serverURL, 5*time.Second); err != nil {
		return nil, &execution.ExecutionError{
			Err: fmt.Errorf("runtime server failed to start: %w, stderr: %s", err, stderr.String()),
			ResourceUsage: &execution.ResourceUsage{
				MaxMemoryMB:    maxMemoryMB,
				MaxCPUTimeMs:   maxCPUTimeMs,
				MemoryUsedMB:   0,
				CPUTimeUsedMs:  int(time.Since(startTime).Milliseconds()),
				WallTimeUsedMs: int(time.Since(startTime).Milliseconds()),
			},
			TerminatedBy: "error",
		}
	}

	// Make HTTP request to execute the function
	reqBody := map[string]interface{}{
		"input": string(input),
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	// Make the request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", serverURL+"/", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		wallTimeUsedMs := int(time.Since(startTime).Milliseconds())
		terminatedBy := "error"
		if ctx.Err() == context.DeadlineExceeded {
			terminatedBy = "timeout"
		}
		return nil, &execution.ExecutionError{
			Err: fmt.Errorf("HTTP request failed: %w", err),
			ResourceUsage: &execution.ResourceUsage{
				MaxMemoryMB:    maxMemoryMB,
				MaxCPUTimeMs:   maxCPUTimeMs,
				MemoryUsedMB:   parseMemoryUsage(stderr.String()),
				CPUTimeUsedMs:  wallTimeUsedMs,
				WallTimeUsedMs: wallTimeUsedMs,
			},
			TerminatedBy: terminatedBy,
		}
	}
	defer resp.Body.Close()

	wallTimeUsedMs := int(time.Since(startTime).Milliseconds())

	// Read response
	var respBody struct {
		Result   string `json:"result"`
		Error    string `json:"error"`
		ExecTime int    `json:"exec_time_ms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, &execution.ExecutionError{
			Err: fmt.Errorf("failed to decode response: %w", err),
			ResourceUsage: &execution.ResourceUsage{
				MaxMemoryMB:    maxMemoryMB,
				MaxCPUTimeMs:   maxCPUTimeMs,
				MemoryUsedMB:   parseMemoryUsage(stderr.String()),
				CPUTimeUsedMs:  wallTimeUsedMs,
				WallTimeUsedMs: wallTimeUsedMs,
			},
			TerminatedBy: "error",
		}
	}

	// Check for execution error in response
	if respBody.Error != "" {
		terminatedBy := "error"
		if strings.Contains(respBody.Error, "timeout") {
			terminatedBy = "timeout"
		} else if strings.Contains(respBody.Error, "memory") {
			terminatedBy = "memory_limit"
		}
		return nil, &execution.ExecutionError{
			Err: fmt.Errorf("function execution error: %s", respBody.Error),
			ResourceUsage: &execution.ResourceUsage{
				MaxMemoryMB:    maxMemoryMB,
				MaxCPUTimeMs:   maxCPUTimeMs,
				MemoryUsedMB:   parseMemoryUsage(stderr.String()),
				CPUTimeUsedMs:  respBody.ExecTime,
				WallTimeUsedMs: wallTimeUsedMs,
			},
			TerminatedBy: terminatedBy,
		}
	}

	return []byte(respBody.Result), nil
}

// getAvailablePort finds an available port on localhost
func getAvailablePort() (int, error) {
	// Use net.Listen to find an available port
	// This is more reliable than checking /dev/tcp
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to find available port: %w", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// waitForServerReady waits for the HTTP server to be ready to accept requests
func waitForServerReady(ctx context.Context, serverURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 100 * time.Millisecond}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to connect to the health endpoint
		resp, err := client.Get(serverURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		// Wait a bit before retrying
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("server not ready after %v", timeout)
}

// executeWithRuntime executes WASM using the Rust runtime subprocess with resource monitoring
func (se *SandboxExecutor) executeWithRuntime(ctx context.Context, wasmPath string, fnVersion *storage.RegistryFunctionVersion, input []byte, timeoutMs int) ([]byte, error) {
	// Get resource limits from context (set by security middleware)
	maxMemoryMB := fnVersion.MemoryMB
	maxCPUTimeMs := timeoutMs

	// Check if we have resource limits from context (set by security middleware)
	if limits := middleware.GetResourceLimitsFromContext(ctx); limits != nil {
		maxMemoryMB = limits.MaxMemoryMB
		maxCPUTimeMs = limits.MaxCPUTimeMs
		timeoutMs = limits.TimeoutMs
	}

	// Prepare command arguments
	args := []string{
		"--wasm", wasmPath,
		"--function", fnVersion.Version, // Use version as function name
		"--runtime", fnVersion.Runtime,
		"--timeout-ms", fmt.Sprintf("%d", maxCPUTimeMs),
		"--memory-mb", fmt.Sprintf("%d", maxMemoryMB),
	}

	// Add capabilities if declared in function version
	if len(fnVersion.Capabilities) > 0 {
		capStr := string(fnVersion.Capabilities)
		// Parse JSON array to comma-separated string
		var caps []string
		if err := json.Unmarshal(fnVersion.Capabilities, &caps); err == nil {
			args = append(args, "--capabilities", strings.Join(caps, ","))
			_ = capStr // suppress unused warning
		}
	}

	// Add deterministic flag if set
	if fnVersion.Deterministic {
		args = append(args, "--deterministic")
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// Create the command with resource limits
	cmd := exec.CommandContext(execCtx, se.runtimePath, args...)
	cmd.Dir = se.tempDir

	// Set resource limits using environment variables (since syscall may not work cross-platform)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("FUNCTIONFLY_MEMORY_LIMIT_MB=%d", maxMemoryMB),
		fmt.Sprintf("FUNCTIONFLY_CPU_LIMIT_MS=%d", maxCPUTimeMs),
		fmt.Sprintf("FUNCTIONFLY_TIMEOUT_MS=%d", timeoutMs),
	)

	// Note: Resource limits are enforced by the Rust runtime via environment variables

	// Set up stdin with input data
	cmd.Stdin = &inputReader{data: input}

	// Capture stdout and stderr
	var stdout, stderr safeBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	startTime := time.Now()
	err := cmd.Run()

	// Parse memory usage from stderr output
	memoryUsedMB := parseMemoryUsage(stderr.String())

	// Basic resource usage
	resourceUsage := &execution.ResourceUsage{
		MaxMemoryMB:    maxMemoryMB,
		MaxCPUTimeMs:   maxCPUTimeMs,
		MemoryUsedMB:   memoryUsedMB,
		CPUTimeUsedMs:  int(time.Since(startTime).Milliseconds()),
		WallTimeUsedMs: int(time.Since(startTime).Milliseconds()),
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return nil, &execution.ExecutionError{
			Err:           fmt.Errorf("execution timeout after %dms", timeoutMs),
			ResourceUsage: resourceUsage,
			TerminatedBy:  "timeout",
		}
	}

	// Check resource limit violations
	if resourceUsage.MemoryUsedMB > float64(maxMemoryMB) {
		return nil, &execution.ExecutionError{
			Err:           fmt.Errorf("memory limit exceeded: used %.2f MB of %d MB", resourceUsage.MemoryUsedMB, maxMemoryMB),
			ResourceUsage: resourceUsage,
			TerminatedBy:  "memory_limit",
		}
	}

	if resourceUsage.CPUTimeUsedMs > maxCPUTimeMs {
		return nil, &execution.ExecutionError{
			Err:           fmt.Errorf("CPU time limit exceeded: used %d ms of %d ms", resourceUsage.CPUTimeUsedMs, maxCPUTimeMs),
			ResourceUsage: resourceUsage,
			TerminatedBy:  "cpu_limit",
		}
	}

	// Check for execution errors
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return nil, &execution.ExecutionError{
				Err:           fmt.Errorf("sandbox execution failed (exit code %d): %s", exitErr.ExitCode(), stderr.String()),
				ResourceUsage: resourceUsage,
				TerminatedBy:  "error",
			}
		}
		return nil, &execution.ExecutionError{
			Err:           fmt.Errorf("failed to execute sandbox: %w", err),
			ResourceUsage: resourceUsage,
			TerminatedBy:  "error",
		}
	}

	// Return the output
	output := stdout.Bytes()
	if len(output) == 0 && stderr.Len() > 0 {
		// If stdout is empty but stderr has content, it might be an error
		return nil, &execution.ExecutionError{
			Err:           fmt.Errorf("sandbox execution error: %s", stderr.String()),
			ResourceUsage: resourceUsage,
			TerminatedBy:  "error",
		}
	}

	return output, nil
}

// Close cleans up the sandbox executor
func (se *SandboxExecutor) Close() error {
	if se.tempDir != "" {
		return os.RemoveAll(se.tempDir)
	}
	return nil
}

// inputReader provides input data to the subprocess
type inputReader struct {
	data []byte
	pos  int
}

func (ir *inputReader) Read(p []byte) (n int, err error) {
	if ir.pos >= len(ir.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, ir.data[ir.pos:])
	ir.pos += n
	return n, nil
}

// safeBuffer is a thread-safe buffer for capturing output.
// It is safe for concurrent use: Write may be called from the OS pipe reader
// goroutine while String/Bytes/Len are called from the main goroutine.
type safeBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (sb *safeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.data = append(sb.data, p...)
	return len(p), nil
}

func (sb *safeBuffer) Bytes() []byte {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	cp := make([]byte, len(sb.data))
	copy(cp, sb.data)
	return cp
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return string(sb.data)
}

func (sb *safeBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return len(sb.data)
}

// parseMemoryUsage attempts to extract memory usage from runtime output
// Returns 0 if no memory usage information is found
func parseMemoryUsage(stderr string) float64 {
	// Look for patterns like "memory used: 15.2 MB" or "peak memory: 12.8MB"
	patterns := []string{
		`memory used:\s*([0-9.]+)\s*MB`,
		`peak memory:\s*([0-9.]+)\s*MB`,
		`memory:\s*([0-9.]+)\s*MB`,
		`max memory:\s*([0-9.]+)\s*MB`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(stderr)
		if len(matches) > 1 {
			if mem, err := strconv.ParseFloat(matches[1], 64); err == nil {
				return mem
			}
		}
	}

	return 0
}

// findLocalRuntime finds the local runtime binary (similar to fly dev)
func findLocalRuntime() (string, error) {
	// Check current directory first
	localPath := filepath.Join(".", "functionfly-local")
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Check GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	gopathBin := filepath.Join(gopath, "bin", "functionfly-local")
	if _, err := os.Stat(gopathBin); err == nil {
		return gopathBin, nil
	}

	// Check PATH
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		binPath := filepath.Join(dir, "functionfly-local")
		if _, err := os.Stat(binPath); err == nil {
			return binPath, nil
		}
	}

	// Check local Rust build
	runtimeDir := filepath.Join(".", "runtimes", "local")
	releasePath := filepath.Join(runtimeDir, "target", "release", "functionfly-local")
	if _, err := os.Stat(releasePath); err == nil {
		return releasePath, nil
	}

	debugPath := filepath.Join(runtimeDir, "target", "debug", "functionfly-local")
	if _, err := os.Stat(debugPath); err == nil {
		return debugPath, nil
	}

	return "", fmt.Errorf("local runtime not found. Please ensure the Rust runtime is built and available")
}
