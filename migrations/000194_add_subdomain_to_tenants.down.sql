-- Remove subdomain column from tenants table
DROP INDEX IF EXISTS idx_tenants_subdomain;
ALTER TABLE tenants DROP COLUMN IF EXISTS subdomain;
