-- Migration: Database Monitoring and Operational Functions
-- Tools for DBA operations, health checks, and performance monitoring
-- Created: 2026-04-19

-- ============================================
-- 1. Connection and Lock Monitoring
-- ============================================

-- View: Active connections by application/user
CREATE OR REPLACE VIEW active_connections AS
SELECT 
    pid,
    usename as username,
    application_name,
    client_addr,
    client_port,
    backend_start,
    xact_start,
    query_start,
    state,
    wait_event_type,
    wait_event,
    left(query, 100) as query_preview
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start;

COMMENT ON VIEW active_connections IS 
'Shows currently active database connections with query previews.';

-- View: Lock conflicts and blocking queries
CREATE OR REPLACE VIEW lock_conflicts AS
SELECT 
    blocked_locks.pid as blocked_pid,
    blocked_activity.usename as blocked_user,
    blocking_locks.pid as blocking_pid,
    blocking_activity.usename as blocking_user,
    blocked_activity.query as blocked_statement,
    blocking_activity.query as blocking_statement,
    blocked_activity.application_name as blocked_app,
    blocking_activity.application_name as blocking_app
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks 
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

COMMENT ON VIEW lock_conflicts IS 
'Shows queries that are blocked by other queries (lock conflicts).';

-- Function to kill a long-running query
CREATE OR REPLACE FUNCTION terminate_long_query(
    p_min_duration_minutes INTEGER DEFAULT 30,
    p_dry_run BOOLEAN DEFAULT true
)
RETURNS TABLE (pid INTEGER, duration INTERVAL, query TEXT, terminated BOOLEAN) AS $$
DECLARE
    v_rec RECORD;
BEGIN
    FOR v_rec IN 
        SELECT 
            a.pid,
            NOW() - a.query_start as duration,
            left(a.query, 200) as query_text
        FROM pg_stat_activity a
        WHERE a.state = 'active'
        AND a.query_start < NOW() - (p_min_duration_minutes || ' minutes')::INTERVAL
        AND a.pid != pg_backend_pid() -- Don't kill ourselves
        AND a.usename = current_user() -- Only our own queries (safety)
        ORDER BY a.query_start
    LOOP
        IF NOT p_dry_run THEN
            PERFORM pg_terminate_backend(v_rec.pid);
        END IF;
        
        RETURN QUERY SELECT v_rec.pid, v_rec.duration, v_rec.query_text, NOT p_dry_run;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION terminate_long_query(INTEGER, BOOLEAN) IS 
'Terminate queries running longer than threshold. Default dry_run=true for safety.';

-- ============================================
-- 2. Table and Index Health
-- ============================================

-- View: Table bloat estimates
CREATE OR REPLACE VIEW table_bloat AS
SELECT
    schemaname,
    tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_tuple_ratio,
    CASE 
        WHEN n_dead_tup > 10000 AND (100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0)) > 10 
        THEN 'VACUUM RECOMMENDED'
        ELSE 'OK'
    END as recommendation
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY n_dead_tup DESC;

COMMENT ON VIEW table_bloat IS 
'Identifies tables with high dead tuple counts that need vacuuming.';

-- View: Index bloat and usage
CREATE OR REPLACE VIEW index_health AS
SELECT
    schemaname,
    tablename,
    indexrelname as index_name,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    pg_relation_size(indexrelid) as index_size_bytes,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED - Consider dropping'
        ELSE 'OK'
    END as usage_status
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan;

COMMENT ON VIEW index_health IS 
'Identifies unused indexes that may be candidates for removal.';

-- Function to estimate table bloat percentage
CREATE OR REPLACE FUNCTION estimate_table_bloat(p_table_name TEXT)
RETURNS TABLE (
    table_size TEXT,
    estimated_bloat_pct FLOAT,
    estimated_wasted_space TEXT
) AS $$
DECLARE
    v_table_oid OID;
    v_table_size BIGINT;
    v_live_tuples BIGINT;
    v_page_size BIGINT := 8192; -- PostgreSQL default
    v_expected_pages BIGINT;
    v_actual_pages BIGINT;
    v_bloat_pct FLOAT;
