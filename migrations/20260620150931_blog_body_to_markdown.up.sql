-- Migration: Convert blog post body from TipTap JSON to Markdown
--
-- The legacy format stored TipTap editor JSON (with 'children' array and
-- kebab-case types like 'bulleted-list'). We convert it to Markdown for
-- simpler, version-control-friendly storage that doesn't depend on TipTap.
--
-- This SQL migration only changes the column type. Run the Go script
-- scripts/migrate-blog-body.go BEFORE applying this migration to convert
-- the data. The Go script handles the complex TipTap JSON to Markdown
-- conversion that would be unwieldy in pure SQL.
--
-- Usage:
--   1. go run scripts/migrate-blog-body.go
--   2. Apply this migration (changes column type to TEXT)

-- Change the column type from JSONB to TEXT
ALTER TABLE blog_posts
    ALTER COLUMN body TYPE TEXT USING body::text;
