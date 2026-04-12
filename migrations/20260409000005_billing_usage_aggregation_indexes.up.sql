-- Billing Usage Aggregation Indexes
-- Optimizes the queries used by RegistryUsageAggregator for aggregating
-- registry_function_executions into usage_events and usage_rollups

-- Index for querying usage events by tenant and timestamp range
-- Used by: GetUsageByTenant, AggregateUsageEventsToRollups
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant_timestamp
    ON usage_events(tenant_id, timestamp DESC);

-- Index for querying usage events by event type and timestamp
-- Used by: AggregateUsageEventsToRollups (filtering by event_type)
CREATE INDEX IF NOT EXISTS idx_usage_events_type_timestamp
    ON usage_events(event_type, timestamp DESC);

-- Composite index for usage rollups - tenant + event_type + period_date
-- Used by: CreateOrUpdateUsageRollup (ON CONFLICT clause), GetUsageByTenant
CREATE INDEX IF NOT EXISTS idx_usage_rollups_tenant_type_date
    ON usage_rollups(tenant_id, event_type, period_date DESC);

-- Optimized index for billing aggregation query in AggregateExecutionsForBilling
-- This covers the join between registry_function_executions, registry_functions,
-- and execution_resource_usage for the billing aggregation query
-- Note: Partial index with NOW() requires immutable function, so we use a regular index
-- The query planner will use timestamp filtering efficiently with this index
CREATE INDEX IF NOT EXISTS idx_registry_executions_billing_agg
    ON registry_function_executions(timestamp, function_id, outcome, cached, duration_ms);

-- Index on execution_resource_usage for efficient joins with executions
-- Used by: AggregateExecutionsForBilling (LEFT JOIN on execution_id)
CREATE INDEX IF NOT EXISTS idx_execution_resource_usage_execution_id
    ON execution_resource_usage(execution_id)
    INCLUDE (cpu_time_used_ms, memory_used_mb);

-- Index for invoices by tenant and period (for GetInvoiceByPeriod)
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_period
    ON invoices(tenant_id, period_start, period_end)
    WHERE status = 'draft';

-- Index on registry_functions for billing aggregation joins
-- Used by: AggregateExecutionsForBilling (JOIN on function_id)
CREATE INDEX IF NOT EXISTS idx_registry_functions_tenant_billing
    ON registry_functions(tenant_id, id)
    WHERE tenant_id IS NOT NULL;

-- Add comment explaining the aggregation flow
COMMENT ON TABLE usage_events IS 'Granular billing events. Populated by execution handler (async) and aggregated into usage_rollups by RegistryUsageAggregator.';
COMMENT ON TABLE usage_rollups IS 'Daily aggregated usage data. Created/updated by RegistryUsageAggregator from usage_events. Used for efficient billing queries and invoice generation.';
COMMENT ON INDEX idx_usage_events_tenant_timestamp IS 'Supports GetUsageByTenant queries for billing dashboard';
COMMENT ON INDEX idx_usage_rollups_tenant_type_date IS 'Supports rollup aggregation and ON CONFLICT updates';
