-- Migration: 20260331000000_trust_scoring_system
-- Description: Trust Scoring System Phase 1 - Database schema for trust scoring
-- Created: 2026-03-31

BEGIN;

-- ============================================
-- Trust History Table
-- Time-series tracking of trust score changes
-- ============================================
CREATE TABLE IF NOT EXISTS trust_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Trust score components
    trust_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    reliability_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    latency_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    error_rate_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    user_rating_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    verification_bonus NUMERIC(5,2) NOT NULL DEFAULT 0,

    -- Execution metrics snapshot
    total_calls INTEGER NOT NULL DEFAULT 0,
    success_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    p50_latency_ms INTEGER NOT NULL DEFAULT 0,
    p95_latency_ms INTEGER NOT NULL DEFAULT 0,
    p99_latency_ms INTEGER NOT NULL DEFAULT 0,
    error_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    timeout_rate NUMERIC(5,2) NOT NULL DEFAULT 0,

    -- Diversity metrics
    consumer_diversity INTEGER NOT NULL DEFAULT 0,
    tenant_diversity INTEGER NOT NULL DEFAULT 0,
    user_diversity INTEGER NOT NULL DEFAULT 0,

    -- Verification status
    is_verified BOOLEAN NOT NULL DEFAULT false,
    verification_level VARCHAR(50) DEFAULT 'none',

    -- Trust tier (computed)
    trust_tier VARCHAR(20) NOT NULL DEFAULT 'untrusted',

    -- Metadata
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    calculation_version INTEGER NOT NULL DEFAULT 1
);

-- Indexes for trust_history
CREATE INDEX idx_trust_history_function_id ON trust_history(function_id);
CREATE INDEX idx_trust_history_calculated_at ON trust_history(calculated_at);
CREATE INDEX idx_trust_history_function_calculated ON trust_history(function_id, calculated_at DESC);
CREATE INDEX idx_trust_history_trust_tier ON trust_history(trust_tier);

-- ============================================
-- Execution Metrics Table
-- Aggregated execution statistics for trust score calculation
-- ============================================
CREATE TABLE IF NOT EXISTS execution_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Metrics window
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    window_type VARCHAR(20) NOT NULL DEFAULT 'hourly', -- 'hourly', 'daily', 'weekly'

    -- Execution counts
    total_calls INTEGER NOT NULL DEFAULT 0,
    successful_calls INTEGER NOT NULL DEFAULT 0,
    failed_calls INTEGER NOT NULL DEFAULT 0,
    timeout_calls INTEGER NOT NULL DEFAULT 0,
    error_calls INTEGER NOT NULL DEFAULT 0,
    cached_calls INTEGER NOT NULL DEFAULT 0,

    -- Latency metrics (in milliseconds)
    latency_min INTEGER NOT NULL DEFAULT 0,
    latency_max INTEGER NOT NULL DEFAULT 0,
    latency_sum BIGINT NOT NULL DEFAULT 0,
    latency_avg NUMERIC(10,2) NOT NULL DEFAULT 0,
    latency_p50 INTEGER NOT NULL DEFAULT 0,
    latency_p95 INTEGER NOT NULL DEFAULT 0,
    latency_p99 INTEGER NOT NULL DEFAULT 0,

    -- Error breakdown
    error_4xx_count INTEGER NOT NULL DEFAULT 0,
    error_5xx_count INTEGER NOT NULL DEFAULT 0,

    -- Unique consumers
    unique_ips INTEGER NOT NULL DEFAULT 0,
    unique_tenants INTEGER NOT NULL DEFAULT 0,
    unique_users INTEGER NOT NULL DEFAULT 0,

    -- Geographic distribution (top countries)
    geo_distribution JSONB DEFAULT '{}',

    -- Computed aggregates
    success_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    error_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    timeout_rate NUMERIC(5,2) NOT NULL DEFAULT 0,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint per function per window
    CONSTRAINT execution_metrics_function_window_unique UNIQUE (function_id, window_start, window_type)
);

-- Indexes for execution_metrics
CREATE INDEX idx_execution_metrics_function_id ON execution_metrics(function_id);
CREATE INDEX idx_execution_metrics_window ON execution_metrics(window_start, window_end);
CREATE INDEX idx_execution_metrics_function_window ON execution_metrics(function_id, window_start DESC);
CREATE INDEX idx_execution_metrics_window_type ON execution_metrics(window_type);

-- ============================================
-- Trust Score Calculation Weights Configuration
-- Allows dynamic adjustment of trust score components
-- ============================================
CREATE TABLE IF NOT EXISTS trust_score_weights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    component VARCHAR(50) NOT NULL UNIQUE,
    weight NUMERIC(5,2) NOT NULL DEFAULT 0,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID
);

-- Default weights (sum to 100)
INSERT INTO trust_score_weights (component, weight, description) VALUES
    ('reliability', 30.00, 'Execution success rate weight'),
    ('latency', 20.00, 'Response time performance weight'),
    ('error_rate', 20.00, 'Error frequency weight'),
    ('user_rating', 15.00, 'User-submitted ratings weight'),
    ('verification', 15.00, 'Verification status bonus weight')
ON CONFLICT (component) DO NOTHING;

-- ============================================
-- Add trust_score columns to existing registry_functions table
-- These are denormalized for quick access
-- ============================================
ALTER TABLE registry_functions
    ADD COLUMN IF NOT EXISTS trust_score NUMERIC(5,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS trust_tier VARCHAR(20) DEFAULT 'untrusted',
    ADD COLUMN IF NOT EXISTS trust_updated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS trust_calculation_version INTEGER DEFAULT 0;

-- Create index for quick trust-based sorting
CREATE INDEX IF NOT EXISTS idx_registry_functions_trust_score ON registry_functions(trust_score DESC);
CREATE INDEX IF NOT EXISTS idx_registry_functions_trust_tier ON registry_functions(trust_tier);

-- ============================================
-- Trust Score Refresh Job Tracking
-- ============================================
CREATE TABLE IF NOT EXISTS trust_score_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type VARCHAR(50) NOT NULL DEFAULT 'scheduled', -- 'scheduled', 'manual', 'full_recalculation'
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    functions_processed INTEGER DEFAULT 0,
    functions_total INTEGER DEFAULT 0,
    errors JSONB DEFAULT '[]',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trust_score_jobs_status ON trust_score_jobs(status);
CREATE INDEX idx_trust_score_jobs_created ON trust_score_jobs(created_at DESC);

COMMIT;
