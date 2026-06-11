-- Drop GoBetterAuth Base Tables
-- Removes the core tables for GoBetterAuth authentication system

-- Drop indexes first
DROP INDEX IF EXISTS idx_gba_users_email_tenant;
DROP INDEX IF EXISTS idx_gba_users_tenant;
DROP INDEX IF EXISTS idx_gba_accounts_user;
DROP INDEX IF EXISTS idx_gba_accounts_provider;
DROP INDEX IF EXISTS idx_gba_sessions_user;
DROP INDEX IF EXISTS idx_gba_sessions_token;
DROP INDEX IF EXISTS idx_gba_sessions_expires;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS gba_sessions;
DROP TABLE IF EXISTS gba_accounts;
DROP TABLE IF EXISTS gba_users;
DROP TABLE IF EXISTS gba_tenants;
