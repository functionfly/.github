-- Rollback WASM execution audit table
-- Migration: 20260325000001_wasm_execution_audit

-- Drop indexes first
DROP INDEX IF EXISTS idx_wasm_audit_tenant_function_status;
DROP INDEX IF EXISTS idx_wasm_audit_runtime;
DROP INDEX IF EXISTS idx_wasm_audit_status;
DROP INDEX IF EXISTS idx_wasm_audit_execution;
DROP INDEX IF EXISTS idx_wasm_audit_function;
DROP INDEX IF EXISTS idx_wasm_audit_tenant;

-- Drop the table
DROP TABLE IF EXISTS wasm_execution_audit;
