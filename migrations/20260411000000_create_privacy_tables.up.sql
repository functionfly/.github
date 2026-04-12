-- Privacy and Compliance Tables
-- Implements GDPR compliance features, PII anonymization, and data export/deletion

-- Privacy Settings Table (per user/tenant)
CREATE TABLE IF NOT EXISTS privacy_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership (user or tenant level)
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    -- Privacy level configuration
    privacy_level VARCHAR(20) NOT NULL DEFAULT 'standard',
    anonymize_ip BOOLEAN NOT NULL DEFAULT FALSE,
    anonymize_user_agent BOOLEAN NOT NULL DEFAULT FALSE,
    log_geo_data BOOLEAN NOT NULL DEFAULT TRUE,
    log_embed_origin BOOLEAN NOT NULL DEFAULT TRUE,
    store_input_output BOOLEAN NOT NULL DEFAULT TRUE,
    retention_days INT NOT NULL DEFAULT 90,
    gdpr_mode BOOLEAN NOT NULL DEFAULT FALSE,
    auto_delete_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    consent_required BOOLEAN NOT NULL DEFAULT FALSE,

    -- PII masking types
    ip_mask_type VARCHAR(20) NOT NULL DEFAULT 'none',
    user_agent_mask_type VARCHAR(20) NOT NULL DEFAULT 'none',

    -- Consent tracking
    consent_given_at TIMESTAMP WITH TIME ZONE,
    consent_version VARCHAR(50),

    -- Lifecycle tracking
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Ensure either tenant_id or user_id is set, but not both
    CONSTRAINT chk_privacy_settings_owner CHECK (
        (tenant_id IS NOT NULL AND user_id IS NULL) OR
        (tenant_id IS NULL AND user_id IS NOT NULL)
    ),

    -- Unique constraint for user settings
    CONSTRAINT unique_user_privacy UNIQUE (user_id)
);

-- Global Privacy Settings Table (system-wide defaults)
CREATE TABLE IF NOT EXISTS global_privacy_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Default privacy configuration
    default_privacy_level VARCHAR(20) NOT NULL DEFAULT 'standard',
    default_ip_mask_type VARCHAR(20) NOT NULL DEFAULT 'none',
    default_user_agent_mask_type VARCHAR(20) NOT NULL DEFAULT 'none',
    default_retention_days INT NOT NULL DEFAULT 90,

    -- Compliance modes
    gdpr_mode_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ccpa_mode_enabled BOOLEAN NOT NULL DEFAULT FALSE,

    -- Automated privacy features
    auto_anonymize_after_days INT DEFAULT 0, -- 0 = disabled
    require_consent BOOLEAN NOT NULL DEFAULT FALSE,
    pii_scanning_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    input_output_redaction BOOLEAN NOT NULL DEFAULT FALSE,

    -- Lifecycle tracking
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Single-row constraint (only one global settings row)
    is_active BOOLEAN NOT NULL DEFAULT TRUE UNIQUE
);

-- Privacy Consent Records Table
CREATE TABLE IF NOT EXISTS privacy_consent_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type VARCHAR(50) NOT NULL, -- 'execution_logging', 'analytics', 'marketing'
    consent_given BOOLEAN NOT NULL,
    consent_version VARCHAR(50) NOT NULL,
    consent_text TEXT,

    -- Audit hashes (not the actual values)
    ip_hash VARCHAR(64),
    user_agent_hash VARCHAR(64),

    given_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    withdrawn_at TIMESTAMP WITH TIME ZONE,
    withdrawn_reason TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Data Export Requests Table (GDPR Article 20)
CREATE TABLE IF NOT EXISTS data_export_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    request_type VARCHAR(20) NOT NULL DEFAULT 'full', -- 'full', 'executions', 'profile', 'audit'

    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Download info (for completed exports)
    download_url TEXT,
    download_token VARCHAR(255),
    file_size BIGINT,
    record_count BIGINT,

    error_message TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Data Deletion Requests Table (GDPR Article 17 - Right to erasure)
