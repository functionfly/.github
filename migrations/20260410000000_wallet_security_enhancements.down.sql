-- Wallet Security Enhancements Down Migration
-- Reverts the security enhancements

-- Drop secondary approval tracking table
DROP TABLE IF EXISTS wallet_secondary_approvals;

-- Drop admin adjustment audit log table
DROP TABLE IF EXISTS wallet_admin_adjustment_audit;

-- Drop encrypted balance column
ALTER TABLE wallets DROP COLUMN IF EXISTS balance_encrypted;

-- Drop index
DROP INDEX IF EXISTS idx_wallets_balance_encrypted;
