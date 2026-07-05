BEGIN;

DROP INDEX IF EXISTS idx_microvm_executions_fly_machine_id;
ALTER TABLE microvm_executions DROP COLUMN IF EXISTS fly_machine_id;

COMMIT;
