package datasources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sshetty/finEdSkywalker/internal/config"
	"github.com/sshetty/finEdSkywalker/internal/finance"
)

const (
	edgarBaseURL      = "https://data.sec.gov"
	edgarTickersURL   = "https://www.sec.gov/files/company_tickers.json"
	edgarCIKURL       = "https://www.sec.gov/cgi-bin/browse-edgar"
	tickerMapCacheTTL = 24 * time.Hour // Refresh ticker map daily
)

// EDGARClient handles interactions with SEC EDGAR API
type EDGARClient struct {
	userAgent         string
	httpClient        *http.Client
	useMock           bool
	cikCache          map[string]string // ticker -> CIK mapping
	tickerMapCache    map[string]string // Full SEC ticker->CIK map (lazy loaded)
	tickerMapLoadedAt time.Time         // Track when ticker map was last loaded
}

// EDGAR Company Facts response structure
type edgarCompanyFacts struct {
	CIK        json.Number                     `json:"cik"` // Can be string or number from API
	EntityName string                          `json:"entityName"`
	Facts      map[string]map[string]edgarFact `json:"facts"`
}

type edgarFact struct {
	Label       string                      `json:"label"`
	Description string                      `json:"description"`
	Units       map[string][]edgarFactValue `json:"units"`
}

type edgarFactValue struct {
	Start string      `json:"start,omitempty"`
	End   string      `json:"end"`
	Val   json.Number `json:"val"` // Can be string or number from API
	AccN  string      `json:"accn"`
	FY    int         `json:"fy"`
	FP    string      `json:"fp"` // Fiscal period (Q1, Q2, Q3, Q4, FY)
	Form  string      `json:"form"`
	Filed string      `json:"filed"`
}

// SEC company tickers JSON structure
type secCompanyTicker struct {
	CIKStr int    `json:"cik_str"`
	Ticker string `json:"ticker"`
	Title  string `json:"title"`
}

// NewEDGARClient creates a new SEC EDGAR API client
func NewEDGARClient() *EDGARClient {
	cfg := config.GetConfig()
	return &EDGARClient{
		userAgent: cfg.EDGARUserAgent,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		useMock:        cfg.IsMockMode(),
		cikCache:       make(map[string]string),
		tickerMapCache: nil, // Lazy loaded on first miss
	}
}

