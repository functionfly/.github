-- Migration: Drop trust infrastructure tables
-- Description: Removes trust revocation, attestation, and policy tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_trust_revocations_updated_at ON trust_revocations;
DROP TRIGGER IF EXISTS update_trust_attestations_updated_at ON trust_attestations;
DROP TRIGGER IF EXISTS update_trust_policies_updated_at ON trust_policies;

-- Drop function
DROP FUNCTION IF EXISTS update_trust_updated_at_column();

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS trust_policy_evaluations;
DROP TABLE IF EXISTS trust_policies;
DROP TABLE IF EXISTS trust_attestations;
DROP TABLE IF EXISTS trust_revocations;
