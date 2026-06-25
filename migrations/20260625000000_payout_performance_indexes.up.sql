-- Add missing index on stripe_connect_accounts.stripe_account_id for webhook lookups
-- The column has UNIQUE constraint but explicit index improves webhook lookup performance

BEGIN;

CREATE INDEX IF NOT EXISTS idx_stripe_connect_accounts_stripe_id
    ON stripe_connect_accounts(stripe_account_id);

-- Add composite index for payout_ledger balance queries (user_id + created_at desc)
-- This optimizes the common pattern of fetching user's recent ledger entries
CREATE INDEX IF NOT EXISTS idx_payout_ledger_user_created
    ON payout_ledger(user_id, created_at DESC);

COMMIT;
