-- Migration: Automated Retention and Archival System
-- Comprehensive data lifecycle management with S3/Parquet archive support
-- Created: 2026-04-19

-- ============================================
-- 1. Enhanced Retention Audit Log
-- Track all data deletions for compliance
-- ============================================

CREATE TABLE IF NOT EXISTS retention_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(255) NOT NULL,
    partition_name VARCHAR(255),
    retention_policy VARCHAR(100) NOT NULL, -- 'detailed_90_day', 'financial_7_year', 'webhook_30_day', etc.
    cutoff_date DATE NOT NULL,
    rows_deleted BIGINT NOT NULL DEFAULT 0,
    bytes_deleted_estimate BIGINT, -- Estimated size freed
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    triggered_by VARCHAR(100) NOT NULL, -- 'scheduled_job', 'manual', 'partition_maintenance', 'legal_hold_override'
    triggered_by_user_id UUID,
    legal_hold_id UUID REFERENCES legal_holds(id),
    archive_path VARCHAR(1000), -- S3 path if archived
    archive_checksum VARCHAR(64), -- SHA-256 of archived file
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_retention_audit_table ON retention_audit_log(table_name);
CREATE INDEX idx_retention_audit_policy ON retention_audit_log(retention_policy);
CREATE INDEX idx_retention_audit_cutoff ON retention_audit_log(cutoff_date);
CREATE INDEX idx_retention_audit_deleted ON retention_audit_log(deleted_at DESC);
CREATE INDEX idx_retention_audit_hold ON retention_audit_log(legal_hold_id) WHERE legal_hold_id IS NOT NULL;

COMMENT ON TABLE retention_audit_log IS 
'Audit log of all data retention policy enforcement. Required for compliance auditing (SOX, GDPR, CCPA).';

-- ============================================
-- 2. Legal Holds Registry
-- Block deletion during litigation/audit
-- ============================================

CREATE TABLE IF NOT EXISTS legal_holds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hold_name VARCHAR(255) NOT NULL,
    description TEXT,
    hold_type VARCHAR(50) NOT NULL, -- 'litigation', 'audit', 'regulatory', 'investigation'
    -- Scope definition
    tenant_ids UUID[] DEFAULT '{}'::UUID[], -- Empty = all tenants
    table_names VARCHAR(255)[] DEFAULT '{}'::VARCHAR[], -- Empty = all tables
    affected_user_ids UUID[] DEFAULT '{}'::UUID[], -- Empty = all users
    -- Dates
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    expiration_date DATE, -- NULL = indefinite
    lifted_at TIMESTAMPTZ,
    lifted_by UUID REFERENCES users(id),
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'expired', 'lifted'
    -- Audit
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_legal_holds_status ON legal_holds(status) WHERE status = 'active';
CREATE INDEX idx_legal_holds_tenant ON legal_holds USING GIN(tenant_ids);
CREATE INDEX idx_legal_holds_tables ON legal_holds USING GIN(table_names);
CREATE INDEX idx_legal_holds_effective ON legal_holds(effective_date, expiration_date);

COMMENT ON TABLE legal_holds IS 
'Registry of legal holds that prevent data deletion. Check is_under_legal_hold() before any retention cleanup.';

-- ============================================
-- 3. Archive Configuration
-- Track S3/Parquet archive settings
-- ============================================

