-- Add Follow System tables
-- This migration adds support for user-to-user and user-to-function follows

-- User Follows table: allows users to follow other users
CREATE TABLE IF NOT EXISTS user_follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Optional: reason for following (e.g., "likes their functions", "colleague")
    follow_reason VARCHAR(255),
    
    -- Notifications preferences for this specific follow
    notify_on_new_function BOOLEAN DEFAULT true,
    notify_on_function_update BOOLEAN DEFAULT true,
    notify_on_new_version BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint: a user can only follow another user once
    CONSTRAINT uq_user_follow UNIQUE (follower_id, followed_user_id),
    
    -- Prevent self-follows
    CONSTRAINT ck_no_self_follow CHECK (follower_id != followed_user_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_followed ON user_follows(followed_user_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_created ON user_follows(created_at DESC);

-- Composite index for checking if user follows
CREATE INDEX IF NOT EXISTS idx_user_follows_lookup ON user_follows(follower_id, followed_user_id);

-- Function Follows table: allows users to follow specific functions
CREATE TABLE IF NOT EXISTS function_follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    
    -- Optional: reason for following
    follow_reason VARCHAR(255),
    
    -- Notifications preferences
    notify_on_new_version BOOLEAN DEFAULT true,
    notify_on_rating_change BOOLEAN DEFAULT false,
    notify_on_trust_change BOOLEAN DEFAULT true,
    notify_on_verification BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint: a user can only follow a function once
    CONSTRAINT uq_function_follow UNIQUE (user_id, function_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_function_follows_user ON function_follows(user_id);
CREATE INDEX IF NOT EXISTS idx_function_follows_function ON function_follows(function_id);
CREATE INDEX IF NOT EXISTS idx_function_follows_created ON function_follows(created_at DESC);

-- Cache for user follower/following counts
CREATE TABLE IF NOT EXISTS user_follow_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Cache for function follower counts
CREATE TABLE IF NOT EXISTS function_follow_stats (
    function_id UUID PRIMARY KEY REFERENCES registry_functions(id) ON DELETE CASCADE,
    followers_count INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
