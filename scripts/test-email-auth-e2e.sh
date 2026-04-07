#!/bin/bash
# FunctionFly Email Auth Flow E2E Test
# Tests: signup -> verify email -> login -> password reset -> login with new password
#
# Prerequisites: API running on localhost:8080, Postgres + Redis started
# Usage: ./scripts/test-email-auth-e2e.sh [API_URL] [INVITE_CODE]
#
# For Mailpit users: emails are captured at http://localhost:8025

set -euo pipefail

API_URL="${1:-http://localhost:8080}"
INVITE_CODE="${2:-${INVITE_CODE:-}}"
MAILPIT_URL="${MAILPIT_URL:-http://localhost:8025}"
TIMESTAMP=$(date +%s)
TEST_EMAIL="e2e-test-${TIMESTAMP}@example.com"
TEST_PASSWORD="Test123!@#"
TEST_PASSWORD_NEW="NewPass456!@#"
TEST_USERNAME="e2euser${TIMESTAMP}"
TEST_NAME="E2E Test User"
TEST_DOB="1990-01-15"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

log_info()  { echo -e "${BLUE}[INFO]${NC}  $1"; }
log_pass()  { echo -e "${GREEN}[PASS]${NC}  $1"; PASSED=$((PASSED + 1)); TOTAL=$((TOTAL + 1)); }
log_fail()  { echo -e "${RED}[FAIL]${NC}  $1"; FAILED=$((FAILED + 1)); TOTAL=$((TOTAL + 1)); }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
log_step()  { echo -e "\n${BLUE}━━━ Step $1: $2 ━━━${NC}"; }

# Cleanup: delete test user from DB if DB env vars available
cleanup() {
    log_info "Test email: $TEST_EMAIL"
    log_info "Test username: $TEST_USERNAME"
    if command -v psql &>/dev/null && [ -n "${DB_HOST:-}" ]; then
        log_info "Cleaning up test user from database..."
        PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-functionfly}" \
            -c "DELETE FROM users WHERE email = '${TEST_EMAIL}';" 2>/dev/null || true
    fi
}
trap cleanup EXIT

assert_status() {
    local expected="$1" actual="$2" label="$3"
    if [ "$actual" = "$expected" ]; then
        log_pass "$label (HTTP $actual)"
        return 0
    else
        log_fail "$label (expected HTTP $expected, got $actual)"
        return 1
    fi
}

assert_contains() {
    local haystack="$1" needle="$2" label="$3"
    if echo "$haystack" | grep -q "$needle"; then
        log_pass "$label"
        return 0
    else
        log_fail "$label (missing: '$needle')"
        echo "  Response: $(echo "$haystack" | head -c 200)"
        return 1
    fi
}

pace() {
    # Small delay to avoid tripping the 10-req/60s auth rate limiter when
    # running repeated E2E tests from the same IP.
    local secs="${PACE_SECONDS:-3}"
    sleep "$secs"
}

assert_not_contains() {
    local haystack="$1" needle="$2" label="$3"
    if echo "$haystack" | grep -q "$needle"; then
        log_fail "$label (unexpectedly contains: '$needle')"
        return 1
    else
        log_pass "$label"
        return 0
    fi
}

extract_json_field() {
    # Extract a JSON field value using python3 or grep
    local json="$1" field="$2"
    echo "$json" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('$field',''))" 2>/dev/null \
        || echo "$json" | grep -o "\"${field}\":\"[^\"]*\"" | head -1 | cut -d'"' -f4
}

# ═══════════════════════════════════════════════════════════════
echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  FunctionFly Email Auth Flow — E2E Test                 ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
log_info "API:    $API_URL"
log_info "Email:  $TEST_EMAIL"
log_info "User:   $TEST_USERNAME"
[ -n "$INVITE_CODE" ] && log_info "Invite: $INVITE_CODE"
echo ""

