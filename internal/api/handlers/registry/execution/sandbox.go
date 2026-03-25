package execution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// EnterpriseExecutionConfig holds config for Enterprise tier (MicroVM) execution
type EnterpriseExecutionConfig struct {
	Enabled                bool
	OrchestratorURL        string
	Tier                   string // "medium" or "high" for enterprise
	TenantID               string
	PythonPackages         []string
	NetworkWhitelist       []string
	StrictNetworkWhitelist bool
	PackageCachingEnabled  bool
}

// SandboxExecutor handles execution of WASM modules in a sandboxed environment
// It communicates with the local runtime via HTTP
type SandboxExecutor struct {
	runtimePath    string
	tempDir        string
	runtimePort    int
	runtimeCmd     *exec.Cmd
	httpClient     *http.Client
	runtimeMu      sync.Mutex
	isRunning      bool
	wasmPath       string
	wasmHash       string // SHA-256 hash of the loaded WASM binary
	fnVersion      *storage.RegistryFunctionVersion
	enterpriseConf *EnterpriseExecutionConfig
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
	return se.ExecuteFunctionWithLimits(fnVersion, input, timeoutMs, fnVersion.MemoryMB, timeoutMs, nil)
}

// ExecuteFunctionWithLimits executes a function version with specific resource limits.
// enterpriseConf is optional; when set for python-microvm runtime, enables MicroVM execution.
func (se *SandboxExecutor) ExecuteFunctionWithLimits(fnVersion *storage.RegistryFunctionVersion, input []byte, timeoutMs, maxMemoryMB, maxCPUTimeMs int, enterpriseConf *EnterpriseExecutionConfig) ([]byte, error) {
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

	se.enterpriseConf = enterpriseConf

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

	// If we have a WASM binary (magic bytes), use "wasm" runtime.
	// For python-microvm, WasmBinary holds Python source - keep python-microvm.
	if len(fnVersion.WasmBinary) >= 4 &&
		fnVersion.WasmBinary[0] == 0x00 && fnVersion.WasmBinary[1] == 0x61 &&
		fnVersion.WasmBinary[2] == 0x73 && fnVersion.WasmBinary[3] == 0x6D {
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

	// Enterprise tier: add args for MicroVM when runtime is python-microvm (must be before exec.Command)
	if se.enterpriseConf != nil && se.enterpriseConf.Enabled && fnVersion.Runtime == plans.RuntimePythonMicroVM {
		args = append(args,
			"--enterprise-enabled",
			"--orchestrator-url", se.enterpriseConf.OrchestratorURL,
			"--tier", se.enterpriseConf.Tier,
		)
		if se.enterpriseConf.TenantID != "" {
			args = append(args, "--tenant-id", se.enterpriseConf.TenantID)
		}
		for _, pkg := range se.enterpriseConf.PythonPackages {
			if strings.TrimSpace(pkg) != "" {
				args = append(args, "--python-packages", pkg)
			}
		}
		for _, host := range se.enterpriseConf.NetworkWhitelist {
			if strings.TrimSpace(host) != "" {
				args = append(args, "--network-whitelist", host)
			}
		}
		if se.enterpriseConf.StrictNetworkWhitelist {
			args = append(args, "--strict-network-whitelist")
		}
		if se.enterpriseConf.PackageCachingEnabled {
			cacheDir := filepath.Join(se.tempDir, "package-cache")
			_ = os.MkdirAll(cacheDir, 0o700)
			args = append(args,
				"--package-caching-enabled",
				"--package-cache-dir", cacheDir,
			)
		}
		logrus.WithFields(logrus.Fields{
			"tenant_id":  se.enterpriseConf.TenantID,
			"packages":   len(se.enterpriseConf.PythonPackages),
			"net_wl_len": len(se.enterpriseConf.NetworkWhitelist),
		}).Debug("Enterprise MicroVM execution enabled")
	}

	// Create the command (args must be complete above)
	se.runtimeCmd = exec.CommandContext(ctx, se.runtimePath, args...)
	se.runtimeCmd.Dir = se.tempDir

	// Build subprocess environment: inherit host env, add resource limits, and
	// forward the MicroVM API token so the local runtime can authenticate to the
	// orchestrator without it appearing in the CLI args (which show up in ps/logs).
	runtimeEnv := append(os.Environ(),
		fmt.Sprintf("FUNCTIONFLY_MEMORY_LIMIT_MB=%d", maxMemoryMB),
		fmt.Sprintf("FUNCTIONFLY_CPU_LIMIT_MS=%d", maxCPUTimeMs),
	)
	if se.enterpriseConf != nil && se.enterpriseConf.Enabled {
		if tok := os.Getenv("FUNCTIONFLY_MICROVM_API_TOKEN"); tok != "" {
			runtimeEnv = append(runtimeEnv, "FUNCTIONFLY_MICROVM_API_TOKEN="+tok)
		}
	}
	se.runtimeCmd.Env = runtimeEnv

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

// ExecuteLocally runs a registry function version locally (sandbox/WASM) with the given resource limits.
// It is the exported entry point for use by flywheel and other callers that need to run registry functions.
// Pass fn=nil, backendRepo=nil when tenant context is not available (enterprise MicroVM will be disabled).
func ExecuteLocally(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, maxMemoryMB, maxCPUTimeMs int) (json.RawMessage, error) {
	return executeLocallyWithLimits(fnVersion, input, maxMemoryMB, maxCPUTimeMs, nil, nil)
}

// executeLocallyWithLimits executes a function locally with specific resource limits.
// fn and backendRepo are optional; when provided and tenant has enterprise plan, enables MicroVM for python-microvm runtime.
func executeLocallyWithLimits(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, maxMemoryMB, maxCPUTimeMs int, fn *storage.RegistryFunction, backendRepo storage.Repository) (json.RawMessage, error) {
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

	enterpriseConf := buildEnterpriseConfig(fnVersion, fn, backendRepo)
	output, err := executor.ExecuteFunctionWithLimits(fnVersion, inputBytes, fnVersion.TimeoutMs, maxMemoryMB, maxCPUTimeMs, enterpriseConf)
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

// microvmManifestExtras parses optional Enterprise / Python fields from stored manifest JSON.
// Supports both `{"function":{...}}` and flat manifests.
type microvmManifestExtras struct {
	Function *struct {
		Python *struct {
			Packages []string `json:"packages"`
		} `json:"python"`
		Enterprise *struct {
			NetworkAllowlist         []string `json:"network_allowlist"`
			PackageCacheEnabled      *bool    `json:"package_cache_enabled"`
			StrictNetworkWhitelist   *bool    `json:"strict_network_whitelist"`
		} `json:"enterprise"`
	} `json:"function"`
	Python *struct {
		Packages []string `json:"packages"`
	} `json:"python"`
	Enterprise *struct {
		NetworkAllowlist       []string `json:"network_allowlist"`
		PackageCacheEnabled    *bool    `json:"package_cache_enabled"`
		StrictNetworkWhitelist *bool    `json:"strict_network_whitelist"`
	} `json:"enterprise"`
}

func parseMicroVMManifest(manifest json.RawMessage) (pkgs []string, net []string, pkgCache, strictNet bool) {
	if len(manifest) == 0 {
		return nil, nil, false, false
	}
	var m microvmManifestExtras
	if err := json.Unmarshal(manifest, &m); err != nil {
		return nil, nil, false, false
	}
	if m.Function != nil {
		if m.Function.Python != nil && len(m.Function.Python.Packages) > 0 {
			pkgs = append(pkgs, m.Function.Python.Packages...)
		}
		if m.Function.Enterprise != nil {
			net = append(net, m.Function.Enterprise.NetworkAllowlist...)
			if m.Function.Enterprise.PackageCacheEnabled != nil {
				pkgCache = *m.Function.Enterprise.PackageCacheEnabled
			}
			if m.Function.Enterprise.StrictNetworkWhitelist != nil {
				strictNet = *m.Function.Enterprise.StrictNetworkWhitelist
			}
		}
	}
	if m.Python != nil && len(m.Python.Packages) > 0 {
		pkgs = append(pkgs, m.Python.Packages...)
	}
	if m.Enterprise != nil {
		net = append(net, m.Enterprise.NetworkAllowlist...)
		if m.Enterprise.PackageCacheEnabled != nil {
			pkgCache = *m.Enterprise.PackageCacheEnabled
		}
		if m.Enterprise.StrictNetworkWhitelist != nil {
			strictNet = *m.Enterprise.StrictNetworkWhitelist
		}
	}
	return pkgs, net, pkgCache, strictNet
}

// buildEnterpriseConfig builds EnterpriseExecutionConfig when tenant has enterprise plan and runtime is python-microvm.
func buildEnterpriseConfig(fnVersion *storage.RegistryFunctionVersion, fn *storage.RegistryFunction, backendRepo storage.Repository) *EnterpriseExecutionConfig {
	if fn == nil || fn.TenantID == nil || backendRepo == nil || fnVersion.Runtime != plans.RuntimePythonMicroVM {
		return nil
	}
	plan := getTenantPlanFromContext(backendRepo, *fn.TenantID)
	if plan != plans.PlanEnterprise {
		return nil
	}
	orchestratorURL := os.Getenv("FUNCTIONFLY_ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:9090"
	}
	pkgs, net, pkgCache, strictNet := parseMicroVMManifest(fnVersion.Manifest)
	return &EnterpriseExecutionConfig{
		Enabled:                true,
		OrchestratorURL:        orchestratorURL,
		Tier:                   "high",
		TenantID:               fn.TenantID.String(),
		PythonPackages:         pkgs,
		NetworkWhitelist:       net,
		StrictNetworkWhitelist: strictNet,
		PackageCachingEnabled:  pkgCache,
	}
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
func executeWithLazyBundling(fnVersion *storage.RegistryFunctionVersion, input json.RawMessage, maxMemoryMB, maxCPUTimeMs int, resourceUsage **ResourceUsage, fn *storage.RegistryFunction, backendRepo storage.Repository) (json.RawMessage, error) {
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

	// python-microvm: use Python source directly (no WASM compilation); CPython runs in Firecracker
	if fnVersion.Runtime == plans.RuntimePythonMicroVM {
		logrus.WithFields(logrus.Fields{
			"function_id": fnVersion.FunctionID,
			"version":     fnVersion.Version,
			"runtime":     fnVersion.Runtime,
		}).Info("Using Python source for MicroVM (no WASM bundle)")

		bundledFnVersion := *fnVersion
		bundledFnVersion.WasmBinary = []byte(sourceCode) // Python source as "binary" for runtime

		output, execErr := executeLocallyWithLimits(&bundledFnVersion, input, maxMemoryMB, maxCPUTimeMs, fn, backendRepo)
		if execErr != nil {
			if execError, ok := execErr.(*ExecutionError); ok {
				*resourceUsage = execError.ResourceUsage
			}
			return nil, execErr
		}
		return output, nil
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
	output, execErr := executeLocallyWithLimits(&bundledFnVersion, input, maxMemoryMB, maxCPUTimeMs, fn, backendRepo)
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
	return executeLocallyWithLimits(fnVersion, input, fnVersion.MemoryMB, fnVersion.TimeoutMs, nil, nil)
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

// ---------------------------------------------------------------------------
// SandboxClient — persistent daemon connection (P1.7)
//
// SandboxClient replaces the per-request SandboxExecutor process-spawn model
// with a single long-lived runtime daemon process.  The daemon handles
// multiple functions concurrently via its internal instance pool, eliminating
// the ~200ms cold-start overhead of spawning a new OS process per request.
//
// Usage:
//
//	client, err := NewSandboxClient(runtimePath)
//	defer client.Close()
//	result, err := client.Execute(fnVersion, input, timeoutMs)
// ---------------------------------------------------------------------------

// SandboxClient manages a single persistent runtime daemon process and
// communicates with it over HTTP.
type SandboxClient struct {
	runtimePath string
	daemonURL   string
	daemonCmd   *exec.Cmd
	httpClient  *http.Client
	mu          sync.Mutex
	isRunning   bool
	tempDir     string
}

// NewSandboxClient creates a SandboxClient and starts the runtime daemon.
// The daemon is started with --daemon-mode which keeps it alive across
// requests and enables the internal Wasm instance pool.
func NewSandboxClient(runtimePath string) (*SandboxClient, error) {
	if runtimePath == "" {
		var err error
		runtimePath, err = findLocalRuntime()
		if err != nil {
			return nil, fmt.Errorf("sandbox runtime not found: %w", err)
		}
	}

	tempDir, err := os.MkdirTemp("", "functionfly-daemon-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Find a free port for the daemon
	port, err := getAvailablePort()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	daemonURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Build daemon command — enable daemon mode and AOT cache
	args := []string{
		"--port", fmt.Sprintf("%d", port),
		"--daemon-mode",
		"--aot-cache-enabled",
	}

	cmd := exec.Command(runtimePath, args...)
	cmd.Dir = tempDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to start runtime daemon: %w", err)
	}

	sc := &SandboxClient{
		runtimePath: runtimePath,
		daemonURL:   daemonURL,
		daemonCmd:   cmd,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				IdleConnTimeout:     90 * time.Second,
				MaxConnsPerHost:     50,
				MaxIdleConnsPerHost: 50,
			},
		},
		isRunning: true,
		tempDir:   tempDir,
	}

	// Wait for the daemon to become ready
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sc.waitForReady(ctx); err != nil {
		sc.Close()
		return nil, fmt.Errorf("runtime daemon did not become ready: %w", err)
	}

	logrus.WithField("url", daemonURL).Info("Runtime daemon started and ready")
	return sc, nil
}

// Close stops the daemon process and cleans up temporary files.
func (sc *SandboxClient) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.daemonCmd != nil && sc.daemonCmd.Process != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		sc.daemonCmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- sc.daemonCmd.Wait() }()
		select {
		case <-done:
		case <-ctx.Done():
			sc.daemonCmd.Process.Kill()
			<-done
		}
		sc.isRunning = false
		sc.daemonCmd = nil
	}

	if sc.tempDir != "" {
		os.RemoveAll(sc.tempDir)
		sc.tempDir = ""
	}
}

// IsRunning reports whether the daemon process is still alive.
func (sc *SandboxClient) IsRunning() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.isRunning
}

