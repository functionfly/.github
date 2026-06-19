-- Migration: Add missing composite index for cost_allocation_entries
-- Created: 2026-06-18
-- Issue: Queries filter by tenant_id + timestamp but no composite index existed
-- Impact: Full table scans on billing reports for large tenants

-- Composite index for tenant+time queries (most common billing filter pattern)
CREATE INDEX IF NOT EXISTS idx_cost_allocation_entries_tenant_timestamp
    ON cost_allocation_entries(tenant_id, timestamp DESC);

-- Additional composite index for function-level cost queries
CREATE INDEX IF NOT EXISTS idx_cost_allocation_entries_function_timestamp
    ON cost_allocation_entries(function_id, timestamp DESC);

-- Composite index for author-level cost queries
CREATE INDEX IF NOT EXISTS idx_cost_allocation_entries_author_timestamp
    ON cost_allocation_entries(function_author, timestamp DESC);

COMMENT ON INDEX idx_cost_allocation_entries_tenant_timestamp IS
    'Optimizes billing reports filtering by tenant and time period';
COMMENT ON INDEX idx_cost_allocation_entries_function_timestamp IS
    'Optimizes per-function cost analysis queries';
COMMENT ON INDEX idx_cost_allocation_entries_author_timestamp IS
    'Optimizes per-author cost aggregation queries';
