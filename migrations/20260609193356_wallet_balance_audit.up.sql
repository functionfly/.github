-- Migration: wallet_balance_audit
-- Purpose: Audit table for tracking balance drift between stored balance and transaction ledger
-- Created: 20260609193356

CREATE TABLE IF NOT EXISTS wallet_balance_audit (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    stored_balance  DECIMAL(15,6) NOT NULL,
    computed_balance DECIMAL(15,6) NOT NULL,
    drift           DECIMAL(15,6) NOT NULL,
    fixed           BOOLEAN DEFAULT FALSE,
    fixed_at        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wallet_balance_audit_wallet ON wallet_balance_audit (wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_balance_audit_created ON wallet_balance_audit (created_at);
