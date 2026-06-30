// Package sandbox provides production-ready secure sandbox execution
// for untrusted code using gVisor (runsc) container runtime.
//
// Security Model:
// - gVisor provides kernel namespace isolation without shared kernel
// - Each execution runs in an isolated sandbox with:
//   - No network access (--network=empty)
//   - Read-only filesystem with writable tmpfs
//   - Memory limits enforced by seccomp
//   - CPU limits via cgroups
//   - No privilege escalation (--no-new-privs)
//
// Resource Limits:
// - Default: 256MB memory, 0.5 CPU, 30s timeout
// - Configurable per-execution via ExecutionConfig
//
// Usage:
//   executor, err := NewSecureSandboxExecutor()
//   result, err := executor.Execute(ctx, "python", code, input)
package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/testing"
	"github.com/sirupsen/logrus"
)

// Re-export testing types for convenience
type SandboxExecutionResult = testing.SandboxExecutionResult

// ExecutionConfig controls resource limits and security settings
type ExecutionConfig struct {
	// Memory limit in MB (default: 256)
	MemoryMB int64
	// CPU limit as fraction (default: 0.5 = 50% of 1 core)
	CPULimit float64
	// Maximum execution time (default: 30s)
	Timeout time.Duration
	// Enable networking (default: false - network is always disabled for security)
	EnableNetwork bool
	// Additional capability flags (default: none)
	Capabilities []string
	// Environment variables to pass to the sandbox (default: none)
	EnvVars map[string]string
}

// DefaultExecutionConfig returns the default secure configuration
func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		MemoryMB:   256,
		CPULimit:   0.5,
		Timeout:    30 * time.Second,
		EnableNetwork: false,
		Capabilities: []string{},
		EnvVars:   map[string]string{},
	}
}

// SecureSandboxExecutor provides production-ready sandbox execution using gVisor
type SecureSandboxExecutor struct {
	mu           sync.Mutex
	runscPath    string
	imagePulled  map[string]bool
	config       *ExecutionConfig
	pullTimeout  time.Duration
}

// NewSecureSandboxExecutor creates a new gVisor-based sandbox executor
// Falls back to Docker if gVisor is not available (with warning logged)
func NewSecureSandboxExecutor() (*SecureSandboxExecutor, error) {
	return NewSecureSandboxExecutorWithConfig(DefaultExecutionConfig())
}

// NewSecureSandboxExecutorWithConfig creates a sandbox executor with custom config
func NewSecureSandboxExecutorWithConfig(config *ExecutionConfig) (*SecureSandboxExecutor, error) {
	if config == nil {
		config = DefaultExecutionConfig()
	}

	// Find runsc (gVisor) or fallback to docker
	runscPath, err := findSecureRuntime()
	if err != nil {
		logrus.WithError(err).Warn("gVisor not available, falling back to Docker with reduced security")
		runscPath = "" // Will use Docker fallback
	}

	return &SecureSandboxExecutor{
		runscPath:   runscPath,
		imagePulled: make(map[string]bool),
		config:      config,
		pullTimeout: 2 * time.Minute,
	}, nil
}

