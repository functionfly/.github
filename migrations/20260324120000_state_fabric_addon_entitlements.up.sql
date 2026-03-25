-- State Fabric add-on entitlements per tenant (Stripe subscription items can be linked for reconciliation).

CREATE TABLE IF NOT EXISTS state_fabric_addon_entitlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    addon_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    stripe_subscription_id VARCHAR(128),
    stripe_subscription_item_id VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_sf_addon_tenant UNIQUE (tenant_id, addon_id)
);

CREATE INDEX IF NOT EXISTS idx_sf_addon_entitlements_tenant
    ON state_fabric_addon_entitlements (tenant_id);

CREATE INDEX IF NOT EXISTS idx_sf_addon_entitlements_addon
    ON state_fabric_addon_entitlements (addon_id);
