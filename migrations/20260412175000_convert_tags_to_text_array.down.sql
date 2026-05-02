-- Migration: Revert tags conversion from jsonb to TEXT[]
-- Reverses: 20260412175000_convert_tags_to_text_array
-- WARNING: This is destructive - converting back to jsonb will lose TEXT[] specific data

DO $$
DECLARE
    rec RECORD;
    json_tags JSONB;
BEGIN
    -- Convert TEXT[] back to JSONB for each row
    FOR rec IN SELECT id, tags FROM blog_posts WHERE tags IS NOT NULL LOOP
        BEGIN
            -- Convert TEXT[] to JSONB
            SELECT to_jsonb(rec.tags::text[]) INTO json_tags;
            UPDATE blog_posts SET tags = json_tags WHERE id = rec.id;
        EXCEPTION WHEN OTHERS THEN
            -- On any error, set to null jsonb
            UPDATE blog_posts SET tags = 'null'::jsonb WHERE id = rec.id;
        END;
    END LOOP;
    
    RAISE NOTICE 'Reverted tags to JSONB format';
END $$;