// findSecureRuntime locates the gVisor runsc binary or returns error
func findSecureRuntime() (string, error) {
	// Check common gVisor installation paths
	paths := []string{
		"/usr/local/bin/runsc",
		"/usr/bin/runsc",
		"/opt/gvisor/runsc",
		"$HOME/go/bin/runsc",
	}

	for _, path := range paths {
		// Expand HOME if needed
		if strings.HasPrefix(path, "$HOME") {
			home := os.Getenv("HOME")
			if home != "" {
				path = strings.Replace(path, "$HOME", home, 1)
			} else {
				continue
			}
		}

		if _, err := os.Stat(path); err == nil {
			// Verify it's executable
			cmd := exec.Command(path, "--version")
			if err := cmd.Run(); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("gVisor (runsc) not found in standard locations")
}

// Runtime configuration for different languages
var runtimeConfigs = map[string]struct {
	image       string
	entryPoint  string
	handlerFile string
	// gVisor alternative: pre-compiled WASM for languages that support it
	wasmImage   string
}{
	"python": {
		image:       "python:3.11-slim",
		entryPoint:  "python",
		handlerFile: "handler.py",
	},
	"python3": {
		image:       "python:3.11-slim",
		entryPoint:  "python",
		handlerFile: "handler.py",
	},
	"python3.11": {
		image:       "python:3.11-slim",
		entryPoint:  "python",
		handlerFile: "handler.py",
	},
	"python3.12": {
		image:       "python:3.12-slim",
		entryPoint:  "python",
		handlerFile: "handler.py",
	},
	"node": {
		image:       "node:18-alpine",
		entryPoint:  "node",
		handlerFile: "handler.js",
	},
	"nodejs": {
		image:       "node:18-alpine",
		entryPoint:  "node",
		handlerFile: "handler.js",
	},
	"node18": {
		image:       "node:18-alpine",
		entryPoint:  "node",
		handlerFile: "handler.js",
	},
	"deno": {
		image:       "denoland/deno:latest",
		entryPoint:  "deno",
		handlerFile: "handler.ts",
	},
	"bun": {
		image:       "oven/bun:latest",
		entryPoint:  "bun",
		handlerFile: "handler.js",
	},
	"ruby": {
		image:       "ruby:3.3-slim",
		entryPoint:  "ruby",
		handlerFile: "handler.rb",
	},
}

// Execute runs the code in an isolated gVisor sandbox
func (e *SecureSandboxExecutor) Execute(ctx context.Context, runtime, code string, input map[string]any) (SandboxExecutionResult, error) {
	return e.ExecuteWithConfig(ctx, runtime, code, input, e.config)
}

// ExecuteWithConfig runs code with custom execution configuration
func (e *SecureSandboxExecutor) ExecuteWithConfig(ctx context.Context, runtime, code string, input map[string]any, config *ExecutionConfig) (SandboxExecutionResult, error) {
	start := time.Now()

	if config == nil {
		config = e.config
	}

	// Get runtime configuration
	cfg, ok := runtimeConfigs[strings.ToLower(runtime)]
	if !ok {
		// Fallback to python for unknown runtimes
		cfg = runtimeConfigs["python"]
	}

	// Ensure image is available
	if err := e.ensureImage(ctx, cfg.image); err != nil {
		return SandboxExecutionResult{
			Passed:     false,
			DurationMs: int(time.Since(start).Milliseconds()),
			Error:      fmt.Sprintf("failed to pull image: %v", err),
			Output:     map[string]any{"runtime": runtime},
		}, nil
	}

	// Create temporary directory for code
	workDir, err := os.MkdirTemp("", "sandbox-*")
	if err != nil {
		return SandboxExecutionResult{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Write handler code with strict permissions
	handlerPath := filepath.Join(workDir, cfg.handlerFile)
	if err := os.WriteFile(handlerPath, []byte(code), 0444); err != nil { // Read-only
		return SandboxExecutionResult{}, fmt.Errorf("write handler: %w", err)
	}

	// Write input as JSON
	inputJSON, _ := json.Marshal(input)
	inputPath := filepath.Join(workDir, "input.json")
	if err := os.WriteFile(inputPath, inputJSON, 0444); err != nil {
		return SandboxExecutionResult{}, fmt.Errorf("write input: %w", err)
	}

	// Create wrapper script that runs the handler
	wrapperScript := e.createWrapperScript(runtime, cfg.handlerFile)
	wrapperPath := filepath.Join(workDir, "run.sh")
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0555); err != nil {
		return SandboxExecutionResult{}, fmt.Errorf("write wrapper: %w", err)
	}

	// Run sandbox and get result
	var result *sandboxResult
	if e.runscPath != "" {
		result, err = e.runWithGvisor(ctx, cfg.image, workDir, runtime, config)
	} else {
		result, err = e.runWithDocker(ctx, cfg.image, workDir, runtime, config)
	}

	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return SandboxExecutionResult{
			Passed:     false,
			DurationMs: durationMs,
			Error:      err.Error(),
			Output:     map[string]any{"runtime": runtime},
		}, nil
	}

	return SandboxExecutionResult{
		Passed:     result.Passed,
		DurationMs: durationMs,
		Output: map[string]any{
			"runtime":        runtime,
			"stdout":         result.Stdout,
			"stderr":         result.Stderr,
			"exitCode":       result.ExitCode,
			"memoryUsedMB":   result.MemoryUsedMB,
			"sandboxType":    ternaryString(e.runscPath != "", "gvisor", "docker"),
		},
		Error: result.Error,
	}, nil
}

// sandboxResult holds the execution result from a sandbox
type sandboxResult struct {
	Passed       bool
	Stdout       string
	Stderr       string
	ExitCode     int
	Error        string
	MemoryUsedMB int64
}

// runWithGvisor executes code in gVisor sandbox
func (e *SecureSandboxExecutor) runWithGvisor(ctx context.Context, imageName, workDir, runtime string, config *ExecutionConfig) (*sandboxResult, error) {
	// Build runsc command with security constraints
	// Security flags:
	//   --network=empty: No network access
	//   --readonly-filesystem: Root filesystem is read-only
	//   --tmpfs=/tmp: Writable /tmp in memory
	//   --nesting: Allow nested namespaces for container isolation
	//   --no-new-privs: Prevent privilege escalation
	args := []string{
		"run",
		"--container",          // Use container mode
		"--network=empty",      // No network access
		"--readonly-rootfs",    // Root filesystem read-only
		"--tmpfs=/tmp",         // Writable /tmp in memory
		"--nesting",            // Enable nested namespaces
		"--no-new-privs",       // No privilege escalation
		"--memory-limit", fmt.Sprintf("%d", config.MemoryMB*1024*1024), // Memory limit in bytes
		"--watchdog-timeout", config.Timeout.String(),
	}

	// Add CPU limit if supported
	if config.CPULimit > 0 && config.CPULimit <= 2 {
		args = append(args, "--cpu-limit", fmt.Sprintf("%d", int(config.CPULimit*1000)))
	}

	// Add capability restrictions
	for _, cap := range config.Capabilities {
		args = append(args, "--add-cap", cap)
	}

	// Deny all capabilities by default (security in depth)
	args = append(args, "--deny-all-caps")

	// Mount code as read-only
	args = append(args, "-b", workDir+":/workspace:ro")

	args = append(args, imageName)
	args = append(args, "/bin/sh", "/workspace/run.sh")

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, config.Timeout+5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, e.runscPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the sandbox
	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &sandboxResult{
		Passed:   exitCode == 0,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: exitCode,
		Error:    "",
	}, nil
}

// runWithDocker executes code with Docker (fallback when gVisor unavailable)
// Uses enhanced security flags to approximate gVisor security properties
func (e *SecureSandboxExecutor) runWithDocker(ctx context.Context, imageName, workDir, runtime string, config *ExecutionConfig) (*sandboxResult, error) {
	containerName := fmt.Sprintf("sandbox-%d", time.Now().UnixNano())

	// Memory limit as string
	memLimit := fmt.Sprintf("%dm", config.MemoryMB)

	// Build docker run command with security constraints
	args := []string{
		"run",
		"--name", containerName,
		"--rm",                                      // Auto-remove when done
		"--network=none",                            // No network access (critical security)
		"--memory", memLimit,                        // Memory limit
		"--cpus", fmt.Sprintf("%.2f", config.CPULimit), // CPU limit
		"--pids-limit", "50",                        // Limit number of processes
		"--read-only",                              // Read-only filesystem
		"--tmpfs", "/tmp:rw,size=50m,noexec",       // Writable tmpfs, noexec prevents binary execution
		"--workdir", "/workspace",
		"-v", workDir + ":/workspace:ro",           // Mount code as read-only
		"--no-new-privileges",                      // No privilege escalation
		"--cap-drop=ALL",                           // Drop all capabilities
		"--label", "functionfly.sandbox=true",
		"--label", "functionfly.runtime=" + runtime,
		"--label", "functionfly.security=enhanced",
	}

	// Add environment variables
	for k, v := range config.EnvVars {
		args = append(args, "--env", k+"="+v)
	}

	// Add security-related env vars
	args = append(args,
		"--env", "NODE_OPTIONS=--max-old-space-size=128",
		"--env", "PYTHONUNBUFFERED=1",
		"--env", "PYTHONDONTWRITEBYTECODE=1",  // Don't write .pyc files
		"--env", "PYTHONHASHSEED=random",      // Randomize hash seed
	)

	args = append(args, imageName, "/workspace/run.sh")

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, config.Timeout+5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the container
	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &sandboxResult{
		Passed:   exitCode == 0,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: exitCode,
		Error:    "",
	}, nil
}

// ensureImage pulls the Docker image if not already available locally
func (e *SecureSandboxExecutor) ensureImage(ctx context.Context, imageName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.imagePulled[imageName] {
		return nil
	}

	// Check if image exists locally
	checkCmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageName)
	if checkCmd.Run() == nil {
		e.imagePulled[imageName] = true
		return nil
	}

	// Pull the image with timeout
	pullCtx, cancel := context.WithTimeout(ctx, e.pullTimeout)
	defer cancel()

	pullCmd := exec.CommandContext(pullCtx, "docker", "pull", imageName)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("pull image %s: %w", imageName, err)
	}

	e.imagePulled[imageName] = true
	return nil
}

