-- Migration: vault_quotas_phase5
-- Created at: 2026-06-12
-- Purpose: Phase 5.2 — per-tenant rate limits + quotas.
-- The vault_rate_limits table holds admin-set overrides; the
-- per-plan defaults live in code (see internal/storage/vault/quota).

BEGIN;

CREATE TABLE IF NOT EXISTS vault_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource VARCHAR(50) NOT NULL,
    requests_per_minute INTEGER NOT NULL DEFAULT 0,
    requests_per_hour INTEGER NOT NULL DEFAULT 0,
    max_secrets INTEGER NOT NULL DEFAULT 0,
    max_dynamic_credentials INTEGER NOT NULL DEFAULT 0,
    audit_exports_per_day INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vault_rate_limits_resource_check
        CHECK (resource IN ('secrets','dynamic_credentials','tokens','audit_exports')),
    CONSTRAINT vault_rate_limits_tenant_resource_unique
        UNIQUE (tenant_id, resource)
);

CREATE INDEX IF NOT EXISTS idx_vault_rate_limits_tenant
    ON vault_rate_limits(tenant_id);

COMMENT ON TABLE vault_rate_limits IS 'Admin-configured overrides for per-tenant rate limits and quotas. Zero values mean "use the plan default".';

-- A "resource" of 'secrets' means: row is the secret-count cap.
-- 'dynamic_credentials': dynamic-credential monthly cap.
-- 'tokens': token-per-secret cap.
-- 'audit_exports': audit exports per day.
-- (Multiple resources per tenant share the same row? No — we use
-- one row per (tenant, resource) pair to keep admin updates atomic.)

COMMIT;
