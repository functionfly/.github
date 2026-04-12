#!/bin/bash
# Comprehensive test script for registry function calling

API_BASE_URL="${API_URL:-http://localhost:8080}"
EMAIL="traseputallaz@gmail.com"
PASSWORD="Dogster1996@"

echo "=========================================="
echo "FunctionFly Registry - Full Test Suite"
echo "=========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Login
echo "Step 1: Authenticating..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    TOKEN=$(echo "$BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$BODY" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✓ Login successful!${NC}"
    echo "  Token: ${TOKEN:0:30}..."
    echo ""
else
    echo -e "${RED}✗ Login failed!${NC}"
    echo "  HTTP Code: $HTTP_CODE"
    echo "  Response: $BODY"
    exit 1
fi

# Step 2: Get user profile
echo "Step 2: Getting user profile..."
PROFILE_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$PROFILE_RESPONSE" | tail -n1)
BODY=$(echo "$PROFILE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    USER_EMAIL=$(echo "$BODY" | grep -o '"email":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}✓ Profile retrieved${NC}"
    echo "  User: $USER_EMAIL"
    echo ""
else
    echo -e "${YELLOW}⚠ Failed to get profile${NC}"
    echo ""
fi

# Step 3: List registry functions
echo "Step 3: Fetching registry functions..."
FUNCTIONS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/v1/registry/functions" \
  -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$FUNCTIONS_RESPONSE" | tail -n1)
BODY=$(echo "$FUNCTIONS_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    # Count functions
    FUNCTION_COUNT=$(echo "$BODY" | grep -o '"function_id"' | wc -l)
    echo -e "${GREEN}✓ Found $FUNCTION_COUNT function(s)${NC}"
    echo ""

    # Extract all function author/name pairs
    echo "Available Functions:"
    echo "$BODY" | grep -o '"author":"[^"]*","name":"[^"]*"' | while read -r line; do
        author=$(echo "$line" | grep -o '"author":"[^"]*"' | cut -d'"' -f4)
        name=$(echo "$line" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
        echo "  - $author/$name"
    done
    echo ""
else
    echo -e "${RED}✗ Failed to fetch functions${NC}"
    echo "  Response: $BODY"
    exit 1
fi

# Step 4: Test multiple functions
echo "Step 4: Testing function execution..."
echo ""

# Get first 3 function names to test
FUNCTIONS=$(echo "$BODY" | grep -o '"author":"[^"]*","name":"[^"]*"' | head -3)

TEST_COUNT=0
SUCCESS_COUNT=0

echo "$FUNCTIONS" | while read -r line; do
    author=$(echo "$line" | grep -o '"author":"[^"]*"' | cut -d'"' -f4)
    name=$(echo "$line" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$author" ] && [ -n "$name" ]; then
        echo "Testing: $author/$name"

        # Try to execute
        EXEC_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/v1/fx/$author/$name" \
          -H "Authorization: Bearer $TOKEN" \
          -H "Content-Type: application/json" \
          -d '{"input":{"test":true,"timestamp":'$(date +%s)'}}')

        HTTP_CODE=$(echo "$EXEC_RESPONSE" | tail -n1)
        BODY=$(echo "$EXEC_RESPONSE" | sed '$d')

        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "  ${GREEN}✓ Success${NC} (HTTP $HTTP_CODE)"
            # Show first 100 chars of response
            echo "  Response: $(echo "$BODY" | cut -c1-100)..."
        else
            echo -e "  ${YELLOW}⚠ Issue${NC} (HTTP $HTTP_CODE)"
            # Extract error message if present
            ERROR_MSG=$(echo "$BODY" | grep -o '"message":"[^"]*"' | head -1 | cut -d'"' -f4)
            if [ -n "$ERROR_MSG" ]; then
                echo "  Error: $ERROR_MSG"
            fi
        fi
        echo ""
    fi
done

# Step 5: Search functions
echo "Step 5: Testing function search..."
SEARCH_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/v1/registry/search?q=test" \
  -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$SEARCH_RESPONSE" | tail -n1)
BODY=$(echo "$SEARCH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    RESULT_COUNT=$(echo "$BODY" | grep -o '"function_id"' | wc -l)
    echo -e "${GREEN}✓ Search working${NC} - Found $RESULT_COUNT results"
    echo ""
else
    echo -e "${YELLOW}⚠ Search failed${NC} (HTTP $HTTP_CODE)"
    echo ""
fi

# Step 6: Get recommendations
echo "Step 6: Getting recommendations..."
REC_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE_URL/v1/recommendations" \
  -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$REC_RESPONSE" | tail -n1)
BODY=$(echo "$REC_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Recommendations retrieved${NC}"
    echo ""
else
    echo -e "${YELLOW}⚠ Recommendations unavailable${NC} (HTTP $HTTP_CODE)"
    echo ""
fi

# Step 7: Logout (optional - revoke token)
echo "Step 7: Revoking token..."
LOGOUT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/v1/auth/logout" \
  -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$LOGOUT_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    echo -e "${GREEN}✓ Logout successful${NC}"
else
    echo -e "${YELLOW}⚠ Logout returned HTTP $HTTP_CODE${NC}"
fi

echo ""
echo "=========================================="
echo "Test Suite Complete!"
echo "=========================================="
