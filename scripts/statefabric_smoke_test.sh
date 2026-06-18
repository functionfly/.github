#!/usr/bin/env bash
# State Fabric Smoke Test
# Tests all CRUD + sub-resource endpoints against a running orchestrator.
# Exits non-zero on any failure.

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="${EMAIL:-admin@functionfly.local}"
PASSWORD="${PASSWORD:-admin123}"
PG_CONN="${PG_CONN:-postgresql://postgres:postgres@localhost:5432/functionfly?sslmode=require}"

PASS=0
FAIL=0
declare -a FAILURES

color() { printf "\033[%sm%s\033[0m" "$1" "$2"; }
ok()   { PASS=$((PASS+1)); printf "  %s %s\n" "$(color '32' '[PASS]')" "$1"; }
bad()  { FAIL=$((FAIL+1)); FAILURES+=("$1: $2"); printf "  %s %s -- %s\n" "$(color '31' '[FAIL]')" "$1" "$2"; }
hdr()  { printf "\n%s %s\n" "$(color '36' '===')" "$1"; }

# Helper: run a request, capture status + body
req() {
  local method=$1 path=$2
  shift 2
  curl -s -o /tmp/sf_body -w "%{http_code}" -X "$method" "${BASE_URL}${path}" "$@"
}

assert_status() {
  local name=$1 expected=$2 actual=$3
  if [[ "$actual" == "$expected" ]]; then
    ok "$name (HTTP $actual)"
  else
    bad "$name" "expected HTTP $expected, got $actual -- body: $(head -c 200 /tmp/sf_body)"
  fi
}

hdr "1. Health & readiness (unauthenticated)"
code=$(req GET /health)
assert_status "GET /health" 200 "$code"
code=$(req GET /healthz)
assert_status "GET /healthz" 200 "$code"
code=$(req GET /v1/state-fabrics/health)
assert_status "GET /v1/state-fabrics/health" 200 "$code"
code=$(req GET /v1/state-fabrics/ready)
# /ready may return 503 if R2 storage is not configured in dev environments
if [[ "$code" == "200" || "$code" == "503" ]]; then
  ok "GET /v1/state-fabrics/ready (HTTP $code, $([ "$code" = "200" ] && echo "R2 available" || echo "R2 unavailable (dev)"))"
else
  bad "GET /v1/state-fabrics/ready" "expected HTTP 200 or 503, got $code"
fi
code=$(req GET /v1/state-fabrics/feature-flags)
assert_status "GET /v1/state-fabrics/feature-flags" 200 "$code"