BEGIN
    -- Get table OID and size
    SELECT pg_relation_size(c.oid), c.relpages, c.reltuples
    INTO v_table_size, v_actual_pages, v_live_tuples
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public' AND c.relname = p_table_name;
    
    IF v_live_tuples = 0 THEN
        v_bloat_pct := 0;
    ELSE
        -- Rough estimate: pages needed vs actual pages
        v_expected_pages := GREATEST(1, v_live_tuples / 200); -- ~200 tuples per page estimate
        v_bloat_pct := 100.0 * (v_actual_pages - v_expected_pages) / v_actual_pages;
    END IF;
    
    RETURN QUERY SELECT 
        pg_size_pretty(v_table_size),
        GREATEST(0, v_bloat_pct)::FLOAT,
        pg_size_pretty(GREATEST(0, v_table_size * v_bloat_pct / 100)::BIGINT);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION estimate_table_bloat(TEXT) IS 
'Estimates bloat percentage for a table. Uses heuristics, not exact measurement.';

-- ============================================
-- 3. Vacuum and Autovacuum Monitoring
-- ============================================

-- View: Autovacuum activity and backlog
CREATE OR REPLACE VIEW autovacuum_status AS
SELECT 
    relname as table_name,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze,
    vacuum_count,
    autovacuum_count,
    CASE 
        WHEN n_dead_tup > 1000 AND last_autovacuum < NOW() - INTERVAL '1 day' 
        THEN 'BACKLOG - Manual vacuum may be needed'
        WHEN last_autovacuum > NOW() - INTERVAL '1 hour' 
        THEN 'RECENT AUTOVACUUM'
        ELSE 'OK'
    END as status
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY n_dead_tup DESC;

COMMENT ON VIEW autovacuum_status IS 
'Monitors autovacuum activity and identifies tables needing attention.';

-- Function to trigger vacuum on bloated tables
CREATE OR REPLACE FUNCTION vacuum_bloated_tables(
    p_min_dead_tuples INTEGER DEFAULT 10000,
    p_dry_run BOOLEAN DEFAULT true
)
RETURNS TABLE (table_name TEXT, dead_tuples BIGINT, vacuum_initiated BOOLEAN) AS $$
DECLARE
    v_rec RECORD;
BEGIN
    FOR v_rec IN 
        SELECT relname, n_dead_tup
        FROM pg_stat_user_tables
        WHERE schemaname = 'public'
        AND n_dead_tup >= p_min_dead_tuples
        AND last_vacuum < NOW() - INTERVAL '1 hour'
        ORDER BY n_dead_tup DESC
        LIMIT 10 -- Don't overwhelm with too many
    LOOP
        IF NOT p_dry_run THEN
            EXECUTE format('VACUUM ANALYZE %I', v_rec.relname);
        END IF;
        
        RETURN QUERY SELECT v_rec.relname::TEXT, v_rec.n_dead_tup::BIGINT, NOT p_dry_run;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION vacuum_bloated_tables(INTEGER, BOOLEAN) IS 
'Vacuum tables with high dead tuple counts. dry_run=true to preview, false to execute.';

-- ============================================
-- 4. Query Performance Monitoring
-- ============================================

-- View: Slow queries from pg_stat_statements (if extension enabled)
CREATE OR REPLACE VIEW slow_queries AS
SELECT 
    substring(query, 1, 100) as query_preview,
    calls,
    ROUND(total_exec_time::numeric, 2) as total_time_ms,
    ROUND(mean_exec_time::numeric, 2) as mean_time_ms,
    ROUND(max_exec_time::numeric, 2) as max_time_ms,
    rows,
    ROUND(100.0 * shared_blks_hit / NULLIF(shared_blks_hit + shared_blks_read, 0), 2) as cache_hit_pct
FROM pg_stat_statements
WHERE mean_exec_time > 100 -- Queries averaging > 100ms
ORDER BY mean_exec_time DESC
LIMIT 50;

COMMENT ON VIEW slow_queries IS 
'Shows slowest queries from pg_stat_statements. Requires pg_stat_statements extension.';

-- View: Missing indexes (sequential scans on large tables)
CREATE OR REPLACE VIEW missing_index_candidates AS
SELECT 
    schemaname,
    relname as table_name,
    seq_scan,
    seq_tup_read,
    idx_scan,
    n_live_tup as table_size,
    CASE 
        WHEN seq_scan > 0 AND idx_scan IS NULL AND n_live_tup > 10000 
        THEN 'HIGH - Add index'
        WHEN seq_scan > idx_scan AND n_live_tup > 1000 
        THEN 'MEDIUM - Review indexing'
        ELSE 'OK'
    END as priority
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY seq_tup_read DESC;

