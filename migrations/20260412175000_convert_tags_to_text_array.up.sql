-- Convert tags from jsonb to TEXT[] format that pq driver expects
-- The Go code uses pq.StringArray which expects PostgreSQL array format like {'tag1','tag2'}

DO $$
DECLARE
    rec RECORD;
    tag_array TEXT[];
    json_tags JSONB;
BEGIN
    -- Process each row with jsonb tags
    FOR rec IN SELECT id, tags FROM blog_posts WHERE tags IS NOT NULL LOOP
        BEGIN
            -- Check if it's a jsonb array
            IF jsonb_typeof(rec.tags::jsonb) = 'array' THEN
                -- Convert jsonb array to TEXT[]
                SELECT array_agg(x::text)
                INTO tag_array
                FROM jsonb_array_elements_text(rec.tags::jsonb) AS x;
                
                UPDATE blog_posts SET tags = tag_array WHERE id = rec.id;
            ELSE
                -- Set empty array for non-array values
                UPDATE blog_posts SET tags = '{}'::TEXT[] WHERE id = rec.id;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            -- On any error, set empty array
            UPDATE blog_posts SET tags = '{}'::TEXT[] WHERE id = rec.id;
        END;
    END LOOP;
    
    RAISE NOTICE 'Converted tags to TEXT[] format';
END $$;
