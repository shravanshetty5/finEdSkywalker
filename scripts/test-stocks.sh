#!/bin/bash

# Test script for stock analysis endpoints (with authentication)
# Usage: ./scripts/test-stocks.sh [ticker] [api_url]
#
# Environment variables:
#   TEST_USERNAME - Username for authentication (default: sshetty)
#   TEST_PASSWORD - Password for authentication (required, or set USER_SSHETTY_PASSWORD)

set -e

TICKER=${1:-AAPL}
API_URL=${2:-http://localhost:8080}
USERNAME="${TEST_USERNAME:-sshetty}"
PASSWORD="${TEST_PASSWORD:-${USER_SSHETTY_PASSWORD}}"

echo "================================"
echo "Stock Analysis API Test"
echo "================================"
echo "Ticker: $TICKER"
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
  echo "Example: TEST_PASSWORD=your_password ./scripts/test-stocks.sh"
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

echo -e "${BLUE}3. Testing Fundamentals Endpoint${NC}"
echo "GET $API_URL/api/stocks/$TICKER/fundamentals"
curl -s "$API_URL/api/stocks/$TICKER/fundamentals" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}4. Testing Valuation Endpoint (Default Assumptions)${NC}"
echo "GET $API_URL/api/stocks/$TICKER/valuation"
curl -s "$API_URL/api/stocks/$TICKER/valuation" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}5. Testing Valuation with Custom Assumptions${NC}"
echo "GET $API_URL/api/stocks/$TICKER/valuation?revenue_growth=0.10&discount_rate=0.12"
curl -s "$API_URL/api/stocks/$TICKER/valuation?revenue_growth=0.10&discount_rate=0.12" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}6. Testing Comprehensive Metrics Endpoint${NC}"
echo "GET $API_URL/api/stocks/$TICKER/metrics"
curl -s "$API_URL/api/stocks/$TICKER/metrics" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

echo -e "${BLUE}7. Testing without authentication (should fail)${NC}"
echo "GET $API_URL/api/stocks/$TICKER/fundamentals (no token)"
curl -s "$API_URL/api/stocks/$TICKER/fundamentals" | jq '.'
echo ""
echo ""

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo -e "${YELLOW}Try other tickers:${NC}"
echo "  ./scripts/test-stocks.sh MSFT"
echo "  ./scripts/test-stocks.sh GOOGL"
echo "  ./scripts/test-stocks.sh TSLA"
echo ""
echo -e "${YELLOW}Test against deployed API:${NC}"
echo "  ./scripts/test-stocks.sh AAPL https://your-api.execute-api.us-east-1.amazonaws.com"


