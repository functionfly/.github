-- Advanced performance indexes for registry and execution queries
-- These indexes optimize complex queries and improve overall database performance

-- ============================================
-- Registry Function Search and Filtering
-- ============================================

-- Trust score search index will be created in a later migration

-- Composite index for function search with popularity ordering
CREATE INDEX IF NOT EXISTS idx_registry_functions_search_popular
ON registry_functions(author, name, visibility, popularity_score DESC)
WHERE visibility = 'public';

-- Multi-criteria and high-trust indexes will be created in a later migration

-- ============================================
-- Function Version Queries
-- ============================================

-- Composite index for version lookups by function and version
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_lookup
ON registry_function_versions(function_id, version);

-- Composite index for latest version queries (most common pattern)
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_latest
ON registry_function_versions(function_id, published_at DESC);

-- Composite index for deterministic functions
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_deterministic
ON registry_function_versions(function_id, deterministic, published_at DESC)
WHERE deterministic = true;

-- ============================================
-- Execution Analytics and Monitoring
-- ============================================

-- Composite index for execution analytics by function and time
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_analytics
ON registry_function_executions(function_id, timestamp DESC, status_code, outcome);

-- Composite index for execution analytics by tenant
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_tenant
ON registry_function_executions(tenant_id, timestamp DESC, outcome)
WHERE tenant_id IS NOT NULL;

-- Composite index for cached execution analysis
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_cached
ON registry_function_executions(function_id, cached, timestamp DESC)
WHERE cached = true;

-- Composite index for error analysis
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_errors
ON registry_function_executions(function_id, outcome, error_code, timestamp DESC)
WHERE outcome = 'error';

-- Composite index for geographic analytics
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_geo
ON registry_function_executions(geo_country, timestamp DESC, outcome)
WHERE geo_country IS NOT NULL;

-- ============================================
-- Verification and Security Queries
-- ============================================

-- Composite index for signature verification queries
CREATE INDEX IF NOT EXISTS idx_registry_function_signatures_verification
ON registry_function_signatures(function_version_id, is_valid, signed_at DESC);

-- Composite index for malware scan queries
CREATE INDEX IF NOT EXISTS idx_registry_function_malware_scans_status
ON registry_function_malware_scans(function_version_id, status, risk_score DESC, scanned_at DESC);

-- Composite index for approval workflow queries
CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_workflow
ON registry_function_approvals(function_version_id, status, priority, assigned_to, review_deadline);

-- Composite index for pending approvals
CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_pending
ON registry_function_approvals(status, review_deadline, priority DESC)
WHERE status = 'pending';

-- Composite index for verification status queries
CREATE INDEX IF NOT EXISTS idx_registry_function_verification_status_overall
ON registry_function_verification_status(overall_status, last_verified_at DESC, malware_risk_score);

-- ============================================
-- Partial Indexes for Common Filters
-- ============================================

-- Index for recently published functions (partial with NOW() not allowed in predicate; use full index)
CREATE INDEX IF NOT EXISTS idx_registry_functions_recent
ON registry_functions(id, author, name, created_at DESC);

-- Partial index for functions with high error rates (for monitoring)
CREATE INDEX IF NOT EXISTS idx_registry_functions_high_errors
ON registry_functions(id, author, name)
WHERE reliability_score < 0.5;

-- Partial index for functions requiring approval
CREATE INDEX IF NOT EXISTS idx_registry_function_verification_requires_approval
ON registry_function_verification_status(function_version_id, approval_required, overall_status)
WHERE approval_required = true;

-- Partial index for blocked functions
CREATE INDEX IF NOT EXISTS idx_registry_function_verification_blocked
ON registry_function_verification_status(function_version_id, overall_status)
WHERE overall_status = 'blocked';

-- ============================================
-- BRIN Indexes for Time-Series Data
-- ============================================

-- BRIN index for execution timestamps (more efficient than B-tree for time-series)
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_timestamp_brin
ON registry_function_executions USING brin(timestamp);

-- BRIN index for performance metrics timestamps
CREATE INDEX IF NOT EXISTS idx_performance_metrics_timestamp_brin
ON performance_metrics USING brin(timestamp);

-- ============================================
-- JSONB Path Indexes for Complex Queries
-- ============================================

-- GIN index for function tags (JSONB array)
CREATE INDEX IF NOT EXISTS idx_registry_functions_tags_gin
ON registry_functions USING gin(tags);

-- GIN index for function capabilities (JSONB)
CREATE INDEX IF NOT EXISTS idx_registry_functions_capabilities_gin
ON registry_functions USING gin(capabilities);

-- GIN index for function version capabilities
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_capabilities_gin
ON registry_function_versions USING gin(capabilities);

-- GIN index for approval required actions
CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_actions_gin
ON registry_function_approvals USING gin(required_actions);

-- ============================================
-- Generated Columns for Computed Fields
-- ============================================

-- Add generated column for search vector (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'registry_functions' AND column_name = 'search_vector') THEN
        ALTER TABLE registry_functions
        ADD COLUMN search_vector tsvector
        GENERATED ALWAYS AS (
            to_tsvector('english',
                coalesce(author, '') || ' ' ||
                coalesce(name, '') || ' ' ||
                coalesce(description, '') || ' ' ||
                coalesce(category, '')
            )
        ) STORED;
    END IF;
END $$;

-- Add GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_registry_functions_search_vector_gin
ON registry_functions USING gin(search_vector);

-- ============================================
-- Covering Indexes for Common Query Patterns
-- ============================================

-- Covering index for function listing will be created in a later migration

-- Covering index for function version listing
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_listing_covering
ON registry_function_versions(function_id, version, published_at DESC, deterministic, timeout_ms, memory_mb);

-- Covering index for execution monitoring
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_monitoring_covering
ON registry_function_executions(function_id, timestamp DESC, outcome, status_code, duration_ms, cached);