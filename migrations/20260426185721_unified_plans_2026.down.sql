-- Migration: 20260426185721_unified_plans_2026.down.sql
-- Description: Rollback unified plans migration - restore separate agent plans

BEGIN;

-- Remove overage rates table
DROP TABLE IF EXISTS overage_rates;

-- Remove plan aliases table
DROP TABLE IF EXISTS plan_aliases;

-- Revert pricing_tiers to old prices
UPDATE pricing_tiers SET price_cents = 2900, annual_price_cents = 27840 WHERE name = 'starter';
UPDATE pricing_tiers SET price_cents = 9900, annual_price_cents = 95040 WHERE name = 'professional';
UPDATE pricing_tiers SET price_cents = 49900, annual_price_cents = 499000, description = 'For large-scale applications and enterprises' WHERE name = 'enterprise';

-- Revert limits
UPDATE pricing_tiers SET max_agents = 2, features = NULL WHERE name = 'starter';
UPDATE pricing_tiers SET max_agents = 10, features = NULL WHERE name = 'professional';
UPDATE pricing_tiers SET max_agents = -1, features = NULL WHERE name = 'enterprise';

-- Remove agent_enterprise tier entry
DELETE FROM pricing_tiers WHERE name = 'agent_enterprise';

-- Revert agent_tier_pricing to old values
UPDATE agent_tier_pricing
SET
    monthly_price_cents = CASE tier_slug
        WHEN 'agent-starter' THEN 2900
        WHEN 'agent-scale' THEN 14900
        WHEN 'agent-pro' THEN 39900
        WHEN 'agent-enterprise' THEN 99000
        ELSE monthly_price_cents
    END,
    display_name = CASE tier_slug
        WHEN 'agent-starter' THEN 'Agent Starter'
        WHEN 'agent-scale' THEN 'Agent Scale'
        WHEN 'agent-pro' THEN 'Agent Pro'
        ELSE display_name
    END
WHERE tier_slug IN ('agent-starter', 'agent-scale', 'agent-pro', 'agent-enterprise');

COMMIT;