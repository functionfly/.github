-- Marketplace purchases: agent function purchases, hirings, idempotency, audit, buyer indexes

BEGIN;

CREATE TABLE IF NOT EXISTS agent_hirings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    hirer_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL,
    task_type VARCHAR(255) NOT NULL,
    task_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    budget_usd DECIMAL(12, 4) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT agent_hirings_status_check
        CHECK (status IN ('pending', 'active', 'completed', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_agent_hirings_tenant_created
    ON agent_hirings (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_hirings_hirer
    ON agent_hirings (hirer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_hirings_agent
    ON agent_hirings (agent_id, created_at DESC);

CREATE TABLE IF NOT EXISTS function_purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    function_author VARCHAR(255) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    published_id UUID NOT NULL,
    price_paid_usd DECIMAL(12, 4) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT function_purchases_status_check
        CHECK (status IN ('completed', 'refunded', 'disputed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_function_purchases_active_unique
    ON function_purchases (agent_id, function_author, function_name)
    WHERE status = 'completed';

CREATE INDEX IF NOT EXISTS idx_function_purchases_agent_created
    ON function_purchases (agent_id, created_at DESC);

CREATE TABLE IF NOT EXISTS marketplace_purchase_idempotency (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    purchase_id UUID NOT NULL REFERENCES function_purchases(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT marketplace_purchase_idempotency_unique UNIQUE (tenant_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS marketplace_purchase_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID,
    agent_id VARCHAR(255) NOT NULL,
    function_author VARCHAR(255) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    purchase_id UUID,
    price_paid_usd DECIMAL(12, 4) NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(128),
    client_ip INET,
    event_type VARCHAR(64) NOT NULL DEFAULT 'function_purchase',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_marketplace_purchase_audit_tenant
    ON marketplace_purchase_audit_log (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_marketplace_license_grants_purchaser_tenant
    ON marketplace_license_grants (purchaser_tenant_id, created_at DESC)
    WHERE purchaser_tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_marketplace_license_grants_purchaser_user
    ON marketplace_license_grants (purchaser_user_id, created_at DESC)
    WHERE purchaser_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_marketplace_plan_subscriptions_subscriber_tenant
    ON marketplace_plan_subscriptions (subscriber_tenant_id, created_at DESC)
    WHERE subscriber_tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_marketplace_plan_subscriptions_subscriber_user
    ON marketplace_plan_subscriptions (subscriber_user_id, created_at DESC)
    WHERE subscriber_user_id IS NOT NULL;

COMMENT ON TABLE function_purchases IS 'Agent wallet purchases of marketplace functions';
COMMENT ON TABLE marketplace_purchase_idempotency IS 'Idempotency keys for POST /v1/marketplace/purchase';
COMMENT ON TABLE marketplace_purchase_audit_log IS 'Audit trail for marketplace purchase events';

COMMIT;