COMMENT ON VIEW missing_index_candidates IS 
'Identifies tables that may benefit from additional indexes based on scan patterns.';

-- ============================================
-- 5. Storage and I/O Monitoring
-- ============================================

-- View: Database size breakdown
CREATE OR REPLACE VIEW database_size_breakdown AS
SELECT 
    schemaname,
    relname as object_name,
    CASE 
        WHEN relkind = 'r' THEN 'table'
        WHEN relkind = 'i' THEN 'index'
        WHEN relkind = 'S' THEN 'sequence'
        WHEN relkind = 't' THEN 'TOAST table'
        WHEN relkind = 'm' THEN 'materialized view'
        WHEN relkind = 'p' THEN 'partitioned table'
        ELSE relkind::TEXT
    END as object_type,
    pg_size_pretty(pg_total_relation_size(relid)) as total_size,
    pg_total_relation_size(relid) as total_size_bytes,
    pg_size_pretty(pg_relation_size(relid)) as data_size,
    pg_size_pretty(pg_indexes_size(relid)) as index_size,
    pg_size_pretty(pg_total_relation_size(relid) - pg_relation_size(relid)) as overhead
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(relid) DESC;

COMMENT ON VIEW database_size_breakdown IS 
'Breakdown of storage usage by table and indexes.';

-- Function to get total database size
CREATE OR REPLACE FUNCTION get_database_size()
RETURNS TABLE (
    database_name TEXT,
    total_size TEXT,
    total_size_bytes BIGINT
) AS $$
BEGIN
    RETURN QUERY 
    SELECT 
        current_database()::TEXT,
        pg_size_pretty(pg_database_size(current_database())),
        pg_database_size(current_database());
END;
$$ LANGUAGE plpgsql;

