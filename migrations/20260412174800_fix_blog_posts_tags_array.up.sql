-- Fix malformed tags array values in blog_posts
UPDATE blog_posts 
SET tags = '{}'::TEXT[] 
WHERE tags IS NULL 
   OR tags::text = '' 
   OR tags::text = '[]'
   OR tags::text NOT LIKE '{%';
