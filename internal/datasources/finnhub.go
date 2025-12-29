package datasources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sshetty/finEdSkywalker/internal/config"
	"github.com/sshetty/finEdSkywalker/internal/finance"
)

const (
	finnhubBaseURL = "https://finnhub.io/api/v1"
)

// FinnhubClient handles interactions with Finnhub API
type FinnhubClient struct {
	apiKey     string
	httpClient *http.Client
	useMock    bool
}

// Finnhub API response structures
type finnhubQuoteResponse struct {
	C  float64 `json:"c"`  // Current price
	D  float64 `json:"d"`  // Change
	DP float64 `json:"dp"` // Percent change
	H  float64 `json:"h"`  // High
	L  float64 `json:"l"`  // Low
	O  float64 `json:"o"`  // Open
	PC float64 `json:"pc"` // Previous close
	T  int64   `json:"t"`  // Timestamp
}

type finnhubProfileResponse struct {
	Country        string  `json:"country"`
	Currency       string  `json:"currency"`
	Exchange       string  `json:"exchange"`
	Name           string  `json:"name"`
	Ticker         string  `json:"ticker"`
	MarketCap      float64 `json:"marketCapitalization"`
	SharesOut      float64 `json:"shareOutstanding"`
	Logo           string  `json:"logo"`
	Phone          string  `json:"phone"`
	WebURL         string  `json:"weburl"`
	IPO            string  `json:"ipo"`
	FinnhubIndustry string `json:"finnhubIndustry"`
}

type finnhubMetricResponse struct {
	Metric struct {
		PE10YearHigh      float64 `json:"10DayAverageTradingVolume"`
		WeekHigh52        float64 `json:"52WeekHigh"`
		WeekLow52         float64 `json:"52WeekLow"`
		Beta              float64 `json:"beta"`
		PERatio           float64 `json:"peBasicExclExtraTTM"`
		PEForward         float64 `json:"peNormalizedAnnual"`
		PEGRatio          float64 `json:"peTTM"`
		DividendYield     float64 `json:"dividendYieldIndicatedAnnual"`
		ROE               float64 `json:"roeTTM"`
		ROA               float64 `json:"roaTTM"`
		QuickRatio        float64 `json:"quickRatioAnnual"`
		CurrentRatio      float64 `json:"currentRatioAnnual"`
		DebtToEquity      float64 `json:"totalDebt/totalEquityAnnual"`
	} `json:"metric"`
}

// NewFinnhubClient creates a new Finnhub API client
func NewFinnhubClient() *FinnhubClient {
	cfg := config.GetConfig()
	return &FinnhubClient{
		apiKey: cfg.FinnhubAPIKey,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		useMock: cfg.IsMockMode(),
	}
}

// GetQuote fetches real-time quote for a ticker
func (c *FinnhubClient) GetQuote(ticker string) (*finance.StockQuote, error) {
	if c.useMock {
		return c.getMockQuote(ticker), nil
	}

	endpoint := fmt.Sprintf("%s/quote", finnhubBaseURL)
	params := url.Values{}
	params.Add("symbol", ticker)
	params.Add("token", c.apiKey)

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("failed to fetch quote: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			Code:    fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	var quoteResp finnhubQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("failed to parse quote response: %v", err),
		}
	}

	// If current price is 0, the ticker might be invalid
	if quoteResp.C == 0 {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("invalid ticker or no data available for %s", ticker),
			Code:    "NO_DATA",
		}
	}

	quote := &finance.StockQuote{
		Ticker:        ticker,
		CurrentPrice:  quoteResp.C,
		Change:        quoteResp.D,
		ChangePercent: quoteResp.DP,
		High:          quoteResp.H,
		Low:           quoteResp.L,
		Open:          quoteResp.O,
		PreviousClose: quoteResp.PC,
		Timestamp:     time.Unix(quoteResp.T, 0),
	}

	return quote, nil
}

// GetProfile fetches company profile including name and market cap
func (c *FinnhubClient) GetProfile(ticker string) (*finnhubProfileResponse, error) {
	if c.useMock {
		return c.getMockProfile(ticker), nil
	}

	endpoint := fmt.Sprintf("%s/stock/profile2", finnhubBaseURL)
	params := url.Values{}
	params.Add("symbol", ticker)
	params.Add("token", c.apiKey)

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("failed to fetch profile: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			Code:    fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	var profile finnhubProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("failed to parse profile response: %v", err),
		}
	}

	return &profile, nil
}

// GetMetrics fetches basic financial metrics
func (c *FinnhubClient) GetMetrics(ticker string) (*finnhubMetricResponse, error) {
	if c.useMock {
		return c.getMockMetrics(ticker), nil
	}

	endpoint := fmt.Sprintf("%s/stock/metric", finnhubBaseURL)
	params := url.Values{}
	params.Add("symbol", ticker)
	params.Add("metric", "all")
	params.Add("token", c.apiKey)

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("failed to fetch metrics: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			Code:    fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	var metrics finnhubMetricResponse
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, &finance.DataSourceError{
			Source:  "Finnhub",
			Message: fmt.Sprintf("failed to parse metrics response: %v", err),
		}
	}

	return &metrics, nil
}

// Mock data for testing without API keys
func (c *FinnhubClient) getMockQuote(ticker string) *finance.StockQuote {
	// Mock data for AAPL
	return &finance.StockQuote{
		Ticker:        ticker,
		CompanyName:   "Mock Company Inc.",
		CurrentPrice:  175.43,
		Change:        2.15,
		ChangePercent: 1.24,
		High:          176.50,
		Low:           173.20,
		Open:          174.00,
		PreviousClose: 173.28,
		Volume:        52000000,
		MarketCap:     2800000000000, // $2.8T
		Timestamp:     time.Now(),
	}
}

func (c *FinnhubClient) getMockProfile(ticker string) *finnhubProfileResponse {
	return &finnhubProfileResponse{
		Name:      "Mock Company Inc.",
		Ticker:    ticker,
		MarketCap: 2800000, // In millions
		SharesOut: 16000,   // In millions
		Country:   "US",
		Currency:  "USD",
		Exchange:  "NASDAQ",
	}
}

func (c *FinnhubClient) getMockMetrics(ticker string) *finnhubMetricResponse {
	metrics := &finnhubMetricResponse{}
	metrics.Metric.PERatio = 28.5
	metrics.Metric.PEForward = 26.2
	metrics.Metric.ROE = 0.47
	metrics.Metric.DebtToEquity = 0.85
	return metrics
}

