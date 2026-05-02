-- Migration: Revert tags column type change from TEXT[] back to JSONB
-- Reverses: 20260412175200_change_tags_to_text_array
-- WARNING: This is destructive - converting back will lose TEXT[] specific data

DO $$
DECLARE
    rec RECORD;
    json_tags JSONB;
BEGIN
    -- Add temporary JSONB column
    ALTER TABLE blog_posts ADD COLUMN tags_jsonb JSONB DEFAULT '[]'::jsonb;

    -- Convert TEXT[] to JSONB
    FOR rec IN SELECT id, tags FROM blog_posts WHERE tags IS NOT NULL LOOP
        BEGIN
            SELECT to_jsonb(rec.tags::text[]) INTO json_tags;
            UPDATE blog_posts SET tags_jsonb = json_tags WHERE id = rec.id;
        EXCEPTION WHEN OTHERS THEN
            UPDATE blog_posts SET tags_jsonb = '[]'::jsonb WHERE id = rec.id;
        END;
    END LOOP;

    -- Drop TEXT[] column
    ALTER TABLE blog_posts DROP COLUMN tags;

    -- Rename JSONB column back to tags
    ALTER TABLE blog_posts RENAME COLUMN tags_jsonb TO tags;

    -- Set default
    ALTER TABLE blog_posts ALTER COLUMN tags SET DEFAULT '[]'::jsonb;

    RAISE NOTICE 'Reverted tags column from TEXT[] back to JSONB';
END $$;