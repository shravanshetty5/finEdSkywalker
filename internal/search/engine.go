package search

import (
	"strings"
	"sync"

	"github.com/sahilm/fuzzy"
	"github.com/sshetty/finEdSkywalker/internal/datasources"
)

var (
	searchIndex     []TickerInfo
	searchIndexOnce sync.Once
	loadError       error
)

// SearchEngine performs fuzzy search on ticker data
type SearchEngine struct {
	tickers []TickerInfo
}

// NewSearchEngine creates a search engine with lazy-loaded SEC data
func NewSearchEngine() (*SearchEngine, error) {
	// Lazy load ticker data once per Lambda container
	searchIndexOnce.Do(func() {
		client := datasources.NewEDGARClient()
		data, err := client.LoadAllTickers()
		if err != nil {
			loadError = err
			return
		}

		// Convert to TickerInfo
		searchIndex = make([]TickerInfo, len(data))
		for i, td := range data {
			searchIndex[i] = TickerInfo{
				Ticker:      td.Ticker,
				CompanyName: td.CompanyName,
				CIK:         td.CIK,
			}
		}
	})

	if loadError != nil {
		return nil, loadError
	}

	return &SearchEngine{tickers: searchIndex}, nil
}

// Search performs fuzzy search on ticker symbols and company names
func (e *SearchEngine) Search(query string, limit int) []SearchResult {
	if query == "" || len(e.tickers) == 0 {
		return []SearchResult{}
	}

	query = strings.TrimSpace(query)
	queryUpper := strings.ToUpper(query)

	// Create searchable strings (ticker + company name for each entry)
	searchableStrings := make([]string, len(e.tickers))
	for i, ticker := range e.tickers {
		// Format: "AAPL Apple Inc." - ticker first for better matching
		searchableStrings[i] = ticker.Ticker + " " + ticker.CompanyName
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, searchableStrings)

	// Convert matches to results
	results := make([]SearchResult, 0, len(matches))
	for _, match := range matches {
		if len(results) >= limit {
			break
		}

		ticker := e.tickers[match.Index]
		results = append(results, SearchResult{
			Ticker:      ticker.Ticker,
			CompanyName: ticker.CompanyName,
		})
	}

	// If no fuzzy matches, try exact prefix matching as fallback
	if len(results) == 0 {
		results = e.prefixSearch(queryUpper, limit)
	}

	return results
}

// prefixSearch performs simple prefix matching as fallback
func (e *SearchEngine) prefixSearch(queryUpper string, limit int) []SearchResult {
	results := make([]SearchResult, 0, limit)

	for _, ticker := range e.tickers {
		if len(results) >= limit {
			break
		}

		// Check if ticker or company name starts with query
		if strings.HasPrefix(ticker.Ticker, queryUpper) ||
			strings.HasPrefix(strings.ToUpper(ticker.CompanyName), queryUpper) {
			results = append(results, SearchResult{
				Ticker:      ticker.Ticker,
				CompanyName: ticker.CompanyName,
			})
		}
	}

	return results
}
