-- Migration: Playground Templates
-- Description: Adds support for storing reusable input templates per function in the database
-- This enables users to save and share input templates for playground testing

CREATE TABLE IF NOT EXISTS playground_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    input_schema JSONB,
    input_example JSONB NOT NULL DEFAULT '{}',
    is_shared BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_function FOREIGN KEY (function_id) REFERENCES registry_functions(id) ON DELETE CASCADE
);

CREATE INDEX idx_playground_templates_tenant ON playground_templates(tenant_id);
CREATE INDEX idx_playground_templates_function ON playground_templates(function_id);
CREATE INDEX idx_playground_templates_created ON playground_templates(created_at DESC);

COMMENT ON TABLE playground_templates IS 'Stores reusable input templates for function playground testing';
COMMENT ON COLUMN playground_templates.input_schema IS 'JSON schema defining expected input fields';
COMMENT ON COLUMN playground_templates.input_example IS 'Example JSON input that conforms to the schema';
COMMENT ON COLUMN playground_templates.is_shared IS 'Whether this template is shared across the tenant';