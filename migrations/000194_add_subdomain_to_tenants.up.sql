-- Add subdomain column to tenants table
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS subdomain VARCHAR(255);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_subdomain ON tenants(subdomain) WHERE subdomain IS NOT NULL;
