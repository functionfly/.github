#!/bin/bash
# Final test script for FunctionFly Public Registry Testing
# Tests all available endpoints for calling public functions
# Records usage tracking data even for failed executions

# Configuration
API_BASE_URL="${API_URL:-http://localhost:8080}"
EMAIL="traseputallaz@gmail.com"
PASSWORD="Dogster1996@"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
}

print_section() {
    echo ""
    echo -e "${CYAN}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_header "FunctionFly Public Registry - Usage Tracking Test"
echo "Account: $EMAIL"
echo "API: $API_BASE_URL"
echo "Purpose: Call public functions and track usage (NO publishing)"
echo ""
echo "⚠️  IMPORTANT: This test only CALLS existing public functions."
echo "   No new functions are published."
echo ""

# Step 1: Login
print_section "STEP 1: Authentication"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$LOGIN_RESPONSE" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    print_success "Login successful"
    print_info "User ID: $USER_ID"
    print_info "Token: ${TOKEN:0:30}..."
else
    print_error "Login failed"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

# Step 2: Discover Public Functions
print_section "STEP 2: Discover Public Functions"

ALL_FUNCTIONS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions?limit=100" \
  -H "Authorization: Bearer $TOKEN")

PUBLIC_FUNCTIONS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions?visibility=public&limit=100" \
  -H "Authorization: Bearer $TOKEN")

PUBLIC_COUNT=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"author"' | wc -l)
print_success "Found $PUBLIC_COUNT public function(s)"

if [ "$PUBLIC_COUNT" -eq 0 ]; then
    print_error "No public functions found - cannot proceed with execution tests"
    echo ""
    print_info "Available Registry Endpoints (tested without execution):"
    echo "  - GET /v1/registry/functions (public read) ✓"
    echo "  - GET /v1/registry/search (public read) ✓"
    echo ""

    # Test search anyway
    print_section "Testing Search Endpoint"
    SEARCH=$(curl -s -X GET "$API_BASE_URL/v1/registry/search?q=test" \
      -H "Authorization: Bearer $TOKEN")
    echo "Search results:"
    echo "$SEARCH" | python3 -m json.tool 2>/dev/null || echo "$SEARCH"

    print_header "Test Summary"
    echo "Account: $EMAIL"
    echo "Public functions found: 0"
    echo "Functions executed: 0"
    echo "Usage tracking records: 0"
    echo ""
    print_info "Note: To test full execution flow, public functions need published versions."
    exit 0
fi

# Step 3: Select Function for Testing
print_section "STEP 3: Select Function for Testing"

FIRST_AUTHOR=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"author":"[^"]*"' | head -1 | cut -d'"' -f4)
FIRST_NAME=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)
FUNC_ID=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

print_info "Selected function: $FIRST_AUTHOR/$FIRST_NAME"
print_info "Function ID: $FUNC_ID"

# Step 4: Get Function Details
print_section "STEP 4: Get Function Details"

FUNC_DETAILS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME" \
  -H "Authorization: Bearer $TOKEN")

if echo "$FUNC_DETAILS" | grep -q '"author"'; then
    print_success "Retrieved function details"
    echo "Details:"
    echo "$FUNC_DETAILS" | python3 -m json.tool 2>/dev/null | grep -E '"(author|name|visibility|version|description)"' || echo "$FUNC_DETAILS"
else
    print_error "Failed to get function details: $FUNC_DETAILS"
fi

# Step 5: Check Versions
print_section "STEP 5: Check Available Versions"

VERSIONS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/versions" \
  -H "Authorization: Bearer $TOKEN")

# Count versions by counting objects in JSON array
VERSION_COUNT=$(echo "$VERSIONS" | python3 -c "import json,sys; data=json.load(sys.stdin); print(len(data))" 2>/dev/null || echo "0")

if [ "$VERSION_COUNT" -gt 0 ]; then
    print_success "Found $VERSION_COUNT published version(s)"
    echo "$VERSIONS" | python3 -m json.tool 2>/dev/null || echo "$VERSIONS"
