-- Migration: 20260615000001_add_blog_tenant_isolation
-- Description: Add tenant_id to blog_posts and blog_page_views for multi-tenant analytics isolation.
-- This enables tenant-scoped queries on blog analytics instead of full table scans.

-- Step 1: Add tenant_id to blog_posts (nullable for back-compat)
ALTER TABLE blog_posts ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;

-- Step 2: Create function to get tenant_id from post author/owner chain
CREATE OR REPLACE FUNCTION get_post_tenant_id(p_post_id UUID)
RETURNS UUID AS $$
DECLARE
    v_tenant_id UUID;
    v_author_tenant_id UUID;
    v_owner_tenant_id UUID;
BEGIN
    -- Try to get tenant_id directly from blog_posts
    SELECT tenant_id INTO v_tenant_id FROM blog_posts WHERE id = p_post_id;
    IF v_tenant_id IS NOT NULL THEN
        RETURN v_tenant_id;
    END IF;

    -- Fallback: get tenant_id via author -> users -> tenant_id
    SELECT u.tenant_id INTO v_author_tenant_id
    FROM blog_posts bp
    JOIN blog_authors ba ON ba.id = bp.author_id
    JOIN users u ON u.id = ba.id  -- blog_authors share IDs with users (via owner_id)
    WHERE bp.id = p_post_id
    LIMIT 1;

    IF v_author_tenant_id IS NOT NULL THEN
        RETURN v_author_tenant_id;
    END IF;

    -- Fallback: get tenant_id via owner_id
    SELECT u.tenant_id INTO v_owner_tenant_id
    FROM blog_posts bp
    JOIN users u ON u.id = bp.owner_id
    WHERE bp.id = p_post_id
    LIMIT 1;

    RETURN v_owner_tenant_id;
END;
$$ LANGUAGE plpgsql;

-- Step 3: Backfill tenant_id for existing blog_posts
UPDATE blog_posts SET tenant_id = get_post_tenant_id(id) WHERE tenant_id IS NULL;

-- Step 4: Add tenant_id to blog_page_views
ALTER TABLE blog_page_views ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;

-- Step 5: Create trigger function to auto-populate tenant_id on insert
CREATE OR REPLACE FUNCTION set_blog_page_view_tenant_id()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.tenant_id IS NULL THEN
        SELECT tenant_id INTO NEW.tenant_id
        FROM blog_posts
        WHERE id = NEW.post_id;

        -- If still null, try via author chain
        IF NEW.tenant_id IS NULL THEN
            SELECT u.tenant_id INTO NEW.tenant_id
            FROM blog_posts bp
            JOIN blog_authors ba ON ba.id = bp.author_id
            JOIN users u ON u.id = ba.id
            WHERE bp.id = NEW.post_id
            LIMIT 1;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if any and recreate
DROP TRIGGER IF EXISTS trg_blog_page_views_tenant_id ON blog_page_views;
CREATE TRIGGER trg_blog_page_views_tenant_id
    BEFORE INSERT ON blog_page_views
    FOR EACH ROW EXECUTE FUNCTION set_blog_page_view_tenant_id();

-- Step 6: Add indexes with IF NOT EXISTS
CREATE INDEX IF NOT EXISTS idx_blog_posts_tenant_id ON blog_posts(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_blog_page_views_tenant_id ON blog_page_views(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_blog_page_views_tenant_post ON blog_page_views(tenant_id, post_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_blog_page_views_tenant_viewed ON blog_page_views(tenant_id, viewed_at DESC) WHERE tenant_id IS NOT NULL;

-- Step 7: Update the daily analytics trigger to include tenant_id
CREATE OR REPLACE FUNCTION update_blog_analytics_daily_with_tenant()
RETURNS TRIGGER AS $$
DECLARE
    v_date DATE := DATE(NEW.viewed_at);
    v_tenant_id UUID;
BEGIN
    -- Get tenant_id from the page view
    v_tenant_id := NEW.tenant_id;

    -- If tenant_id is still null, try to get it from post
    IF v_tenant_id IS NULL THEN
        SELECT tenant_id INTO v_tenant_id FROM blog_posts WHERE id = NEW.post_id;
    END IF;

    -- Insert or update daily analytics (include tenant_id if available)
    INSERT INTO blog_analytics_daily (post_id, date, views, unique_visitors, tenant_id)
    VALUES (NEW.post_id, v_date, 1, 1, v_tenant_id)
    ON CONFLICT (post_id, date) DO UPDATE SET
        views = blog_analytics_daily.views + 1,
        unique_visitors = blog_analytics_daily.unique_visitors + 1,
        updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_blog_page_views_analytics ON blog_page_views;
CREATE TRIGGER trg_blog_page_views_analytics
AFTER INSERT ON blog_page_views
FOR EACH ROW EXECUTE FUNCTION update_blog_analytics_daily_with_tenant();

-- Step 8: Add tenant_id to blog_analytics_daily (nullable for back-compat)
ALTER TABLE blog_analytics_daily ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;

-- Step 9: Backfill tenant_id on blog_analytics_daily from blog_posts
UPDATE blog_analytics_daily bad
SET tenant_id = bp.tenant_id
FROM blog_analytics_daily bad
JOIN blog_posts bp ON bp.id = bad.post_id
WHERE bad.tenant_id IS NULL AND bp.tenant_id IS NOT NULL;

-- Step 10: Add index on blog_analytics_daily for tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_blog_analytics_daily_tenant ON blog_analytics_daily(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_blog_analytics_daily_tenant_post ON blog_analytics_daily(tenant_id, post_id) WHERE tenant_id IS NOT NULL;

COMMENT ON COLUMN blog_posts.tenant_id IS 'Tenant ID for analytics isolation. Derived from author/owner chain if not explicitly set.';
COMMENT ON COLUMN blog_page_views.tenant_id IS 'Tenant ID populated via trigger from blog_posts. Enables tenant-scoped analytics queries.';
COMMENT ON COLUMN blog_analytics_daily.tenant_id IS 'Denormalized tenant_id for efficient tenant-scoped daily analytics queries.';
