-- Add missing user profile fields to users table
-- These fields are used in the user profile API endpoints

ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS location VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS website VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS job_title VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS social_links JSONB DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS twitter_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS github_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS linkedin_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS cover_image_url VARCHAR(500);

-- Add indexes for better performance on profile queries
CREATE INDEX IF NOT EXISTS idx_users_bio ON users(bio);
CREATE INDEX IF NOT EXISTS idx_users_location ON users(location);
CREATE INDEX IF NOT EXISTS idx_users_job_title ON users(job_title);
CREATE INDEX IF NOT EXISTS idx_users_social_links ON users USING GIN (social_links);

-- Add comments for documentation
COMMENT ON COLUMN users.bio IS 'User biography/description text';
COMMENT ON COLUMN users.location IS 'User location information';
COMMENT ON COLUMN users.website IS 'User personal/professional website URL';
COMMENT ON COLUMN users.job_title IS 'User job title/position';
COMMENT ON COLUMN users.social_links IS 'Additional social media links as JSON';
COMMENT ON COLUMN users.twitter_url IS 'Twitter profile URL';
COMMENT ON COLUMN users.github_url IS 'GitHub profile URL';
COMMENT ON COLUMN users.linkedin_url IS 'LinkedIn profile URL';
COMMENT ON COLUMN users.cover_image_url IS 'Cover image URL for profile';