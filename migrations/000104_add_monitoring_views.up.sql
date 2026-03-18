-- Add database performance monitoring views and metrics
-- These views provide insights into database performance, usage patterns, and potential issues

-- Table size and growth monitoring
CREATE OR REPLACE VIEW db_table_sizes AS
SELECT
    schemaname,
    relname as tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||relname)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname) - pg_relation_size(schemaname||'.'||relname)) as index_size,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_rows,
    n_dead_tup as dead_rows,
    CASE WHEN n_live_tup > 0 THEN round(((n_dead_tup::float / n_live_tup::float) * 100)::numeric, 2) ELSE 0 END as dead_tuple_ratio,
    NULL::timestamp as last_vacuum, -- Not available in this PostgreSQL version
    NULL::timestamp as last_autovacuum,
    NULL::timestamp as last_analyze,
    NULL::timestamp as last_autoanalyze,
    0 as vacuum_count, -- Not available in this PostgreSQL version
    0 as autovacuum_count,
    0 as analyze_count,
    0 as autoanalyze_count
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||relname) DESC;

-- Index usage and efficiency
CREATE OR REPLACE VIEW db_index_usage AS
SELECT
    schemaname,
    relname as tablename,
    indexrelname as indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read_via_index,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(schemaname||'.'||indexrelname)) as index_size,
    CASE WHEN idx_scan > 0 THEN round((idx_tup_read::float / idx_scan::float)::numeric, 2) ELSE 0 END as avg_tuples_per_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC, pg_relation_size(schemaname||'.'||indexrelname) DESC;

-- Unused indexes (potential cleanup candidates)
CREATE OR REPLACE VIEW db_unused_indexes AS
SELECT
    schemaname,
    relname as tablename,
    indexrelname as indexname,
    idx_scan as scans,
    pg_size_pretty(pg_relation_size(schemaname||'.'||indexrelname)) as size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(schemaname||'.'||indexrelname) DESC;

-- Query performance monitoring (DROP first so column types can change, e.g. integer -> bigint)
DROP VIEW IF EXISTS db_query_performance CASCADE;
DO $$
BEGIN
    BEGIN
        EXECUTE $view$
            CREATE VIEW db_query_performance AS
            SELECT
                query,
                calls,
                total_exec_time as total_time,
                mean_exec_time as mean_time,
                rows,
                round((total_exec_time / calls)::numeric, 2) as avg_time_per_call,
                round((rows::float / calls)::numeric, 2) as avg_rows_per_call,
                0::bigint as temp_blks_read,
                0::bigint as temp_blks_written,
                0::double precision as blk_read_time,
                0::double precision as blk_write_time
            FROM pg_stat_statements
            WHERE calls > 10
            ORDER BY mean_exec_time DESC
            LIMIT 50
        $view$;
    EXCEPTION WHEN undefined_table THEN
        EXECUTE $view$
            CREATE VIEW db_query_performance AS
            SELECT
                NULL::text AS query,
                0::bigint AS calls,
                0::double precision AS total_time,
                0::double precision AS mean_time,
                0::bigint AS rows,
                0::numeric AS avg_time_per_call,
                0::numeric AS avg_rows_per_call,
                0::bigint AS temp_blks_read,
                0::bigint AS temp_blks_written,
                0::double precision AS blk_read_time,
                0::double precision AS blk_write_time
            WHERE FALSE
        $view$;
    END;
END $$;

-- Connection and session monitoring
CREATE OR REPLACE VIEW db_connection_stats AS
SELECT
    datname as database,
    usename as username,
    client_addr,
    client_port,
    backend_start,
    query_start,
    state_change,
    state,
    CASE
        WHEN state = 'active' THEN extract(epoch from (now() - query_start))
        WHEN state = 'idle in transaction' THEN extract(epoch from (now() - state_change))
        ELSE NULL
    END as duration_seconds
FROM pg_stat_activity
WHERE pid <> pg_backend_pid()
ORDER BY backend_start;