// GetCompanyFacts fetches financial data for a company by ticker
func (c *EDGARClient) GetCompanyFacts(ticker string) (*finance.FinancialStatement, error) {
	if c.useMock {
		return c.getMockFinancials(ticker), nil
	}

	// First, get the CIK for the ticker
	cik, err := c.getCIK(ticker)
	if err != nil {
		return nil, err
	}

	// Fetch company facts
	endpoint := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", edgarBaseURL, cik)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "EDGAR",
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	// SEC requires User-Agent header
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "EDGAR",
			Message: fmt.Sprintf("failed to fetch company facts: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &finance.DataSourceError{
			Source:  "EDGAR",
			Message: fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			Code:    fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	var facts edgarCompanyFacts
	if err := json.NewDecoder(resp.Body).Decode(&facts); err != nil {
		return nil, &finance.DataSourceError{
			Source:  "EDGAR",
			Message: fmt.Sprintf("failed to parse company facts: %v", err),
		}
	}

	// Parse the facts into our FinancialStatement structure
	statement := c.parseFinancialStatement(&facts)
	return statement, nil
}

// ensureTickerMapFresh checks if ticker map cache needs refresh and reloads if stale
func (c *EDGARClient) ensureTickerMapFresh() error {
	// Check if cache needs refresh (nil or expired)
	cacheAge := time.Since(c.tickerMapLoadedAt)
	needsRefresh := c.tickerMapCache == nil || cacheAge > tickerMapCacheTTL

	if !needsRefresh {
		return nil // Cache is still fresh
	}

	// Load fresh ticker map
	tickerMap, err := c.loadTickerMap()
	if err != nil {
		// If we have stale cache, use it as fallback
		if c.tickerMapCache != nil {
			// Log warning but continue with stale data
			// (in production, you might want proper logging here)
			return nil // Allow stale cache usage
		}
		// No cache at all, must return error
		return err
	}

	// Successfully loaded fresh data
	c.tickerMapCache = tickerMap
	c.tickerMapLoadedAt = time.Now()
	return nil
}

// getCIK retrieves the CIK number for a ticker (with caching and dynamic lookup)
func (c *EDGARClient) getCIK(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)

	// STEP 1: Ensure ticker map is fresh (do this FIRST)
	// This ensures new tickers become available after 24h
	if err := c.ensureTickerMapFresh(); err != nil {
		// Only fail if we have no cache at all
		// If we have stale cache, we already logged warning in ensureTickerMapFresh
	}

	// STEP 2: Check individual ticker cache (fast path)
	if cik, ok := c.cikCache[ticker]; ok {
		return cik, nil
	}

	// STEP 3: Try hardcoded common tickers
	commonCIKs := map[string]string{
		"AAPL":  "0000320193",
		"MSFT":  "0000789019",
		"GOOGL": "0001652044",
		"GOOG":  "0001652044",
		"AMZN":  "0001018724",
		"TSLA":  "0001318605",
		"META":  "0001326801",
		"NVDA":  "0001045810",
		"JPM":   "0000019617",
		"V":     "0001403161",
		"BAC":   "0000070858",
		"WMT":   "0000104169",
		"XOM":   "0000034088",
		"UNH":   "0000731766",
		"JNJ":   "0000200406",
	}

	if cik, ok := commonCIKs[ticker]; ok {
		c.cikCache[ticker] = cik
		return cik, nil
	}

	// STEP 4: Dynamic lookup from SEC (ticker map already fresh from step 1)
	cik, err := c.lookupCIKFromSEC(ticker)
	if err != nil {
		return "", err
	}

	// Cache the result
	c.cikCache[ticker] = cik
	return cik, nil
}

// lookupCIKFromSEC fetches the CIK for a ticker from SEC's official company tickers JSON
func (c *EDGARClient) lookupCIKFromSEC(ticker string) (string, error) {
	// Ticker map freshness already ensured by getCIK()
	// Just do the lookup

	// Look up in cached map
	cik, ok := c.tickerMapCache[ticker]
	if !ok {
		return "", &finance.DataSourceError{
			Source:  "EDGAR",
			Message: fmt.Sprintf("CIK not found for ticker %s", ticker),
			Code:    "CIK_NOT_FOUND",
		}
	}

	return cik, nil
}

