-- Migration: Drop dunning management tables
-- Created: 2026-04-19

DROP TABLE IF EXISTS service_suspensions CASCADE;
DROP TABLE IF EXISTS dunning_notifications CASCADE;
DROP TABLE IF EXISTS payment_retries CASCADE;
DROP TABLE IF EXISTS payment_retry_schedules CASCADE;

DROP FUNCTION IF EXISTS update_payment_retry_updated_at();
DROP FUNCTION IF EXISTS audit_payment_retry_changes();
