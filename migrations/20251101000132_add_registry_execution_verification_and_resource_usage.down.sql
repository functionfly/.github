-- Remove execution_resource_usages table and verification/tenant columns from registry_function_executions

DROP TABLE IF EXISTS execution_resource_usages;

DROP INDEX IF EXISTS idx_registry_function_executions_verified_at;
DROP INDEX IF EXISTS idx_registry_function_executions_verification_status;
DROP INDEX IF EXISTS idx_registry_function_executions_tenant_id;
DROP INDEX IF EXISTS idx_registry_function_executions_user_id;

ALTER TABLE registry_function_executions DROP COLUMN IF EXISTS replayed_duration_ms;
ALTER TABLE registry_function_executions DROP COLUMN IF EXISTS verification_error;
ALTER TABLE registry_function_executions DROP COLUMN IF EXISTS verification_status;
ALTER TABLE registry_function_executions DROP COLUMN IF EXISTS verified_at;
ALTER TABLE registry_function_executions DROP COLUMN IF EXISTS user_id;
ALTER TABLE registry_function_executions DROP COLUMN IF EXISTS tenant_id;
