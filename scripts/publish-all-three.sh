#!/usr/bin/env bash
# Publish base64-decode, json-prettify, uuid-generate. Start API first (e.g. make dev).
set -e
cd "$(dirname "$0")/.."
for f in examples/stdlib-publish/publish_base64_decode.json examples/stdlib-publish/publish_json_prettify.json examples/stdlib-publish/publish_uuid_generate.json; do
  echo "--- $f ---"
  ./scripts/publish-from-json.sh "$f" "${1:-http://localhost:8080}" || exit 1
done
echo "Done. All three published."
