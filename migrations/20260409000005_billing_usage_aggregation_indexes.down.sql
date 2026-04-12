-- Drop indexes for billing usage aggregation
DROP INDEX IF EXISTS idx_usage_events_tenant_timestamp;
DROP INDEX IF EXISTS idx_usage_events_type_timestamp;
DROP INDEX IF EXISTS idx_usage_rollups_tenant_type_date;
DROP INDEX IF EXISTS idx_registry_executions_billing_agg;
