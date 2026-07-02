DROP INDEX IF EXISTS idx_community_posts_slug;
ALTER TABLE community_posts DROP COLUMN IF EXISTS slug;
