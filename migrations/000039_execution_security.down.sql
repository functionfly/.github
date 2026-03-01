-- Remove execution security tables and related objects

-- Drop triggers
DROP TRIGGER IF EXISTS update_user_execution_quotas_updated_at ON user_execution_quotas;
DROP TRIGGER IF EXISTS update_abuse_patterns_updated_at ON abuse_patterns;
DROP TRIGGER IF EXISTS update_function_input_schemas_updated_at ON function_input_schemas;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (due to foreign key constraints)
DROP TABLE IF EXISTS execution_resource_usage;
DROP TABLE IF EXISTS function_input_schemas;
DROP TABLE IF EXISTS execution_security_events;
DROP TABLE IF EXISTS abuse_patterns;
DROP TABLE IF EXISTS user_execution_quotas;