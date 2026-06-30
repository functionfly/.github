// Package wasm provides WebAssembly runtime support for FunctionFly
// Tests for security functions (works with and without CGO)
package wasm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestWASMConfigDefaults tests that default security config values are correct
func TestWASMConfigDefaults(t *testing.T) {
	config := NewDefaultSecurityConfig()

	if config.MaxMemory != DefaultMaxMemory {
		t.Errorf("Expected MaxMemory=%d, got %d", DefaultMaxMemory, config.MaxMemory)
	}

	if config.MaxExecutionTime != DefaultMaxExecutionTime {
		t.Errorf("Expected MaxExecutionTime=%v, got %v", DefaultMaxExecutionTime, config.MaxExecutionTime)
	}

	if config.MaxInstructions != DefaultMaxInstructions {
		t.Errorf("Expected MaxInstructions=%d, got %d", DefaultMaxInstructions, config.MaxInstructions)
	}

	if config.PoolSize != DefaultPoolSize {
		t.Errorf("Expected PoolSize=%d, got %d", DefaultPoolSize, config.PoolSize)
	}

	if config.EnableWASI != DefaultWASIEnabled {
		t.Errorf("Expected EnableWASI=%v, got %v", DefaultWASIEnabled, config.EnableWASI)
	}

	if config.AllowRawPointers != DefaultAllowRawPointers {
		t.Errorf("Expected AllowRawPointers=%v, got %v", DefaultAllowRawPointers, config.AllowRawPointers)
	}

	if !config.InstancePoolPerTenant {
		t.Error("Expected InstancePoolPerTenant=true by default")
	}
}

// TestWASMConfigValidateInputSize tests input size validation
func TestWASMConfigValidateInputSize(t *testing.T) {
	config := NewDefaultSecurityConfig()

	// Test valid input sizes
	tests := []struct {
		size    uint32
		want    bool
	}{
		{0, true},
		{1024, true},
		{1024 * 1024, true},        // 1MB - default max
		{1024 * 1024 + 1, false},   // 1MB + 1 byte - should fail
		{10 * 1024 * 1024, false},  // 10MB - should fail
	}

	for _, tt := range tests {
		got := config.ValidateInputSize(tt.size)
		if got != tt.want {
			t.Errorf("ValidateInputSize(%d) = %v, want %v", tt.size, got, tt.want)
		}
	}
}

// TestWASMConfigValidateOutputSize tests output size validation
func TestWASMConfigValidateOutputSize(t *testing.T) {
	config := NewDefaultSecurityConfig()

	// Test valid output sizes
	tests := []struct {
		size    uint32
		want    bool
	}{
		{0, true},
		{1024, true},
		{1024 * 1024, true},        // 1MB - default max
		{1024 * 1024 + 1, false},   // 1MB + 1 byte - should fail
		{10 * 1024 * 1024, false},  // 10MB - should fail
	}

	for _, tt := range tests {
		got := config.ValidateOutputSize(tt.size)
		if got != tt.want {
			t.Errorf("ValidateOutputSize(%d) = %v, want %v", tt.size, got, tt.want)
		}
	}
}

