-- Migration: Add billing performance indexes for dunning and reporting
-- Created: 2026-04-19

-- ============================================================
-- 11. Missing Indexes for Dunning Scheduler
-- ============================================================

-- Composite index on payment_retries.status and next_retry_at for dunning scheduler queries
-- This optimizes queries that find retries needing to be scheduled/processed
CREATE INDEX IF NOT EXISTS idx_payment_retries_status_next_retry 
    ON payment_retries(status, next_retry_at) 
    WHERE status IN ('active', 'paused');

-- Additional index for grace period queries (service suspension check)
CREATE INDEX IF NOT EXISTS idx_payment_retries_grace_period_status 
    ON payment_retries(grace_period_ends_at, status) 
    WHERE status = 'active' AND grace_period_ends_at < NOW() + INTERVAL '7 days';

-- ============================================================
-- 11. Missing Indexes for Reporting Queries
-- ============================================================

-- Composite index on invoices.period_start, period_end, tenant_id for reporting queries
-- Optimizes period-based billing reports and aggregations
CREATE INDEX IF NOT EXISTS idx_invoices_period_tenant 
    ON invoices(period_start, period_end, tenant_id) 
    WHERE status IN ('open', 'paid', 'uncollectible');

-- Index for invoice status filtering in reporting
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_period_status 
    ON invoices(tenant_id, period_start DESC, period_end DESC, status);

-- ============================================================
-- Additional performance indexes for cost allocation cleanup
-- ============================================================

-- Index for data retention cleanup queries on cost_allocation_entries
-- Optimizes the DELETE queries for old entries beyond retention period
CREATE INDEX IF NOT EXISTS idx_cost_allocation_entries_timestamp 
    ON cost_allocation_entries(timestamp DESC);

-- Partial index for entries older than 90 days (cleanup candidates)
CREATE INDEX IF NOT EXISTS idx_cost_allocation_entries_old 
    ON cost_allocation_entries(timestamp) 
    WHERE timestamp < NOW() - INTERVAL '90 days';

-- ============================================================
-- Index comments for documentation
-- ============================================================
COMMENT ON INDEX idx_payment_retries_status_next_retry IS 
    'Optimizes dunning scheduler queries for finding retries needing processing';
COMMENT ON INDEX idx_payment_retries_grace_period_status IS 
    'Optimizes grace period monitoring and service suspension decisions';
COMMENT ON INDEX idx_invoices_period_tenant IS 
    'Optimizes period-based billing reports and revenue recognition queries';
COMMENT ON INDEX idx_invoices_tenant_period_status IS 
    'Optimizes tenant invoice history and status-based reporting';
COMMENT ON INDEX idx_cost_allocation_entries_timestamp IS 
    'Optimizes cost allocation queries and cleanup operations';
COMMENT ON INDEX idx_cost_allocation_entries_old IS 
    'Partial index for identifying cost entries beyond retention period for cleanup';
