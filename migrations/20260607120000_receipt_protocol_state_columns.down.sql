-- Rollback P1: receipt protocol/state columns

DROP INDEX IF EXISTS idx_exec_public_parent;
DROP INDEX IF EXISTS idx_exec_public_state;
DROP INDEX IF EXISTS idx_exec_public_protocol;

ALTER TABLE registry_executions_public
  DROP COLUMN IF EXISTS fallback_chain;

ALTER TABLE registry_executions_public
  DROP COLUMN IF EXISTS parent_task_id;

ALTER TABLE registry_executions_public
  DROP COLUMN IF EXISTS state;

ALTER TABLE registry_executions_public
  DROP COLUMN IF EXISTS protocol;
