-- Plugin Manager Schema
-- Phase 1: Core plugin tables with sandbox tiers and permissions

CREATE TABLE IF NOT EXISTS plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    manifest JSONB NOT NULL,
    plugin_type VARCHAR(50) NOT NULL DEFAULT 'ui',
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    author_name VARCHAR(255) NOT NULL,
    author_email VARCHAR(255),
    author_website VARCHAR(500),
    category VARCHAR(100) NOT NULL DEFAULT 'custom',
    status VARCHAR(50) NOT NULL DEFAULT 'disabled',
    icon_url VARCHAR(500),
    homepage_url VARCHAR(500),
    repository_url VARCHAR(500),
    license VARCHAR(100),
    size_bytes INTEGER DEFAULT 0,
    signature TEXT,
    verified BOOLEAN DEFAULT false,
    config JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    installed_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    enabled_at TIMESTAMPTZ,
    error_message TEXT,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_plugins_tenant ON plugins(tenant_id);
CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_type ON plugins(plugin_type);
CREATE INDEX idx_plugins_status ON plugins(status);
CREATE INDEX idx_plugins_category ON plugins(category);
CREATE INDEX idx_plugins_installed_at ON plugins(installed_at DESC);

COMMENT ON TABLE plugins IS 'Plugin registry for FunctionFly Studio - zero-knowledge server stores only ciphertext and metadata';