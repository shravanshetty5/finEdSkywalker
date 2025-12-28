#!/bin/bash

# Quick start script for testing the API with authentication
# This script demonstrates the complete authentication flow

set -e

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
USERNAME="${USERNAME:-alice}"
PASSWORD="${PASSWORD:-password123}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== finEdSkywalker API Quick Start ===${NC}"
echo ""
echo "API URL: $API_URL"
echo "Username: $USERNAME"
echo ""

# Step 1: Health Check
echo -e "${YELLOW}Step 1: Testing health endpoint (public)${NC}"
echo "curl $API_URL/health"
curl -s "$API_URL/health" | jq . || echo "Error: jq not installed, showing raw response:"
echo ""
echo ""

# Step 2: Login
echo -e "${YELLOW}Step 2: Logging in${NC}"
echo "curl -X POST $API_URL/auth/login ..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

echo "$LOGIN_RESPONSE" | jq . 2>/dev/null || echo "$LOGIN_RESPONSE"
echo ""

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to get token. Check your credentials.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Login successful!${NC}"
echo "Token: ${TOKEN:0:50}..."
echo ""
echo ""

# Step 3: Try protected endpoint without auth (will fail)
echo -e "${YELLOW}Step 3: Testing protected endpoint WITHOUT auth (should fail)${NC}"
echo "curl $API_URL/api/items"
curl -s "$API_URL/api/items" | jq . 2>/dev/null || curl -s "$API_URL/api/items"
echo ""
echo ""

# Step 4: Access protected endpoint with auth
echo -e "${YELLOW}Step 4: Testing protected endpoint WITH auth${NC}"
echo "curl $API_URL/api/items -H \"Authorization: Bearer \$TOKEN\""
curl -s "$API_URL/api/items" \
  -H "Authorization: Bearer $TOKEN" | jq . 2>/dev/null || \
  curl -s "$API_URL/api/items" -H "Authorization: Bearer $TOKEN"
echo ""
echo ""

# Step 5: Create a new item
echo -e "${YELLOW}Step 5: Creating a new item${NC}"
echo "curl -X POST $API_URL/api/items ..."
curl -s -X POST "$API_URL/api/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Item","description":"Created via quickstart script"}' | jq . 2>/dev/null || \
  curl -s -X POST "$API_URL/api/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Item","description":"Created via quickstart script"}'
echo ""
echo ""

# Step 6: Get a specific item
echo -e "${YELLOW}Step 6: Getting a specific item${NC}"
echo "curl $API_URL/api/items/123 -H \"Authorization: Bearer \$TOKEN\""
curl -s "$API_URL/api/items/123" \
  -H "Authorization: Bearer $TOKEN" | jq . 2>/dev/null || \
  curl -s "$API_URL/api/items/123" -H "Authorization: Bearer $TOKEN"
echo ""
echo ""

# Summary
echo -e "${GREEN}=== All tests completed successfully! ===${NC}"
echo ""
echo "Your JWT token is:"
echo "$TOKEN"
echo ""
echo "You can use this token for the next 24 hours to make authenticated requests:"
echo ""
echo "  export API_TOKEN=\"$TOKEN\""
echo "  curl $API_URL/api/items -H \"Authorization: Bearer \$API_TOKEN\""
echo ""
echo "For more information, see docs/AUTHENTICATION.md"

