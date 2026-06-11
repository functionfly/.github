-- Migration rollback: Remove team collaboration features
-- Created: 2026-02-19

-- Remove real-time notification channels
DELETE FROM pg_notification_channels WHERE channel_name IN (
    'team_updates',
    'team_member_joined',
    'team_member_left',
    'team_permission_granted',
    'team_permission_revoked'
);

-- Drop indexes
DROP INDEX IF EXISTS idx_team_permissions_resource;
DROP INDEX IF EXISTS idx_team_permissions_team_id;
DROP INDEX IF EXISTS idx_team_memberships_role;
DROP INDEX IF EXISTS idx_team_memberships_user_id;
DROP INDEX IF EXISTS idx_team_memberships_team_id;
DROP INDEX IF EXISTS idx_teams_created_by;
DROP INDEX IF EXISTS idx_teams_tenant_id;

-- Drop tables in reverse order (due to foreign key constraints)
DROP TABLE IF EXISTS team_permissions;
DROP TABLE IF EXISTS team_memberships;
DROP TABLE IF EXISTS teams;