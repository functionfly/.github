ALTER TABLE agent_execution_records
    DROP COLUMN IF EXISTS reasoning_tokens,
    DROP COLUMN IF EXISTS total_tokens,
    DROP COLUMN IF EXISTS completion_tokens,
    DROP COLUMN IF EXISTS prompt_tokens,
    DROP COLUMN IF EXISTS provider,
    DROP COLUMN IF EXISTS model_name;

DROP INDEX IF EXISTS idx_agent_exec_model_name;
