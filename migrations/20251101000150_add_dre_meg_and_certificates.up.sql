-- DRE 2.0: execution_meg_records, execution_certificates, drift_reports, function_execution_passports
-- Required for GET /registry/{author}/{name}/executions (list MEG records) and related DRE APIs.

CREATE TABLE IF NOT EXISTS execution_meg_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    version VARCHAR(20) NOT NULL,

    execution_root_hash TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    environment_hash TEXT NOT NULL,
    dependency_hash TEXT NOT NULL,
    trace_hash TEXT,
    resource_hash TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    metadata_hash TEXT NOT NULL,

    capsule_descriptor_hash TEXT NOT NULL,
    determinism_tier TEXT NOT NULL DEFAULT 'full',
    protocol_version TEXT NOT NULL DEFAULT 'dre/1.0',

    replay_root_hash TEXT,
    replay_verified_at TIMESTAMP WITH TIME ZONE,
    replay_node_id TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_execution_meg_records_function_id ON execution_meg_records(function_id);
CREATE INDEX IF NOT EXISTS idx_execution_meg_records_execution_id ON execution_meg_records(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_meg_records_execution_root_hash ON execution_meg_records(execution_root_hash);
CREATE INDEX IF NOT EXISTS idx_execution_meg_records_created_at ON execution_meg_records(created_at DESC);

CREATE TABLE IF NOT EXISTS execution_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    certificate_id TEXT NOT NULL UNIQUE,
    execution_id UUID NOT NULL,
    meg_record_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    cert_level TEXT NOT NULL DEFAULT 'standard',
    cert_json JSONB NOT NULL,
    execution_root_hash TEXT NOT NULL,
    certificate_hash TEXT NOT NULL,

    node_signature TEXT,
    platform_signature TEXT,

    anchored BOOLEAN NOT NULL DEFAULT false,
    anchor_chain TEXT,
    anchor_block_number BIGINT,
    anchor_tx_hash TEXT,
    anchor_merkle_root TEXT,
    anchored_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_execution_certificates_function_id ON execution_certificates(function_id);
CREATE INDEX IF NOT EXISTS idx_execution_certificates_execution_id ON execution_certificates(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_certificates_execution_root_hash ON execution_certificates(execution_root_hash);

CREATE TABLE IF NOT EXISTS drift_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    version VARCHAR(20) NOT NULL,

    original_root_hash TEXT NOT NULL,
    replay_root_hash TEXT NOT NULL,
    drift_category TEXT NOT NULL,
    component_diff JSONB,
    trust_penalty DOUBLE PRECISION NOT NULL DEFAULT 0,

    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_drift_reports_function_id ON drift_reports(function_id);
CREATE INDEX IF NOT EXISTS idx_drift_reports_detected_at ON drift_reports(detected_at);

CREATE TABLE IF NOT EXISTS function_execution_passports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL UNIQUE REFERENCES registry_functions(id) ON DELETE CASCADE,

    deterministic_reliability DOUBLE PRECISION NOT NULL DEFAULT 0,
    replay_drift_incidents INTEGER NOT NULL DEFAULT 0,
    verified_executions_total BIGINT NOT NULL DEFAULT 0,
    total_executions BIGINT NOT NULL DEFAULT 0,

    determinism_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    replay_integrity_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    performance_stability_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    drift_score DOUBLE PRECISION NOT NULL DEFAULT 1,

    capsule_versions_used JSONB,
    last_verified_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_function_execution_passports_function_id ON function_execution_passports(function_id);
