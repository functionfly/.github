-- Migration: 20260621200000_city_wars.up.sql
-- Description: Create the city_wars and city_war_matches tables for the
-- quarterly "City Wars" bracket (plan §8 #8). A war is a single-elimination
-- bracket of 8 metros paired by current rank. Each match is a row in
-- city_war_matches; the winner advances to the next round until one
-- champion remains.

CREATE TABLE IF NOT EXISTS city_wars (
  id             BIGSERIAL PRIMARY KEY,
  slug           TEXT UNIQUE NOT NULL,
  name           TEXT NOT NULL,
  season         TEXT NOT NULL,             -- "2026-Q3" / "2026-Q4" / "2027-Q1"
  status         TEXT NOT NULL DEFAULT 'scheduled',
  round          TEXT NOT NULL DEFAULT 'quarterfinal',  -- 'quarterfinal' | 'semifinal' | 'final' | 'complete'
  starts_at      TIMESTAMPTZ NOT NULL,
  ends_at        TIMESTAMPTZ NOT NULL,
  champion_metro_id BIGINT REFERENCES metro_areas(id) ON DELETE SET NULL,
  total_matches  INTEGER NOT NULL DEFAULT 0,
  total_active_users INTEGER NOT NULL DEFAULT 0,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT city_wars_status_check
    CHECK (status = ANY (ARRAY['scheduled','active','complete','cancelled'])),
  CONSTRAINT city_wars_round_check
    CHECK (round = ANY (ARRAY['quarterfinal','semifinal','final','complete']))
);

CREATE INDEX IF NOT EXISTS idx_city_wars_status ON city_wars (status, starts_at DESC);
CREATE INDEX IF NOT EXISTS idx_city_wars_season ON city_wars (season);

-- One war per (season, round) is enforced implicitly: a new war starts a
-- new row with the same season but only one row at each round. The UI
-- reads "active" wars and "complete" wars separately.

CREATE TABLE IF NOT EXISTS city_war_matches (
  id               BIGSERIAL PRIMARY KEY,
  war_id           BIGINT NOT NULL REFERENCES city_wars(id) ON DELETE CASCADE,
  round            TEXT NOT NULL,
  position         INTEGER NOT NULL,         -- 1..4 (quarterfinal), 1..2 (semifinal), 1 (final)
  metro_a_id       BIGINT NOT NULL REFERENCES metro_areas(id) ON DELETE CASCADE,
  metro_b_id       BIGINT NOT NULL REFERENCES metro_areas(id) ON DELETE CASCADE,
  score_a          NUMERIC(20,6) NOT NULL DEFAULT 0,  -- per-capita at decision time
  score_b          NUMERIC(20,6) NOT NULL DEFAULT 0,
  active_users_a   INTEGER NOT NULL DEFAULT 0,
  active_users_b   INTEGER NOT NULL DEFAULT 0,
  winner_metro_id  BIGINT REFERENCES metro_areas(id) ON DELETE SET NULL,  -- NULL = pending
  decided_at       TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT city_war_matches_round_check
    CHECK (round = ANY (ARRAY['quarterfinal','semifinal','final'])),
  CONSTRAINT city_war_matches_different_metros
    CHECK (metro_a_id <> metro_b_id),
  CONSTRAINT city_war_matches_position_range
    CHECK (position >= 1 AND position <= 4),
  UNIQUE (war_id, round, position)
);

CREATE INDEX IF NOT EXISTS idx_city_war_matches_war ON city_war_matches (war_id, round, position);
CREATE INDEX IF NOT EXISTS idx_city_war_matches_metro_a ON city_war_matches (metro_a_id);
CREATE INDEX IF NOT EXISTS idx_city_war_matches_metro_b ON city_war_matches (metro_b_id);

-- Down:
-- DROP TABLE IF EXISTS city_war_matches CASCADE;
-- DROP TABLE IF EXISTS city_wars CASCADE;
