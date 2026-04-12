#!/bin/bash
# Test script for calling public registry functions and tracking usage
# This script only CALLS public functions, does NOT publish anything

# Configuration
API_BASE_URL="${API_URL:-http://localhost:8080}"
EMAIL="traseputallaz@gmail.com"
PASSWORD="Dogster1996@"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "FunctionFly Public Registry Usage Test"
echo "=========================================="
echo "API URL: $API_BASE_URL"
echo "Email: $EMAIL"
echo "Purpose: Call public functions and track usage"
echo ""

# Step 1: Login
echo -e "${YELLOW}Step 1: Authenticating...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

# Check if login succeeded
if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$LOGIN_RESPONSE" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✓ Login successful!${NC}"
    echo "  Token: ${TOKEN:0:20}..."
    echo "  User ID: $USER_ID"
    echo ""
else
    echo -e "${RED}✗ Login failed!${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

# Step 2: List public functions
echo -e "${YELLOW}Step 2: Fetching public functions from registry...${NC}"
FUNCTIONS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions?visibility=public&limit=10" \
  -H "Authorization: Bearer $TOKEN")

# Check if any functions exist
if echo "$FUNCTIONS_RESPONSE" | grep -q '"author"'; then
    FUNCTION_COUNT=$(echo "$FUNCTIONS_RESPONSE" | grep -o '"author"' | wc -l)
    echo -e "${GREEN}✓ Found $FUNCTION_COUNT public function(s) in registry${NC}"
    echo ""

    # Extract first public function author and name
    FIRST_AUTHOR=$(echo "$FUNCTIONS_RESPONSE" | grep -o '"author":"[^"]*"' | head -1 | cut -d'"' -f4)
    FIRST_NAME=$(echo "$FUNCTIONS_RESPONSE" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -n "$FIRST_AUTHOR" ] && [ -n "$FIRST_NAME" ]; then
        echo -e "${YELLOW}Testing with public function: $FIRST_AUTHOR/$FIRST_NAME${NC}"
        echo ""

        # Step 3: Get function details
