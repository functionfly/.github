-- Add is_coming_soon to cert_tiers table
ALTER TABLE cert_tiers ADD COLUMN IF NOT EXISTS is_coming_soon BOOLEAN NOT NULL DEFAULT false;