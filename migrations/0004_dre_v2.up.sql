-- Migration: Create DRE 2.0 tables for Merkle Execution Graph (MEG) and Execution Certificates
-- This implements the Deterministic Replay Engine 2.0 protocol for cryptographically
-- verifiable function executions.

-- ============================================
-- MERKLE EXECUTION GRAPH RECORDS
-- ============================================

-- Merkle Execution Graph records (one per execution)
CREATE TABLE IF NOT EXISTS execution_meg_records (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id         UUID NOT NULL REFERENCES registry_function_executions(id) ON DELETE CASCADE,
    function_id          UUID NOT NULL,
    version              TEXT NOT NULL,

    -- MEG component hashes
    execution_root_hash  TEXT NOT NULL,
    input_hash           TEXT NOT NULL,
    environment_hash     TEXT NOT NULL,
    dependency_hash      TEXT NOT NULL,
    trace_hash           TEXT,          -- NULL in lite tier
    resource_hash        TEXT NOT NULL,
    output_hash          TEXT NOT NULL,
    metadata_hash        TEXT NOT NULL,

    -- Capsule
    capsule_descriptor_hash TEXT NOT NULL,
    determinism_tier     TEXT NOT NULL DEFAULT 'full',  -- 'full' | 'lite'
    protocol_version     TEXT NOT NULL DEFAULT 'dre/1.0',

    -- Replay state
    replay_root_hash     TEXT,          -- set after successful replay
    replay_verified_at   TIMESTAMPTZ,
    replay_node_id       TEXT,

    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for MEG records
CREATE INDEX IF NOT EXISTS idx_meg_execution_id ON execution_meg_records(execution_id);
CREATE INDEX IF NOT EXISTS idx_meg_function_id ON execution_meg_records(function_id);
CREATE INDEX IF NOT EXISTS idx_meg_root_hash ON execution_meg_records(execution_root_hash);

-- ============================================
-- EXECUTION CERTIFICATES
-- ============================================

-- Execution certificates (.fxcert)
CREATE TABLE IF NOT EXISTS execution_certificates (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    certificate_id       TEXT NOT NULL UNIQUE,          -- "fxc_01H..."
    execution_id         UUID NOT NULL REFERENCES registry_function_executions(id),
    meg_record_id        UUID NOT NULL REFERENCES execution_meg_records(id),
    function_id          UUID NOT NULL,

    cert_level           TEXT NOT NULL DEFAULT 'standard', -- 'lite'|'standard'|'legal_grade'
    cert_json            JSONB NOT NULL,                -- full FXCERT document
    execution_root_hash  TEXT NOT NULL,
    certificate_hash     TEXT NOT NULL,

    -- Signatures
    node_signature       TEXT,
    platform_signature   TEXT,

    -- Blockchain anchoring (optional)
    anchored             BOOLEAN NOT NULL DEFAULT FALSE,
    anchor_chain         TEXT,
    anchor_block_number  BIGINT,
    anchor_tx_hash       TEXT,
    anchor_merkle_root   TEXT,
    anchored_at          TIMESTAMPTZ,

    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for certificates
CREATE INDEX IF NOT EXISTS idx_cert_execution_id ON execution_certificates(execution_id);
CREATE INDEX IF NOT EXISTS idx_cert_function_id ON execution_certificates(function_id);
CREATE INDEX IF NOT EXISTS idx_cert_root_hash ON execution_certificates(execution_root_hash);
CREATE INDEX IF NOT EXISTS idx_cert_certificate_id ON execution_certificates(certificate_id);

-- ============================================
-- DRIFT REPORTS
-- ============================================

-- Drift reports (when replay diverges)
CREATE TABLE IF NOT EXISTS drift_reports (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id         UUID NOT NULL REFERENCES registry_function_executions(id),
    function_id          UUID NOT NULL,
    version              TEXT NOT NULL,

    original_root_hash   TEXT NOT NULL,
    replay_root_hash     TEXT NOT NULL,
    drift_category       TEXT NOT NULL,  -- DriftCategory enum
    component_diff       JSONB,          -- which component hashes differ
    trust_penalty        FLOAT NOT NULL DEFAULT 0,

    detected_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for drift reports
CREATE INDEX IF NOT EXISTS idx_drift_function_id ON drift_reports(function_id);
CREATE INDEX IF NOT EXISTS idx_drift_detected_at ON drift_reports(detected_at);
CREATE INDEX IF NOT EXISTS idx_drift_execution_id ON drift_reports(execution_id);

-- ============================================
-- EXECUTION PASSPORTS
-- ============================================

-- Execution Passport (per-function aggregate)
CREATE TABLE IF NOT EXISTS function_execution_passports (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id                 UUID NOT NULL UNIQUE REFERENCES registry_functions(id),

    -- Determinism stats
    deterministic_reliability   FLOAT NOT NULL DEFAULT 0,  -- 0.0-1.0
    replay_drift_incidents      INT NOT NULL DEFAULT 0,
    verified_executions_total   BIGINT NOT NULL DEFAULT 0,
    total_executions            BIGINT NOT NULL DEFAULT 0,

    -- DRE sub-scores (feed into TrustScore v2)
    determinism_score           FLOAT NOT NULL DEFAULT 0,
    replay_integrity_score      FLOAT NOT NULL DEFAULT 0,
    performance_stability_score FLOAT NOT NULL DEFAULT 0,
    drift_score                 FLOAT NOT NULL DEFAULT 1,  -- 1.0 = no drift

    -- Capsule version history
    capsule_versions_used       JSONB,  -- array of capsule descriptor hashes seen

    -- Last verification timestamp
    last_verified_at           TIMESTAMPTZ,

    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for execution passports
CREATE INDEX IF NOT EXISTS idx_passport_function_id ON function_execution_passports(function_id);

-- ============================================
-- EXTEND REGISTRY_FUNCTION_RATINGS
-- ============================================

-- Extend registry_function_ratings with DRE v2 scores
ALTER TABLE registry_function_ratings
    ADD COLUMN IF NOT EXISTS determinism_score        FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS replay_integrity_score   FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS performance_stability_score FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS drift_score              FLOAT DEFAULT 1,
    ADD COLUMN IF NOT EXISTS trust_score_v2           FLOAT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS trust_v2_updated_at      TIMESTAMPTZ;

-- ============================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================

COMMENT ON TABLE execution_meg_records IS 'Merkle Execution Graph records for deterministic function executions (DRE 2.0)';
COMMENT ON TABLE execution_certificates IS 'FXCERT execution certificates for cryptographically verifiable function executions';
COMMENT ON TABLE drift_reports IS 'Drift reports when replay verification fails (DRE 2.0 anti-manipulation)';
COMMENT ON TABLE function_execution_passports IS 'Public-facing determinism passports for marketplace display';
