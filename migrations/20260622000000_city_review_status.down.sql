-- Migration: 20260622000000_city_review_status.down.sql
-- Description: Remove review workflow columns from cities table

ALTER TABLE cities DROP COLUMN IF EXISTS auto_review_pop_threshold;

DROP INDEX IF EXISTS idx_cities_review_status;

ALTER TABLE cities DROP COLUMN IF EXISTS review_notes,
  DROP COLUMN IF EXISTS reviewed_by,
  DROP COLUMN IF EXISTS reviewed_at,
  DROP COLUMN IF EXISTS review_status;
