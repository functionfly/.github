-- Fix tags column - ensure all values are valid jsonb arrays
UPDATE blog_posts 
SET tags = '[]'::jsonb 
WHERE tags IS NULL 
   OR tags::text = '' 
   OR tags = 'null'::jsonb
   OR jsonb_typeof(tags) != 'array';

-- Also handle any non-array jsonb values by converting them to single-element arrays
UPDATE blog_posts 
SET tags = jsonb_build_array(tags) 
WHERE jsonb_typeof(tags) != 'array' AND tags IS NOT NULL;
