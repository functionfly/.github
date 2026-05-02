-- Migration: Remove additional blog categories
-- Reverses: 20260412175400_add_more_blog_categories
-- Note: Only removes categories that have no associated posts

DELETE FROM blog_categories WHERE slug IN (
    'ai-ml', 'devops', 'api-sdks', 'releases', 'case-studies'
) AND NOT EXISTS (
    SELECT 1 FROM blog_posts WHERE blog_posts.category_id = blog_categories.id
);