// TestWASMConfigIsDomainAllowed tests domain allowlist functionality
func TestWASMConfigIsDomainAllowed(t *testing.T) {
	config := NewDefaultSecurityConfig()

	// Empty list should DENY all domains (default-deny)
	config.AllowedDomains = []string{}
	if config.IsDomainAllowed("any-domain.com") {
		t.Error("Expected all domains to be DENIED when AllowedDomains is empty (default-deny)")
	}

	// Explicit wildcard should allow all domains
	config.AllowedDomains = []string{"*"}
	if !config.IsDomainAllowed("any-domain.com") {
		t.Error("Expected all domains to be allowed when AllowedDomains contains '*'")
	}

	// Single domain
	config.AllowedDomains = []string{"api.example.com"}
	tests := []struct {
		domain string
		want   bool
	}{
		{"api.example.com", true},
		{"api.example.com.", false}, // trailing dot not supported in this simple implementation
		{"sub.api.example.com", true},
		{"other.com", false},
		{"example.com", false},
		{"fake-api.example.com", false},
	}

	for _, tt := range tests {
		got := config.IsDomainAllowed(tt.domain)
		if got != tt.want {
			t.Errorf("IsDomainAllowed(%q) = %v, want %v", tt.domain, got, tt.want)
		}
	}

	// Multiple domains with wildcard matching
	config.AllowedDomains = []string{"api.example.com", "cdn.jsdelivr.net"}
	domainTests := []struct {
		domain string
		want   bool
	}{
		{"api.example.com", true},
		{"cdn.jsdelivr.net", true},
		{"sub.cdn.jsdelivr.net", true},
		{"other.com", false},
		{"api.example.com.", false},
	}

	for _, tt := range domainTests {
		got := config.IsDomainAllowed(tt.domain)
		if got != tt.want {
			t.Errorf("IsDomainAllowed(%q) = %v, want %v", tt.domain, got, tt.want)
		}
	}
}

// TestWASMConfigClone tests config cloning
func TestWASMConfigClone(t *testing.T) {
	config := NewDefaultSecurityConfig()
	config.AllowedDomains = []string{"example.com"}
	config.MaxMemory = 128 * 1024 * 1024
	config.MaxExecutionTime = 60 * time.Second

	clone := config.Clone()

	// Verify values are copied
	if clone.MaxMemory != config.MaxMemory {
		t.Errorf("Clone MaxMemory = %d, want %d", clone.MaxMemory, config.MaxMemory)
	}
	if clone.MaxExecutionTime != config.MaxExecutionTime {
		t.Errorf("Clone MaxExecutionTime = %v, want %v", clone.MaxExecutionTime, config.MaxExecutionTime)
	}
	if len(clone.AllowedDomains) != len(config.AllowedDomains) {
		t.Errorf("Clone AllowedDomains length = %d, want %d", len(clone.AllowedDomains), len(config.AllowedDomains))
	}

	// Verify it's a deep copy
	clone.AllowedDomains[0] = "modified.com"
	if config.AllowedDomains[0] == "modified.com" {
		t.Error("Clone should be independent from original")
	}
}

// TestWASMConfigFromEnv tests environment variable loading
func TestWASMConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("WASM_MAX_MEMORY", "128")
	os.Setenv("WASM_MAX_TIMEOUT", "60s")
	os.Setenv("WASM_POOL_SIZE", "20")
	os.Setenv("WASM_ENABLE_DETERMINISTIC", "true")
	os.Setenv("WASM_ENABLE_WASI", "false")
	os.Setenv("WASM_ALLOWED_DOMAINS", "api.example.com,cdn.jsdelivr.net")
	os.Setenv("WASM_MAX_INPUT_SIZE", "2")
	os.Setenv("WASM_MAX_OUTPUT_SIZE", "4")

	defer func() {
		os.Unsetenv("WASM_MAX_MEMORY")
		os.Unsetenv("WASM_MAX_TIMEOUT")
		os.Unsetenv("WASM_POOL_SIZE")
		os.Unsetenv("WASM_ENABLE_DETERMINISTIC")
		os.Unsetenv("WASM_ENABLE_WASI")
		os.Unsetenv("WASM_ALLOWED_DOMAINS")
		os.Unsetenv("WASM_MAX_INPUT_SIZE")
		os.Unsetenv("WASM_MAX_OUTPUT_SIZE")
	}()

	config := NewSecurityConfigFromEnv()

	// Verify values from environment
	expectedMemory := uint32(128 * 1024 * 1024)
	if config.MaxMemory != expectedMemory {
		t.Errorf("MaxMemory = %d, want %d", config.MaxMemory, expectedMemory)
	}

	expectedTimeout := 60 * time.Second
	if config.MaxExecutionTime != expectedTimeout {
		t.Errorf("MaxExecutionTime = %v, want %v", config.MaxExecutionTime, expectedTimeout)
	}

	if config.PoolSize != 20 {
		t.Errorf("PoolSize = %d, want 20", config.PoolSize)
	}

	if !config.EnableDeterministic {
		t.Error("EnableDeterministic should be true")
	}

	if config.EnableWASI {
		t.Error("EnableWASI should be false")
	}

	if len(config.AllowedDomains) != 2 {
		t.Errorf("AllowedDomains length = %d, want 2", len(config.AllowedDomains))
	}

	expectedInputSize := uint32(2 * 1024 * 1024)
	if config.MaxInputSize != expectedInputSize {
		t.Errorf("MaxInputSize = %d, want %d", config.MaxInputSize, expectedInputSize)
	}

	expectedOutputSize := uint32(4 * 1024 * 1024)
	if config.MaxOutputSize != expectedOutputSize {
		t.Errorf("MaxOutputSize = %d, want %d", config.MaxOutputSize, expectedOutputSize)
	}
}

