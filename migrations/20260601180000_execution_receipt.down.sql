-- Rollback: execution_receipt
-- Created at: 2026-06-01T20:35:00
-- Reverses the up migration. All operations are idempotent so this is safe
-- to re-run on a partially-rolled-back database.

-- Down migration (reverses the up migration)
BEGIN;

-- Reverse F: drop user opt-out column
ALTER TABLE users
  DROP COLUMN IF EXISTS receipt_milestones_enabled;

-- Reverse E: drop milestone / revocation tables (CASCADE removes their indexes)
DROP TABLE IF EXISTS receipt_revocations CASCADE;
DROP TABLE IF EXISTS receipt_milestone_events CASCADE;

-- Reverse D: drop indexes
DROP INDEX IF EXISTS idx_rcpt_function_total;
DROP INDEX IF EXISTS idx_rcpt_trending;
DROP INDEX IF EXISTS idx_rcpt_function_created;
DROP INDEX IF EXISTS idx_rcpt_public_id_active;

-- Reverse C: relax NOT NULL
ALTER TABLE registry_executions_public
  ALTER COLUMN function_name   DROP NOT NULL,
  ALTER COLUMN function_author DROP NOT NULL,
  ALTER COLUMN runtime         DROP NOT NULL;

-- Reverse A: drop added columns
ALTER TABLE registry_executions_public
  DROP COLUMN IF EXISTS revoked_at,
  DROP COLUMN IF EXISTS last_viewed_at,
  DROP COLUMN IF EXISTS view_count,
  DROP COLUMN IF EXISTS fork_count,
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS function_visibility,
  DROP COLUMN IF EXISTS output_schema,
  DROP COLUMN IF EXISTS input_schema,
  DROP COLUMN IF EXISTS runtime,
  DROP COLUMN IF EXISTS function_author,
  DROP COLUMN IF EXISTS function_name;

COMMIT;
