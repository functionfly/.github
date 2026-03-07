-- Rollback tenant status
DROP INDEX IF EXISTS idx_tenants_status;
ALTER TABLE tenants DROP COLUMN IF EXISTS status;
ALTER TABLE tenants DROP COLUMN IF EXISTS plan;