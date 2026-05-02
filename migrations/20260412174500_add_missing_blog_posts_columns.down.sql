-- Migration: Remove added columns from blog_posts
-- Reverses: 20260412174500_add_missing_blog_posts_columns
-- Note: Data in these columns will be lost

ALTER TABLE blog_posts DROP COLUMN IF EXISTS excerpt;
ALTER TABLE blog_posts DROP COLUMN IF EXISTS featured_image;
ALTER TABLE blog_posts DROP COLUMN IF EXISTS sanity_id;
ALTER TABLE blog_posts DROP COLUMN IF EXISTS is_published;
ALTER TABLE blog_posts DROP COLUMN IF EXISTS published_at;
ALTER TABLE blog_posts DROP COLUMN IF EXISTS tags;