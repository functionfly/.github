-- Migration: Remove content column from blog_posts
-- Reverses: 20260412173800_add_blog_posts_content_column
-- Note: This is a data-destructive operation - content data will be lost

ALTER TABLE blog_posts DROP COLUMN IF EXISTS content;