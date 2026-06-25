-- Migration: 20260622000000_city_review_status.up.sql
-- Description: Add review_status to cities table for admin review workflow
--   - 'seed'       : from CSV seed (pre-approved)
--   - 'approved'   : manually approved city (admin or from geocode when population < threshold)
--   - 'pending'    : awaiting admin review (geocoded city with high population)
--   - 'rejected'   : rejected by admin review
-- Also adds reviewed_at, reviewed_by, and review_notes columns for audit trail.

ALTER TABLE cities
  ADD COLUMN IF NOT EXISTS review_status TEXT NOT NULL DEFAULT 'seed',
  ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS reviewed_by UUID,
  ADD COLUMN IF NOT EXISTS review_notes TEXT;

CREATE INDEX IF NOT EXISTS idx_cities_review_status
  ON cities(review_status) WHERE review_status != 'approved';

-- Backfill existing geocoded cities (those with source = 'user_geocode' in aliases) as 'approved'
-- since they've been in production use
UPDATE cities c
SET review_status = 'approved'
FROM city_aliases a
WHERE a.city_id = c.id AND a.source = 'user_geocode' AND c.review_status = 'seed';

-- FunctionFly: High-population threshold for automatic approval
-- Cities with population below this threshold are auto-approved on geocode
-- Cities at or above this threshold require admin review
ALTER TABLE cities
  ADD COLUMN IF NOT EXISTS auto_review_pop_threshold INTEGER NOT NULL DEFAULT 100000;
