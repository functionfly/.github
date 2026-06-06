-- Migration: local_runtime_cleanup_index
-- Created at: 2026-06-05T17:03:00+00:00
-- Purpose: Add composite index for local_runtime_instances cleanup queries
--          Optimizes DELETE WHERE last_heartbeat < ? on a table that grows
--          with each registered local runtime heartbeat update.

BEGIN;

-- Composite index: supports CleanupStaleLocalRuntimes(ctx, maxAge) query
-- WHERE clause matches the cleanup path which only targets 'running' instances
CREATE INDEX IF NOT EXISTS idx_local_runtime_instances_heartbeat_status
    ON local_runtime_instances(last_heartbeat, status)
    WHERE status = 'running';

COMMIT;