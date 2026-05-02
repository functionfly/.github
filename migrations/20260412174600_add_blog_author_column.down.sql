-- Migration: Remove author column from blog_posts
-- Reverses: 20260412174600_add_blog_author_column
-- Note: Data will be lost

ALTER TABLE blog_posts DROP COLUMN IF EXISTS author;