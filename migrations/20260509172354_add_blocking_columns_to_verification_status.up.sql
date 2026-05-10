-- Migration: add_blocking_columns_to_verification_status
-- Created at: 2026-05-09T17:23:54-05:00
-- Purpose: Add block_reason and blocked_at columns to registry_function_verification_status table

-- Up migration
BEGIN;

ALTER TABLE registry_function_verification_status ADD COLUMN IF NOT EXISTS block_reason TEXT;
ALTER TABLE registry_function_verification_status ADD COLUMN IF NOT EXISTS blocked_at TIMESTAMP WITH TIME ZONE;

COMMIT;
