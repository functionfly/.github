-- Migration: Add user profile fields for extended profile support
-- Created: 2026-03-02

-- Add new columns to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS location VARCHAR(255),
    ADD COLUMN IF NOT EXISTS website VARCHAR(500),
    ADD COLUMN IF NOT EXISTS job_title VARCHAR(255),
    ADD COLUMN IF NOT EXISTS social_links JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS twitter_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS github_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS linkedin_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS cover_image_url VARCHAR(500);

-- Create user_skills table
CREATE TABLE IF NOT EXISTS user_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    level VARCHAR(20) NOT NULL DEFAULT 'intermediate', -- 'beginner', 'intermediate', 'advanced', 'expert'
    category VARCHAR(50), -- 'language', 'framework', 'tool', 'platform', 'soft'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

-- Create achievements table (definitions)
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    icon VARCHAR(100), -- Icon name or URL
    color VARCHAR(20), -- Badge color theme
    category VARCHAR(50) NOT NULL, -- 'publisher', 'community', 'usage', 'milestone'
    requirement_type VARCHAR(50) NOT NULL, -- 'function_count', 'execution_count', 'rating', 'days_active', etc.
    requirement_value INTEGER NOT NULL, -- Threshold value
    points INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT FALSE, -- Secret achievements
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create user_achievements table (earned achievements)
CREATE TABLE IF NOT EXISTS user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    progress INTEGER DEFAULT 0, -- Current progress toward the achievement
    is_completed BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}', -- Additional context (e.g., which function triggered it)
    UNIQUE(user_id, achievement_id)
);

-- Create user_activity table (activity feed)
CREATE TABLE IF NOT EXISTS user_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL, -- 'function_published', 'function_updated', 'badge_earned', 'profile_updated', etc.
    title VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}', -- Additional details (function_id, version, etc.)
    is_public BOOLEAN DEFAULT TRUE, -- Whether this activity is visible on public profile
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_skills_user_id ON user_skills(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_achievement_id ON user_achievements(achievement_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_user_id ON user_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_created_at ON user_activity(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_activity_type ON user_activity(activity_type);
CREATE INDEX IF NOT EXISTS idx_achievements_category ON achievements(category);
CREATE INDEX IF NOT EXISTS idx_achievements_slug ON achievements(slug);

-- Enable RLS on new tables
ALTER TABLE user_skills ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_achievements ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_activity ENABLE ROW LEVEL SECURITY;
ALTER TABLE achievements ENABLE ROW LEVEL SECURITY;

-- RLS Policies for user_skills
CREATE POLICY user_skills_owner_access ON user_skills
    FOR ALL USING (user_id = current_user_id());

CREATE POLICY user_skills_public_read ON user_skills
    FOR SELECT USING (true); -- Public for profile viewing

-- RLS Policies for user_achievements
CREATE POLICY user_achievements_owner_access ON user_achievements
    FOR ALL USING (user_id = current_user_id());

CREATE POLICY user_achievements_public_read ON user_achievements
    FOR SELECT USING (true); -- Public for profile viewing

-- RLS Policies for user_activity
CREATE POLICY user_activity_owner_access ON user_activity
    FOR ALL USING (user_id = current_user_id());

CREATE POLICY user_activity_public_read ON user_activity
    FOR SELECT USING (is_public = true); -- Only public activities visible

-- RLS Policies for achievements (public read-only)
CREATE POLICY achievements_public_read ON achievements
    FOR SELECT USING (true);

-- Insert default achievements
INSERT INTO achievements (slug, name, description, icon, color, category, requirement_type, requirement_value, points) VALUES
    ('first_function', 'First Function', 'Published your first function to the registry', 'Zap', 'blue', 'publisher', 'function_count', 1, 10),
    ('function_10', 'Function Publisher', 'Published 10 functions', 'Package', 'blue', 'publisher', 'function_count', 10, 50),
    ('function_50', 'Function Factory', 'Published 50 functions', 'Boxes', 'blue', 'publisher', 'function_count', 50, 200),
    ('function_100', 'Function Master', 'Published 100 functions', 'Crown', 'purple', 'publisher', 'function_count', 100, 500),
    ('execution_1k', 'Rising Star', 'Reached 1,000 total executions', 'TrendingUp', 'green', 'usage', 'execution_count', 1000, 25),
    ('execution_10k', 'Popular Choice', 'Reached 10,000 total executions', 'BarChart3', 'green', 'usage', 'execution_count', 10000, 100),
    ('execution_100k', 'Heavy Hitter', 'Reached 100,000 total executions', 'Activity', 'green', 'usage', 'execution_count', 100000, 500),
    ('execution_1m', 'Millionaire', 'Reached 1,000,000 total executions', 'Star', 'gold', 'usage', 'execution_count', 1000000, 2000),
    ('rating_4_5', 'Highly Rated', 'Maintained a 4.5+ average rating', 'Heart', 'pink', 'community', 'rating', 45, 100),
    ('rating_5', 'Perfect Score', 'Achieved a perfect 5.0 rating', 'Award', 'gold', 'community', 'rating', 50, 500),
    ('member_30', 'Regular', 'Active member for 30 days', 'Calendar', 'blue', 'milestone', 'days_active', 30, 25),
    ('member_365', 'Veteran', 'Active member for 1 year', 'Calendar', 'purple', 'milestone', 'days_active', 365, 200),
    ('early_adopter', 'Early Adopter', 'Joined during the beta period', 'Rocket', 'orange', 'milestone', 'early_adopter', 1, 100)
ON CONFLICT (slug) DO NOTHING;

-- Comments
COMMENT ON TABLE user_skills IS 'User skills and expertise areas';
COMMENT ON TABLE achievements IS 'Achievement/badge definitions';
COMMENT ON TABLE user_achievements IS 'User-earned achievements with progress tracking';
COMMENT ON TABLE user_activity IS 'User activity feed for profile timeline';