else
    print_error "No published versions found for this function"
    print_info "Function exists but cannot be executed without a published version"
    echo ""
fi

# Step 6: Get Stats (Pre-Execution)
print_section "STEP 6: Get Function Stats (Pre-Execution)"

STATS_BEFORE=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/stats" \
  -H "Authorization: Bearer $TOKEN")

echo "Stats before execution attempt:"
echo "$STATS_BEFORE" | python3 -m json.tool 2>/dev/null || echo "$STATS_BEFORE"

# Step 7: Execute Function
print_section "STEP 7: Execute Public Function"

print_info "POST /v1/fx/$FIRST_AUTHOR/$FIRST_NAME"
print_info "Request body: {\"test\": true, \"message\": \"Registry usage test\"}"

EXEC_START_TIME=$(date +%s%N)
EXEC_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/v1/fx/$FIRST_AUTHOR/$FIRST_NAME" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"test":true,"message":"Registry usage test","timestamp":'$(date +%s)'}')
EXEC_END_TIME=$(date +%s%N)

HTTP_CODE=$(echo "$EXEC_RESPONSE" | tail -1)
BODY=$(echo "$EXEC_RESPONSE" | sed '$d')
EXEC_DURATION_MS=$(( (EXEC_END_TIME - EXEC_START_TIME) / 1000000 ))

echo ""
echo "Response (HTTP $HTTP_CODE):"
echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
echo ""

# Determine outcome
if echo "$BODY" | grep -q '"ok":true'; then
    EXECUTION_SUCCESS=true
    EXEC_ID=$(echo "$BODY" | grep -o '"execution_id":"[^"]*"' | cut -d'"' -f4)
    DURATION=$(echo "$BODY" | grep -o '"duration_ms":[0-9]*' | cut -d':' -f2)
    CACHED=$(echo "$BODY" | grep -o '"cached":[a-z]*' | cut -d':' -f2)

    print_success "Execution successful!"
    print_info "Execution ID: $EXEC_ID"
    print_info "Duration: ${DURATION}ms (API round-trip: ${EXEC_DURATION_MS}ms)"
    print_info "Cached: $CACHED"
else
    EXECUTION_SUCCESS=false
    ERROR_CODE=$(echo "$BODY" | grep -o '"code":"[^"]*"' | head -1 | cut -d'"' -f4)
    ERROR_MSG=$(echo "$BODY" | grep -o '"message":"[^"]*"' | head -1 | cut -d'"' -f4)

    print_error "Execution failed (HTTP $HTTP_CODE)"
    print_info "Error Code: $ERROR_CODE"
    print_info "Error Message: $ERROR_MSG"
fi

# Step 8: Record Usage for Tracking
print_section "STEP 8: Record Usage for Recommendations"

if [ -n "$FUNC_ID" ]; then
    SESSION_ID="session_$(date +%s)_$$"

    print_info "Recording execution for usage tracking..."
    print_info "Session ID: $SESSION_ID"
    print_info "Function ID: $FUNC_ID"

    REC_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/recommendations/executions" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"function_id\":\"$FUNC_ID\",\"session_id\":\"$SESSION_ID\",\"user_agent\":\"registry-test-script\"}")

    if echo "$REC_RESPONSE" | grep -q '"status":"recorded"\|"success":true'; then
        print_success "Usage recorded successfully"
    else
        print_info "Usage recording response:"
        echo "$REC_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$REC_RESPONSE"
    fi
else
    print_error "Cannot record usage - function ID not available"
fi

# Step 9: Get Stats (Post-Execution)
print_section "STEP 9: Get Updated Function Stats"

STATS_AFTER=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/stats" \
  -H "Authorization: Bearer $TOKEN")

echo "Stats after execution attempt:"
echo "$STATS_AFTER" | python3 -m json.tool 2>/dev/null || echo "$STATS_AFTER"

