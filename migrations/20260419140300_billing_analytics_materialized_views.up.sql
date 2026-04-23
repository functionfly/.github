-- Migration: Materialized Views for Billing Analytics
-- Pre-computed aggregations for fast dashboard queries
-- Created: 2026-04-19

-- ============================================
-- 1. Tenant Daily Billing Summary
-- Core view for tenant billing dashboards
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_tenant_daily_billing_summary AS
WITH daily_stats AS (
    SELECT 
        tenant_id,
        DATE(timestamp) as billing_date,
        COUNT(*) as execution_count,
        COUNT(DISTINCT function_id) as unique_functions,
        SUM(total_cost_cents) as total_cost_cents,
        SUM(execution_cost_cents) as execution_cost_cents,
        SUM(compute_cost_cents) as compute_cost_cents,
        SUM(platform_fee_cents) as platform_fee_cents,
        SUM(data_transfer_cents) as data_transfer_cents,
        SUM(CASE WHEN execution_outcome = 'success' THEN 1 ELSE 0 END) as success_count,
        SUM(CASE WHEN execution_outcome != 'success' THEN 1 ELSE 0 END) as error_count,
        SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_count,
        AVG(duration_ms)::BIGINT as avg_duration_ms,
        MAX(timestamp) as last_execution_at
    FROM cost_allocation_entries
    WHERE timestamp > CURRENT_DATE - INTERVAL '90 days'
    GROUP BY tenant_id, DATE(timestamp)
),
cumulative AS (
    SELECT 
        tenant_id,
        billing_date,
        execution_count,
        unique_functions,
        total_cost_cents,
        execution_cost_cents,
        compute_cost_cents,
        platform_fee_cents,
        data_transfer_cents,
        success_count,
        error_count,
        cached_count,
        avg_duration_ms,
        last_execution_at,
        SUM(total_cost_cents) OVER (
            PARTITION BY tenant_id 
            ORDER BY billing_date 
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) as cumulative_cost_cents,
        SUM(execution_count) OVER (
            PARTITION BY tenant_id 
            ORDER BY billing_date 
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) as cumulative_executions
    FROM daily_stats
)
SELECT * FROM cumulative
ORDER BY tenant_id, billing_date DESC;

-- Index for tenant lookup
CREATE UNIQUE INDEX idx_mv_tenant_daily_tenant_date 
ON mv_tenant_daily_billing_summary (tenant_id, billing_date);

-- Index for recent data queries
CREATE INDEX idx_mv_tenant_daily_recent 
ON mv_tenant_daily_billing_summary (billing_date DESC, tenant_id);

COMMENT ON MATERIALIZED VIEW mv_tenant_daily_billing_summary IS 
'Daily billing summary per tenant. Last 90 days. Refresh nightly.';

-- ============================================
-- 2. Function Usage Statistics
-- Per-function analytics for publisher dashboards
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_function_usage_stats AS
SELECT 
    ca.function_id,
    ca.function_name,
    ca.function_author,
    COUNT(*) as total_executions,
    COUNT(DISTINCT ca.tenant_id) as unique_tenants,
    SUM(ca.total_cost_cents) as total_revenue_cents,
    SUM(CASE WHEN ca.execution_outcome = 'success' THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN ca.execution_outcome != 'success' THEN 1 ELSE 0 END) as error_count,
    AVG(ca.duration_ms)::BIGINT as avg_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY ca.duration_ms) as p95_duration_ms,
    MAX(ca.timestamp) as last_executed_at,
    MIN(ca.timestamp) as first_executed_at,
    SUM(CASE WHEN ca.cached THEN 1 ELSE 0 END) as cache_hit_count,
    ROUND(
        100.0 * SUM(CASE WHEN ca.cached THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0), 
        2
    ) as cache_hit_rate_pct
FROM cost_allocation_entries ca
WHERE ca.timestamp > CURRENT_DATE - INTERVAL '90 days'
GROUP BY ca.function_id, ca.function_name, ca.function_author
ORDER BY total_executions DESC;

CREATE UNIQUE INDEX idx_mv_function_stats_function 
ON mv_function_usage_stats (function_id);

CREATE INDEX idx_mv_function_stats_author 
ON mv_function_usage_stats (function_author, total_executions DESC);

