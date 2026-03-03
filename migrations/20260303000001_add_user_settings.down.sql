-- Migration: Drop user settings column
-- Created: 2026-03-03

-- Remove settings column from users table
ALTER TABLE users DROP COLUMN IF EXISTS settings;
