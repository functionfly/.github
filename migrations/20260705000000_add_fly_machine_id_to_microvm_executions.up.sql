BEGIN;

ALTER TABLE microvm_executions
    ADD COLUMN IF NOT EXISTS fly_machine_id TEXT;

CREATE INDEX IF NOT EXISTS idx_microvm_executions_fly_machine_id
    ON microvm_executions(fly_machine_id);

COMMIT;
