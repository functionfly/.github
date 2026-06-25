-- Migration: 20260621180000_university_rankings.up.sql
-- Description: Create the university rankings schema (universities,
-- university_aliases, university_rankings) and add the users.university_id
-- column. Plan #5: University Rankings.

-- ============================================================================
-- universities
-- The primary ranking unit for the university leaderboard. Carries the
-- canonical name, country, enrollment, and an "is_active" flag so a
-- university can be retired without losing historical rankings.
-- ============================================================================
CREATE TABLE IF NOT EXISTS universities (
  id              BIGSERIAL PRIMARY KEY,
  slug            TEXT UNIQUE NOT NULL,
  name            TEXT NOT NULL,
  short_name      TEXT,
  country_code    TEXT NOT NULL,
  state_code      TEXT,
  city_id         BIGINT REFERENCES cities(id) ON DELETE SET NULL,
  student_count   INTEGER NOT NULL,
  institution_type TEXT NOT NULL DEFAULT 'university',
  website         TEXT,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT universities_type_check
    CHECK (institution_type = ANY (ARRAY['university','college','bootcamp','institute']))
);

CREATE INDEX IF NOT EXISTS idx_universities_country
  ON universities(country_code) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_universities_active_students
  ON universities(student_count DESC) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_universities_city
  ON universities(city_id);

-- ============================================================================
-- university_aliases
-- Fuzzy match for user-typed "University" / "School" entries. Same shape
-- as city_aliases so the resolver code can be shared.
-- ============================================================================
CREATE TABLE IF NOT EXISTS university_aliases (
  id            BIGSERIAL PRIMARY KEY,
  university_id BIGINT NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
  alias         TEXT NOT NULL,
  source        TEXT NOT NULL,
  UNIQUE(alias, source)
);

CREATE INDEX IF NOT EXISTS idx_university_aliases_alias_lower
  ON university_aliases(LOWER(alias));
CREATE INDEX IF NOT EXISTS idx_university_aliases_university
  ON university_aliases(university_id);

-- ============================================================================
-- users.university_id and users.university_ranking_opted_out
-- Same opt-out model as the city leaderboard.
-- ============================================================================
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS university_id BIGINT REFERENCES universities(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS university_ranking_opted_out BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_users_university
  ON users(university_id) WHERE NOT university_ranking_opted_out;

-- ============================================================================
-- university_rankings
-- Materialized score cache, rewritten by the recompute job. Same UNIQUE
-- shape as city_rankings (one row per university / period_end / category).
-- ============================================================================
CREATE TABLE IF NOT EXISTS university_rankings (
  id                 BIGSERIAL PRIMARY KEY,
  university_id      BIGINT NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
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
  UNIQUE(university_id, period_end, ranking_category),
  CONSTRAINT university_rankings_category_check
    CHECK (ranking_category = ANY (ARRAY['composite','agents','automation','startups','open_source']))
);

CREATE INDEX IF NOT EXISTS idx_university_rankings_category_rank
  ON university_rankings(ranking_category, period_end DESC, rank_position);
CREATE INDEX IF NOT EXISTS idx_university_rankings_university_period
  ON university_rankings(university_id, period_end DESC);
CREATE INDEX IF NOT EXISTS idx_university_rankings_period_rank
  ON university_rankings(period_end DESC, rank_position);
