-- Migration: 20260615000002_add_missing_billing_indexes
-- Description: Add missing indexes for billing query performance and fix schema issues.
-- Fixes: usage_events_v2 composite index, cost_allocation_entries timestamp index,
-- invoices subscription_id FK action, invoices currency CHECK constraint

-- ============================================
-- 1. usage_events_v2: Add composite index for tenant + timestamp queries
-- ============================================
CREATE INDEX IF NOT EXISTS idx_usage_events_v2_tenant_timestamp
    ON usage_events_v2(tenant_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_usage_events_v2_tenant_event_type
    ON usage_events_v2(tenant_id, event_type, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_usage_events_v2_tenant_ai_model
    ON usage_events_v2(tenant_id, ai_model, timestamp DESC)
    WHERE ai_model IS NOT NULL;

-- ============================================
-- 2. cost_allocation_entries: Add timestamp index for batch deletion cursor
-- ============================================
CREATE INDEX IF NOT EXISTS idx_cost_allocation_entries_timestamp
    ON cost_allocation_entries(timestamp DESC);

-- ============================================
-- 3. invoices: Add ON DELETE SET NULL to subscription_id FK
-- ============================================
-- First, drop the existing FK constraint (if exists) and recreate with proper action
-- This is safe because we're adding ON DELETE SET NULL which preserves invoices
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    -- Find the current constraint name
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'invoices'::regclass
      AND confrelid = 'subscriptions'::regclass
      AND contype = 'f'
    LIMIT 1;

    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE invoices DROP CONSTRAINT %I', constraint_name);
        ALTER TABLE invoices
            ADD CONSTRAINT invoices_subscription_id_fkey
            FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL;
    END IF;
END
$$;

-- ============================================
-- 4. invoices: Add CHECK constraint on currency
-- ============================================
ALTER TABLE invoices
    ADD CONSTRAINT invoices_currency_check
    CHECK (currency ~ '^[A-Z]{3}$' OR currency IS NULL);

-- ============================================
-- 5. usage_rollups_v2: Add missing indexes
-- ============================================
CREATE INDEX IF NOT EXISTS idx_usage_rollups_v2_tenant_period
    ON usage_rollups_v2(tenant_id, period_date DESC);

CREATE INDEX IF NOT EXISTS idx_usage_rollups_v2_tenant_event_type
    ON usage_rollups_v2(tenant_id, event_type, period_date DESC);

-- ============================================
-- 6. pending_usage_charges: Add indexes for dunning processing
-- ============================================
CREATE INDEX IF NOT EXISTS idx_pending_usage_charges_tenant_status
    ON pending_usage_charges(tenant_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_pending_usage_charges_retry
    ON pending_usage_charges(retry_count, next_retry_at)
    WHERE status = 'failed' AND retry_count < 5;

-- ============================================
-- 7. blog_page_views: Fix indexes with IF NOT EXISTS (redo idempotently)
-- ============================================
-- Drop existing indexes if they exist without IF NOT EXISTS
DROP INDEX IF EXISTS idx_blog_page_views_post_id;
DROP INDEX IF EXISTS idx_blog_page_views_viewed_at;
DROP INDEX IF EXISTS idx_blog_page_views_post_viewed;

-- Recreate with IF NOT EXISTS
CREATE INDEX IF NOT EXISTS idx_blog_page_views_post_id ON blog_page_views(post_id);
CREATE INDEX IF NOT EXISTS idx_blog_page_views_viewed_at ON blog_page_views(viewed_at DESC);
CREATE INDEX IF NOT EXISTS idx_blog_page_views_post_viewed ON blog_page_views(post_id, viewed_at DESC);

COMMENT ON INDEX idx_usage_events_v2_tenant_timestamp IS 'Composite index for tenant-scoped usage queries with time ordering';
COMMENT ON INDEX idx_cost_allocation_entries_timestamp IS 'Supports batch deletion cursor in CleanupCostAllocationByRetention';
