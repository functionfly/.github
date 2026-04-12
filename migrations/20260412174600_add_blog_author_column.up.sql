-- Add missing columns to blog_posts table
DO $$
BEGIN
    -- Add author column if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'author'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN author VARCHAR(255) NOT NULL DEFAULT 'System';
        RAISE NOTICE 'Added author column to blog_posts';
    END IF;
END $$;
