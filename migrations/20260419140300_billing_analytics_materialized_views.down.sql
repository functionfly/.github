-- Migration: Rollback Billing Analytics Materialized Views

-- Drop refresh functions
DROP FUNCTION IF EXISTS refresh_regional_materialized_views();
DROP FUNCTION IF EXISTS refresh_hourly_materialized_views();
DROP FUNCTION IF EXISTS refresh_billing_materialized_views();

-- Drop materialized views (automatically drops indexes)
DROP MATERIALIZED VIEW IF EXISTS mv_team_cost_allocation;
DROP MATERIALIZED VIEW IF EXISTS mv_platform_daily_metrics;
DROP MATERIALIZED VIEW IF EXISTS mv_subscription_revenue_summary;
DROP MATERIALIZED VIEW IF EXISTS mv_api_key_usage_summary;
DROP MATERIALIZED VIEW IF EXISTS mv_regional_performance_summary;
DROP MATERIALIZED VIEW IF EXISTS mv_tenant_cohort_analysis;
DROP MATERIALIZED VIEW IF EXISTS mv_function_usage_stats;
DROP MATERIALIZED VIEW IF EXISTS mv_tenant_daily_billing_summary;