CREATE TABLE IF NOT EXISTS archive_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(255) NOT NULL UNIQUE,
    archive_enabled BOOLEAN NOT NULL DEFAULT true,
    archive_bucket VARCHAR(255) NOT NULL DEFAULT 'functionfly-archives',
    archive_prefix VARCHAR(500) NOT NULL DEFAULT 'data-retention/',
    archive_format VARCHAR(20) NOT NULL DEFAULT 'parquet', -- 'parquet', 'csv', 'jsonl'
    compression VARCHAR(20) NOT NULL DEFAULT 'zstd', -- 'none', 'gzip', 'zstd', 'snappy'
    -- Retention before archive
    days_before_archive INTEGER NOT NULL DEFAULT 90, -- Move to S3 after N days
    days_before_delete INTEGER NOT NULL DEFAULT 2555, -- Delete S3 after 7 years (2555 days)
    -- Last run tracking
    last_archive_at TIMESTAMPTZ,
    last_archive_path VARCHAR(1000),
    last_archive_rows BIGINT,
    last_archive_bytes BIGINT,
    last_error TEXT,
    last_error_at TIMESTAMPTZ,
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO archive_configurations (table_name, days_before_archive, days_before_delete)
VALUES 
    ('cost_allocation_entries', 90, 2555),
    ('registry_function_executions', 30, 2555),
    ('routing_events', 30, 2555),
    ('health_checks', 90, 2555),
    ('performance_metrics', 30, 2555),
    ('function_logs', 30, 2555),
    ('audit_events', 30, 2555),
    ('email_events', 90, 2555),
    ('stored_webhook_payloads', 30, 90) -- Shorter retention for webhooks
ON CONFLICT (table_name) DO NOTHING;

COMMENT ON TABLE archive_configurations IS 
'Configuration for S3-based data archiving. Defines when to move to cold storage and when to delete.';

-- ============================================
-- 4. Core Legal Hold Functions
-- ============================================

-- Function to check if a table/tenant is under legal hold
CREATE OR REPLACE FUNCTION is_under_legal_hold(
    p_table_name TEXT,
    p_tenant_id UUID DEFAULT NULL,
    p_user_id UUID DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    v_active_hold BOOLEAN := FALSE;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM legal_holds
        WHERE status = 'active'
        AND effective_date <= CURRENT_DATE
        AND (expiration_date IS NULL OR expiration_date >= CURRENT_DATE)
        AND (
            -- Check table scope
            table_names = '{}'::VARCHAR[] OR p_table_name = ANY(table_names)
        )
        AND (
            -- Check tenant scope
            tenant_ids = '{}'::UUID[] OR p_tenant_id = ANY(tenant_ids) OR p_tenant_id IS NULL
        )
        AND (
            -- Check user scope
            affected_user_ids = '{}'::UUID[] OR p_user_id = ANY(affected_user_ids) OR p_user_id IS NULL
        )
    ) INTO v_active_hold;
    
    RETURN v_active_hold;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to create a legal hold
