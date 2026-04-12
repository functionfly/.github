-- Migration: 20260409000003_populate_pricing_tiers (down)
-- Description: Remove the populated pricing tiers

BEGIN;

-- Delete the pricing tiers we added
DELETE FROM pricing_tiers
WHERE id IN (
    'f47ac10b-58cc-4372-a567-0e02b2c3d479', -- Starter tier
    '550e8400-e29b-41d4-a716-446655440000', -- Professional tier
    '6ba7b810-9dad-11d1-80b4-00c04fd430c8'  -- Free tier
);

COMMIT;
