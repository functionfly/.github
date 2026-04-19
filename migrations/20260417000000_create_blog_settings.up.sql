-- Blog settings: platform-wide settings for blog display and behavior
CREATE TABLE IF NOT EXISTS blog_settings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_title  VARCHAR(255) NOT NULL DEFAULT 'FunctionFly Blog',
    posts_per_page INTEGER NOT NULL DEFAULT 10,
    meta_description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings row (single-row table)
INSERT INTO blog_settings (id, blog_title, posts_per_page, meta_description)
VALUES (gen_random_uuid(), 'FunctionFly Blog', 10, '')
ON CONFLICT DO NOTHING;

-- Allow authenticated admin users to read and write blog settings
-- (handled by auth middleware on the API routes, not RLS)
