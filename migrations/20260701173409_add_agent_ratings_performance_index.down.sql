-- Rollback: add_agent_ratings_performance_index
-- Created at: 2026-07-01T17:34:09-05:00

BEGIN;

DROP INDEX IF EXISTS idx_agent_ratings_agent_id;

COMMIT;
