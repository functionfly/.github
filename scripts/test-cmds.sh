#!/usr/bin/env bash
# Quick API test commands. Usage: ./scripts/test-cmds.sh [base_url]
# Default base: http://localhost:8080 (set API_URL or pass as first arg)

BASE="${1:-${API_URL:-http://localhost:8080}}"
V1="$BASE/v1"
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "=== FunctionFly API test commands (BASE=$BASE) ==="
echo ""

# 1. Health
echo "1. Health"
echo "   curl -s $BASE/health"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/health" 2>/dev/null || echo "000")
if [ "$CODE" = "200" ]; then
  echo -e "   ${GREEN}→ $CODE OK${NC}"
else
  echo -e "   ${RED}→ $CODE (API not reachable?)${NC}"
fi
echo ""

# 2. Login (use credentials from create-admin or environment variables)
echo "2. Login"
EMAIL="${TEST_EMAIL:-admin@functionfly.local}"
PASSWORD="${TEST_PASSWORD:-}"
if [ -z "$PASSWORD" ]; then
  echo "   ERROR: Set TEST_PASSWORD environment variable or pass credentials"
  echo "   Usage: TEST_EMAIL=user@example.com TEST_PASSWORD=secret ./scripts/test-cmds.sh"
  exit 1
fi
echo "   curl -s -X POST $V1/auth/login -H 'Content-Type: application/json' -d '{\"email\":\"$EMAIL\",\"password\":\"***\"}'"
LOGIN=$(curl -s -X POST "$V1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
if echo "$LOGIN" | grep -q '"token"'; then
  echo -e "   ${GREEN}→ Login OK, token present${NC}"
  TOKEN=$(echo "$LOGIN" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
  echo "   Token (first 40 chars): ${TOKEN:0:40}..."
else
  echo -e "   ${RED}→ Login failed${NC}"
  echo "$LOGIN" | jq . 2>/dev/null || echo "   $LOGIN"
fi
echo ""

# 3. Session (if we got a token)
if echo "$LOGIN" | grep -q '"token"'; then
  TOKEN=$(echo "$LOGIN" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
  echo "3. Get session (with token)"
  echo "   curl -s $V1/auth/session -H 'Authorization: Bearer <token>'"
  SESS=$(curl -s "$V1/auth/session" -H "Authorization: Bearer $TOKEN")
  if echo "$SESS" | grep -q "session\|user\|data"; then
    echo -e "   ${GREEN}→ Session OK${NC}"
  else
    echo "   Response: $(echo "$SESS" | head -c 200)"
  fi
fi

echo ""
echo "=== Copy-paste one-liners ==="
echo "  Health:  curl -s $BASE/health"
echo "  Login:   TEST_EMAIL=user@example.com TEST_PASSWORD=secret curl -s -X POST $V1/auth/login -H 'Content-Type: application/json' -d '{\"email\":\"\$TEST_EMAIL\",\"password\":\"\$TEST_PASSWORD\"}'"
echo "  (Set TEST_EMAIL and TEST_PASSWORD environment variables)"
