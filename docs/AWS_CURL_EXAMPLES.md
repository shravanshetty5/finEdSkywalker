# Testing AWS Stock Analysis API - Curl Examples

## Quick Start

### 1. Get Your API Gateway URL

```bash
cd terraform
terraform output api_gateway_url
```

Example output: `https://abc123xyz.execute-api.us-east-1.amazonaws.com`

### 2. Run the Automated Test Script

```bash
./scripts/test-aws-stocks.sh
```

Or with a custom URL:
```bash
./scripts/test-aws-stocks.sh https://your-api.execute-api.us-east-1.amazonaws.com
```

Or test different ticker:
```bash
TICKER=MSFT ./scripts/test-aws-stocks.sh
```

---

## Manual Curl Examples

### Setup

```bash
# Set your API URL
export API_URL="https://abc123xyz.execute-api.us-east-1.amazonaws.com"

# Or get it from Terraform
export API_URL=$(cd terraform && terraform output -raw api_gateway_url)
```

---

## 1. Health Check (Public)

```bash
curl -s "$API_URL/health" | jq
```

**Expected Response:**
```json
{
  "data": {
    "service": "finEdSkywalker",
    "status": "ok",
    "version": "1.0.0"
  },
  "message": "Service is healthy"
}
```

---

## 2. Login to Get JWT Token

```bash
curl -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' | jq
```

**Expected Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "sshetty",
  "message": "Login successful"
}
```

**Save the token:**
```bash
export TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' | jq -r '.token')

echo "Token: $TOKEN"
```

---

## 3. Get Stock Fundamentals

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/fundamentals" | jq
```

**With different ticker:**
```bash
# Microsoft
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/MSFT/fundamentals" | jq

# Google
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/GOOGL/fundamentals" | jq

# Tesla
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/TSLA/fundamentals" | jq
```

**Extract specific metrics:**
```bash
# Get just the overall score
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/fundamentals" | \
  jq '.fundamental_scorecard | {overall_score, summary}'

# Get just P/E ratio
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/fundamentals" | \
  jq '.fundamental_scorecard.pe_ratio'
```

---

## 4. Get DCF Valuation

### Default Assumptions

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/valuation" | jq
```

### Custom Assumptions

```bash
# Conservative scenario (low growth, high discount rate)
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/valuation?revenue_growth=0.05&discount_rate=0.15" | jq

# Aggressive scenario (high growth, low discount rate)
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/TSLA/valuation?revenue_growth=0.25&profit_margin=0.15&discount_rate=0.10" | jq

# Custom all parameters
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/valuation?revenue_growth=0.08&profit_margin=0.25&fcf_margin=0.20&discount_rate=0.12&terminal_growth=0.03" | jq
```

**Available Parameters:**
- `revenue_growth` - Annual revenue growth rate (0.08 = 8%)
- `profit_margin` - Net profit margin (0.15 = 15%)
- `fcf_margin` - Free cash flow margin (0.12 = 12%)
- `discount_rate` - Required rate of return (0.10 = 10%)
- `terminal_growth` - Perpetual growth rate (0.025 = 2.5%)

**Extract valuation summary:**
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/valuation" | \
  jq '{ticker, current_price, fair_value: .valuation.fair_value_per_share, upside: .valuation.upside_percent}'
```

---

## 5. Get Comprehensive Metrics

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/metrics" | jq
```

**With custom DCF parameters:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/metrics?revenue_growth=0.10&discount_rate=0.12" | jq
```

---

## 6. Batch Analysis

### Compare Multiple Stocks

```bash
#!/bin/bash
export TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' | jq -r '.token')

echo "Stock Analysis Summary"
echo "====================="
echo ""

for TICKER in AAPL MSFT GOOGL AMZN TSLA; do
  echo "Analyzing $TICKER..."
  
  # Get fundamentals summary
  SUMMARY=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/stocks/$TICKER/fundamentals" | \
    jq -r '.fundamental_scorecard.summary')
  
  # Get valuation
  UPSIDE=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/stocks/$TICKER/valuation" | \
    jq -r '.valuation.upside_percent // "N/A"')
  
  echo "  Fundamentals: $SUMMARY"
  echo "  Upside: $UPSIDE%"
  echo ""
done
```

### Create Comparison Table

```bash
#!/bin/bash
export TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' | jq -r '.token')

printf "%-8s %-12s %-10s %-10s %-10s\n" "Ticker" "Price" "P/E" "ROE" "Upside"
printf "%-8s %-12s %-10s %-10s %-10s\n" "------" "-----" "---" "---" "------"

for TICKER in AAPL MSFT GOOGL AMZN; do
  DATA=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/stocks/$TICKER/metrics")
  
  PRICE=$(echo $DATA | jq -r '.current_price')
  PE=$(echo $DATA | jq -r '.fundamental_scorecard.pe_ratio.current')
  ROE=$(echo $DATA | jq -r '.fundamental_scorecard.roe.current')
  UPSIDE=$(echo $DATA | jq -r '.valuation.upside_percent // "N/A"')
  
  printf "%-8s \$%-11.2f %-10.2f %-10.2f %-10s\n" "$TICKER" "$PRICE" "$PE" "$ROE" "$UPSIDE"
done
```

