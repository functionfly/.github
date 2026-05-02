-- Migration: Update all Stripe Price IDs to live values
-- Run this to sync Stripe Price IDs across the system

-- ============================================================================
-- MAIN PLAN PRICE IDS (LIVE)
-- ============================================================================

-- Starter Plan ($24/mo)
-- Monthly: price_1TRjMX4GpTfnozQjrPTg8dnr
-- Annual: price_1TRjMX4GpTfnozQjndh17IEN
UPDATE pricing_tiers SET stripe_price_id = 'price_1TRjMX4GpTfnozQjrPTg8dnr' WHERE name = 'Starter';
UPDATE pricing_tiers SET stripe_price_id_annual = 'price_1TRjMX4GpTfnozQjndh17IEN' WHERE name = 'Starter';
UPDATE pricing_tiers SET price_cents = 2400 WHERE name = 'Starter';
UPDATE pricing_tiers SET annual_price_cents = 24000 WHERE name = 'Starter';

-- Professional Plan ($79/mo)
-- Monthly: price_1TRjMY4GpTfnozQjUpMEEre4
-- Annual: price_1TRjMY4GpTfnozQj2p5NmSsq
UPDATE pricing_tiers SET stripe_price_id = 'price_1TRjMY4GpTfnozQjUpMEEre4' WHERE name = 'Professional';
UPDATE pricing_tiers SET stripe_price_id_annual = 'price_1TRjMY4GpTfnozQj2p5NmSsq' WHERE name = 'Professional';
UPDATE pricing_tiers SET price_cents = 7900 WHERE name = 'Professional';
UPDATE pricing_tiers SET annual_price_cents = 79000 WHERE name = 'Professional';

-- Enterprise Plan ($299/mo) - Add if not exists
INSERT INTO pricing_tiers (name, description, price_cents, currency, features, is_active, tier_type, max_agents, max_functions, max_executions_per_month, billing_cycle, annual_price_cents, stripe_price_id, stripe_price_id_annual)
SELECT 'enterprise', 'For large-scale applications with AI agents', 29900, 'USD', '{"ai_calls_per_month": 10000000}'::jsonb, true, 'subscription', 500, -1, -1, 'monthly', 299000, 'price_1TRjMg4GpTfnozQjaNJRaem3', 'price_1TRjMh4GpTfnozQj3re0rSkx'
WHERE NOT EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'enterprise');

-- Agent Enterprise ($499/mo) - Add if not exists
INSERT INTO pricing_tiers (name, description, price_cents, currency, features, is_active, tier_type, max_agents, max_functions, max_executions_per_month, billing_cycle, annual_price_cents, stripe_price_id, stripe_price_id_annual)
SELECT 'agent_enterprise', 'Unlimited AI agents for enterprise scale', 49900, 'USD', '{"unlimited": true}'::jsonb, true, 'subscription', -1, -1, -1, 'monthly', 499000, 'price_1TRjMh4GpTfnozQjLe8VpO2o', 'price_1TRjMh4GpTfnozQj0pPeldsb'
WHERE NOT EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'agent_enterprise');

-- ============================================================================
-- BACKEND-IN-A-BOX BUNDLE PRICE IDS (LIVE)
-- ============================================================================

-- SaaS Starter Pack ($29/mo): price_1TRjFz4GpTfnozQjrOx1eJMz
-- Marketplace Pack ($49/mo): price_1TRjGA4GpTfnozQjwMLI9dfF
-- AI App Pack ($39/mo): price_1TRjGM4GpTfnozQjU554K57I
UPDATE pricing_bundles SET stripe_price_id = 'price_1TRjFz4GpTfnozQjrOx1eJMz' WHERE slug = 'saas-starter';
UPDATE pricing_bundles SET stripe_price_id = 'price_1TRjGA4GpTfnozQjwMLI9dfF' WHERE slug = 'marketplace';
UPDATE pricing_bundles SET stripe_price_id = 'price_1TRjGM4GpTfnozQjU554K57I' WHERE slug = 'ai-app';

-- ============================================================================
-- STATE FABRIC PRICE IDS (LIVE)
-- ============================================================================

-- State Fabric Starter ($19/mo): price_1TRjJX4GpTfnozQjtQ2fRosA
-- State Fabric Pro ($99/mo): price_1TRjJX4GpTfnozQjZUIPr46U
-- State Fabric Business ($499/mo): price_1TRjJY4GpTfnozQj6TZJ40WN
-- Hot Cache Booster ($49/mo): price_1TRjJg4GpTfnozQjvDKCN6oO
-- Advanced Security Pack ($99/mo): price_1TRjJg4GpTfnozQj6ljWGcTJ
-- AI Memory Pack ($149/mo): price_1TRjJh4GpTfnozQjl94nX5ZN
-- Advanced Insights ($79/mo): price_1TRjJh4GpTfnozQjn53ONtd3

-- ============================================================================
-- TRUST API PRICE IDS (LIVE)
-- ============================================================================

-- Trust API Startup ($49/mo): price_1TRjJP4GpTfnozQjEY0Y6USJ
-- Trust API Business ($199/mo): price_1TRjJP4GpTfnozQjEq9DpWNb
-- Trust API Overage Startup ($5/unit): price_1TRjJQ4GpTfnozQjxlIMLNSf
-- Trust API Overage Business ($3/unit): price_1TRjJP4GpTfnozQjlLnbKcD7
