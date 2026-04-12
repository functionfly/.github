-- Add unique constraint to fee_transactions.reference column to prevent duplicate wallet credits
-- This fixes a race condition where duplicate webhooks could credit a wallet twice

-- Add partial unique index: reference must be unique when it's not NULL (for credit transactions)
CREATE UNIQUE INDEX IF NOT EXISTS idx_fee_transactions_reference_unique
    ON fee_transactions (reference)
    WHERE reference IS NOT NULL AND reference != '';

-- Add a comment documenting the purpose of this constraint
COMMENT ON INDEX idx_fee_transactions_reference_unique IS 'Prevents duplicate wallet credits from race conditions during webhook processing';
