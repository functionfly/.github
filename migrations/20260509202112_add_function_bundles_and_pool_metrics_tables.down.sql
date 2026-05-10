-- Migration: add_function_bundles_and_pool_metrics_tables
-- Created at: 2026-05-09T20:21:12-05:00
-- Purpose: Rollback function_bundles and execution_pool_metrics tables

BEGIN;

DROP TABLE IF EXISTS execution_pool_metrics;
DROP TABLE IF EXISTS function_bundles;

COMMIT;