echo -e "${YELLOW}Step 3: Getting function details...${NC}"
        FUNCTION_DETAILS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME" \
          -H "Authorization: Bearer $TOKEN")

        if echo "$FUNCTION_DETAILS" | grep -q '"author"'; then
            echo -e "${GREEN}✓ Function details retrieved${NC}"
            echo "  Author: $(echo "$FUNCTION_DETAILS" | grep -o '"author":"[^"]*"' | head -1 | cut -d'"' -f4)"
            echo "  Name: $(echo "$FUNCTION_DETAILS" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)"
            echo "  Visibility: $(echo "$FUNCTION_DETAILS" | grep -o '"visibility":"[^"]*"' | head -1 | cut -d'"' -f4)"
            echo ""

            # Step 4: Get function stats before execution
echo -e "${YELLOW}Step 4: Getting initial function stats...${NC}"
            STATS_BEFORE=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/stats" \
              -H "Authorization: Bearer $TOKEN")
            echo "Stats before execution:"
            echo "$STATS_BEFORE" | python3 -m json.tool 2>/dev/null || echo "$STATS_BEFORE"
            echo ""

            # Step 5: Execute the public function via API
echo -e "${YELLOW}Step 5: Executing public function via API...${NC}"
            echo "  POST /v1/fx/$FIRST_AUTHOR/$FIRST_NAME"
            EXEC_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/fx/$FIRST_AUTHOR/$FIRST_NAME" \
              -H "Authorization: Bearer $TOKEN" \
              -H "Content-Type: application/json" \
              -d '{"test":true,"message":"Hello from public registry test","timestamp":'$(date +%s)'}')

            echo -e "${GREEN}Execution Response:${NC}"
            echo "$EXEC_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$EXEC_RESPONSE"
            echo ""

            # Check if execution was successful
            if echo "$EXEC_RESPONSE" | grep -q '"ok":true'; then
                EXECUTION_ID=$(echo "$EXEC_RESPONSE" | grep -o '"execution_id":"[^"]*"' | cut -d'"' -f4)
                DURATION=$(echo "$EXEC_RESPONSE" | grep -o '"duration_ms":[0-9]*' | cut -d':' -f2)
                CACHED=$(echo "$EXEC_RESPONSE" | grep -o '"cached":[a-z]*' | cut -d':' -f2)

                echo -e "${GREEN}✓ Function executed successfully${NC}"
                echo "  Execution ID: $EXECUTION_ID"
                echo "  Duration: ${DURATION}ms"
                echo "  Cached: $CACHED"
                echo ""

                # Step 6: Get function stats after execution
echo -e "${YELLOW}Step 6: Getting updated function stats...${NC}"
                STATS_AFTER=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/stats" \
                  -H "Authorization: Bearer $TOKEN")
                echo "Stats after execution:"
                echo "$STATS_AFTER" | python3 -m json.tool 2>/dev/null || echo "$STATS_AFTER"
                echo ""

                # Step 7: Get execution replay if available
echo -e "${YELLOW}Step 7: Checking execution replay...${NC}"
                if [ -n "$EXECUTION_ID" ]; then
                    REPLAY_RESPONSE=$(curl -s -X GET "$API_BASE_URL/v1/registry/replay/$EXECUTION_ID" \
                      -H "Authorization: Bearer $TOKEN")
                    echo "Replay Response:"
                    echo "$REPLAY_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$REPLAY_RESPONSE"
                    echo ""
                fi

                # Step 8: Record execution for recommendations/usage tracking
echo -e "${YELLOW}Step 8: Recording execution for usage tracking...${NC}"
                FUNCTION_ID=$(echo "$FUNCTION_DETAILS" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
                if [ -n "$FUNCTION_ID" ]; then
                    # Get recommendations interactions to track usage
                    REC_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/recommendations/executions" \
                      -H "Authorization: Bearer $TOKEN" \
                      -H "Content-Type: application/json" \
                      -d "{\"function_id\":\"$FUNCTION_ID\",\"session_id\":\"test_session_$(date +%s)\"}")
                    echo "Usage tracking response:"
                    echo "$REC_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$REC_RESPONSE"
                    echo ""
                fi

                # Step 9: Execute multiple times to test caching
echo -e "${YELLOW}Step 9: Executing again to test caching...${NC}"
                EXEC_RESPONSE_2=$(curl -s -X POST "$API_BASE_URL/v1/fx/$FIRST_AUTHOR/$FIRST_NAME" \
                  -H "Authorization: Bearer $TOKEN" \
                  -H "Content-Type: application/json" \
                  -d '{"test":true,"message":"Second execution for cache test","timestamp":'$(date +%s)'}')

                echo "Second Execution Response:"
                echo "$EXEC_RESPONSE_2" | python3 -m json.tool 2>/dev/null || echo "$EXEC_RESPONSE_2"

                # Check cache status
                if echo "$EXEC_RESPONSE_2" | grep -q '"cached":true'; then
                    echo -e "${GREEN}✓ Second execution was served from cache${NC}"
                else
                    echo -e "${YELLOW}Note: Second execution was not cached${NC}"
                fi
                echo ""

            else
                echo -e "${RED}✗ Function execution failed${NC}"
                ERROR_MSG=$(echo "$EXEC_RESPONSE" | grep -o '"message":"[^"]*"' | head -1 | cut -d'"' -f4)
                echo "  Error: $ERROR_MSG"
                echo ""
            fi

        else
            echo -e "${RED}✗ Failed to get function details${NC}"
            echo "Response: $FUNCTION_DETAILS"
        fi
    else
        echo -e "${RED}No public functions found in registry to test${NC}"
    fi
else
    echo -e "${YELLOW}No public functions found or error occurred${NC}"
    echo "Response: $FUNCTIONS_RESPONSE"
fi

echo "=========================================="
echo "Test Complete!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Tested account: $EMAIL"
echo "  - Target: Public registry functions"
echo "  - Actions: List, Get Details, Execute (multiple times)"
echo "  - Usage tracking: Recorded via recommendations API"
echo ""
echo "This test only CALLS public functions. No publishing occurred."
