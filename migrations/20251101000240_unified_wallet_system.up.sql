-- Unified Wallet System Migration
-- Merges user_wallets (registry platform fees) and agent_billing_controls.credit_balance_usd (execution credits)
-- into a single unified wallet system with wallet_transactions ledger

-- ============================================
-- 1. Create unified wallets table
-- ============================================
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Owner identification (polymorphic)
    owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'agent')),
    owner_id TEXT NOT NULL,  -- user_id (UUID as string) or agent_id

    -- For user wallets: links to user
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    -- For agent wallets: links to agent
    agent_id TEXT REFERENCES agent_identities(agent_id) ON DELETE CASCADE,

    -- Wallet purpose (what this wallet can be used for)
    wallet_type TEXT NOT NULL DEFAULT 'unified' CHECK (wallet_type IN ('unified', 'registry', 'execution')),

    -- Core balance (combined from old systems)
    balance_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,

    -- Legacy tracking fields (migrated from user_wallets)
    lifetime_earnings_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,  -- From user_wallets
    lifetime_spent_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,     -- From user_wallets (renamed from lifetime_fees)

    -- Agent-specific controls (migrated from agent_billing_controls)
    spend_cap_monthly_usd DECIMAL(10, 2),
    spend_cap_daily_usd DECIMAL(10, 2),
    alert_thresholds DECIMAL[] NOT NULL DEFAULT '{0.5, 0.8, 0.95}',

    -- Billing mode for agents
    billing_mode TEXT NOT NULL DEFAULT 'per_wallet' CHECK (billing_mode IN ('per_wallet', 'per_agent', 'per_tenant', 'per_team')),
    team_id UUID,

    -- Soft delete and status
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
    closed_at TIMESTAMPTZ,
    closure_reason TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT unique_owner UNIQUE (owner_type, owner_id),
    CONSTRAINT user_id_required_for_user CHECK (
        (owner_type = 'user' AND user_id IS NOT NULL AND agent_id IS NULL) OR
        (owner_type = 'agent' AND agent_id IS NOT NULL AND user_id IS NULL) OR
        (owner_type NOT IN ('user', 'agent'))
    )
);

-- Indexes for wallets
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets (user_id) WHERE owner_type = 'user';
CREATE INDEX IF NOT EXISTS idx_wallets_agent_id ON wallets (agent_id) WHERE owner_type = 'agent';
CREATE INDEX IF NOT EXISTS idx_wallets_owner_type ON wallets (owner_type);
CREATE INDEX IF NOT EXISTS idx_wallets_status ON wallets (status);
CREATE INDEX IF NOT EXISTS idx_wallets_team_id ON wallets (team_id) WHERE team_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_wallets_low_balance ON wallets (balance_usd) WHERE balance_usd < 10;

-- ============================================
-- 2. Create unified wallet_transactions ledger
-- ============================================
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Wallet reference
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,

    -- Transaction categorization
    transaction_type TEXT NOT NULL CHECK (transaction_type IN (
        'credit',           -- Adding funds (Stripe checkout)
        'debit',            -- Generic deduction
        'fee_payment',      -- Paying platform fees (publish, version update)
        'execution_charge', -- Paying for function execution
        'commission',       -- Platform commission on sales
        'refund',           -- Refund to wallet
        'transfer_in',      -- Transfer from another wallet
        'transfer_out',     -- Transfer to another wallet
        'adjustment'        -- Manual/admin adjustment
    )),

    -- Amount (always positive; direction determined by type)
    amount_usd DECIMAL(14, 4) NOT NULL,

    -- Running balance snapshot (for audit)
    balance_before_usd DECIMAL(14, 4) NOT NULL,
    balance_after_usd DECIMAL(14, 4) NOT NULL,

    -- Status
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed', 'reversed')),

    -- References for idempotency and tracing
    reference TEXT,  -- External reference (Stripe payment intent, checkout session, etc.)
    parent_transaction_id UUID REFERENCES wallet_transactions(id) ON DELETE SET NULL,

    -- Context (who/what triggered this)
    triggered_by_type TEXT CHECK (triggered_by_type IN ('user', 'agent', 'system', 'admin', 'webhook')),
    triggered_by_id TEXT,  -- user_id or agent_id

    -- For execution charges: what was executed
    execution_id UUID,
    function_id UUID,

    -- For fee payments: what fee was paid
    fee_type TEXT CHECK (fee_type IN ('publish', 'version_update', 'commission')),

    -- Rich metadata
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    reversed_at TIMESTAMPTZ,

    -- Idempotency key (unique per external operation)
    idempotency_key TEXT
);

