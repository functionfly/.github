-- Unified Wallet System Down Migration
-- Reverts to separate user_wallets and agent_billing_controls tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP FUNCTION IF EXISTS update_wallet_updated_at();
DROP FUNCTION IF EXISTS validate_wallet_balance();

-- Drop views
DROP VIEW IF EXISTS wallet_summary;
DROP VIEW IF EXISTS user_wallets_compat;
DROP VIEW IF EXISTS agent_billing_controls_compat;

-- Drop indexes (automatically dropped with tables, but explicit for safety)
DROP INDEX IF EXISTS idx_wallet_transactions_idempotency_unique;

-- Drop tables (in dependency order)
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;

-- Note: Data migration back to old tables is NOT automatically performed.
-- If you need to rollback with data preservation, run the data migration script first:
-- cmd/migrate-wallets/rollback.go
