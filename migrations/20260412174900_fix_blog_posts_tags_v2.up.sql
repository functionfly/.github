-- Fix malformed tags column in blog_posts (handle both JSONB and TEXT[] types)
DO $$
DECLARE
    col_type TEXT;
BEGIN
    -- Get the actual column type
    SELECT data_type INTO col_type
    FROM information_schema.columns
    WHERE table_name = 'blog_posts' AND column_name = 'tags';

    -- If it's JSONB, fix JSONB format
    IF col_type = 'ARRAY' THEN
        UPDATE blog_posts 
        SET tags = '{}'::TEXT[] 
        WHERE tags IS NULL 
           OR tags::text = '' 
           OR tags::text = '[]'
           OR tags::text NOT LIKE '{%';
        RAISE NOTICE 'Fixed TEXT[] tags column';
    ELSE
        -- For JSONB or other types, set to empty JSON array
        UPDATE blog_posts 
        SET tags = '[]'::jsonb 
        WHERE tags IS NULL 
           OR tags::text = ''
           OR tags = 'null'::jsonb;
        RAISE NOTICE 'Fixed JSONB tags column';
    END IF;
END $$;