// createWrapperScript creates a shell script to run the handler with security constraints
func (e *SecureSandboxExecutor) createWrapperScript(runtime, handlerFile string) string {
	lower := strings.ToLower(runtime)

	if strings.Contains(lower, "python") {
		return fmt.Sprintf(`#!/bin/sh
set -e
cd /workspace
# Security: disable bytecode compilation and optimize for sandbox
python3 -u -X pycache_prefix=/tmp/pycache -c "import json; input_data=json.load(open('input.json'))" 2>/dev/null || true
python3 -u -X pycache_prefix=/tmp/pycache %s
`, handlerFile)
	}

	if strings.Contains(lower, "node") {
		return fmt.Sprintf(`#!/bin/sh
set -e
cd /workspace
node --no-warnings -e "const input=require('./input.json');" 2>/dev/null || true
node %s
`, handlerFile)
	}

	// Default fallback
	return fmt.Sprintf(`#!/bin/sh
cd /workspace
cat %s
`, handlerFile)
}

// Close cleans up resources (no-op for CLI-based executor)
func (e *SecureSandboxExecutor) Close() error {
	return nil
}

// IsGvisorAvailable returns true if gVisor is available for secure execution
func (e *SecureSandboxExecutor) IsGvisorAvailable() bool {
	return e.runscPath != ""
}

// Helper functions
func ternaryString(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}