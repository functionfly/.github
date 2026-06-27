-- Migration: 20260626170600_consolidate_state_fabric_plans.down.sql
-- Revert State Fabric Plans Consolidation

BEGIN;

-- Remove sf_active_add_ons column
ALTER TABLE tenants DROP COLUMN IF EXISTS sf_active_add_ons;

-- Drop state_fabric_add_ons table
DROP TABLE IF EXISTS state_fabric_add_ons;

-- Restore original plan constraint (without agent_enterprise)
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_plan_check;

ALTER TABLE tenants ADD CONSTRAINT tenants_plan_check
    CHECK (plan IN ('free', 'starter', 'professional', 'enterprise'));

COMMIT;
