#!/bin/bash

# Test script for JWT authentication
# This script tests the authentication flow end-to-end

set -e

API_URL="${API_URL:-http://localhost:8080}"
echo "Testing API at: $API_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name="$1"
    local expected_status="$2"
    shift 2
    local response
    local status
    
    echo -n "Testing: $name... "
    response=$(curl -s -w "\n%{http_code}" "$@")
    status=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $status)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected HTTP $expected_status, got HTTP $status)"
        echo "Response: $body"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo "=== Public Endpoints ==="
echo ""

# Test 1: Health check
test_endpoint "Health check (public)" "200" \
    "$API_URL/health"

echo ""

# Test 2: Login with invalid credentials
test_endpoint "Login with invalid credentials" "401" \
    -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"wrongpassword"}'

echo ""

# Test 3: Login with valid credentials
echo -n "Testing: Login with valid credentials... "
response=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"password123"}')

TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✓ PASS${NC}"
    echo "  Token received: ${TOKEN:0:50}..."
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "  No token received"
    echo "  Response: $response"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    exit 1
fi

echo ""
echo "=== Protected Endpoints Without Auth ==="
echo ""

# Test 4: Access protected endpoint without token
test_endpoint "GET /api/items without token" "401" \
    "$API_URL/api/items"

echo ""

# Test 5: Access protected endpoint with invalid token
test_endpoint "GET /api/items with invalid token" "401" \
    -H "Authorization: Bearer invalid_token_here" \
    "$API_URL/api/items"

echo ""
echo "=== Protected Endpoints With Auth ==="
echo ""

# Test 6: Access protected endpoint with valid token
test_endpoint "GET /api/items with valid token" "200" \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/items"

echo ""

# Test 7: Create item with valid token
test_endpoint "POST /api/items with valid token" "201" \
    -X POST "$API_URL/api/items" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"Test Item","description":"Created via authenticated API"}'

echo ""

# Test 8: Get specific item with valid token
test_endpoint "GET /api/items/123 with valid token" "200" \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/items/123"

echo ""

# Test 9: Token refresh
test_endpoint "POST /auth/refresh with valid token" "200" \
    -X POST "$API_URL/auth/refresh" \
    -H "Authorization: Bearer $TOKEN"

echo ""
echo "=== Test Summary ==="
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC} 🎉"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi

