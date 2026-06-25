-- Migration: 20260621234014_company_rankings.up.sql
-- Description: Create company rankings schema for Top Agent Companies leaderboard.
-- Companies are businesses on FunctionFly, ranked by their aggregate activity.
-- This mirrors the city/university ranking pattern with per-capita scoring.

-- ============================================================================
-- companies
-- Primary ranking unit. Companies (tenants) on the platform.
-- ============================================================================
CREATE TABLE IF NOT EXISTS companies (
  id            BIGSERIAL PRIMARY KEY,
  slug          TEXT UNIQUE NOT NULL,
  name          TEXT NOT NULL,
  city_id       BIGINT REFERENCES cities(id) ON DELETE SET NULL,
  country_code  TEXT NOT NULL,
  employee_count INTEGER,
  industry      TEXT,
  website       TEXT,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_companies_country ON companies(country_code) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_companies_city ON companies(city_id) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_companies_slug ON companies(slug);

-- Aliases for fuzzy matching company names
CREATE TABLE IF NOT EXISTS company_aliases (
  id         BIGSERIAL PRIMARY KEY,
  company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  alias      TEXT NOT NULL,
  source     TEXT NOT NULL,
  UNIQUE(alias, source)
);

CREATE INDEX IF NOT EXISTS idx_company_aliases_alias_lower ON company_aliases(LOWER(alias));
CREATE INDEX IF NOT EXISTS idx_company_aliases_company ON company_aliases(company_id);

-- ============================================================================
-- company_rankings
-- Materialized score rows. UNIQUE(company_id, period_end, ranking_category).
-- Rankings are per-capita: company employees as the denominator.
-- ============================================================================
CREATE TABLE IF NOT EXISTS company_rankings (
  id                  BIGSERIAL PRIMARY KEY,
  company_id          BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  rank_position       INTEGER NOT NULL,
  prev_rank_position  INTEGER,
  score_raw           NUMERIC(20,6) NOT NULL,
  score_per_capita    NUMERIC(20,6) NOT NULL,
  active_users        INTEGER NOT NULL,
  deployments        INTEGER NOT NULL,
  executions_30d     BIGINT NOT NULL,
  revenue_cents      BIGINT NOT NULL DEFAULT 0,
  new_users_30d      INTEGER NOT NULL,
  period_start       TIMESTAMPTZ NOT NULL,
  period_end         TIMESTAMPTZ NOT NULL,
  computed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ranking_category   TEXT NOT NULL DEFAULT 'composite',
  CONSTRAINT company_rankings_category_check
    CHECK (ranking_category = ANY (ARRAY['composite'::text, 'agents'::text, 'automation'::text, 'startups'::text, 'open_source'::text, 'robotics'::text])),
  UNIQUE(company_id, period_end, ranking_category)
);

CREATE INDEX IF NOT EXISTS idx_company_rankings_period_rank ON company_rankings(period_end DESC, rank_position);
CREATE INDEX IF NOT EXISTS idx_company_rankings_company_period ON company_rankings(company_id, period_end DESC);
CREATE INDEX IF NOT EXISTS idx_company_rankings_category_rank ON company_rankings(ranking_category, period_end DESC, rank_position);

-- Add company_id to tenants (links a tenant to a company record)
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS company_id BIGINT REFERENCES companies(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_company ON tenants(company_id) WHERE company_id IS NOT NULL;
