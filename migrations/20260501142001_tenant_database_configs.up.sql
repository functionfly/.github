-- Tenant database configs table
-- Stores credentials and connection info for per-tenant dedicated databases

CREATE TABLE IF NOT EXISTS tenant_database_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    db_name VARCHAR(100) NOT NULL,
    db_host VARCHAR(255),
    db_port INT DEFAULT 5432,
    db_user VARCHAR(100),
    db_password_encrypted TEXT,  -- Encrypted with platform vault key
    encryption_version INT DEFAULT 1,  -- For key rotation support
    connection_string_template TEXT,  -- Template for connection (password redacted)
    max_connections INT DEFAULT 10,
    min_connections INT DEFAULT 2,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'provisioning'  -- 'provisioning', 'active', 'suspended', 'failed', 'deleting'
);

CREATE INDEX IF NOT EXISTS idx_tenant_database_configs_tenant_id ON tenant_database_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_database_configs_status ON tenant_database_configs(status);

-- Template database for cloning (optional optimization)
CREATE TABLE IF NOT EXISTS tenant_db_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_tenant_db_templates_name ON tenant_db_templates(name);

-- Migration history for tenant databases (tracks schema version per tenant DB)
CREATE TABLE IF NOT EXISTS tenant_db_migrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    migration_version INT NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    duration_ms INT,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_tenant_db_migrations_tenant_id ON tenant_db_migrations(tenant_id);