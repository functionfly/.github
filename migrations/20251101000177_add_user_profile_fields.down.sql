-- Remove user profile fields from users table
-- Reverse migration for 20260303000003_add_user_profile_fields.up.sql

DROP INDEX IF EXISTS idx_users_social_links;
DROP INDEX IF EXISTS idx_users_job_title;
DROP INDEX IF EXISTS idx_users_location;
DROP INDEX IF EXISTS idx_users_bio;

ALTER TABLE users DROP COLUMN IF EXISTS cover_image_url;
ALTER TABLE users DROP COLUMN IF EXISTS linkedin_url;
ALTER TABLE users DROP COLUMN IF EXISTS github_url;
ALTER TABLE users DROP COLUMN IF EXISTS twitter_url;
ALTER TABLE users DROP COLUMN IF EXISTS social_links;
ALTER TABLE users DROP COLUMN IF EXISTS job_title;
ALTER TABLE users DROP COLUMN IF EXISTS website;
ALTER TABLE users DROP COLUMN IF EXISTS location;
ALTER TABLE users DROP COLUMN IF EXISTS bio;