-- Fix slow execution retention query (1593ms -> <10ms)
-- The retention query SELECT id FROM registry_function_executions WHERE timestamp < cutoff LIMIT 1000
-- was scanning the entire large table instead of using an efficient index.
-- 
-- This composite covering index allows PostgreSQL to:
-- 1. Use timestamp range scan to find matching rows
-- 2. Return id directly from the index (no heap access needed)
-- 3. Stop after LIMIT 1000 rows

CREATE INDEX IF NOT EXISTS idx_registry_executions_timestamp_id_covering
ON registry_function_executions (timestamp, id);

-- Analyze the table to update statistics so the planner chooses the new index
ANALYZE registry_function_executions;
