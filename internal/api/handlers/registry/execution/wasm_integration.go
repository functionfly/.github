package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/wasm"
	"github.com/google/uuid"
)

// WASMExecutionConfig contains configuration for WASM execution
type WASMExecutionConfig struct {
	// EnableDeterministic enables deterministic execution mode
	EnableDeterministic bool

	// EnableStreaming enables streaming execution for large inputs
	EnableStreaming bool

	// EnableMetrics enables metrics collection
	EnableMetrics bool

	// EnableAudit enables audit logging
	EnableAudit bool
}

// DefaultWASMExecutionConfig returns default WASM execution configuration
func DefaultWASMExecutionConfig() *WASMExecutionConfig {
	return &WASMExecutionConfig{
		EnableDeterministic: false,
		EnableStreaming:     false,
		EnableMetrics:      true,
		EnableAudit:        true,
	}
}

// WASMExecutor handles WASM execution with advanced features
type WASMExecutor struct {
	pool        *wasm.InstancePool
	router      *wasm.RuntimeRouter
	auditLogger wasm.AuditLogger
	metrics     *wasm.MetricsRecorder
	config      *WASMExecutionConfig
}

// NewWASMExecutor creates a new WASM executor with advanced features
func NewWASMExecutor(pool *wasm.InstancePool, router *wasm.RuntimeRouter, auditLogger wasm.AuditLogger, config *WASMExecutionConfig) *WASMExecutor {
	if config == nil {
		config = DefaultWASMExecutionConfig()
	}

	return &WASMExecutor{
		pool:        pool,
		router:      router,
		auditLogger: auditLogger,
		config:      config,
	}
}

// ExecuteResult contains the result of a WASM execution
type ExecuteResult struct {
	Output           json.RawMessage
	ExecutionTimeMs  int64
	Status           string
	ErrorMessage     string
	MemoryUsed       uint64
	DeterministicID  string
}

// Execute executes a WASM function with the given input
func (e *WASMExecutor) Execute(ctx context.Context, tenantID, functionID uuid.UUID, runtimeType string, input json.RawMessage) (*ExecuteResult, error) {
	startTime := time.Now()

	// Get or determine runtime type
	rt := wasm.RuntimeTypeFromString(runtimeType)
	if rt == wasm.RuntimeUnknown {
		rt = wasm.RuntimePythonWASM // Default to Python WASM
	}

	// Create metrics recorder if enabled
	var recorder *wasm.MetricsRecorder
	if e.config.EnableMetrics {
		recorder = wasm.NewMetricsRecorder(rt.String(), tenantID.String())
	}

	// Get instance from pool
	inst, err := e.pool.Get(ctx, tenantID.String(), rt.String())
	if err != nil {
		return e.handleError(ctx, tenantID, functionID, rt.String(), input, startTime, "pool_get_failed", err, recorder)
	}

	// Execute with context
	output, execErr := inst.Instance.ExecuteWithContext(ctx, input)

	// Calculate execution time
	executionTimeMs := time.Since(startTime).Milliseconds()

	// Get memory usage
	memoryUsed := inst.Instance.GetMemoryUsage()

	// Return instance to pool
	e.pool.Put(inst)

	// Handle execution error
	if execErr != nil {
		return e.handleError(ctx, tenantID, functionID, rt.String(), input, startTime, "execution_failed", execErr, recorder)
	}

	// Record success metrics
	if recorder != nil {
		recorder.RecordExecutionWithSizes(
			time.Since(startTime),
			wasm.StatusSuccess,
			len(input),
			len(output),
		)
		recorder.RecordMemoryUsage("default", memoryUsed)
	}

	// Log audit if enabled
	if e.config.EnableAudit && e.auditLogger != nil {
		audit := &wasm.ExecutionAudit{
			TenantID:        tenantID,
			FunctionID:      functionID,
			ExecutionID:     uuid.New(),
			Runtime:         rt.String(),
			InputSize:       len(input),
			OutputSize:      len(output),
			ExecutionTimeMs: executionTimeMs,
			Status:          wasm.StatusSuccess,
			MemoryUsed:      memoryUsed,
			CreatedAt:       time.Now(),
		}
		e.auditLogger.LogExecution(ctx, audit)
	}

	return &ExecuteResult{
		Output:          output,
		ExecutionTimeMs: executionTimeMs,
		Status:          wasm.StatusSuccess,
		MemoryUsed:      memoryUsed,
	}, nil
}