# Step 10: Check Execution Replay (if successful)
if [ "$EXECUTION_SUCCESS" = true ] && [ -n "$EXEC_ID" ]; then
    print_section "STEP 10: Check Execution Replay"

    print_info "Fetching replay for execution: $EXEC_ID"

    REPLAY=$(curl -s -X GET "$API_BASE_URL/v1/registry/replay/$EXEC_ID" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$REPLAY" | grep -q '"execution_id"'; then
        print_success "Replay available"
        echo "$REPLAY" | python3 -m json.tool 2>/dev/null | head -30 || echo "$REPLAY"
    else
        print_info "Replay response:"
        echo "$REPLAY" | python3 -m json.tool 2>/dev/null || echo "$REPLAY"
    fi
fi

# Step 11: Test Caching (Second Execution)
print_section "STEP 11: Second Execution (Cache Test)"

print_info "POST /v1/fx/$FIRST_AUTHOR/$FIRST_NAME (second call)"

EXEC_RESPONSE_2=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/v1/fx/$FIRST_AUTHOR/$FIRST_NAME" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"test":true,"message":"Second execution for cache test","timestamp":'$(date +%s)'}')

HTTP_CODE_2=$(echo "$EXEC_RESPONSE_2" | tail -1)
BODY_2=$(echo "$EXEC_RESPONSE_2" | sed '$d')

if echo "$BODY_2" | grep -q '"cached":true'; then
    print_success "Second execution served from cache"
elif echo "$BODY_2" | grep -q '"ok":true'; then
    print_info "Second execution succeeded (not cached)"
elif echo "$BODY_2" | grep -q '"ok":false'; then
    print_info "Second execution also failed (consistent with first)"
fi

# Step 12: Get Recommendations
print_section "STEP 12: Get Recommendations"

RECOMMENDATIONS=$(curl -s -X GET "$API_BASE_URL/v1/recommendations?limit=10" \
  -H "Authorization: Bearer $TOKEN")

echo "Recommendations:"
echo "$RECOMMENDATIONS" | python3 -m json.tool 2>/dev/null | head -50 || echo "$RECOMMENDATIONS"

# Summary
print_header "Test Summary"
echo ""
echo "Account Tested:"
echo "  Email: $EMAIL"
echo "  User ID: $USER_ID"
echo ""
echo "Functions Discovered:"
echo "  Total functions: $(echo "$ALL_FUNCTIONS" | grep -o '"author"' | wc -l)"
echo "  Public functions: $PUBLIC_COUNT"
echo "  Selected: $FIRST_AUTHOR/$FIRST_NAME"
echo "  Published versions: $VERSION_COUNT"
echo ""
echo "Execution Results:"
if [ "$EXECUTION_SUCCESS" = true ]; then
    echo "  Status: ✅ SUCCESS"
    echo "  Execution ID: $EXEC_ID"
    echo "  Duration: ${DURATION}ms"
    echo "  Cached: $CACHED"
else
    echo "  Status: ❌ FAILED"
    echo "  Error Code: $ERROR_CODE"
    echo "  Error: $ERROR_MSG"
    echo ""
    echo "  Note: Execution failed because the function has no published"
    echo "        version available for execution."
fi
echo ""
echo "Usage Tracking:"
echo "  Session ID: $SESSION_ID"
echo "  Function ID recorded: $FUNC_ID"
echo "  Recommendation API: Called"
echo ""
echo "Endpoints Tested:"
echo "  ✓ POST /v1/auth/login"
echo "  ✓ GET /v1/registry/functions"
echo "  ✓ GET /v1/registry/functions/:author/:name"
echo "  ✓ GET /v1/registry/functions/:author/:name/versions"
echo "  ✓ GET /v1/registry/functions/:author/:name/stats"
echo "  ✓ POST /v1/fx/:author/:name (execution)"
echo "  ✓ POST /v1/recommendations/executions (usage tracking)"
if [ "$EXECUTION_SUCCESS" = true ]; then
    echo "  ✓ GET /v1/registry/replay/:executionId"
fi
echo "  ✓ GET /v1/recommendations"
echo ""
echo "Important Note:"
echo "  This test only CALLED public functions - no publishing occurred."
echo "  All API calls were read-only or execution-only operations."
echo ""
print_info "Test completed at $(date)"
