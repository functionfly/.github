-- Migration: Remove default blog categories
-- Reverses: 20260412175300_seed_default_blog_categories
-- Note: Only removes categories that have no associated posts

DELETE FROM blog_categories WHERE slug IN (
    'engineering', 'product', 'company-news', 'tutorials', 'community', 'security'
) AND NOT EXISTS (
    SELECT 1 FROM blog_posts WHERE blog_posts.category_id = blog_categories.id
);