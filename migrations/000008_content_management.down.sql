-- Drop Content Management System tables
DROP TRIGGER IF EXISTS update_blog_posts_updated_at ON blog_posts;
DROP TRIGGER IF EXISTS update_changelog_changes_updated_at ON changelog_changes;
DROP TRIGGER IF EXISTS update_changelog_entries_updated_at ON changelog_entries;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_blog_posts_tags;
DROP INDEX IF EXISTS idx_blog_posts_published_at;
DROP INDEX IF EXISTS idx_blog_posts_published;
DROP INDEX IF EXISTS idx_blog_posts_slug;
DROP INDEX IF EXISTS idx_changelog_changes_entry_id;
DROP INDEX IF EXISTS idx_changelog_entries_date;
DROP INDEX IF EXISTS idx_changelog_entries_published;
DROP INDEX IF EXISTS idx_changelog_entries_version;

DROP TABLE IF EXISTS blog_posts;
DROP TABLE IF EXISTS changelog_changes;
DROP TABLE IF EXISTS changelog_entries;