-- Economic Memory Layer - Phase 3 Implementation
-- Database migration for cost-per-quality metrics tracking

-- ============================================================================
-- Execution Records Table
-- Records individual LLM executions with cost and quality metrics
-- ============================================================================
CREATE TABLE IF NOT EXISTS economic_memory_executions (
    id SERIAL PRIMARY KEY,
    execution_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100),
    function_id VARCHAR(100),
    
    -- Cost metrics (in USD)
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    cost_usd DECIMAL(12, 8) DEFAULT 0.0,
    
    -- Quality metrics
    latency_ms DECIMAL(10, 2) DEFAULT 0.0,
    success BOOLEAN DEFAULT TRUE,
    error_type VARCHAR(100),
    output_quality_score DECIMAL(3, 2),  -- 0.00-1.00 scale
    user_rating DECIMAL(2, 1),  -- 0.0-5.0 scale
    
    -- Metadata
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_economic_exec_timestamp 
    ON economic_memory_executions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_economic_exec_tenant 
    ON economic_memory_executions(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_economic_exec_provider 
    ON economic_memory_executions(provider, model, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_economic_exec_function 
    ON economic_memory_executions(function_id, timestamp DESC);

-- Partial index for failed executions (for analysis)
CREATE INDEX IF NOT EXISTS idx_economic_exec_failures 
    ON economic_memory_executions(timestamp DESC) 
    WHERE success = FALSE;

-- ============================================================================
-- Aggregated Scores Table
-- Maintains running cost-quality scores per provider/model
-- ============================================================================
CREATE TABLE IF NOT EXISTS economic_memory_scores (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    
    -- Cost metrics
    avg_cost_per_1k_tokens DECIMAL(12, 8) DEFAULT 0.0,
    avg_cost_per_request DECIMAL(12, 8) DEFAULT 0.0,
    
    -- Quality metrics (0-1 scale)
    quality_score DECIMAL(3, 2) DEFAULT 0.0,
    response_time_score DECIMAL(3, 2) DEFAULT 0.0,
    token_efficiency_score DECIMAL(5, 2) DEFAULT 0.0,
    success_rate DECIMAL(3, 2) DEFAULT 1.0,
    
    -- Composite Cost-Quality Index (CQI)
    -- Higher is better - more quality per dollar spent
    cost_quality_index DECIMAL(5, 2) DEFAULT 0.0,
    
    -- Totals
    total_executions INTEGER DEFAULT 0,
    total_cost_usd DECIMAL(12, 4) DEFAULT 0.0,
    total_tokens BIGINT DEFAULT 0,
    
    -- Metadata
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(provider, model)
);

-- Index for CQI-based lookups (for routing decisions)
CREATE INDEX IF NOT EXISTS idx_economic_scores_cqi 
    ON economic_memory_scores(cost_quality_index DESC);

-- Index for provider lookups
CREATE INDEX IF NOT EXISTS idx_economic_scores_provider 
    ON economic_memory_scores(provider, model);

-- ============================================================================
-- Tenant Cost Summary View
-- Provides aggregated cost data per tenant for billing/insights
-- ============================================================================
CREATE OR REPLACE VIEW economic_memory_tenant_summary AS
SELECT 
    tenant_id,
    DATE_TRUNC('day', timestamp) AS day,
    provider,
    model,
    COUNT(*) AS executions,
    SUM(cost_usd) AS total_cost_usd,
    SUM(total_tokens) AS total_tokens,
    AVG(latency_ms) AS avg_latency_ms,
    SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) AS success_rate,
    AVG(output_quality_score) AS avg_quality_score
FROM economic_memory_executions
WHERE tenant_id IS NOT NULL
GROUP BY tenant_id, DATE_TRUNC('day', timestamp), provider, model;

-- ============================================================================
-- Functions for Updating Aggregated Scores
-- ============================================================================

-- Function to update economic scores after an execution record is inserted
CREATE OR REPLACE FUNCTION update_economic_memory_scores()
RETURNS TRIGGER AS $$
DECLARE
    v_total_executions INTEGER;
    v_total_cost DECIMAL(12, 4);
    v_total_tokens BIGINT;
    v_avg_cost_per_1k DECIMAL(12, 8);
    v_avg_latency DECIMAL(10, 2);
    v_success_rate DECIMAL(3, 2);
    v_quality_score DECIMAL(3, 2);
    v_response_time_score DECIMAL(3, 2);
    v_cqi DECIMAL(5, 2);
BEGIN
    -- Calculate aggregated metrics for this provider/model
    SELECT 
        COUNT(*),
        SUM(cost_usd),
        SUM(total_tokens),
        AVG(CASE WHEN total_tokens > 0 THEN (cost_usd / total_tokens * 1000) ELSE 0 END),
        AVG(latency_ms),
        SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*),
        AVG(COALESCE(output_quality_score, 0.75))  -- Default to 0.75 if no quality score
    INTO 
        v_total_executions,
        v_total_cost,
        v_total_tokens,
        v_avg_cost_per_1k,
        v_avg_latency,
        v_success_rate,
        v_quality_score
    FROM economic_memory_executions
    WHERE provider = NEW.provider AND model = NEW.model;
    
    -- Calculate response time score (inverse of latency, normalized)
    v_response_time_score := GREATEST(0.0, 1.0 - (v_avg_latency / 1000.0));
    
    -- Calculate CQI: (quality * response_time * success_rate) / cost_per_1k * 1000
    IF v_avg_cost_per_1k > 0 THEN
        v_cqi := (v_quality_score * v_response_time_score * v_success_rate * 100) / (v_avg_cost_per_1k * 1000);
        v_cqi := GREATEST(0.0, LEAST(100.0, v_cqi));  -- Clamp to 0-100
    ELSE
        v_cqi := v_quality_score * 100;  -- Max quality score when free
    END IF;
    
    -- Insert or update the scores record
    INSERT INTO economic_memory_scores (
        provider,
        model,
        avg_cost_per_1k_tokens,
        avg_cost_per_request,
        quality_score,
        response_time_score,
        success_rate,
        cost_quality_index,
        total_executions,
        total_cost_usd,
        total_tokens,
        last_updated
    )
    VALUES (
        NEW.provider,
        NEW.model,
        v_avg_cost_per_1k,
        COALESCE(v_total_cost / NULLIF(v_total_executions, 0), 0),
        v_quality_score,
        v_response_time_score,
        v_success_rate,
        v_cqi,
        v_total_executions,
        v_total_cost,
        v_total_tokens,
        NOW()
    )
    ON CONFLICT (provider, model) DO UPDATE SET
        avg_cost_per_1k_tokens = EXCLUDED.avg_cost_per_1k_tokens,
        avg_cost_per_request = EXCLUDED.avg_cost_per_request,
        quality_score = EXCLUDED.quality_score,
        response_time_score = EXCLUDED.response_time_score,
        success_rate = EXCLUDED.success_rate,
        cost_quality_index = EXCLUDED.cost_quality_index,
        total_executions = EXCLUDED.total_executions,
        total_cost_usd = EXCLUDED.total_cost_usd,
        total_tokens = EXCLUDED.total_tokens,
        last_updated = EXCLUDED.last_updated;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update scores on new execution record
