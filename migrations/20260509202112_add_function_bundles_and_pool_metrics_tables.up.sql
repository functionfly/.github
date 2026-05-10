-- Migration: add_function_bundles_and_pool_metrics_tables
-- Created at: 2026-05-09T20:21:12-05:00
-- Purpose: Add function_bundles and execution_pool_metrics tables for 2026 execution architecture

BEGIN;

-- Table: function_bundles
-- Stores pre-compiled WASM / JS bundles for fast retrieval at execution time.
CREATE TABLE IF NOT EXISTS function_bundles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    runtime VARCHAR(50) NOT NULL,
    bundle_hash VARCHAR(64) NOT NULL, -- SHA-256 of source_code
    wasm_binary BYTEA,               -- NULL for JS/Deno (stored in S3/R2)
    compiled_size_bytes INT,
    compilation_duration_ms INT,
    compiled_at TIMESTAMPTZ DEFAULT now(),
    is_valid BOOLEAN DEFAULT true,
    UNIQUE(function_version_id, runtime, bundle_hash)
);

CREATE INDEX IF NOT EXISTS idx_function_bundles_lookup
    ON function_bundles(function_version_id, runtime, bundle_hash);

-- Table: execution_pool_metrics
-- Tracks pool health for autoscaling and observability.
CREATE TABLE IF NOT EXISTS execution_pool_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime VARCHAR(50) NOT NULL,
    pool_size INT NOT NULL,
    warm_instances INT NOT NULL,
    cold_starts_1m INT NOT NULL,
    avg_latency_ms INT,
    recorded_at TIMESTAMPTZ DEFAULT now()
);

COMMIT;
