-- Migration: Add billing operational readiness tables
-- Created: 2026-04-19

-- ============================================================
-- 1. Stored Webhook Payloads (for replay capability)
-- ============================================================
CREATE TABLE IF NOT EXISTS stored_webhook_payloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    signature VARCHAR(255),
    processing_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processed, failed, replayed
    processed_at TIMESTAMP WITH TIME ZONE,
    replayed_at TIMESTAMP WITH TIME ZONE,
    replayed_by UUID,
    replay_reason VARCHAR(500),
    processing_error VARCHAR(1000),
    attempts INT DEFAULT 0,
    webhook_secret_hash VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    CONSTRAINT chk_processing_status CHECK (processing_status IN ('pending', 'processed', 'failed', 'replayed'))
);

-- Indexes for stored_webhook_payloads
CREATE INDEX idx_webhook_event_id ON stored_webhook_payloads(stripe_event_id);
CREATE INDEX idx_webhook_event_type ON stored_webhook_payloads(event_type);
CREATE INDEX idx_webhook_status ON stored_webhook_payloads(processing_status);
CREATE INDEX idx_webhook_expires ON stored_webhook_payloads(expires_at);
CREATE INDEX idx_webhook_created_at ON stored_webhook_payloads(created_at DESC);

-- Partial index for unprocessed webhooks (useful for operational monitoring)
CREATE INDEX idx_webhook_pending ON stored_webhook_payloads(processing_status, created_at) 
    WHERE processing_status IN ('pending', 'failed');

-- ============================================================
-- 2. Webhook Replay Requests (audit trail)
-- ============================================================
CREATE TABLE IF NOT EXISTS webhook_replay_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_payload_id UUID NOT NULL REFERENCES stored_webhook_payloads(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    reason VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, completed, failed
    result_message VARCHAR(1000),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_replay_status CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_webhook_replay_payload ON webhook_replay_requests(webhook_payload_id);
CREATE INDEX idx_webhook_replay_requested_by ON webhook_replay_requests(requested_by);
CREATE INDEX idx_webhook_replay_status ON webhook_replay_requests(status);

-- ============================================================
-- 3. Tax Exemption Certificates (US entities)
-- ============================================================
CREATE TABLE IF NOT EXISTS tax_exemption_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    
    -- Certificate details
    certificate_number VARCHAR(100) NOT NULL,
    state VARCHAR(2) NOT NULL, -- US state code
    exemption_type VARCHAR(50) NOT NULL, -- resale, nonprofit, government, agricultural, etc.
    exemption_reason VARCHAR(500),
    
    -- File storage
    file_url VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash VARCHAR(64) NOT NULL, -- SHA-256 hash for integrity
    
    -- Validity period
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE,
    
    -- Review workflow
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, expired
    reviewed_by UUID,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes VARCHAR(1000),
    rejection_reason VARCHAR(500),
    
    -- Stripe integration
    stripe_exemption_id VARCHAR(255),
    applied_to_stripe_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_cert_status CHECK (status IN ('pending', 'approved', 'rejected', 'expired'))
);

-- Indexes for tax_exemption_certificates
CREATE INDEX idx_tax_cert_tenant ON tax_exemption_certificates(tenant_id);
CREATE INDEX idx_tax_cert_user ON tax_exemption_certificates(user_id);
CREATE INDEX idx_tax_cert_status ON tax_exemption_certificates(status);
CREATE INDEX idx_tax_cert_state ON tax_exemption_certificates(state);
CREATE INDEX idx_tax_cert_number ON tax_exemption_certificates(certificate_number);
CREATE INDEX idx_tax_cert_valid_dates ON tax_exemption_certificates(valid_from, valid_until);

-- Partial index for pending certificates (for admin review queue)
CREATE INDEX idx_tax_cert_pending ON tax_exemption_certificates(status, created_at) 
    WHERE status = 'pending';

-- ============================================================
-- 4. EU VAT Validations (VIES integration)
-- ============================================================
CREATE TABLE IF NOT EXISTS eu_vat_validations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    
    -- VAT ID details
    vat_id VARCHAR(20) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    
    -- VIES API response
    is_valid BOOLEAN NOT NULL,
    request_date TIMESTAMP WITH TIME ZONE NOT NULL,
    validation_source VARCHAR(20) NOT NULL DEFAULT 'vies', -- vies, fallback
    
    -- VIES response details (stored for audit)
    vies_request_id VARCHAR(100),
    vies_response_code VARCHAR(50),
    vies_trader_name VARCHAR(255),
    vies_trader_address VARCHAR(500),
    
    -- Error handling
    error_code VARCHAR(50),
    error_message VARCHAR(500),
    
    -- Retry logic for VIES unavailability
    retry_count INT DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, valid, invalid, error, timeout
    
    -- Applied to tenant tax settings
    applied_to_settings BOOLEAN DEFAULT FALSE,
    applied_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_vat_status CHECK (status IN ('pending', 'valid', 'invalid', 'error', 'timeout'))
);

-- Indexes for eu_vat_validations
CREATE INDEX idx_vat_tenant ON eu_vat_validations(tenant_id);
CREATE INDEX idx_vat_user ON eu_vat_validations(user_id);
CREATE INDEX idx_vat_id ON eu_vat_validations(vat_id);
CREATE INDEX idx_vat_status ON eu_vat_validations(status);
CREATE INDEX idx_vat_valid ON eu_vat_validations(is_valid, status);
CREATE INDEX idx_vat_retry ON eu_vat_validations(next_retry_at) WHERE next_retry_at IS NOT NULL;

-- ============================================================
-- 5. Automated cleanup for 30-day webhook retention
-- ============================================================

-- Create a function to clean up expired webhook payloads
CREATE OR REPLACE FUNCTION cleanup_expired_webhook_payloads()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM stored_webhook_payloads 
    WHERE expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE stored_webhook_payloads IS 'Raw webhook payloads stored for 30 days for replay capability';
COMMENT ON TABLE webhook_replay_requests IS 'Audit log of manual webhook replay requests';
COMMENT ON TABLE tax_exemption_certificates IS 'US tax exemption certificates uploaded by customers';
COMMENT ON TABLE eu_vat_validations IS 'EU VAT ID validation results from VIES API';
