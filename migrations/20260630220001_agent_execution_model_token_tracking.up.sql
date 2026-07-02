-- Add model and token tracking columns to agent_execution_records
ALTER TABLE agent_execution_records
    ADD COLUMN IF NOT EXISTS model_name TEXT,
    ADD COLUMN IF NOT EXISTS provider TEXT,
    ADD COLUMN IF NOT EXISTS prompt_tokens INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS completion_tokens INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_tokens INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reasoning_tokens INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_agent_exec_model_name ON agent_execution_records(agent_id, model_name) WHERE model_name IS NOT NULL;
