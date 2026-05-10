-- Migration: add_blocking_columns_to_verification_status
-- Created at: 2026-05-09T17:23:54-05:00
-- Purpose: Add block_reason and blocked_at columns to registry_function_verification_status table

-- Down migration
BEGIN;

ALTER TABLE registry_function_verification_status DROP COLUMN IF EXISTS blocked_at;
ALTER TABLE registry_function_verification_status DROP COLUMN IF EXISTS block_reason;

COMMIT;
