#!/bin/bash

# Test AWS Stock Analysis API Endpoints
# This script tests the deployed Lambda function via API Gateway

set -e

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get API Gateway URL (priority order: argument > env var > terraform)
API_URL=""

# 1. Check if provided as first argument
if [ -n "$1" ]; then
  API_URL="$1"
fi

# 2. Check if provided as environment variable
if [ -z "$API_URL" ] && [ -n "$SKYWALKER_API_URL" ]; then
  API_URL="$SKYWALKER_API_URL"
fi

# Error if no URL found
if [ -z "$API_URL" ]; then
  echo -e "${RED}Error: No API Gateway URL provided${NC}"
  echo ""
  echo -e "${YELLOW}Please provide the API URL in one of these ways:${NC}"
  echo ""
  echo "1. As a command argument:"
  echo "   $0 https://abc123.execute-api.us-east-1.amazonaws.com"
  echo ""
  echo "2. As an environment variable:"
  echo "   export API_GATEWAY_URL=https://abc123.execute-api.us-east-1.amazonaws.com"
  echo "   $0"
  echo ""
  echo "3. Deploy infrastructure so Terraform has the URL:"
  echo "   make deploy"
  echo ""
  echo "To test locally instead:"
  echo "   make test-stocks"
  exit 1
fi

# Default credentials (update these if different)
USERNAME="${AWS_USERNAME:-sshetty}"
PASSWORD="${AWS_PASSWORD:-Utd@Pogba6}"
TICKER="${TICKER:-AAPL}"

echo "================================"
echo "AWS Stock Analysis API Test"
echo "================================"
echo "API URL: $API_URL"
echo "Ticker: $TICKER"
echo ""

# Test 1: Health Check
echo -e "${BLUE}1. Testing Health Endpoint${NC}"
echo "GET $API_URL/health"
echo ""
curl -s "$API_URL/health" | jq '.'
echo ""
echo ""

# Test 2: Login
echo -e "${BLUE}2. Logging in to get JWT token${NC}"
echo "POST $API_URL/auth/login"
echo ""
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}Error: Failed to get JWT token${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}Successfully logged in!${NC}"
echo "Token: ${TOKEN:0:30}..."
echo ""
echo ""

# Test 3: Stock Fundamentals
echo -e "${BLUE}3. Testing Stock Fundamentals Endpoint${NC}"
echo "GET $API_URL/api/stocks/$TICKER/fundamentals"
echo ""
curl -s "$API_URL/api/stocks/$TICKER/fundamentals" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

# Test 4: Stock Valuation (Default)
echo -e "${BLUE}4. Testing Stock Valuation (Default Assumptions)${NC}"
echo "GET $API_URL/api/stocks/$TICKER/valuation"
echo ""
curl -s "$API_URL/api/stocks/$TICKER/valuation" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

# Test 5: Stock Valuation (Custom)
echo -e "${BLUE}5. Testing Stock Valuation (Custom Assumptions)${NC}"
echo "GET $API_URL/api/stocks/$TICKER/valuation?revenue_growth=0.10&discount_rate=0.12"
echo ""
curl -s "$API_URL/api/stocks/$TICKER/valuation?revenue_growth=0.10&discount_rate=0.12" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

# Test 6: Comprehensive Metrics
echo -e "${BLUE}6. Testing Comprehensive Metrics Endpoint${NC}"
echo "GET $API_URL/api/stocks/$TICKER/metrics"
echo ""
curl -s "$API_URL/api/stocks/$TICKER/metrics" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo ""

# Test 7: Unauthorized Access
echo -e "${BLUE}7. Testing Unauthorized Access (Should Fail)${NC}"
echo "GET $API_URL/api/stocks/$TICKER/fundamentals (no token)"
echo ""
curl -s "$API_URL/api/stocks/$TICKER/fundamentals" | jq '.'
echo ""
echo ""

# Test 8: Multiple Tickers
echo -e "${BLUE}8. Testing Multiple Tickers${NC}"
for STOCK in MSFT GOOGL TSLA; do
  echo "Fetching fundamentals for $STOCK..."
  SUMMARY=$(curl -s "$API_URL/api/stocks/$STOCK/fundamentals" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.fundamental_scorecard.summary // "N/A"')
  echo "  $STOCK: $SUMMARY"
done
echo ""
echo ""

# Summary
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo -e "${YELLOW}API Gateway URL:${NC} $API_URL"
echo -e "${YELLOW}Token valid for:${NC} ~1 hour"
echo ""

