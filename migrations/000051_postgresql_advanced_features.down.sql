-- Drop advanced PostgreSQL features

-- Drop generated columns
ALTER TABLE registry_functions DROP COLUMN IF EXISTS search_rank_score;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS efficiency_score;
ALTER TABLE registry_function_verification_status DROP COLUMN IF EXISTS risk_score;

-- Drop enhanced JSONB indexes
DROP INDEX IF EXISTS idx_registry_functions_tags_array;
DROP INDEX IF EXISTS idx_registry_functions_capabilities_keys;
DROP INDEX IF EXISTS idx_registry_function_versions_capabilities_keys;
DROP INDEX IF EXISTS idx_registry_functions_category_path;
DROP INDEX IF EXISTS idx_registry_functions_runtime_path;

-- Drop BRIN indexes
DROP INDEX IF EXISTS idx_audit_events_timestamp_brin;
DROP INDEX IF EXISTS idx_usage_events_timestamp_brin;
DROP INDEX IF EXISTS idx_alerts_created_at_brin;

-- Drop expression indexes
DROP INDEX IF EXISTS idx_registry_functions_age;
DROP INDEX IF EXISTS idx_registry_function_executions_duration_category;
DROP INDEX IF EXISTS idx_registry_function_versions_size_category;

-- Drop partial expression indexes
DROP INDEX IF EXISTS idx_registry_functions_recently_active;
DROP INDEX IF EXISTS idx_registry_functions_high_usage;
DROP INDEX IF EXISTS idx_registry_functions_error_prone;

-- Drop weighted full-text search index
DROP INDEX IF EXISTS idx_registry_functions_weighted_search;

-- Drop triggers and functions
DROP TRIGGER IF EXISTS validate_trust_score_trigger ON registry_function_ratings;
DROP FUNCTION IF EXISTS validate_trust_score();
DROP FUNCTION IF EXISTS update_function_popularity();

-- Drop views
DROP VIEW IF EXISTS function_performance_analytics;
DROP VIEW IF EXISTS tenant_usage_analytics;