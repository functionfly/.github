-- Migration: 20260621180000_university_rankings.down.sql
-- Description: Drop the university rankings schema and users columns added by
-- the up migration. Safe to run only if no other consumers depend on these
-- tables.

DROP TABLE IF EXISTS university_rankings CASCADE;
DROP TABLE IF EXISTS university_aliases CASCADE;
DROP TABLE IF EXISTS universities CASCADE;

DROP INDEX IF EXISTS idx_users_university;

ALTER TABLE users
  DROP COLUMN IF EXISTS university_id,
  DROP COLUMN IF EXISTS university_ranking_opted_out;