// ExecuteDeterministic executes a WASM function with deterministic mode using DeterministicExecutor.
func (e *WASMExecutor) ExecuteDeterministic(ctx context.Context, tenantID, functionID uuid.UUID, runtimeType string, input json.RawMessage, detConfig *wasm.DeterministicConfig) (*ExecuteResult, error) {
	if detConfig == nil {
		detConfig = wasm.DefaultDeterministicConfig()
	}

	rt := wasm.RuntimeTypeFromString(runtimeType)
	if rt == wasm.RuntimeUnknown {
		rt = wasm.RuntimePythonWASM
	}

	var recorder *wasm.MetricsRecorder
	if e.config.EnableMetrics {
		recorder = wasm.NewMetricsRecorder(rt.String(), tenantID.String())
		recorder.RecordDeterministicExecution()
	}

	detExecutor := wasm.NewDeterministicExecutor(e.pool, detConfig)
	execIDSuffix := "none"
	if len(input) >= 8 {
		execIDSuffix = fmt.Sprintf("%x", input[:8])
	} else if len(input) > 0 {
		execIDSuffix = fmt.Sprintf("%x", input)
	}
	executionID := fmt.Sprintf("%s-%s", functionID.String(), execIDSuffix)
	res, err := detExecutor.Execute(ctx, tenantID.String(), functionID.String(), executionID, rt.String(), input)
	if err != nil {
		startTime := time.Now()
		if res != nil {
			startTime = time.Now().Add(-res.ExecutionTime)
		}
		return e.handleError(ctx, tenantID, functionID, rt.String(), input, startTime, "execution_failed", err, recorder)
	}

	if recorder != nil {
		recorder.RecordExecutionWithSizes(res.ExecutionTime, res.Status, len(input), len(res.Output))
	}

	return &ExecuteResult{
		Output:          res.Output,
		ExecutionTimeMs: res.ExecutionTime.Milliseconds(),
		Status:          res.Status,
		MemoryUsed:      0,
		DeterministicID: res.DeterministicID,
	}, nil
}

// ExecuteStreaming executes a WASM function with streaming for large inputs
func (e *WASMExecutor) ExecuteStreaming(ctx context.Context, tenantID, functionID uuid.UUID, runtimeType string, input json.RawMessage) (*ExecuteResult, error) {
	rt := wasm.RuntimeTypeFromString(runtimeType)
	if rt == wasm.RuntimeUnknown {
		rt = wasm.RuntimePythonWASM
	}

	// Use streaming config
	streamingConfig := wasm.DefaultStreamingConfig()

	// Get instance from pool
	inst, err := e.pool.Get(ctx, tenantID.String(), rt.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Execute with large input handling
	result, execErr := inst.Instance.ExecuteLargeInput(ctx, input, streamingConfig)

	e.pool.Put(inst)

	if execErr != nil {
		return nil, fmt.Errorf("streaming execution failed: %w", execErr)
	}

	return &ExecuteResult{
		Output:          result.Output,
		ExecutionTimeMs: result.Duration.Milliseconds(),
		Status:          wasm.StatusSuccess,
	}, nil
}

// handleError handles execution errors
func (e *WASMExecutor) handleError(ctx context.Context, tenantID, functionID uuid.UUID, runtimeType string, input []byte, startTime time.Time, errorType string, err error, recorder *wasm.MetricsRecorder) (*ExecuteResult, error) {
	executionTimeMs := time.Since(startTime).Milliseconds()

	// Record error metrics
	if recorder != nil {
		recorder.RecordError(errorType)
		recorder.RecordExecution(time.Since(startTime), wasm.StatusError)
	}

	// Log audit if enabled
	if e.config.EnableAudit && e.auditLogger != nil {
		audit := &wasm.ExecutionAudit{
			TenantID:        tenantID,
			FunctionID:      functionID,
			ExecutionID:     uuid.New(),
			Runtime:         runtimeType,
			InputSize:       len(input),
			ExecutionTimeMs: executionTimeMs,
			Status:          wasm.StatusError,
			ErrorMessage:    err.Error(),
			CreatedAt:       time.Now(),
		}
		e.auditLogger.LogExecution(ctx, audit)
	}

	return &ExecuteResult{
		Output:          nil,
		ExecutionTimeMs: executionTimeMs,
		Status:          wasm.StatusError,
		ErrorMessage:    err.Error(),
	}, err
}

// ExecuteWASMWithIntegration runs the function via the WASM executor when non-nil,
// otherwise falls back to executeLocallyWithLimits for compatibility with existing handlers.
func ExecuteWASMWithIntegration(
	fnVersion *storage.RegistryFunctionVersion,
	input json.RawMessage,
	maxMemoryMB, maxCPUTimeMs int,
	executor *WASMExecutor,
	tenantID, functionID uuid.UUID,
) (json.RawMessage, error) {
	if executor == nil {
		// Fall back to original execution
		return executeLocallyWithLimits(fnVersion, input, maxMemoryMB, maxCPUTimeMs)
	}

	// Determine runtime type from function version
	runtimeType := determineRuntimeType(fnVersion)

	// Execute with integration
	result, err := executor.Execute(context.Background(), tenantID, functionID, runtimeType, input)
	if err != nil {
		return nil, err
	}

	return result.Output, nil
}

// determineRuntimeType returns the runtime type from the function version (stored on the version record).
func determineRuntimeType(fnVersion *storage.RegistryFunctionVersion) string {
	if fnVersion.Runtime != "" {
		return fnVersion.Runtime
	}

	// Default to Python WASM for backward compatibility
	return string(wasm.RuntimePythonWASM)
}
