-- Down migration: Drop Trust API core tables
-- Description: Removes the foundational Trust API tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_trust_api_partners_updated_at ON trust_api_partners;
DROP TRIGGER IF EXISTS update_trust_api_keys_updated_at ON trust_api_keys;
DROP TRIGGER IF EXISTS update_trust_api_reports_updated_at ON trust_api_reports;
DROP TRIGGER IF EXISTS update_trust_api_verifications_updated_at ON trust_api_verifications;

-- Drop functions
DROP FUNCTION IF EXISTS update_trust_api_core_updated_at_column();

-- Drop RLS policies
DROP POLICY IF EXISTS trust_partners_select ON trust_api_partners;
DROP POLICY IF EXISTS trust_partners_update ON trust_api_partners;
DROP POLICY IF EXISTS trust_api_keys_select ON trust_api_keys;
DROP POLICY IF EXISTS trust_api_keys_insert ON trust_api_keys;
DROP POLICY IF EXISTS trust_api_keys_update ON trust_api_keys;
DROP POLICY IF EXISTS trust_api_keys_delete ON trust_api_keys;
DROP POLICY IF EXISTS trust_api_usage_select ON trust_api_usage;
DROP POLICY IF EXISTS trust_api_usage_insert ON trust_api_usage;
DROP POLICY IF EXISTS trust_api_rate_limits_select ON trust_api_rate_limits;
DROP POLICY IF EXISTS trust_api_rate_limits_insert ON trust_api_rate_limits;
DROP POLICY IF EXISTS trust_api_rate_limits_update ON trust_api_rate_limits;
DROP POLICY IF EXISTS trust_api_reports_select ON trust_api_reports;
DROP POLICY IF EXISTS trust_api_reports_insert ON trust_api_reports;
DROP POLICY IF EXISTS trust_api_reports_update ON trust_api_reports;
DROP POLICY IF EXISTS trust_api_verifications_select ON trust_api_verifications;
DROP POLICY IF EXISTS trust_api_verifications_insert ON trust_api_verifications;
DROP POLICY IF EXISTS trust_api_verifications_update ON trust_api_verifications;

-- Disable RLS before dropping tables
ALTER TABLE trust_api_verifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_reports DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_rate_limits DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_usage DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_keys DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_partners DISABLE ROW LEVEL SECURITY;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS trust_api_verifications;
DROP TABLE IF EXISTS trust_api_reports;
DROP TABLE IF EXISTS trust_api_rate_limits;
DROP TABLE IF EXISTS trust_api_usage;
DROP TABLE IF EXISTS trust_api_keys;
DROP TABLE IF EXISTS trust_api_partners;
