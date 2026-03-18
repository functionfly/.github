#!/usr/bin/env bash
# Create FunctionFly company and initial org chart agents in Paperclip.
# Prereqs: Paperclip running, board user logged in. Set PAPERCLIP_BASE_URL (default http://localhost:3100)
# and either PAPERCLIP_COOKIE (session cookie) or PAPERCLIP_API_KEY.
# Usage: ./deploy/paperclip/scripts/setup-functionfly-company.sh

set -e
BASE_URL="${PAPERCLIP_BASE_URL:-http://localhost:3100}"
AUTH_HEADER=""
if [[ -n "$PAPERCLIP_API_KEY" ]]; then
  AUTH_HEADER="Authorization: Bearer $PAPERCLIP_API_KEY"
elif [[ -n "$PAPERCLIP_COOKIE" ]]; then
  AUTH_HEADER="Cookie: $PAPERCLIP_COOKIE"
else
  echo "Set PAPERCLIP_API_KEY or PAPERCLIP_COOKIE"
  exit 1
fi

# Create company
echo "Creating company FunctionFly..."
RESP=$(curl -s -X POST "$BASE_URL/api/companies" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"name":"FunctionFly","description":"Internal agent control plane for FunctionFly platform"}')
COMPANY_ID=$(echo "$RESP" | jq -r '.id // .companyId // empty')
if [[ -z "$COMPANY_ID" ]]; then
  echo "Failed to create company or get ID. Response: $RESP"
  exit 1
fi
echo "Company ID: $COMPANY_ID"

# Create CTO (no reports_to)
echo "Creating agent CTO..."
CTO_RESP=$(curl -s -X POST "$BASE_URL/api/companies/$COMPANY_ID/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"name":"CTO","role":"cto","title":"CTO","adapterType":"http","adapterConfig":{}}')
CTO_ID=$(echo "$CTO_RESP" | jq -r '.id // .agentId // empty')
if [[ -z "$CTO_ID" ]]; then
  echo "CTO response: $CTO_RESP"
fi

# Create ICs (reports_to CTO if we have CTO_ID)
for name in PlatformEngineer DevOps SupportTriage Security; do
  role=$(echo "$name" | tr '[:upper:]' '[:lower:]')
  title=$(echo "$name" | sed 's/\([A-Z]\)/ \1/g' | xargs)
  payload="{\"name\":\"$name\",\"role\":\"$role\",\"title\":\"$title\",\"adapterType\":\"http\",\"adapterConfig\":{}}"
  if [[ -n "$CTO_ID" ]]; then
    payload=$(echo "$payload" | jq --arg cto "$CTO_ID" '. + {reportsTo: $cto}')
  fi
  echo "Creating agent $name..."
  curl -s -X POST "$BASE_URL/api/companies/$COMPANY_ID/agents" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$payload" > /dev/null || true
done

echo "Done. Create API keys for agents in the Paperclip UI (Agents → select agent → Create API key)."