// Execute sends a function execution request to the persistent daemon.
// The daemon looks up the function in its internal pool and executes it,
// returning the result without spawning a new process.
func (sc *SandboxClient) Execute(fnVersion *storage.RegistryFunctionVersion, input []byte, timeoutMs int) ([]byte, error) {
	if len(fnVersion.WasmBinary) == 0 {
		return nil, fmt.Errorf("function version has no WASM binary")
	}

	// Build the per-function execution URL:
	//   POST /execute/{functionID}/{version}
	// The daemon routes by function ID + version and uses its pool.
	execURL := fmt.Sprintf("%s/execute/%s/%s",
		sc.daemonURL,
		fnVersion.FunctionID.String(),
		fnVersion.Version,
	)

	// Request body: wasm_binary (base64) + input
	type execRequest struct {
		WasmBinary string `json:"wasm_binary"` // base64-encoded
		Input      string `json:"input"`
		TimeoutMs  int    `json:"timeout_ms"`
		MemoryMB   int    `json:"memory_mb"`
	}

	reqBody := execRequest{
		WasmBinary: base64.StdEncoding.EncodeToString(fnVersion.WasmBinary),
		Input:      string(input),
		TimeoutMs:  timeoutMs,
		MemoryMB:   fnVersion.MemoryMB,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execution request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs+5000)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", execURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("execution timeout after %dms", timeoutMs)
		}
		return nil, fmt.Errorf("daemon request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Error != "" {
			return nil, fmt.Errorf("daemon error (status %d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}

	var result struct {
		Result     string `json:"result"`
		ExecTimeMs uint64 `json:"exec_time_ms"`
		CacheHit   bool   `json:"cache_hit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode daemon response: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"exec_time_ms": result.ExecTimeMs,
		"cache_hit":    result.CacheHit,
	}).Debug("SandboxClient: function executed via daemon")

	return []byte(result.Result), nil
}

// waitForReady polls the daemon health endpoint until it responds.
func (sc *SandboxClient) waitForReady(ctx context.Context) error {
	healthURL := sc.daemonURL + "/health"
	healthClient := &http.Client{Timeout: 200 * time.Millisecond}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resp, err := healthClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
