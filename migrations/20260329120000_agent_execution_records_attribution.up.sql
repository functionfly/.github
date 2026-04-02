-- Align agent_execution_records with attribution.AgentExecutionRecord (GORM) and cost breakdown queries.

BEGIN;

ALTER TABLE agent_execution_records
    ADD COLUMN IF NOT EXISTS function_uri TEXT,
    ADD COLUMN IF NOT EXISTS input_hash TEXT,
    ADD COLUMN IF NOT EXISTS output_hash TEXT,
    ADD COLUMN IF NOT EXISTS memory_before_hash TEXT,
    ADD COLUMN IF NOT EXISTS memory_after_hash TEXT,
    ADD COLUMN IF NOT EXISTS retry_count INT,
    ADD COLUMN IF NOT EXISTS policy_violation TEXT;

COMMIT;
