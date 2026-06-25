-- Migration: 20260623220000_fwos_remaining_tables.down.sql
-- Description: Drop FWOS remaining tables (feedback, goals, documents, certs, packages)

DROP TABLE IF EXISTS package_versions;
DROP TABLE IF EXISTS package_registry;
DROP TABLE IF EXISTS org_chart_imports;
DROP TABLE IF EXISTS wallet_pass_templates;
DROP TABLE IF EXISTS certificate_keys;
DROP TABLE IF EXISTS document_signatures;

ALTER TABLE performance_goals
  DROP COLUMN IF EXISTS cascade_visibility,
  DROP COLUMN IF EXISTS goal_level,
  DROP COLUMN IF EXISTS parent_goal_id;

DROP TABLE IF EXISTS feedback_round_responses;
DROP TABLE IF EXISTS feedback_round_assignments;
DROP TABLE IF EXISTS feedback_rounds;
