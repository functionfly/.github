-- Function DNA: Row-Level Security Policies Rollback
-- Migration: 20260623230000_dna_row_level_security (down)

-- ============================================================================
-- Drop RLS policies from function_dna_analysis_queue
-- ============================================================================

DROP POLICY IF EXISTS dna_strict_isolation_update_queue ON function_dna_analysis_queue;
DROP POLICY IF EXISTS dna_strict_isolation_insert_queue ON function_dna_analysis_queue;
DROP POLICY IF EXISTS dna_strict_isolation_select_queue ON function_dna_analysis_queue;

-- ============================================================================
-- Drop RLS policies from function_dna_mutations
-- ============================================================================

DROP POLICY IF EXISTS dna_strict_isolation_update_mutation ON function_dna_mutations;
DROP POLICY IF EXISTS dna_strict_isolation_insert_mutation ON function_dna_mutations;
DROP POLICY IF EXISTS dna_strict_isolation_select_mutation ON function_dna_mutations;

-- ============================================================================
-- Drop RLS policies from function_dna_profiles
-- ============================================================================

DROP POLICY IF EXISTS dna_strict_isolation_update_profile ON function_dna_profiles;
DROP POLICY IF EXISTS dna_strict_isolation_insert_profile ON function_dna_profiles;
DROP POLICY IF EXISTS dna_strict_isolation_select_profile ON function_dna_profiles;

-- ============================================================================
-- Disable Row-Level Security on DNA tables
-- ============================================================================

ALTER TABLE function_dna_analysis_queue DISABLE ROW LEVEL SECURITY;
ALTER TABLE function_dna_mutations DISABLE ROW LEVEL SECURITY;
ALTER TABLE function_dna_profiles DISABLE ROW LEVEL SECURITY;

-- ============================================================================
-- Drop helper functions
-- ============================================================================

DROP FUNCTION IF EXISTS clear_dna_tenant();
DROP FUNCTION IF EXISTS set_dna_tenant(TEXT);