-- Indexes for wallet_transactions
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions (wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions (transaction_type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference ON wallet_transactions (reference) WHERE reference IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_status ON wallet_transactions (status);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_idempotency ON wallet_transactions (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_triggered ON wallet_transactions (triggered_by_type, triggered_by_id) WHERE triggered_by_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_execution ON wallet_transactions (execution_id) WHERE execution_id IS NOT NULL;

-- Unique constraint on idempotency key to prevent double-processing
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_transactions_idempotency_unique ON wallet_transactions (idempotency_key) WHERE idempotency_key IS NOT NULL;

-- ============================================
-- 3. Create wallet transaction summary view
-- ============================================
CREATE OR REPLACE VIEW wallet_summary AS
SELECT
    w.id AS wallet_id,
    w.owner_type,
    w.owner_id,
    w.user_id,
    w.agent_id,
    w.balance_usd,
    w.status,

    -- Transaction aggregates
    COALESCE(SUM(CASE WHEN wt.transaction_type = 'credit' AND wt.status = 'completed' THEN wt.amount_usd END), 0) AS total_credits_usd,
    COALESCE(SUM(CASE WHEN wt.transaction_type = 'debit' AND wt.status = 'completed' THEN wt.amount_usd END), 0) AS total_debits_usd,
    COALESCE(SUM(CASE WHEN wt.transaction_type = 'fee_payment' AND wt.status = 'completed' THEN wt.amount_usd END), 0) AS total_fees_paid_usd,
    COALESCE(SUM(CASE WHEN wt.transaction_type = 'execution_charge' AND wt.status = 'completed' THEN wt.amount_usd END), 0) AS total_execution_charges_usd,
    COALESCE(SUM(CASE WHEN wt.transaction_type = 'commission' AND wt.status = 'completed' THEN wt.amount_usd END), 0) AS total_commissions_usd,

    -- Counts
    COUNT(CASE WHEN wt.status = 'completed' THEN 1 END) AS total_transactions,
    COUNT(CASE WHEN wt.status = 'pending' THEN 1 END) AS pending_transactions,

    -- Last activity
    MAX(CASE WHEN wt.status = 'completed' THEN wt.created_at END) AS last_transaction_at,

    w.created_at,
    w.updated_at

FROM wallets w
LEFT JOIN wallet_transactions wt ON wt.wallet_id = w.id
GROUP BY w.id, w.owner_type, w.owner_id, w.user_id, w.agent_id, w.balance_usd, w.status, w.created_at, w.updated_at;

-- ============================================
-- 4. Create function to auto-update updated_at
-- ============================================
CREATE OR REPLACE FUNCTION update_wallet_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_wallet_updated_at();

-- ============================================
-- 5. Create function to validate wallet balance consistency
-- ============================================
CREATE OR REPLACE FUNCTION validate_wallet_balance()
RETURNS TRIGGER AS $$
DECLARE
    expected_balance DECIMAL(14, 4);
    actual_balance DECIMAL(14, 4);
BEGIN
    -- Calculate expected balance from transaction history
    SELECT COALESCE(SUM(
        CASE
            WHEN transaction_type IN ('credit', 'refund', 'transfer_in') AND status = 'completed' THEN amount_usd
            WHEN transaction_type IN ('debit', 'fee_payment', 'execution_charge', 'commission', 'transfer_out') AND status = 'completed' THEN -amount_usd
            ELSE 0
        END
    ), 0)
    INTO expected_balance
    FROM wallet_transactions
    WHERE wallet_id = NEW.wallet_id AND status = 'completed';

    -- Get actual balance from wallet
    SELECT balance_usd INTO actual_balance
    FROM wallets
    WHERE id = NEW.wallet_id;

    -- Allow small floating point differences
    IF ABS(expected_balance - actual_balance) > 0.0001 THEN
        RAISE WARNING 'Wallet balance mismatch for wallet %: expected %, actual %',
            NEW.wallet_id, expected_balance, actual_balance;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to validate balance after transaction completion (optional, can be disabled for performance)
-- CREATE TRIGGER validate_wallet_balance_after_transaction
--     AFTER INSERT OR UPDATE ON wallet_transactions
--     FOR EACH ROW
--     WHEN (NEW.status = 'completed')
--     EXECUTE FUNCTION validate_wallet_balance();

-- ============================================
-- 6. Migration data preservation notes
-- ============================================
-- NOTE: Data migration from old tables should be performed by a separate migration script
-- that runs after this schema migration:
--
-- 1. Migrate user_wallets -> wallets (owner_type='user')
-- 2. Migrate agent_billing_controls.credit_balance_usd -> wallets (owner_type='agent')
-- 3. Migrate fee_transactions -> wallet_transactions (transaction_type='fee_payment' or 'credit')
-- 4. Migrate agent_financial_transactions -> wallet_transactions (transaction_type='credit' or 'execution_charge')

-- ============================================
-- 7. Create backup/denormalized views for backward compatibility
-- ============================================

-- View that mimics old user_wallets table structure
CREATE OR REPLACE VIEW user_wallets_compat AS
SELECT
    w.user_id,
    w.balance_usd,
    w.lifetime_earnings_usd AS lifetime_earnings_usd,
    w.lifetime_spent_usd AS lifetime_fees_usd,
    w.created_at,
    w.updated_at
FROM wallets w
WHERE w.owner_type = 'user';

-- View that mimics agent_billing_controls structure (without credit_balance_usd since it's in wallets)
CREATE OR REPLACE VIEW agent_billing_controls_compat AS
SELECT
    w.id,
    w.agent_id,
    w.spend_cap_monthly_usd,
    w.spend_cap_daily_usd,
    w.balance_usd AS credit_balance_usd,
    w.billing_mode,
    w.team_id,
    w.alert_thresholds,
    w.created_at,
    w.updated_at
FROM wallets w
WHERE w.owner_type = 'agent';

-- ============================================
-- 8. Comments for documentation
-- ============================================
COMMENT ON TABLE wallets IS 'Unified wallet system - merged from user_wallets and agent_billing_controls';
COMMENT ON TABLE wallet_transactions IS 'Unified transaction ledger - merged from fee_transactions and agent_financial_transactions';
COMMENT ON COLUMN wallets.owner_type IS 'Type of wallet owner: user (for registry fees) or agent (for execution credits)';
COMMENT ON COLUMN wallets.wallet_type IS 'Purpose of wallet: unified (both purposes), registry (publish/version fees), or execution (function calls)';
COMMENT ON COLUMN wallet_transactions.reference IS 'External reference for idempotency (Stripe payment intent, checkout session ID, etc.)';
COMMENT ON COLUMN wallet_transactions.idempotency_key IS 'Unique key to prevent duplicate transaction processing';
