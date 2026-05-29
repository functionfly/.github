-- Add index on tenants.created_at for faster list queries with ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at DESC);
