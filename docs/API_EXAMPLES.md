# Stock Analysis API - Example Responses

This document shows example API responses for the stock analysis endpoints.

**Note:** All stock analysis endpoints require JWT authentication. See the [Authentication](#authentication) section below.

## Authentication

All stock analysis endpoints require a JWT token. First, login to obtain a token:

**Login Request:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"Utd@Pogba6"}'
```

**Login Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "sshetty",
  "message": "Login successful"
}
```

Use the token in all subsequent requests:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/stocks/AAPL/fundamentals
```

---

## GET /api/stocks/{ticker}/fundamentals

Returns the "Big 5" fundamental scorecard for a stock.

**Authentication:** Required (JWT Bearer token)

**Example Request:**
```bash
TOKEN="your_jwt_token_here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/AAPL/fundamentals
```

**Example Response:**
```json
{
  "ticker": "AAPL",
  "company_name": "Apple Inc.",
  "current_price": 175.43,
  "last_updated": "2025-12-28T10:30:00Z",
  "fundamental_scorecard": {
    "pe_ratio": {
      "current": 28.5,
      "five_year_avg": 25.2,
      "rating": "YELLOW",
      "message": "P/E (28.50) is slightly above 5-year average (25.20)",
      "available": true
    },
    "debt_to_equity": {
      "current": 0.85,
      "rating": "YELLOW",
      "message": "Moderate debt levels (0.85) - acceptable",
      "available": true
    },
    "fcf_yield": {
      "current": 3.56,
      "rating": "RED",
      "message": "Low FCF yield (3.56%) - limited cash generation",
      "available": true
    },
    "peg_ratio": {
      "current": 3.56,
      "rating": "RED",
      "message": "PEG of 3.56 suggests overvalued relative to growth (assuming 8% growth)",
      "available": true
    },
    "roe": {
      "current": 155.7,
      "rating": "GREEN",
      "message": "Excellent ROE (155.70%) - highly efficient management",
      "available": true
    },
    "overall_score": "3/5 metrics healthy",
    "summary": "Mixed fundamentals - Proceed with caution"
  },
  "warnings": [],
  "data_freshness": {
    "price": "real-time",
    "fundamentals": "2024-FY"
  }
}
```

**Rating System:**
- `GREEN`: Healthy/Good value
- `YELLOW`: Moderate/Acceptable
- `RED`: Concerning/Risky
- `N/A`: Data not available

---

## GET /api/stocks/{ticker}/valuation

Returns DCF (Discounted Cash Flow) intrinsic value calculation.

**Authentication:** Required (JWT Bearer token)

**Example Request (with defaults):**
```bash
TOKEN="your_jwt_token_here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/AAPL/valuation
```

**Example Request (with custom assumptions):**
```bash
TOKEN="your_jwt_token_here"
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/stocks/AAPL/valuation?revenue_growth=0.10&profit_margin=0.25&discount_rate=0.12"
```

**Example Response:**
```json
{
  "ticker": "AAPL",
  "company_name": "Apple Inc.",
  "current_price": 175.43,
  "last_updated": "2025-12-28T10:30:00Z",
  "valuation": {
    "fair_value_per_share": 192.30,
    "current_price": 175.43,
    "upside_percent": 9.6,
    "model": "DCF",
    "assumptions": {
      "revenue_growth_rate": 0.08,
      "profit_margin": 0.246,
      "fcf_margin": 0.252,
      "discount_rate": 0.10,
      "terminal_growth_rate": 0.025,
      "projection_years": 5,
      "source": "defaults"
    },
    "projections": [
      {
        "year": 1,
        "revenue": 425874240000,
        "net_income": 104765062400,
        "free_cash_flow": 107320228800,
        "discount_factor": 1.1,
        "present_value": 97563844363.64
      },
      {
        "year": 2,
        "revenue": 459944179200,
        "net_income": 113146068123,
        "free_cash_flow": 115905932198,
        "discount_factor": 1.21,
        "present_value": 95789232725.21
      }
      // ... years 3-5 omitted for brevity
    ],
    "terminal_value": 2845671234567,
    "enterprise_value": 2954823456789,
    "shares_outstanding": 16000.0
  },
  "warnings": [],
  "data_freshness": {
    "price": "real-time",
    "fundamentals": "2024-FY"
  }
}
```

**Interpretation:**
- `upside_percent > 0`: Stock is undervalued (potential buy)
- `upside_percent < 0`: Stock is overvalued (potential sell)
- `upside_percent ≈ 0`: Stock is fairly valued

**Query Parameters:**
- `revenue_growth`: Annual revenue growth rate (0.08 = 8%)
- `profit_margin`: Net profit margin (0.15 = 15%)
- `fcf_margin`: Free cash flow margin (0.12 = 12%)
- `discount_rate`: Required rate of return (0.10 = 10%)
- `terminal_growth`: Perpetual growth rate (0.025 = 2.5%)

---

## GET /api/stocks/{ticker}/metrics

Returns comprehensive analysis combining fundamentals and valuation.

**Authentication:** Required (JWT Bearer token)

**Example Request:**
```bash
TOKEN="your_jwt_token_here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/AAPL/metrics
```

**Example Response:**
```json
{
  "ticker": "AAPL",
  "company_name": "Apple Inc.",
  "current_price": 175.43,
  "last_updated": "2025-12-28T10:30:00Z",
  "fundamental_scorecard": {
    "pe_ratio": {
      "current": 28.5,
      "rating": "YELLOW",
      "message": "P/E (28.50) is moderate",
      "available": true
    },
    "debt_to_equity": {
      "current": 0.85,
      "rating": "YELLOW",
      "message": "Moderate debt levels (0.85) - acceptable",
      "available": true
    },
    "fcf_yield": {
      "current": 3.56,
      "rating": "RED",
      "message": "Low FCF yield (3.56%) - limited cash generation",
      "available": true
    },
    "peg_ratio": {
      "current": 3.56,
      "rating": "RED",
      "message": "PEG of 3.56 suggests overvalued relative to growth (assuming 8% growth)",
      "available": true
    },
    "roe": {
      "current": 155.7,
      "rating": "GREEN",
      "message": "Excellent ROE (155.70%) - highly efficient management",
      "available": true
    },
    "overall_score": "3/5 metrics healthy",
    "summary": "Mixed fundamentals - Proceed with caution"
  },
  "valuation": {
    "fair_value_per_share": 192.30,
    "current_price": 175.43,
    "upside_percent": 9.6,
    "model": "DCF",
    "assumptions": {
      "revenue_growth_rate": 0.08,
      "profit_margin": 0.246,
      "fcf_margin": 0.252,
      "discount_rate": 0.10,
      "terminal_growth_rate": 0.025,
      "projection_years": 5,
      "source": "defaults"
    }
  },
  "warnings": [],
  "data_freshness": {
    "price": "real-time",
    "fundamentals": "2024-FY"
  }
}
```

---

## Error Responses

### Unauthorized Access (401)

**Request (missing token):**
```bash
curl http://localhost:8080/api/stocks/AAPL/metrics
```

**Response (401 Unauthorized):**
```json
{
  "error": "Unauthorized",
  "message": "Missing or invalid authorization token"
}
```

### Invalid Ticker

**Request:**
```bash
TOKEN="your_jwt_token_here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/INVALID/metrics
```

**Response (400 Bad Request):**
```json
{
  "error": "Invalid request",
  "message": "invalid ticker or no data available for INVALID"
}
```

### Missing Data (Graceful Degradation)

**Request:**
```bash
TOKEN="your_jwt_token_here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/XYZ/fundamentals
```

**Response (200 OK with warnings):**
```json
{
  "ticker": "XYZ",
  "company_name": "XYZ Corp",
  "current_price": 0,
  "last_updated": "2025-12-28T10:30:00Z",
  "fundamental_scorecard": {
    "pe_ratio": {
      "current": 0,
      "rating": "N/A",
      "message": "Insufficient data to calculate P/E ratio",
      "available": false
    },
    "debt_to_equity": {
      "current": 0,
      "rating": "N/A",
      "message": "No financial data available",
      "available": false
    },
    "fcf_yield": {
      "current": 0,
      "rating": "N/A",
      "message": "Insufficient data to calculate FCF Yield",
      "available": false
    },
    "peg_ratio": {
      "current": 0,
      "rating": "N/A",
      "message": "P/E ratio not available",
      "available": false
    },
    "roe": {
      "current": 0,
      "rating": "N/A",
      "message": "No financial data available",
      "available": false
    },
    "overall_score": "0/5 metrics available",
    "summary": "Insufficient data for analysis"
  },
  "warnings": [
    "Price data unavailable: Finnhub: invalid ticker or no data available for XYZ",
    "Company profile unavailable: Finnhub: API error (status 404)",
    "Fundamental data unavailable: EDGAR: CIK not found for ticker XYZ (limited ticker support in MVP)"
  ],
  "data_freshness": {
    "price": "unavailable",
    "fundamentals": "unavailable"
  }
}
```

---

## Mock Data Mode

When running with `USE_MOCK_DATA=true`, the API returns synthetic data for testing.

**Request:**
```bash
export USE_MOCK_DATA=true
make run-local