DROP TRIGGER IF EXISTS trg_update_economic_scores ON economic_memory_executions;
CREATE TRIGGER trg_update_economic_scores
    AFTER INSERT ON economic_memory_executions
    FOR EACH ROW
    EXECUTE FUNCTION update_economic_memory_scores();

-- ============================================================================
-- Partitioning (optional - for high-volume installations)
-- Uncomment if expecting > 1M records per month
-- ============================================================================
-- CREATE TABLE economic_memory_executions_partitioned (
--     LIKE economic_memory_executions INCLUDING ALL
-- ) PARTITION BY RANGE (timestamp);

-- Create monthly partitions
-- CREATE TABLE economic_memory_executions_y2024m01 
--     PARTITION OF economic_memory_executions_partitioned
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- ============================================================================
-- Comments for documentation
-- ============================================================================
COMMENT ON TABLE economic_memory_executions IS 
    'Individual LLM execution records with cost and quality metrics for economic routing';
COMMENT ON TABLE economic_memory_scores IS 
    'Aggregated cost-quality metrics per provider/model for routing decisions';
COMMENT ON COLUMN economic_memory_scores.cost_quality_index IS 
    'Cost-Quality Index: Higher = better value (quality per dollar spent)';

-- ============================================================================
-- Sample data for testing (optional - remove for production)
-- ============================================================================
-- INSERT INTO economic_memory_scores (
--     provider, model, avg_cost_per_1k_tokens, quality_score, 
--     response_time_score, success_rate, cost_quality_index, total_executions
-- ) VALUES 
-- ('openai', 'gpt-4o-mini', 0.000375, 0.82, 0.95, 0.99, 82.0, 1000),
-- ('groq', 'llama-3.1-8b', 0.0001, 0.75, 0.98, 0.97, 75.0, 500),
-- ('anthropic', 'claude-3-haiku', 0.000625, 0.80, 0.92, 0.98, 64.0, 800)
-- ON CONFLICT DO NOTHING;
