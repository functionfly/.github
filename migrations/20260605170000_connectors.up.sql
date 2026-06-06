-- Migration: Connectors catalog and user connectors (OAuth-linked accounts)
-- Part of the Connectors + Brain System

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS connectors (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       VARCHAR(50) UNIQUE NOT NULL,
    name       VARCHAR(100) NOT NULL,
    icon_url   TEXT,
    oauth_url  TEXT,
    scopes     TEXT,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO connectors (slug, name, icon_url, scopes) VALUES
    ('github',  'GitHub',  '/icons/github.svg',  'repo,read:user,notifications'),
    ('notion',  'Notion',  '/icons/notion.svg',  'read_content,read_database'),
    ('slack',   'Slack',   '/icons/slack.svg',   'channels:history,users:read,reactions:read'),
    ('gmail',   'Gmail',   '/icons/gmail.svg',   'gmail.readonly'),
    ('linear',  'Linear',  '/icons/linear.svg',  'read,issues:read')
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_connectors (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    connector_id          UUID NOT NULL REFERENCES connectors(id) ON DELETE CASCADE,
    display_name          VARCHAR(255),
    status                VARCHAR(30) NOT NULL DEFAULT 'active',
    encrypted_credentials JSONB NOT NULL DEFAULT '{}',
    last_sync_at          TIMESTAMPTZ,
    sync_error            TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, connector_id)
);

CREATE INDEX IF NOT EXISTS idx_user_connectors_tenant ON user_connectors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_connectors_status ON user_connectors(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_user_connectors_connector ON user_connectors(connector_id);
