-- Drop index on tenants.created_at (rollback of performance optimization)
DROP INDEX IF EXISTS idx_tenants_created_at;