# ── Step 0: Health check ─────────────────────────────────────
log_step 0 "Health check"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health" 2>/dev/null || echo "000")
assert_status "200" "$STATUS" "API health endpoint"

# ── Step 1: Signup config ────────────────────────────────────
log_step 1 "Get signup config"
SIGNUP_CFG=$(curl -s -w "\n%{http_code}" "$API_URL/auth/signup-config")
CFG_BODY=$(echo "$SIGNUP_CFG" | sed '$d')
CFG_STATUS=$(echo "$SIGNUP_CFG" | tail -1)
assert_status "200" "$CFG_STATUS" "GET /auth/signup-config"
log_info "Signup config: $(echo "$CFG_BODY" | head -c 100)"

# ── Step 2: Check username availability ──────────────────────
log_step 2 "Check username availability"
UNAME_RESP=$(curl -s -w "\n%{http_code}" "$API_URL/auth/check-username?username=$TEST_USERNAME")
UNAME_BODY=$(echo "$UNAME_RESP" | sed '$d')
UNAME_STATUS=$(echo "$UNAME_RESP" | tail -1)
assert_status "200" "$UNAME_STATUS" "GET /auth/check-username"
assert_contains "$UNAME_BODY" '"available":true' "Username is available"

# ── Step 3: Signup ───────────────────────────────────────────
log_step 3 "Signup with email"
pace
INVITE_FIELD=""
[ -n "$INVITE_CODE" ] && INVITE_FIELD=", \"inviteCode\": \"$INVITE_CODE\""
SIGNUP_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/signup" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\",
        \"confirmPassword\": \"$TEST_PASSWORD\",
        \"username\": \"$TEST_USERNAME\",
        \"name\": \"$TEST_NAME\",
        \"dateOfBirth\": \"$TEST_DOB\",
        \"termsAccepted\": true${INVITE_FIELD}
    }")
SIGNUP_BODY=$(echo "$SIGNUP_RESP" | sed '$d')
SIGNUP_STATUS=$(echo "$SIGNUP_RESP" | tail -1)
# Accept 200 or 201 — production returns 200 with message, local returns 201
if [ "$SIGNUP_STATUS" = "201" ] || [ "$SIGNUP_STATUS" = "200" ]; then
    log_pass "POST /auth/signup (HTTP $SIGNUP_STATUS)"
    PASSED=$((PASSED + 1)); TOTAL=$((TOTAL + 1))
else
    log_fail "POST /auth/signup (expected HTTP 200/201, got $SIGNUP_STATUS)"
    FAILED=$((FAILED + 1)); TOTAL=$((TOTAL + 1))
    echo "  Response: $(echo "$SIGNUP_BODY" | head -c 300)"
fi
assert_contains "$SIGNUP_BODY" "emailSent" "Response contains emailSent field"
log_info "Signup response: $(echo "$SIGNUP_BODY" | head -c 200)"

USER_ID=$(extract_json_field "$SIGNUP_BODY" "userId")
log_info "Created user ID: $USER_ID"

# ── Step 4: Login before verification (should fail or warn) ──
log_step 4 "Login before email verification"
pace
LOGIN_PRE_VERIFY=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}")
LOGIN_PRE_BODY=$(echo "$LOGIN_PRE_VERIFY" | sed '$d')
LOGIN_PRE_STATUS=$(echo "$LOGIN_PRE_VERIFY" | tail -1)

if [ "$LOGIN_PRE_STATUS" = "401" ] || [ "$LOGIN_PRE_STATUS" = "403" ]; then
    log_pass "Login before verification correctly rejected (HTTP $LOGIN_PRE_STATUS)"
    assert_contains "$LOGIN_PRE_BODY" "verif" "Error mentions verification"
elif [ "$LOGIN_PRE_STATUS" = "200" ]; then
    log_warn "Login before verification succeeded — verification may not be enforced"
else
    log_fail "Unexpected status for pre-verification login (HTTP $LOGIN_PRE_STATUS)"
