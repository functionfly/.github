-- Revert: Update free tier limits back to previous values
BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'Free') THEN
        UPDATE pricing_tiers
        SET
            description = 'For side projects and experimentation. 100,000 executions/month, 1 active function, public-only.',
            features = '{"functions": 1, "providers": 2, "requests": 100000, "agents": 0, "storage_gb": 0, "state_fabrics": 0}'::jsonb,
            max_functions = 1,
            max_executions_per_month = 100000,
            updated_at = NOW()
        WHERE name = 'Free';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_registry_functions_tenant_id_status;

COMMIT;
