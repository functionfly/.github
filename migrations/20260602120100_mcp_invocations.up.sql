-- 20260602120100_mcp_invocations.up.sql
--
-- Observability for MCP tool invocations. Mirrors the partitioning strategy
-- of registry_function_executions (monthly RANGE partitions).
--
-- Note: the parent table is partitioned; we do NOT add a primary key on the
-- parent because Postgres requires partition keys to be part of the PK. We
-- use a composite identity (function_id, timestamp) and a sequence on the
-- child partition via BIGSERIAL on a per-partition sequence substitute.
--
-- The cron job in internal/observability/partition_maintenance.go (already
-- in the codebase for other partitioned tables) will auto-create next-month
-- partitions and drop partitions older than 90 days.

BEGIN;

CREATE TABLE IF NOT EXISTS registry_mcp_invocations (
    function_id    UUID         NOT NULL,
    -- Server-assigned invocation id (uuid v4). Used as cursor for poll-based
    -- long-running execution recovery (see PR-A.4 tools/call handler).
    invocation_id  UUID         NOT NULL DEFAULT gen_random_uuid(),
    caller_id      TEXT,        -- Hashed caller identity (MCP client user-agent or API key prefix)
    caller_origin  TEXT,        -- Origin URL (for streamable-HTTP CORS audit)
    transport      TEXT         NOT NULL CHECK (transport IN ('streamable-http', 'stdio')),
    method         TEXT         NOT NULL, -- 'tools/call', 'tools/list', etc.
    duration_ms    INTEGER      NOT NULL CHECK (duration_ms >= 0),
    status_code    INTEGER      NOT NULL,  -- HTTP status returned to client (e.g. 200, 401, 429)
    error_code     TEXT,                  -- JSON-RPC error code if applicable
    request_bytes  INTEGER      CHECK (request_bytes IS NULL OR request_bytes >= 0),
    response_bytes INTEGER      CHECK (response_bytes IS NULL OR response_bytes >= 0),
    timestamp      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (function_id, timestamp, invocation_id)
) PARTITION BY RANGE (timestamp);

-- Default partition catches anything before/after explicit monthly partitions.
CREATE TABLE IF NOT EXISTS registry_mcp_invocations_default
    PARTITION OF registry_mcp_invocations DEFAULT;

-- Pre-create the current and next month so writes never land in the default
-- partition (which would break partition pruning for analytics queries).
DO $$
DECLARE
    cur_start   TIMESTAMPTZ := date_trunc('month', now());
    next_start  TIMESTAMPTZ := date_trunc('month', now()) + interval '1 month';
    cur_name    TEXT        := 'registry_mcp_invocations_' || to_char(cur_start,  'YYYYMM');
    next_name   TEXT        := 'registry_mcp_invocations_' || to_char(next_start, 'YYYYMM');
BEGIN
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF registry_mcp_invocations
         FOR VALUES FROM (%L) TO (%L)',
        cur_name, cur_start, next_start
    );
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF registry_mcp_invocations
         FOR VALUES FROM (%L) TO (%L)',
        next_name, next_start, next_start + interval '1 month'
    );
END$$;

CREATE INDEX IF NOT EXISTS idx_mcp_invocations_function_time
    ON registry_mcp_invocations (function_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_invocations_method_time
    ON registry_mcp_invocations (method, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_invocations_invocation_id
    ON registry_mcp_invocations (invocation_id);

COMMIT;
