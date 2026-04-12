-- Migration: 20260410223826_three_layer_pricing_strategy (down)
-- Description: Rollback 3-layer pricing changes

-- Drop indexes
DROP INDEX IF EXISTS idx_usage_events_v2_tenant_id;
DROP INDEX IF EXISTS idx_usage_events_v2_event_type;
DROP INDEX IF EXISTS idx_usage_events_v2_timestamp;
DROP INDEX IF EXISTS idx_usage_events_v2_tenant_event_time;
DROP INDEX IF EXISTS idx_usage_rollups_v2_tenant_id;
DROP INDEX IF EXISTS idx_usage_rollups_v2_period;
DROP INDEX IF EXISTS idx_usage_rollups_v2_tenant_period;
DROP INDEX IF EXISTS idx_pending_charges_tenant;
DROP INDEX IF EXISTS idx_pending_charges_period;
DROP INDEX IF EXISTS idx_pending_charges_uninvoiced;

-- Drop function
DROP FUNCTION IF EXISTS calculate_usage_charges(UUID, DATE, DATE);

-- Drop tables
DROP TABLE IF EXISTS pending_usage_charges;
DROP TABLE IF EXISTS usage_events_v2;
DROP TABLE IF EXISTS usage_rollups_v2;
DROP TABLE IF EXISTS usage_pricing_config;

-- Revert Free tier to basic limits (you may want to adjust)
UPDATE pricing_tiers SET
    max_functions = 5,
    max_executions_per_month = 1000,
    features = jsonb_build_object(
        'requests', 1000,
        'included_compute_ms', 3600000,
        'included_compute_hours', 1.0,
        'functions', 5
    )
WHERE name = 'Free';

-- Remove Pro and Team tiers (optional - comment out if you want to keep them)
-- DELETE FROM pricing_tiers WHERE name IN ('Pro', 'Team');
