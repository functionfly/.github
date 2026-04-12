#!/bin/bash
# Test script for logging in and calling a registry function

# Configuration
API_BASE_URL="${API_URL:-http://localhost:8080}"
EMAIL="traseputallaz@gmail.com"
PASSWORD="Dogster1996@"

echo "=========================================="
echo "FunctionFly Registry Function Call Test"
echo "=========================================="
echo "API URL: $API_BASE_URL"
echo "Email: $EMAIL"
echo ""

# Step 1: Login
echo "Step 1: Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

# Check if login succeeded
if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "✓ Login successful!"
    echo "Token: ${TOKEN:0:20}..."
    echo ""
else
    echo "✗ Login failed!"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

# Step 2: List registry functions
echo "Step 2: Fetching registry functions..."
FUNCTIONS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions" \
  -H "Authorization: Bearer $TOKEN")

# Check if any functions exist
if echo "$FUNCTIONS_RESPONSE" | grep -q '\['; then
    FUNCTION_COUNT=$(echo "$FUNCTIONS_RESPONSE" | grep -o '"author"' | wc -l)
    echo "✓ Found $FUNCTION_COUNT function(s) in registry"
    echo ""

    # Extract first function author and name
    FIRST_AUTHOR=$(echo "$FUNCTIONS_RESPONSE" | grep -o '"author":"[^"]*"' | head -1 | cut -d'"' -f4)
    FIRST_NAME=$(echo "$FUNCTIONS_RESPONSE" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -n "$FIRST_AUTHOR" ] && [ -n "$FIRST_NAME" ]; then
        echo "Testing with function: $FIRST_AUTHOR/$FIRST_NAME"
        echo ""

        # Step 3: Get function details
        echo "Step 3: Getting function details..."
        FUNCTION_DETAILS=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME" \
          -H "Authorization: Bearer $TOKEN")

        if echo "$FUNCTION_DETAILS" | grep -q '"author"'; then
            echo "✓ Function details retrieved"
            echo ""

            # Step 4: Test the function
            echo "Step 4: Testing function execution..."
            TEST_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/test" \
              -H "Authorization: Bearer $TOKEN" \
              -H "Content-Type: application/json" \
              -d '{"test":"validation","timestamp":'$(date +%s)'}')

            echo "Test Response:"
            echo "$TEST_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$TEST_RESPONSE"
            echo ""

            # Step 5: Execute the function (if available)
            echo "Step 5: Executing function..."
            EXEC_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/fx/$FIRST_AUTHOR/$FIRST_NAME" \
              -H "Authorization: Bearer $TOKEN" \
              -H "Content-Type: application/json" \
              -d '{"input":{"test":true,"message":"Hello from test script"}}')

            echo "Execution Response:"
            echo "$EXEC_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$EXEC_RESPONSE"
            echo ""

            # Step 6: Get function stats
            echo "Step 6: Getting function stats..."
            STATS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/v1/registry/functions/$FIRST_AUTHOR/$FIRST_NAME/stats" \
              -H "Authorization: Bearer $TOKEN")

            echo "Stats Response:"
            echo "$STATS_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$STATS_RESPONSE"
            echo ""

        else
            echo "✗ Failed to get function details"
            echo "Response: $FUNCTION_DETAILS"
        fi
    else
        echo "No functions found in registry to test"
    fi
else
    echo "No functions found in registry or error occurred"
    echo "Response: $FUNCTIONS_RESPONSE"
fi

echo "=========================================="
echo "Test Complete!"
echo "=========================================="