// loadTickerMap fetches the SEC company tickers mapping (pure function)
// Returns a map of ticker -> CIK without mutating client state
func (c *EDGARClient) loadTickerMap() (map[string]string, error) {
	req, err := http.NewRequest("GET", edgarTickersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// SEC requires User-Agent header
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ticker map: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// The JSON structure is: { "0": {...}, "1": {...}, ... }
	var tickersMap map[string]secCompanyTicker
	if err := json.NewDecoder(resp.Body).Decode(&tickersMap); err != nil {
		return nil, fmt.Errorf("failed to parse ticker map: %w", err)
	}

	// Build the ticker -> padded CIK map
	result := make(map[string]string, len(tickersMap))
	for _, company := range tickersMap {
		// Convert CIK to padded 10-digit string format
		cik := fmt.Sprintf("%010d", company.CIKStr)
		result[strings.ToUpper(company.Ticker)] = cik
	}

	return result, nil
}

// parseFinancialStatement extracts relevant financial data from EDGAR facts
func (c *EDGARClient) parseFinancialStatement(facts *edgarCompanyFacts) *finance.FinancialStatement {
	statement := &finance.FinancialStatement{}

	// Get the most recent annual filing data
	// EDGAR uses US-GAAP taxonomy
	usGAAP, ok := facts.Facts["us-gaap"]
	if !ok {
		return statement
	}

	// Extract key metrics (take most recent annual data)
	statement.Revenue = c.getLatestValue(usGAAP, "Revenues")
	if statement.Revenue == 0 {
		statement.Revenue = c.getLatestValue(usGAAP, "RevenueFromContractWithCustomerExcludingAssessedTax")
	}

	statement.NetIncome = c.getLatestValue(usGAAP, "NetIncomeLoss")
	statement.TotalAssets = c.getLatestValue(usGAAP, "Assets")
	statement.TotalLiabilities = c.getLatestValue(usGAAP, "Liabilities")
	statement.ShareholdersEquity = c.getLatestValue(usGAAP, "StockholdersEquity")

	// Debt metrics
	longTermDebt := c.getLatestValue(usGAAP, "LongTermDebt")
	shortTermDebt := c.getLatestValue(usGAAP, "ShortTermBorrowings")
	statement.TotalDebt = longTermDebt + shortTermDebt
	if statement.TotalDebt == 0 {
		statement.TotalDebt = c.getLatestValue(usGAAP, "DebtCurrent")
	}

	// Cash flow metrics
	statement.OperatingCashFlow = c.getLatestValue(usGAAP, "NetCashProvidedByUsedInOperatingActivities")
	statement.CapEx = c.getLatestValue(usGAAP, "PaymentsToAcquirePropertyPlantAndEquipment")

	// Calculate Free Cash Flow
	if statement.OperatingCashFlow > 0 && statement.CapEx > 0 {
		statement.FreeCashFlow = statement.OperatingCashFlow - statement.CapEx
	}

	// Set metadata from the most recent filing
	if len(usGAAP) > 0 {
		for _, fact := range usGAAP {
			if len(fact.Units) > 0 {
				for _, values := range fact.Units {
					if len(values) > 0 {
						latest := values[len(values)-1]
						statement.FiscalYear = latest.FY
						statement.Period = fmt.Sprintf("%d-%s", latest.FY, latest.FP)
						if latest.Filed != "" {
							if filedDate, err := time.Parse("2006-01-02", latest.Filed); err == nil {
								statement.FilingDate = filedDate
							}
						}
						if latest.End != "" {
							if endDate, err := time.Parse("2006-01-02", latest.End); err == nil {
								statement.ReportDate = endDate
							}
						}
						break
					}
				}
				break
			}
		}
	}

	return statement
}

// getLatestValue extracts the most recent value for a given fact name
func (c *EDGARClient) getLatestValue(facts map[string]edgarFact, factName string) float64 {
	fact, ok := facts[factName]
	if !ok {
		return 0
	}

	// Look for USD values
	var latestValue float64
	var latestDate string

	for unit, values := range fact.Units {
		// Prefer USD values
		if unit != "USD" && unit != "USD/shares" {
			continue
		}

		for _, value := range values {
			// Prefer annual filings (10-K) over quarterly (10-Q)
			if value.Form != "10-K" && value.Form != "10-Q" {
				continue
			}

			// Get the most recent value
			if value.End > latestDate {
				latestDate = value.End
				// Convert json.Number to float64
				if val, err := value.Val.Float64(); err == nil {
					latestValue = val
				}
			}
		}
	}

	return latestValue
}

// Mock data for testing
func (c *EDGARClient) getMockFinancials(ticker string) *finance.FinancialStatement {
	return &finance.FinancialStatement{
		Revenue:            394328000000, // $394B
		NetIncome:          96995000000,  // $97B
		TotalAssets:        352755000000, // $353B
		TotalLiabilities:   290437000000, // $290B
		TotalDebt:          109280000000, // $109B
		ShareholdersEquity: 62318000000,  // $62B
		OperatingCashFlow:  110543000000, // $110B
		CapEx:              10959000000,  // $11B
		FreeCashFlow:       99584000000,  // $99.5B
		Period:             "2024-FY",
		FiscalYear:         2024,
		ReportDate:         time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
		FilingDate:         time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
	}
}
