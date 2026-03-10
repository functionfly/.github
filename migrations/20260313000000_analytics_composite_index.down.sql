-- Migration: 20260313000000_analytics_composite_index
-- Description: Rollback composite index for analytics events

-- Drop the composite index
DROP INDEX IF EXISTS idx_analytics_events_tenant_type_time;
