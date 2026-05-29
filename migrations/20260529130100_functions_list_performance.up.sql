-- Add composite index on functions(tenant_id, created_at) for faster tenant function listing
CREATE INDEX IF NOT EXISTS idx_functions_tenant_created ON functions(tenant_id, created_at DESC);
