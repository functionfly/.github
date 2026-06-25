-- Migration: 20260621170000_city_rankings.up.sql
-- Description: Create city rankings schema (metro_areas, cities, city_aliases,
-- city_rankings) and the users columns required to attribute activity to a city.
-- Created: 2026-06-21
-- Plan: .kilo/plans/1782018734195-city-rankings-plan.md

-- ============================================================================
-- metro_areas
-- Primary ranking aggregation unit. MSAs (US) or equivalent metros worldwide.
-- ============================================================================
CREATE TABLE IF NOT EXISTS metro_areas (
  id            BIGSERIAL PRIMARY KEY,
  slug          TEXT UNIQUE NOT NULL,
  name          TEXT NOT NULL,
  country_code  TEXT NOT NULL,
  population    INTEGER NOT NULL,
  latitude      NUMERIC(9,6),
  longitude     NUMERIC(9,6),
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metro_areas_country
  ON metro_areas(country_code) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_metro_areas_active_pop
  ON metro_areas(population DESC) WHERE is_active = TRUE;

-- ============================================================================
-- cities
-- Individual city records. FK to metro_areas (nullable for cities without a
-- metro in the seed data). The canonical self-reported "Location" string.
-- ============================================================================
CREATE TABLE IF NOT EXISTS cities (
  id            BIGSERIAL PRIMARY KEY,
  slug          TEXT UNIQUE NOT NULL,
  name          TEXT NOT NULL,
  state_code    TEXT NOT NULL,
  state_name    TEXT NOT NULL,
  country_code  TEXT NOT NULL,
  country_name  TEXT NOT NULL,
  latitude      NUMERIC(9,6) NOT NULL,
  longitude     NUMERIC(9,6) NOT NULL,
  population    INTEGER,
  metro_area_id BIGINT REFERENCES metro_areas(id) ON DELETE SET NULL,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cities_country_state
  ON cities(country_code, state_code);
CREATE INDEX IF NOT EXISTS idx_cities_metro
  ON cities(metro_area_id);
CREATE INDEX IF NOT EXISTS idx_cities_slug
  ON cities(slug);
CREATE INDEX IF NOT EXISTS idx_cities_name_lower
  ON cities(LOWER(name));

-- ============================================================================
-- city_aliases
-- Fuzzy-match lookup for user-typed "Location" strings. Source distinguishes
-- "seed" (from CSV) vs "user_input" vs "manual".
-- ============================================================================
CREATE TABLE IF NOT EXISTS city_aliases (
  id      BIGSERIAL PRIMARY KEY,
  city_id BIGINT NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
  alias   TEXT NOT NULL,
  source  TEXT NOT NULL,
  UNIQUE(alias, source)
);

CREATE INDEX IF NOT EXISTS idx_city_aliases_alias_lower
  ON city_aliases(LOWER(alias));
CREATE INDEX IF NOT EXISTS idx_city_aliases_city
  ON city_aliases(city_id);

-- ============================================================================
-- users.city_id and users.city_ranking_opted_out
-- Opt-out model: users are counted by default. Setting the column to TRUE
-- excludes the user from every aggregation.
-- ============================================================================
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS city_id BIGINT REFERENCES cities(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS city_ranking_opted_out BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_users_city
  ON users(city_id) WHERE NOT city_ranking_opted_out;
CREATE INDEX IF NOT EXISTS idx_users_opted_out
  ON users(city_ranking_opted_out);

-- ============================================================================
-- city_rankings
-- Materialized score cache, rewritten by the hourly recompute job. One row
-- per (metro, period_end, category). ranking_category is a CHECK-constrained
-- enum so invalid categories fail at the DB level.
-- ============================================================================
CREATE TABLE IF NOT EXISTS city_rankings (
  id                 BIGSERIAL PRIMARY KEY,
  metro_area_id      BIGINT NOT NULL REFERENCES metro_areas(id) ON DELETE CASCADE,
  rank_position      INTEGER NOT NULL,
  prev_rank_position INTEGER,
  score_raw          NUMERIC(20,6) NOT NULL,
  score_per_capita   NUMERIC(20,6) NOT NULL,
  active_users       INTEGER NOT NULL,
  deployments        INTEGER NOT NULL,
  executions_30d     BIGINT NOT NULL,
  founder_earnings   BIGINT NOT NULL,
  new_users_30d      INTEGER NOT NULL,
  period_start       TIMESTAMPTZ NOT NULL,
  period_end         TIMESTAMPTZ NOT NULL,
  computed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ranking_category   TEXT NOT NULL DEFAULT 'composite',
  UNIQUE(metro_area_id, period_end, ranking_category),
  CONSTRAINT city_rankings_category_check
    CHECK (ranking_category = ANY (ARRAY['composite','agents','automation','startups','open_source','robotics']))
);

CREATE INDEX IF NOT EXISTS idx_city_rankings_category_rank
  ON city_rankings(ranking_category, period_end DESC, rank_position);
CREATE INDEX IF NOT EXISTS idx_city_rankings_metro_period
  ON city_rankings(metro_area_id, period_end DESC);
CREATE INDEX IF NOT EXISTS idx_city_rankings_period_rank
  ON city_rankings(period_end DESC, rank_position);
