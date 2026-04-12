-- Migration: Add wallet balance history table for time-series tracking
-- This enables analytics, auditing, and reconciliation capabilities

CREATE TABLE IF NOT EXISTS wallet_balance_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    balance_usd DECIMAL(14,4) NOT NULL,
    change_amount_usd DECIMAL(14,4) NOT NULL DEFAULT 0,
    transaction_id UUID REFERENCES wallet_transactions(id) ON DELETE SET NULL,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    recorded_date DATE NOT NULL DEFAULT CURRENT_DATE,
    snapshot_type VARCHAR(20) NOT NULL DEFAULT 'transactional', -- 'transactional', 'scheduled', 'manual', 'reconciliation'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_wallet_balance_history_wallet_id ON wallet_balance_history(wallet_id);
CREATE INDEX idx_wallet_balance_history_recorded_at ON wallet_balance_history(recorded_at DESC);
CREATE INDEX idx_wallet_balance_history_recorded_date ON wallet_balance_history(recorded_date DESC);
CREATE INDEX idx_wallet_balance_history_transaction_id ON wallet_balance_history(transaction_id);
CREATE INDEX idx_wallet_balance_history_snapshot_type ON wallet_balance_history(snapshot_type);

-- Composite index for time-series queries (wallet + date range)
CREATE INDEX idx_wallet_balance_history_wallet_date ON wallet_balance_history(wallet_id, recorded_date DESC);

-- Unique constraint to prevent duplicate snapshots for the same transaction
CREATE UNIQUE INDEX idx_wallet_balance_history_unique_tx 
ON wallet_balance_history(wallet_id, transaction_id) 
WHERE transaction_id IS NOT NULL;

-- Partial index for latest balance lookups per wallet
CREATE INDEX idx_wallet_balance_history_latest 
ON wallet_balance_history(wallet_id, recorded_at DESC) 
WHERE snapshot_type = 'scheduled';

-- Table comment
COMMENT ON TABLE wallet_balance_history IS 'Time-series record of wallet balance changes for analytics and auditing';
