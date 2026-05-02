-- Migration: Vault security hardening
-- 1. Reconcile secret_versions columns with Go model (add change_type, actor_type)
-- 2. Add audit log immutability trigger
-- 3. Add token cleanup helper function

-- =====================================================
-- 1. Reconcile secret_versions with Go model
--    The table may use changed_by instead of actor_id/actor_type,
--    and may lack change_type.
-- =====================================================

-- Add change_type column if missing
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'secret_versions' AND column_name = 'change_type'
    ) THEN
        ALTER TABLE secret_versions ADD COLUMN change_type VARCHAR(20) NOT NULL DEFAULT 'update';
    END IF;
END $$;

-- Add actor_type column if missing
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'secret_versions' AND column_name = 'actor_type'
    ) THEN
        ALTER TABLE secret_versions ADD COLUMN actor_type VARCHAR(50) NOT NULL DEFAULT 'user';
    END IF;
END $$;

-- Add actor_id column if missing (maps from changed_by)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'secret_versions' AND column_name = 'actor_id'
    ) THEN
        ALTER TABLE secret_versions ADD COLUMN actor_id UUID;
        UPDATE secret_versions SET actor_id = changed_by WHERE actor_id IS NULL;
        ALTER TABLE secret_versions ALTER COLUMN actor_id SET NOT NULL;
    END IF;
END $$;

-- Add rotate to change_type CHECK if constraint exists
DO $$
BEGIN
    -- Drop old constraint if it exists (may have different name)
    BEGIN
        ALTER TABLE secret_versions DROP CONSTRAINT secret_versions_change_type_check;
    EXCEPTION WHEN undefined_object THEN
        NULL;
    END;
    -- Add updated constraint
    ALTER TABLE secret_versions ADD CONSTRAINT secret_versions_change_type_check
        CHECK (change_type IN ('create', 'update', 'rollback', 'rotate'));
EXCEPTION WHEN duplicate_object THEN
    NULL;
END $$;

-- Add index for token lookup if missing
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_lookup ON secret_access_tokens(token_hash, is_revoked, expires_at);

-- =====================================================
-- 2. Audit log immutability trigger
--    Prevents UPDATE and DELETE on secrets_audit_log
-- =====================================================
CREATE OR REPLACE FUNCTION prevent_audit_log_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit logs are immutable and cannot be modified or deleted'
        USING ERRCODE = 'prohibited_sql_statement_attempted';
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_audit_log_immutable ON secrets_audit_log;
CREATE TRIGGER trg_audit_log_immutable
    BEFORE UPDATE OR DELETE ON secrets_audit_log
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_log_modification();

-- =====================================================
-- 3. Token cleanup helper function
--    Can be called by cron or application code
-- =====================================================
CREATE OR REPLACE FUNCTION cleanup_expired_vault_tokens(older_than INTERVAL DEFAULT '30 days')
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM secret_access_tokens
    WHERE (expires_at < NOW() AND expires_at < NOW() - older_than)
       OR (is_revoked = TRUE AND revoked_at < NOW() - older_than);

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_vault_tokens IS 'Removes expired and old revoked tokens. Call periodically.';
