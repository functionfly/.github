-- Database encryption support
-- +migrate Up

-- Create encryption keys table for key management
CREATE TABLE IF NOT EXISTS encryption_keys (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(50) NOT NULL,
    encrypted_key TEXT NOT NULL,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
    purpose VARCHAR(20) NOT NULL CHECK (purpose IN ('master', 'data', 'field')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    active BOOLEAN DEFAULT TRUE,
    rotated_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(version, purpose)
);

-- Create encrypted fields table to track which fields are encrypted
CREATE TABLE IF NOT EXISTS encrypted_fields (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL,
    column_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(50) NOT NULL,
    encryption_key_version VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(table_name, column_name)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_encryption_keys_version_purpose ON encryption_keys(version, purpose);
CREATE INDEX IF NOT EXISTS idx_encryption_keys_active ON encryption_keys(active);
CREATE INDEX IF NOT EXISTS idx_encrypted_fields_table_column ON encrypted_fields(table_name, column_name);

-- Add comments for documentation
COMMENT ON TABLE encryption_keys IS 'Stores encrypted encryption keys for database encryption at rest';
COMMENT ON TABLE encrypted_fields IS 'Tracks which database fields are encrypted and with which key version';
COMMENT ON COLUMN encryption_keys.purpose IS 'Key purpose: master (encrypts other keys), data (encrypts data), field (encrypts specific fields)';

-- Grant permissions to application user (if different from owner)
-- GRANT SELECT, INSERT, UPDATE ON encryption_keys TO functionfly_app;
-- GRANT SELECT, INSERT, UPDATE ON encrypted_fields TO functionfly_app;