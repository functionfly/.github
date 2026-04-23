-- Migration: Optimized Indexes for Performance
-- GIN indexes for JSONB, covering indexes, computed columns
-- Created: 2026-04-19

-- ============================================
-- 1. GIN Indexes for JSONB Columns
-- For flexible querying of settings, metadata, configs
-- ============================================

-- Users table - GIN on settings JSONB
CREATE INDEX IF NOT EXISTS idx_users_settings_gin 
ON users USING GIN (settings jsonb_path_ops);

-- Functions table - GIN on playground_config
CREATE INDEX IF NOT EXISTS idx_functions_playground_config_gin 
ON functions USING GIN (playground_config jsonb_path_ops);

-- Functions table - GIN on env_vars
CREATE INDEX IF NOT EXISTS idx_functions_env_vars_gin 
ON functions USING GIN (env_vars);

-- Functions table - GIN on metadata
CREATE INDEX IF NOT EXISTS idx_functions_metadata_gin 
ON functions USING GIN (metadata jsonb_path_ops);

-- Registry functions - GIN on tags
CREATE INDEX IF NOT EXISTS idx_registry_functions_tags_gin 
ON registry_functions USING GIN (tags);

-- Agent memories - GIN on structured_data
CREATE INDEX IF NOT EXISTS idx_agent_memories_structured_data_gin 
ON agent_memories USING GIN (structured_data jsonb_path_ops);

-- Graph definitions - GIN on node_refs
CREATE INDEX IF NOT EXISTS idx_graph_definitions_node_refs_gin 
ON graph_definitions USING GIN (node_refs);

-- Graph definitions - GIN on edges
CREATE INDEX IF NOT EXISTS idx_graph_definitions_edges_gin 
ON graph_definitions USING GIN (edges);

-- Graph definitions - GIN on trigger_config
CREATE INDEX IF NOT EXISTS idx_graph_definitions_trigger_config_gin 
ON graph_definitions USING GIN (trigger_config jsonb_path_ops);

-- Graph execution instances - GIN on node_states
CREATE INDEX IF NOT EXISTS idx_graph_exec_instances_node_states_gin 
ON graph_execution_instances USING GIN (node_states jsonb_path_ops);

-- Graph execution instances - GIN on frozen_nodes and frozen_edges
CREATE INDEX IF NOT EXISTS idx_graph_exec_instances_frozen_gin 
ON graph_execution_instances USING GIN (frozen_nodes, frozen_edges);

-- Factory pipeline runs - GIN on metadata
CREATE INDEX IF NOT EXISTS idx_factory_pipeline_runs_metadata_gin 
ON factory_pipeline_runs USING GIN (metadata jsonb_path_ops);

-- Factory pipeline jobs - GIN on metadata
CREATE INDEX IF NOT EXISTS idx_factory_pipeline_jobs_metadata_gin 
ON factory_pipeline_jobs USING GIN (metadata jsonb_path_ops);

-- Trust API scopes - GIN on scopes JSONB
CREATE INDEX IF NOT EXISTS idx_trustapi_client_keys_scopes_gin 
ON trustapi_client_keys USING GIN (scopes);

CREATE INDEX IF NOT EXISTS idx_trustapi_service_keys_scopes_gin 
ON trustapi_service_keys USING GIN (scopes);

-- Team memories - GIN on tags
CREATE INDEX IF NOT EXISTS idx_team_memories_tags_gin 
ON team_memories USING GIN (tags);

-- Performance metrics - GIN on labels
CREATE INDEX IF NOT EXISTS idx_performance_metrics_labels_gin 
ON performance_metrics USING GIN (labels);

-- Monitoring alerts - GIN on metadata
CREATE INDEX IF NOT EXISTS idx_monitoring_alerts_metadata_gin 
ON monitoring_alerts USING GIN (metadata jsonb_path_ops);

-- Webhook payloads - GIN on payload
CREATE INDEX IF NOT EXISTS idx_stored_webhook_payloads_payload_gin 
ON stored_webhook_payloads USING GIN (payload jsonb_path_ops);

-- Email events - GIN on event_data and metadata
CREATE INDEX IF NOT EXISTS idx_email_events_event_data_gin 
ON email_events USING GIN (event_data jsonb_path_ops);

