ALTER TABLE community_posts ADD COLUMN IF NOT EXISTS slug VARCHAR(320);

UPDATE community_posts SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(title, '[^a-zA-Z0-9 -]', '', 'g'), '\s+', '-', 'g'))
  WHERE slug IS NULL;

UPDATE community_posts SET slug = slug || '-' || SUBSTRING(id::text, 1, 8)
  WHERE slug IN (SELECT slug FROM community_posts GROUP BY slug HAVING COUNT(*) > 1);

ALTER TABLE community_posts ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_community_posts_slug ON community_posts (slug);
