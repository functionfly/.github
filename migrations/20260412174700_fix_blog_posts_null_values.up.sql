-- Fix NULL values in blog_posts that cause scan errors
UPDATE blog_posts SET content = '' WHERE content IS NULL;
UPDATE blog_posts SET excerpt = '' WHERE excerpt IS NULL;
UPDATE blog_posts SET featured_image = '' WHERE featured_image IS NULL;
UPDATE blog_posts SET sanity_id = '' WHERE sanity_id IS NULL;
UPDATE blog_posts SET author = 'System' WHERE author IS NULL OR author = '';
