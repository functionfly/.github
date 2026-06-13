-- FOM: Function Outcome Model Database
-- Down Migration: 001_fom_initial.down.sql
-- Date: 2026-06-12

DROP TABLE IF EXISTS fom_privacy_budget;
DROP TABLE IF EXISTS fom_workflow_hints;
DROP TABLE IF EXISTS fom_events;
DROP TABLE IF EXISTS fom_training_records;
DROP TABLE IF EXISTS fom_function_stats;
DROP TABLE IF EXISTS fom_workflow_patterns;
DROP TABLE IF EXISTS fom_results;
DROP TABLE IF EXISTS fom_actions;
DROP TABLE IF EXISTS fom_plans;
DROP TABLE IF EXISTS fom_goals;
DROP TABLE IF EXISTS fom_failure_types;