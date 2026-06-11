-- Drop advanced performance indexes

-- Drop search and filtering indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_search_trust;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_search_popular;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_multi_criteria;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_high_trust;

-- Drop function version indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_versions_lookup;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_versions_latest;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_versions_deterministic;

-- Drop execution analytics indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_analytics;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_tenant;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_cached;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_errors;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_geo;

-- Drop verification and security indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_signatures_verification;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_malware_scans_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_approvals_workflow;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_approvals_pending;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_verification_status_overall;

-- Drop partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_recent;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_high_errors;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_verification_requires_approval;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_verification_blocked;

-- Drop BRIN indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_timestamp_brin;
DROP INDEX CONCURRENTLY IF EXISTS idx_performance_metrics_timestamp_brin;

-- Drop JSONB GIN indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_tags_gin;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_capabilities_gin;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_versions_capabilities_gin;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_approvals_actions_gin;

-- Drop full-text search index
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_search_vector_gin;

-- Drop covering indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_functions_listing_covering;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_versions_listing_covering;
DROP INDEX CONCURRENTLY IF EXISTS idx_registry_function_executions_monitoring_covering;

-- Drop generated column (if it exists)
ALTER TABLE registry_functions DROP COLUMN IF EXISTS search_vector;