-- Migration: Finalize unified wallet system
-- This migration migrates all data from legacy tables into the unified wallets table
-- and sets wallet_transactions from fee_transactions
--
-- NOTE: Before running this, ensure 000254_unified_wallet_system.up.sql has been applied

BEGIN;

-- ============================================
-- 1. Migrate user_wallets to wallets (user type)
-- ============================================

INSERT INTO wallets (
    id,
    owner_type,
    owner_id,
    user_id,
    agent_id,
    wallet_type,
    balance_usd,
    lifetime_earnings_usd,
    lifetime_spent_usd,
    billing_mode,
    status,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() AS id,
    'user' AS owner_type,
    user_id::TEXT AS owner_id,
    user_id AS user_id,
    NULL::TEXT AS agent_id,
    'unified' AS wallet_type,
    balance_usd,
    lifetime_earnings_usd,
    lifetime_fees_usd AS lifetime_spent_usd,
    'per_wallet' AS billing_mode,
    'active' AS status,
    created_at,
    updated_at
FROM user_wallets uw
WHERE NOT EXISTS (
    SELECT 1 FROM wallets w WHERE w.owner_type = 'user' AND w.owner_id = uw.user_id::TEXT
);

-- ============================================
-- 2. Migrate agent_billing_controls to wallets (agent type)
-- ============================================

INSERT INTO wallets (
    id,
    owner_type,
    owner_id,
    user_id,
    agent_id,
    wallet_type,
    balance_usd,
    lifetime_earnings_usd,
    lifetime_spent_usd,
    spend_cap_monthly_usd,
    spend_cap_daily_usd,
    alert_thresholds,
    billing_mode,
    team_id,
    status,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() AS id,
    'agent' AS owner_type,
    agent_id AS owner_id,
    NULL::UUID AS user_id,
    agent_id AS agent_id,
    'unified' AS wallet_type,
    credit_balance_usd AS balance_usd,
    0 AS lifetime_earnings_usd,
    0 AS lifetime_spent_usd,
    spend_cap_monthly_usd,
    spend_cap_daily_usd,
    COALESCE(alert_thresholds, '{0.5, 0.8, 0.95}'::DECIMAL[]),
    COALESCE(billing_mode, 'per_agent') AS billing_mode,
    team_id,
    'active' AS status,
    created_at,
    updated_at
FROM agent_billing_controls abc
WHERE NOT EXISTS (
    SELECT 1 FROM wallets w WHERE w.owner_type = 'agent' AND w.owner_id = abc.agent_id
);

-- ============================================
-- 3. Migrate fee_transactions to wallet_transactions (for user wallets)
-- ============================================

INSERT INTO wallet_transactions (
    id,
    wallet_id,
    transaction_type,
    amount_usd,
    balance_before_usd,
    balance_after_usd,
    status,
    reference,
    triggered_by_type,
    triggered_by_id,
    fee_type,
    metadata,
    created_at,
    completed_at
)
SELECT
    ft.id,
    w.id AS wallet_id,
    CASE ft.kind
        WHEN 'credit' THEN 'credit'
        WHEN 'debit' THEN 'debit'
        WHEN 'fee_payment' THEN 'fee_payment'
        WHEN 'commission' THEN 'commission'
        ELSE 'adjustment'
    END AS transaction_type,
    ft.amount_usd,
    0 AS balance_before_usd,
    0 AS balance_after_usd,
    ft.status,
    ft.reference,
    'user' AS triggered_by_type,
    ft.user_id::TEXT AS triggered_by_id,
    CASE
        WHEN ft.reference LIKE 'fee_payment_publish%' THEN 'publish'
        WHEN ft.reference LIKE 'fee_payment_version_update%' THEN 'version_update'
        WHEN ft.reference LIKE 'commission%' THEN 'commission'
        ELSE NULL
    END AS fee_type,
    ft.metadata,
    ft.created_at,
    ft.created_at AS completed_at
FROM fee_transactions ft
JOIN wallets w ON w.user_id = ft.user_id AND w.owner_type = 'user'
WHERE NOT EXISTS (
    SELECT 1 FROM wallet_transactions wt WHERE wt.id = ft.id
);

-- ============================================
-- 4. Migrate agent_revenue_transactions to wallet_transactions (for agent wallets)
-- Note: agent_financial_transactions was renamed to agent_revenue_transactions
-- ============================================

INSERT INTO wallet_transactions (
    id,
    wallet_id,
    transaction_type,
    amount_usd,
    balance_before_usd,
    balance_after_usd,
    status,
    reference,
    triggered_by_type,
    triggered_by_id,
    execution_id,
    function_id,
    metadata,
    created_at,
    completed_at
)
SELECT
    art.id,
    w.id AS wallet_id,
    CASE art.transaction_type
        WHEN 'delegation_payment' THEN 'commission'
        WHEN 'function_call' THEN 'execution_charge'
        WHEN 'revenue_share' THEN 'commission'
        WHEN 'refund' THEN 'refund'
        ELSE 'debit'
    END AS transaction_type,
    art.amount_usd,
    0 AS balance_before_usd,
    0 AS balance_after_usd,
    art.status,
    art.execution_id AS reference,
    'agent' AS triggered_by_type,
    art.to_agent_id AS triggered_by_id,
    art.execution_id::UUID,
    art.function_id,
    jsonb_build_object(
        'session_id', art.session_id,
        'parent_execution_id', art.parent_execution_id
    ) AS metadata,
    art.created_at,
    art.created_at AS completed_at
FROM agent_revenue_transactions art
JOIN wallets w ON w.agent_id = art.to_agent_id AND w.owner_type = 'agent'
WHERE NOT EXISTS (
    SELECT 1 FROM wallet_transactions wt WHERE wt.id = art.id
);

-- ============================================
-- 5. Verify migration with summary query
-- ============================================

DO $$
DECLARE
    user_wallets_count INTEGER;
    agent_controls_count INTEGER;
    wallets_count INTEGER;
    wallets_user_count INTEGER;
    wallets_agent_count INTEGER;
    transactions_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_wallets_count FROM user_wallets;
    SELECT COUNT(*) INTO agent_controls_count FROM agent_billing_controls;
    SELECT COUNT(*) INTO wallets_count FROM wallets;
    SELECT COUNT(*) INTO wallets_user_count FROM wallets WHERE owner_type = 'user';
    SELECT COUNT(*) INTO wallets_agent_count FROM wallets WHERE owner_type = 'agent';
    SELECT COUNT(*) INTO transactions_count FROM wallet_transactions;

    RAISE NOTICE '=== Unified Wallet Migration Complete ===';
    RAISE NOTICE 'User wallets migrated: %', wallets_user_count;
    RAISE NOTICE 'Agent billing controls migrated: %', wallets_agent_count;
    RAISE NOTICE 'Total wallets in unified system: %', wallets_count;
    RAISE NOTICE 'Total wallet transactions: %', transactions_count;

    -- Validate: wallets count should equal user_wallets + agent_billing_controls
    IF wallets_count != (user_wallets_count + agent_controls_count) THEN
        RAISE WARNING 'Wallet count mismatch: expected % but got %', (user_wallets_count + agent_controls_count), wallets_count;
    END IF;
END $$;

COMMIT;