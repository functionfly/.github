-- Rollback function registry schema
DROP TABLE IF EXISTS registry_function_ratings CASCADE;
DROP TABLE IF EXISTS registry_function_executions CASCADE;
DROP TABLE IF EXISTS registry_function_versions CASCADE;
DROP TABLE IF EXISTS registry_functions CASCADE;
