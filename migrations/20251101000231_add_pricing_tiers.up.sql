-- Add pricing tiers support
-- Add plan column to tenants table with default 'starter'
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS plan VARCHAR(20) NOT NULL DEFAULT 'starter' CHECK (plan IN ('starter', 'pro'));

-- Add priority column to backends table for Pro plan priority failover
ALTER TABLE backends ADD COLUMN IF NOT EXISTS priority INTEGER;
