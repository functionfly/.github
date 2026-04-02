BEGIN;

ALTER TABLE agent_execution_records
    DROP COLUMN IF EXISTS policy_violation,
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS memory_after_hash,
    DROP COLUMN IF EXISTS memory_before_hash,
    DROP COLUMN IF EXISTS output_hash,
    DROP COLUMN IF EXISTS input_hash,
    DROP COLUMN IF EXISTS function_uri;

COMMIT;
