-- Migration: Create embedding audit logs table
-- Purpose: Comprehensive audit logging for embedding operations

-- Enable pgcrypto if not already enabled (for gen_random_uuid if needed)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the main audit logs table
CREATE TABLE IF NOT EXISTS embedding_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL,
    user_id UUID,
    api_key_id VARCHAR(32) NOT NULL,
    operation VARCHAR(32) NOT NULL,
    function_id UUID,
    model VARCHAR(100),
    dimensions INTEGER,
    text_hash VARCHAR(64) NOT NULL,  -- SHA256 hash for deduplication
    success BOOLEAN NOT NULL DEFAULT true,
    status VARCHAR(16) NOT NULL DEFAULT 'success',  -- success, failure, blocked
    latency_ms INTEGER,
    error_message TEXT,
    client_ip INET,
    request_id VARCHAR(64),
    token_count INTEGER,
    cost_usd NUMERIC(10, 6),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_operation CHECK (operation IN (
        'embed_generate', 'embed_batch_generate', 'embed_search', 
        'embed_query', 'rag_retrieve'
    )),
    CONSTRAINT chk_status CHECK (status IN ('success', 'failure', 'blocked'))
);

-- Indexes for efficient querying
CREATE INDEX idx_audit_tenant_created ON embedding_audit_logs (tenant_id, created_at);
CREATE INDEX idx_audit_operation_created ON embedding_audit_logs (operation, created_at);
CREATE INDEX idx_audit_text_hash ON embedding_audit_logs (text_hash);
CREATE INDEX idx_audit_api_key ON embedding_audit_logs (api_key_id, created_at);
CREATE INDEX idx_audit_request_id ON embedding_audit_logs (request_id);
CREATE INDEX idx_audit_success ON embedding_audit_logs (success, created_at);

-- Partial index for failed operations (useful for debugging)
CREATE INDEX idx_audit_failures ON embedding_audit_logs (tenant_id, created_at) 
    WHERE success = false;

-- Index for cost tracking queries
CREATE INDEX idx_audit_cost ON embedding_audit_logs (tenant_id, cost_usd, created_at) 
    WHERE cost_usd IS NOT NULL;

-- Create a summary view for quick statistics
CREATE OR REPLACE VIEW embedding_audit_summary AS
SELECT 
    tenant_id,
    DATE_TRUNC('day', created_at) as date,
    operation,
    COUNT(*) as total_count,
    SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) as failure_count,
    AVG(latency_ms) as avg_latency_ms,
    SUM(cost_usd) as total_cost_usd,
    SUM(token_count) as total_tokens
FROM embedding_audit_logs
GROUP BY tenant_id, DATE_TRUNC('day', created_at), operation;

-- Create a function to clean up old audit logs (90-day retention)
CREATE OR REPLACE FUNCTION cleanup_old_embedding_audit_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM embedding_audit_logs 
    WHERE created_at < NOW() - INTERVAL '90 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a materialized view for daily cost reports (refresh nightly)
CREATE MATERIALIZED VIEW IF NOT EXISTS embedding_daily_costs AS
SELECT 
    tenant_id,
    DATE(created_at) as date,
    model,
    SUM(cost_usd) as daily_cost,
    COUNT(*) as request_count,
    SUM(token_count) as total_tokens
FROM embedding_audit_logs
WHERE cost_usd IS NOT NULL
GROUP BY tenant_id, DATE(created_at), model;

-- Create index on materialized view
CREATE UNIQUE INDEX idx_daily_costs_pkey ON embedding_daily_costs (tenant_id, date, model);

-- Add comment for documentation
COMMENT ON TABLE embedding_audit_logs IS 'Audit trail for all embedding operations including generation, search, and RAG retrieval. Text content is stored as SHA256 hashes for privacy.';
COMMENT ON COLUMN embedding_audit_logs.text_hash IS 'SHA256 hash of the embedded text. Used for deduplication and correlation without storing sensitive content.';
COMMENT ON COLUMN embedding_audit_logs.status IS 'Event status: success, failure (error occurred), blocked (rate limit or policy violation).';