CREATE INDEX IF NOT EXISTS idx_email_events_metadata_gin 
ON email_events USING GIN (metadata jsonb_path_ops);

-- Provider data - GIN on provider_data
CREATE INDEX IF NOT EXISTS idx_users_provider_data_gin 
ON users USING GIN (provider_data jsonb_path_ops);

-- ============================================
-- 2. Covering Indexes (INCLUDE clause)
-- Reduces heap lookups for common queries
-- ============================================

-- Billing: Cost allocation tenant queries (includes all needed columns)
CREATE INDEX IF NOT EXISTS idx_cost_allocation_tenant_covering 
ON cost_allocation_entries (tenant_id, timestamp DESC)
INCLUDE (function_id, function_name, total_cost_cents, execution_outcome, cached);

-- Billing: Cost allocation function queries
CREATE INDEX IF NOT EXISTS idx_cost_allocation_function_covering 
ON cost_allocation_entries (function_id, timestamp DESC)
INCLUDE (tenant_id, total_cost_cents, execution_outcome, duration_ms);

-- Registry executions: Tenant dashboard queries
CREATE INDEX IF NOT EXISTS idx_registry_exec_tenant_covering 
ON registry_function_executions (tenant_id, timestamp DESC)
INCLUDE (function_id, outcome, duration_ms, cached, status_code)
WHERE tenant_id IS NOT NULL;

-- Registry executions: Function detail queries
CREATE INDEX IF NOT EXISTS idx_registry_exec_function_covering 
ON registry_function_executions (function_id, timestamp DESC)
INCLUDE (outcome, duration_ms, cached, status_code, error_code);

-- User activity feed queries (user_id, created_at)
CREATE INDEX IF NOT EXISTS idx_user_activity_user_covering 
ON user_activity (user_id, created_at DESC)
INCLUDE (activity_type, title, metadata);

-- Apps by tenant (common admin query)
CREATE INDEX IF NOT EXISTS idx_apps_tenant_covering 
ON apps (tenant_id, created_at DESC)
INCLUDE (name, slug);

-- Backends by app (routing queries)
CREATE INDEX IF NOT EXISTS idx_backends_app_covering 
ON backends (app_id, enabled DESC, last_used_at DESC)
INCLUDE (provider, region, url, shared_secret);

-- Subscriptions by tenant (billing queries)
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_covering 
ON subscriptions (tenant_id, current_period_end DESC)
INCLUDE (status, pricing_tier_id, stripe_subscription_id);

-- Invoices by tenant (billing queries)
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_covering 
ON invoices (tenant_id, created_at DESC)
INCLUDE (status, amount_due_cents, currency, period_start, period_end);

-- Team memberships by team (team detail page)
CREATE INDEX IF NOT EXISTS idx_team_memberships_team_covering 
ON team_memberships (team_id, added_at DESC)
INCLUDE (user_id, role, added_by);

