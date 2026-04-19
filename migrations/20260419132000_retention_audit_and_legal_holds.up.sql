-- Migration: Create retention audit log table for compliance tracking
-- Created: 2026-04-19
-- This tracks what data was deleted during retention policy enforcement

-- ============================================================
-- Retention Audit Log
-- ============================================================
-- Records compliance-level data deletion events for audit purposes
-- Required for SOX, PCI-DSS, and GDPR Article 17 (right to erasure) compliance

CREATE TABLE IF NOT EXISTS retention_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- What was deleted
    table_name VARCHAR(100) NOT NULL,
    retention_policy VARCHAR(50) NOT NULL, -- 'financial_7_year', 'detailed_90_day', 'gdpr_erasure', etc.
    
    -- When and how much
    cutoff_date TIMESTAMP WITH TIME ZONE NOT NULL,
    deletion_completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Impact metrics
    records_affected BIGINT NOT NULL DEFAULT 0,
    tenant_count INTEGER NOT NULL DEFAULT 0,
    financial_impact_cents BIGINT NOT NULL DEFAULT 0, -- Total value of deleted records
    
    -- Date range of deleted data
    oldest_record TIMESTAMP WITH TIME ZONE,
    newest_record TIMESTAMP WITH TIME ZONE,
    
    -- Execution details
    batch_size INTEGER DEFAULT 10000,
    total_batches INTEGER DEFAULT 1,
    execution_duration_ms INTEGER,
    
    -- Initiator (system or user)
    triggered_by VARCHAR(50) NOT NULL DEFAULT 'system', -- 'system', 'admin', 'gdpr_request'
    triggered_by_user_id UUID,
    
    -- Request context (for GDPR or legal holds)
    request_reference VARCHAR(255), -- External ticket ID or legal case reference
    legal_hold_exemptions INTEGER DEFAULT 0, -- Count of records skipped due to legal hold
    
    -- Verification
    verification_hash VARCHAR(64), -- Hash of summary data for tamper detection
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for retention audit log
CREATE INDEX IF NOT EXISTS idx_retention_audit_table ON retention_audit_log(table_name);
CREATE INDEX IF NOT EXISTS idx_retention_audit_policy ON retention_audit_log(retention_policy);
CREATE INDEX IF NOT EXISTS idx_retention_audit_cutoff ON retention_audit_log(cutoff_date);
CREATE INDEX IF NOT EXISTS idx_retention_audit_created ON retention_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_retention_audit_triggered_by ON retention_audit_log(triggered_by, triggered_by_user_id);

-- Partial index for pending deletions (verification not yet completed)
CREATE INDEX IF NOT EXISTS idx_retention_audit_pending 
    ON retention_audit_log(table_name, cutoff_date) 
    WHERE deletion_completed_at IS NULL;

-- ============================================================
-- Legal Hold Registry (prevents accidental deletion)
-- ============================================================

CREATE TABLE IF NOT EXISTS legal_holds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- What is under hold
    hold_name VARCHAR(255) NOT NULL,
    hold_description TEXT,
    
    -- Scope
    table_name VARCHAR(100), -- NULL = all tables
    tenant_id UUID,
    
    -- Date range covered by hold
    hold_date_from TIMESTAMP WITH TIME ZONE,
    hold_date_to TIMESTAMP WITH TIME ZONE,
    
    -- Who placed the hold
    requested_by VARCHAR(255) NOT NULL, -- Name or email
    requested_by_user_id UUID,
    legal_case_reference VARCHAR(255),
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'released', 'expired'
    
    -- Timeline
    placed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    released_at TIMESTAMP WITH TIME ZONE,
    released_by UUID,
    release_reason TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for legal holds
CREATE INDEX IF NOT EXISTS idx_legal_holds_status ON legal_holds(status);
CREATE INDEX IF NOT EXISTS idx_legal_holds_table ON legal_holds(table_name) WHERE table_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_legal_holds_tenant ON legal_holds(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_legal_holds_date_range ON legal_holds(hold_date_from, hold_date_to) WHERE status = 'active';

-- Active holds index (most common query - checking if deletion should be blocked)
CREATE INDEX IF NOT EXISTS idx_legal_holds_active 
    ON legal_holds(table_name, hold_date_from, hold_date_to) 
    WHERE status = 'active';

-- ============================================================
-- Function to check if a date range is under legal hold
-- ============================================================

CREATE OR REPLACE FUNCTION is_under_legal_hold(
    p_table_name VARCHAR(100),
    p_date_from TIMESTAMP WITH TIME ZONE,
    p_date_to TIMESTAMP WITH TIME ZONE
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM legal_holds
        WHERE status = 'active'
          AND (table_name IS NULL OR table_name = p_table_name)
          AND (expires_at IS NULL OR expires_at > NOW())
          AND (
              -- Overlapping date range check
              (hold_date_from IS NULL OR hold_date_from <= p_date_to)
              AND (hold_date_to IS NULL OR hold_date_to >= p_date_from)
          )
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Comments for documentation
-- ============================================================

COMMENT ON TABLE retention_audit_log IS 
    'Audit log of data retention policy enforcement. Records what data was deleted, when, and by whom. Required for compliance auditing.';
COMMENT ON TABLE legal_holds IS 
    'Registry of legal holds that prevent data deletion. Check this before any retention cleanup operation.';
COMMENT ON FUNCTION is_under_legal_hold IS 
    'Checks if a given date range for a table is under an active legal hold. Returns true if deletion should be blocked.';

-- ============================================================
-- Trigger to update legal_holds.updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION update_legal_hold_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_legal_holds_updated_at ON legal_holds;
CREATE TRIGGER trg_legal_holds_updated_at
    BEFORE UPDATE ON legal_holds
    FOR EACH ROW
    EXECUTE FUNCTION update_legal_hold_updated_at();
