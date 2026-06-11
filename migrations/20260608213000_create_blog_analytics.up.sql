-- Create blog page views tracking table
CREATE TABLE IF NOT EXISTS blog_page_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    visitor_id VARCHAR(255),
    referrer TEXT,
    user_agent TEXT,
    ip_address INET,
    country VARCHAR(100),
    city VARCHAR(255),
    device_type VARCHAR(50),
    browser VARCHAR(100),
    os VARCHAR(100),
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for analytics queries
CREATE INDEX idx_blog_page_views_post_id ON blog_page_views(post_id);
CREATE INDEX idx_blog_page_views_viewed_at ON blog_page_views(viewed_at);
CREATE INDEX idx_blog_page_views_post_viewed ON blog_page_views(post_id, viewed_at);

-- Blog daily analytics aggregate table for fast dashboard queries
CREATE TABLE IF NOT EXISTS blog_daily_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(post_id, date)
);

CREATE INDEX idx_blog_daily_analytics_post_date ON blog_daily_analytics(post_id, date);
CREATE INDEX idx_blog_daily_analytics_date ON blog_daily_analytics(date);

-- Function to record a page view
CREATE OR REPLACE FUNCTION record_blog_page_view(
    p_post_id UUID,
    p_visitor_id VARCHAR(255),
    p_referrer TEXT,
    p_user_agent TEXT,
    p_ip_address INET,
    p_country VARCHAR(100),
    p_city VARCHAR(255),
    p_device_type VARCHAR(50),
    p_browser VARCHAR(100),
    p_os VARCHAR(100)
) RETURNS void AS $$
DECLARE
    v_today DATE := CURRENT_DATE;
BEGIN
    -- Insert individual page view record
    INSERT INTO blog_page_views (post_id, visitor_id, referrer, user_agent, ip_address, country, city, device_type, browser, os)
    VALUES (p_post_id, p_visitor_id, p_referrer, p_user_agent, p_ip_address, p_country, p_city, p_device_type, p_browser, p_os);

    -- Update daily analytics
    INSERT INTO blog_daily_analytics (post_id, date, views, unique_visitors)
    VALUES (p_post_id, v_today, 1, 1)
    ON CONFLICT (post_id, date)
    DO UPDATE SET
        views = blog_daily_analytics.views + 1,
        unique_visitors = CASE
            WHEN p_visitor_id IS NOT NULL THEN
                (SELECT COUNT(DISTINCT visitor_id) FROM blog_page_views WHERE post_id = p_post_id AND DATE(viewed_at) = v_today)
            ELSE blog_daily_analytics.unique_visitors
        END,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- View for top posts by views
CREATE OR REPLACE VIEW blog_top_posts_view AS
SELECT
    bp.id,
    bp.title,
    bp.slug,
    bp.author,
    bp.published_at,
    COALESCE(SUM(bda.views), 0) as total_views,
    COUNT(DISTINCT bpv.visitor_id) as unique_visitors,
    MAX(bpv.viewed_at) as last_viewed_at
FROM blog_posts bp
LEFT JOIN blog_daily_analytics bda ON bp.id = bda.post_id
LEFT JOIN blog_page_views bpv ON bp.id = bpv.post_id
GROUP BY bp.id, bp.title, bp.slug, bp.author, bp.published_at
ORDER BY total_views DESC;

-- Analytics summary function
CREATE OR REPLACE FUNCTION get_blog_analytics_summary(
    p_days INTEGER DEFAULT 30
) RETURNS TABLE (
    total_views BIGINT,
    total_posts INTEGER,
    published_posts INTEGER,
    top_post_id UUID,
    top_post_title TEXT,
    top_post_views BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(SUM(bda.views), 0)::BIGINT as total_views,
        COUNT(DISTINCT bp.id)::INTEGER as total_posts,
        COUNT(DISTINCT CASE WHEN bp.is_published THEN bp.id END)::INTEGER as published_posts,
        (
            SELECT bp2.id FROM blog_posts bp2
            LEFT JOIN blog_daily_analytics bda2 ON bp2.id = bda2.post_id
            WHERE bp2.is_published = true
            GROUP BY bp2.id
            ORDER BY COALESCE(SUM(bda2.views), 0) DESC
            LIMIT 1
        ) as top_post_id,
        (
            SELECT bp2.title FROM blog_posts bp2
            LEFT JOIN blog_daily_analytics bda2 ON bp2.id = bda2.post_id
            WHERE bp2.is_published = true
            GROUP BY bp2.id, bp2.title
            ORDER BY COALESCE(SUM(bda2.views), 0) DESC
            LIMIT 1
        ) as top_post_title,
        (
            SELECT COALESCE(SUM(bda3.views), 0) FROM blog_daily_analytics bda3
            JOIN blog_posts bp3 ON bda3.post_id = bp3.id
            WHERE bp3.is_published = true
            GROUP BY bp3.id
            ORDER BY COALESCE(SUM(bda3.views), 0) DESC
            LIMIT 1
        )::BIGINT as top_post_views;
END;
$$ LANGUAGE plpgsql;

-- Time series for views
CREATE OR REPLACE FUNCTION get_blog_views_timeseries(
    p_days INTEGER DEFAULT 30
) RETURNS TABLE (
    date DATE,
    views INTEGER,
    unique_visitors INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        d::DATE as date,
        COALESCE(SUM(bda.views), 0)::INTEGER as views,
        COALESCE(SUM(bda.unique_visitors), 0)::INTEGER as unique_visitors
    FROM generate_series(
        CURRENT_DATE - (p_days - 1)::INTEGER,
        CURRENT_DATE,
        '1 day'::INTERVAL
    ) d
    LEFT JOIN blog_daily_analytics bda ON d::DATE = bda.date
    GROUP BY d::DATE
    ORDER BY d::DATE;
END;
$$ LANGUAGE plpgsql;