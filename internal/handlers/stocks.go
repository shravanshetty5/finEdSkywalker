package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sshetty/finEdSkywalker/internal/auth"
	"github.com/sshetty/finEdSkywalker/internal/calculator"
	"github.com/sshetty/finEdSkywalker/internal/datasources"
	"github.com/sshetty/finEdSkywalker/internal/finance"
)

// StockService aggregates data from multiple sources
type StockService struct {
	finnhub  *datasources.FinnhubClient
	edgar    *datasources.EDGARClient
	openfigi *datasources.OpenFIGIClient
}

// NewStockService creates a new stock service
func NewStockService() *StockService {
	return &StockService{
		finnhub:  datasources.NewFinnhubClient(),
		edgar:    datasources.NewEDGARClient(),
		openfigi: datasources.NewOpenFIGIClient(),
	}
}

// GetCompanyData aggregates data from all sources
func (s *StockService) GetCompanyData(ticker string) (*finance.CompanyData, []string) {
	warnings := []string{}
	companyData := &finance.CompanyData{
		Ticker: strings.ToUpper(ticker),
	}

	// 1. Get stock quote from Finnhub
	quote, err := s.finnhub.GetQuote(ticker)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Price data unavailable: %v", err))
		log.Printf("Finnhub quote error for %s: %v", ticker, err)
	} else {
		companyData.Quote = quote
	}

	// 2. Get company profile for name and market cap
	profile, err := s.finnhub.GetProfile(ticker)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Company profile unavailable: %v", err))
		log.Printf("Finnhub profile error for %s: %v", ticker, err)
	} else {
		companyData.CompanyName = profile.Name
		if companyData.Quote != nil {
			companyData.Quote.CompanyName = profile.Name
			companyData.Quote.MarketCap = profile.MarketCap * 1_000_000 // Convert to actual value
		}
		companyData.SharesOutstanding = profile.SharesOut // In millions
	}

	// 3. Get fundamental data from SEC EDGAR
	financials, err := s.edgar.GetCompanyFacts(ticker)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Fundamental data unavailable: %v", err))
		log.Printf("EDGAR error for %s: %v", ticker, err)
	} else {
		companyData.LatestFinancials = financials
	}

	// 4. Get FIGI mapping (optional)
	figi, name, err := s.openfigi.MapTicker(ticker)
	if err != nil {
		log.Printf("OpenFIGI error for %s: %v", ticker, err)
		// Not adding to warnings as FIGI is optional
	} else {
		companyData.FIGI = figi
		if companyData.CompanyName == "" {
			companyData.CompanyName = name
		}
	}

	// Set default company name if still not set
	if companyData.CompanyName == "" {
		companyData.CompanyName = ticker
	}

	return companyData, warnings
}

// handleStockFundamentalsAuth is the authenticated version of handleStockFundamentals
func handleStockFundamentalsAuth(request events.APIGatewayV2HTTPRequest, authCtx *auth.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ticker from path
	parts := strings.Split(request.RawPath, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Ticker symbol is required")
	}
	ticker := strings.ToUpper(parts[3])

	log.Printf("User %s (%s) requesting fundamentals for %s", authCtx.Username, authCtx.UserID, ticker)
	return handleStockFundamentals(request)
}

// handleStockFundamentals returns the Big 5 fundamental scorecard
func handleStockFundamentals(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ticker from path
	parts := strings.Split(request.RawPath, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Ticker symbol is required")
	}
	ticker := strings.ToUpper(parts[3])

	log.Printf("Fetching fundamentals for ticker: %s", ticker)

	// Get company data
	service := NewStockService()
	companyData, warnings := service.GetCompanyData(ticker)

	// Calculate scorecard
	scorecard := calculator.CalculateScorecard(companyData)

	// Build response
	response := finance.StockAnalysisResponse{
		Ticker:               ticker,
		CompanyName:          companyData.CompanyName,
		LastUpdated:          time.Now(),
		FundamentalScorecard: scorecard,
		Warnings:             warnings,
		DataFreshness:        buildDataFreshness(companyData),
	}

	if companyData.Quote != nil {
		response.CurrentPrice = companyData.Quote.CurrentPrice
	}

	return jsonResponse(200, response)
}

// handleStockValuationAuth is the authenticated version of handleStockValuation
func handleStockValuationAuth(request events.APIGatewayV2HTTPRequest, authCtx *auth.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ticker from path
	parts := strings.Split(request.RawPath, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Ticker symbol is required")
	}
	ticker := strings.ToUpper(parts[3])

	log.Printf("User %s (%s) requesting valuation for %s", authCtx.Username, authCtx.UserID, ticker)
	return handleStockValuation(request)
}