-- Function to get table sizes over threshold
CREATE OR REPLACE FUNCTION get_large_tables(
    p_min_size_mb INTEGER DEFAULT 100
)
RETURNS TABLE (
    table_name TEXT,
    total_size TEXT,
    total_size_bytes BIGINT,
    row_estimate BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        relname::TEXT,
        pg_size_pretty(pg_total_relation_size(relid)),
        pg_total_relation_size(relid),
        n_live_tup::BIGINT
    FROM pg_stat_user_tables
    WHERE schemaname = 'public'
    AND pg_total_relation_size(relid) > (p_min_size_mb * 1024 * 1024)
    ORDER BY pg_total_relation_size(relid) DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 6. Replication Lag Monitoring (if replicas exist)
-- ============================================

-- View: Replication status
CREATE OR REPLACE VIEW replication_status AS
SELECT 
    client_addr as replica_address,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_size_pretty(pg_wal_lsn_diff(sent_lsn, replay_lsn)) as replication_lag,
    reply_time
FROM pg_stat_replication;

COMMENT ON VIEW replication_status IS 
'Shows streaming replication status and lag. Requires replicas to be connected.';

-- Function to check replication lag
CREATE OR REPLACE FUNCTION check_replication_lag()
RETURNS TABLE (
    replica_address INET,
    lag_bytes BIGINT,
    lag_size TEXT,
    lag_seconds INTERVAL,
    status TEXT
) AS $$
DECLARE
    v_rec RECORD;
    v_lag_bytes BIGINT;
    v_lag_seconds INTERVAL;
    v_status TEXT;
BEGIN
    FOR v_rec IN 
        SELECT client_addr, sent_lsn, replay_lsn, reply_time
        FROM pg_stat_replication
        WHERE state = 'streaming'
    LOOP
        v_lag_bytes := pg_wal_lsn_diff(v_rec.sent_lsn, v_rec.replay_lsn);
        v_lag_seconds := NOW() - v_rec.reply_time;
        
        v_status := CASE 
            WHEN v_lag_bytes > 100000000 OR v_lag_seconds > INTERVAL '30 seconds' THEN 'CRITICAL'
            WHEN v_lag_bytes > 10000000 OR v_lag_seconds > INTERVAL '5 seconds' THEN 'WARNING'
            ELSE 'OK'
        END;
        
        RETURN QUERY SELECT 
            v_rec.client_addr,
            v_lag_bytes,
            pg_size_pretty(v_lag_bytes),
            v_lag_seconds,
            v_status;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_replication_lag() IS 
'Checks replication lag with severity levels (OK/WARNING/CRITICAL).';

-- ============================================
-- 7. Health Check Functions
-- ============================================

-- Function: Quick health check
CREATE OR REPLACE FUNCTION quick_health_check()
RETURNS TABLE (
    check_name TEXT,
    status TEXT,
    details TEXT
) AS $$
DECLARE
    v_count INTEGER;
    v_lag BIGINT;
BEGIN
    -- Check 1: Connection count
    SELECT count(*) INTO v_count FROM pg_stat_activity;
    RETURN QUERY SELECT 
        'connections'::TEXT,
        CASE WHEN v_count > 200 THEN 'WARNING' ELSE 'OK' END,
        format('%s connections (max_connections: %s)', v_count, current_setting('max_connections'))::TEXT;
    
    -- Check 2: Lock conflicts
    SELECT count(*) INTO v_count FROM lock_conflicts;
    RETURN QUERY SELECT 
        'locks'::TEXT,
        CASE WHEN v_count > 0 THEN 'WARNING' ELSE 'OK' END,
        format('%s blocked queries', v_count)::TEXT;
    
    -- Check 3: Replication lag
    IF EXISTS (SELECT 1 FROM pg_stat_replication) THEN
        SELECT max(pg_wal_lsn_diff(sent_lsn, replay_lsn)) INTO v_lag FROM pg_stat_replication;
        RETURN QUERY SELECT 
            'replication'::TEXT,
            CASE WHEN v_lag > 10000000 THEN 'WARNING' ELSE 'OK' END,
            format('Lag: %s', pg_size_pretty(v_lag))::TEXT;
    ELSE
        RETURN QUERY SELECT 
            'replication'::TEXT,
            'N/A'::TEXT,
            'No active replication'::TEXT;
    END IF;
    
    -- Check 4: Dead tuples
    SELECT count(*) INTO v_count 
    FROM pg_stat_user_tables 
    WHERE n_dead_tup > 10000 AND schemaname = 'public';
    RETURN QUERY SELECT 
        'bloat'::TEXT,
        CASE WHEN v_count > 5 THEN 'WARNING' ELSE 'OK' END,
        format('%s tables with >10k dead tuples', v_count)::TEXT;
    
    -- Check 5: Long running queries
    SELECT count(*) INTO v_count 
    FROM pg_stat_activity 
    WHERE state = 'active' 
    AND query_start < NOW() - INTERVAL '10 minutes'
    AND pid != pg_backend_pid();
    RETURN QUERY SELECT 
        'long_queries'::TEXT,
        CASE WHEN v_count > 0 THEN 'WARNING' ELSE 'OK' END,
        format('%s queries > 10min', v_count)::TEXT;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION quick_health_check() IS 
'Quick health check returning 5 key metrics. Use for monitoring dashboards.';

-- Function: Detailed health report
CREATE OR REPLACE FUNCTION detailed_health_report()
RETURNS JSONB AS $$
DECLARE
    v_report JSONB;
BEGIN
    SELECT jsonb_build_object(
        'timestamp', NOW(),
        'database', current_database(),
        'version', version(),
        'connections', (SELECT jsonb_agg(jsonb_build_object(
            'state', state,
            'count', cnt
        )) FROM (SELECT state, count(*) as cnt FROM pg_stat_activity GROUP BY state) s),
        'locks', (SELECT jsonb_agg(jsonb_build_object(
            'blocked_pid', blocked_pid,
            'blocking_pid', blocking_pid
        )) FROM lock_conflicts LIMIT 10),
        'table_bloat', (SELECT jsonb_agg(jsonb_build_object(
            'table', tablename,
            'dead_ratio', dead_tuple_ratio
        )) FROM table_bloat WHERE dead_tuple_ratio > 10 LIMIT 10),
        'slow_queries', (SELECT jsonb_agg(jsonb_build_object(
            'query', query_preview,
            'mean_time', mean_time_ms
        )) FROM slow_queries LIMIT 10),
        'large_tables', (SELECT jsonb_agg(jsonb_build_object(
            'table', object_name,
            'size', total_size
        )) FROM database_size_breakdown WHERE object_type = 'table' LIMIT 20)
    ) INTO v_report;
    
    RETURN v_report;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION detailed_health_report() IS 
'Returns comprehensive health report as JSON for detailed analysis.';

-- ============================================
-- 8. Maintenance Scheduling Helpers
-- ============================================

-- Table to track scheduled maintenance jobs
CREATE TABLE IF NOT EXISTS db_maintenance_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_name VARCHAR(255) NOT NULL UNIQUE,
    job_type VARCHAR(100) NOT NULL, -- 'vacuum', 'reindex', 'analyze', 'refresh_mv', 'retention_cleanup'
    target_table VARCHAR(255),
    schedule_cron VARCHAR(100), -- Cron expression if using pg_cron
    last_run_at TIMESTAMPTZ,
    last_run_duration INTERVAL,
    last_run_status VARCHAR(50), -- 'success', 'failed', 'running'
    last_error TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_maintenance_jobs_enabled ON db_maintenance_jobs(is_enabled, last_run_status) WHERE is_enabled = true;

-- Function to log maintenance job execution
CREATE OR REPLACE FUNCTION log_maintenance_job(
    p_job_name TEXT,
    p_status TEXT,
    p_duration INTERVAL DEFAULT NULL,
    p_error TEXT DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO db_maintenance_jobs (job_name, job_type, last_run_at, last_run_duration, last_run_status, last_error)
    VALUES (p_job_name, 'adhoc', NOW(), p_duration, p_status, p_error)
    ON CONFLICT (job_name) DO UPDATE SET
        last_run_at = NOW(),
        last_run_duration = p_duration,
        last_run_status = p_status,
        last_error = p_error,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to run vacuum on all user tables
CREATE OR REPLACE FUNCTION maintenance_vacuum_all()
RETURNS TABLE (table_name TEXT, status TEXT, duration INTERVAL) AS $$
DECLARE
    v_start TIMESTAMPTZ;
    v_rec RECORD;
BEGIN
    v_start := clock_timestamp();
    
    FOR v_rec IN 
        SELECT relname
        FROM pg_stat_user_tables
        WHERE schemaname = 'public'
        AND n_dead_tup > 1000
        ORDER BY n_dead_tup DESC
        LIMIT 50
    LOOP
        BEGIN
            EXECUTE format('VACUUM ANALYZE %I', v_rec.relname);
            RETURN QUERY SELECT v_rec.relname::TEXT, 'success'::TEXT, clock_timestamp() - v_start;
        EXCEPTION WHEN OTHERS THEN
            RETURN QUERY SELECT v_rec.relname::TEXT, 'failed'::TEXT, NULL::INTERVAL;
        END;
    END LOOP;
    
    PERFORM log_maintenance_job(
        'vacuum_all',
        'success',
        clock_timestamp() - v_start
    );
END;
$$ LANGUAGE plpgsql;

-- Function to reindex bloated indexes
CREATE OR REPLACE FUNCTION maintenance_reindex_bloated()
RETURNS TABLE (index_name TEXT, status TEXT, duration INTERVAL) AS $$
DECLARE
    v_start TIMESTAMPTZ;
    v_rec RECORD;
BEGIN
    v_start := clock_timestamp();
    
    -- Reindex top 10 most bloated indexes
    FOR v_rec IN 
        SELECT schemaname, tablename, indexrelname
        FROM pg_stat_user_indexes
        JOIN pg_class ON pg_class.oid = indexrelid
        WHERE schemaname = 'public'
        AND idx_scan = 0
        ORDER BY pg_relation_size(indexrelid) DESC
        LIMIT 10
    LOOP
        BEGIN
            EXECUTE format('REINDEX INDEX CONCURRENTLY %I', v_rec.indexrelname);
            RETURN QUERY SELECT v_rec.indexrelname::TEXT, 'success'::TEXT, clock_timestamp() - v_start;
        EXCEPTION WHEN OTHERS THEN
            RETURN QUERY SELECT v_rec.indexrelname::TEXT, 'failed'::TEXT, NULL::INTERVAL;
        END;
    END LOOP;
    
    PERFORM log_maintenance_job(
        'reindex_bloated',
        'success',
        clock_timestamp() - v_start
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 9. Comments for Documentation
-- ============================================

COMMENT ON TABLE db_maintenance_jobs IS 
'Tracking table for scheduled database maintenance jobs.';

COMMENT ON FUNCTION maintenance_vacuum_all() IS 
'Run VACUUM ANALYZE on user tables with dead tuples. Run weekly or via cron.';

COMMENT ON FUNCTION maintenance_reindex_bloated() IS 
'Run REINDEX CONCURRENTLY on bloated indexes. Run monthly or via cron.';

COMMENT ON FUNCTION quick_health_check() IS 
'Quick database health check for monitoring dashboards. Returns 5 key metrics.';
