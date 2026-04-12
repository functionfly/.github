#!/bin/bash
# Comprehensive test script for calling public registry functions and tracking usage
# This script only CALLS public functions, does NOT publish anything

# Configuration
API_BASE_URL="${API_URL:-http://localhost:8080}"
EMAIL="traseputallaz@gmail.com"
PASSWORD="Dogster1996@"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "FunctionFly Registry Usage Test - Complete"
echo "=========================================="
echo "API URL: $API_BASE_URL"
echo "Email: $EMAIL"
echo "Purpose: Test all public function endpoints and track usage"
echo ""

# Step 1: Login
echo -e "${YELLOW}Step 1: Authenticating...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

# Check if login succeeded
if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✓ Login successful!${NC}"
    echo "  Token: ${TOKEN:0:30}..."
    echo ""
else
    echo -e "${RED}✗ Login failed!${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

# Step 2: Get user apps (with auth)
echo -e "${YELLOW}Step 2: Fetching user's apps (for app-based execution)...${NC}"
APPS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/v1/apps" \
  -H "Authorization: Bearer $TOKEN")

if echo "$APPS_RESPONSE" | grep -q '\['; then
    echo -e "${GREEN}✓ Apps endpoint accessible${NC}"
    echo "$APPS_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$APPS_RESPONSE"
    echo ""
else
    echo -e "${YELLOW}Note: Could not retrieve apps${NC}"
    echo "Response: $APPS_RESPONSE"
    echo ""
fi

# Step 3: List all registry functions
echo -e "${YELLOW}Step 3: Listing all registry functions...${NC}"
ALL_FUNCTIONS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions?limit=100" \
  -H "Authorization: Bearer $TOKEN")

FUNCTION_COUNT=$(echo "$ALL_FUNCTIONS" | grep -o '"author"' | wc -l)
echo -e "${GREEN}Found $FUNCTION_COUNT function(s) total${NC}"
echo ""

# Step 4: List public functions
echo -e "${YELLOW}Step 4: Filtering public functions...${NC}"
PUBLIC_FUNCTIONS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions?visibility=public&limit=100" \
  -H "Authorization: Bearer $TOKEN")

PUBLIC_COUNT=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"author"' | wc -l)
echo -e "${GREEN}Found $PUBLIC_COUNT public function(s)${NC}"

if [ "$PUBLIC_COUNT" -gt 0 ]; then
    # Extract function details
    echo ""
    echo "Public functions:"
    echo "$PUBLIC_FUNCTIONS" | python3 -m json.tool 2>/dev/null | head -50 || echo "$PUBLIC_FUNCTIONS"
    echo ""

    # Get first public function
    FIRST_AUTHOR=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"author":"[^"]*"' | head -1 | cut -d'"' -f4)
    FIRST_NAME=$(echo "$PUBLIC_FUNCTIONS" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)

    echo -e "${BLUE}Selected function: $FIRST_AUTHOR/$FIRST_NAME${NC}"
    echo ""

    # Step 5: Get detailed function info
