-- Drop triggers
DROP TRIGGER IF EXISTS update_function_deployments_updated_at ON function_deployments;
DROP TRIGGER IF EXISTS update_functions_updated_at ON functions;

-- Drop indexes
DROP INDEX IF EXISTS idx_function_logs_level;
DROP INDEX IF EXISTS idx_function_logs_timestamp;
DROP INDEX IF EXISTS idx_function_logs_deployment_id;
DROP INDEX IF EXISTS idx_function_logs_function_id;
DROP INDEX IF EXISTS idx_function_deployments_status;
DROP INDEX IF EXISTS idx_function_deployments_function_id;
DROP INDEX IF EXISTS idx_functions_status;
DROP INDEX IF EXISTS idx_functions_tenant_id;

-- Drop tables
DROP TABLE IF EXISTS function_logs;
DROP TABLE IF EXISTS function_deployments;
DROP TABLE IF EXISTS functions;