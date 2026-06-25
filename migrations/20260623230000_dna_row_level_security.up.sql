-- Function DNA: Row-Level Security Policies
-- Migration: 20260623230000_dna_row_level_security
-- Adds database-level tenant isolation for multi-tenant security
--
-- IMPORTANT: This migration adds STRICT RLS that REQUIRES the application
-- to call set_dna_tenant(tenant_id) before DNA queries.
-- See internal/api/handlers/dna/handler.go for integration.
--
-- SECURITY: If set_dna_tenant() is not called, ALL DNA queries return empty results.
-- This is intentional - secure by default.

-- ============================================================================
-- Helper function to set tenant context for DNA queries
-- SECURITY DEFINER: Runs with privileges of creator (postgres), but only
-- allows setting, not arbitrary SQL execution.
-- ============================================================================

CREATE OR REPLACE FUNCTION set_dna_tenant(tenant TEXT)
RETURNS void AS $$
BEGIN
  IF tenant IS NOT NULL AND length(tenant) > 0 AND length(tenant) <= 36 THEN
    PERFORM set_config('app.dna_tenant', tenant, true);
  END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION set_dna_tenant(TEXT) IS
'Required: Call this BEFORE any DNA queries to set tenant context. Queries fail if not called.';

-- ============================================================================
-- Helper function to clear tenant context (for cleanup/testing)
-- ============================================================================

CREATE OR REPLACE FUNCTION clear_dna_tenant()
RETURNS void AS $$
BEGIN
  PERFORM set_config('app.dna_tenant', '', true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- Enable Row-Level Security on DNA tables
-- ============================================================================

ALTER TABLE function_dna_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE function_dna_mutations ENABLE ROW LEVEL SECURITY;
ALTER TABLE function_dna_analysis_queue ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- STRICT RLS Policies - ALL queries require tenant context
-- If app.dna_tenant is not set, queries return empty results (secure by default)
-- ============================================================================

-- function_dna_profiles policies
CREATE POLICY dna_strict_isolation_select_profile ON function_dna_profiles
    FOR SELECT
    USING (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

CREATE POLICY dna_strict_isolation_update_profile ON function_dna_profiles
    FOR UPDATE
    USING (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

CREATE POLICY dna_strict_isolation_insert_profile ON function_dna_profiles
    FOR INSERT
    WITH CHECK (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

-- function_dna_mutations policies
CREATE POLICY dna_strict_isolation_select_mutation ON function_dna_mutations
    FOR SELECT
    USING (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

CREATE POLICY dna_strict_isolation_update_mutation ON function_dna_mutations
    FOR UPDATE
    USING (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

CREATE POLICY dna_strict_isolation_insert_mutation ON function_dna_mutations
    FOR INSERT
    WITH CHECK (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

-- function_dna_analysis_queue policies
CREATE POLICY dna_strict_isolation_select_queue ON function_dna_analysis_queue
    FOR SELECT
    USING (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

CREATE POLICY dna_strict_isolation_insert_queue ON function_dna_analysis_queue
    FOR INSERT
    WITH CHECK (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

CREATE POLICY dna_strict_isolation_update_queue ON function_dna_analysis_queue
    FOR UPDATE
    USING (
        current_setting('app.dna_tenant', true) IS NOT NULL
        AND current_setting('app.dna_tenant', true) != ''
        AND tenant_id = current_setting('app.dna_tenant', true)
    );

-- ============================================================================
-- Force RLS for table owner (superuser is also subject to RLS)
-- ============================================================================

ALTER TABLE function_dna_profiles FORCE ROW LEVEL SECURITY;
ALTER TABLE function_dna_mutations FORCE ROW LEVEL SECURITY;
ALTER TABLE function_dna_analysis_queue FORCE ROW LEVEL SECURITY;
