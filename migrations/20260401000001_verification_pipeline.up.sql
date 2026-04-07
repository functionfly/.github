-- Migration: 20260401000000_verification_pipeline
-- Description: Verification Pipeline - Job queue and results for automated function verification
-- Created: 2026-04-01
-- Part of: Moat Competitive Analysis Phase 1 - Trust Foundation

BEGIN;

-- ============================================
-- Verification Jobs Table
-- Queue for verification pipeline jobs
-- ============================================
CREATE TABLE IF NOT EXISTS verification_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,

    -- Verification level
    level VARCHAR(20) NOT NULL DEFAULT 'basic', -- 'unverified', 'basic', 'standard', 'full'

    -- Job status
    status VARCHAR(20) NOT NULL DEFAULT 'queued', -- 'queued', 'running', 'completed', 'failed', 'cancelled'
    priority VARCHAR(20) NOT NULL DEFAULT 'normal', -- 'low', 'normal', 'high', 'urgent'

    -- Timing
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Result summary
    result_status VARCHAR(20), -- 'pass', 'fail', 'pending'
    result_data JSONB DEFAULT '{}',
    error TEXT,

    -- Requester info
    requested_by UUID REFERENCES users(id),

    -- Auto-verify flags
    is_auto_verify BOOLEAN NOT NULL DEFAULT false, -- true if triggered automatically on publish

    -- Retry tracking
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for verification_jobs
CREATE INDEX idx_verification_jobs_function_id ON verification_jobs(function_id);
CREATE INDEX idx_verification_jobs_status ON verification_jobs(status);
CREATE INDEX idx_verification_jobs_level ON verification_jobs(level);
CREATE INDEX idx_verification_jobs_priority ON verification_jobs(priority);
CREATE INDEX idx_verification_jobs_requested_at ON verification_jobs(requested_at);
CREATE INDEX idx_verification_jobs_is_auto_verify ON verification_jobs(is_auto_verify);
CREATE INDEX idx_verification_jobs_function_status ON verification_jobs(function_id, status);

-- ============================================
-- Verification Results Table
-- Stores detailed results of verification pipeline runs
-- ============================================
CREATE TABLE IF NOT EXISTS verification_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES verification_jobs(id) ON DELETE CASCADE,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,

    -- Verification level that was run
    level VARCHAR(20) NOT NULL,

    -- Overall result
    status VARCHAR(20) NOT NULL, -- 'pass', 'fail', 'pending'

    -- Stage results (JSONB for flexibility)
    stages_data JSONB DEFAULT '{}',

    -- Individual stage summaries
    malware_scan_passed BOOLEAN,
    malware_scan_risk_score NUMERIC(5,4),
    dre_passed BOOLEAN,
    dre_pass_rate NUMERIC(5,4),
    dre_is_deterministic BOOLEAN,
    fxcert_passed BOOLEAN,
    fxcert_valid_until TIMESTAMPTZ,
    manual_review_status VARCHAR(20), -- 'pending', 'approved', 'rejected'

    -- Execution metrics at time of verification
    total_executions INTEGER DEFAULT 0,
    success_rate NUMERIC(5,4) DEFAULT 0,
    avg_latency_ms INTEGER DEFAULT 0,
    error_rate NUMERIC(5,4) DEFAULT 0,

    -- Trust score impact
    trust_score_impact NUMERIC(5,2) DEFAULT 0,

    -- Timing
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,

    -- Error if failed
    error TEXT,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for verification_results
CREATE INDEX idx_verification_results_job_id ON verification_results(job_id);
CREATE INDEX idx_verification_results_function_id ON verification_results(function_id);
CREATE INDEX idx_verification_results_status ON verification_results(status);
CREATE INDEX idx_verification_results_level ON verification_results(level);
CREATE INDEX idx_verification_results_created_at ON verification_results(created_at DESC);
CREATE INDEX idx_verification_results_function_level ON verification_results(function_id, level);
CREATE INDEX idx_verification_results_valid_until ON verification_results(fxcert_valid_until) WHERE fxcert_valid_until IS NOT NULL;

-- ============================================
-- Add verification_level to registry_functions
-- Quick access to current verification level
-- ============================================
ALTER TABLE registry_functions
    ADD COLUMN IF NOT EXISTS verification_level VARCHAR(20) DEFAULT 'unverified',
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_verified_at TIMESTAMPTZ;

-- Index for quick trust-based sorting
CREATE INDEX IF NOT EXISTS idx_registry_functions_verification_level ON registry_functions(verification_level);

-- ============================================
-- Manual Review Queue Table
-- Human review queue for Level 3 verification
-- ============================================
CREATE TABLE IF NOT EXISTS manual_review_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    verification_job_id UUID REFERENCES verification_jobs(id),

    -- Review status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'in_review', 'approved', 'rejected', 'escalated'

    -- Priority and assignment
    priority VARCHAR(20) NOT NULL DEFAULT 'normal', -- 'low', 'normal', 'high', 'urgent'
    assigned_to UUID REFERENCES users(id),

    -- Review details
    review_type VARCHAR(50) NOT NULL, -- 'security', 'compliance', 'accuracy', 'quality'
    review_notes TEXT,
    review_comments TEXT,

    -- Decision
    decision_at TIMESTAMPTZ,
    decision_by UUID REFERENCES users(id),
    decision_reason TEXT,

    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Auto-review threshold
    auto_approve_if_no_response_days INTEGER DEFAULT 7
);

