-- Migration: Remove billing performance indexes (rollback)
-- Created: 2026-04-19

-- Drop indexes added in the up migration
DROP INDEX IF EXISTS idx_payment_retries_status_next_retry;
DROP INDEX IF EXISTS idx_payment_retries_grace_period_status;
DROP INDEX IF EXISTS idx_invoices_period_tenant;
DROP INDEX IF EXISTS idx_invoices_tenant_period_status;
DROP INDEX IF EXISTS idx_cost_allocation_entries_timestamp;
DROP INDEX IF EXISTS idx_cost_allocation_entries_old;
