-- Agent wallet / billing ledger (idempotent Stripe and internal events)

CREATE TABLE IF NOT EXISTS agent_financial_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN (
        'credit_purchase',
        'execution_debit',
        'transfer_in',
        'transfer_out',
        'adjustment',
        'refund'
    )),
    amount_usd DECIMAL(14, 4) NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed')),
    provider TEXT,
    provider_ref TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_financial_tx_agent_created
    ON agent_financial_transactions (agent_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_financial_tx_tenant
    ON agent_financial_transactions (tenant_id);

-- Stripe rows always set both; Postgres treats NULLs as distinct so multiple internal rows may omit provider.
CREATE UNIQUE INDEX IF NOT EXISTS uq_agent_financial_tx_provider_ref
    ON agent_financial_transactions (provider, provider_ref);
