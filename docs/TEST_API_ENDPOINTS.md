# Direct API Testing - Finnhub & SEC EDGAR

## Testing Instructions

Run these curl commands and share the responses so we can validate the data models.

---

## 1. Finnhub API Endpoints

### Prerequisites
```bash
# Set your Finnhub API key
export FINNHUB_API_KEY="your_api_key_here"
```

### A. Get Stock Quote (Real-time Price)
```bash
curl -s "https://finnhub.io/api/v1/quote?symbol=AAPL&token=$FINNHUB_API_KEY" | jq
```

**Expected Response Structure:**
```json
{
  "c": 175.43,    // Current price
  "d": 2.15,      // Change
  "dp": 1.24,     // Percent change
  "h": 176.50,    // High
  "l": 173.20,    // Low
  "o": 174.00,    // Open
  "pc": 173.28,   // Previous close
  "t": 1703780400 // Timestamp
}
```

### B. Get Company Profile
```bash
curl -s "https://finnhub.io/api/v1/stock/profile2?symbol=AAPL&token=$FINNHUB_API_KEY" | jq
```

**Expected Response Structure:**
```json
{
  "country": "US",
  "currency": "USD",
  "exchange": "NASDAQ",
  "name": "Apple Inc.",
  "ticker": "AAPL",
  "marketCapitalization": 2800000,  // In millions
  "shareOutstanding": 16000,        // In millions
  "logo": "https://...",
  "weburl": "https://www.apple.com",
  "finnhubIndustry": "Technology"
}
```

### C. Get Basic Metrics
```bash
curl -s "https://finnhub.io/api/v1/stock/metric?symbol=AAPL&metric=all&token=$FINNHUB_API_KEY" | jq
```

**Expected Response Structure:**
```json
{
  "metric": {
    "52WeekHigh": 199.62,
    "52WeekLow": 164.08,
    "beta": 1.29,
    "peBasicExclExtraTTM": 28.5,
    "roeTTM": 1.47,
    "roaTTM": 0.22,
    "totalDebt/totalEquityAnnual": 1.96,
    // ... many more metrics
  },
  "series": {
    // Historical data
  }
}
```

---

## 2. SEC EDGAR API Endpoints

### A. Get Company Facts (Main Endpoint)
```bash
# For Apple Inc. (CIK: 0000320193)
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | jq '.' | head -100
```

**To see the structure only (without data):**
```bash
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | jq 'keys'
```

**To see CIK type:**
```bash
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | jq '.cik, .entityName'
```

**To see a specific fact structure (Revenues):**
```bash
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | \
  jq '.facts."us-gaap".Revenues.units.USD[0]'
```

**Expected Response Structure:**
```json
{
  "cik": 320193,  // NOTE: Can be number OR string "0000320193"
  "entityName": "Apple Inc.",
  "facts": {
    "us-gaap": {
      "Revenues": {
        "label": "Revenue",
        "description": "...",
        "units": {
          "USD": [
            {
              "end": "2024-09-30",
              "val": 394328000000,  // NOTE: Can be number OR string
              "accn": "0000320193-24-000123",
              "fy": 2024,
              "fp": "FY",
              "form": "10-K",
              "filed": "2024-11-01"
            }
          ]
        }
      },
      "NetIncomeLoss": { /* similar structure */ },
      "Assets": { /* similar structure */ }
    }
  }
}
```

### B. Other EDGAR Tickers to Test
```bash
# Microsoft (CIK: 0000789019)
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000789019.json" | jq '.cik, .entityName'

# Google/Alphabet (CIK: 0001652044)
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0001652044.json" | jq '.cik, .entityName'

# Tesla (CIK: 0001318605)
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0001318605.json" | jq '.cik, .entityName'
```

---

## 3. Test All at Once

### Quick Test Script
```bash
#!/bin/bash

echo "=== Testing Finnhub API ==="
echo ""
echo "1. Quote for AAPL:"
curl -s "https://finnhub.io/api/v1/quote?symbol=AAPL&token=$FINNHUB_API_KEY" | jq
echo ""

echo "2. Profile for AAPL:"
curl -s "https://finnhub.io/api/v1/stock/profile2?symbol=AAPL&token=$FINNHUB_API_KEY" | jq
echo ""

echo "=== Testing SEC EDGAR API ==="
echo ""
echo "3. CIK and Entity Name for Apple:"
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | \
  jq '{cik: .cik, entityName: .entityName, cik_type: (.cik | type)}'
echo ""

echo "4. Sample Revenue Data:"
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | \
  jq '.facts."us-gaap".Revenues.units.USD[0] | {end, val, val_type: (.val | type), fy, fp, form}'
```

Save this as `test-apis.sh` and run:
```bash
chmod +x test-apis.sh
./test-apis.sh
```

---

## 4. What to Share

Please run these specific commands and share the output:

### Critical Tests:

**Test 1 - Check CIK type:**
```bash
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | \
  jq '{cik: .cik, cik_type: (.cik | type)}'
```

**Test 2 - Check Val type:**
```bash
curl -s \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | \
  jq '.facts."us-gaap".Revenues.units.USD[0] | {val: .val, val_type: (.val | type)}'
```

**Test 3 - Finnhub Quote:**
```bash
curl -s "https://finnhub.io/api/v1/quote?symbol=AAPL&token=$FINNHUB_API_KEY" | jq
```

**Test 4 - Finnhub Profile:**
```bash
curl -s "https://finnhub.io/api/v1/stock/profile2?symbol=AAPL&token=$FINNHUB_API_KEY" | jq
```

---

## 5. Known Type Issues

Based on the error you saw, here's what we know:

### CIK Field
- **Can be**: `number` (e.g., `320193`) OR `string` (e.g., `"0000320193"`)
- **Our fix**: Using `json.Number` which handles both

### Val Field  
- **Can be**: `number` (e.g., `394328000000`) OR `string` (e.g., `"394328000000"`)
- **Our fix**: Using `json.Number` which handles both

---

## 6. Rate Limits

**Finnhub:**
- Free tier: 60 calls/minute
- Wait a few seconds between calls if testing many times

**SEC EDGAR:**
- Limit: 10 requests/second per IP
- Must include User-Agent header (enforced)
- Be respectful of their servers

---

## 7. Troubleshooting

### If Finnhub returns empty:
```bash
# Check if your API key is set
echo $FINNHUB_API_KEY

# Try a different ticker
curl -s "https://finnhub.io/api/v1/quote?symbol=MSFT&token=$FINNHUB_API_KEY" | jq
```

### If EDGAR returns error:
```bash
# Make sure User-Agent is included
curl -v \
  -H "User-Agent: finEdSkywalker/1.0 (test@example.com)" \
  "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json" | head
```

---

## Next Steps

1. Run the critical tests (Test 1-4 above)
2. Share the output here
3. I'll verify our Go structs match the actual API responses
4. We'll fix any mismatches if needed

This will ensure our data models are 100% correct! 🎯

