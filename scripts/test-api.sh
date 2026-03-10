#!/bin/bash

# FunctionFly MVP1 API Test Script
# Run this after starting PostgreSQL and running migrations + setup

echo "=== FunctionFly MVP1 API Test ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080/v1"

# Test health endpoint
echo "Testing health endpoint..."
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200" && echo -e "${GREEN}✓ Health check passed${NC}" || echo -e "${RED}✗ Health check failed${NC}"

echo

# Login
echo "Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}')

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
  echo -e "${GREEN}✓ Login successful${NC}"

  # Extract token
  TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

  echo "Token: ${TOKEN:0:50}..."
  echo

  # Create app
  echo "Testing app creation..."
  TIMESTAMP=$(date +%s)
  APP_RESPONSE=$(curl -s -X POST $BASE_URL/apps \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"name\":\"Test App\",\"slug\":\"test-app-$TIMESTAMP\"}")

  if echo "$APP_RESPONSE" | grep -q "id"; then
    echo -e "${GREEN}✓ App created successfully${NC}"

    # Extract app ID
    APP_ID=$(echo "$APP_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4 | head -1)
    echo "App ID: $APP_ID"
    echo

    # Create backend
    echo "Testing backend creation..."
    BACKEND_RESPONSE=$(curl -s -X POST $BASE_URL/apps/$APP_ID/backends \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{
        "provider": "workers",
        "region": "us-east-1",
        "url": "https://test-backend.example.com"
      }')

    if echo "$BACKEND_RESPONSE" | grep -q "id"; then
      echo -e "${GREEN}✓ Backend created successfully${NC}"

      # Extract backend ID
      BACKEND_ID=$(echo "$BACKEND_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4 | head -1)
      echo "Backend ID: $BACKEND_ID"
      echo

      # List backends
      echo "Testing backend listing..."
      LIST_RESPONSE=$(curl -s $BASE_URL/apps/$APP_ID/backends \
        -H "Authorization: Bearer $TOKEN")

      if echo "$LIST_RESPONSE" | grep -q "backends"; then
        echo -e "${GREEN}✓ Backend listing successful${NC}"
        echo "Backends: $(echo "$LIST_RESPONSE" | grep -o '"backends":\[[^]]*\]' | head -1)"
        echo
      else
        echo -e "${RED}✗ Backend listing failed${NC}"
        echo "Response: $LIST_RESPONSE"
      fi

      # Test routing decision
      echo "Testing routing decision..."
      ROUTE_RESPONSE=$(curl -s "$BASE_URL/apps/$APP_ID/route?method=GET&request_id=test-123" \
        -H "Authorization: Bearer $TOKEN")

      if echo "$ROUTE_RESPONSE" | grep -q "reason"; then
        echo -e "${GREEN}✓ Routing decision successful${NC}"
        echo "Decision: $(echo "$ROUTE_RESPONSE" | grep -o '"reason":"[^"]*"' | cut -d'"' -f4)"
      else
        echo -e "${YELLOW}⚠ Routing decision returned: $ROUTE_RESPONSE${NC}"
      fi

    else
      echo -e "${RED}✗ Backend creation failed${NC}"
      echo "Response: $BACKEND_RESPONSE"
    fi

  else
    echo -e "${RED}✗ App creation failed${NC}"
    echo "Response: $APP_RESPONSE"
  fi

else
  echo -e "${RED}✗ Login failed${NC}"
  echo "Response: $LOGIN_RESPONSE"
fi

echo
echo "=== Test Complete ==="
echo
echo "To start the services:"
echo "1. Start PostgreSQL: docker run -d --name postgres -e POSTGRES_DB=functionfly -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:15"
echo "2. Run migrations: make migrate"
echo "3. Run setup: make setup"
echo "4. Start API: make api"
echo "5. Start health monitor (in another terminal): make health-monitor"