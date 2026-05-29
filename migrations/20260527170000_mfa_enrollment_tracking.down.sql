-- Migration: Remove MFA enrollment tracking tables

DROP TRIGGER IF EXISTS trg_mfa_enrollments_updated_at ON mfa_enrollments;
DROP FUNCTION IF EXISTS update_mfa_enrollment_updated_at();
DROP TABLE IF EXISTS mfa_backup_codes;
DROP TABLE IF EXISTS mfa_enrollments;