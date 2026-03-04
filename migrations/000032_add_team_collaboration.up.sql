-- Migration: Add team collaboration features
-- Created: 2026-02-19

-- Create teams table
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create team_memberships table
CREATE TABLE IF NOT EXISTS team_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    added_by UUID NOT NULL REFERENCES users(id),
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(team_id, user_id)
);

-- Create team_permissions table
CREATE TABLE IF NOT EXISTS team_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL, -- 'app', 'function', 'backend', 'deployment'
    resource_id UUID NOT NULL,
    permissions TEXT NOT NULL, -- JSON array of permissions like '["read", "write", "delete"]'
    granted_by UUID NOT NULL REFERENCES users(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(team_id, resource_type, resource_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_teams_tenant_id ON teams(tenant_id);
CREATE INDEX IF NOT EXISTS idx_teams_created_by ON teams(created_by);
CREATE INDEX IF NOT EXISTS idx_team_memberships_team_id ON team_memberships(team_id);
CREATE INDEX IF NOT EXISTS idx_team_memberships_user_id ON team_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_team_memberships_role ON team_memberships(role);
CREATE INDEX IF NOT EXISTS idx_team_permissions_team_id ON team_permissions(team_id);
CREATE INDEX IF NOT EXISTS idx_team_permissions_resource ON team_permissions(resource_type, resource_id);

-- Update users table to include teams relationship (already done via GORM tags)

-- Create pg_notification_channels table if it doesn't exist
CREATE TABLE IF NOT EXISTS pg_notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add real-time notification channels for team events
-- These will be used by the WebSocket system for broadcasting team updates
INSERT INTO pg_notification_channels (channel_name, description) VALUES
('team_updates', 'Team membership and permission changes'),
('team_member_joined', 'New team member notifications'),
('team_member_left', 'Team member removal notifications'),
('team_permission_granted', 'Permission grant notifications'),
('team_permission_revoked', 'Permission revocation notifications')
ON CONFLICT (channel_name) DO NOTHING;