-- Lock monitoring
CREATE OR REPLACE VIEW db_lock_monitoring AS
SELECT
    blocked_locks.pid as blocked_pid,
    blocked_activity.usename as blocked_user,
    blocked_activity.query as blocked_query,
    blocked_activity.state as blocked_state,
    blocking_locks.pid as blocking_pid,
    blocking_activity.usename as blocking_user,
    blocking_activity.query as blocking_query,
    blocked_locks.locktype,
    blocked_locks.mode as blocked_mode,
    blocking_locks.mode as blocking_mode,
    extract(epoch from (now() - blocked_activity.query_start)) as blocked_duration_seconds
FROM pg_locks blocked_locks
JOIN pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_locks blocking_locks
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database = blocked_locks.database
    AND blocking_locks.relation = blocked_locks.relation
    AND blocking_locks.page = blocked_locks.page
    AND blocking_locks.tuple = blocked_locks.tuple
    AND blocking_locks.virtualxid = blocked_locks.virtualxid
    AND blocking_locks.transactionid = blocked_locks.transactionid
    AND blocking_locks.classid = blocked_locks.classid
    AND blocking_locks.objid = blocked_locks.objid
    AND blocking_locks.objsubid = blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Database growth trends (requires historical data collection)
CREATE OR REPLACE VIEW db_growth_trends AS
SELECT
    schemaname,
    relname as tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) as current_size,
    n_tup_ins as total_inserts,
    n_tup_upd as total_updates,
    n_tup_del as total_deletes,
    n_live_tup as current_live_rows,
    CASE WHEN n_tup_ins > 0 THEN
        round(((n_live_tup::float / n_tup_ins::float) * 100)::numeric, 2)
    ELSE 0 END as data_retention_ratio
FROM pg_stat_user_tables
WHERE n_tup_ins > 1000  -- Only show tables with significant activity
ORDER BY n_tup_ins DESC;

-- Tenant-specific usage patterns
CREATE OR REPLACE VIEW tenant_usage_summary AS
SELECT
    t.id as tenant_id,
    t.name as tenant_name,
    COALESCE(t.plan, 'free') as plan,
    COUNT(DISTINCT u.id) as user_count,
    COUNT(DISTINCT a.id) as app_count,
    COUNT(DISTINCT d.id) as deployment_count,
    COALESCE(SUM(ue.quantity), 0) as total_usage_events,
    MAX(ue.timestamp) as last_activity,
    pg_size_pretty(0::bigint) as estimated_tenant_size -- Placeholder for now
FROM tenants t
LEFT JOIN users u ON u.tenant_id = t.id
LEFT JOIN apps a ON a.tenant_id = t.id
LEFT JOIN deployments d ON d.app_id = a.id
LEFT JOIN usage_events ue ON ue.tenant_id = t.id
GROUP BY t.id, t.name, t.plan
ORDER BY total_usage_events DESC;

-- Replication lag monitoring (for replicas)
CREATE OR REPLACE VIEW replication_status AS
SELECT
    client_addr,
    usename,
    application_name,
    state,
    sync_state,
    pg_wal_lsn_diff(pg_current_wal_lsn(), sent_lsn) as sent_lag_bytes,
    pg_wal_lsn_diff(sent_lsn, flush_lsn) as flush_lag_bytes,
    pg_wal_lsn_diff(flush_lsn, replay_lsn) as replay_lag_bytes,
    extract(epoch from (now() - write_lag)) as write_lag_seconds,
    extract(epoch from (now() - flush_lag)) as flush_lag_seconds,
    extract(epoch from (now() - replay_lag)) as replay_lag_seconds
FROM pg_stat_replication;

-- Database health score (composite metric) - simplified version
CREATE OR REPLACE VIEW db_health_score AS
WITH metrics AS (
    SELECT
        COALESCE((SELECT count(*) FROM db_unused_indexes), 0) as unused_indexes,
        COALESCE((SELECT avg(dead_tuple_ratio) FROM db_table_sizes WHERE dead_tuple_ratio > 20), 0) as avg_dead_tuple_ratio,
        COALESCE((SELECT count(*) FROM db_lock_monitoring), 0) as blocking_locks,
        COALESCE((SELECT count(*) FROM db_connection_stats WHERE duration_seconds > 300), 0) as long_running_queries,
        (SELECT count(*) FROM pg_stat_activity WHERE state = 'idle in transaction') as idle_in_transaction
)
SELECT
    CASE
        WHEN unused_indexes > 10 THEN 30
        WHEN unused_indexes > 5 THEN 20
        ELSE 10
    END +
    CASE
        WHEN avg_dead_tuple_ratio > 50 THEN 30
        WHEN avg_dead_tuple_ratio > 20 THEN 20
        ELSE 10
    END +
    CASE
        WHEN blocking_locks > 0 THEN 30
        ELSE 10
    END +
    CASE
        WHEN long_running_queries > 5 THEN 30
        WHEN long_running_queries > 2 THEN 20
        ELSE 10
    END +
    CASE
        WHEN idle_in_transaction > 10 THEN 30
        WHEN idle_in_transaction > 5 THEN 20
        ELSE 10
    END as health_score,
    unused_indexes,
    round(avg_dead_tuple_ratio::numeric, 2) as avg_dead_tuple_ratio,
    blocking_locks,
    long_running_queries,
    idle_in_transaction
