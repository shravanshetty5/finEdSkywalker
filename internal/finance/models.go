package finance

import "time"

// StockQuote represents real-time price data for a stock
type StockQuote struct {
	Ticker      string    `json:"ticker"`
	CompanyName string    `json:"company_name"`
	CurrentPrice float64  `json:"current_price"`
	Change      float64   `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Open        float64   `json:"open"`
	PreviousClose float64 `json:"previous_close"`
	Volume      int64     `json:"volume"`
	MarketCap   float64   `json:"market_cap,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// FinancialStatement represents a company's financial data
type FinancialStatement struct {
	// Income Statement
	Revenue       float64 `json:"revenue"`
	NetIncome     float64 `json:"net_income"`
	EPS           float64 `json:"eps,omitempty"`
	
	// Balance Sheet
	TotalAssets          float64 `json:"total_assets"`
	TotalLiabilities     float64 `json:"total_liabilities"`
	TotalDebt            float64 `json:"total_debt"`
	ShareholdersEquity   float64 `json:"shareholders_equity"`
	
	// Cash Flow
	OperatingCashFlow    float64 `json:"operating_cash_flow"`
	CapEx                float64 `json:"capex"`
	FreeCashFlow         float64 `json:"free_cash_flow"`
	
	// Metadata
	Period      string    `json:"period"`       // e.g., "2024-Q3", "2024"
	FiscalYear  int       `json:"fiscal_year"`
	ReportDate  time.Time `json:"report_date"`
	FilingDate  time.Time `json:"filing_date,omitempty"`
}

// HistoricalMetrics represents historical data for trend analysis
type HistoricalMetrics struct {
	PERatios        []float64 `json:"pe_ratios,omitempty"`
	PERatioAvg5Year float64   `json:"pe_ratio_avg_5year,omitempty"`
	ROEHistory      []float64 `json:"roe_history,omitempty"`
	FCFYieldHistory []float64 `json:"fcf_yield_history,omitempty"`
}

// MetricRating represents a colored rating for a financial metric
type MetricRating string

const (
	RatingGreen  MetricRating = "GREEN"
	RatingYellow MetricRating = "YELLOW"
	RatingRed    MetricRating = "RED"
	RatingNA     MetricRating = "N/A"
)

// FundamentalMetric represents a single metric with its rating
type FundamentalMetric struct {
	Current      float64      `json:"current"`
	FiveYearAvg  *float64     `json:"five_year_avg,omitempty"`
	Rating       MetricRating `json:"rating"`
	Message      string       `json:"message"`
	Available    bool         `json:"available"`
}

// FundamentalScorecard represents the "Big 5" fundamental metrics
type FundamentalScorecard struct {
	PERatio       FundamentalMetric `json:"pe_ratio"`
	ForwardPE     FundamentalMetric `json:"forward_pe,omitempty"`
	DebtToEquity  FundamentalMetric `json:"debt_to_equity"`
	FCFYield      FundamentalMetric `json:"fcf_yield"`
	PEGRatio      FundamentalMetric `json:"peg_ratio"`
	ROE           FundamentalMetric `json:"roe"`
	
	OverallScore  string `json:"overall_score"`  // e.g., "4/5 metrics healthy"
	Summary       string `json:"summary"`
}

// DCFAssumptions represents the inputs for DCF valuation
type DCFAssumptions struct {
	RevenueGrowthRate  float64 `json:"revenue_growth_rate"`   // e.g., 0.08 for 8%
	ProfitMargin       float64 `json:"profit_margin"`          // e.g., 0.15 for 15%
	FCFMargin          float64 `json:"fcf_margin"`             // e.g., 0.12 for 12%
	DiscountRate       float64 `json:"discount_rate"`          // e.g., 0.10 for 10%
	TerminalGrowthRate float64 `json:"terminal_growth_rate"`   // e.g., 0.025 for 2.5%
	ProjectionYears    int     `json:"projection_years"`       // typically 5
	Source             string  `json:"source"`                 // "user_input", "analyst_consensus", "defaults"
}

// DCFProjection represents a single year's projection
type DCFProjection struct {
	Year              int     `json:"year"`
	Revenue           float64 `json:"revenue"`
	NetIncome         float64 `json:"net_income"`
	FreeCashFlow      float64 `json:"free_cash_flow"`
	DiscountFactor    float64 `json:"discount_factor"`
	PresentValue      float64 `json:"present_value"`
}

// ValuationResult represents the DCF valuation output
type ValuationResult struct {
	FairValuePerShare float64         `json:"fair_value_per_share"`
	CurrentPrice      float64         `json:"current_price"`
	UpsidePercent     float64         `json:"upside_percent"`     // Positive = undervalued, Negative = overvalued
	Model             string          `json:"model"`              // "DCF"
	Assumptions       DCFAssumptions  `json:"assumptions"`
	Projections       []DCFProjection `json:"projections,omitempty"`
	TerminalValue     float64         `json:"terminal_value,omitempty"`
	EnterpriseValue   float64         `json:"enterprise_value,omitempty"`
	SharesOutstanding float64         `json:"shares_outstanding,omitempty"`
}

// CompanyData represents aggregated company data from all sources
type CompanyData struct {
	Ticker            string              `json:"ticker"`
	CompanyName       string              `json:"company_name"`
	CIK               string              `json:"cik,omitempty"`
	FIGI              string              `json:"figi,omitempty"`
	Quote             *StockQuote         `json:"quote,omitempty"`
	LatestFinancials  *FinancialStatement `json:"latest_financials,omitempty"`
	HistoricalData    *HistoricalMetrics  `json:"historical_data,omitempty"`
	SharesOutstanding float64             `json:"shares_outstanding,omitempty"`
}

// StockAnalysisResponse represents the complete API response
type StockAnalysisResponse struct {
	Ticker               string                `json:"ticker"`
	CompanyName          string                `json:"company_name"`
	CurrentPrice         float64               `json:"current_price"`
	LastUpdated          time.Time             `json:"last_updated"`
	FundamentalScorecard *FundamentalScorecard `json:"fundamental_scorecard,omitempty"`
	Valuation            *ValuationResult      `json:"valuation,omitempty"`
	Warnings             []string              `json:"warnings,omitempty"`
	DataFreshness        map[string]string     `json:"data_freshness,omitempty"`
}

// DataSourceError represents an error from a data source
type DataSourceError struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (e *DataSourceError) Error() string {
	return e.Source + ": " + e.Message
}