-- Indexes for manual_review_queue
CREATE INDEX idx_manual_review_queue_status ON manual_review_queue(status);
CREATE INDEX idx_manual_review_queue_priority ON manual_review_queue(priority);
CREATE INDEX idx_manual_review_queue_assigned_to ON manual_review_queue(assigned_to);
CREATE INDEX idx_manual_review_queue_function_id ON manual_review_queue(function_id);
CREATE INDEX idx_manual_review_queue_created_at ON manual_review_queue(created_at);
CREATE INDEX idx_manual_review_queue_due_at ON manual_review_queue(due_at) WHERE due_at IS NOT NULL;

-- ============================================
-- Verification Level Configuration
-- Define requirements for each verification level
-- ============================================
CREATE TABLE IF NOT EXISTS verification_level_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(20) NOT NULL UNIQUE,

    -- Stage requirements (which stages must pass)
    requires_malware_scan BOOLEAN NOT NULL DEFAULT true,
    requires_dre BOOLEAN NOT NULL DEFAULT false,
    requires_fxcert BOOLEAN NOT NULL DEFAULT false,
    requires_manual_review BOOLEAN NOT NULL DEFAULT false,

    -- Thresholds
    min_dre_pass_rate NUMERIC(5,4) DEFAULT 0.95,
    max_latency_ms INTEGER DEFAULT 5000,
    min_success_rate NUMERIC(5,4) DEFAULT 0.99,

    -- Auto-upgrade settings
    auto_upgrade_from_level VARCHAR(20), -- Level to auto-upgrade from
    auto_upgrade_after_days INTEGER, -- Days after which to auto-upgrade

    -- Trust score bonuses
    trust_bonus NUMERIC(5,2) DEFAULT 0,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Metadata
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id)
);

-- Insert default verification level configurations
INSERT INTO verification_level_config (level, requires_malware_scan, requires_dre, requires_fxcert, requires_manual_review, min_dre_pass_rate, max_latency_ms, min_success_rate, trust_bonus, description) VALUES
    ('unverified', false, false, false, false, 0, 0, 0, 0, 'No verification performed'),
    ('basic', true, false, false, false, 0, 0, 0, 5.00, 'Basic verification - malware scan only'),
    ('standard', true, true, true, false, 0.95, 5000, 0.99, 15.00, 'Standard verification - malware + DRE + FXCERT'),
    ('full', true, true, true, true, 0.98, 2000, 0.995, 25.00, 'Full verification - all checks + manual review')
ON CONFLICT (level) DO NOTHING;

-- ============================================
-- Verification Audit Log
-- Audit trail for all verification activities
-- ============================================
CREATE TABLE IF NOT EXISTS verification_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID REFERENCES registry_functions(id) ON DELETE SET NULL,
    verification_job_id UUID REFERENCES verification_jobs(id) ON DELETE SET NULL,
    verification_result_id UUID REFERENCES verification_results(id) ON DELETE SET NULL,

    -- Action details
    action VARCHAR(50) NOT NULL, -- 'job_created', 'job_started', 'job_completed', 'job_failed', 'level_changed', 'review_requested', 'review_decision'
    actor_type VARCHAR(20) NOT NULL, -- 'system', 'user', 'admin'
    actor_id UUID,
    actor_email VARCHAR(255),

    -- Change details
    old_value JSONB,
    new_value JSONB,

    -- Context
    ip_address INET,
    user_agent TEXT,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for verification_audit_log
CREATE INDEX idx_verification_audit_log_function_id ON verification_audit_log(function_id);
CREATE INDEX idx_verification_audit_log_job_id ON verification_audit_log(verification_job_id);
CREATE INDEX idx_verification_audit_log_action ON verification_audit_log(action);
CREATE INDEX idx_verification_audit_log_created_at ON verification_audit_log(created_at DESC);

-- ============================================
-- Periodic Verification Schedule
-- Track scheduled periodic re-verification
-- ============================================
CREATE TABLE IF NOT EXISTS verification_schedule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Schedule configuration
    verification_level VARCHAR(20) NOT NULL, -- Level to verify at
    frequency VARCHAR(20) NOT NULL DEFAULT 'monthly', -- 'weekly', 'monthly', 'quarterly'
    next_verification_at TIMESTAMPTZ NOT NULL,
    last_verification_at TIMESTAMPTZ,
    last_verification_result VARCHAR(20),

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_paused BOOLEAN NOT NULL DEFAULT false,
    pause_reason TEXT,

    -- Notifications
    notify_on_failure BOOLEAN NOT NULL DEFAULT true,
    notification_email VARCHAR(255),

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Indexes for verification_schedule
CREATE INDEX idx_verification_schedule_function_id ON verification_schedule(function_id);
CREATE INDEX idx_verification_schedule_next_verification ON verification_schedule(next_verification_at);
CREATE INDEX idx_verification_schedule_is_active ON verification_schedule(is_active);
CREATE INDEX idx_verification_schedule_paused ON verification_schedule(is_paused) WHERE is_paused = false;

-- ============================================
-- Update existing trust_history to reference verification levels
-- ============================================
ALTER TABLE trust_history
    ADD COLUMN IF NOT EXISTS verification_level VARCHAR(20) DEFAULT 'none';

COMMIT;