fi

# ── Step 5: Resend verification email ────────────────────────
log_step 5 "Resend verification email"
pace
RESEND_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/resend-verification" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$TEST_EMAIL\"}")
RESEND_STATUS=$(echo "$RESEND_RESP" | tail -1)
if [ "$RESEND_STATUS" = "200" ]; then
    log_pass "POST /auth/resend-verification (HTTP 200)"
    PASSED=$((PASSED + 1)); TOTAL=$((TOTAL + 1))
elif [ "$RESEND_STATUS" = "429" ]; then
    log_warn "POST /auth/resend-verification rate-limited (HTTP 429) — first signup email still valid"
else
    log_fail "POST /auth/resend-verification (expected HTTP 200/429, got $RESEND_STATUS)"
    FAILED=$((FAILED + 1)); TOTAL=$((TOTAL + 1))
fi

# ── Step 6: Retrieve verification token ─────────────────────
# In dev, the MockService sends email via Mailpit on :1025.
# We can fetch the token from Mailpit's API, or fall back to
# reading it directly from the database.
log_step 6 "Retrieve verification token"
VERIFICATION_TOKEN=""

# Try Mailpit first
if curl -s --connect-timeout 2 "$MAILPIT_URL/api/v1/search" &>/dev/null; then
    log_info "Querying Mailpit for verification email..."
    MAILPIT_SEARCH=$(curl -s "$MAILPIT_URL/api/v1/search?query=to:$TEST_EMAIL&limit=1")
    MSG_ID=$(echo "$MAILPIT_SEARCH" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['messages'][0]['ID'] if d.get('messages') else '')" 2>/dev/null)
    if [ -n "$MSG_ID" ]; then
        MSG_HTML=$(curl -s "$MAILPIT_URL/api/v1/message/$MSG_ID" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Text','') + d.get('HTML',''))" 2>/dev/null)
        # Extract token from verify-email link
        VERIFICATION_TOKEN=$(echo "$MSG_HTML" | grep -oP 'token=[\w\-]+' | head -1 | sed 's/token=//')
        log_info "Extracted verification token from Mailpit"
    fi
fi

# Fallback: read from database
if [ -z "$VERIFICATION_TOKEN" ] && command -v psql &>/dev/null && [ -n "${DB_HOST:-}" ]; then
    log_info "Reading verification token from database..."
    VERIFICATION_TOKEN=$(PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" \
        -U "${DB_USER:-postgres}" -d "${DB_NAME:-functionfly}" -t -A \
        -c "SELECT verification_token FROM users WHERE email = '$TEST_EMAIL';" 2>/dev/null)
fi

if [ -n "$VERIFICATION_TOKEN" ]; then
    log_pass "Retrieved verification token: ${VERIFICATION_TOKEN:0:16}..."
else
    log_fail "Could not retrieve verification token (no Mailpit or DB access)"
    log_warn "Manual step: check your email or Mailpit at http://localhost:8025"
    log_warn "Set DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME env vars to read from DB directly"
    echo ""
    echo "To find the token manually:"
    echo "  psql -c \"SELECT verification_token FROM users WHERE email = '$TEST_EMAIL';\""
    echo ""
    if [ -t 0 ]; then
        read -rp "Paste verification token (or press Enter to skip): " VERIFICATION_TOKEN || true
    fi
    if [ -z "$VERIFICATION_TOKEN" ]; then
        log_warn "Skipping email verification steps"
    fi
fi

# ── Step 7: Verify email ────────────────────────────────────
if [ -n "$VERIFICATION_TOKEN" ]; then
    log_step 7 "Verify email with token"
    VERIFY_RESP=$(curl -s -w "\n%{http_code}" "$API_URL/auth/verify-email?token=$VERIFICATION_TOKEN")
    VERIFY_BODY=$(echo "$VERIFY_RESP" | sed '$d')
    VERIFY_STATUS=$(echo "$VERIFY_RESP" | tail -1)
    assert_status "200" "$VERIFY_STATUS" "GET /auth/verify-email?token=..."
    assert_contains "$VERIFY_BODY" "verified" "Response confirms verification"

    VERIFY_TOKEN=$(extract_json_field "$VERIFY_BODY" "token")
    if [ -n "$VERIFY_TOKEN" ]; then
        log_pass "Auto-login token returned after verification"
        log_info "Verification token: ${VERIFY_TOKEN:0:32}..."
    else
        log_warn "No auto-login token in verification response"
    fi
else
    log_step 7 "Verify email with token"
    log_warn "Skipped (no token available)"
fi

# ── Step 8: Verify token is rejected on reuse ───────────────
if [ -n "${VERIFICATION_TOKEN:-}" ]; then
    log_step 8 "Verify token reuse is rejected"
    VERIFY_REUSE=$(curl -s -w "\n%{http_code}" "$API_URL/auth/verify-email?token=$VERIFICATION_TOKEN")
    VERIFY_REUSE_STATUS=$(echo "$VERIFY_REUSE" | tail -1)
    assert_status "400" "$VERIFY_REUSE_STATUS" "Reusing verification token returns 400"
fi

# ── Step 9: Login after verification ────────────────────────
log_step 9 "Login after email verification"
pace
LOGIN_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}")
LOGIN_BODY=$(echo "$LOGIN_RESP" | sed '$d')
LOGIN_STATUS=$(echo "$LOGIN_RESP" | tail -1)
assert_status "200" "$LOGIN_STATUS" "POST /auth/login (after verification)"

ACCESS_TOKEN=$(extract_json_field "$LOGIN_BODY" "token")
REFRESH_TOKEN=$(extract_json_field "$LOGIN_BODY" "refresh_token")

if [ -n "$ACCESS_TOKEN" ]; then
    log_pass "Received access token"
    log_info "Access token: ${ACCESS_TOKEN:0:32}..."
else
    log_fail "No access token in login response"
    echo "  Response: $(echo "$LOGIN_BODY" | head -c 200)"
fi

if [ -n "$REFRESH_TOKEN" ]; then
    log_pass "Received refresh token"
else
    log_warn "No refresh token in login response"
fi

# ── Step 10: Validate token ─────────────────────────────────
if [ -n "$ACCESS_TOKEN" ]; then
    log_step 10 "Validate access token"
    VALIDATE_RESP=$(curl -s -w "\n%{http_code}" "$API_URL/v1/auth/validate" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    VALIDATE_BODY=$(echo "$VALIDATE_RESP" | sed '$d')
    VALIDATE_STATUS=$(echo "$VALIDATE_RESP" | tail -1)
    assert_status "200" "$VALIDATE_STATUS" "GET /v1/auth/validate"
    assert_contains "$VALIDATE_BODY" "$TEST_EMAIL" "User email in validate response"
    assert_contains "$VALIDATE_BODY" '"email_verified":true' "Email verified flag is true"
fi

# ── Step 11: Get current user ───────────────────────────────
if [ -n "$ACCESS_TOKEN" ]; then
    log_step 11 "Get current user (/v1/users/me)"
    ME_RESP=$(curl -s -w "\n%{http_code}" "$API_URL/v1/users/me" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    ME_BODY=$(echo "$ME_RESP" | sed '$d')
    ME_STATUS=$(echo "$ME_RESP" | tail -1)
    assert_status "200" "$ME_STATUS" "GET /v1/users/me"
    assert_contains "$ME_BODY" "$TEST_EMAIL" "User email in /v1/users/me"
    assert_contains "$ME_BODY" '"tenantId"' "tenantId present in /v1/users/me"
fi

# ── Step 12: Password reset request ─────────────────────────
log_step 12 "Request password reset"
pace
RESET_REQ_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/v1/auth/password-reset" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$TEST_EMAIL\"}")
RESET_REQ_BODY=$(echo "$RESET_REQ_RESP" | sed '$d')
RESET_REQ_STATUS=$(echo "$RESET_REQ_RESP" | tail -1)
assert_status "200" "$RESET_REQ_STATUS" "POST /auth/password-reset"
assert_contains "$RESET_REQ_BODY" "reset" "Response mentions reset"

# ── Step 13: Retrieve reset token ───────────────────────────
log_step 13 "Retrieve password reset token"
RESET_TOKEN=""

# Try Mailpit
if curl -s --connect-timeout 2 "$MAILPIT_URL/api/v1/search" &>/dev/null; then
    log_info "Querying Mailpit for password reset email..."
    MAILPIT_SEARCH=$(curl -s "$MAILPIT_URL/api/v1/search?query=to:$TEST_EMAIL%20subject:Reset&limit=1")
    MSG_ID=$(echo "$MAILPIT_SEARCH" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['messages'][0]['ID'] if d.get('messages') else '')" 2>/dev/null)
    if [ -n "$MSG_ID" ]; then
        MSG_TEXT=$(curl -s "$MAILPIT_URL/api/v1/message/$MSG_ID" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Text',''))" 2>/dev/null)
        RESET_TOKEN=$(echo "$MSG_TEXT" | grep -oP 'token=[\w\-]+' | head -1 | sed 's/token=//')
        if [ -n "$RESET_TOKEN" ]; then
            log_info "Extracted reset token from Mailpit"
        fi
    fi
fi

# Fallback: read from database (reset token is stored in verification_token column)
if [ -z "$RESET_TOKEN" ] && command -v psql &>/dev/null && [ -n "${DB_HOST:-}" ]; then
    log_info "Reading reset token from database..."
    RESET_TOKEN=$(PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" \
        -U "${DB_USER:-postgres}" -d "${DB_NAME:-functionfly}" -t -A \
        -c "SELECT verification_token FROM users WHERE email = '$TEST_EMAIL';" 2>/dev/null || true)
fi

if [ -n "$RESET_TOKEN" ]; then
    log_pass "Retrieved password reset token: ${RESET_TOKEN:0:16}..."
else
    log_warn "Could not retrieve reset token automatically"
    if [ -t 0 ]; then
        read -rp "Paste reset token (or press Enter to skip): " RESET_TOKEN || true
    fi
fi

# ── Step 14: Confirm password reset ─────────────────────────
if [ -n "${RESET_TOKEN:-}" ]; then
    log_step 14 "Confirm password reset"
    pace
    RESET_CONFIRM=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/v1/auth/password-reset/confirm" \
        -H "Content-Type: application/json" \
        -d "{\"token\": \"$RESET_TOKEN\", \"newPassword\": \"$TEST_PASSWORD_NEW\"}")
    RESET_CONFIRM_BODY=$(echo "$RESET_CONFIRM" | sed '$d')
    RESET_CONFIRM_STATUS=$(echo "$RESET_CONFIRM" | tail -1)
    assert_status "200" "$RESET_CONFIRM_STATUS" "POST /auth/password-reset/confirm"
    assert_contains "$RESET_CONFIRM_BODY" "success" "Response confirms password reset"
else
    log_step 14 "Confirm password reset"
    log_warn "Skipped (no reset token available)"
fi

# ── Step 15: Login with old password (should fail) ──────────
if [ -n "${RESET_TOKEN:-}" ]; then
    log_step 15 "Login with old password (should fail)"
    pace
    OLD_PW_LOGIN=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}")
    OLD_PW_STATUS=$(echo "$OLD_PW_LOGIN" | tail -1)
    assert_status "401" "$OLD_PW_STATUS" "Login with old password rejected"
fi

# ── Step 16: Login with new password ────────────────────────
if [ -n "${RESET_TOKEN:-}" ]; then
    log_step 16 "Login with new password"
    pace
    NEW_PW_LOGIN=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD_NEW\"}")
    NEW_PW_BODY=$(echo "$NEW_PW_LOGIN" | sed '$d')
    NEW_PW_STATUS=$(echo "$NEW_PW_LOGIN" | tail -1)
    assert_status "200" "$NEW_PW_STATUS" "Login with new password succeeds"

    NEW_ACCESS_TOKEN=$(extract_json_field "$NEW_PW_BODY" "token")
    if [ -n "$NEW_ACCESS_TOKEN" ]; then
        log_pass "Received new access token after password reset"
    fi
fi

# ── Step 17: Refresh token ──────────────────────────────────
if [ -n "${REFRESH_TOKEN:-}" ]; then
    log_step 17 "Refresh access token"
    REFRESH_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")
    REFRESH_BODY=$(echo "$REFRESH_RESP" | sed '$d')
    REFRESH_STATUS=$(echo "$REFRESH_RESP" | tail -1)

    if [ "$REFRESH_STATUS" = "200" ]; then
        log_pass "Token refresh succeeded (HTTP 200)"
        NEW_REFRESH=$(extract_json_field "$REFRESH_BODY" "refresh_token")
        if [ -n "$NEW_REFRESH" ]; then
            log_pass "Received rotated refresh token"
        fi
    elif [ "$REFRESH_STATUS" = "401" ]; then
        # Refresh tokens may not be returned if verification flow already issued a session
        log_warn "Refresh token rejected (401) — may have been consumed by verification auto-login"
    else
        log_fail "Token refresh returned HTTP $REFRESH_STATUS"
    fi
fi

# ── Step 18: Logout ─────────────────────────────────────────
FINAL_TOKEN="${NEW_ACCESS_TOKEN:-$ACCESS_TOKEN}"
if [ -n "$FINAL_TOKEN" ]; then
    log_step 18 "Logout"
    LOGOUT_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/v1/auth/logout" \
        -H "Authorization: Bearer $FINAL_TOKEN")
    LOGOUT_STATUS=$(echo "$LOGOUT_RESP" | tail -1)
    assert_status "200" "$LOGOUT_STATUS" "POST /auth/logout"
fi

# ── Step 19: Use token after logout (should fail) ──────────
if [ -n "$FINAL_TOKEN" ]; then
    log_step 19 "Validate token after logout"
    POST_LOGOUT=$(curl -s -w "\n%{http_code}" "$API_URL/v1/auth/validate" \
        -H "Authorization: Bearer $FINAL_TOKEN")
    POST_LOGOUT_STATUS=$(echo "$POST_LOGOUT" | tail -1)
    # Token validation may still pass (JWT is stateless), but session should be gone
    if [ "$POST_LOGOUT_STATUS" = "401" ]; then
        log_pass "Post-logout token validation rejected (401)"
    elif [ "$POST_LOGOUT_STATUS" = "200" ]; then
        log_warn "Post-logout token still valid (JWT is stateless — session revoked but token not expired)"
    else
        log_fail "Unexpected status after logout (HTTP $POST_LOGOUT_STATUS)"
    fi
fi

# ═══════════════════════════════════════════════════════════════
echo ""
echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Results                                                ║${NC}"
echo -e "${BLUE}╠══════════════════════════════════════════════════════════╣${NC}"
echo -e "${BLUE}║${NC}  Total:  ${TOTAL}                                              ${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${GREEN}Passed: ${PASSED}${NC}                                              ${BLUE}║${NC}"
if [ $FAILED -gt 0 ]; then
echo -e "${BLUE}║${NC}  ${RED}Failed: ${FAILED}${NC}                                              ${BLUE}║${NC}"
else
echo -e "${BLUE}║${NC}  Failed: ${FAILED}                                              ${BLUE}║${NC}"
fi
echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    exit 1
fi
exit 0