CREATE TABLE IF NOT EXISTS data_deletion_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed', 'partial'
    request_type VARCHAR(20) NOT NULL DEFAULT 'full', -- 'full', 'executions', 'audit_logs'

    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    records_deleted BIGINT DEFAULT 0,
    records_anonymized BIGINT DEFAULT 0,
    error_message TEXT,

    -- Verification hash to prove deletion
    verification_hash VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Privacy Audit Log Table
CREATE TABLE IF NOT EXISTS privacy_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    action VARCHAR(50) NOT NULL, -- 'export_requested', 'export_completed', 'deletion_requested', 'deletion_completed', 'consent_given', 'consent_withdrawn', 'settings_updated'
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_id UUID, -- References export or deletion request

    ip_address VARCHAR(45),
    user_agent TEXT,
    details TEXT,

    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance

-- Privacy settings indexes
CREATE INDEX IF NOT EXISTS idx_privacy_settings_user_id ON privacy_settings (user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_settings_tenant_id ON privacy_settings (tenant_id);
CREATE INDEX IF NOT EXISTS idx_privacy_settings_privacy_level ON privacy_settings (privacy_level);

-- Global privacy settings indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_global_privacy_settings_single_active
ON global_privacy_settings (is_active) WHERE is_active = TRUE;

-- Consent records indexes
CREATE INDEX IF NOT EXISTS idx_privacy_consent_user_id ON privacy_consent_records (user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_consent_type ON privacy_consent_records (consent_type);
CREATE INDEX IF NOT EXISTS idx_privacy_consent_active ON privacy_consent_records (user_id, consent_type) WHERE withdrawn_at IS NULL;

-- Export requests indexes
CREATE INDEX IF NOT EXISTS idx_data_export_user_id ON data_export_requests (user_id);
CREATE INDEX IF NOT EXISTS idx_data_export_status ON data_export_requests (status);
CREATE INDEX IF NOT EXISTS idx_data_export_created ON data_export_requests (created_at DESC);

-- Deletion requests indexes
CREATE INDEX IF NOT EXISTS idx_data_deletion_user_id ON data_deletion_requests (user_id);
CREATE INDEX IF NOT EXISTS idx_data_deletion_status ON data_deletion_requests (status);
CREATE INDEX IF NOT EXISTS idx_data_deletion_created ON data_deletion_requests (created_at DESC);

-- Privacy audit log indexes
CREATE INDEX IF NOT EXISTS idx_privacy_audit_user_id ON privacy_audit_log (user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_audit_action ON privacy_audit_log (action);
CREATE INDEX IF NOT EXISTS idx_privacy_audit_created ON privacy_audit_log (created_at DESC);

-- Insert default global privacy settings
INSERT INTO global_privacy_settings (
    default_privacy_level,
    default_ip_mask_type,
    default_user_agent_mask_type,
    default_retention_days,
    gdpr_mode_enabled,
    ccpa_mode_enabled,
    auto_anonymize_after_days,
    require_consent,
    pii_scanning_enabled,
    input_output_redaction,
    is_active
) VALUES (
    'standard', -- default_privacy_level
    'none',     -- default_ip_mask_type
    'none',     -- default_user_agent_mask_type
    90,         -- default_retention_days
    FALSE,      -- gdpr_mode_enabled
    FALSE,      -- ccpa_mode_enabled
    0,          -- auto_anonymize_after_days (disabled)
    FALSE,      -- require_consent
    FALSE,      -- pii_scanning_enabled
    FALSE,      -- input_output_redaction
    TRUE        -- is_active
)
ON CONFLICT (is_active) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE privacy_settings IS 'Per-user or per-tenant privacy configuration for PII handling and GDPR compliance';
COMMENT ON TABLE global_privacy_settings IS 'System-wide default privacy settings';
COMMENT ON TABLE privacy_consent_records IS 'Records of user consent for data processing under GDPR Article 7';
COMMENT ON TABLE data_export_requests IS 'GDPR Article 20 data portability export requests';
COMMENT ON TABLE data_deletion_requests IS 'GDPR Article 17 right to erasure (right to be forgotten) requests';
COMMENT ON TABLE privacy_audit_log IS 'Audit trail for privacy-related actions';
