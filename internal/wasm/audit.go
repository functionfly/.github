// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains execution audit types (works with and without CGO)
package wasm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Execution status constants
const (
	StatusSuccess    = "success"
	StatusError     = "error"
	StatusTimeout   = "timeout"
	StatusMemoryLimit = "memory_limit"
	StatusTerminated = "terminated"
	StatusInvalidInput = "invalid_input"
)

// ExecutionAudit represents a single WASM execution audit record
type ExecutionAudit struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	FunctionID     uuid.UUID `json:"function_id"`
	ExecutionID    uuid.UUID `json:"execution_id"`
	Runtime        string    `json:"runtime"`
	InputSize      int       `json:"input_size"`
	OutputSize     int       `json:"output_size"`
	ExecutionTimeMs int64    `json:"execution_time_ms"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	MemoryUsed     uint64    `json:"memory_used,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// AuditLogger handles execution audit logging
type AuditLogger interface {
	LogExecution(ctx context.Context, audit *ExecutionAudit) error
	GetExecution(ctx context.Context, executionID uuid.UUID) (*ExecutionAudit, error)
	ListExecutions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]ExecutionAudit, error)
}

// InMemoryAuditLogger provides in-memory audit logging (for testing/dev)
type InMemoryAuditLogger struct {
	mu        sync.RWMutex
	executions map[uuid.UUID]*ExecutionAudit
	byTenant  map[uuid.UUID][]uuid.UUID
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger() *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		executions: make(map[uuid.UUID]*ExecutionAudit),
		byTenant:   make(map[uuid.UUID][]uuid.UUID),
	}
}

// LogExecution logs an execution to memory
func (l *InMemoryAuditLogger) LogExecution(ctx context.Context, audit *ExecutionAudit) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if audit.ID == uuid.Nil {
		audit.ID = uuid.New()
	}
	if audit.ExecutionID == uuid.Nil {
		audit.ExecutionID = uuid.New()
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = time.Now()
	}

	// Store by both ID and ExecutionID for lookups
	l.executions[audit.ID] = audit
	l.executions[audit.ExecutionID] = audit
	l.byTenant[audit.TenantID] = append(l.byTenant[audit.TenantID], audit.ID)

	return nil
}

// GetExecution retrieves an execution by ID
func (l *InMemoryAuditLogger) GetExecution(ctx context.Context, executionID uuid.UUID) (*ExecutionAudit, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, audit := range l.executions {
		if audit.ExecutionID == executionID {
			return audit, nil
		}
	}
	return nil, fmt.Errorf("execution not found: %s", executionID)
}

// ListExecutions lists executions for a tenant
func (l *InMemoryAuditLogger) ListExecutions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]ExecutionAudit, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	ids, exists := l.byTenant[tenantID]
	if !exists {
		return []ExecutionAudit{}, nil
	}

	if offset >= len(ids) {
		return []ExecutionAudit{}, nil
	}

	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}

	executions := make([]ExecutionAudit, 0, end-offset)
	for i := offset; i < end; i++ {
		if audit, exists := l.executions[ids[i]]; exists {
			executions = append(executions, *audit)
		}
	}

	return executions, nil
}

// DBAuditLogger provides database-backed audit logging
type DBAuditLogger struct {
	db *sql.DB
}

// NewDBAuditLogger creates a new database audit logger
func NewDBAuditLogger(db *sql.DB) *DBAuditLogger {
	return &DBAuditLogger{db: db}
}

