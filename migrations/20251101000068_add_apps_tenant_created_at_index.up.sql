-- Support ListAppsByTenant ORDER BY created_at DESC with an index
CREATE INDEX IF NOT EXISTS idx_apps_tenant_id_created_at_desc
ON apps(tenant_id, created_at DESC);
