-- Migration: execution_receipt
-- Created at: 2026-06-01T20:35:00
-- Purpose: Adds the "Execution Receipt" feature — a public, shareable, denormalized
--          view of every successful function execution. Extends registry_executions_public
--          with function metadata, schemas, view/fork counters, and adds a milestone
--          events table for the "your function just ran 100x" notification fan-out.

-- ============================================================================
-- Up migration
-- ============================================================================
BEGIN;

-- ---------------------------------------------------------------------------
-- A. Extend the existing shareable-execution table with denormalized fields
--    so the public /v1/receipts/:id read path needs zero joins.
-- ---------------------------------------------------------------------------
ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS function_name       TEXT,
  ADD COLUMN IF NOT EXISTS function_author     TEXT,
  ADD COLUMN IF NOT EXISTS runtime             TEXT,
  ADD COLUMN IF NOT EXISTS input_schema        JSONB,
  ADD COLUMN IF NOT EXISTS output_schema       JSONB,
  ADD COLUMN IF NOT EXISTS function_visibility TEXT DEFAULT 'public',
  ADD COLUMN IF NOT EXISTS description         TEXT,
  ADD COLUMN IF NOT EXISTS fork_count          INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS view_count          INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_viewed_at      TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS revoked_at          TIMESTAMPTZ;

-- ---------------------------------------------------------------------------
-- B. Backfill from registry_functions / registry_function_versions.
--    Uses a CTE to avoid the "ep cannot be referenced from this part of the
--    query" issue when joining registry_function_versions inside FROM.
-- ---------------------------------------------------------------------------
WITH backfill AS (
  SELECT
    ep.public_id,
    f.name          AS bf_name,
    f.author        AS bf_author,
    v.runtime       AS bf_runtime,
    f.visibility    AS bf_visibility,
    f.description   AS bf_description
  FROM registry_executions_public ep
  JOIN registry_functions f
    ON f.id = ep.function_id
  LEFT JOIN registry_function_versions v
    ON v.function_id = f.id AND v.version = ep.version
  WHERE ep.function_name IS NULL OR ep.runtime IS NULL
)
UPDATE registry_executions_public ep
SET
  function_name       = COALESCE(ep.function_name, bf.bf_name),
  function_author     = COALESCE(ep.function_author, bf.bf_author),
  runtime             = COALESCE(ep.runtime, bf.bf_runtime),
  function_visibility = COALESCE(ep.function_visibility, bf.bf_visibility, 'public'),
  description         = COALESCE(ep.description, bf.bf_description)
FROM backfill bf
WHERE ep.public_id = bf.public_id;

-- ---------------------------------------------------------------------------
-- C. Enforce NOT NULL on the columns every receipt row must have.
--    Safe because the backfill above covers every existing row.
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  -- Only set NOT NULL if no NULLs remain (defensive — handles the
  -- edge case of a function row that no longer exists).
  IF NOT EXISTS (
    SELECT 1 FROM registry_executions_public
    WHERE function_name IS NULL OR function_author IS NULL OR runtime IS NULL
  ) THEN
    ALTER TABLE registry_executions_public
      ALTER COLUMN function_name   SET NOT NULL,
      ALTER COLUMN function_author SET NOT NULL,
      ALTER COLUMN runtime         SET NOT NULL;
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- D. Indexes for the two hot public read paths
-- ---------------------------------------------------------------------------

-- (a) Single-receipt lookup (the hottest public read; partial so we only
--     index shareable rows).
CREATE INDEX IF NOT EXISTS idx_rcpt_public_id_active
  ON registry_executions_public (public_id)
  WHERE shareable = TRUE AND revoked_at IS NULL;

-- (b) Per-function "all receipts for this function" list (drives the
--     trending widget and the function-page embed strip).
CREATE INDEX IF NOT EXISTS idx_rcpt_function_created
  ON registry_executions_public (function_id, created_at DESC)
  WHERE shareable = TRUE;

-- (c) Trending receipts — top-viewed, recent first. Used by the
--     /v1/receipts/trending endpoint that powers the marketing site
--     and the dashboard "trending" widget.
CREATE INDEX IF NOT EXISTS idx_rcpt_trending
  ON registry_executions_public (view_count DESC, created_at DESC)
  WHERE shareable = TRUE AND revoked_at IS NULL;

-- (d) Milestone detection — fast lookup of "has this function been
--     counted at all", used by the sweep scheduler.
CREATE INDEX IF NOT EXISTS idx_rcpt_function_total
  ON registry_executions_public (function_id);

-- ---------------------------------------------------------------------------
-- E. New table: per-function, per-threshold milestone log
--    Idempotent fan-out: dedupe_key UNIQUE means INSERT ON CONFLICT
--    DO NOTHING is safe to retry from any worker.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS receipt_milestone_events (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  function_id       UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
  tenant_id         UUID,
  threshold         INTEGER NOT NULL CHECK (threshold > 0),
  total_runs_at     INTEGER NOT NULL CHECK (total_runs_at > 0),
  public_id         TEXT NOT NULL,
  fired_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  channels_fired    TEXT[] NOT NULL DEFAULT '{}',
  dedupe_key        TEXT NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT receipt_milestone_dedupe_key_unique UNIQUE (dedupe_key)
);

CREATE INDEX IF NOT EXISTS idx_milestone_function
  ON receipt_milestone_events (function_id, fired_at DESC);

CREATE INDEX IF NOT EXISTS idx_milestone_threshold
  ON receipt_milestone_events (threshold, fired_at DESC);

-- ---------------------------------------------------------------------------
-- F. User opt-out column for milestone notifications
-- ---------------------------------------------------------------------------
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS receipt_milestones_enabled BOOLEAN NOT NULL DEFAULT TRUE;

-- ---------------------------------------------------------------------------
-- G. Owner revocation audit (for analytics on revocations)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS receipt_revocations (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  public_id       TEXT NOT NULL,
  function_id     UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
  revoked_by      UUID NOT NULL,
  revoked_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  reason          TEXT
);

CREATE INDEX IF NOT EXISTS idx_revocations_function
  ON receipt_revocations (function_id, revoked_at DESC);

COMMIT;
