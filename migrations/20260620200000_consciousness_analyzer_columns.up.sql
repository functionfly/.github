-- +go migration
-- Add missing columns for consciousness analyzers
-- Fixes: traffic analyzer (function_id in usage_events), health analyzer (cold_start_rate in function_dna_profiles)

BEGIN;

-- Add function_id column to usage_events for traffic pattern analysis
ALTER TABLE usage_events ADD COLUMN IF NOT EXISTS function_id uuid;
CREATE INDEX IF NOT EXISTS idx_usage_events_function_id ON usage_events(function_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant_function_timestamp ON usage_events(tenant_id, function_id, timestamp DESC);

-- Add cold_start_rate column to function_dna_profiles for health analysis
ALTER TABLE function_dna_profiles ADD COLUMN IF NOT EXISTS cold_start_rate double precision DEFAULT 0.0;

COMMIT;
