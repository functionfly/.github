-- Down migration: Drop agent function tables
DROP TABLE IF EXISTS agent_function_policies CASCADE;
DROP TABLE IF EXISTS agent_function_executions CASCADE;
DROP TABLE IF EXISTS agent_functions CASCADE;
DROP VIEW IF EXISTS agent_exclusive_functions CASCADE;
DROP TYPE IF EXISTS agent_function_category CASCADE;