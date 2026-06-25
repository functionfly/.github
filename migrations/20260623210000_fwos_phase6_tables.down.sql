-- Migration: 20260623210000_fwos_phase6_tables.down.sql
-- Description: Drop FWOS phase 6 tables (email accounts, devices, SSO, wallet passes, notifications)

DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS push_subscriptions;
DROP TABLE IF EXISTS wallet_passes;
DROP TABLE IF EXISTS sso_provisioning_logs;
DROP TABLE IF EXISTS sso_provisioning_configs;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS email_accounts;
