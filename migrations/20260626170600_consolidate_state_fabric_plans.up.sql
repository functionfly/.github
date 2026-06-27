-- Migration: 20260626170600_consolidate_state_fabric_plans.up.sql
-- ============================================================================
-- State Fabric Plans Consolidation into Platform Plans
-- ============================================================================
-- SF tiers are now bundled into platform plans:
--   SF Sandbox → Free (1 state fabric object)
--   SF Starter → Starter (3 state fabric objects)
--   SF Pro → Professional (10 state fabric objects)
--   SF Business → Enterprise (unlimited state fabrics)
--   SF Enterprise → Enterprise (unlimited state fabrics)
--
-- This migration:
-- 1. Updates tenants.plan constraint to include 'agent_enterprise'
-- 2. Creates state_fabric_add_ons table for tracking premium add-on subscriptions

BEGIN;

-- ============================================================================
-- Step 1: Update tenants.plan constraint to include agent_enterprise
-- ============================================================================

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_plan_check;

ALTER TABLE tenants ADD CONSTRAINT tenants_plan_check
    CHECK (plan IN ('free', 'starter', 'professional', 'enterprise', 'agent_enterprise'));

-- ============================================================================
-- Step 2: Create state_fabric_add_ons table for tracking premium add-ons
-- SF add-ons are stackable premium features available on any paid plan
-- ============================================================================

CREATE TABLE IF NOT EXISTS state_fabric_add_ons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    add_on_id VARCHAR(100) NOT NULL, -- 'sf_hot_cache', 'sf_multi_region', 'sf_ai_recall', etc.
    stripe_subscription_id VARCHAR(255),
    stripe_price_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'cancelled', 'past_due'
    activated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT valid_add_on_id CHECK (add_on_id IN (
        'sf_hot_cache',
        'sf_multi_region',
        'sf_ai_recall',
        'sf_advanced_insights',
        'sf_advanced_security'
    )),
    CONSTRAINT unique_active_add_on_per_tenant UNIQUE (tenant_id, add_on_id)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_sf_add_ons_tenant_id ON state_fabric_add_ons(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sf_add_ons_status ON state_fabric_add_ons(status);
CREATE INDEX IF NOT EXISTS idx_sf_add_ons_add_on_id ON state_fabric_add_ons(add_on_id);

-- ============================================================================
-- Step 3: Add sf_add_ons column to tenants (for backward compat, tracks active add-ons)
-- This is a denormalized count for quick access
-- ============================================================================

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS sf_active_add_ons TEXT[] DEFAULT '{}';

COMMENT ON COLUMN tenants.sf_active_add_ons IS 'Array of active SF add-on IDs for quick access (denormalized)';

COMMIT;
