#!/usr/bin/env bash
# Update the "Introducing State Fabric" blog post content from the markdown file.
# Uses same DB env as seed-blog (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME).
# Run from repo root: ./scripts/update-blog-post-from-markdown.sh
# With Docker Postgres: DB_PORT=5434 ./scripts/update-blog-post-from-markdown.sh

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MD_FILE="${REPO_ROOT}/content/blog/introducing-state-fabric.md"
SLUG="introducing-state-fabric"

if [[ ! -f "$MD_FILE" ]]; then
  echo "Error: Markdown file not found: $MD_FILE"
  exit 1
fi

# Build SQL with dollar-quoting so content does not need escaping.
# Delimiter chosen to not appear in the markdown body.
TMP_SQL=$(mktemp)
trap "rm -f $TMP_SQL" EXIT

{
  echo "UPDATE blog_posts SET content = \$BLOGMD\$"
  cat "$MD_FILE"
  echo "\$BLOGMD\$ WHERE slug = '$SLUG';"
} > "$TMP_SQL"

PGPASSWORD="${DB_PASSWORD:-postgres}" psql \
  -h "${DB_HOST:-localhost}" \
  -p "${DB_PORT:-5432}" \
  -U "${DB_USER:-postgres}" \
  -d "${DB_NAME:-functionfly}" \
  -f "$TMP_SQL"

echo "Updated blog post content for slug: $SLUG"
