-- +migrate Up
-- DRE 2.0: Deterministic Replay Engine Tables
-- This migration creates the core tables for the DRE 2.0 protocol:
-- - MEG records for cryptographic execution verification
-- - Execution certificates (FXCERT) for legal-grade proof
-- - Drift reports for tracking replay divergence
-- - Execution passports for marketplace display
-- - Resource hash history for performance stability tracking

-- ============================================
-- MEG Records Table
-- ============================================
-- Stores the Merkle Execution Graph hash for each deterministic function execution.
-- One record per execution, linked to the execution record.

CREATE TABLE IF NOT EXISTS execution_meg_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES registry_function_executions(id) ON DELETE CASCADE,
    function_id UUID NOT NULL,
    version TEXT NOT NULL,

    -- MEG component hashes (DRE/1.0 leaf ordering — fixed forever)
    execution_root_hash TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    environment_hash TEXT NOT NULL,
    dependency_hash TEXT NOT NULL,
    trace_hash TEXT,  -- NULL in lite tier
    resource_hash TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    metadata_hash TEXT NOT NULL,

    -- Capsule descriptor
    capsule_descriptor_hash TEXT NOT NULL,
    determinism_tier TEXT NOT NULL DEFAULT 'full',  -- 'full' | 'lite'
    protocol_version TEXT NOT NULL DEFAULT 'dre/1.0',

    -- Replay verification state
    replay_root_hash TEXT,
    replay_verified_at TIMESTAMPTZ,
    replay_node_id TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_meg_execution_id ON execution_meg_records(execution_id);
CREATE INDEX idx_meg_function_id ON execution_meg_records(function_id);
CREATE INDEX idx_meg_root_hash ON execution_meg_records(execution_root_hash);
CREATE INDEX idx_meg_created_at ON execution_meg_records(created_at DESC);

-- ============================================
-- Execution Certificates Table
-- ============================================
-- Stores FXCERT execution certificates for legal-grade proof of execution.
-- Certificates can be retrieved publicly for verification.

CREATE TABLE IF NOT EXISTS execution_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    certificate_id TEXT NOT NULL UNIQUE,  -- "fxc_01H..." format
    execution_id UUID NOT NULL REFERENCES registry_function_executions(id),
    meg_record_id UUID NOT NULL REFERENCES execution_meg_records(id),
    function_id UUID NOT NULL,

    cert_level TEXT NOT NULL DEFAULT 'standard',  -- 'lite' | 'standard' | 'legal_grade'
    cert_json JSONB NOT NULL,  -- Full FXCERT document
    execution_root_hash TEXT NOT NULL,
    certificate_hash TEXT NOT NULL,

    -- Signatures
    node_signature TEXT,
    platform_signature TEXT,

    -- Blockchain anchoring (optional)
    anchored BOOLEAN NOT NULL DEFAULT FALSE,
    anchor_chain TEXT,
    anchor_block_number BIGINT,
    anchor_tx_hash TEXT,
    anchor_merkle_root TEXT,
    anchored_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cert_execution_id ON execution_certificates(execution_id);
CREATE INDEX idx_cert_function_id ON execution_certificates(function_id);
CREATE INDEX idx_cert_root_hash ON execution_certificates(execution_root_hash);
CREATE INDEX idx_cert_certificate_id ON execution_certificates(certificate_id);
CREATE INDEX idx_cert_created_at ON execution_certificates(created_at DESC);

-- ============================================
-- Drift Reports Table
-- ============================================
-- Stores divergence reports when replay verification fails.
-- Used for anti-manipulation and trust scoring.

CREATE TABLE IF NOT EXISTS drift_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES registry_function_executions(id),
    function_id UUID NOT NULL,
    version TEXT NOT NULL,

    original_root_hash TEXT NOT NULL,
    replay_root_hash TEXT NOT NULL,
    drift_category TEXT NOT NULL,  -- DriftCategory enum
    component_diff JSONB,  -- Which component hashes differ
    trust_penalty FLOAT NOT NULL DEFAULT 0,

    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_drift_execution_id ON drift_reports(execution_id);
CREATE INDEX idx_drift_function_id ON drift_reports(function_id);
CREATE INDEX idx_drift_detected_at ON drift_reports(detected_at DESC);
CREATE INDEX idx_drift_category ON drift_reports(drift_category);

-- ============================================
-- Execution Passports Table
-- ============================================
-- Per-function aggregate of DRE statistics.
-- This is the public-facing "determinism passport" shown on the marketplace.

CREATE TABLE IF NOT EXISTS function_execution_passports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL UNIQUE REFERENCES registry_functions(id),

    -- Determinism statistics
    deterministic_reliability FLOAT NOT NULL DEFAULT 0,  -- 0.0-1.0
    replay_drift_incidents INT NOT NULL DEFAULT 0,
    verified_executions_total BIGINT NOT NULL DEFAULT 0,
    total_executions BIGINT NOT NULL DEFAULT 0,

    -- DRE sub-scores (feed into TrustScore v2)
    determinism_score FLOAT NOT NULL DEFAULT 0,
    replay_integrity_score FLOAT NOT NULL DEFAULT 0,
    performance_stability_score FLOAT NOT NULL DEFAULT 0,
    drift_score FLOAT NOT NULL DEFAULT 1,

    -- Capsule version history (array of capsule descriptor hashes)
    capsule_versions_used JSONB,

    -- Metadata
    last_verified_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_passport_function_id ON function_execution_passports(function_id);
CREATE INDEX idx_passport_deterministic_rel ON function_execution_passports(deterministic_reliability DESC);
CREATE INDEX idx_passport_trust_scores ON function_execution_passports(determinism_score, replay_integrity_score);

-- ============================================
-- Resource Hash History Table (for Performance Stability)
-- ============================================
-- Stores resource hash history for each function to compute
-- performance stability score over time.

CREATE TABLE IF NOT EXISTS resource_hash_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL,
    resource_hashes JSONB NOT NULL DEFAULT '[]',  -- Array of resource hashes
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_resource_hash_history_function_id ON resource_hash_history(function_id);

-- ============================================
-- Extend registry_function_ratings with DRE v2 fields
-- ============================================

ALTER TABLE registry_function_ratings
    ADD COLUMN IF NOT EXISTS determinism_score FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS replay_integrity_score FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS performance_stability_score FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS drift_score FLOAT DEFAULT 1,
    ADD COLUMN IF NOT EXISTS trust_score_v2 FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS trust_v2_updated_at TIMESTAMPTZ;

-- +migrate Down
-- Rollback DRE 2.0 tables

DROP TABLE IF EXISTS execution_meg_records CASCADE;
DROP TABLE IF EXISTS execution_certificates CASCADE;
DROP TABLE IF EXISTS drift_reports CASCADE;
DROP TABLE IF EXISTS function_execution_passports CASCADE;
DROP TABLE IF EXISTS resource_hash_history CASCADE;

ALTER TABLE registry_function_ratings
    DROP COLUMN IF EXISTS determinism_score,
    DROP COLUMN IF EXISTS replay_integrity_score,
    DROP COLUMN IF EXISTS performance_stability_score,
    DROP COLUMN IF EXISTS drift_score,
    DROP COLUMN IF EXISTS trust_score_v2,
    DROP COLUMN IF EXISTS trust_v2_updated_at;
