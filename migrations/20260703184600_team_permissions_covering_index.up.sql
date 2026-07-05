-- Add covering index to team_permissions to eliminate heap fetches
-- The query `SELECT * FROM team_permissions WHERE team_id IN (...)` now uses
-- an index-only scan, avoiding the main table heap entirely.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_permissions_team_id_covering
    ON team_permissions(team_id)
    INCLUDE (id, resource_type, resource_id, permissions, granted_by, granted_at);
