-- Add exam attempt limits for certification system
-- Allows tracking and limiting how many times a user can attempt each exam tier

ALTER TABLE cert_exams 
  ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3;

-- Backfill existing exams with default max_attempts
UPDATE cert_exams 
SET max_attempts = 3 
WHERE max_attempts = 0;

-- Index for fast lookups of user+exam+status combinations
CREATE INDEX IF NOT EXISTS idx_cert_exams_user_tier_status 
  ON cert_exams(user_id, tier_id, status);

-- Add exam_metadata to tiers for configurable limits per tier
ALTER TABLE cert_tiers 
  ADD COLUMN IF NOT EXISTS exam_metadata JSONB DEFAULT '{"max_attempts": 3, "allow_retake_after_days": 30}'::jsonb;

-- Allow admins to override per-user attempt limits (for special cases)
ALTER TABLE cert_exams 
  ADD COLUMN IF NOT EXISTS admin_override_max_attempts INTEGER;
