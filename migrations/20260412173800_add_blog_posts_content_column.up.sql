-- Add content column to blog_posts if missing
-- Fixes ERROR: column "content" does not exist (SQLSTATE 42703)

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'blog_posts'
        AND column_name = 'content'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN content TEXT;
    END IF;
END $$;
