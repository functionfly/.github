-- Create WASM execution audit table for Phase 3 production features
-- Migration: 20260325000001_wasm_execution_audit

-- Create the main WASM execution audit table
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

-- Index for querying by tenant (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_wasm_audit_tenant ON wasm_execution_audit(tenant_id, created_at DESC);

-- Index for querying by function
CREATE INDEX IF NOT EXISTS idx_wasm_audit_function ON wasm_execution_audit(function_id, created_at DESC);

-- Index for querying by execution
CREATE INDEX IF NOT EXISTS idx_wasm_audit_execution ON wasm_execution_audit(execution_id);

-- Index for status filtering (useful for error analysis)
CREATE INDEX IF NOT EXISTS idx_wasm_audit_status ON wasm_execution_audit(status, created_at DESC);

-- Index for runtime filtering (useful for performance analysis)
CREATE INDEX IF NOT EXISTS idx_wasm_audit_runtime ON wasm_execution_audit(runtime, created_at DESC);

-- Composite index for tenant + function + status (common query for function health)
CREATE INDEX IF NOT EXISTS idx_wasm_audit_tenant_function_status
    ON wasm_execution_audit(tenant_id, function_id, status, created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE wasm_execution_audit IS 'WASM execution audit log for FunctionFly production monitoring';
COMMENT ON COLUMN wasm_execution_audit.tenant_id IS 'Tenant identifier';
COMMENT ON COLUMN wasm_execution_audit.function_id IS 'Function identifier';
COMMENT ON COLUMN wasm_execution_audit.execution_id IS 'Unique execution identifier';
COMMENT ON COLUMN wasm_execution_audit.runtime IS 'Runtime type (python, python-wasm, javascript, typescript-wasm)';
COMMENT ON COLUMN wasm_execution_audit.input_size IS 'Input size in bytes';
COMMENT ON COLUMN wasm_execution_audit.output_size IS 'Output size in bytes';
COMMENT ON COLUMN wasm_execution_audit.execution_time_ms IS 'Execution time in milliseconds';
COMMENT ON COLUMN wasm_execution_audit.status IS 'Execution status (success, error, timeout, memory_limit, terminated, invalid_input)';
COMMENT ON COLUMN wasm_execution_audit.error_message IS 'Error message if status is error';
COMMENT ON COLUMN wasm_execution_audit.memory_used IS 'Memory used in bytes';

