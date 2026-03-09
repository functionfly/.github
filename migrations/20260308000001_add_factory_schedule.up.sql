-- Ensure factory_config table exists (may live in internal/storage/sql/migrations in some setups)
CREATE TABLE IF NOT EXISTS factory_config (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    discovery_batch_size INTEGER NOT NULL DEFAULT 10,
    minimum_quality_score DECIMAL(5,2) NOT NULL DEFAULT 70,
    minimum_test_score DECIMAL(5,2) NOT NULL DEFAULT 80,
    require_all_tests_pass BOOLEAN NOT NULL DEFAULT true,
    auto_publish BOOLEAN NOT NULL DEFAULT true,
    max_opportunities_per_run INTEGER NOT NULL DEFAULT 3,
    retry_attempts INTEGER NOT NULL DEFAULT 1,
    retry_backoff_ms INTEGER NOT NULL DEFAULT 500,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_factory_config_agent_id ON factory_config(agent_id);

-- Add scheduling fields to factory_config table
ALTER TABLE factory_config
ADD COLUMN IF NOT EXISTS schedule_enabled BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS schedule_cron TEXT,
ADD COLUMN IF NOT EXISTS schedule_timezone TEXT DEFAULT 'UTC';
