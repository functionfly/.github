-- Rollback: add_slow_query_performance_indexes
-- Created at: 2026-06-11T21:21:00-05:00

-- Down migration (reverses the up migration)
BEGIN;

DROP INDEX IF EXISTS CONCURRENTLY idx_secrets_vault_active_tenant;
DROP INDEX IF EXISTS CONCURRENTLY idx_registry_function_executions_timestamp_id;

COMMIT;