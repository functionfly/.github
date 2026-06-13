-- Migration: vault_enterprise_phase4
-- Created at: 2026-06-12
-- Purpose: Phase 4 enterprise features for the secrets vault
--   4.1 Role-based access control (roles, permissions, assignments)
--   4.2 Audit log compliance export (CEF + HMAC + SIEM webhooks)
--   4.3 Hierarchical secret namespaces
--   4.4 Cross-tenant secret sharing
--   4.5 SAML/SSO configuration for vault access

BEGIN;

-- =====================================================
-- 4.3 vault_namespaces — hierarchical paths
-- =====================================================
CREATE TABLE IF NOT EXISTS vault_namespaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    path VARCHAR(512) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES vault_namespaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vault_namespaces_path_check
        CHECK (path ~ '^[a-z0-9][a-z0-9/_-]*$'),
    CONSTRAINT vault_namespaces_tenant_path_unique UNIQUE (tenant_id, path)
);

CREATE INDEX IF NOT EXISTS idx_vault_namespaces_tenant ON vault_namespaces(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vault_namespaces_parent ON vault_namespaces(parent_id);

COMMENT ON TABLE vault_namespaces IS 'Hierarchical secret organization. Paths use /-separated lowercase segments (e.g. production/api-gateway).';

ALTER TABLE secrets_vault
    ADD COLUMN IF NOT EXISTS namespace VARCHAR(512) NOT NULL DEFAULT 'default',
    ADD COLUMN IF NOT EXISTS is_shared BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_secrets_vault_namespace
    ON secrets_vault(tenant_id, namespace) WHERE deleted_at IS NULL;

-- =====================================================
-- 4.1 vault_roles + vault_role_assignments
-- =====================================================
CREATE TABLE IF NOT EXISTS vault_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_builtin BOOLEAN NOT NULL DEFAULT false,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vault_roles_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_vault_roles_tenant ON vault_roles(tenant_id);

COMMENT ON TABLE vault_roles IS 'Named, JSONB-permissioned vault roles for RBAC. Built-in roles (admin, operator, reader) are seeded per-tenant on first use.';
COMMENT ON COLUMN vault_roles.permissions IS 'JSONB object mapping permission keys (e.g. secrets:read, secrets:create) to true or a string array of scope filters.';

CREATE TABLE IF NOT EXISTS vault_role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES vault_roles(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID, -- future: link to teams table
    scope VARCHAR(512) NOT NULL DEFAULT 'all',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vault_role_assignments_target_check
        CHECK (user_id IS NOT NULL OR team_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_vault_role_assignments_user
    ON vault_role_assignments(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_vault_role_assignments_role
    ON vault_role_assignments(role_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vault_role_assignments_unique_user
    ON vault_role_assignments(tenant_id, role_id, user_id)
    WHERE user_id IS NOT NULL;

-- =====================================================
-- 4.4 vault_shares — cross-tenant secret sharing
-- =====================================================
CREATE TABLE IF NOT EXISTS vault_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets_vault(id) ON DELETE CASCADE,
    source_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    granted_to_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    granted_by_user UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permissions VARCHAR(20) NOT NULL DEFAULT 'read',
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vault_shares_permissions_check
        CHECK (permissions IN ('read', 'read-write')),
    CONSTRAINT vault_shares_source_dest_check
        CHECK (source_tenant_id <> granted_to_tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_vault_shares_source ON vault_shares(source_tenant_id);
CREATE INDEX IF NOT EXISTS idx_vault_shares_dest ON vault_shares(granted_to_tenant_id);
CREATE INDEX IF NOT EXISTS idx_vault_shares_secret ON vault_shares(secret_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vault_shares_unique
    ON vault_shares(secret_id, granted_to_tenant_id) WHERE revoked_at IS NULL;

COMMENT ON TABLE vault_shares IS 'Cross-tenant secret shares. Grantee sees the secret under the shared/ namespace; can read decrypted value (client-side) but cannot rotate, delete, or re-share.';

-- =====================================================
-- 4.5 vault_sso_config — SAML/SSO configuration
-- =====================================================
CREATE TABLE IF NOT EXISTS vault_sso_config (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    saml_metadata_url TEXT,
    saml_entity_id VARCHAR(255),
    saml_sso_url TEXT,
    saml_slo_url TEXT,
    saml_x509_cert TEXT,
    default_role_id UUID REFERENCES vault_roles(id) ON DELETE SET NULL,
    jit_provisioning_enabled BOOLEAN NOT NULL DEFAULT true,
    attribute_role_mapping JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE vault_sso_config IS 'Per-tenant SAML SSO configuration for vault access. Map SAML attributes to vault_role_id for JIT-provisioned users.';

-- =====================================================
-- 4.2 SIEM webhook for audit streaming
-- =====================================================
CREATE TABLE IF NOT EXISTS vault_siem_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    url TEXT NOT NULL,
    secret_hmac BYTEA NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'json',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_delivery_at TIMESTAMPTZ,
    last_delivery_status INTEGER,
    last_delivery_error TEXT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vault_siem_webhooks_format_check
        CHECK (format IN ('json', 'cef'))
);

CREATE INDEX IF NOT EXISTS idx_vault_siem_webhooks_tenant
    ON vault_siem_webhooks(tenant_id) WHERE enabled = true;

COMMENT ON TABLE vault_siem_webhooks IS 'Push audit events to external SIEMs in real time. The secret_hmac is the shared secret used to compute X-Signature for the outbound request.';

-- =====================================================
-- 4.0 Backfill: existing secrets get namespace='default'
-- =====================================================
UPDATE secrets_vault
   SET namespace = 'default'
 WHERE namespace IS NULL OR namespace = '';

COMMIT;
