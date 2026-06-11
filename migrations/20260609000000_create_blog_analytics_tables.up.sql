-- Blog Page Views Analytics Table
-- Tracks individual page views for blog posts

CREATE TABLE IF NOT EXISTS blog_page_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    visitor_id VARCHAR(64),
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country VARCHAR(2),
    city VARCHAR(128),
    device_type VARCHAR(32), -- desktop, mobile, tablet
    browser VARCHAR(64),
    os VARCHAR(64),
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for analytics queries
CREATE INDEX idx_blog_page_views_post_id ON blog_page_views(post_id);
CREATE INDEX idx_blog_page_views_viewed_at ON blog_page_views(viewed_at);
CREATE INDEX idx_blog_page_views_post_viewed ON blog_page_views(post_id, viewed_at);

-- Blog Analytics Daily Aggregates
-- Pre-computed daily view counts per post for fast dashboard queries

CREATE TABLE IF NOT EXISTS blog_analytics_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views INTEGER NOT NULL DEFAULT 0,
    unique_visitors INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(post_id, date)
);

CREATE INDEX idx_blog_analytics_daily_post_id ON blog_analytics_daily(post_id);
CREATE INDEX idx_blog_analytics_daily_date ON blog_analytics_daily(date);

-- Function to update daily analytics
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

-- Trigger to automatically update daily analytics
DROP TRIGGER IF EXISTS trg_blog_page_views_analytics ON blog_page_views;
CREATE TRIGGER trg_blog_page_views_analytics
AFTER INSERT ON blog_page_views
FOR EACH ROW EXECUTE FUNCTION update_blog_analytics_daily();