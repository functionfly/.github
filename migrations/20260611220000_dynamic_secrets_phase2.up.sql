-- Migration: dynamic_secrets_phase2
-- Created at: 2026-06-11
-- Purpose: Phase 2 dynamic secret credentials with leasing.
--   2.1 On-demand credential generation (PostgreSQL first, then MySQL)
--   2.2 Lease model with renewal + revocation
--   2.3 Database target configuration (admin-managed connection settings)
--
-- The vault never stores plaintext target admin passwords: the
-- encrypted_admin_password column is encrypted with the tenant's
-- server-side envelope (see internal/crypto/server_encryption.go).

BEGIN;

-- =====================================================
-- 2.3 Database targets (PostgreSQL / MySQL)
-- =====================================================
CREATE TABLE IF NOT EXISTS dynamic_secret_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    db_type VARCHAR(20) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    admin_username VARCHAR(255) NOT NULL,
    encrypted_admin_password BYTEA NOT NULL,
    password_nonce BYTEA NOT NULL,
    password_key_version INTEGER NOT NULL DEFAULT 1,
    ssl_mode VARCHAR(20) NOT NULL DEFAULT 'require',
    allowed_roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_ttl_seconds INTEGER NOT NULL DEFAULT 3600,
    max_ttl_seconds INTEGER NOT NULL DEFAULT 86400,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    last_error TEXT,
    CONSTRAINT dynamic_secret_targets_db_type_check
        CHECK (db_type IN ('postgres', 'mysql')),
    CONSTRAINT dynamic_secret_targets_status_check
        CHECK (status IN ('active', 'disabled')),
    CONSTRAINT dynamic_secret_targets_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_dynamic_secret_targets_tenant
    ON dynamic_secret_targets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dynamic_secret_targets_active
    ON dynamic_secret_targets(tenant_id) WHERE status = 'active';

COMMENT ON TABLE dynamic_secret_targets IS 'Admin-managed database connection used to mint dynamic credentials';
COMMENT ON COLUMN dynamic_secret_targets.db_type IS 'postgres or mysql';
COMMENT ON COLUMN dynamic_secret_targets.allowed_roles IS 'JSON array of role names the dynamic user can be granted';
COMMENT ON COLUMN dynamic_secret_targets.encrypted_admin_password IS 'Server-side encrypted admin password used to mint credentials';

-- =====================================================
-- 2.1 Dynamic credential definitions (named, reusable)
-- =====================================================
CREATE TABLE IF NOT EXISTS dynamic_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES dynamic_secret_targets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    role_template VARCHAR(100),
    ttl_seconds INTEGER NOT NULL DEFAULT 3600,
    max_ttl_seconds INTEGER NOT NULL DEFAULT 86400,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT dynamic_credentials_status_check
        CHECK (status IN ('active', 'disabled')),
    CONSTRAINT dynamic_credentials_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_dynamic_credentials_tenant
    ON dynamic_credentials(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dynamic_credentials_target
    ON dynamic_credentials(target_id);

-- =====================================================
-- 2.2 Leases (one row per issuance / renewal)
-- =====================================================
CREATE TABLE IF NOT EXISTS dynamic_credential_leases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lease_id VARCHAR(64) NOT NULL UNIQUE,
    credential_id UUID NOT NULL REFERENCES dynamic_credentials(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES dynamic_secret_targets(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    db_username VARCHAR(128) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    renewed_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revocation_reason VARCHAR(100),
    last_used_at TIMESTAMPTZ,
    use_count INTEGER NOT NULL DEFAULT 0,
    issued_to UUID REFERENCES users(id) ON DELETE SET NULL,
    issued_ip VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dynamic_leases_tenant
    ON dynamic_credential_leases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dynamic_leases_credential
    ON dynamic_credential_leases(credential_id);
CREATE INDEX IF NOT EXISTS idx_dynamic_leases_active
    ON dynamic_credential_leases(expires_at)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dynamic_leases_username
    ON dynamic_credential_leases(target_id, db_username);

COMMENT ON TABLE dynamic_credential_leases IS 'Each issued dynamic credential has a lease row. Background worker drops users whose lease has expired.';

COMMIT;
