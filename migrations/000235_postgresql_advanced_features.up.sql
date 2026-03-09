-- Advanced PostgreSQL features for better performance and functionality
-- This migration adds generated columns, enhanced JSONB operators, and other PostgreSQL-specific optimizations

-- ============================================
-- Generated Columns for Computed Fields
-- ============================================

-- Note: Complex generated columns removed due to PostgreSQL limitations
-- Simple generated columns would be added here if needed

-- ============================================
-- Enhanced JSONB Operators and Indexes
-- ============================================

-- Create functional indexes for JSONB path operations
CREATE INDEX IF NOT EXISTS idx_registry_functions_tags_array
ON registry_functions USING gin ((tags->'tags'));

-- registry_functions.capabilities is TEXT[]; use GIN on array. Skip jsonb_object_keys.
CREATE INDEX IF NOT EXISTS idx_registry_functions_capabilities_gin
ON registry_functions USING gin (capabilities);

-- registry_function_versions.capabilities is JSONB
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_capabilities_keys
ON registry_function_versions USING gin (capabilities);

-- Create indexes for specific JSONB paths (registry_functions.capabilities is TEXT[] so skip these for registry_functions)
-- Only for registry_function_versions if it has JSONB capabilities; handled by GIN on capabilities above

-- ============================================
-- BRIN Indexes for Large Time-Series Tables
-- ============================================

-- BRIN indexes for audit events (more efficient than B-tree for time-series)
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp_brin
ON audit_events USING brin(timestamp)
WHERE timestamp > '2024-01-01';

-- BRIN indexes for usage events
CREATE INDEX IF NOT EXISTS idx_usage_events_timestamp_brin
ON usage_events USING brin(timestamp)
WHERE timestamp > '2024-01-01';

-- BRIN indexes for alerts
CREATE INDEX IF NOT EXISTS idx_alerts_created_at_brin
ON alerts USING brin(created_at)
WHERE created_at > '2024-01-01';

-- ============================================
-- Expression Indexes for Computed Values
-- ============================================

-- Expression index for function age skipped (now() is not IMMUTABLE)

-- Expression index for execution duration categories (expression must be in double parentheses)
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_duration_category
ON registry_function_executions ((
    CASE
        WHEN duration_ms < 100 THEN 'fast'
        WHEN duration_ms < 1000 THEN 'medium'
        WHEN duration_ms < 10000 THEN 'slow'
        ELSE 'very_slow'
    END
));

-- Expression index for function size categories (skip if bundle_size column does not exist)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registry_function_versions' AND column_name = 'bundle_size') THEN
    CREATE INDEX IF NOT EXISTS idx_registry_function_versions_size_category
    ON registry_function_versions ((
        CASE
            WHEN bundle_size < 1024 THEN 'tiny'
            WHEN bundle_size < 10240 THEN 'small'
            WHEN bundle_size < 102400 THEN 'medium'
            WHEN bundle_size < 1048576 THEN 'large'
            ELSE 'huge'
        END
    ));
  END IF;
END $$;

-- ============================================
-- Partial Indexes with Expressions
-- ============================================

-- Index for recently active functions (avoid NOW() in predicate for IMMUTABILITY)
CREATE INDEX IF NOT EXISTS idx_registry_functions_recently_active
ON registry_functions (id, updated_at DESC);

-- Partial index for functions with high usage (executions in last 30 days)
CREATE INDEX IF NOT EXISTS idx_registry_functions_high_usage
ON registry_functions (id, popularity_score DESC)
WHERE popularity_score > 1000;

-- Partial index for functions with recent errors (error rate > 10% in last 24h)
CREATE INDEX IF NOT EXISTS idx_registry_functions_error_prone
ON registry_functions (id, reliability_score ASC)
WHERE reliability_score < 0.9;

-- ============================================
-- Advanced Text Search Features
-- ============================================

-- Add weighted full-text search configuration (expression in double parentheses for GIN)
CREATE INDEX IF NOT EXISTS idx_registry_functions_weighted_search
ON registry_functions USING gin ((
    setweight(to_tsvector('english', coalesce(author, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(name, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'C') ||
    setweight(to_tsvector('english', coalesce(category, '')), 'D')
))
WHERE visibility = 'public';

-- ============================================
-- Advanced Constraints and Triggers
-- ============================================

-- Function to validate trust score ranges
CREATE OR REPLACE FUNCTION validate_trust_score()
RETURNS trigger AS $$
BEGIN
    IF NEW.trust_score < 0 OR NEW.trust_score > 1 THEN
        RAISE EXCEPTION 'trust_score must be between 0 and 1';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for trust score validation
DROP TRIGGER IF EXISTS validate_trust_score_trigger ON registry_function_ratings;
CREATE TRIGGER validate_trust_score_trigger
    BEFORE INSERT OR UPDATE ON registry_function_ratings
    FOR EACH ROW EXECUTE FUNCTION validate_trust_score();

-- Function to automatically update function popularity based on executions
CREATE OR REPLACE FUNCTION update_function_popularity()
RETURNS trigger AS $$
BEGIN
    -- Increment popularity score when function is executed
    UPDATE registry_functions
    SET popularity_score = popularity_score + 1,
        updated_at = NOW()
    WHERE id = NEW.function_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for popularity updates (optional - enable if needed)
-- DROP TRIGGER IF EXISTS update_popularity_on_execution ON registry_function_executions;
-- CREATE TRIGGER update_popularity_on_execution
--     AFTER INSERT ON registry_function_executions
--     FOR EACH ROW EXECUTE FUNCTION update_function_popularity();

-- ============================================
-- Advanced Views for Analytics
-- ============================================

-- Create a view for function performance analytics (omit trust_score/popularity_score for schema compatibility)
DROP VIEW IF EXISTS function_performance_analytics;
CREATE VIEW function_performance_analytics AS
SELECT
    f.id,
    f.author,
    f.name,
    COUNT(e.id) as total_executions,
    AVG(e.duration_ms) as avg_execution_time,
    MIN(e.duration_ms) as min_execution_time,
    MAX(e.duration_ms) as max_execution_time,
    COUNT(CASE WHEN e.outcome = 'success' THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0) as success_rate,
    COUNT(CASE WHEN e.cached = true THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0) as cache_hit_rate,
    MAX(e.timestamp) as last_execution_at
FROM registry_functions f
LEFT JOIN registry_function_executions e ON f.id = e.function_id
WHERE f.visibility = 'public'
GROUP BY f.id, f.author, f.name;

-- Create a view for tenant usage analytics
DROP VIEW IF EXISTS tenant_usage_analytics;
CREATE VIEW tenant_usage_analytics AS
SELECT
    t.id as tenant_id,
    t.name as tenant_name,
    COUNT(DISTINCT f.id) as functions_count,
    COUNT(e.id) as total_executions,
    SUM(e.duration_ms) as total_execution_time,
    AVG(e.duration_ms) as avg_execution_time,
    COUNT(DISTINCT DATE(e.timestamp)) as active_days,
    MAX(e.timestamp) as last_activity_at
FROM tenants t
LEFT JOIN registry_functions f ON t.id = f.tenant_id
LEFT JOIN registry_function_executions e ON f.id = e.function_id
GROUP BY t.id, t.name;

-- ============================================
-- Comments for Documentation
-- ============================================

COMMENT ON VIEW function_performance_analytics IS 'Aggregated view of function performance metrics for analytics and monitoring';
COMMENT ON VIEW tenant_usage_analytics IS 'Aggregated view of tenant usage patterns for billing and analytics';