// TestInMemoryAuditLogger tests the in-memory audit logger
func TestInMemoryAuditLogger(t *testing.T) {
	logger := NewInMemoryAuditLogger()
	ctx := context.Background()

	tenantID := uuid.New()
	functionID := uuid.New()
	executionID := uuid.New()

	// Test logging an execution
	audit := &ExecutionAudit{
		ID:             executionID,
		TenantID:       tenantID,
		FunctionID:     functionID,
		ExecutionID:    executionID,
		Runtime:        "python",
		InputSize:      100,
		OutputSize:     200,
		ExecutionTimeMs: 150,
		Status:         StatusSuccess,
	}

	err := logger.LogExecution(ctx, audit)
	if err != nil {
		t.Errorf("LogExecution failed: %v", err)
	}

	// Test retrieving execution
	retrieved, err := logger.GetExecution(ctx, executionID)
	if err != nil {
		t.Errorf("GetExecution failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved execution is nil")
	}
	if retrieved.Status != StatusSuccess {
		t.Errorf("Retrieved status = %q, want %q", retrieved.Status, StatusSuccess)
	}
	if retrieved.ExecutionTimeMs != 150 {
		t.Errorf("Retrieved execution time = %d, want 150", retrieved.ExecutionTimeMs)
	}

	// Test listing executions
	executions, err := logger.ListExecutions(ctx, tenantID, 10, 0)
	if err != nil {
		t.Errorf("ListExecutions failed: %v", err)
	}
	if len(executions) != 1 {
		t.Errorf("List executions count = %d, want 1", len(executions))
	}

	// Test listing for non-existent tenant
	executions, err = logger.ListExecutions(ctx, uuid.New(), 10, 0)
	if err != nil {
		t.Errorf("ListExecutions for non-existent tenant failed: %v", err)
	}
	if len(executions) != 0 {
		t.Errorf("List executions for non-existent tenant = %d, want 0", len(executions))
	}
}

