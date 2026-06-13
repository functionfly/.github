-- Migration: vault_security_hardening_phase1
-- Created at: 2026-06-11
-- Purpose: Phase 1 production security hardening for the secrets vault
--   1.1 MFA enforcement for vault operations
--   1.2 IP allowlist / denylist on access tokens
--   1.3 Secret expiration enforcement
--   1.4 Break-glass emergency access + optional escrowed recovery
--   1.5 Hardened key derivation (Argon2id support)

BEGIN;

-- =====================================================
-- 1.1 vault_mfa_config — per-tenant vault MFA policy
-- =====================================================
CREATE TABLE IF NOT EXISTS vault_mfa_config (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    mfa_required BOOLEAN NOT NULL DEFAULT false,
    mfa_method VARCHAR(20) NOT NULL DEFAULT 'totp',
    enforce_for_tokens BOOLEAN NOT NULL DEFAULT false,
    enforce_for_api BOOLEAN NOT NULL DEFAULT false,
    mfa_session_ttl_seconds INTEGER NOT NULL DEFAULT 900,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE vault_mfa_config IS 'Per-tenant policy that gates vault operations on MFA verification';
COMMENT ON COLUMN vault_mfa_config.mfa_method IS 'totp, webauthn, or both';
COMMENT ON COLUMN vault_mfa_config.mfa_session_ttl_seconds IS 'How long a verified MFA assertion is honored (default 15m)';

-- =====================================================
-- 1.2 IP allowlist / denylist on secret_access_tokens
-- =====================================================
ALTER TABLE secret_access_tokens
    ADD COLUMN IF NOT EXISTS allowed_ips JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS denied_ips JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS ip_restriction_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_ip_restricted
    ON secret_access_tokens(tenant_id) WHERE ip_restriction_enabled = true;

-- =====================================================
-- 1.3 Secret expiration
-- =====================================================
ALTER TABLE secrets_vault
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS auto_expire BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS expire_after_days INTEGER,
    ADD COLUMN IF NOT EXISTS last_expiry_warning_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS expired_notified_at TIMESTAMPTZ;

-- Constraint: status must be one of these values
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'secrets_vault_status_check'
    ) THEN
        ALTER TABLE secrets_vault
            ADD CONSTRAINT secrets_vault_status_check
            CHECK (status IN ('active', 'expiring_soon', 'expired', 'revoked'));
    END IF;
END$$;

-- Index for background expiration sweep
CREATE INDEX IF NOT EXISTS idx_secrets_vault_expires_active
    ON secrets_vault(expires_at) WHERE status = 'active' AND expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_secrets_vault_status
    ON secrets_vault(tenant_id, status);

-- =====================================================
-- 1.4 Break-glass emergency access
-- =====================================================
CREATE TABLE IF NOT EXISTS break_glass_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    duration_minutes INTEGER NOT NULL DEFAULT 60,
    expires_at TIMESTAMPTZ NOT NULL,
    approved_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT break_glass_status_check CHECK (status IN ('pending', 'approved', 'denied', 'expired', 'revoked'))
);

CREATE INDEX IF NOT EXISTS idx_break_glass_tenant_status
    ON break_glass_requests(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_break_glass_expires
    ON break_glass_requests(expires_at) WHERE status = 'approved';

CREATE TABLE IF NOT EXISTS break_glass_config (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    max_duration_minutes INTEGER NOT NULL DEFAULT 60,
    required_approver_count INTEGER NOT NULL DEFAULT 1,
    approver_user_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 1.4b Optional escrowed recovery (enterprise tier)
CREATE TABLE IF NOT EXISTS vault_escrow_config (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    security_question_hashes JSONB NOT NULL DEFAULT '[]'::jsonb,
    kdf_salt BYTEA NOT NULL,
    kdf_method VARCHAR(20) NOT NULL DEFAULT 'argon2id',
    kdf_params JSONB NOT NULL DEFAULT '{}'::jsonb,
    encrypted_recovery_blob BYTEA NOT NULL,
    blob_iv BYTEA NOT NULL,
    blob_auth_tag BYTEA NOT NULL,
    blob_key_version INTEGER NOT NULL DEFAULT 2,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_recovered_at TIMESTAMPTZ
);

COMMENT ON TABLE vault_escrow_config IS 'Optional enterprise escrow for lost-passphrase recovery. Stores only an encrypted recovery blob, never the master passphrase.';

-- =====================================================
-- 1.5 key_derivation metadata on secrets
-- =====================================================
-- The existing key_version column is reused (1=PBKDF2 legacy, 2=Argon2id).
-- Add explicit metadata column for KDF parameters so future migration
-- can pick the right algorithm without guessing.
ALTER TABLE secrets_vault
    ADD COLUMN IF NOT EXISTS kdf_method VARCHAR(20) NOT NULL DEFAULT 'pbkdf2-sha256',
    ADD COLUMN IF NOT EXISTS kdf_params JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE secrets_vault
   SET kdf_method = 'pbkdf2-sha256',
       kdf_params = '{"iterations":100000,"key_length":32}'::jsonb
 WHERE kdf_method IS NULL OR kdf_method = '';

COMMIT;
