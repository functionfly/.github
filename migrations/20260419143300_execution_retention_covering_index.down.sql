-- Revert execution retention covering index
DROP INDEX IF EXISTS idx_registry_executions_timestamp_id_covering;
