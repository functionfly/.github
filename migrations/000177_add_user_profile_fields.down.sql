-- Migration: Revert user profile fields

-- Drop indexes
DROP INDEX IF EXISTS idx_user_skills_user_id;
DROP INDEX IF EXISTS idx_user_achievements_user_id;
DROP INDEX IF EXISTS idx_user_achievements_achievement_id;
DROP INDEX IF EXISTS idx_user_activity_user_id;
DROP INDEX IF EXISTS idx_user_activity_created_at;
DROP INDEX IF EXISTS idx_user_activity_type;
DROP INDEX IF EXISTS idx_achievements_category;
DROP INDEX IF EXISTS idx_achievements_slug;

-- Drop tables (order matters for foreign keys)
DROP TABLE IF EXISTS user_activity;
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS user_skills;
DROP TABLE IF EXISTS achievements;

-- Drop columns from users table
ALTER TABLE users
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS website,
    DROP COLUMN IF EXISTS job_title,
    DROP COLUMN IF EXISTS social_links,
    DROP COLUMN IF EXISTS twitter_url,
    DROP COLUMN IF EXISTS github_url,
    DROP COLUMN IF EXISTS linkedin_url,
    DROP COLUMN IF EXISTS cover_image_url;
