-- Drop registry performance indexes
-- This migration removes all the performance indexes added for registry tables

-- ============================================
-- Registry Functions Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_functions_visibility_popularity;
DROP INDEX IF EXISTS idx_registry_functions_category_popularity;
DROP INDEX IF EXISTS idx_registry_functions_reliability_popularity;
DROP INDEX IF EXISTS idx_registry_functions_trust_popularity;
DROP INDEX IF EXISTS idx_registry_functions_search_popularity;
DROP INDEX IF EXISTS idx_registry_functions_author_name_lookup;
DROP INDEX IF EXISTS idx_registry_functions_owner_user;
DROP INDEX IF EXISTS idx_registry_functions_tenant_owner;
DROP INDEX IF EXISTS idx_registry_functions_active_public;
DROP INDEX IF EXISTS idx_registry_functions_high_trust;

-- ============================================
-- Registry Function Versions Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_function_versions_function_version;
DROP INDEX IF EXISTS idx_registry_function_versions_runtime_published;
DROP INDEX IF EXISTS idx_registry_function_versions_deterministic;
DROP INDEX IF EXISTS idx_registry_function_versions_backend;
DROP INDEX IF EXISTS idx_registry_function_versions_deterministic_runtime;

-- ============================================
-- Registry Function Executions Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_executions_function_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_function_version_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_timestamp_outcome;
DROP INDEX IF EXISTS idx_registry_executions_outcome_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_tenant_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_user_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_geo_country_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_verified_at;
DROP INDEX IF EXISTS idx_registry_executions_verification_status;
DROP INDEX IF EXISTS idx_registry_executions_cached_timestamp;
DROP INDEX IF EXISTS idx_registry_executions_recent;
DROP INDEX IF EXISTS idx_registry_executions_successful;
DROP INDEX IF EXISTS idx_registry_executions_cached_recent;

-- ============================================
-- Registry Function Ratings Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_function_ratings_overall_score;
DROP INDEX IF EXISTS idx_registry_function_ratings_reliability;
DROP INDEX IF EXISTS idx_registry_function_ratings_latency;
DROP INDEX IF EXISTS idx_registry_function_ratings_trust_components;

-- ============================================
-- Registry Executions Public Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_executions_public_shareable_created;
DROP INDEX IF EXISTS idx_registry_executions_public_function_created;
DROP INDEX IF EXISTS idx_registry_executions_public_verified;

-- ============================================
-- Execution Resource Usage Indexes
-- ============================================
DROP INDEX IF EXISTS idx_execution_resource_usage_execution;
DROP INDEX IF EXISTS idx_execution_resource_usage_created;

-- ============================================
-- Function Signature Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_function_signatures_function_version;
DROP INDEX IF EXISTS idx_registry_function_signatures_key_id;
DROP INDEX IF EXISTS idx_registry_function_signatures_valid;

-- ============================================
-- Malware Scan Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_function_malware_scans_function_version;
DROP INDEX IF EXISTS idx_registry_function_malware_scans_status;
DROP INDEX IF EXISTS idx_registry_function_malware_scans_risk;

-- ============================================
-- Approval Workflow Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_function_approvals_function_version;
DROP INDEX IF EXISTS idx_registry_function_approvals_status;
DROP INDEX IF EXISTS idx_registry_function_approvals_assigned;
DROP INDEX IF EXISTS idx_registry_function_approvals_trust_level;
DROP INDEX IF EXISTS idx_registry_function_approvals_deadline;

-- ============================================
-- Verification Status Indexes
-- ============================================
DROP INDEX IF EXISTS idx_registry_function_verification_status_function_version;
DROP INDEX IF EXISTS idx_registry_function_verification_status_overall;
DROP INDEX IF EXISTS idx_registry_function_verification_status_next_check;