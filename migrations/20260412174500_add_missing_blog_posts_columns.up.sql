-- Add missing columns to blog_posts table
DO $$
BEGIN
    -- Add excerpt column if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'excerpt'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN excerpt TEXT;
        RAISE NOTICE 'Added excerpt column to blog_posts';
    END IF;

    -- Add featured_image column if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'featured_image'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN featured_image VARCHAR(500);
        RAISE NOTICE 'Added featured_image column to blog_posts';
    END IF;

    -- Add sanity_id column if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'sanity_id'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN sanity_id VARCHAR(255);
        RAISE NOTICE 'Added sanity_id column to blog_posts';
    END IF;

    -- Add is_published column if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'is_published'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN is_published BOOLEAN DEFAULT false;
        RAISE NOTICE 'Added is_published column to blog_posts';
    END IF;

    -- Add published_at column if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'published_at'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN published_at TIMESTAMP WITH TIME ZONE;
        RAISE NOTICE 'Added published_at column to blog_posts';
    END IF;

    -- Add tags column if missing (TEXT[])
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'blog_posts' AND column_name = 'tags'
    ) THEN
        ALTER TABLE blog_posts ADD COLUMN tags TEXT[] DEFAULT '{}';
        RAISE NOTICE 'Added tags column to blog_posts';
    END IF;
END $$;