// LogExecution logs an execution to the database
func (l *DBAuditLogger) LogExecution(ctx context.Context, audit *ExecutionAudit) error {
	if audit.ID == uuid.Nil {
		audit.ID = uuid.New()
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO wasm_execution_audit (
			id, tenant_id, function_id, execution_id, runtime,
			input_size, output_size, execution_time_ms, status,
			error_message, memory_used, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := l.db.ExecContext(ctx, query,
		audit.ID,
		audit.TenantID,
		audit.FunctionID,
		audit.ExecutionID,
		audit.Runtime,
		audit.InputSize,
		audit.OutputSize,
		audit.ExecutionTimeMs,
		audit.Status,
		audit.ErrorMessage,
		audit.MemoryUsed,
		audit.CreatedAt,
	)

	return err
}

// GetExecution retrieves an execution by ID from the database
func (l *DBAuditLogger) GetExecution(ctx context.Context, executionID uuid.UUID) (*ExecutionAudit, error) {
	query := `
		SELECT id, tenant_id, function_id, execution_id, runtime,
		       input_size, output_size, execution_time_ms, status,
		       error_message, memory_used, created_at
		FROM wasm_execution_audit
		WHERE execution_id = $1
	`

	audit := &ExecutionAudit{}
	err := l.db.QueryRowContext(ctx, query, executionID).Scan(
		&audit.ID,
		&audit.TenantID,
		&audit.FunctionID,
		&audit.ExecutionID,
		&audit.Runtime,
		&audit.InputSize,
		&audit.OutputSize,
		&audit.ExecutionTimeMs,
		&audit.Status,
		&audit.ErrorMessage,
		&audit.MemoryUsed,
		&audit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return audit, nil
}

// ListExecutions lists executions for a tenant from the database
func (l *DBAuditLogger) ListExecutions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]ExecutionAudit, error) {
	query := `
		SELECT id, tenant_id, function_id, execution_id, runtime,
		       input_size, output_size, execution_time_ms, status,
		       error_message, memory_used, created_at
		FROM wasm_execution_audit
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := l.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []ExecutionAudit
	for rows.Next() {
		var audit ExecutionAudit
		err := rows.Scan(
			&audit.ID,
			&audit.TenantID,
			&audit.FunctionID,
			&audit.ExecutionID,
			&audit.Runtime,
			&audit.InputSize,
			&audit.OutputSize,
			&audit.ExecutionTimeMs,
			&audit.Status,
			&audit.ErrorMessage,
			&audit.MemoryUsed,
			&audit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		executions = append(executions, audit)
	}

	return executions, rows.Err()
}

// ExecutionRecorder wraps an audit logger and provides convenient recording methods
type ExecutionRecorder struct {
	logger   AuditLogger
	tenantID uuid.UUID
	functionID uuid.UUID
	executionID uuid.UUID
	runtime  string
}

// NewExecutionRecorder creates a new execution recorder
func NewExecutionRecorder(logger AuditLogger, tenantID, functionID uuid.UUID, runtime string) *ExecutionRecorder {
	return &ExecutionRecorder{
		logger:     logger,
		tenantID:   tenantID,
		functionID: functionID,
		executionID: uuid.New(),
		runtime:   runtime,
	}
}

// Start begins recording an execution
func (r *ExecutionRecorder) Start() *ExecutionStart {
	return &ExecutionStart{
		recorder:    r,
		startTime:   time.Now(),
		inputSize:   0,
		outputSize:  0,
		memoryUsed:  0,
	}
}

// ExecutionStart holds execution metrics during execution
type ExecutionStart struct {
	recorder    *ExecutionRecorder
	startTime   time.Time
	inputSize   int
	outputSize  int
	memoryUsed  uint64
}

// WithInputSize sets the input size
func (s *ExecutionStart) WithInputSize(size int) *ExecutionStart {
	s.inputSize = size
	return s
}

// WithMemoryUsed sets the memory used
func (s *ExecutionStart) WithMemoryUsed(mem uint64) *ExecutionStart {
	s.memoryUsed = mem
	return s
}

// End completes the execution recording with success
func (s *ExecutionStart) End(output []byte) error {
	elapsed := time.Since(s.startTime)

	audit := &ExecutionAudit{
		ID:              s.recorder.executionID,
		TenantID:        s.recorder.tenantID,
		FunctionID:      s.recorder.functionID,
		ExecutionID:     s.recorder.executionID,
		Runtime:         s.recorder.runtime,
		InputSize:       s.inputSize,
		OutputSize:      len(output),
		ExecutionTimeMs: elapsed.Milliseconds(),
		Status:          StatusSuccess,
		MemoryUsed:      s.memoryUsed,
		CreatedAt:       time.Now(),
	}

	return s.recorder.logger.LogExecution(context.Background(), audit)
}

// EndWithError completes the execution recording with an error
func (s *ExecutionStart) EndWithError(err error) error {
	elapsed := time.Since(s.startTime)

	audit := &ExecutionAudit{
		ID:              s.recorder.executionID,
		TenantID:       s.recorder.tenantID,
		FunctionID:     s.recorder.functionID,
		ExecutionID:    s.recorder.executionID,
		Runtime:        s.recorder.runtime,
		InputSize:      s.inputSize,
		OutputSize:     0,
		ExecutionTimeMs: elapsed.Milliseconds(),
		Status:         StatusError,
		ErrorMessage:   err.Error(),
		MemoryUsed:     s.memoryUsed,
		CreatedAt:      time.Now(),
	}

	return s.recorder.logger.LogExecution(context.Background(), audit)
}

// EndWithStatus completes the execution recording with a specific status
func (s *ExecutionStart) EndWithStatus(status string, output []byte, err error) error {
	elapsed := time.Since(s.startTime)

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	audit := &ExecutionAudit{
		ID:              s.recorder.executionID,
		TenantID:       s.recorder.tenantID,
		FunctionID:     s.recorder.functionID,
		ExecutionID:    s.recorder.executionID,
		Runtime:        s.recorder.runtime,
		InputSize:      s.inputSize,
		OutputSize:     len(output),
		ExecutionTimeMs: elapsed.Milliseconds(),
		Status:         status,
		ErrorMessage:   errorMsg,
		MemoryUsed:     s.memoryUsed,
		CreatedAt:      time.Now(),
	}

	return s.recorder.logger.LogExecution(context.Background(), audit)
}

// globalAuditLogger is the global audit logger
var globalAuditLogger AuditLogger
var auditLoggerOnce sync.Once

// InitAuditLogger initializes the global audit logger
func InitAuditLogger(logger AuditLogger) {
	auditLoggerOnce.Do(func() {
		globalAuditLogger = logger
	})
}

// GetAuditLogger returns the global audit logger
func GetAuditLogger() AuditLogger {
	if globalAuditLogger == nil {
		// Fallback to in-memory logger if not initialized
		globalAuditLogger = NewInMemoryAuditLogger()
	}
	return globalAuditLogger
}

// LogExecution is a convenience function for logging execution
func LogExecution(ctx context.Context, audit *ExecutionAudit) error {
	return GetAuditLogger().LogExecution(ctx, audit)
}

// CreateExecutionAudit creates a new execution audit record
func CreateExecutionAudit(tenantID, functionID, executionID uuid.UUID, runtime string) *ExecutionRecorder {
	return NewExecutionRecorder(GetAuditLogger(), tenantID, functionID, runtime)
}

// MarshalJSON implements custom JSON marshaling
func (a *ExecutionAudit) MarshalJSON() ([]byte, error) {
	type Alias ExecutionAudit
	return json.Marshal(&struct {
		Alias
		CreatedAt string `json:"created_at"`
	}{
		Alias:     Alias(*a),
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	})
}

// CreateAuditMigration generates SQL for creating the audit table
func CreateAuditMigration() string {
	return `
-- Create WASM execution audit table
CREATE TABLE IF NOT EXISTS wasm_execution_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id UUID NOT NULL,
    execution_id UUID NOT NULL,
    runtime VARCHAR(50) NOT NULL,
    input_size INTEGER,
    output_size INTEGER,
    execution_time_ms INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    memory_used BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for querying by tenant
CREATE INDEX IF NOT EXISTS idx_wasm_audit_tenant ON wasm_execution_audit(tenant_id, created_at DESC);

-- Index for querying by function
CREATE INDEX IF NOT EXISTS idx_wasm_audit_function ON wasm_execution_audit(function_id, created_at DESC);

-- Index for querying by execution
CREATE INDEX IF NOT EXISTS idx_wasm_audit_execution ON wasm_execution_audit(execution_id);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_wasm_audit_status ON wasm_execution_audit(status, created_at DESC);
`
}

// LogWASMExecution logs a WASM execution with all relevant details
func LogWASMExecution(
	ctx context.Context,
	tenantID, functionID, executionID uuid.UUID,
	runtime string,
	inputSize, outputSize int,
	execTimeMs int64,
	status, errorMsg string,
	memUsed uint64,
) error {
	audit := &ExecutionAudit{
		ID:              uuid.New(),
		TenantID:        tenantID,
		FunctionID:      functionID,
		ExecutionID:     executionID,
		Runtime:         runtime,
		InputSize:       inputSize,
		OutputSize:      outputSize,
		ExecutionTimeMs: execTimeMs,
		Status:          status,
		ErrorMessage:    errorMsg,
		MemoryUsed:      memUsed,
		CreatedAt:       time.Now(),
	}

	return LogExecution(ctx, audit)
}

// LogExecutionStart logs the start of a WASM execution (helper for debugging)
func LogExecutionStart(tenantID, functionID, executionID uuid.UUID, runtime string, inputSize int) {
	log.Printf("[WASM Audit] Execution started: tenant=%s function=%s execution=%s runtime=%s input_size=%d",
		tenantID, functionID, executionID, runtime, inputSize)
}

// LogExecutionEnd logs the end of a WASM execution (helper for debugging)
func LogExecutionEnd(executionID uuid.UUID, status string, execTimeMs int64, outputSize int, err error) {
	if err != nil {
		log.Printf("[WASM Audit] Execution completed: execution=%s status=%s time=%dms error=%s",
			executionID, status, execTimeMs, err.Error())
	} else {
		log.Printf("[WASM Audit] Execution completed: execution=%s status=%s time=%dms output_size=%d",
			executionID, status, execTimeMs, outputSize)
	}
}
