-- Migration: 20260313000000_analytics_composite_index
-- Description: Add composite index for analytics events query optimization
--              Supports filtering by tenant_id + resource_type + occurred_at

-- Add composite index for common query pattern: tenant_id + resource_type + occurred_at
CREATE INDEX IF NOT EXISTS idx_analytics_events_tenant_type_time
ON analytics_events(tenant_id, resource_type, occurred_at);