CREATE OR REPLACE FUNCTION create_legal_hold(
    p_hold_name TEXT,
    p_description TEXT,
    p_hold_type TEXT,
    p_tenant_ids UUID[] DEFAULT '{}'::UUID[],
    p_table_names TEXT[] DEFAULT '{}'::TEXT[],
    p_affected_user_ids UUID[] DEFAULT '{}'::UUID[],
    p_effective_date DATE DEFAULT CURRENT_DATE,
    p_expiration_date DATE DEFAULT NULL,
    p_created_by UUID DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_hold_id UUID;
BEGIN
    INSERT INTO legal_holds (
        hold_name, description, hold_type,
        tenant_ids, table_names, affected_user_ids,
        effective_date, expiration_date,
        created_by, created_at, updated_at
    ) VALUES (
        p_hold_name, p_description, p_hold_type,
        p_tenant_ids, p_table_names, p_affected_user_ids,
        p_effective_date, p_expiration_date,
        p_created_by, NOW(), NOW()
    )
    RETURNING id INTO v_hold_id;
    
    -- Log the creation
    INSERT INTO retention_audit_log (
        table_name, retention_policy, cutoff_date,
        rows_deleted, triggered_by, triggered_by_user_id,
        metadata
    ) VALUES (
        'legal_holds',
        'legal_hold_created',
        p_effective_date,
        0,
        'manual',
        p_created_by,
        jsonb_build_object(
            'hold_id', v_hold_id,
            'hold_name', p_hold_name,
            'hold_type', p_hold_type,
            'affected_tenants', p_tenant_ids,
            'affected_tables', p_table_names
        )
    );
    
    RETURN v_hold_id;
END;
$$ LANGUAGE plpgsql;

-- Function to lift a legal hold
CREATE OR REPLACE FUNCTION lift_legal_hold(
    p_hold_id UUID,
    p_lifted_by UUID,
    p_reason TEXT DEFAULT 'Investigation complete'
)
RETURNS VOID AS $$
BEGIN
    UPDATE legal_holds
    SET status = 'lifted',
        lifted_at = NOW(),
        lifted_by = p_lifted_by,
        updated_at = NOW()
    WHERE id = p_hold_id;
    
    INSERT INTO retention_audit_log (
        table_name, retention_policy, cutoff_date,
        rows_deleted, triggered_by, triggered_by_user_id,
        metadata
    ) VALUES (
        'legal_holds',
        'legal_hold_lifted',
        CURRENT_DATE,
        0,
        'manual',
        p_lifted_by,
        jsonb_build_object(
            'hold_id', p_hold_id,
            'lift_reason', p_reason
        )
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 5. Archive Metadata Tracking
-- ============================================

CREATE TABLE IF NOT EXISTS archive_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(255) NOT NULL,
    archive_path VARCHAR(1000) NOT NULL,
    -- Date range in this archive
    min_timestamp TIMESTAMPTZ NOT NULL,
    max_timestamp TIMESTAMPTZ NOT NULL,
    -- Stats
    row_count BIGINT NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    file_count INTEGER NOT NULL DEFAULT 1,
    -- Compression/Format
    compression VARCHAR(20) NOT NULL,
    archive_format VARCHAR(20) NOT NULL,
    -- Verification
    checksum_sha256 VARCHAR(64) NOT NULL,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'verified', 'failed', 'restored'
    verified_at TIMESTAMPTZ,
    -- Status
    deleted_from_source BOOLEAN NOT NULL DEFAULT false,
    deleted_from_source_at TIMESTAMPTZ,
    -- S3 lifecycle
    s3_storage_class VARCHAR(20) NOT NULL DEFAULT 'STANDARD', -- 'STANDARD', 'GLACIER', 'DEEP_ARCHIVE'
    s3_lifecycle_updated_at TIMESTAMPTZ,
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_archive_batches_table ON archive_batches(table_name);
CREATE INDEX idx_archive_batches_timestamp ON archive_batches(min_timestamp, max_timestamp);
CREATE INDEX idx_archive_batches_status ON archive_batches(verification_status, deleted_from_source);

COMMENT ON TABLE archive_batches IS 
'Tracking for archived data batches in S3. Used for verification and restore operations.';

-- ============================================
-- 6. Retention Cleanup Functions
-- ============================================

-- Function to get retention cutoff for a table
CREATE OR REPLACE FUNCTION get_retention_cutoff(
    p_table_name TEXT,
    p_custom_days INTEGER DEFAULT NULL
)
RETURNS DATE AS $$
DECLARE
    v_config RECORD;
    v_cutoff DATE;
BEGIN
    -- Get archive configuration
    SELECT * INTO v_config
    FROM archive_configurations
    WHERE table_name = p_table_name;
    
    IF v_config IS NULL THEN
        -- Default retention
        v_cutoff := CURRENT_DATE - COALESCE(p_custom_days, 90);
    ELSE
        v_cutoff := CURRENT_DATE - v_config.days_before_archive;
    END IF;
    
    RETURN v_cutoff;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to cleanup cost_allocation_entries with legal hold check
CREATE OR REPLACE FUNCTION cleanup_cost_allocation_entries(
    p_batch_size INTEGER DEFAULT 10000,
    p_dry_run BOOLEAN DEFAULT true
)
RETURNS TABLE (
    cutoff_date DATE,
    rows_affected BIGINT,
    archived BOOLEAN,
    legal_holds_blocked BOOLEAN
) AS $$
DECLARE
    v_cutoff DATE;
    v_hold_exists BOOLEAN;
    v_count BIGINT := 0;
    v_total_count BIGINT := 0;
    v_batch_count INTEGER := 0;
BEGIN
    -- Get cutoff
    v_cutoff := get_retention_cutoff('cost_allocation_entries');
    
    -- Check legal holds
    v_hold_exists := is_under_legal_hold('cost_allocation_entries');
    
    IF v_hold_exists THEN
        RETURN QUERY SELECT v_cutoff, 0::BIGINT, false, true;
        RETURN;
    END IF;
    
    IF p_dry_run THEN
        -- Count only
        SELECT COUNT(*) INTO v_total_count
        FROM cost_allocation_entries
        WHERE timestamp < v_cutoff;
        
        RETURN QUERY SELECT v_cutoff, v_total_count, false, false;
        RETURN;
    END IF;
    
    -- Real deletion in batches
    LOOP
        WITH to_delete AS (
            SELECT id, timestamp
            FROM cost_allocation_entries
            WHERE timestamp < v_cutoff
            AND NOT EXISTS (
                SELECT 1 FROM legal_holds
                WHERE status = 'active'
                AND (tenant_ids = '{}'::UUID[] OR tenant_id = ANY(tenant_ids))
            )
            LIMIT p_batch_size
            FOR UPDATE SKIP LOCKED
        )
        DELETE FROM cost_allocation_entries
        WHERE (id, timestamp) IN (SELECT id, timestamp FROM to_delete);
        
        GET DIAGNOSTICS v_count = ROW_COUNT;
        v_total_count := v_total_count + v_count;
        v_batch_count := v_batch_count + 1;
        
        EXIT WHEN v_count = 0 OR v_batch_count >= 100; -- Safety limit
        
        COMMIT;
        PERFORM pg_sleep(0.1); -- Brief pause between batches
    END LOOP;
    
    -- Log the cleanup
    INSERT INTO retention_audit_log (
        table_name, retention_policy, cutoff_date,
        rows_deleted, triggered_by, metadata
    ) VALUES (
        'cost_allocation_entries',
        'detailed_90_day',
        v_cutoff,
        v_total_count,
        'scheduled_job',
        jsonb_build_object('batches', v_batch_count, 'dry_run', p_dry_run)
    );
    
    RETURN QUERY SELECT v_cutoff, v_total_count, false, false;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup registry_function_executions
CREATE OR REPLACE FUNCTION cleanup_registry_executions(
    p_batch_size INTEGER DEFAULT 10000,
    p_dry_run BOOLEAN DEFAULT true
)
RETURNS TABLE (
    cutoff_date DATE,
    rows_affected BIGINT,
    archived BOOLEAN,
    legal_holds_blocked BOOLEAN
) AS $$
DECLARE
    v_cutoff DATE;
    v_hold_exists BOOLEAN;
    v_count BIGINT := 0;
    v_total_count BIGINT := 0;
BEGIN
    v_cutoff := get_retention_cutoff('registry_function_executions');
    v_hold_exists := is_under_legal_hold('registry_function_executions');
    
    IF v_hold_exists THEN
        RETURN QUERY SELECT v_cutoff, 0::BIGINT, false, true;
        RETURN;
    END IF;
    
    IF p_dry_run THEN
        SELECT COUNT(*) INTO v_total_count
        FROM registry_function_executions
        WHERE timestamp < v_cutoff;
        
        RETURN QUERY SELECT v_cutoff, v_total_count, false, false;
        RETURN;
    END IF;
    
    LOOP
        WITH to_delete AS (
            SELECT id, timestamp
            FROM registry_function_executions
            WHERE timestamp < v_cutoff
            LIMIT p_batch_size
            FOR UPDATE SKIP LOCKED
        )
        DELETE FROM registry_function_executions
        WHERE (id, timestamp) IN (SELECT id, timestamp FROM to_delete);
        
        GET DIAGNOSTICS v_count = ROW_COUNT;
        v_total_count := v_total_count + v_count;
        
        EXIT WHEN v_count = 0;
        
        COMMIT;
        PERFORM pg_sleep(0.1);
    END LOOP;
    
    INSERT INTO retention_audit_log (
        table_name, retention_policy, cutoff_date,
        rows_deleted, triggered_by
    ) VALUES (
        'registry_function_executions',
        'detailed_30_day',
        v_cutoff,
        v_total_count,
        'scheduled_job'
    );
    
    RETURN QUERY SELECT v_cutoff, v_total_count, false, false;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 7. Master Retention Cleanup Function
-- ============================================

CREATE OR REPLACE FUNCTION execute_retention_cleanup(
    p_dry_run BOOLEAN DEFAULT true,
    p_tables TEXT[] DEFAULT NULL -- NULL = all configured tables
)
RETURNS TABLE (
    table_name TEXT,
    cutoff_date DATE,
    rows_affected BIGINT,
    archived BOOLEAN,
    legal_holds_blocked BOOLEAN,
    error_message TEXT
) AS $$
DECLARE
    v_tables TEXT[] := COALESCE(p_tables, ARRAY[
        'cost_allocation_entries',
        'registry_function_executions',
        'routing_events',
        'health_checks',
        'performance_metrics',
        'function_logs'
    ]);
    v_tbl TEXT;
    v_result RECORD;
BEGIN
    FOREACH v_tbl IN ARRAY v_tables
    LOOP
        BEGIN
            CASE v_tbl
                WHEN 'cost_allocation_entries' THEN
                    SELECT * INTO v_result FROM cleanup_cost_allocation_entries(10000, p_dry_run);
                    RETURN QUERY SELECT v_tbl::TEXT, v_result.cutoff_date, v_result.rows_affected, 
                                        v_result.archived, v_result.legal_holds_blocked, NULL::TEXT;
                WHEN 'registry_function_executions' THEN
                    SELECT * INTO v_result FROM cleanup_registry_executions(10000, p_dry_run);
                    RETURN QUERY SELECT v_tbl::TEXT, v_result.cutoff_date, v_result.rows_affected,
                                        v_result.archived, v_result.legal_holds_blocked, NULL::TEXT;
                ELSE
                    RETURN QUERY SELECT v_tbl::TEXT, NULL::DATE, 0::BIGINT,
                                        false, false, 'Cleanup function not implemented'::TEXT;
            END CASE;
        EXCEPTION WHEN OTHERS THEN
            RETURN QUERY SELECT v_tbl::TEXT, NULL::DATE, 0::BIGINT,
                                false, false, SQLERRM::TEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION execute_retention_cleanup(BOOLEAN, TEXT[]) IS 
'Master function to run retention cleanup on configured tables. Set dry_run=false for actual deletion.';

-- ============================================
-- 8. Archive Trigger Functions (for S3 integration)
-- ============================================

-- Function to prepare archive batch metadata
CREATE OR REPLACE FUNCTION prepare_archive_batch(
    p_table_name TEXT,
    p_min_timestamp TIMESTAMPTZ,
    p_max_timestamp TIMESTAMPTZ,
    p_row_count BIGINT,
    p_file_size_bytes BIGINT,
    p_checksum_sha256 TEXT,
    p_archive_path TEXT,
    p_compression TEXT DEFAULT 'zstd',
    p_format TEXT DEFAULT 'parquet'
)
RETURNS UUID AS $$
DECLARE
    v_batch_id UUID;
BEGIN
    INSERT INTO archive_batches (
        table_name, min_timestamp, max_timestamp,
        row_count, file_size_bytes, checksum_sha256,
        archive_path, compression, archive_format,
        verification_status, created_at
    ) VALUES (
        p_table_name, p_min_timestamp, p_max_timestamp,
        p_row_count, p_file_size_bytes, p_checksum_sha256,
        p_archive_path, p_compression, p_format,
        'pending', NOW()
    )
    RETURNING id INTO v_batch_id;
    
    -- Update archive config with last run info
    UPDATE archive_configurations
    SET last_archive_at = NOW(),
        last_archive_path = p_archive_path,
        last_archive_rows = p_row_count,
        last_archive_bytes = p_file_size_bytes,
        last_error = NULL,
        last_error_at = NULL,
        updated_at = NOW()
    WHERE table_name = p_table_name;
    
    RETURN v_batch_id;
END;
$$ LANGUAGE plpgsql;

-- Function to mark archive as verified
CREATE OR REPLACE FUNCTION verify_archive_batch(
    p_batch_id UUID,
    p_verified BOOLEAN,
    p_error_message TEXT DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    UPDATE archive_batches
    SET verification_status = CASE WHEN p_verified THEN 'verified' ELSE 'failed' END,
        verified_at = NOW(),
        updated_at = NOW()
    WHERE id = p_batch_id;
    
    -- If verification failed, update config with error
    IF NOT p_verified THEN
        UPDATE archive_configurations
        SET last_error = p_error_message,
            last_error_at = NOW(),
            updated_at = NOW()
        WHERE table_name = (
            SELECT table_name FROM archive_batches WHERE id = p_batch_id
        );
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to mark source data as deleted after archive verification
CREATE OR REPLACE FUNCTION confirm_source_deleted(
    p_batch_id UUID
)
RETURNS VOID AS $$
BEGIN
    UPDATE archive_batches
    SET deleted_from_source = true,
        deleted_from_source_at = NOW(),
        updated_at = NOW()
    WHERE id = p_batch_id;
    
    -- Log retention
    INSERT INTO retention_audit_log (
        table_name, partition_name, retention_policy,
        cutoff_date, rows_deleted, triggered_by, metadata
    )
    SELECT 
        ab.table_name,
        ab.archive_path,
        'archived_to_s3',
        ab.max_timestamp::DATE,
        ab.row_count,
        'archive_process',
        jsonb_build_object(
            'batch_id', p_batch_id,
            'archive_path', ab.archive_path,
            'storage_class', ab.s3_storage_class
        )
    FROM archive_batches ab
    WHERE ab.id = p_batch_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 9. Retention Summary View
-- ============================================

CREATE OR REPLACE VIEW retention_summary AS
WITH table_stats AS (
    SELECT 
        'cost_allocation_entries'::TEXT as table_name,
        COUNT(*) as total_rows,
        MIN(timestamp) as oldest_record,
        MAX(timestamp) as newest_record
    FROM cost_allocation_entries
    UNION ALL
    SELECT 
        'registry_function_executions'::TEXT,
        COUNT(*),
        MIN(timestamp),
        MAX(timestamp)
    FROM registry_function_executions
    UNION ALL
    SELECT 
        'routing_events'::TEXT,
        COUNT(*),
        MIN(timestamp),
        MAX(timestamp)
    FROM routing_events
)
SELECT 
    ts.table_name,
    ts.total_rows,
    ts.oldest_record,
    ts.newest_record,
    ac.days_before_archive,
    ac.days_before_delete,
    ac.archive_enabled,
    get_retention_cutoff(ts.table_name) as next_cleanup_cutoff,
    EXISTS (
        SELECT 1 FROM legal_holds
        WHERE status = 'active'
        AND table_names = '{}'::VARCHAR[] OR ts.table_name = ANY(table_names)
    ) as has_active_legal_hold,
    ac.last_archive_at,
    ac.last_archive_rows
FROM table_stats ts
LEFT JOIN archive_configurations ac ON ac.table_name = ts.table_name;

COMMENT ON VIEW retention_summary IS 
'Dashboard view for retention status across all managed tables.';

-- ============================================
-- 10. Comments for Documentation
-- ============================================

COMMENT ON FUNCTION is_under_legal_hold(TEXT, UUID, UUID) IS 
'Check if a table/tenant is under an active legal hold. Must be called before any retention deletion.';

COMMENT ON FUNCTION cleanup_cost_allocation_entries(INTEGER, BOOLEAN) IS 
'Clean up old cost allocation entries respecting legal holds. Use dry_run first to estimate impact.';

COMMENT ON FUNCTION prepare_archive_batch(TEXT, TIMESTAMPTZ, TIMESTAMPTZ, BIGINT, BIGINT, TEXT, TEXT, TEXT, TEXT) IS 
'Called by external archive process (Go Lambda) to register a new S3 archive batch.';
