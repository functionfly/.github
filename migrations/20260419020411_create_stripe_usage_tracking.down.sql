-- Migration: Drop Stripe usage reporting tracking tables
-- Created: 2026-04-19

DROP TABLE IF EXISTS billing_usage_reconciliation CASCADE;
DROP TABLE IF EXISTS stripe_usage_reports CASCADE;

DROP FUNCTION IF EXISTS update_stripe_usage_report_updated_at();
