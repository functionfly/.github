-- Verification queries for unified wallet migration
-- Run these to validate the migration was successful

-- ============================================
-- 1. Basic wallet counts verification
-- ============================================

-- Expected: wallets count should equal user_wallets + agent_billing_controls
SELECT
    'Legacy Tables' as source,
    (SELECT COUNT(*) FROM user_wallets) +
    (SELECT COUNT(*) FROM agent_billing_controls) as total
UNION ALL
SELECT
    'Unified Wallet Table' as source,
    COUNT(*) as total
FROM wallets;

-- ============================================
-- 2. Balance verification
-- ============================================

-- Check that unified wallet balances match legacy balances
-- Run for user wallets
SELECT
    uw.user_id,
    uw.balance_usd as legacy_balance,
    w.balance_usd as unified_balance,
    CASE WHEN uw.balance_usd = w.balance_usd THEN 'MATCH' ELSE 'MISMATCH' END as status
FROM user_wallets uw
JOIN wallets w ON w.user_id = uw.user_id AND w.owner_type = 'user'
WHERE uw.balance_usd != w.balance_usd;

-- ============================================
-- 3. Transaction counts verification
-- ============================================

-- User wallet transactions should match fee_transactions count
SELECT
    'fee_transactions' as source,
    COUNT(*) as total
FROM fee_transactions ft
WHERE EXISTS (
    SELECT 1 FROM wallets w
    WHERE w.user_id = ft.user_id AND w.owner_type = 'user'
)
UNION ALL
SELECT
    'wallet_transactions (user)',
    COUNT(*)
FROM wallet_transactions wt
JOIN wallets w ON w.id = wt.wallet_id
WHERE w.owner_type = 'user';

-- ============================================
-- 4. Check for orphan wallets (wallets without corresponding legacy data)
-- ============================================

-- User wallets without legacy user_wallets entry (acceptable if created via new flow)
SELECT
    w.id,
    w.owner_id,
    w.balance_usd,
    'No legacy user_wallets' as note
FROM wallets w
WHERE w.owner_type = 'user'
AND NOT EXISTS (
    SELECT 1 FROM user_wallets uw
    WHERE uw.user_id::TEXT = w.owner_id
);

-- ============================================
-- 5. Verify transaction types in wallet_transactions
-- ============================================

SELECT
    transaction_type,
    COUNT(*) as count,
    SUM(amount_usd) as total_amount
FROM wallet_transactions
GROUP BY transaction_type
ORDER BY transaction_type;

-- ============================================
-- 6. Sample recent transactions for audit
-- ============================================

SELECT
    wt.id,
    wt.transaction_type,
    wt.amount_usd,
    wt.balance_before_usd,
    wt.balance_after_usd,
    wt.status,
    wt.created_at,
    w.owner_id
FROM wallet_transactions wt
JOIN wallets w ON w.id = wt.wallet_id
ORDER BY wt.created_at DESC
LIMIT 10;

-- ============================================
-- 7. Quick balance check for admin review
-- ============================================

SELECT
    owner_type,
    COUNT(*) as wallet_count,
    SUM(balance_usd) as total_balance,
    AVG(balance_usd) as avg_balance,
    MIN(balance_usd) as min_balance,
    MAX(balance_usd) as max_balance
FROM wallets
GROUP BY owner_type;

-- ============================================
-- 8. Check for wallets with zero balance but lifetime activity
-- ============================================

SELECT
    w.id,
    w.owner_type,
    w.owner_id,
    w.balance_usd,
    w.lifetime_earnings_usd,
    w.lifetime_spent_usd,
    CASE
        WHEN w.lifetime_earnings_usd > 0 AND w.balance_usd = 0 THEN 'FULLY SPENT'
        WHEN w.lifetime_earnings_usd = 0 AND w.lifetime_spent_usd = 0 AND w.balance_usd = 0 THEN 'NEVER FUNDED'
        ELSE 'NORMAL'
    END as status
FROM wallets w
WHERE w.balance_usd = 0
AND (w.lifetime_earnings_usd > 0 OR w.lifetime_spent_usd > 0)
ORDER BY w.owner_type, w.owner_id;