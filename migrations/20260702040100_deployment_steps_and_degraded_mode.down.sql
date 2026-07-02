DROP TABLE IF EXISTS deployment_steps;

ALTER TABLE tenants DROP COLUMN IF EXISTS degraded_mode;
ALTER TABLE tenants DROP COLUMN IF EXISTS degradation_reason;
ALTER TABLE tenants DROP COLUMN IF EXISTS degradation_updated_at;
