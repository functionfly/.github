-- Change tags column from jsonb to TEXT[] to match Go code expectations
-- First, extract values if jsonb array, then drop and recreate column

-- Step 1: Create temporary column to store converted values
ALTER TABLE blog_posts ADD COLUMN tags_new TEXT[] DEFAULT '{}';

-- Step 2: Convert existing jsonb data to TEXT[]
UPDATE blog_posts 
SET tags_new = COALESCE(
    (SELECT array_agg(x::text) 
     FROM jsonb_array_elements_text(tags::jsonb) AS x),
    '{}'::TEXT[]
)
WHERE tags IS NOT NULL 
  AND tags::text != ''
  AND jsonb_typeof(tags::jsonb) = 'array';

-- Step 3: Drop old jsonb column
ALTER TABLE blog_posts DROP COLUMN tags;

-- Step 4: Rename new column to tags
ALTER TABLE blog_posts RENAME COLUMN tags_new TO tags;

-- Step 5: Ensure default value
ALTER TABLE blog_posts ALTER COLUMN tags SET DEFAULT '{}';
