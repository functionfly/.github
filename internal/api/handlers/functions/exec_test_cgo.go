//go:build cgo

package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/flypy"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/wasm"
)

func executeTestFunctionImpl(ctx context.Context, h *Handler, req *types.TestFunctionRequest, user *storage.User) (*types.TestFunctionResponse, error) {
	startTime := time.Now()

	var functionCode string
	var functionName string

	if req.FunctionId != nil {
		function, err := h.repo.GetFunctionByID(ctx, *req.FunctionId)
		if err != nil {
			return nil, fmt.Errorf("failed to get function: %w", err)
		}

		if function.TenantID != user.TenantID {
			return nil, fmt.Errorf("function does not belong to user")
		}

		functionCode = function.Code
		functionName = function.Name
	} else {
		functionCode = req.Input
		functionName = "test-function"
	}

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("test-function-%d", time.Now().Unix()))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	config := &flypy.Config{
		Mode:      flypy.CompatibleMode,
		OutputDir: tempDir,
		Verbose:   false,
	}

	compiler := flypy.NewCompiler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create compiler: %w", err)
	}

	compileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := compiler.Compile(compileCtx, functionCode, functionName)
	if err != nil {
		return &types.TestFunctionResponse{
			Success:         false,
			Output:          nil,
			ExecutionTimeMs: int(time.Since(startTime).Milliseconds()),
			Logs:            []*storage.FunctionLog{{Message: fmt.Sprintf("Compilation failed: %v", err)}},
		}, nil
	}

	if len(result.Warnings) > 0 {
		logs := make([]*storage.FunctionLog, len(result.Warnings))
		for i, warning := range result.Warnings {
			logs[i] = &storage.FunctionLog{Message: fmt.Sprintf("Warning: %s", warning)}
		}
		return &types.TestFunctionResponse{
			Success:         false,
			Output:          nil,
			ExecutionTimeMs: int(time.Since(startTime).Milliseconds()),
			Logs:            logs,
		}, nil
	}

	wasmPath := filepath.Join(tempDir, "state_transition.wasm")

	var stdoutBuf, stderrBuf bytes.Buffer

	runtime, err := wasm.NewPythonRuntimeWithDebug(wasmPath, &stdoutBuf, &stderrBuf, nil, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}
	defer runtime.Close()

	if err := runtime.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize runtime: %w", err)
	}

	if err := runtime.LoadCode(functionCode); err != nil {
		return nil, fmt.Errorf("failed to load code: %w", err)
	}

	var inputData []byte
	if req.FunctionId != nil {
		inputData = []byte(req.Input)
	} else {
		inputData = []byte("{}")
	}

	output, err := runtime.Execute(inputData)
	executionTime := time.Since(startTime)

	var response *types.TestFunctionResponse
	if err != nil {
		response = &types.TestFunctionResponse{
			Success:         false,
			Output:          nil,
			ExecutionTimeMs: int(executionTime.Milliseconds()),
			Logs:            []*storage.FunctionLog{{Message: fmt.Sprintf("Execution failed: %v", err)}},
		}
	} else {
		var parsedOutput interface{}
		if err := json.Unmarshal(output, &parsedOutput); err != nil {
			parsedOutput = string(output)
		}

		response = &types.TestFunctionResponse{
			Success:         true,
			Output:          parsedOutput,
			ExecutionTimeMs: int(executionTime.Milliseconds()),
			Logs:            []*storage.FunctionLog{},
		}
	}

	if stdoutBuf.Len() > 0 {
		response.Logs = append(response.Logs, &storage.FunctionLog{
			Message: fmt.Sprintf("stdout: %s", stdoutBuf.String()),
		})
	}
	if stderrBuf.Len() > 0 {
		response.Logs = append(response.Logs, &storage.FunctionLog{
			Message: fmt.Sprintf("stderr: %s", stderrBuf.String()),
		})
	}

	return response, nil
}