hdr "2. Authentication"
code=$(req POST /v1/auth/login -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
assert_status "POST /v1/auth/login" 200 "$code"
TOKEN=$(jq -r '.token' /tmp/sf_body)
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  bad "auth token extraction" "token is empty -- body: $(head -c 200 /tmp/sf_body)"
  echo "FATAL: cannot continue without auth token" >&2
  exit 1
fi
ok "auth token obtained (${#TOKEN} chars)"

AUTH=(-H "Authorization: Bearer $TOKEN")

hdr "3. Create fabric (CRUD)"
UNIQ=$(date +%s%N)
FAB_NAME="sf-smoke-$UNIQ"
code=$(req POST /v1/state-fabrics -H "Content-Type: application/json" "${AUTH[@]}" \
  -d "{\"name\":\"$FAB_NAME\",\"description\":\"smoke test\",\"fabric_type\":\"custom\"}")
assert_status "POST /v1/state-fabrics (create)" 201 "$code"
# Extract fabric ID from DB since response body is empty (known bug)
FABRIC_ID=$(PGPASSWORD=postgres psql -h localhost -U postgres -d functionfly -tA -c \
  "SELECT id::text FROM states WHERE name='$FAB_NAME' AND tenant_id='ad23adc3-2f26-4708-9641-5157a1423174' LIMIT 1" 2>/dev/null)
if [[ -n "$FABRIC_ID" ]]; then
  ok "fabric created in DB (id=$FABRIC_ID)"
else
  bad "fabric DB lookup" "no fabric found in DB with name=$FAB_NAME"
  echo "FATAL: cannot continue without fabric ID" >&2
  exit 1
fi

hdr "4. Get / list fabrics"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}" 200 "$code"
code=$(req GET "/v1/state-fabrics" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics (list)" 200 "$code"

hdr "5. Update fabric"
code=$(req PATCH "/v1/state-fabrics/$FABRIC_ID" -H "Content-Type: application/json" "${AUTH[@]}" \
  -d '{"description":"updated by smoke test"}')
assert_status "PATCH /v1/state-fabrics/{id}" 200 "$code"

hdr "6. Metrics"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/metrics" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/metrics" 200 "$code"

hdr "7. Stores sub-resource"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/stores" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/stores (list)" 200 "$code"
code=$(req POST "/v1/state-fabrics/$FABRIC_ID/stores" -H "Content-Type: application/json" "${AUTH[@]}" \
  -d '{"name":"smoke-store","store_type":"persistent","region":"us-east-1"}')
assert_status "POST /v1/state-fabrics/{id}/stores (create)" 201 "$code"
STORE_ID=$(PGPASSWORD=postgres psql -h localhost -U postgres -d functionfly -tA -c \
  "SELECT id::text FROM states WHERE id='$FABRIC_ID' AND tags->>'region'='us-east-1' LIMIT 1" 2>/dev/null)
if [[ -n "$STORE_ID" ]]; then
  ok "store created in DB (id=$STORE_ID)"
  code=$(req DELETE "/v1/state-fabrics/$FABRIC_ID/stores/$STORE_ID" "${AUTH[@]}")
  assert_status "DELETE /v1/state-fabrics/{id}/stores/{storeId}" 204 "$code"
else
  bad "store DB lookup" "no store found in DB"
fi

hdr "8. Pipelines sub-resource"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/pipelines" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/pipelines (list)" 200 "$code"
code=$(req POST "/v1/state-fabrics/$FABRIC_ID/pipelines" -H "Content-Type: application/json" "${AUTH[@]}" \
  -d '{"name":"smoke-pipeline","description":"smoke test pipeline","condition":{}}')
assert_status "POST /v1/state-fabrics/{id}/pipelines (create)" 201 "$code"
PIPELINE_ID=$(PGPASSWORD=postgres psql -h localhost -U postgres -d functionfly -tA -c \
  "SELECT id::text FROM state_triggers WHERE target_function='smoke-pipeline' AND source_state_id='$FABRIC_ID' LIMIT 1" 2>/dev/null)
if [[ -n "$PIPELINE_ID" ]]; then
  ok "pipeline created in DB (id=$PIPELINE_ID)"
  code=$(req PATCH "/v1/state-fabrics/$FABRIC_ID/pipelines/$PIPELINE_ID" -H "Content-Type: application/json" "${AUTH[@]}" \
    -d '{"description":"updated pipeline"}')
  assert_status "PATCH /v1/state-fabrics/{id}/pipelines/{pipelineId}" 200 "$code"
  code=$(req DELETE "/v1/state-fabrics/$FABRIC_ID/pipelines/$PIPELINE_ID" "${AUTH[@]}")
  assert_status "DELETE /v1/state-fabrics/{id}/pipelines/{pipelineId}" 204 "$code"
else
  bad "pipeline DB lookup" "no pipeline found in DB"
fi

hdr "9. Snapshots sub-resource"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/snapshots" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/snapshots (list)" 200 "$code"
code=$(req POST "/v1/state-fabrics/$FABRIC_ID/snapshots" -H "Content-Type: application/json" "${AUTH[@]}" \
  -d '{"name":"smoke-snapshot"}')
assert_status "POST /v1/state-fabrics/{id}/snapshots (create)" 201 "$code"
SNAPSHOT_ID=$(PGPASSWORD=postgres psql -h localhost -U postgres -d functionfly -tA -c \
  "SELECT id::text FROM state_snapshots WHERE label='smoke-snapshot' AND state_id='$FABRIC_ID' LIMIT 1" 2>/dev/null)
if [[ -n "$SNAPSHOT_ID" ]]; then
  ok "snapshot created in DB (id=$SNAPSHOT_ID)"
  code=$(req DELETE "/v1/state-fabrics/$FABRIC_ID/snapshots/$SNAPSHOT_ID" "${AUTH[@]}")
  assert_status "DELETE /v1/state-fabrics/{id}/snapshots/{snapshotId}" 204 "$code"
else
  bad "snapshot DB lookup" "no snapshot found in DB"
fi

hdr "10. Events sub-resource"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/events" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/events" 200 "$code"

hdr "11. Replays sub-resource"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/replays" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/replays (list)" 200 "$code"

hdr "12. Triggers sub-resource"
code=$(req GET "/v1/state-fabrics/$FABRIC_ID/triggers" "${AUTH[@]}")
assert_status "GET /v1/state-fabrics/{id}/triggers (list)" 200 "$code"

hdr "13. Delete fabric (cleanup)"
code=$(req DELETE "/v1/state-fabrics/$FABRIC_ID" "${AUTH[@]}")
assert_status "DELETE /v1/state-fabrics/{id}" 204 "$code"

# Verify deletion
EXISTS=$(PGPASSWORD=postgres psql -h localhost -U postgres -d functionfly -tA -c \
  "SELECT count(*) FROM states WHERE id='$FABRIC_ID'" 2>/dev/null | tr -d ' ')
if [[ "$EXISTS" == "0" ]]; then
  ok "fabric deleted from DB"
else
  bad "fabric deletion verification" "fabric still exists in DB (count=$EXISTS)"
fi

hdr "14. Admin endpoints"
code=$(req GET "/v1/admin/state-fabrics" "${AUTH[@]}")
assert_status "GET /v1/admin/state-fabrics (list all)" 200 "$code"
code=$(req GET "/v1/admin/state-fabrics/stats" "${AUTH[@]}")
assert_status "GET /v1/admin/state-fabrics/stats" 200 "$code"
code=$(req GET "/v1/admin/state-fabrics/settings" "${AUTH[@]}")
assert_status "GET /v1/admin/state-fabrics/settings" 200 "$code"
code=$(req GET "/v1/admin/state-fabrics/cleanup/stats" "${AUTH[@]}")
assert_status "GET /v1/admin/state-fabrics/cleanup/stats" 200 "$code"

hdr "Summary"
TOTAL=$((PASS + FAIL))
echo "  Total: $TOTAL  Passed: $PASS  Failed: $FAIL"
if [[ $FAIL -gt 0 ]]; then
  echo ""
  echo "Failures:"
  for f in "${FAILURES[@]}"; do
    echo "  - $f"
  done
  exit 1
fi
echo "  All smoke tests passed."
exit 0
