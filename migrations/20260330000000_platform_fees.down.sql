-- Rollback platform fees migration

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS fee_transactions;
DROP TABLE IF EXISTS platform_fees;
DROP TABLE IF EXISTS user_wallets;

-- Remove columns from registry_functions
ALTER TABLE registry_functions
    DROP COLUMN IF EXISTS platform_fee_paid,
    DROP COLUMN IF EXISTS platform_fee_amount_usd,
    DROP COLUMN IF EXISTS last_fee_charged_at;
