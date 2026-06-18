-- Rollback: 20260615000001_add_blog_tenant_isolation
-- Removes tenant_id columns added for multi-tenant analytics isolation

-- Remove indexes
DROP INDEX IF EXISTS idx_blog_posts_tenant_id;
DROP INDEX IF EXISTS idx_blog_page_views_tenant_id;
DROP INDEX IF EXISTS idx_blog_page_views_tenant_post;
DROP INDEX IF EXISTS idx_blog_page_views_tenant_viewed;
DROP INDEX IF EXISTS idx_blog_analytics_daily_tenant;
DROP INDEX IF EXISTS idx_blog_analytics_daily_tenant_post;

-- Remove tenant_id columns
ALTER TABLE blog_posts DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE blog_page_views DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE blog_analytics_daily DROP COLUMN IF EXISTS tenant_id;

-- Remove functions
DROP FUNCTION IF EXISTS get_post_tenant_id(UUID);
DROP FUNCTION IF EXISTS set_blog_page_view_tenant_id();
DROP FUNCTION IF EXISTS update_blog_analytics_daily_with_tenant();

-- Restore original trigger (without tenant_id handling)
DROP TRIGGER IF EXISTS trg_blog_page_views_analytics ON blog_page_views;
CREATE OR REPLACE FUNCTION update_blog_analytics_daily()
RETURNS TRIGGER AS $$
DECLARE
    v_date DATE := DATE(NEW.viewed_at);
BEGIN
    INSERT INTO blog_analytics_daily (post_id, date, views, unique_visitors)
    VALUES (NEW.post_id, v_date, 1, 1)
    ON CONFLICT (post_id, date) DO UPDATE SET
        views = blog_analytics_daily.views + 1,
        unique_visitors = blog_analytics_daily.unique_visitors + 1,
        updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_blog_page_views_analytics
AFTER INSERT ON blog_page_views
FOR EACH ROW EXECUTE FUNCTION update_blog_analytics_daily();
