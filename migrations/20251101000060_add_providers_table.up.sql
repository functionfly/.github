-- Migration: Add providers table for user-connected cloud providers (Vercel, Fly, Cloudflare, etc.)
-- Created: 2026-02-27

CREATE TABLE IF NOT EXISTS providers (
    id VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    is_shared BOOLEAN NOT NULL DEFAULT false,
    team_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_providers_user_id ON providers(user_id);
CREATE INDEX IF NOT EXISTS idx_providers_provider ON providers(provider);
CREATE INDEX IF NOT EXISTS idx_providers_team_id ON providers(team_id);
CREATE INDEX IF NOT EXISTS idx_providers_status ON providers(status);
