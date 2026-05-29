-- Drop index on functions(tenant_id, created_at) (rollback of performance optimization)
DROP INDEX IF EXISTS idx_functions_tenant_created;
