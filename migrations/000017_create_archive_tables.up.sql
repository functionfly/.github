-- Create archive_data table for storing compressed and encrypted archive data
CREATE TABLE IF NOT EXISTS archive_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_key VARCHAR(255) UNIQUE NOT NULL,
    archive_type VARCHAR(100) NOT NULL, -- e.g., 'audit_logs', 'user_data', 'compliance_reports'
    compressed_data BYTEA NOT NULL, -- Compressed and encrypted archive data
    encryption_key BYTEA NOT NULL, -- DEK: per-archive key bytes. In production use envelope encryption (store KMS-encrypted DEK or key reference elsewhere).
    metadata_json JSONB NOT NULL, -- Archive metadata as JSON
    checksum VARCHAR(64) NOT NULL, -- SHA256 checksum of original data
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'completed', 'failed', 'deleted'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

    -- Check constraints
    CONSTRAINT chk_archive_status CHECK (status IN ('pending', 'completed', 'failed', 'deleted')),
    CONSTRAINT chk_archive_type CHECK (archive_type IN ('audit_logs', 'user_data', 'compliance_reports', 'system_logs'))
);

-- Indexes for efficient querying
CREATE INDEX idx_archive_data_type ON archive_data (archive_type);
CREATE INDEX idx_archive_data_status ON archive_data (status);
CREATE INDEX idx_archive_data_created_at ON archive_data (created_at);
CREATE INDEX idx_archive_data_storage_key ON archive_data (storage_key);

-- Create archive_cleanup_log table to track cleanup operations
CREATE TABLE IF NOT EXISTS archive_cleanup_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archive_id UUID REFERENCES archive_data(id) ON DELETE CASCADE,
    operation VARCHAR(100) NOT NULL, -- 'cleanup_started', 'cleanup_completed', 'cleanup_failed'
    records_processed INTEGER DEFAULT 0,
    records_deleted INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_archive_cleanup_archive_id ON archive_cleanup_log (archive_id);
CREATE INDEX idx_archive_cleanup_operation ON archive_cleanup_log (operation);
CREATE INDEX idx_archive_cleanup_started_at ON archive_cleanup_log (started_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_archive_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_update_archive_updated_at
    BEFORE UPDATE ON archive_data
    FOR EACH ROW
    EXECUTE FUNCTION update_archive_updated_at();

-- Create index on metadata JSON for efficient querying
CREATE INDEX idx_archive_data_metadata_gin ON archive_data USING GIN (metadata_json);

-- Add comments for documentation
COMMENT ON TABLE archive_data IS 'Stores compressed and encrypted archive data with metadata';
COMMENT ON TABLE archive_cleanup_log IS 'Tracks archive cleanup and maintenance operations';
COMMENT ON COLUMN archive_data.compressed_data IS 'Gzip compressed and AES-256-GCM encrypted data';
COMMENT ON COLUMN archive_data.encryption_key IS 'Data encryption key (DEK). In production use envelope encryption: store KMS-encrypted DEK here or reference key via metadata.encryption_key_id';
COMMENT ON COLUMN archive_data.metadata_json IS 'JSON metadata including record count, date ranges, compression stats';