# In another terminal, login first
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"Utd@Pogba6"}' | jq -r '.token')

# Then test with mock data
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/TEST/metrics
```

**Response:**
Mock data will be returned with realistic values, useful for:
- Local development without API keys
- CI/CD testing
- Demo purposes

---

## Common Use Cases

### 1. Complete Workflow
```bash
# Step 1: Login
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"Utd@Pogba6"}' | jq -r '.token')

# Step 2: Check if fundamentals are strong
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/MSFT/fundamentals | \
  jq '.fundamental_scorecard.summary'
# Output: "Strong fundamentals - Good investment candidate"

# Step 3: Check valuation upside
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/GOOGL/valuation | \
  jq '.valuation.upside_percent'
# Output: 12.5 (indicating 12.5% upside)
```

### 2. Quick Stock Health Check
```bash
TOKEN="your_token"
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/MSFT/fundamentals | \
  jq '.fundamental_scorecard.summary'
```

### 3. Find Undervalued Stocks
```bash
TOKEN="your_token"
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/stocks/GOOGL/valuation | \
  jq '.valuation.upside_percent'
```

### 4. Custom DCF Analysis
```bash
TOKEN="your_token"
# Run aggressive growth scenario
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/stocks/NVDA/valuation?revenue_growth=0.25&profit_margin=0.30&discount_rate=0.15"
```

### 5. Batch Analysis (using jq)
```bash
# Login once
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"Utd@Pogba6"}' | jq -r '.token')

# Analyze multiple stocks
for ticker in AAPL MSFT GOOGL AMZN; do
  echo "$ticker:"
  curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8080/api/stocks/$ticker/fundamentals" | \
    jq -r '.fundamental_scorecard.summary'
  echo ""
done
```

---

## Data Sources

The API aggregates data from:
- **Finnhub**: Real-time prices, market cap, shares outstanding
- **SEC EDGAR**: Financial statements (revenue, net income, cash flow, balance sheet)
- **OpenFIGI**: Ticker validation (optional)

For detailed information about data sources, see [DATA_SOURCES.md](DATA_SOURCES.md).

---

## Authentication Details

For complete authentication documentation, including:
- User management
- Password hashing
- Token refresh
- Security best practices

See [AUTHENTICATION.md](AUTHENTICATION.md)

---

## Rate Limits

- **Local Development**: No limits
- **Production**: Depends on external API rate limits
  - Finnhub free tier: 60 calls/minute
  - SEC EDGAR: 10 requests/second
  - OpenFIGI: 25 requests/minute (free)

Recommended: Implement caching for production use.