---

## 7. Error Testing

### Test without token (should fail)

```bash
curl "$API_URL/api/stocks/AAPL/fundamentals" | jq
```

**Expected Response:**
```json
{
  "error": "Unauthorized",
  "message": "Missing authorization header"
}
```

### Test with invalid token

```bash
curl -H "Authorization: Bearer invalid_token_here" \
  "$API_URL/api/stocks/AAPL/fundamentals" | jq
```

### Test with invalid ticker

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/INVALID/fundamentals" | jq
```

---

## 8. Performance Testing

### Measure Response Time

```bash
time curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/metrics" > /dev/null
```

### Detailed Timing

```bash
curl -w "\nTime total: %{time_total}s\nTime connect: %{time_connect}s\nTime start transfer: %{time_starttransfer}s\n" \
  -H "Authorization: Bearer $TOKEN" \
  -s -o /dev/null \
  "$API_URL/api/stocks/AAPL/fundamentals"
```

---

## 9. One-Liners

### Quick Stock Check

```bash
# Set once
API_URL="https://your-api.execute-api.us-east-1.amazonaws.com"
TOKEN=$(curl -s -X POST "$API_URL/auth/login" -H "Content-Type: application/json" -d '{"username":"your_username","password":"your_password"}' | jq -r '.token')

# Then use for any stock
curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/api/stocks/AAPL/metrics" | jq '{ticker, price: .current_price, summary: .fundamental_scorecard.summary, upside: .valuation.upside_percent}'
```

### Investment Decision Helper

```bash
# Get key metrics for decision making
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/metrics" | \
  jq '{
    ticker,
    current_price,
    fundamentals_score: .fundamental_scorecard.overall_score,
    fundamentals_summary: .fundamental_scorecard.summary,
    fair_value: .valuation.fair_value_per_share,
    upside_percent: .valuation.upside_percent,
    pe_ratio: .fundamental_scorecard.pe_ratio.current,
    debt_to_equity: .fundamental_scorecard.debt_to_equity.current,
    roe: .fundamental_scorecard.roe.current
  }'
```

---

## 10. Save Results

### Save to File

```bash
# Save full analysis
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/api/stocks/AAPL/metrics" > aapl_analysis_$(date +%Y%m%d).json

# Save multiple stocks
for TICKER in AAPL MSFT GOOGL; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/stocks/$TICKER/metrics" > "${TICKER}_analysis_$(date +%Y%m%d).json"
done
```

### Create CSV Report

```bash
#!/bin/bash
export TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' | jq -r '.token')

# Header
echo "Ticker,Price,PE,Debt/Equity,FCF_Yield,ROE,Fair_Value,Upside%" > stock_report.csv

# Data
for TICKER in AAPL MSFT GOOGL AMZN TSLA META NVDA; do
  DATA=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$API_URL/api/stocks/$TICKER/metrics")
  
  echo "$TICKER,$(echo $DATA | jq -r '[.current_price, .fundamental_scorecard.pe_ratio.current, .fundamental_scorecard.debt_to_equity.current, .fundamental_scorecard.fcf_yield.current, .fundamental_scorecard.roe.current, .valuation.fair_value_per_share, .valuation.upside_percent] | @csv')" >> stock_report.csv
done

echo "Report saved to stock_report.csv"
cat stock_report.csv
```

---

## Troubleshooting

### Get API URL if lost

```bash
cd terraform
terraform output api_gateway_url
```

### Refresh expired token

```bash
# Tokens expire after 1 hour, get a new one
export TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' | jq -r '.token')
```

### Check CloudWatch Logs

```bash
# From Makefile
make logs

# Or directly
aws logs tail /aws/lambda/finEdSkywalker-api --follow
```

### Verify deployment

```bash
cd terraform
terraform output
```

---

## Rate Limits

Be mindful of:
- **Finnhub**: 60 calls/minute (free tier)
- **SEC EDGAR**: 10 requests/second
- **API Gateway**: Throttled based on your Terraform settings

Add delays between batch requests:
```bash
for TICKER in AAPL MSFT GOOGL; do
  curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/api/stocks/$TICKER/metrics" | jq
  sleep 2  # Wait 2 seconds between requests
done
```

---

## Next Steps

1. ✅ Test health endpoint
2. ✅ Get JWT token
3. ✅ Test single stock analysis
4. ✅ Test multiple tickers
5. ✅ Try custom DCF parameters
6. ✅ Create comparison reports

For automated testing, use:
```bash
./scripts/test-aws-stocks.sh
```

