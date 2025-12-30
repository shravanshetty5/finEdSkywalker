#!/bin/bash

# Test script for ticker search endpoint (with authentication)
# Usage: ./scripts/test-search.sh [api_url]
# 
# Environment variables:
#   TEST_USERNAME - Username for authentication (default: sshetty)
#   TEST_PASSWORD - Password for authentication (required, or set USER_SSHETTY_PASSWORD)

set -e

API_URL=${1:-http://localhost:8080}
USERNAME="${TEST_USERNAME:-sshetty}"
PASSWORD="${TEST_PASSWORD:-${USER_SSHETTY_PASSWORD}}"

echo "================================"
echo "Ticker Search API Test"
echo "================================"
echo "API URL: $API_URL"
echo "Username: $USERNAME"
echo ""

# Color output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}1. Testing Health Endpoint${NC}"
echo "GET $API_URL/health"
curl -s "$API_URL/health" | jq '.'
echo ""
echo ""

echo -e "${BLUE}2. Logging in to get JWT token${NC}"
echo "POST $API_URL/auth/login"

if [ -z "$PASSWORD" ]; then
  echo -e "${RED}Error: No password provided${NC}"
  echo "Set TEST_PASSWORD or USER_SSHETTY_PASSWORD environment variable"
  echo "Example: TEST_PASSWORD=your_password ./scripts/test-search.sh"
  exit 1
fi

LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}Error: Failed to get JWT token${NC}"
  echo "Response: $LOGIN_RESPONSE"
  echo ""
  echo -e "${YELLOW}Make sure the user exists and password is correct${NC}"
  echo "Set environment variables: TEST_USERNAME and TEST_PASSWORD"
  exit 1
fi

echo -e "${GREEN}Successfully logged in!${NC}"
echo "Token: ${TOKEN:0:20}..."
echo ""
echo ""

echo -e "${BLUE}3. Testing Search - Single Letter${NC}"
echo "GET $API_URL/api/search/tickers?q=A"
curl -s "$API_URL/api/search/tickers?q=A" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}4. Testing Search - Common Ticker${NC}"
echo "GET $API_URL/api/search/tickers?q=AAPL"
curl -s "$API_URL/api/search/tickers?q=AAPL" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}5. Testing Search - Partial Match${NC}"
echo "GET $API_URL/api/search/tickers?q=App"
curl -s "$API_URL/api/search/tickers?q=App" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}6. Testing Search - Company Name${NC}"
echo "GET $API_URL/api/search/tickers?q=Microsoft"
curl -s "$API_URL/api/search/tickers?q=Microsoft" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}7. Testing Search - With Custom Limit${NC}"
echo "GET $API_URL/api/search/tickers?q=tech&limit=5"
curl -s "$API_URL/api/search/tickers?q=tech&limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}8. Testing Search - Case Insensitive${NC}"
echo "GET $API_URL/api/search/tickers?q=tesla"
curl -s "$API_URL/api/search/tickers?q=tesla" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}9. Testing Search - Multiple Results${NC}"
echo "GET $API_URL/api/search/tickers?q=bank"
curl -s "$API_URL/api/search/tickers?q=bank" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}10. Testing Edge Cases${NC}"

echo -e "${YELLOW}  a) Empty query (should fail)${NC}"
echo "GET $API_URL/api/search/tickers?q="
curl -s "$API_URL/api/search/tickers?q=" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

echo -e "${YELLOW}  b) Missing query parameter (should fail)${NC}"
echo "GET $API_URL/api/search/tickers"
curl -s "$API_URL/api/search/tickers" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

echo -e "${YELLOW}  c) Very long query (should be rejected)${NC}"
echo "GET $API_URL/api/search/tickers?q=<101 characters>"
LONG_QUERY=$(printf 'a%.0s' {1..101})
curl -s "$API_URL/api/search/tickers?q=$LONG_QUERY" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

echo -e "${YELLOW}  d) Limit exceeding max (should cap at 50)${NC}"
echo "GET $API_URL/api/search/tickers?q=A&limit=100"
curl -s "$API_URL/api/search/tickers?q=A&limit=100" \
  -H "Authorization: Bearer $TOKEN" | jq '. | {query, total, results_count: (.results | length)}'
echo ""
echo ""

echo -e "${BLUE}11. Testing without authentication (should fail)${NC}"
echo "GET $API_URL/api/search/tickers?q=AAPL (no token)"
curl -s "$API_URL/api/search/tickers?q=AAPL" | jq '.'
echo ""
echo ""

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo -e "${YELLOW}Usage examples:${NC}"
echo "  ./scripts/test-search.sh                                      # Test local (http://localhost:8080)"
echo "  ./scripts/test-search.sh https://your-api-url.com             # Test specific URL"
echo "  SKYWALKER_API_URL=https://your-api-url.com make test-search  # Via environment variable"
echo ""
echo -e "${YELLOW}Quick search test:${NC}"
echo "  curl -H \"Authorization: Bearer \$TOKEN\" \"$API_URL/api/search/tickers?q=YOUR_QUERY\" | jq"