FROM metrics;

-- Create a function to collect historical database metrics
CREATE OR REPLACE FUNCTION collect_database_metrics()
RETURNS void AS $$
BEGIN
    -- Insert current database size metrics
    INSERT INTO database_metrics (metric_type, value, unit, recorded_at, created_at)
    SELECT
        'database_size_gb',
        round(pg_database_size(current_database()) / 1024.0 / 1024.0 / 1024.0, 2),
        'gb',
        now(),
        now();

    -- Insert connection count
    INSERT INTO database_metrics (metric_type, value, unit, recorded_at, created_at)
    SELECT
        'active_connections',
        count(*),
        'count',
        now(),
        now()
    FROM pg_stat_activity;

    -- Insert cache hit ratio
    INSERT INTO database_metrics (metric_type, value, unit, recorded_at, created_at)
    SELECT
        'cache_hit_ratio',
        round(sum(blks_hit) / (sum(blks_hit) + sum(blks_read)) * 100, 2),
        'ratio',
        now(),
        now()
    FROM pg_stat_database
    WHERE datname = current_database();

    -- Insert transaction rate (per minute)
    INSERT INTO database_metrics (metric_type, value, unit, recorded_at, created_at)
    SELECT
        'transactions_per_minute',
        round((sum(xact_commit) + sum(xact_rollback)) / extract(epoch from (now() - stats_reset)) * 60, 2),
        'tpm',
        now(),
        now()
    FROM pg_stat_database
    WHERE datname = current_database();
END;
$$ LANGUAGE plpgsql;

-- Create a function to get database performance recommendations
CREATE OR REPLACE FUNCTION get_db_recommendations()
RETURNS TABLE(recommendation text, severity text, details text) AS $$
BEGIN
    -- Unused indexes
    RETURN QUERY
    SELECT
        'Remove unused indexes'::text,
        CASE WHEN count(*) > 10 THEN 'high' WHEN count(*) > 5 THEN 'medium' ELSE 'low' END,
        'Found ' || count(*) || ' unused indexes consuming space'::text
    FROM db_unused_indexes;

    -- High dead tuple ratio
    RETURN QUERY
    SELECT
        'Run VACUUM on bloated tables'::text,
        CASE WHEN avg(dead_tuple_ratio) > 50 THEN 'high' WHEN avg(dead_tuple_ratio) > 20 THEN 'medium' ELSE 'low' END,
        'Average dead tuple ratio is ' || round(avg(dead_tuple_ratio), 1) || '%'::text
    FROM db_table_sizes
    WHERE dead_tuple_ratio > 10;

    -- Long running queries
    RETURN QUERY
    SELECT
        'Review long-running queries'::text,
        'medium'::text,
        'Found ' || count(*) || ' queries running longer than 5 minutes'::text
    FROM db_connection_stats
    WHERE duration_seconds > 300;

    -- Blocking locks
    RETURN QUERY
    SELECT
        'Resolve blocking locks'::text,
        'high'::text,
        'Found ' || count(*) || ' blocking lock situations'::text
    FROM db_lock_monitoring;

    -- Connection pool saturation
    RETURN QUERY
    SELECT
        'Increase connection pool size'::text,
        'medium'::text,
        'High connection utilization detected'::text
    WHERE (SELECT count(*) FROM db_connection_stats) > 80; -- Assuming 80% of max_connections
END;
$$ LANGUAGE plpgsql;
