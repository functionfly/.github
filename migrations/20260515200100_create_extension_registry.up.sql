-- Create extension_registry table for studio extensions

CREATE TABLE IF NOT EXISTS extension_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    author_name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'custom',
    status VARCHAR(50) NOT NULL DEFAULT 'disabled',
    permissions JSONB DEFAULT '[]',
    hooks JSONB DEFAULT '[]',
    size_kb INTEGER DEFAULT 0,
    config JSONB DEFAULT '{}',
    installed_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    enabled_at TIMESTAMPTZ,
    error_message TEXT,
    UNIQUE(tenant_id, name, version)
);

-- Indexes for extension lookups
CREATE INDEX IF NOT EXISTS idx_extension_registry_tenant ON extension_registry(tenant_id);
CREATE INDEX IF NOT EXISTS idx_extension_registry_name ON extension_registry(name);
CREATE INDEX IF NOT EXISTS idx_extension_registry_status ON extension_registry(status);
CREATE INDEX IF NOT EXISTS idx_extension_registry_category ON extension_registry(category);
CREATE INDEX IF NOT EXISTS idx_extension_registry_installed_at ON extension_registry(installed_at DESC);

COMMENT ON TABLE extension_registry IS 'Extension registry for FunctionFly Studio integrations';