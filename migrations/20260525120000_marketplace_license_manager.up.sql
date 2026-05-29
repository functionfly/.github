-- Marketplace License Manager
-- Function license policies (SPDX/commercial) and issued license grants

BEGIN;

CREATE TABLE IF NOT EXISTS marketplace_function_license_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id VARCHAR(255) NOT NULL,
    spdx_license VARCHAR(50) NOT NULL DEFAULT 'mit',
    custom_license_text TEXT,
    commercial_type VARCHAR(20) NOT NULL DEFAULT 'open',
    max_activations_default INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, function_id),
    CONSTRAINT marketplace_function_license_policies_commercial_type_check
        CHECK (commercial_type IN ('open', 'restricted', 'commercial')),
    CONSTRAINT marketplace_function_license_policies_spdx_check
        CHECK (spdx_license IN ('mit', 'apache', 'gpl', 'proprietary', 'custom'))
);

CREATE INDEX IF NOT EXISTS idx_marketplace_license_policies_tenant
    ON marketplace_function_license_policies (tenant_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_license_policies_function
    ON marketplace_function_license_policies (function_id);

CREATE TABLE IF NOT EXISTS marketplace_license_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id VARCHAR(255) NOT NULL,
    function_name VARCHAR(255) NOT NULL DEFAULT '',
    license_key_hash VARCHAR(64) NOT NULL,
    license_key_prefix VARCHAR(16) NOT NULL,
    license_type VARCHAR(20) NOT NULL DEFAULT 'commercial',
    purchaser_tenant_id UUID,
    purchaser_user_id UUID,
    purchaser_name VARCHAR(255) NOT NULL DEFAULT '',
    issued_by_user_id UUID NOT NULL,
    expires_at TIMESTAMPTZ,
    max_activations INTEGER,
    activation_count INTEGER NOT NULL DEFAULT 0,
    revoked_at TIMESTAMPTZ,
    revoked_by_user_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT marketplace_license_grants_type_check
        CHECK (license_type IN ('open', 'restricted', 'commercial'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_marketplace_license_grants_key_hash
    ON marketplace_license_grants (license_key_hash);
CREATE INDEX IF NOT EXISTS idx_marketplace_license_grants_tenant
    ON marketplace_license_grants (tenant_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_license_grants_function
    ON marketplace_license_grants (tenant_id, function_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_license_grants_active
    ON marketplace_license_grants (tenant_id)
    WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS marketplace_license_activations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grant_id UUID NOT NULL REFERENCES marketplace_license_grants(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    user_id UUID,
    activation_label VARCHAR(255),
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_marketplace_license_activations_grant
    ON marketplace_license_activations (grant_id);

COMMENT ON TABLE marketplace_function_license_policies IS 'Per-function SPDX and commercial licensing policy for marketplace publishers';
COMMENT ON TABLE marketplace_license_grants IS 'Issued license keys for restricted/commercial marketplace functions';
COMMENT ON TABLE marketplace_license_activations IS 'Audit trail of license key activations';

COMMIT;
