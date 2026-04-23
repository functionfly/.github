-- Migration: Create deploy_keys table for SSH deploy key management
-- Description: Adds SSH public key infrastructure for function deployments

CREATE TABLE IF NOT EXISTS deploy_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_deploy_keys_tenant ON deploy_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_deploy_keys_fingerprint ON deploy_keys(fingerprint);
CREATE INDEX IF NOT EXISTS idx_deploy_keys_created_by ON deploy_keys(created_by);

COMMENT ON TABLE deploy_keys IS 'SSH public keys for authenticating function deployments';
COMMENT ON COLUMN deploy_keys.public_key IS 'OpenSSH format: ssh-ed25519 AAAA..., ssh-rsa AAAA..., ecdsa-sha2-nistp256 AAAA...';
COMMENT ON COLUMN deploy_keys.fingerprint IS 'SHA256:base64 fingerprint of the raw key, used for authentication lookup';
