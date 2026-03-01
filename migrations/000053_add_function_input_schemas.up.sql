-- Add function input schemas table for input validation
-- This table stores JSON schemas for validating function inputs

CREATE TABLE IF NOT EXISTS function_input_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    schema JSONB NOT NULL, -- JSON Schema for input validation
    is_strict BOOLEAN DEFAULT false, -- Whether to enforce strict schema validation
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_function_input_schema_version UNIQUE (function_version_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_function_input_schemas_function_version ON function_input_schemas(function_version_id);
CREATE INDEX IF NOT EXISTS idx_function_input_schemas_created_at ON function_input_schemas(created_at);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_function_input_schemas_updated_at ON function_input_schemas;
CREATE TRIGGER update_function_input_schemas_updated_at
    BEFORE UPDATE ON function_input_schemas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
