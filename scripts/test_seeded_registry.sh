#!/bin/bash
# Test the 1021 seeded public functions

API_BASE_URL="http://localhost:8080"
EMAIL="traseputallaz@gmail.com"
PASSWORD="Dogster1996@"

echo "=========================================="
echo "TESTING SEEDED REGISTRY (1021 functions)"
echo "=========================================="
echo ""

# 1. Login
echo "1. Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "   FAILED: Login error"
    echo "   Response: $LOGIN_RESPONSE"
    exit 1
fi
echo "   SUCCESS: Got JWT token"
echo ""

# 2. Get a specific function
echo "2. Get function details (functionfly/crypto-price)..."
FUNC_RESPONSE=$(curl -s "$API_BASE_URL/v1/registry/functions/functionfly/crypto-price" \
  -H "Authorization: Bearer $TOKEN")
if echo "$FUNC_RESPONSE" | grep -q "functionfly"; then
    echo "   SUCCESS: Found function"
    echo "   Name: $(echo "$FUNC_RESPONSE" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)"
    echo "   Visibility: $(echo "$FUNC_RESPONSE" | grep -o '"visibility":"[^"]*"' | head -1 | cut -d'"' -f4)"
else
    echo "   Response: $FUNC_RESPONSE"
fi
echo ""

# 3. Execute a function
echo "3. Execute function (functionfly/crypto-price)..."
EXEC_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/fx/functionfly/crypto-price" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC"}')
echo "   Response: $EXEC_RESPONSE"
echo ""

# 4. Get function stats
echo "4. Get function stats..."
STATS_RESPONSE=$(curl -s "$API_BASE_URL/v1/registry/functions/functionfly/crypto-price/stats" \
  -H "Authorization: Bearer $TOKEN")
echo "   Response: $STATS_RESPONSE"
echo ""

# 5. Search functions
echo "5. Search for functions (query: 'crypto')..."
SEARCH_RESPONSE=$(curl -s "$API_BASE_URL/v1/registry/search?q=crypto" \
  -H "Authorization: Bearer $TOKEN")
echo "   Response: $SEARCH_RESPONSE" | head -300
echo ""

# 6. List some functions
echo "6. List first 5 public functions..."
LIST_RESPONSE=$(curl -s "$API_BASE_URL/v1/registry/functions?visibility=public&limit=5" \
  -H "Authorization: Bearer $TOKEN")
echo "   Response: $LIST_RESPONSE" | head -500
echo ""

# 7. Execute another function
echo "7. Execute text-classification function..."
CLASSIFY_RESPONSE=$(curl -s -X POST "$API_BASE_URL/v1/fx/functionfly/text-classification" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"The stock market rallied today"}')
echo "   Response: $CLASSIFY_RESPONSE"
echo ""

echo "=========================================="
echo "TEST COMPLETE"
echo "=========================================="
