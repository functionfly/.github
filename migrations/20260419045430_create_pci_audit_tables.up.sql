-- PCI DSS Audit Logging Tables
-- These tables support PCI DSS compliance requirements:
-- - Req 10.3: Record audit trail entries for all system components
-- - Req 3.6: Key management with proper audit trails
-- - Req 8.2: Authentication and access control audit trails

-- ============================================
-- Main PCI Audit Events Table
-- Immutable audit trail for all cardholder data environment (CDE) access
-- ============================================
CREATE TABLE IF NOT EXISTS pci_audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',

    -- Actor Information
    actor_user_id UUID NULL,
    actor_email VARCHAR(255) NULL,
    actor_role VARCHAR(50) NULL,
    actor_ip INET NULL,
    actor_user_agent TEXT NULL,
    session_id VARCHAR(255) NULL,

    -- Resource Information
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NULL,
    tenant_id UUID NULL,

    -- Cardholder Data Context (PCI-safe: only last 4 digits, never full PAN or CVV)
    card_last_four VARCHAR(4) NULL,
    card_brand VARCHAR(50) NULL,
    card_expiry_month INT NULL,
    card_expiry_year INT NULL,
    token_id VARCHAR(255) NULL,

    -- Encryption Key Context
    key_id UUID NULL,
    key_algorithm VARCHAR(50) NULL,
    key_operation VARCHAR(20) NULL,

    -- Event Details
    description TEXT NOT NULL,
    request_id VARCHAR(255) NULL,
    transaction_id VARCHAR(255) NULL,
    stripe_event_id VARCHAR(255) NULL,

    -- Security Context
    auth_method VARCHAR(50) NULL,
    mfa_used BOOLEAN NOT NULL DEFAULT FALSE,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    failure_reason TEXT NULL,

    -- Compliance Metadata (JSON for flexibility)
    compliance_data JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',

    -- Retention and Tamper Protection
    retention_until TIMESTAMP NOT NULL,
    tamper_hash VARCHAR(64) NULL,
    chain_hash VARCHAR(64) NULL,

    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- PCI Audit Event Indexes
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_event_type ON pci_audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_severity ON pci_audit_events(severity);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_actor_user_id ON pci_audit_events(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_tenant_id ON pci_audit_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_resource_type ON pci_audit_events(resource_type);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_resource_id ON pci_audit_events(resource_id);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_transaction_id ON pci_audit_events(transaction_id);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_created_at ON pci_audit_events(created_at);
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_retention_until ON pci_audit_events(retention_until);

-- Composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_pci_audit_events_tenant_created ON pci_audit_events(tenant_id, created_at DESC);

-- ============================================
-- Encryption Key Tracking
-- For PCI DSS Req 3.6: Key management with proper lifecycle tracking
-- ============================================
CREATE TABLE IF NOT EXISTS pci_encryption_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_type VARCHAR(50) NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    key_fingerprint VARCHAR(128) NOT NULL,
    key_status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Key Lifecycle
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    created_by UUID NOT NULL,
    activated_at TIMESTAMP NULL,
    rotated_at TIMESTAMP NULL,
    rotated_from_key UUID NULL,
    retired_at TIMESTAMP NULL,
    retired_reason TEXT NULL,
    destroyed_at TIMESTAMP NULL,

    -- Key Splitting/Sharing (for custodian requirements)
    has_key_shares BOOLEAN NOT NULL DEFAULT FALSE,
    key_shares_total INT NOT NULL DEFAULT 0,
    key_shares_required INT NOT NULL DEFAULT 0,

    -- HSM/External KMS
    is_hsm_backed BOOLEAN NOT NULL DEFAULT FALSE,
    hsm_key_id VARCHAR(255) NULL,
    kms_provider VARCHAR(50) NULL,

    -- Rotation Schedule
    rotation_due_at TIMESTAMP NULL,
    rotation_interval_days INT NOT NULL DEFAULT 365,

    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Encryption Key Indexes
CREATE INDEX IF NOT EXISTS idx_pci_encryption_keys_key_status ON pci_encryption_keys(key_status);
CREATE INDEX IF NOT EXISTS idx_pci_encryption_keys_key_fingerprint ON pci_encryption_keys(key_fingerprint);
CREATE INDEX IF NOT EXISTS idx_pci_encryption_keys_created_by ON pci_encryption_keys(created_by);
CREATE INDEX IF NOT EXISTS idx_pci_encryption_keys_rotated_from ON pci_encryption_keys(rotated_from_key);
CREATE INDEX IF NOT EXISTS idx_pci_encryption_keys_rotation_due ON pci_encryption_keys(rotation_due_at);

-- ============================================
-- Key Access Logging
-- For PCI DSS: Track all key custodian access
-- ============================================
CREATE TABLE IF NOT EXISTS pci_key_access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id UUID NOT NULL,
    user_id UUID NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    access_type VARCHAR(50) NOT NULL,
    access_reason TEXT NOT NULL,

    -- Context
    ip_address INET NULL,
    user_agent TEXT NULL,
    session_id VARCHAR(255) NULL,
    mfa_verified BOOLEAN NOT NULL DEFAULT FALSE,

    -- Approval for sensitive operations
    approved_by UUID NULL,
    approval_ticket VARCHAR(255) NULL,

    success BOOLEAN NOT NULL DEFAULT TRUE,
    failure_reason TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Key Access Log Indexes
CREATE INDEX IF NOT EXISTS idx_pci_key_access_logs_key_id ON pci_key_access_logs(key_id);
CREATE INDEX IF NOT EXISTS idx_pci_key_access_logs_user_id ON pci_key_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_pci_key_access_logs_access_type ON pci_key_access_logs(access_type);
CREATE INDEX IF NOT EXISTS idx_pci_key_access_logs_created_at ON pci_key_access_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_pci_key_access_logs_key_created ON pci_key_access_logs(key_id, created_at DESC);

-- ============================================
-- Cardholder Data Access Log
-- Detailed tracking of all CDE access
-- ============================================
CREATE TABLE IF NOT EXISTS pci_cardholder_data_access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_type VARCHAR(50) NOT NULL,
    data_type VARCHAR(50) NOT NULL,

    -- Actor (user or system)
    user_id UUID NULL,
    system_process VARCHAR(255) NULL,
    api_key_id UUID NULL,

    -- Context
    tenant_id UUID NULL,
    payment_method_id UUID NULL,
    transaction_id VARCHAR(255) NULL,

    -- Data Identification (safely logged)
    card_last_four VARCHAR(4) NULL,
    token_reference VARCHAR(255) NULL,

    -- Access Context
    ip_address INET NULL,
    request_id VARCHAR(255) NULL,
    purpose TEXT NOT NULL,

    -- CDE Tracking
    cde_section VARCHAR(100) NOT NULL,
    data_flow_step VARCHAR(100) NOT NULL,

    success BOOLEAN NOT NULL DEFAULT TRUE,
    failure_reason TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Cardholder Data Access Log Indexes
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_access_type ON pci_cardholder_data_access_logs(access_type);
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_user_id ON pci_cardholder_data_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_tenant_id ON pci_cardholder_data_access_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_payment_method ON pci_cardholder_data_access_logs(payment_method_id);
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_cde_section ON pci_cardholder_data_access_logs(cde_section);
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_created_at ON pci_cardholder_data_access_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_pci_cde_access_logs_tenant_created ON pci_cardholder_data_access_logs(tenant_id, created_at DESC);

-- ============================================
-- Environment Controls (Network segmentation, access lists)
-- ============================================
CREATE TABLE IF NOT EXISTS pci_environment_controls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    control_type VARCHAR(50) NOT NULL,
    control_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,

    -- Network controls
    source_ip_range CIDR NULL,
    destination_ip_range CIDR NULL,
    port_range VARCHAR(50) NULL,
    protocol VARCHAR(20) NULL,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    approved_by UUID NOT NULL,
    approved_at TIMESTAMP NOT NULL DEFAULT now(),

    -- Review cycle
    last_reviewed_at TIMESTAMP NULL,
    reviewed_by UUID NULL,
    next_review_at TIMESTAMP NULL,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Environment Control Indexes
CREATE INDEX IF NOT EXISTS idx_pci_env_controls_type ON pci_environment_controls(control_type);
CREATE INDEX IF NOT EXISTS idx_pci_env_controls_active ON pci_environment_controls(is_active);
CREATE INDEX IF NOT EXISTS idx_pci_env_controls_next_review ON pci_environment_controls(next_review_at);

-- ============================================
-- Comments explaining table purposes
-- ============================================
COMMENT ON TABLE pci_audit_events IS 'Immutable audit trail for PCI DSS compliance. Records all access to cardholder data environment. Retention: 1 year minimum, 3 years for critical events';
COMMENT ON TABLE pci_encryption_keys IS 'Encryption key lifecycle tracking for PCI DSS Req 3.6. Never stores actual keys, only fingerprints and metadata';
COMMENT ON TABLE pci_key_access_logs IS 'Detailed logging of all encryption key access by custodians for PCI compliance';
COMMENT ON TABLE pci_cardholder_data_access_logs IS 'Detailed tracking of all cardholder data access within the CDE for PCI DSS Req 10';
COMMENT ON TABLE pci_environment_controls IS 'Network segmentation and access control documentation for PCI DSS Req 1 and 2';

-- ============================================
-- Enable Row Level Security (RLS) for PCI tables
-- ============================================
ALTER TABLE pci_audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE pci_encryption_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE pci_key_access_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE pci_cardholder_data_access_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE pci_environment_controls ENABLE ROW LEVEL SECURITY;

-- ============================================
-- Create RLS Policies
-- ============================================

-- PCI Audit Events: Admin-only access for audit reviews
CREATE POLICY pci_audit_events_admin_select ON pci_audit_events
    FOR SELECT USING (current_user = 'admin' OR current_user LIKE '%admin%');

-- Encryption Keys: Admin and key custodians
CREATE POLICY pci_encryption_keys_admin_all ON pci_encryption_keys
    FOR ALL USING (current_user = 'admin' OR current_user LIKE '%admin%');

-- Key Access Logs: Admin-only (immutable)
CREATE POLICY pci_key_access_logs_admin_select ON pci_key_access_logs
    FOR SELECT USING (current_user = 'admin' OR current_user LIKE '%admin%');

-- CDE Access Logs: Admin and tenant-scoped access
CREATE POLICY pci_cde_access_logs_admin_select ON pci_cardholder_data_access_logs
    FOR SELECT USING (current_user = 'admin' OR current_user LIKE '%admin%');

CREATE POLICY pci_cde_access_logs_tenant_select ON pci_cardholder_data_access_logs
    FOR SELECT USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Environment Controls: Admin-only
CREATE POLICY pci_env_controls_admin_all ON pci_environment_controls
    FOR ALL USING (current_user = 'admin' OR current_user LIKE '%admin%');

-- ============================================
-- Create helper function for retention management
-- ============================================
CREATE OR REPLACE FUNCTION pci_purge_expired_audit_events()
RETURNS INTEGER AS $$
DECLARE
    purged_count INTEGER;
BEGIN
    DELETE FROM pci_audit_events
    WHERE retention_until < NOW()
    RETURNING COUNT(*) INTO purged_count;

    RETURN purged_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION pci_purge_expired_audit_events() IS 'Removes PCI audit events past their retention period. Run as scheduled job.';
