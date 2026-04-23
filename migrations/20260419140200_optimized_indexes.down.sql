-- Migration: Rollback Optimized Indexes
-- Removes all GIN, covering, partial, BRIN, and expression indexes
-- Preserves basic B-tree indexes on primary/foreign keys

-- ============================================
-- 1. Drop Functions
-- ============================================

DROP FUNCTION IF EXISTS safe_reindex_table(TEXT, BOOLEAN);
DROP FUNCTION IF EXISTS get_bloated_indexes(FLOAT, BIGINT);

-- ============================================
-- 2. Drop Views
-- ============================================

DROP VIEW IF EXISTS index_usage_stats;

-- ============================================
-- 3. Drop Generated Columns
-- ============================================

-- Note: PostgreSQL doesn't allow dropping generated columns directly
-- if they have dependencies. Order matters here.

-- First drop indexes on generated columns
DROP INDEX IF EXISTS idx_graph_definitions_published;
DROP INDEX IF EXISTS idx_graph_definitions_trigger_type;
DROP INDEX IF EXISTS idx_registry_functions_determinism;
DROP INDEX IF EXISTS idx_email_events_category;
DROP INDEX IF EXISTS idx_users_profile_visibility;
DROP INDEX IF EXISTS idx_functions_playground_runtime;

-- Then we can drop the columns (but they're automatically dropped with dependencies)
-- These will be removed when the migration runs

-- ============================================
-- 4. Drop BRIN Indexes
-- ============================================

DROP INDEX IF EXISTS idx_email_events_brin;
DROP INDEX IF EXISTS idx_function_logs_brin;
DROP INDEX IF EXISTS idx_perf_metrics_brin;
DROP INDEX IF EXISTS idx_health_checks_brin;
DROP INDEX IF EXISTS idx_routing_events_brin;
DROP INDEX IF EXISTS idx_registry_exec_brin;
DROP INDEX IF EXISTS idx_cost_allocation_brin;

-- ============================================
-- 5. Drop Expression/Full-Text Indexes
-- ============================================

DROP INDEX IF EXISTS idx_team_memories_content_search;
DROP INDEX IF EXISTS idx_graph_definitions_description_search;
DROP INDEX IF EXISTS idx_registry_functions_description_search;
DROP INDEX IF EXISTS idx_registry_functions_name_lower;
DROP INDEX IF EXISTS idx_users_username_lower;
DROP INDEX IF EXISTS idx_users_email_lower;

-- ============================================
-- 6. Drop Partial Indexes
-- ============================================

DROP INDEX IF EXISTS idx_team_invites_pending;
DROP INDEX IF EXISTS idx_magic_links_unexpired;
DROP INDEX IF EXISTS idx_pending_username_pending;
DROP INDEX IF EXISTS idx_dunning_campaigns_active;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_invoices_uncollected;
DROP INDEX IF EXISTS idx_subscriptions_active;
DROP INDEX IF EXISTS idx_monitoring_alerts_active;
DROP INDEX IF EXISTS idx_notifications_unread;
DROP INDEX IF EXISTS idx_webhook_payloads_failed;
DROP INDEX IF EXISTS idx_webhook_payloads_pending;
DROP INDEX IF EXISTS idx_registry_exec_pending_verification;
DROP INDEX IF EXISTS idx_registry_exec_cached;
DROP INDEX IF EXISTS idx_registry_exec_failed;
DROP INDEX IF EXISTS idx_sessions_active;

-- ============================================
-- 7. Drop Covering Indexes (INCLUDE clause)
-- ============================================

DROP INDEX IF EXISTS idx_usage_rollups_tenant_covering;
DROP INDEX IF EXISTS idx_audit_events_tenant_covering;
DROP INDEX IF EXISTS idx_provider_tokens_user_covering;
DROP INDEX IF EXISTS idx_refresh_tokens_user_covering;
DROP INDEX IF EXISTS idx_sessions_user_covering;
DROP INDEX IF EXISTS idx_magic_links_email_covering;
DROP INDEX IF EXISTS idx_team_memberships_user_covering;
DROP INDEX IF EXISTS idx_team_memberships_team_covering;
DROP INDEX IF EXISTS idx_invoices_tenant_covering;
DROP INDEX IF EXISTS idx_subscriptions_tenant_covering;
DROP INDEX IF EXISTS idx_backends_app_covering;
DROP INDEX IF EXISTS idx_apps_tenant_covering;
DROP INDEX IF EXISTS idx_user_activity_user_covering;
DROP INDEX IF EXISTS idx_registry_exec_function_covering;
DROP INDEX IF EXISTS idx_registry_exec_tenant_covering;
DROP INDEX IF EXISTS idx_cost_allocation_function_covering;
DROP INDEX IF EXISTS idx_cost_allocation_tenant_covering;

-- ============================================
-- 8. Drop GIN Indexes
-- ============================================

DROP INDEX IF EXISTS idx_email_events_metadata_gin;
DROP INDEX IF EXISTS idx_email_events_event_data_gin;
DROP INDEX IF EXISTS idx_stored_webhook_payloads_payload_gin;
DROP INDEX IF EXISTS idx_monitoring_alerts_metadata_gin;
DROP INDEX IF EXISTS idx_performance_metrics_labels_gin;
DROP INDEX IF EXISTS idx_team_memories_tags_gin;
DROP INDEX IF EXISTS idx_trustapi_service_keys_scopes_gin;
DROP INDEX IF EXISTS idx_trustapi_client_keys_scopes_gin;
DROP INDEX IF EXISTS idx_factory_pipeline_jobs_metadata_gin;
DROP INDEX IF EXISTS idx_factory_pipeline_runs_metadata_gin;
DROP INDEX IF EXISTS idx_graph_exec_instances_frozen_gin;
DROP INDEX IF EXISTS idx_graph_exec_instances_node_states_gin;
DROP INDEX IF EXISTS idx_graph_definitions_trigger_config_gin;
DROP INDEX IF EXISTS idx_graph_definitions_edges_gin;
DROP INDEX IF EXISTS idx_graph_definitions_node_refs_gin;
DROP INDEX IF EXISTS idx_agent_memories_structured_data_gin;
DROP INDEX IF EXISTS idx_registry_functions_tags_gin;
DROP INDEX IF EXISTS idx_functions_metadata_gin;
DROP INDEX IF EXISTS idx_functions_env_vars_gin;
DROP INDEX IF EXISTS idx_functions_playground_config_gin;
DROP INDEX IF EXISTS idx_users_settings_gin;
DROP INDEX IF EXISTS idx_users_provider_data_gin;

-- ============================================
-- 9. Drop Generated Columns (must drop indexes first)
-- ============================================

-- Note: In a real rollback, you'd want to drop generated columns
-- However, if data was written assuming these columns exist,
-- we should keep them and just remove the indexes.
-- Uncomment below to actually drop generated columns:

-- ALTER TABLE graph_definitions DROP COLUMN IF EXISTS is_published_boolean;
-- ALTER TABLE graph_definitions DROP COLUMN IF EXISTS trigger_type;
-- ALTER TABLE registry_functions DROP COLUMN IF EXISTS determinism_level;
-- ALTER TABLE email_events DROP COLUMN IF EXISTS event_category;
-- ALTER TABLE users DROP COLUMN IF EXISTS profile_visibility;
-- ALTER TABLE functions DROP COLUMN IF EXISTS playground_runtime_type;
