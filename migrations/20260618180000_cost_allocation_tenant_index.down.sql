-- Down migration: Remove composite indexes added for cost_allocation_entries

DROP INDEX IF EXISTS idx_cost_allocation_entries_tenant_timestamp;
DROP INDEX IF EXISTS idx_cost_allocation_entries_function_timestamp;
DROP INDEX IF EXISTS idx_cost_allocation_entries_author_timestamp;
