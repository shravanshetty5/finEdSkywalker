# Data Sources Documentation

This document explains the external data sources used by finEdSkywalker and how they are integrated.

## Overview

The stock analysis platform aggregates data from three primary sources:

```
┌─────────────┐
│   Finnhub   │ ──▶ Real-time prices, company profile
└─────────────┘

┌─────────────┐
│ SEC EDGAR   │ ──▶ Financial statements (XBRL data)
└─────────────┘

┌─────────────┐
│  OpenFIGI   │ ──▶ Ticker mapping (optional)
└─────────────┘
```

## Data Source Details

### 1. Finnhub

**Purpose**: Real-time stock prices, company profiles, and basic metrics

**Base URL**: `https://finnhub.io/api/v1`

**API Key Required**: Yes (free tier available)

**Endpoints Used**:
- `/quote` - Real-time price data
- `/stock/profile2` - Company profile (name, market cap, shares outstanding)
- `/stock/metric` - Financial metrics (P/E, ROE, etc.)

**Data Retrieved**:
- Current stock price
- Daily high/low/open/close
- Volume
- Market capitalization
- Shares outstanding
- Company name and basic info

**Rate Limits**: 
- Free tier: 60 API calls/minute
- Premium tiers available for higher limits

**Registration**: [https://finnhub.io/register](https://finnhub.io/register)

**Sample Response** (`/quote`):
```json
{
  "c": 175.43,
  "d": 2.15,
  "dp": 1.24,
  "h": 176.50,
  "l": 173.20,
  "o": 174.00,
  "pc": 173.28,
  "t": 1703780400
}
```

### 2. SEC EDGAR (Electronic Data Gathering, Analysis, and Retrieval)

**Purpose**: Official financial statements filed with the U.S. Securities and Exchange Commission

**Base URL**: `https://data.sec.gov`

**API Key Required**: No, but User-Agent header is **required**

**Endpoints Used**:
- `/api/xbrl/companyfacts/CIK{cik}.json` - Company financial facts in XBRL format

**Data Retrieved**:
- Revenue (annual and quarterly)
- Net Income
- Total Assets
- Total Liabilities
- Shareholders' Equity
- Total Debt
- Operating Cash Flow
- Capital Expenditures (CapEx)
- Free Cash Flow (calculated)

**Important Requirements**:
- **User-Agent Header**: SEC requires a User-Agent header with contact information
- Example: `finEdSkywalker/1.0 (your-email@example.com)`
- Requests without User-Agent will be blocked

**Rate Limits**:
- 10 requests per second per IP address
- Exceeding limits may result in IP blocking

**CIK Lookup**:
- Companies are identified by CIK (Central Index Key)
- **Automatic Ticker-to-CIK Mapping**: The system now dynamically looks up CIK numbers for any US public company ticker
- Fast path: Common tickers (AAPL, MSFT, GOOGL, etc.) are cached in memory
- Fallback: Unknown tickers are automatically looked up from SEC's official company tickers JSON file
- Source: `https://www.sec.gov/files/company_tickers.json` (updated daily by SEC)
- **Cache TTL**: Ticker map is automatically refreshed every 24 hours per Lambda instance
- **Graceful Degradation**: If refresh fails, stale cache is used to maintain availability
- Example: AAPL → CIK 0000320193
- **No manual CIK configuration needed** - just provide any valid ticker symbol

**Caching Strategy**:
- Two-level cache for optimal performance:
  - **Individual Ticker Cache**: Previously looked-up tickers are cached indefinitely
  - **Ticker Map Cache**: Full SEC ticker database (~12,000+ companies) cached with 24-hour TTL
- On each request, the ticker map is checked for freshness before any lookups
- New IPOs and ticker changes become available within 24 hours of SEC update
- Lambda cold starts always get fresh data
- Warm Lambda instances refresh automatically after 24 hours

**Data Taxonomy**:
- Uses US-GAAP (Generally Accepted Accounting Principles) taxonomy
- Different companies may use slightly different field names
- Our implementation handles common variations

**Sample CIK Mapping**:
```go
AAPL  → 0000320193  (Apple Inc.)
MSFT  → 0000789019  (Microsoft Corporation)
GOOGL → 0001652044  (Alphabet Inc.)
AMZN  → 0001018724  (Amazon.com Inc.)
TSLA  → 0001318605  (Tesla Inc.)
```

**Sample Response Structure**:
```json
{
  "cik": "0000320193",
  "entityName": "Apple Inc.",
  "facts": {
    "us-gaap": {
      "Revenues": {
        "label": "Revenue",
        "units": {
          "USD": [
            {
              "end": "2024-09-30",
              "val": 394328000000,
              "fy": 2024,
              "fp": "FY",
              "form": "10-K"
            }
          ]
        }
      }
    }
  }
}
```

### 3. OpenFIGI

**Purpose**: Financial Instrument Global Identifier mapping

**Base URL**: `https://api.openfigi.com/v3`

**API Key Required**: No (optional for higher rate limits)

**Endpoints Used**:
- `/mapping` - Map ticker symbols to FIGI identifiers

**Usage in MVP**:
- Currently optional - most APIs work directly with ticker symbols
- Used for ticker validation and company name lookup
- Can be disabled without affecting core functionality

**Rate Limits**:
- Free tier: 25 requests per minute
- API key provides 250 requests per minute

**Note**: For MVP, OpenFIGI is primarily used as a fallback for company name lookup when Finnhub profile is unavailable.

## Data Flow

### Request Flow for Stock Analysis

```
User Request (GET /api/stocks/AAPL/metrics)
    │
    ├─▶ Finnhub API
    │   ├─ /quote → Current price, volume
    │   └─ /stock/profile2 → Company name, market cap
    │
    ├─▶ SEC EDGAR API
    │   └─ /api/xbrl/companyfacts/CIK... → Financial statements
    │
    └─▶ OpenFIGI API (optional)
        └─ /mapping → Ticker validation
```

### Data Aggregation Strategy

1. **Parallel Requests**: All data sources are queried simultaneously for performance
2. **Graceful Degradation**: If one source fails, return partial data with warnings
3. **Error Handling**: Each data source has independent error handling
4. **Timeout Management**: 10-second timeout per external API call

## Error Handling

### Common Error Scenarios

1. **Invalid Ticker**:
   - Finnhub returns 0 for price
   - EDGAR CIK not found
   - Response: 400 Bad Request with clear error message

2. **API Rate Limit**:
   - Response: 429 Too Many Requests
   - Solution: Wait and retry, or upgrade API tier

3. **Missing Data**:
   - Not all companies have complete financial data
   - Response: Include available data with warnings array

4. **API Downtime**:
   - One source failing doesn't block entire response
   - Return partial data with data_freshness indicators

### Example Error Response

```json
{
  "ticker": "AAPL",
  "company_name": "Apple Inc.",
  "current_price": 175.43,
  "warnings": [
    "Free cash flow data is from Q3 2024, Q4 not yet available",
    "P/E ratio calculated without 5-year historical average"
  ],
  "data_freshness": {
    "price": "real-time",
    "fundamentals": "2024-Q3"
  }
}
```

## Mock Data Mode

For development and testing without API keys:

```bash
export USE_MOCK_DATA=true
```

This enables:
- Synthetic price data
- Sample financial statements
- No external API calls
- Instant responses

**Use Cases**:
- Local development
- CI/CD testing
- API key not yet obtained

## Setting Up API Keys

### 1. Finnhub API Key

```bash
# Get free API key
1. Visit https://finnhub.io/register
2. Sign up for free account
3. Navigate to Dashboard
4. Copy your API key

# Set environment variable
export FINNHUB_API_KEY="your_api_key_here"

# Or in Terraform
export TF_VAR_finnhub_api_key="your_api_key_here"
```

### 2. SEC EDGAR User-Agent

```bash
# Format: AppName/Version (contact-email)
export EDGAR_USER_AGENT="finEdSkywalker/1.0 (your-email@example.com)"
```

**Important**: Replace with your actual email. SEC may contact you if there are issues with your requests.

### 3. OpenFIGI (Optional)

No API key required for MVP. For higher rate limits:

```bash
# Visit https://www.openfigi.com/api
# Request API key (optional)
```

## Data Freshness and Accuracy

### Price Data (Finnhub)
- **Frequency**: Real-time (15-minute delay for free tier)
- **Market Hours**: Updates during trading hours
- **After Hours**: Last closing price

### Financial Data (SEC EDGAR)
- **Frequency**: Updated when companies file (10-K annual, 10-Q quarterly)
- **Typical Delay**: 
  - Annual: 60-90 days after fiscal year end
  - Quarterly: 40-45 days after quarter end
- **Historical Data**: Available back to 2009

### Calculations
- **P/E Ratio**: Calculated in real-time from current price and latest earnings
- **FCF Yield**: Based on latest annual or TTM (trailing twelve months) data
- **DCF Valuation**: Uses latest annual data for projections

## API Response Times

Typical latency (with good network):
- Finnhub: 200-500ms
- SEC EDGAR: 500-1500ms
- OpenFIGI: 300-600ms

**Total request time**: 1-2 seconds for comprehensive analysis (parallel requests)

## Compliance and Legal

### SEC EDGAR Terms
- Data is public domain
- Must include User-Agent header
- Rate limiting applies
- See: https://www.sec.gov/privacy

### Finnhub Terms
- Free tier for personal/non-commercial use
- Review Terms of Service: https://finnhub.io/terms
- Attribution may be required for certain use cases

### OpenFIGI Terms
- Free for non-commercial use
- Review Terms: https://www.openfigi.com/terms

## Future Enhancements

### Planned Data Sources

1. **Alpha Vantage**: Additional fundamental data
2. **Yahoo Finance**: Historical prices, splits, dividends
3. **News APIs**: Sentiment analysis data
4. **Insider Trading Data**: SEC Form 4 filings

### Caching Strategy

For production, implement:
- Redis/DynamoDB for caching EDGAR data (updates infrequently)
- Price data: 1-minute cache
- Financial data: 24-hour cache
- Async background jobs to refresh data

## Troubleshooting

### "CIK not found for ticker"
- Ticker may not be supported in MVP
- Add mapping to `internal/datasources/edgar.go`
- Or implement dynamic CIK lookup

### "API error (status 403)"
- Check EDGAR User-Agent header
- Ensure it includes valid email

### "No data available"
- Verify ticker symbol is correct
- Check if company has filed with SEC (US companies only)
- Try mock mode: `USE_MOCK_DATA=true`

### "Rate limit exceeded"
- Wait before retrying
- Consider API key upgrade
- Implement request caching

## Support

- **Finnhub Support**: https://finnhub.io/contact
- **SEC EDGAR Help**: https://www.sec.gov/edgar/searchedgar/webusers.htm
- **OpenFIGI Support**: Via their website contact form

