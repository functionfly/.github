#!/usr/bin/env bash
# Logical backup of a Neon database using the Neon CLI for the connection string only.
# Uses a direct (non-pooled) endpoint — recommended for pg_dump (see docs/NEON.md).
#
# Prerequisites: neon CLI (make neon-install), PostgreSQL client (pg_dump, gunzip).
# Auth: neon auth  OR  NEON_API_KEY
#
# Env:
#   NEON_BRANCH         Branch name (default: production)
#   NEON_DATABASE_NAME  Database name (default: functionfly)
#   NEON_PROJECT_ID     Optional; passed to neon if set
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BRANCH="${NEON_BRANCH:-production}"
DBNAME="${NEON_DATABASE_NAME:-functionfly}"
PROJ=()
if [[ -n "${NEON_PROJECT_ID:-}" ]]; then
  PROJ=(--project-id "$NEON_PROJECT_ID")
fi

mkdir -p "$ROOT/backups"
STAMP=$(date +%Y%m%d_%H%M%S)
OUT="$ROOT/backups/functionfly_${STAMP}.sql.gz"

if ! command -v neon >/dev/null 2>&1; then
  echo "Neon CLI not found. Install: make neon-install  (or: npm i -g neonctl)" >&2
  exit 1
fi
if ! command -v pg_dump >/dev/null 2>&1; then
  echo "pg_dump not found. Install postgresql-client (e.g. apt install postgresql-client)." >&2
  exit 1
fi

echo "→ Resolving connection string: branch=$BRANCH database=$DBNAME"
CS="$(neon "${PROJ[@]}" connection-string "$BRANCH" --database-name "$DBNAME")"

echo "→ pg_dump | gzip → $OUT"
pg_dump "$CS" | gzip >"$OUT"

gunzip -t "$OUT"
echo "OK: gzip integrity verified"
ls -la "$OUT"
