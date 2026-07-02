-- Migration: add_agent_ratings_performance_index
-- Created at: 2026-07-01T17:34:09-05:00
-- Purpose: Add index on agent_ratings.agent_id to speed up the update_agent_listing_rating
--          trigger that runs AVG() on ratings after INSERT/UPDATE/DELETE on agent_ratings.
--          Without this index, the trigger's subquery does a sequential scan on every rating operation.

BEGIN;

CREATE INDEX IF NOT EXISTS idx_agent_ratings_agent_id ON agent_ratings(agent_id);

COMMIT;
