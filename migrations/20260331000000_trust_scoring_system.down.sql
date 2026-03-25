-- Migration: 20260331000000_trust_scoring_system
-- Description: Rollback Trust Scoring System Phase 1
-- Created: 2026-03-31

BEGIN;

-- Drop indexes first
DROP INDEX IF EXISTS idx_registry_functions_trust_score;
DROP INDEX IF EXISTS idx_registry_functions_trust_tier;
DROP INDEX IF EXISTS idx_trust_history_function_id;
DROP INDEX IF EXISTS idx_trust_history_calculated_at;
DROP INDEX IF EXISTS idx_trust_history_function_calculated;
DROP INDEX IF EXISTS idx_trust_history_trust_tier;
DROP INDEX IF EXISTS idx_execution_metrics_function_id;
DROP INDEX IF EXISTS idx_execution_metrics_window;
DROP INDEX IF EXISTS idx_execution_metrics_function_window;
DROP INDEX IF EXISTS idx_execution_metrics_window_type;
DROP INDEX IF EXISTS idx_trust_score_jobs_status;
DROP INDEX IF EXISTS idx_trust_score_jobs_created;

-- Remove trust columns from registry_functions
ALTER TABLE registry_functions
    DROP COLUMN IF EXISTS trust_score,
    DROP COLUMN IF EXISTS trust_tier,
    DROP COLUMN IF EXISTS trust_updated_at,
    DROP COLUMN IF EXISTS trust_calculation_version;

-- Drop tables in reverse order of creation (due to foreign keys)
DROP TABLE IF EXISTS trust_score_jobs;
DROP TABLE IF EXISTS trust_score_weights;
DROP TABLE IF EXISTS execution_metrics;
DROP TABLE IF EXISTS trust_history;

COMMIT;
