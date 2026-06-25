-- Migration: 20260621170000_city_rankings.down.sql
-- Description: Drop the city rankings schema and users columns added by the up
-- migration. Safe to run only if no other consumers depend on these tables.

DROP TABLE IF EXISTS city_rankings CASCADE;
DROP TABLE IF EXISTS city_aliases CASCADE;
DROP TABLE IF EXISTS cities CASCADE;
DROP TABLE IF EXISTS metro_areas CASCADE;

-- Drop indexes tied to the new users columns before dropping the columns.
DROP INDEX IF EXISTS idx_users_city;
DROP INDEX IF EXISTS idx_users_opted_out;

ALTER TABLE users
  DROP COLUMN IF EXISTS city_id,
  DROP COLUMN IF EXISTS city_ranking_opted_out;
