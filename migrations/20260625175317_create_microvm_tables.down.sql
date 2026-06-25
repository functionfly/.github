-- Rollback MicroVM tables

BEGIN;

DROP TRIGGER IF EXISTS trigger_update_microvm_usage ON microvm_executions;
DROP FUNCTION IF EXISTS cleanup_microvm_executions(INTEGER);
DROP FUNCTION IF EXISTS update_microvm_usage_after_execution();

DROP TABLE IF EXISTS microvm_audit_log;
DROP TABLE IF EXISTS microvm_billing_records;
DROP TABLE IF EXISTS microvm_tenant_quotas;
DROP TABLE IF EXISTS microvm_executions;

COMMIT;
