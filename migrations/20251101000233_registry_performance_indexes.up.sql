-- Add comprehensive performance indexes for registry tables
-- These indexes target the most common query patterns in the function registry

-- ============================================
-- Registry Functions Indexes
-- ============================================
-- Indexes that use trust_score/popularity_score/reliability_score only when those columns exist

DO $$
DECLARE
  has_trust boolean;
  has_popularity boolean;
  has_reliability boolean;
BEGIN
  SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_functions' AND column_name = 'trust_score') INTO has_trust;
  SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_functions' AND column_name = 'popularity_score') INTO has_popularity;
  SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_functions' AND column_name = 'reliability_score') INTO has_reliability;

  IF has_popularity THEN
    CREATE INDEX IF NOT EXISTS idx_registry_functions_visibility_popularity ON registry_functions(visibility, popularity_score DESC) WHERE visibility = 'public';
    CREATE INDEX IF NOT EXISTS idx_registry_functions_category_popularity ON registry_functions(category, popularity_score DESC) WHERE visibility = 'public';
  END IF;
  IF has_reliability AND has_popularity THEN
    CREATE INDEX IF NOT EXISTS idx_registry_functions_reliability_popularity ON registry_functions(reliability_score DESC, popularity_score DESC) WHERE visibility = 'public';
  END IF;
  IF has_trust AND has_popularity THEN
    CREATE INDEX IF NOT EXISTS idx_registry_functions_trust_popularity ON registry_functions(trust_score DESC, popularity_score DESC) WHERE visibility = 'public';
  END IF;
  IF has_trust THEN
    CREATE INDEX IF NOT EXISTS idx_registry_functions_high_trust ON registry_functions(trust_score DESC) WHERE trust_score >= 0.8 AND visibility = 'public';
  END IF;
END $$;

-- Full-text search optimization (composite with popularity for ranking)
-- Full-text search (name + description only for schema compatibility)
CREATE INDEX IF NOT EXISTS idx_registry_functions_search_popularity
ON registry_functions USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')))
WHERE visibility = 'public';

