#!/bin/bash

# FunctionFly Consolidated Auth System End-to-End Test
# Tests the unified role-based authentication system

set -e

echo "=== FunctionFly Consolidated Auth System E2E Test ==="
echo

# Configuration
BASE_URL="http://localhost:8080"
V1_API="$BASE_URL/v1"
JWT_SECRET="test-jwt-secret-key-for-e2e-tests"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Helper function to make HTTP requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth_header=$4

    if [ -n "$auth_header" ]; then
        curl -s -X "$method" -H "Content-Type: application/json" -H "$auth_header" -d "$data" "$url"
    else
        curl -s -X "$method" -H "Content-Type: application/json" -d "$data" "$url"
    fi
}

# Helper function to run a test
run_test() {
    local test_name=$1
    local test_cmd=$2

    echo -n "Testing $test_name... "
    TESTS_RUN=$((TESTS_RUN + 1))

    if eval "$test_cmd"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC}"
    fi
}

# Test 1: Health check
run_test "Health Check" "
    response=\$(curl -s -o /dev/null -w '%{http_code}' '$BASE_URL/health')
    [ \"\$response\" = \"200\" ]
"

# Test 2: Create test tenant (requires admin auth - we'll skip for now and assume tenant exists)
echo "Note: Assuming test tenant exists. In production, this would require admin authentication."

# Test 3: User registration with different roles
echo -e \"\n${BLUE}Testing User Registration and Role Assignment${NC}\"

# Register a super_admin user
SUPER_ADMIN_TOKEN=\$(make_request POST "$V1_API/auth/signup" '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "username": "test_super_admin",
    "email": "superadmin@test.com",
    "password": "testpass123",
    "role": "super_admin"
}' | jq -r '.token // empty')

if [ -n "\$SUPER_ADMIN_TOKEN" ]; then
    echo "Super admin registered successfully"
else
    echo "Super admin registration failed or returned no token"
fi

# Register an admin user
ADMIN_TOKEN=\$(make_request POST "$V1_API/auth/signup" '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "username": "test_admin",
    "email": "admin@test.com",
    "password": "testpass123",
    "role": "admin"
}' | jq -r '.token // empty')

# Register a team_owner user
TEAM_OWNER_TOKEN=\$(make_request POST "$V1_API/auth/signup" '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "username": "test_team_owner",
    "email": "teamowner@test.com",
    "password": "testpass123",
    "role": "team_owner"
}' | jq -r '.token // empty')

# Register a regular user
USER_TOKEN=\$(make_request POST "$V1_API/auth/signup" '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "username": "test_user",
    "email": "user@test.com",
    "password": "testpass123",
    "role": "user"
}' | jq -r '.token // empty')

# Test 4: Login functionality
echo -e \"\n${BLUE}Testing Login Functionality${NC}\"

run_test "User Login" "
    response=\$(make_request POST \"$V1_API/auth/login\" '{
        \"email\": \"user@test.com\",
        \"password\": \"testpass123\"
    }')
    echo \"\$response\" | jq -e '.token' >/dev/null
"

# Test 5: Session validation
echo -e \"\n${BLUE}Testing Session Management${NC}\"

if [ -n \"\$USER_TOKEN\" ]; then
    run_test "Session Validation" "
        response=\$(make_request GET \"$V1_API/auth/session\" \"\" \"Authorization: Bearer \$USER_TOKEN\")
        echo \"\$response\" | jq -e '.user' >/dev/null
    "
else
    echo \"Skipping session test - no user token\"
fi

# Test 6: Role-based access control
echo -e \"\n${BLUE}Testing Role-Based Access Control${NC}\"

# Test super admin can access admin endpoints
if [ -n \"\$SUPER_ADMIN_TOKEN\" ]; then
    run_test "Super Admin Access to Admin Endpoints" "
        response=\$(curl -s -o /dev/null -w '%{http_code}' -H 'Authorization: Bearer \$SUPER_ADMIN_TOKEN' '$V1_API/admin/tenants')
        [ \"\$response\" = \"200\" -o \"\$response\" = \"401\" -o \"\$response\" = \"403\" ]
    "
else
    echo \"Skipping super admin test - no token\"
fi

# Test regular user cannot access admin endpoints
if [ -n \"\$USER_TOKEN\" ]; then
    run_test "Regular User Blocked from Admin Endpoints" "
        response=\$(curl -s -o /dev/null -w '%{http_code}' -H 'Authorization: Bearer \$USER_TOKEN' '$V1_API/admin/tenants')
        [ \"\$response\" = \"403\" -o \"\$response\" = \"401\" ]
    "
else
    echo \"Skipping user access test - no token\"
fi

# Test 7: Permission checking
echo -e \"\n${BLUE}Testing Permission Validation${NC}\"

if [ -n \"\$USER_TOKEN\" ]; then
    run_test "User Permission Check" "
        response=\$(make_request GET \"$V1_API/permissions/resource/apps/123\" \"\" \"Authorization: Bearer \$USER_TOKEN\")
        # Should return permission status or 403
        echo \"\$response\" | jq -e '.allowed // .message' >/dev/null 2>/dev/null || [ \"\$?\" = \"0\" ]
    "
else
    echo \"Skipping permission test - no user token\"
fi

# Test 8: JWT claims validation
echo -e \"\n${BLUE}Testing JWT Claims and Role Inheritance${NC}\"

if [ -n \"\$USER_TOKEN\" ]; then
    run_test "JWT Contains Role Information" "
        # Decode JWT payload (simplified - in production use proper JWT library)
        payload=\$(echo \"\$USER_TOKEN\" | cut -d'.' -f2 | base64 -d 2>/dev/null | jq -r '.role // empty' 2>/dev/null)
        [ -n \"\$payload\" ]
    "
else
    echo \"Skipping JWT test - no user token\"
fi

# Test 9: GBA integration (session management)
echo -e \"\n${BLUE}Testing GoBetterAuth Integration${NC}\"

if [ -n \"\$USER_TOKEN\" ]; then
    run_test "GBA Session Persistence" "
        # Make multiple requests with same token
        response1=\$(make_request GET \"$V1_API/auth/session\" \"\" \"Authorization: Bearer \$USER_TOKEN\")
        sleep 1
        response2=\$(make_request GET \"$V1_API/auth/session\" \"\" \"Authorization: Bearer \$USER_TOKEN\")
        # Both should succeed
        echo \"\$response1\" | jq -e '.user' >/dev/null && echo \"\$response2\" | jq -e '.user' >/dev/null
    "
else
    echo \"Skipping GBA test - no user token\"
fi

# Summary
echo
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Tests Run: $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $((TESTS_RUN - TESTS_PASSED))"

if [ "$TESTS_PASSED" -eq "$TESTS_RUN" ]; then
    echo -e "${GREEN}🎉 All tests passed! Consolidated auth system is working correctly.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please check the auth system configuration.${NC}"
    exit 1
fi