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
	openFIGIBaseURL = "https://api.openfigi.com/v3"
)

// OpenFIGIClient handles interactions with OpenFIGI API
type OpenFIGIClient struct {
	httpClient *http.Client
	useMock    bool
}

// OpenFIGI API structures
type openFIGIRequest struct {
	IDType string `json:"idType"`
	IDValue string `json:"idValue"`
}

type openFIGIResponse struct {
	Data []openFIGIData `json:"data"`
	Error string         `json:"error,omitempty"`
}

type openFIGIData struct {
	FIGI            string `json:"figi"`
	SecurityType    string `json:"securityType"`
	MarketSector    string `json:"marketSector"`
	Ticker          string `json:"ticker"`
	Name            string `json:"name"`
	ExchangeCode    string `json:"exchCode"`
	CompositeFIGI   string `json:"compositeFIGI"`
	SecurityType2   string `json:"securityType2"`
}

// NewOpenFIGIClient creates a new OpenFIGI API client
func NewOpenFIGIClient() *OpenFIGIClient {
	cfg := config.GetConfig()
	return &OpenFIGIClient{
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		useMock: cfg.IsMockMode(),
	}
}

// MapTicker converts a ticker symbol to FIGI identifier
func (c *OpenFIGIClient) MapTicker(ticker string) (string, string, error) {
	if c.useMock {
		return c.getMockFIGI(ticker)
	}

	// For MVP, OpenFIGI mapping is optional - many APIs work with ticker directly
	// Return the ticker as-is for now
	return "", ticker, nil
}

// SearchTicker looks up company information by ticker
func (c *OpenFIGIClient) SearchTicker(ticker string) (*openFIGIData, error) {
	if c.useMock {
		figi, name, _ := c.getMockFIGI(ticker)
		return &openFIGIData{
			FIGI:   figi,
			Ticker: ticker,
			Name:   name,
		}, nil
	}

	// OpenFIGI API endpoint for mapping
	endpoint := fmt.Sprintf("%s/mapping", openFIGIBaseURL)
	
	// Create request body
	reqBody := []openFIGIRequest{
		{
			IDType:  "TICKER",
			IDValue: strings.ToUpper(ticker),
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(reqJSON)))
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: fmt.Sprintf("failed to fetch mapping: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)),
			Code:    fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	var responses []openFIGIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: fmt.Sprintf("failed to parse response: %v", err),
		}
	}

	if len(responses) == 0 || len(responses[0].Data) == 0 {
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: fmt.Sprintf("no mapping found for ticker %s", ticker),
			Code:    "NO_MAPPING",
		}
	}

	if responses[0].Error != "" {
		return nil, &finance.DataSourceError{
			Source:  "OpenFIGI",
			Message: responses[0].Error,
			Code:    "API_ERROR",
		}
	}

	// Return the first match (usually the primary exchange listing)
	return &responses[0].Data[0], nil
}

// Mock data for testing
func (c *OpenFIGIClient) getMockFIGI(ticker string) (string, string, error) {
	mockData := map[string]struct {
		FIGI string
		Name string
	}{
		"AAPL":  {"BBG000B9XRY4", "Apple Inc."},
		"MSFT":  {"BBG000BPH459", "Microsoft Corporation"},
		"GOOGL": {"BBG009S39JX6", "Alphabet Inc."},
		"AMZN":  {"BBG000BVPV84", "Amazon.com Inc."},
		"TSLA":  {"BBG000N9MNX3", "Tesla Inc."},
	}

	data, ok := mockData[strings.ToUpper(ticker)]
	if !ok {
		return "", "Mock Company Inc.", nil
	}

	return data.FIGI, data.Name, nil
}