CREATE INDEX idx_mv_function_stats_revenue 
ON mv_function_usage_stats (total_revenue_cents DESC) 
WHERE total_revenue_cents > 0;

COMMENT ON MATERIALIZED VIEW mv_function_usage_stats IS 
'Function-level usage statistics. Last 90 days. Refresh daily.';

-- ============================================
-- 3. Tenant Monthly Cohort Analysis
-- For retention and MRR analysis
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_tenant_cohort_analysis AS
WITH first_usage AS (
    SELECT 
        tenant_id,
        MIN(DATE(timestamp)) as first_usage_date,
        DATE_TRUNC('month', MIN(DATE(timestamp))) as cohort_month
    FROM cost_allocation_entries
    GROUP BY tenant_id
),
monthly_activity AS (
    SELECT 
        tenant_id,
        DATE_TRUNC('month', DATE(timestamp)) as activity_month,
        COUNT(*) as execution_count,
        SUM(total_cost_cents) as total_cost_cents
    FROM cost_allocation_entries
    GROUP BY tenant_id, DATE_TRUNC('month', DATE(timestamp))
)
SELECT 
    f.cohort_month,
    f.first_usage_date,
    f.tenant_id,
    m.activity_month,
    EXTRACT(YEAR FROM AGE(m.activity_month, f.cohort_month)) * 12 +
    EXTRACT(MONTH FROM AGE(m.activity_month, f.cohort_month)) as months_since_first,
    m.execution_count,
    m.total_cost_cents,
    CASE 
        WHEN m.activity_month = f.cohort_month THEN 'new'
        ELSE 'retained'
    END as status
FROM first_usage f
JOIN monthly_activity m ON f.tenant_id = m.tenant_id
ORDER BY f.cohort_month DESC, m.activity_month;

CREATE INDEX idx_mv_cohort_analysis_cohort 
ON mv_tenant_cohort_analysis (cohort_month, months_since_first);

CREATE INDEX idx_mv_cohort_analysis_tenant 
ON mv_tenant_cohort_analysis (tenant_id, activity_month);

COMMENT ON MATERIALIZED VIEW mv_tenant_cohort_analysis IS 
'Cohort analysis for tenant retention. Refresh daily.';

-- ============================================
-- 4. Regional Performance Summary
-- For capacity planning and regional optimization
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_regional_performance_summary AS
SELECT 
    COALESCE(region, 'unknown') as region,
    DATE(timestamp) as stat_date,
    COUNT(*) as execution_count,
    COUNT(DISTINCT tenant_id) as active_tenants,
    COUNT(DISTINCT function_id) as active_functions,
    AVG(duration_ms)::BIGINT as avg_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms) as p95_duration_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms) as p99_duration_ms,
    SUM(total_cost_cents) as total_cost_cents,
    SUM(CASE WHEN execution_outcome != 'success' THEN 1 ELSE 0 END) as error_count,
    ROUND(
        100.0 * SUM(CASE WHEN execution_outcome = 'success' THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0),
        4
    ) as success_rate_pct,
    AVG(memory_used_mb)::BIGINT as avg_memory_mb,
    MAX(memory_used_mb) as peak_memory_mb
FROM cost_allocation_entries
WHERE timestamp > CURRENT_DATE - INTERVAL '30 days'
GROUP BY COALESCE(region, 'unknown'), DATE(timestamp)
ORDER BY stat_date DESC, region;

CREATE UNIQUE INDEX idx_mv_regional_summary_region_date 
ON mv_regional_performance_summary (region, stat_date);

CREATE INDEX idx_mv_regional_summary_date 
ON mv_regional_performance_summary (stat_date DESC);

COMMENT ON MATERIALIZED VIEW mv_regional_performance_summary IS 
'Regional performance metrics. Last 30 days. Refresh 4x daily.';

-- ============================================
-- 5. API Key Usage Summary
-- For API key management and quota tracking
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_api_key_usage_summary AS
SELECT 
    ca.api_key_id,
    ca.tenant_id,
    DATE_TRUNC('hour', ca.timestamp) as usage_hour,
    COUNT(*) as request_count,
    SUM(ca.total_cost_cents) as total_cost_cents,
    COUNT(DISTINCT ca.function_id) as unique_functions,
    AVG(ca.duration_ms)::BIGINT as avg_duration_ms,
    SUM(CASE WHEN ca.execution_outcome != 'success' THEN 1 ELSE 0 END) as error_count
