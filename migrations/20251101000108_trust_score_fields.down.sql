-- Rollback: Remove trust score fields from registry_function_ratings
-- Date: 2024-02-19
-- Description: Removes trust score related columns

-- Drop indexes first
DROP INDEX IF EXISTS idx_registry_function_executions_user_id;
DROP INDEX IF EXISTS idx_registry_function_executions_tenant_id;
DROP INDEX IF EXISTS idx_registry_function_ratings_trust_score;

-- Remove columns from registry_function_executions
ALTER TABLE registry_function_executions
DROP COLUMN IF EXISTS user_id;

ALTER TABLE registry_function_executions
DROP COLUMN IF EXISTS tenant_id;

-- Remove columns from registry_function_ratings
ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS trust_updated_at;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS trust_score;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS user_diversity;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS tenant_diversity;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS consumer_diversity;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS error_rate;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS timeout_rate;

ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS p50_latency_ms;