// handleStockValuation returns DCF valuation
func handleStockValuation(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ticker from path
	parts := strings.Split(request.RawPath, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Ticker symbol is required")
	}
	ticker := strings.ToUpper(parts[3])

	log.Printf("Calculating valuation for ticker: %s", ticker)

	// Parse query parameters for DCF inputs
	dcfInput := parseDCFInput(request.QueryStringParameters)

	// Get company data
	service := NewStockService()
	companyData, warnings := service.GetCompanyData(ticker)

	// Calculate DCF valuation
	valuation, err := calculator.CalculateDCF(companyData, dcfInput)
	if err != nil {
		return errorResponse(400, "Valuation failed", err.Error())
	}

	// Build response
	response := finance.StockAnalysisResponse{
		Ticker:        ticker,
		CompanyName:   companyData.CompanyName,
		LastUpdated:   time.Now(),
		Valuation:     valuation,
		Warnings:      warnings,
		DataFreshness: buildDataFreshness(companyData),
	}

	if companyData.Quote != nil {
		response.CurrentPrice = companyData.Quote.CurrentPrice
	}

	return jsonResponse(200, response)
}

// handleStockMetricsAuth is the authenticated version of handleStockMetrics
func handleStockMetricsAuth(request events.APIGatewayV2HTTPRequest, authCtx *auth.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ticker from path
	parts := strings.Split(request.RawPath, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Ticker symbol is required")
	}
	ticker := strings.ToUpper(parts[3])

	log.Printf("User %s (%s) requesting metrics for %s", authCtx.Username, authCtx.UserID, ticker)
	return handleStockMetrics(request)
}

// handleStockMetrics returns comprehensive metrics (fundamentals + valuation)
func handleStockMetrics(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ticker from path
	parts := strings.Split(request.RawPath, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Ticker symbol is required")
	}
	ticker := strings.ToUpper(parts[3])

	log.Printf("Fetching comprehensive metrics for ticker: %s", ticker)

	// Parse query parameters for DCF inputs
	dcfInput := parseDCFInput(request.QueryStringParameters)

	// Get company data
	service := NewStockService()
	companyData, warnings := service.GetCompanyData(ticker)

	// Calculate scorecard
	scorecard := calculator.CalculateScorecard(companyData)

	// Calculate DCF valuation
	valuation, err := calculator.CalculateDCF(companyData, dcfInput)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Valuation calculation failed: %v", err))
		log.Printf("DCF error for %s: %v", ticker, err)
	}

	// Build comprehensive response
	response := finance.StockAnalysisResponse{
		Ticker:               ticker,
		CompanyName:          companyData.CompanyName,
		LastUpdated:          time.Now(),
		FundamentalScorecard: scorecard,
		Valuation:            valuation,
		Warnings:             warnings,
		DataFreshness:        buildDataFreshness(companyData),
	}

	if companyData.Quote != nil {
		response.CurrentPrice = companyData.Quote.CurrentPrice
	}

	return jsonResponse(200, response)
}

// parseDCFInput extracts DCF parameters from query string
func parseDCFInput(params map[string]string) *calculator.DCFInput {
	input := &calculator.DCFInput{}

	if val, ok := params["revenue_growth"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			input.RevenueGrowthRate = &f
		}
	}

	if val, ok := params["profit_margin"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			input.ProfitMargin = &f
		}
	}

	if val, ok := params["fcf_margin"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			input.FCFMargin = &f
		}
	}

	if val, ok := params["discount_rate"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			input.DiscountRate = &f
		}
	}

	if val, ok := params["terminal_growth"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			input.TerminalGrowthRate = &f
		}
	}

	return input
}

// buildDataFreshness creates a map of data source freshness
func buildDataFreshness(data *finance.CompanyData) map[string]string {
	freshness := make(map[string]string)

	if data.Quote != nil {
		freshness["price"] = "real-time"
	} else {
		freshness["price"] = "unavailable"
	}

	if data.LatestFinancials != nil {
		if !data.LatestFinancials.ReportDate.IsZero() {
			freshness["fundamentals"] = data.LatestFinancials.Period
		} else {
			freshness["fundamentals"] = "available"
		}
	} else {
		freshness["fundamentals"] = "unavailable"
	}

	return freshness
}

// Helper to parse JSON request body
func parseJSONBody(body string, target interface{}) error {
	if body == "" {
		return fmt.Errorf("request body is empty")
	}
	return json.Unmarshal([]byte(body), target)
}
