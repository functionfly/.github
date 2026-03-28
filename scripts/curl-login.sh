#!/usr/bin/env bash
# Quick login test. Use the same email/password you used with create-admin.
# Example: ./scripts/curl-login.sh
# Or:     ./scripts/curl-login.sh admin@functionfly.local mypassword

set -e

# SECURITY: In production (DEVELOPMENT != true), require credentials as arguments
if [ "$DEVELOPMENT" != "true" ]; then
    if [ -z "$1" ] || [ -z "$2" ]; then
        echo "ERROR: In production, you must provide email and password as arguments."
        echo "       Usage: $0 <email> <password>"
        exit 1
    fi
fi

EMAIL="${1:-admin@functionfly.local}"
PASSWORD="${2:-admin123}"
# Direct API (use 8080 when orchestrator-api is running)
BASE="${API_URL:-http://localhost:8080}"
echo "POST $BASE/v1/auth/login (email=$EMAIL)"
echo ""
RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
BODY=$(echo "$RESP" | head -n -1)
CODE=$(echo "$RESP" | tail -n 1)
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
echo ""
echo "HTTP $CODE"
if [ "$CODE" = "200" ]; then
  echo "$BODY" | grep -q token && echo "Login OK: token present" || exit 1
else
  exit 1
fi
