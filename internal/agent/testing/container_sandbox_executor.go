package testing

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
)

// ContainerSandboxExecutor runs generated code in isolated Docker containers
// using the Docker CLI for reliability and simplicity
type ContainerSandboxExecutor struct {
	mu           sync.Mutex
	imagePulled  map[string]bool
	pullTimeout  time.Duration
	execTimeout  time.Duration
	defaultMemMB int64
	defaultCPU   float64
}

// NewContainerSandboxExecutor creates a new container-based sandbox executor
func NewContainerSandboxExecutor() (*ContainerSandboxExecutor, error) {
	return &ContainerSandboxExecutor{
		imagePulled:  make(map[string]bool),
		pullTimeout:  2 * time.Minute,
		execTimeout:  30 * time.Second,
		defaultMemMB: 256,
		defaultCPU:   0.5,
	}, nil
}

// Runtime configuration for different languages
var runtimeConfigs = map[string]struct {
	image       string
	entryPoint  string
	handlerFile string
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
}

// Execute runs the code in an isolated container
func (e *ContainerSandboxExecutor) Execute(ctx context.Context, runtime, code string, input map[string]any) (SandboxExecutionResult, error) {
	start := time.Now()

	// Get runtime configuration
	config, ok := runtimeConfigs[strings.ToLower(runtime)]
	if !ok {
		// Fallback to python for unknown runtimes
		config = runtimeConfigs["python"]
	}

	// Ensure image is available
	if err := e.ensureImage(ctx, config.image); err != nil {
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

	// Write handler code
	handlerPath := filepath.Join(workDir, config.handlerFile)
	if err := os.WriteFile(handlerPath, []byte(code), 0644); err != nil {
		return SandboxExecutionResult{}, fmt.Errorf("write handler: %w", err)
	}

	// Write input as JSON
	inputJSON, _ := json.Marshal(input)
	inputPath := filepath.Join(workDir, "input.json")
	if err := os.WriteFile(inputPath, inputJSON, 0644); err != nil {
		return SandboxExecutionResult{}, fmt.Errorf("write input: %w", err)
	}

	// Create wrapper script that runs the handler
	wrapperScript := e.createWrapperScript(runtime, config.handlerFile)
	wrapperPath := filepath.Join(workDir, "run.sh")
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0755); err != nil {
		return SandboxExecutionResult{}, fmt.Errorf("write wrapper: %w", err)
	}

	// Run container and get result
	result, err := e.runContainer(ctx, config.image, workDir, runtime)
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
			"runtime":      runtime,
			"stdout":       result.Stdout,
			"stderr":       result.Stderr,
			"exitCode":     result.ExitCode,
			"memoryUsedMB": result.MemoryUsedMB,
		},
		Error: result.Error,
	}, nil
}

// containerResult holds the execution result from a container
type containerResult struct {
	Passed       bool
	Stdout       string
	Stderr       string
	ExitCode     int
	Error        string
	MemoryUsedMB int64
}

// ensureImage pulls the Docker image if not already available locally
func (e *ContainerSandboxExecutor) ensureImage(ctx context.Context, imageName string) error {
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

// createWrapperScript creates a shell script to run the handler
func (e *ContainerSandboxExecutor) createWrapperScript(runtime, handlerFile string) string {
	lower := strings.ToLower(runtime)

	if strings.Contains(lower, "python") {
		return fmt.Sprintf(`#!/bin/sh
set -e
cd /workspace
python3 -c "import json; input_data=json.load(open('input.json'))" 2>/dev/null || true
python3 %s
`, handlerFile)
	}

	if strings.Contains(lower, "node") {
		return fmt.Sprintf(`#!/bin/sh
set -e
cd /workspace
node -e "const input=require('./input.json');" 2>/dev/null || true
node %s
`, handlerFile)
	}

	// Default fallback
	return fmt.Sprintf(`#!/bin/sh
cd /workspace
cat %s
`, handlerFile)
}

// runContainer creates and executes a container with the code
func (e *ContainerSandboxExecutor) runContainer(ctx context.Context, imageName, workDir, runtime string) (*containerResult, error) {
	// Container name for tracking
	containerName := fmt.Sprintf("sandbox-%d", time.Now().UnixNano())

	// Memory limit as string
	memLimit := fmt.Sprintf("%dm", e.defaultMemMB)

	// Build docker run command with security constraints
	args := []string{
		"run",
		"--name", containerName,
		"--rm",              // Auto-remove when done
		"--network", "none", // No network access for security
		"--memory", memLimit, // Memory limit
		"--cpus", fmt.Sprintf("%.2f", e.defaultCPU), // CPU limit
		"--read-only",                 // Read-only filesystem
		"--tmpfs", "/tmp:rw,size=50m", // Writable tmpfs
		"--workdir", "/workspace",
		"-v", workDir + ":/workspace:ro", // Mount code as read-only
		"--env", "NODE_OPTIONS=--max-old-space-size=128",
		"--env", "PYTHONUNBUFFERED=1",
		"--label", "functionfly.sandbox=true",
		"--label", "functionfly.runtime=" + runtime,
		imageName,
		"/workspace/run.sh",
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.execTimeout)
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

	return &containerResult{
		Passed:   exitCode == 0,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: exitCode,
		Error:    "",
	}, nil
}

// Close implements the io.Closer interface - no-op for CLI-based executor
func (e *ContainerSandboxExecutor) Close() error {
	return nil
}
