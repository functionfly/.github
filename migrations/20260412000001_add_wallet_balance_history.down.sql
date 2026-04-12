-- Down migration: Remove wallet balance history table
DROP TABLE IF EXISTS wallet_balance_history;
