#!/usr/bin/env bash
# Upload CDN assets (SDK, docs, static) to Cloudflare R2 for cdn.functionfly.com.
#
# The API generates these on the fly. This script fetches each asset from the
# origin API and uploads it to R2 so the CDN can serve them without hitting the API.
#
# Prerequisites:
#   - wrangler CLI: npm i -g wrangler && wrangler login
#   - R2 bucket created and custom domain cdn.functionfly.com attached (see docs/CDN_SETUP.md)
#
# Usage:
#   ORIGIN=https://api.functionfly.com R2_BUCKET=functionfly-cdn ./scripts/upload-cdn-to-r2.sh
#
# Important: Use an origin where CACHE_CDN_ENABLED is false (or unset), so the API
# returns 200 with the generated content instead of redirecting to the CDN.
# For first-time upload you can run the API locally with CDN disabled:
#   ORIGIN=http://localhost:8080 R2_BUCKET=functionfly-cdn ./scripts/upload-cdn-to-r2.sh
#
set -euo pipefail

ORIGIN="${ORIGIN:-}"
R2_BUCKET="${R2_BUCKET:-functionfly-cdn}"

if [[ -z "$ORIGIN" ]]; then
  echo "Usage: ORIGIN=<api-base-url> [R2_BUCKET=<bucket>] $0" >&2
  echo "Example: ORIGIN=https://api.functionfly.com R2_BUCKET=functionfly-cdn $0" >&2
  echo "" >&2
  echo "ORIGIN must be the API base URL (no trailing slash), e.g. https://api.functionfly.com or http://localhost:8080" >&2
  echo "Use an origin with CACHE_CDN_ENABLED=false so the API serves content instead of redirecting." >&2
  exit 1
fi

ORIGIN="${ORIGIN%/}"
WORKDIR="${WORKDIR:-$(mktemp -d)}"
cleanup() { rm -rf "$WORKDIR"; }
trap cleanup EXIT

# R2 object key -> API path under /v1/
# Key is the path as served by the CDN (no leading slash). API path is /v1/<key>.
# js/latest is a duplicate of javascript/latest so /sdk/js/latest/functionfly.js works in the UI.
ASSETS=(
  "sdk/javascript/latest/functionfly.js"
  "sdk/js/latest/functionfly.js"
  "sdk/python/latest/functionfly.py"
  "sdk/go/latest/functionfly.go"
  "docs/api/latest/index.html"
  "docs/sdk/latest/javascript.html"
  "docs/guides/latest/getting-started.md"
  "static/images/logo.svg"
  "static/css/styles.css"
)

echo "Origin: $ORIGIN"
echo "Bucket: $R2_BUCKET"
echo "Uploading ${#ASSETS[@]} objects..."
echo ""

for key in "${ASSETS[@]}"; do
  url="${ORIGIN}/v1/${key}"
  tmp="$WORKDIR/$(echo "$key" | tr '/' '_')"
  if curl -sf -o "$tmp" "$url"; then
    if [[ ! -s "$tmp" ]]; then
      echo "  skip $key (empty response; is the API redirecting to CDN?)"
      continue
    fi
    if npx wrangler r2 object put "$R2_BUCKET/$key" --file="$tmp" 2>/dev/null; then
      echo "  ok   $key"
    else
      echo "  fail $key (wrangler error)"
      exit 1
    fi
  else
    echo "  fail $key (curl error)"
    exit 1
  fi
done

echo ""
echo "Done. Verify: curl -sI https://cdn.functionfly.com/sdk/javascript/latest/functionfly.js"
