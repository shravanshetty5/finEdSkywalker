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
	edgarBaseURL = "https://data.sec.gov"
	edgarCIKURL  = "https://www.sec.gov/cgi-bin/browse-edgar"
)

// EDGARClient handles interactions with SEC EDGAR API
type EDGARClient struct {
	userAgent  string
	httpClient *http.Client
	useMock    bool
	cikCache   map[string]string // ticker -> CIK mapping
}

// EDGAR Company Facts response structure
type edgarCompanyFacts struct {
	CIK        json.Number                  `json:"cik"` // Can be string or number from API
	EntityName string                       `json:"entityName"`
	Facts      map[string]map[string]edgarFact `json:"facts"`
}

type edgarFact struct {
	Label       string           `json:"label"`
	Description string           `json:"description"`
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

// NewEDGARClient creates a new SEC EDGAR API client
func NewEDGARClient() *EDGARClient {
	cfg := config.GetConfig()
	return &EDGARClient{
		userAgent: cfg.EDGARUserAgent,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		useMock:  cfg.IsMockMode(),
		cikCache: make(map[string]string),
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

// getCIK retrieves the CIK number for a ticker (with caching)
func (c *EDGARClient) getCIK(ticker string) (string, error) {
	// Check cache first
	if cik, ok := c.cikCache[ticker]; ok {
		return cik, nil
	}

	// For MVP, use a hardcoded mapping for common tickers
	// In production, you'd want to use OpenFIGI or SEC's ticker lookup
	commonCIKs := map[string]string{
		"AAPL": "0000320193",
		"MSFT": "0000789019",
		"GOOGL": "0001652044",
		"GOOG": "0001652044",
		"AMZN": "0001018724",
		"TSLA": "0001318605",
		"META": "0001326801",
		"NVDA": "0001045810",
		"JPM":  "0000019617",
		"V":    "0001403161",
	}

	cik, ok := commonCIKs[strings.ToUpper(ticker)]
	if !ok {
		return "", &finance.DataSourceError{
			Source:  "EDGAR",
			Message: fmt.Sprintf("CIK not found for ticker %s (limited ticker support in MVP)", ticker),
			Code:    "CIK_NOT_FOUND",
		}
	}

	// Cache the result
	c.cikCache[ticker] = cik
	return cik, nil
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
		Revenue:              394328000000,  // $394B
		NetIncome:            96995000000,   // $97B
		TotalAssets:          352755000000,  // $353B
		TotalLiabilities:     290437000000,  // $290B
		TotalDebt:            109280000000,  // $109B
		ShareholdersEquity:   62318000000,   // $62B
		OperatingCashFlow:    110543000000,  // $110B
		CapEx:                10959000000,   // $11B
		FreeCashFlow:         99584000000,   // $99.5B
		Period:               "2024-FY",
		FiscalYear:           2024,
		ReportDate:           time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
		FilingDate:           time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
	}
}