// TestInMemoryAuditLoggerErrorStatus tests error status logging
func TestInMemoryAuditLoggerErrorStatus(t *testing.T) {
	logger := NewInMemoryAuditLogger()
	ctx := context.Background()

	tenantID := uuid.New()
	functionID := uuid.New()
	executionID := uuid.New()

	// Log an error execution
	audit := &ExecutionAudit{
		TenantID:        tenantID,
		FunctionID:      functionID,
		ExecutionID:     executionID,
		Runtime:         "python",
		InputSize:       100,
		OutputSize:      0,
		ExecutionTimeMs: 5000,
		Status:          StatusError,
		ErrorMessage:   "execution failed: out of memory",
	}

	err := logger.LogExecution(ctx, audit)
	if err != nil {
		t.Errorf("LogExecution failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := logger.GetExecution(ctx, executionID)
	if err != nil {
		t.Errorf("GetExecution failed: %v", err)
	}
	if retrieved.Status != StatusError {
		t.Errorf("Status = %q, want %q", retrieved.Status, StatusError)
	}
	if retrieved.ErrorMessage != "execution failed: out of memory" {
		t.Errorf("ErrorMessage = %q, want %q", retrieved.ErrorMessage, "execution failed: out of memory")
	}
}

// TestInMemoryAuditLoggerTimeout tests timeout status logging
func TestInMemoryAuditLoggerTimeout(t *testing.T) {
	logger := NewInMemoryAuditLogger()
	ctx := context.Background()

	tenantID := uuid.New()
	functionID := uuid.New()
	executionID := uuid.New()

	// Log a timeout execution
	audit := &ExecutionAudit{
		TenantID:        tenantID,
		FunctionID:      functionID,
		ExecutionID:     executionID,
		Runtime:         "python",
		InputSize:       100,
		OutputSize:      0,
		ExecutionTimeMs: 30000,
		Status:          StatusTimeout,
		ErrorMessage:   "execution timed out after 30s",
	}

	err := logger.LogExecution(ctx, audit)
	if err != nil {
		t.Errorf("LogExecution failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := logger.GetExecution(ctx, executionID)
	if err != nil {
		t.Errorf("GetExecution failed: %v", err)
	}
	if retrieved.Status != StatusTimeout {
		t.Errorf("Status = %q, want %q", retrieved.Status, StatusTimeout)
	}
}

// TestExecutionRecorder tests the execution recorder
func TestExecutionRecorder(t *testing.T) {
	logger := NewInMemoryAuditLogger()
	tenantID := uuid.New()
	functionID := uuid.New()
	runtime := "python"

	recorder := NewExecutionRecorder(logger, tenantID, functionID, runtime)
	start := recorder.Start()

	// Simulate execution
	start = start.WithInputSize(100)
	start = start.WithMemoryUsed(1024 * 1024) // 1MB

	// End with success
	err := start.End([]byte(`{"result": "success"}`))
	if err != nil {
		t.Errorf("ExecutionEnd failed: %v", err)
	}

	// Verify audit was logged
	executions, err := logger.ListExecutions(context.Background(), tenantID, 10, 0)
	if err != nil {
		t.Errorf("ListExecutions failed: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("Expected 1 execution, got %d", len(executions))
	}

	exec := executions[0]
	if exec.InputSize != 100 {
		t.Errorf("InputSize = %d, want 100", exec.InputSize)
	}
	if exec.OutputSize != 21 { // len(`{"result": "success"}`)
		t.Errorf("OutputSize = %d, want 21", exec.OutputSize)
	}
	if exec.Status != StatusSuccess {
		t.Errorf("Status = %q, want %q", exec.Status, StatusSuccess)
	}
}

// TestExecutionRecorderError tests the execution recorder with error
func TestExecutionRecorderError(t *testing.T) {
	logger := NewInMemoryAuditLogger()
	tenantID := uuid.New()
	functionID := uuid.New()
	runtime := "python"

	recorder := NewExecutionRecorder(logger, tenantID, functionID, runtime)
	start := recorder.Start()

	start = start.WithInputSize(100)

	// End with error
	err := start.EndWithError(fmt.Errorf("test error"))
	if err != nil {
		t.Errorf("EndWithError failed: %v", err)
	}

	// Verify error was logged
	executions, err := logger.ListExecutions(context.Background(), tenantID, 10, 0)
	if err != nil {
		t.Errorf("ListExecutions failed: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("Expected 1 execution, got %d", len(executions))
	}

	exec := executions[0]
	if exec.Status != StatusError {
		t.Errorf("Status = %q, want %q", exec.Status, StatusError)
	}
	if exec.ErrorMessage != "test error" {
		t.Errorf("ErrorMessage = %q, want %q", exec.ErrorMessage, "test error")
	}
}

// TestCreateAuditMigration tests SQL generation
func TestCreateAuditMigration(t *testing.T) {
	sql := CreateAuditMigration()

	// Basic validation
	if len(sql) == 0 {
		t.Error("Migration SQL should not be empty")
	}

	// Check for key table creation
	if !contains(sql, "wasm_execution_audit") {
		t.Error("Migration should create wasm_execution_audit table")
	}

	// Check for indexes
	if !contains(sql, "idx_wasm_audit_tenant") {
		t.Error("Migration should create tenant index")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
