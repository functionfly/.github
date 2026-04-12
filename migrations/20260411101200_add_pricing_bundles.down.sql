-- Rollback: Backend-in-a-Box pricing bundles
DROP TABLE IF EXISTS bundle_subscriptions;
DROP TABLE IF EXISTS founder_mode_registrations;
DROP TABLE IF EXISTS deferred_billing_configs;
DROP TABLE IF EXISTS pricing_bundles;
