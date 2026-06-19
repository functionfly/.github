-- Migration: 20260602000000_update_free_tier_limits
-- Description: Update free tier limits in pricing_tiers to match production config
-- Problem: Free tier was set to 100,000 executions/month and 1 function max
-- Solution: Update to 500 executions/month and 3 functions max

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'Free') THEN
        UPDATE pricing_tiers
        SET
            description = 'For side projects and experimentation. 500 executions/month, 3 active functions, public-only.',
            features = '{"functions": 3, "providers": 2, "requests": 500, "agents": 0, "storage_gb": 0, "state_fabrics": 0}'::jsonb,
            max_functions = 3,
            max_executions_per_month = 500,
            updated_at = NOW()
        WHERE name = 'Free';
    END IF;
END $$;

-- Ensure index exists for active function counting (used by publish guard)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE schemaname = 'public'
        AND tablename = 'registry_functions'
        AND indexname = 'idx_registry_functions_tenant_id_status'
    ) THEN
        CREATE INDEX idx_registry_functions_tenant_id_status
            ON registry_functions (tenant_id, status)
            WHERE tenant_id IS NOT NULL;
    END IF;
END $$;

COMMIT;
