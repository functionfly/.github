-- Migration: wallet_performance_indexes
-- Purpose: Additional indexes for wallet performance optimization
-- Created: 20260609193357

-- Index for general balance range queries (not just low balance)
CREATE INDEX IF NOT EXISTS idx_wallets_balance_range
    ON wallets (balance_usd)
    WHERE balance_usd >= 0;

-- Index for wallet type filtering with balance
CREATE INDEX IF NOT EXISTS idx_wallets_type_balance
    ON wallets (wallet_type, balance_usd)
    WHERE status = 'active';

-- Composite index for spend cap queries
CREATE INDEX IF NOT EXISTS idx_wallets_owner_spend_caps
    ON wallets (owner_type, owner_id)
    INCLUDE (spend_cap_daily_usd, spend_cap_monthly_usd);

-- Index for transaction lookups by wallet and status
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_status
    ON wallet_transactions (wallet_id, status)
    WHERE status = 'completed';
