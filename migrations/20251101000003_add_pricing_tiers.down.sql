-- Remove pricing tiers support
-- Remove priority column from backends table
ALTER TABLE backends DROP COLUMN IF EXISTS priority;

-- Remove plan column from tenants table
ALTER TABLE tenants DROP COLUMN IF EXISTS plan;