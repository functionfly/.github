-- Migration: Create MFA enrollment tracking table
-- Tracks MFA setup lifecycle: initial setup -> verification -> enabled/disabled
-- This provides audit trail and enables MFA method tracking (TOTP, WebAuthn, etc.)

CREATE TABLE IF NOT EXISTS mfa_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL DEFAULT 'totp', -- totp, webauthn, sms, email
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, active, disabled, expired, revoked
    secret_encrypted TEXT, -- AES-256-GCM encrypted TOTP secret (NULL for WebAuthn)
    secret_iv TEXT,        -- Initialization vector for secret encryption
    secret_tag TEXT,       -- Authentication tag for secret encryption
    backup_codes_encrypted TEXT, -- Encrypted backup codes (JSON array)
    backup_codes_iv TEXT,
    backup_codes_tag TEXT,
    metadata JSONB DEFAULT '{}',
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ, -- When MFA was successfully verified
    disabled_at TIMESTAMPTZ, -- When MFA was disabled
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_method UNIQUE (user_id, method) -- One enrollment per method per user
);

CREATE INDEX IF NOT EXISTS idx_mfa_enrollments_user_id ON mfa_enrollments(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_enrollments_tenant_id ON mfa_enrollments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mfa_enrollments_status ON mfa_enrollments(status);
CREATE INDEX IF NOT EXISTS idx_mfa_enrollments_enrolled_at ON mfa_enrollments(enrolled_at);

-- Table to track individual backup codes (for one-time use tracking)
CREATE TABLE IF NOT EXISTS mfa_backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES mfa_enrollments(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL, -- bcrypt hash of the code
    used_at TIMESTAMPTZ,           -- NULL = unused
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_enrollment_code UNIQUE (enrollment_id, code_hash)
);

CREATE INDEX IF NOT EXISTS idx_mfa_backup_codes_enrollment_id ON mfa_backup_codes(enrollment_id);
CREATE INDEX IF NOT EXISTS idx_mfa_backup_codes_used_at ON mfa_backup_codes(used_at) WHERE used_at IS NULL;

-- Trigger to update updated_at on mfa_enrollments
CREATE OR REPLACE FUNCTION update_mfa_enrollment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_mfa_enrollments_updated_at ON mfa_enrollments;
CREATE TRIGGER trg_mfa_enrollments_updated_at
    BEFORE UPDATE ON mfa_enrollments
    FOR EACH ROW EXECUTE FUNCTION update_mfa_enrollment_updated_at();

COMMENT ON TABLE mfa_enrollments IS 'Tracks MFA enrollment lifecycle per user per method';
COMMENT ON TABLE mfa_backup_codes IS 'Individual backup codes with one-time use tracking';
COMMENT ON COLUMN mfa_enrollments.method IS 'MFA method: totp, webauthn, sms, email';
COMMENT ON COLUMN mfa_enrollments.status IS 'Enrollment status: pending, active, disabled, expired, revoked';
COMMENT ON COLUMN mfa_enrollments.secret_encrypted IS 'Zero-knowledge encrypted secret - server never sees plaintext';