FROM cost_allocation_entries ca
WHERE ca.api_key_id IS NOT NULL
    AND ca.timestamp > CURRENT_DATE - INTERVAL '7 days'
GROUP BY ca.api_key_id, ca.tenant_id, DATE_TRUNC('hour', ca.timestamp)
ORDER BY usage_hour DESC, request_count DESC;

CREATE UNIQUE INDEX idx_mv_api_key_usage_lookup 
ON mv_api_key_usage_summary (api_key_id, usage_hour);

CREATE INDEX idx_mv_api_key_usage_tenant 
ON mv_api_key_usage_summary (tenant_id, usage_hour DESC);

COMMENT ON MATERIALIZED VIEW mv_api_key_usage_summary IS 
'Hourly API key usage summary. Last 7 days. Refresh hourly.';

-- ============================================
-- 6. Subscription Revenue Summary (MRR Tracking)
-- For financial reporting
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_subscription_revenue_summary AS
SELECT 
    s.id as subscription_id,
    s.tenant_id,
    t.name as tenant_name,
    pt.name as tier_name,
    pt.price_cents as monthly_price_cents,
    s.status,
    s.current_period_start,
    s.current_period_end,
    CASE 
        WHEN s.status = 'active' AND s.cancel_at_period_end = false THEN pt.price_cents
        WHEN s.status = 'trialing' THEN 0
        ELSE 0
    END as recognized_mrr_cents,
    CASE 
        WHEN s.cancel_at_period_end THEN 'canceling'
        WHEN s.status = 'past_due' THEN 'at_risk'
        WHEN s.status = 'trialing' THEN 'trial'
        ELSE 'healthy'
    END as revenue_status,
    s.created_at as subscription_created_at,
    s.canceled_at,
    s.cancel_at_period_end
FROM subscriptions s
JOIN tenants t ON s.tenant_id = t.id
JOIN pricing_tiers pt ON s.pricing_tier_id = pt.id
WHERE s.status IN ('active', 'trialing', 'past_due')
ORDER BY recognized_mrr_cents DESC, tenant_id;

CREATE UNIQUE INDEX idx_mv_subscription_revenue_sub 
ON mv_subscription_revenue_summary (subscription_id);

CREATE INDEX idx_mv_subscription_revenue_tenant 
ON mv_subscription_revenue_summary (tenant_id);

CREATE INDEX idx_mv_subscription_revenue_status 
ON mv_subscription_revenue_summary (revenue_status, recognized_mrr_cents DESC);

COMMENT ON MATERIALIZED VIEW mv_subscription_revenue_summary IS 
'Subscription MRR summary for financial reporting. Refresh daily.';

-- ============================================
-- 7. Platform Daily Metrics Summary
-- For platform health dashboards
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_platform_daily_metrics AS
SELECT 
    DATE(timestamp) as metric_date,
    COUNT(DISTINCT tenant_id) as active_tenants,
    COUNT(DISTINCT function_id) as active_functions,
    COUNT(*) as total_executions,
    SUM(total_cost_cents) as total_cost_cents,
    SUM(CASE WHEN execution_outcome = 'success' THEN 1 ELSE 0 END) as successful_executions,
    SUM(CASE WHEN execution_outcome != 'success' THEN 1 ELSE 0 END) as failed_executions,
    SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_executions,
    AVG(duration_ms)::BIGINT as avg_duration_ms,
    MAX(duration_ms) as max_duration_ms,
    COUNT(DISTINCT CASE WHEN timestamp > CURRENT_DATE - INTERVAL '1 day' THEN tenant_id END) as new_daily_active_tenants,
    COUNT(DISTINCT CASE WHEN timestamp > CURRENT_DATE - INTERVAL '1 hour' THEN tenant_id END) as new_hourly_active_tenants
FROM cost_allocation_entries
WHERE timestamp > CURRENT_DATE - INTERVAL '90 days'
GROUP BY DATE(timestamp)
ORDER BY metric_date DESC;

CREATE UNIQUE INDEX idx_mv_platform_metrics_date 
ON mv_platform_daily_metrics (metric_date);

COMMENT ON MATERIALIZED VIEW mv_platform_daily_metrics IS 
'Platform-wide daily metrics. Last 90 days. Refresh daily.';

