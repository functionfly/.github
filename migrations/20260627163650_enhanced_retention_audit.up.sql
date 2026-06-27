-- Migration: Enhanced retention audit fields for compliance
-- Created: 2026-06-27
-- Adds tenant-level tracking and verification hash for tamper-evident audit logs

-- Add tenant_ids column to track which tenants were affected by each cleanup
ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS tenant_ids UUID[];

-- Add archive reference fields for long-term storage of deleted data
ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS archive_path VARCHAR(500);

ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS archive_checksum VARCHAR(64);

-- Add execution metadata for audit trail
ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS execution_id UUID;

ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS parent_audit_id UUID REFERENCES retention_audit_log(id);

-- Add deletion batch sequence for multi-batch operations
ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS batch_sequence INTEGER DEFAULT 1;

-- Add policy metadata for detailed compliance tracking
ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS retention_days INTEGER;

ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS jurisdiction VARCHAR(50) DEFAULT 'US';

-- Add tenant-specific legal hold tracking
ALTER TABLE retention_audit_log
ADD COLUMN IF NOT EXISTS tenant_legal_holds JSONB DEFAULT '[]';

-- Index for efficient tenant-based audit queries
CREATE INDEX IF NOT EXISTS idx_retention_audit_tenant_ids
ON retention_audit_log USING GIN(tenant_ids);

-- Index for verification hash lookups (tamper detection)
CREATE INDEX IF NOT EXISTS idx_retention_audit_verification_hash
ON retention_audit_log(verification_hash);

-- Index for execution chain tracking (parent -> child audit entries)
CREATE INDEX IF NOT EXISTS idx_retention_audit_parent
ON retention_audit_log(parent_audit_id) WHERE parent_audit_id IS NOT NULL;

-- Index for archive reference lookups
CREATE INDEX IF NOT EXISTS idx_retention_audit_archive
ON retention_audit_log(archive_path) WHERE archive_path IS NOT NULL;

-- Comments for documentation
COMMENT ON COLUMN retention_audit_log.tenant_ids IS
    'Array of tenant UUIDs affected by this retention cleanup operation';
COMMENT ON COLUMN retention_audit_log.archive_path IS
    'S3/GCS path to archived data before deletion (for compliance recovery)';
COMMENT ON COLUMN retention_audit_log.archive_checksum IS
    'SHA-256 checksum of archived data for integrity verification';
COMMENT ON COLUMN retention_audit_log.execution_id IS
    'Unique execution ID for this cleanup run (groups all batches)';
COMMENT ON COLUMN retention_audit_log.parent_audit_id IS
    'Reference to parent audit entry for multi-phase cleanup operations';
COMMENT ON COLUMN retention_audit_log.batch_sequence IS
    'Sequence number for batched deletion operations (1, 2, 3...)';
COMMENT ON COLUMN retention_audit_log.retention_days IS
    'Configured retention period in days for this policy';
COMMENT ON COLUMN retention_audit_log.jurisdiction IS
    'Legal jurisdiction governing data retention requirements (SOX, GDPR, etc.)';
COMMENT ON COLUMN retention_audit_log.tenant_legal_holds IS
    'JSON array of tenant-specific legal holds that affected this cleanup';

-- Verify audit log integrity function
CREATE OR REPLACE FUNCTION verify_retention_audit_integrity(p_audit_id UUID)
RETURNS TABLE(field_name VARCHAR, expected_value TEXT, actual_value TEXT, is_valid BOOLEAN) AS $$
DECLARE
    v_record RECORD;
    v_computed_hash VARCHAR;
BEGIN
    SELECT * INTO v_record FROM retention_audit_log WHERE id = p_audit_id;
    IF v_record IS NULL THEN
        RETURN;
    END IF;

    v_computed_hash := encode(
        sha256(
            concat(
                v_record.table_name, '|',
                v_record.retention_policy, '|',
                v_record.cutoff_date::TEXT, '|',
                v_record.records_affected::TEXT, '|',
                v_record.tenant_count::TEXT, '|',
                v_record.financial_impact_cents::TEXT
            )::bytea
        ),
        'hex'
    );

    IF v_record.verification_hash = v_computed_hash THEN
        RETURN QUERY SELECT
            'verification_hash'::VARCHAR as field_name,
            v_computed_hash as expected_value,
            v_record.verification_hash as actual_value,
            TRUE as is_valid;
    ELSE
        RETURN QUERY SELECT
            'verification_hash'::VARCHAR as field_name,
            v_computed_hash as expected_value,
            v_record.verification_hash as actual_value,
            FALSE as is_valid;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to get retention audit summary by tenant
CREATE OR REPLACE FUNCTION get_retention_audit_by_tenant(p_tenant_id UUID)
RETURNS TABLE(
    audit_id UUID,
    table_name VARCHAR,
    retention_policy VARCHAR,
    cutoff_date TIMESTAMPTZ,
    records_affected BIGINT,
    financial_impact_cents BIGINT,
    verification_hash VARCHAR,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ral.id,
        ral.table_name,
        ral.retention_policy,
        ral.cutoff_date,
        ral.records_affected,
        ral.financial_impact_cents,
        ral.verification_hash,
        ral.created_at
    FROM retention_audit_log ral
    WHERE p_tenant_id = ANY(ral.tenant_ids)
    ORDER BY ral.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- Add check constraint for verification hash format
ALTER TABLE retention_audit_log
ADD CONSTRAINT chk_verification_hash_length
CHECK (verification_hash IS NULL OR length(verification_hash) = 64);

-- Add check constraint for batch sequence
ALTER TABLE retention_audit_log
ADD CONSTRAINT chk_batch_sequence_positive
CHECK (batch_sequence IS NULL OR batch_sequence > 0);