-- Function lookup by author/name (already exists but ensure it's optimized)
CREATE INDEX IF NOT EXISTS idx_registry_functions_author_name_lookup
ON registry_functions(author, name)
WHERE visibility = 'public';

-- Owner-based queries for management
CREATE INDEX IF NOT EXISTS idx_registry_functions_owner_user
ON registry_functions(owner_user_id)
WHERE owner_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registry_functions_tenant_owner
ON registry_functions(tenant_id, owner_user_id);

-- ============================================
-- Registry Function Versions Indexes
-- ============================================

-- Version lookup optimization
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_function_version
ON registry_function_versions(function_id, version);

-- Runtime-based filtering for search
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_runtime_published
ON registry_function_versions(runtime, published_at DESC);

-- Deterministic functions for caching optimization
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_deterministic
ON registry_function_versions(deterministic)
WHERE deterministic = true;

-- Backend deployment queries
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_backend
ON registry_function_versions(backend_id)
WHERE backend_id IS NOT NULL;

-- ============================================
-- Registry Function Executions Indexes
-- ============================================

-- High-volume execution analytics (composite indexes for time-series queries)
CREATE INDEX IF NOT EXISTS idx_registry_executions_function_timestamp
ON registry_function_executions(function_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_registry_executions_function_version_timestamp
ON registry_function_executions(function_id, version, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_registry_executions_timestamp_outcome
ON registry_function_executions(timestamp DESC, outcome);

CREATE INDEX IF NOT EXISTS idx_registry_executions_outcome_timestamp
ON registry_function_executions(outcome, timestamp DESC);

-- User/tenant analytics for billing and quotas
CREATE INDEX IF NOT EXISTS idx_registry_executions_tenant_timestamp
ON registry_function_executions(tenant_id, timestamp DESC)
WHERE tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registry_executions_user_timestamp
ON registry_function_executions(user_id, timestamp DESC)
WHERE user_id IS NOT NULL;

-- Geographic analytics
CREATE INDEX IF NOT EXISTS idx_registry_executions_geo_country_timestamp
ON registry_function_executions(geo_country, timestamp DESC)
WHERE geo_country IS NOT NULL;

-- Verification queries (only when columns exist)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'verified_at') THEN
    CREATE INDEX IF NOT EXISTS idx_registry_executions_verified_at ON registry_function_executions(verified_at DESC) WHERE verified_at IS NOT NULL;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'verification_status') THEN
    CREATE INDEX IF NOT EXISTS idx_registry_executions_verification_status ON registry_function_executions(verification_status) WHERE verification_status IS NOT NULL;
  END IF;
END $$;

-- Cached execution queries (for cache hit analysis)
CREATE INDEX IF NOT EXISTS idx_registry_executions_cached_timestamp
ON registry_function_executions(cached, timestamp DESC)
WHERE cached = true;

-- ============================================
-- Registry Function Ratings Indexes
-- ============================================

-- Trust score sorting (already exists but ensure it's optimized)
-- The existing index should be sufficient, but let's ensure it's a DESC index
DROP INDEX IF EXISTS idx_registry_function_ratings_score;
CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_overall_score
ON registry_function_ratings(overall_score DESC);

-- Performance-based sorting for discovery
CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_reliability
ON registry_function_ratings(reliability_score DESC);

CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_latency
ON registry_function_ratings(p95_latency_ms ASC, avg_latency_ms ASC);

-- Trust score components for analytics
CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_trust_components
ON registry_function_ratings(trust_score DESC, reliability_score DESC, consumer_diversity DESC);

-- ============================================
-- Registry Executions Public Indexes
-- ============================================

-- Public execution lookup and sharing
CREATE INDEX IF NOT EXISTS idx_registry_executions_public_shareable_created
ON registry_executions_public(shareable, created_at DESC)
WHERE shareable = true;

CREATE INDEX IF NOT EXISTS idx_registry_executions_public_function_created
ON registry_executions_public(function_id, created_at DESC);

-- Replay verification queries (only when verified_at exists)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_executions_public' AND column_name = 'verified_at') THEN
    CREATE INDEX IF NOT EXISTS idx_registry_executions_public_verified ON registry_executions_public(verified_at DESC) WHERE verified_at IS NOT NULL;
  END IF;
END $$;

-- ============================================
-- Execution Resource Usage Indexes
-- ============================================

-- Resource usage analytics
CREATE INDEX IF NOT EXISTS idx_execution_resource_usage_execution
ON execution_resource_usage(execution_id)
WHERE execution_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_execution_resource_usage_created
ON execution_resource_usage(created_at DESC);

-- ============================================
-- Function Signature Indexes
-- ============================================

-- Signature verification queries
CREATE INDEX IF NOT EXISTS idx_registry_function_signatures_function_version
ON registry_function_signatures(function_version_id);

CREATE INDEX IF NOT EXISTS idx_registry_function_signatures_key_id
ON registry_function_signatures(key_id);

-- Only when verified_at exists on registry_function_signatures
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_function_signatures' AND column_name = 'verified_at') THEN
    CREATE INDEX IF NOT EXISTS idx_registry_function_signatures_valid ON registry_function_signatures(is_valid, verified_at DESC);
  END IF;
END $$;

-- ============================================
-- Malware Scan Indexes
-- ============================================

-- Security scan queries
CREATE INDEX IF NOT EXISTS idx_registry_function_malware_scans_function_version
ON registry_function_malware_scans(function_version_id);

CREATE INDEX IF NOT EXISTS idx_registry_function_malware_scans_status
ON registry_function_malware_scans(status, scanned_at DESC);

CREATE INDEX IF NOT EXISTS idx_registry_function_malware_scans_risk
ON registry_function_malware_scans(risk_score DESC)
WHERE status = 'completed';

-- ============================================
-- Approval Workflow Indexes
-- ============================================

-- Approval management queries
CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_function_version
ON registry_function_approvals(function_version_id);

CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_status
ON registry_function_approvals(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_assigned
ON registry_function_approvals(assigned_to, status)
WHERE assigned_to IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_trust_level
ON registry_function_approvals(trust_level, status);

CREATE INDEX IF NOT EXISTS idx_registry_function_approvals_deadline
ON registry_function_approvals(review_deadline)
WHERE review_deadline IS NOT NULL AND status = 'pending';

-- ============================================
-- Verification Status Indexes
-- ============================================

-- Overall verification status queries
CREATE INDEX IF NOT EXISTS idx_registry_function_verification_status_function_version
ON registry_function_verification_status(function_version_id);

CREATE INDEX IF NOT EXISTS idx_registry_function_verification_status_overall
ON registry_function_verification_status(overall_status, last_verified_at DESC);

CREATE INDEX IF NOT EXISTS idx_registry_function_verification_status_next_check
ON registry_function_verification_status(next_verification_at)
WHERE next_verification_at IS NOT NULL;

-- ============================================
-- Partial Indexes for Common Filtered Queries
-- ============================================

-- Active public functions (most common query)
CREATE INDEX IF NOT EXISTS idx_registry_functions_active_public
ON registry_functions(created_at DESC)
WHERE visibility = 'public';

-- Recent executions (for analytics dashboards)
-- Note: Partial index with NOW() would require IMMUTABLE; use full index instead for compatibility
CREATE INDEX IF NOT EXISTS idx_registry_executions_recent
ON registry_function_executions(timestamp DESC);

-- Successful executions (for success rate calculations)
CREATE INDEX IF NOT EXISTS idx_registry_executions_successful
ON registry_function_executions(function_id, timestamp DESC)
WHERE outcome = 'success';

-- Cached executions (for cache performance monitoring)
CREATE INDEX IF NOT EXISTS idx_registry_executions_cached_recent
ON registry_function_executions(timestamp DESC)
WHERE cached = true;

-- Deterministic functions (for caching opportunities)
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_deterministic_runtime
ON registry_function_versions(runtime)
WHERE deterministic = true;

-- High-trust functions index created in DO block above when trust_score exists