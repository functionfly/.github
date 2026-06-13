-- Rollback: vault_quotas_phase5
BEGIN;
DROP TABLE IF EXISTS vault_rate_limits;
COMMIT;
