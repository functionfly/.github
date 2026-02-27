#!/usr/bin/env bash
# Publish a function from a JSON file (e.g. publish_slugify.json).
# Usage: ./scripts/publish-from-json.sh <path-to-publish.json> [SERVER_URL]
# Example: ./scripts/publish-from-json.sh publish_slugify.json
#
# Token: uses JWT from generate_token.go (loads .env for JWT_SECRET). If you get 401,
# use login: PUBLISH_USE_LOGIN=1 ./scripts/publish-from-json.sh publish_slugify.json
# (set PUBLISH_EMAIL and PUBLISH_PASSWORD, or defaults admin@functionfly.local / admin123)
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
JSON_FILE="${1:?Usage: $0 <publish.json> [SERVER_URL]}"
SERVER_URL="${2:-http://localhost:8080}"

cd "$ROOT_DIR"

if [[ ! -f "$JSON_FILE" ]]; then
  echo "Error: File not found: $JSON_FILE"
  exit 1
fi

# Load .env so JWT_SECRET matches server (if server was started with same .env)
if [[ -f "$ROOT_DIR/.env" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "$ROOT_DIR/.env" 2>/dev/null || true
  set +a
fi

echo "Getting token..."
TOKEN=""
if [[ -n "${PUBLISH_USE_LOGIN:-}" ]]; then
  EMAIL="${PUBLISH_EMAIL:-admin@functionfly.local}"
  PASS="${PUBLISH_PASSWORD:-admin123}"
  RESP=$(curl -s -X POST "$SERVER_URL/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" 2>/dev/null || true)
  TOKEN=$(echo "$RESP" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
  if [[ -z "$TOKEN" ]]; then
    echo "Error: Login failed. Check PUBLISH_EMAIL / PUBLISH_PASSWORD or create admin (create-admin)."
    echo "Response: $RESP"
    exit 1
  fi
else
  TOKEN=$(go run generate_token.go 2>/dev/null | tr -d '\n\r')
fi

if [[ -z "$TOKEN" ]]; then
  echo "Error: Failed to generate token. Try: PUBLISH_USE_LOGIN=1 $0 $*"
  exit 1
fi

echo "Publishing from $JSON_FILE to $SERVER_URL/v1/registry/publish ..."
HTTP_CODE=$(curl -s -w "%{http_code}" -o /tmp/publish_out.json \
  -X POST "$SERVER_URL/v1/registry/publish" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @"$JSON_FILE")

if [[ "$HTTP_CODE" =~ ^2 ]]; then
  echo "OK (HTTP $HTTP_CODE)"
  cat /tmp/publish_out.json
else
  echo "Publish failed (HTTP $HTTP_CODE)"
  cat /tmp/publish_out.json
  exit 1
fi