-- ============================================
-- 8. Team Cost Allocation Summary
-- For team-based billing and chargeback
-- ============================================

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_team_cost_allocation AS
WITH team_members AS (
    SELECT DISTINCT 
        tm.team_id,
        tm.user_id,
        t.tenant_id,
        t.name as team_name
    FROM team_memberships tm
    JOIN teams t ON tm.team_id = t.id
    WHERE tm.role IN ('owner', 'admin', 'member')
)
SELECT 
    tm.team_id,
    tm.tenant_id,
    tm.team_name,
    DATE(ca.timestamp) as usage_date,
    COUNT(*) as team_executions,
    SUM(ca.total_cost_cents) as team_cost_cents,
    SUM(ca.execution_cost_cents) as execution_cost_cents,
    SUM(ca.compute_cost_cents) as compute_cost_cents,
    SUM(ca.platform_fee_cents) as platform_fee_cents,
    COUNT(DISTINCT ca.function_id) as team_functions_used,
    COUNT(DISTINCT ca.api_key_id) as team_api_keys_used,
    ARRAY_AGG(DISTINCT ca.function_name ORDER BY ca.function_name) as top_functions
FROM cost_allocation_entries ca
JOIN team_members tm ON ca.tenant_id = tm.tenant_id
WHERE ca.timestamp > CURRENT_DATE - INTERVAL '30 days'
GROUP BY tm.team_id, tm.tenant_id, tm.team_name, DATE(ca.timestamp)
ORDER BY usage_date DESC, team_cost_cents DESC;

CREATE UNIQUE INDEX idx_mv_team_cost_team_date 
ON mv_team_cost_allocation (team_id, usage_date);

CREATE INDEX idx_mv_team_cost_tenant 
ON mv_team_cost_allocation (tenant_id, usage_date DESC);

COMMENT ON MATERIALIZED VIEW mv_team_cost_allocation IS 
'Team-level cost allocation for chargeback. Last 30 days. Refresh daily.';

-- ============================================
-- 9. Refresh Functions
-- ============================================

-- Function to refresh all billing materialized views
CREATE OR REPLACE FUNCTION refresh_billing_materialized_views()
RETURNS TABLE (view_name TEXT, refreshed_at TIMESTAMPTZ, duration_ms INTEGER) AS $$
DECLARE
    start_time TIMESTAMPTZ;
    end_time TIMESTAMPTZ;
BEGIN
    -- Refresh daily views (order matters for dependencies)
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_tenant_daily_billing_summary;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_tenant_daily_billing_summary'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_function_usage_stats;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_function_usage_stats'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_tenant_cohort_analysis;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_tenant_cohort_analysis'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_subscription_revenue_summary;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_subscription_revenue_summary'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_platform_daily_metrics;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_platform_daily_metrics'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
    
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_team_cost_allocation;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_team_cost_allocation'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Function to refresh hourly views only
CREATE OR REPLACE FUNCTION refresh_hourly_materialized_views()
RETURNS TABLE (view_name TEXT, refreshed_at TIMESTAMPTZ, duration_ms INTEGER) AS $$
DECLARE
    start_time TIMESTAMPTZ;
    end_time TIMESTAMPTZ;
BEGIN
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_api_key_usage_summary;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_api_key_usage_summary'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Function to refresh regional metrics (4x daily)
CREATE OR REPLACE FUNCTION refresh_regional_materialized_views()
RETURNS TABLE (view_name TEXT, refreshed_at TIMESTAMPTZ, duration_ms INTEGER) AS $$
DECLARE
    start_time TIMESTAMPTZ;
    end_time TIMESTAMPTZ;
BEGIN
    start_time := clock_timestamp();
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_regional_performance_summary;
    end_time := clock_timestamp();
    RETURN QUERY SELECT 'mv_regional_performance_summary'::TEXT, NOW(), 
        EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION refresh_billing_materialized_views() IS 
'Refresh all daily billing materialized views. Run via cron at 3 AM daily.';

COMMENT ON FUNCTION refresh_hourly_materialized_views() IS 
'Refresh hourly materialized views. Run via cron every hour.';

COMMENT ON FUNCTION refresh_regional_materialized_views() IS 
'Refresh regional performance views. Run via cron every 6 hours.';
