-- Add environment column to functions table for environment-scoped telemetry queries
ALTER TABLE functions ADD COLUMN IF NOT EXISTS environment VARCHAR(100) NOT NULL DEFAULT '';

-- Add index for environment-filtered queries
CREATE INDEX IF NOT EXISTS idx_functions_tenant_environment ON functions(tenant_id, environment);

-- Backfill empty string for existing rows (they were created without environment context)
-- No action needed as default is already empty string