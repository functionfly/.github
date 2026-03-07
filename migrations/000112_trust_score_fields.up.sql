-- Migration: Add trust score fields to registry_function_ratings
-- Date: 2024-02-19
-- Description: Adds new columns for trust score calculation

-- Add new columns to registry_function_ratings table
ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS p50_latency_ms INTEGER DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS timeout_rate FLOAT DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS error_rate FLOAT DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS consumer_diversity FLOAT DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS tenant_diversity INTEGER DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS user_diversity INTEGER DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS trust_score FLOAT DEFAULT 0;

ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS trust_updated_at TIMESTAMP;

-- Create index on trust_score for fast sorting
CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_trust_score
ON registry_function_ratings(trust_score DESC);

-- Add new columns to registry_function_executions for diversity tracking
ALTER TABLE registry_function_executions
ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE registry_function_executions
ADD COLUMN IF NOT EXISTS user_id UUID;

-- Create indexes for diversity queries
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_tenant_id
ON registry_function_executions(tenant_id)
WHERE tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registry_function_executions_user_id
ON registry_function_executions(user_id)
WHERE user_id IS NOT NULL;