echo -e "${YELLOW}Step 5: Getting detailed function info...${NC}"
    FUNC_DETAILS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$FUNC_DETAILS" | grep -q '"author"'; then
        echo -e "${GREEN}✓ Function details retrieved${NC}"
        echo "$FUNC_DETAILS" | python3 -m json.tool 2>/dev/null || echo "$FUNC_DETAILS"
        echo ""

        # Check if function has versions
        VERSIONS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/versions" \
          -H "Authorization: Bearer $TOKEN")
        VERSION_COUNT=$(echo "$VERSIONS" | grep -c '{' || echo "0")

        echo -e "${YELLOW}Step 6: Checking function versions...${NC}"
        if [ "$VERSION_COUNT" -gt 0 ]; then
            echo -e "${GREEN}✓ Found $VERSION_COUNT version(s)${NC}"
            echo "$VERSIONS" | python3 -m json.tool 2>/dev/null || echo "$VERSIONS"
            echo ""

            # Execute the function
            echo -e "${YELLOW}Step 7: Executing public function...${NC}"
            echo "  POST /v1/fx/$FIRST_AUTHOR/$FIRST_NAME"

            EXEC_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL/v1/fx/$FIRST_AUTHOR/$FIRST_NAME" \
              -H "Authorization: Bearer $TOKEN" \
              -H "Content-Type: application/json" \
              -d '{"test":true,"message":"Hello from registry test","timestamp":'$(date +%s)'}')

            HTTP_CODE=$(echo "$EXEC_RESPONSE" | tail -1)
            BODY=$(echo "$EXEC_RESPONSE" | sed '$d')

            echo "  HTTP Status: $HTTP_CODE"
            echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
            echo ""

            # Check execution result
            if echo "$BODY" | grep -q '"ok":true'; then
                EXEC_ID=$(echo "$BODY" | grep -o '"execution_id":"[^"]*"' | cut -d'"' -f4)
                echo -e "${GREEN}✓ Execution successful (ID: $EXEC_ID)${NC}"
                echo ""

                # Get stats
                echo -e "${YELLOW}Step 8: Getting function stats...${NC}"
                STATS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/stats" \
                  -H "Authorization: Bearer $TOKEN")
                echo "$STATS" | python3 -m json.tool 2>/dev/null || echo "$STATS"
                echo ""

                # Record for recommendations/usage tracking
                echo -e "${YELLOW}Step 9: Recording for usage tracking...${NC}"
                FUNC_ID=$(echo "$FUNC_DETAILS" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
                if [ -n "$FUNC_ID" ]; then
                    REC_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/recommendations/executions" \
                      -H "Authorization: Bearer $TOKEN" \
                      -H "Content-Type: application/json" \
                      -d "{\"function_id\":\"$FUNC_ID\",\"session_id\":\"test_session_$(date +%s)\"}")
                    echo "$REC_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$REC_RESPONSE"
                    echo ""
                fi

                # Check replay if we have execution ID
                if [ -n "$EXEC_ID" ]; then
                    echo -e "${YELLOW}Step 10: Checking execution replay...${NC}"
                    REPLAY=$(curl -s -X GET "$API_BASE_URL/v1/registry/replay/$EXEC_ID" \
                      -H "Authorization: Bearer $TOKEN")
                    echo "$REPLAY" | python3 -m json.tool 2>/dev/null || echo "$REPLAY"
                    echo ""
                fi

            else
                echo -e "${RED}✗ Execution failed (HTTP $HTTP_CODE)${NC}"
                ERROR_MSG=$(echo "$BODY" | grep -o '"message":"[^"]*"' | head -1 | cut -d'"' -f4)
                if [ -n "$ERROR_MSG" ]; then
                    echo "  Error: $ERROR_MSG"
                fi
                echo ""

                # Try the test endpoint
                echo -e "${YELLOW}Attempting test endpoint...${NC}"
                TEST_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/test" \
                  -H "Authorization: Bearer $TOKEN" \
                  -H "Content-Type: application/json" \
                  -d '{"test":"validation"}')
                echo "Test Response:"
                echo "$TEST_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$TEST_RESPONSE"
                echo ""
            fi
        else
            echo -e "${YELLOW}Function has no published versions${NC}"
            echo "Available versions: $VERSIONS"
            echo ""

            # Try to use playground endpoint
            echo -e "${YELLOW}Attempting playground UI endpoint...${NC}"
            PLAYGROUND=$(curl -s -X GET "$API_BASE_URL/v1/run/$FIRST_AUTHOR/$FIRST_NAME" \
              -H "Authorization: Bearer $TOKEN")
            echo "Playground Response (HTML, first 500 chars):"
            echo "$PLAYGROUND" | head -500
            echo ""
        fi
    else
        echo -e "${RED}✗ Failed to get function details${NC}"
        echo "Response: $FUNC_DETAILS"
    fi
else
    echo -e "${YELLOW}No public functions found${NC}"
fi

# Step 11: Search for any available functions
echo -e "${YELLOW}Step 11: Searching for available functions...${NC}"
SEARCH=$(curl -s -X GET "$API_BASE_URL/v1/registry/search?q=a" \
  -H "Authorization: Bearer $TOKEN")
SEARCH_COUNT=$(echo "$SEARCH" | grep -o '"author"' | wc -l)
echo -e "${GREEN}Search found $SEARCH_COUNT function(s)${NC}"
if [ "$SEARCH_COUNT" -gt 0 ]; then
    echo "$SEARCH" | python3 -m json.tool 2>/dev/null | head -50 || echo "$SEARCH"
fi
echo ""

# Step 12: Check recommendations
echo -e "${YELLOW}Step 12: Getting recommendations...${NC}"
RECOMMENDATIONS=$(curl -s -X GET "$API_BASE_URL/v1/recommendations?limit=10" \
  -H "Authorization: Bearer $TOKEN")
echo "$RECOMMENDATIONS" | python3 -m json.tool 2>/dev/null | head -100 || echo "$RECOMMENDATIONS"
echo ""

echo "=========================================="
echo "Test Complete!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Account: $EMAIL"
echo "  - Public functions found: $PUBLIC_COUNT"
echo "  - Total functions found: $FUNCTION_COUNT"
echo "  - Actions: List, Search, Get Details, Execute (attempted)"
echo ""
echo -e "${BLUE}Note: This test only CALLS public functions. No publishing occurred.${NC}"
