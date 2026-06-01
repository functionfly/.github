-- Rollback attempt limits migration
DROP INDEX IF EXISTS idx_cert_exams_user_tier_status;
ALTER TABLE cert_exams DROP COLUMN IF EXISTS attempts;
ALTER TABLE cert_exams DROP COLUMN IF EXISTS max_attempts;
ALTER TABLE cert_exams DROP COLUMN IF EXISTS admin_override_max_attempts;
ALTER TABLE cert_tiers DROP COLUMN IF EXISTS exam_metadata;
