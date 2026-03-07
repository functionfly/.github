DROP TRIGGER IF EXISTS update_blog_authors_updated_at ON blog_authors;
DROP TRIGGER IF EXISTS update_blog_categories_updated_at ON blog_categories;
DROP INDEX IF EXISTS idx_blog_authors_active;
DROP INDEX IF EXISTS idx_blog_authors_slug;
DROP INDEX IF EXISTS idx_blog_categories_order;
DROP INDEX IF EXISTS idx_blog_categories_slug;
DROP TABLE IF EXISTS blog_authors;
DROP TABLE IF EXISTS blog_categories;
