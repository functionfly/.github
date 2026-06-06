-- Rollback: local_runtime_cleanup_index
-- Created at: 2026-06-05T17:03:00+00:00

BEGIN;

DROP INDEX IF EXISTS idx_local_runtime_instances_heartbeat_status;

COMMIT;