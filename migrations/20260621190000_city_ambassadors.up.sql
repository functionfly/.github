-- Migration: 20260621190000_city_ambassadors.up.sql
-- Description: Create the city_ambassadors table. The top user in each
-- metro (by activity score) gets promoted to ambassador after every
-- recompute cycle. Only one active ambassador per metro at a time.

CREATE TABLE IF NOT EXISTS city_ambassadors (
  id           BIGSERIAL PRIMARY KEY,
  metro_id     BIGINT NOT NULL REFERENCES metro_areas(id) ON DELETE CASCADE,
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  promoted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  revoked_at   TIMESTAMPTZ,
  source       TEXT NOT NULL DEFAULT 'auto',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (metro_id, user_id),
  CONSTRAINT city_ambassadors_source_check
    CHECK (source = ANY (ARRAY['auto','manual']))
);

-- One active ambassador per metro: enforce via a partial unique index.
-- Postgres allows multiple NULL revoked_at? No — but with a partial
-- UNIQUE we exclude revoked rows so we can keep history.
CREATE UNIQUE INDEX IF NOT EXISTS uq_city_ambassadors_active
  ON city_ambassadors (metro_id)
  WHERE revoked_at IS NULL;

-- Fast lookup of "is this user an ambassador anywhere?".
CREATE INDEX IF NOT EXISTS idx_city_ambassadors_user
  ON city_ambassadors (user_id)
  WHERE revoked_at IS NULL;

-- "Who is the current ambassador for this metro?" — covered by the
-- partial unique index above.

-- Drop in reverse: city_ambassadors only.