-- Team memberships by user (user's teams)
CREATE INDEX IF NOT EXISTS idx_team_memberships_user_covering 
ON team_memberships (user_id)
INCLUDE (team_id, role);

-- Magic links by email (auth flow)
CREATE INDEX IF NOT EXISTS idx_magic_links_email_covering 
ON magic_links (email, created_at DESC)
INCLUDE (token, used, expires_at)
WHERE NOT used;

-- Sessions by user (security/admin queries)
CREATE INDEX IF NOT EXISTS idx_sessions_user_covering 
ON sessions (user_id, created_at DESC)
INCLUDE (token, expires_at, last_active_at, is_valid);

-- Refresh tokens by user
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_covering 
ON refresh_tokens (user_id, created_at DESC)
INCLUDE (token_hash, expires_at, revoked)
WHERE NOT revoked;

-- Provider tokens by user
CREATE INDEX IF NOT EXISTS idx_provider_tokens_user_covering 
ON provider_tokens (user_id, provider)
INCLUDE (token_encrypted, status, last_used_at);

-- Audit events by tenant (compliance queries)
CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_covering 
ON audit_events (tenant_id, timestamp DESC)
INCLUDE (action, resource_type, resource_id, actor_user_id, success)
WHERE tenant_id IS NOT NULL;

-- Usage rollups by tenant (billing aggregation)
CREATE INDEX IF NOT EXISTS idx_usage_rollups_tenant_covering 
ON usage_rollups (tenant_id, period_date DESC)
INCLUDE (event_type, total_quantity);

-- ============================================
-- 3. Computed Columns (Generated Columns)
-- For frequently queried JSONB paths
-- ============================================

-- Users: Add generated column for profile visibility
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS profile_visibility VARCHAR(50) 
GENERATED ALWAYS AS (settings->>'profileVisibility') STORED;

CREATE INDEX IF NOT EXISTS idx_users_profile_visibility 
ON users (profile_visibility) 
WHERE profile_visibility IS NOT NULL;

-- Functions: Add generated column for runtime type from playground_config
ALTER TABLE functions 
ADD COLUMN IF NOT EXISTS playground_runtime_type VARCHAR(50) 
GENERATED ALWAYS AS (playground_config->>'runtime') STORED;

CREATE INDEX IF NOT EXISTS idx_functions_playground_runtime 
ON functions (playground_runtime_type)
WHERE playground_runtime_type IS NOT NULL;

-- Graph definitions: Add generated column for trigger type
ALTER TABLE graph_definitions 
ADD COLUMN IF NOT EXISTS trigger_type VARCHAR(50) 
GENERATED ALWAYS AS (trigger_config->>'type') STORED;

CREATE INDEX IF NOT EXISTS idx_graph_definitions_trigger_type 
ON graph_definitions (trigger_type)
WHERE trigger_type IS NOT NULL;

-- Graph definitions: Add generated column for publish status
ALTER TABLE graph_definitions 
ADD COLUMN IF NOT EXISTS is_published_boolean BOOLEAN 
GENERATED ALWAYS AS (COALESCE((trigger_config->>'published')::boolean, false)) STORED;

CREATE INDEX IF NOT EXISTS idx_graph_definitions_published 
ON graph_definitions (is_published_boolean, created_at DESC)
WHERE is_published_boolean = true;

-- Registry functions: Add generated column for determinism classification
ALTER TABLE registry_functions 
ADD COLUMN IF NOT EXISTS determinism_level VARCHAR(50) 
GENERATED ALWAYS AS (metadata->>'determinism_level') STORED;

CREATE INDEX IF NOT EXISTS idx_registry_functions_determinism 
ON registry_functions (determinism_level)
WHERE determinism_level IS NOT NULL;

-- Email events: Add generated column for event category
ALTER TABLE email_events 
ADD COLUMN IF NOT EXISTS event_category VARCHAR(50) 
GENERATED ALWAYS AS (split_part(event_type, '.', 1)) STORED;

CREATE INDEX IF NOT EXISTS idx_email_events_category 
ON email_events (event_category, timestamp DESC);

-- ============================================
-- 4. Partial Indexes for Common Filtered Queries
-- ============================================

-- Active sessions only
CREATE INDEX IF NOT EXISTS idx_sessions_active 
ON sessions (user_id, last_active_at DESC)
WHERE is_valid = true AND expires_at > NOW();

-- Failed executions (error investigation)
CREATE INDEX IF NOT EXISTS idx_registry_exec_failed 
ON registry_function_executions (function_id, timestamp DESC)
WHERE outcome != 'success';

-- Cached executions (cache hit analysis)
CREATE INDEX IF NOT EXISTS idx_registry_exec_cached 
ON registry_function_executions (function_id, timestamp DESC)
WHERE cached = true;

-- Pending verification (trust verification queue)
CREATE INDEX IF NOT EXISTS idx_registry_exec_pending_verification 
ON registry_function_executions (function_id, timestamp DESC)
WHERE verification_status = 'pending';

-- Unprocessed webhook payloads
CREATE INDEX IF NOT EXISTS idx_webhook_payloads_pending 
ON stored_webhook_payloads (processing_status, created_at)
WHERE processing_status = 'pending';

-- Failed webhooks (for replay)
CREATE INDEX IF NOT EXISTS idx_webhook_payloads_failed 
ON stored_webhook_payloads (event_type, created_at DESC)
WHERE processing_status = 'failed';

-- Unread notifications
CREATE INDEX IF NOT EXISTS idx_notifications_unread 
ON notifications (user_id, created_at DESC)
WHERE read_at IS NULL;

-- Active alerts
CREATE INDEX IF NOT EXISTS idx_monitoring_alerts_active 
ON monitoring_alerts (tenant_id, severity, triggered_at DESC)
WHERE resolved_at IS NULL;

-- Active subscriptions (billing)
CREATE INDEX IF NOT EXISTS idx_subscriptions_active 
ON subscriptions (tenant_id, current_period_end DESC)
WHERE status = 'active' AND cancel_at_period_end = false;

-- Uncollected invoices
CREATE INDEX IF NOT EXISTS idx_invoices_uncollected 
ON invoices (tenant_id, due_date)
WHERE status IN ('open', 'uncollectible');

-- Active API keys
CREATE INDEX IF NOT EXISTS idx_api_keys_active 
ON api_keys (tenant_id, last_used_at DESC)
WHERE revoked_at IS NULL AND (expires_at IS NULL OR expires_at > NOW());

-- Active dunning campaigns
CREATE INDEX IF NOT EXISTS idx_dunning_campaigns_active 
ON dunning_campaigns (tenant_id, status, next_retry_at)
WHERE status IN ('active', 'retrying');

-- Pending username changes
CREATE INDEX IF NOT EXISTS idx_pending_username_pending 
ON pending_username_changes (user_id, created_at DESC)
WHERE status = 'pending';

-- Unexpired magic links
CREATE INDEX IF NOT EXISTS idx_magic_links_unexpired 
ON magic_links (email, created_at DESC)
WHERE used = false AND expires_at > NOW();

-- Active team invites
CREATE INDEX IF NOT EXISTS idx_team_invites_pending 
ON team_invites (team_id, created_at DESC)
WHERE status = 'pending' AND expires_at > NOW();

-- ============================================
-- 5. BRIN Indexes for Time-Series Data (Large Tables)
-- Block Range INdexes for very large time-series tables
-- ============================================

-- BRIN for cost_allocation_entries (if > 100M rows)
CREATE INDEX IF NOT EXISTS idx_cost_allocation_brin 
ON cost_allocation_entries USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- BRIN for registry_function_executions (if > 100M rows)
CREATE INDEX IF NOT EXISTS idx_registry_exec_brin 
ON registry_function_executions USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- BRIN for routing_events
CREATE INDEX IF NOT EXISTS idx_routing_events_brin 
ON routing_events USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- BRIN for health_checks
CREATE INDEX IF NOT EXISTS idx_health_checks_brin 
ON health_checks USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- BRIN for performance_metrics
CREATE INDEX IF NOT EXISTS idx_perf_metrics_brin 
ON performance_metrics USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- BRIN for function_logs
CREATE INDEX IF NOT EXISTS idx_function_logs_brin 
ON function_logs USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- BRIN for email_events
CREATE INDEX IF NOT EXISTS idx_email_events_brin 
ON email_events USING BRIN (timestamp)
WITH (pages_per_range = 128);

-- ============================================
-- 6. Expression Indexes for Text Search
-- ============================================

-- Lowercase email search (case-insensitive lookups)
CREATE INDEX IF NOT EXISTS idx_users_email_lower 
ON users (LOWER(email));

-- Lowercase username search
CREATE INDEX IF NOT EXISTS idx_users_username_lower 
ON users (LOWER(username))
WHERE username IS NOT NULL;

-- Function name search (registry search)
CREATE INDEX IF NOT EXISTS idx_registry_functions_name_lower 
ON registry_functions (LOWER(name));

-- Graph name search
CREATE INDEX IF NOT EXISTS idx_graph_definitions_name_lower 
ON graph_definitions (LOWER(name));

-- Full-text search on function descriptions
CREATE INDEX IF NOT EXISTS idx_registry_functions_description_search 
ON registry_functions USING GIN (to_tsvector('english', COALESCE(description, '')));

-- Full-text search on graph descriptions  
CREATE INDEX IF NOT EXISTS idx_graph_definitions_description_search 
ON graph_definitions USING GIN (to_tsvector('english', COALESCE(description, '')));

-- Full-text search on team memory content
CREATE INDEX IF NOT EXISTS idx_team_memories_content_search 
ON team_memories USING GIN (to_tsvector('english', COALESCE(content, '')));

-- ============================================
-- 7. Re-indexing Strategy Helpers
-- ============================================

-- Function to identify bloated indexes
CREATE OR REPLACE FUNCTION get_bloated_indexes(
    p_min_bloat_ratio FLOAT DEFAULT 0.3,
    p_min_wasted_bytes BIGINT DEFAULT 104857600  -- 100MB
)
RETURNS TABLE (
    schema_name TEXT,
    table_name TEXT,
    index_name TEXT,
    bloat_ratio FLOAT,
    wasted_bytes BIGINT,
    wasted_size TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH index_stats AS (
        SELECT
            schemaname::TEXT,
            relname::TEXT as tblname,
            indexrelname::TEXT as idxname,
            pg_relation_size(indexrelid) as idx_bytes,
            pg_relation_size(indrelid) as tbl_bytes
        FROM pg_stat_user_indexes
        JOIN pg_index USING (indexrelid)
        WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
    )
    SELECT 
        schemaname,
        tblname,
        idxname,
        CASE WHEN tbl_bytes > 0 
            THEN (idx_bytes::FLOAT / tbl_bytes) - 1 
            ELSE 0 
        END as bloat_ratio,
        idx_bytes as wasted_bytes,
        pg_size_pretty(idx_bytes) as wasted_size
    FROM index_stats
    WHERE (idx_bytes::FLOAT / NULLIF(tbl_bytes, 0)) - 1 > p_min_bloat_ratio
    AND idx_bytes > p_min_wasted_bytes
    ORDER BY idx_bytes DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to re-index tables safely
CREATE OR REPLACE FUNCTION safe_reindex_table(
    p_table_name TEXT,
    p_concurrent BOOLEAN DEFAULT true
)
RETURNS TABLE (index_name TEXT, success BOOLEAN, error_message TEXT) AS $$
DECLARE
    idx RECORD;
    sql_cmd TEXT;
BEGIN
    FOR idx IN 
        SELECT indexname::TEXT as idx_name
        FROM pg_indexes
        WHERE tablename = p_table_name
        AND schemaname = 'public'
    LOOP
        BEGIN
            IF p_concurrent THEN
                sql_cmd := format('REINDEX INDEX CONCURRENTLY %I', idx.idx_name);
            ELSE
                sql_cmd := format('REINDEX INDEX %I', idx.idx_name);
            END IF;
            
            EXECUTE sql_cmd;
            RETURN QUERY SELECT idx.idx_name, TRUE, NULL::TEXT;
        EXCEPTION WHEN OTHERS THEN
            RETURN QUERY SELECT idx.idx_name, FALSE, SQLERRM;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_bloated_indexes(FLOAT, BIGINT) IS 
'Returns indexes with estimated bloat above threshold. Run periodically to identify reindex candidates.';

COMMENT ON FUNCTION safe_reindex_table(TEXT, BOOLEAN) IS 
'Safely reindexes all indexes on a table. Use CONCURRENTLY to avoid locks (slower but non-blocking).';

-- ============================================
-- 8. Index Usage Monitoring
-- ============================================

-- View to track index usage statistics
CREATE OR REPLACE VIEW index_usage_stats AS
SELECT 
    schemaname,
    relname as table_name,
    indexrelname as index_name,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    CASE 
        WHEN idx_scan > 0 THEN 
            pg_size_pretty(pg_relation_size(indexrelid)::BIGINT / idx_scan)
        ELSE 'N/A'
    END as bytes_per_scan
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;

COMMENT ON VIEW index_usage_stats IS 
'Monitoring view for index usage. Indexes with 0 scans may be candidates for removal.';

-- ============================================
-- 9. Comments for Documentation
-- ============================================

COMMENT ON INDEX idx_users_settings_gin IS 'GIN index for JSONB user settings queries';
COMMENT ON INDEX idx_cost_allocation_tenant_covering IS 'Covering index avoids heap lookups for tenant billing queries';
COMMENT ON INDEX idx_registry_exec_tenant_covering IS 'Covering index for tenant execution dashboard queries';
COMMENT ON INDEX idx_users_profile_visibility IS 'Generated column index for profile privacy settings';
COMMENT ON INDEX idx_sessions_active IS 'Partial index for currently active sessions only';
COMMENT ON INDEX idx_cost_allocation_brin IS 'BRIN index for efficient time-series range scans on large tables';
