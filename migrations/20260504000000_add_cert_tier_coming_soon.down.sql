-- Remove is_coming_soon from cert_tiers table
ALTER TABLE cert_tiers DROP COLUMN IF EXISTS is_coming_soon;