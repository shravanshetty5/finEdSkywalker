package search

// TickerInfo represents searchable ticker data
type TickerInfo struct {
	Ticker      string `json:"ticker"`
	CompanyName string `json:"company_name"`
	CIK         string `json:"cik,omitempty"` // Optional for response
}

// SearchResult represents a single search result with relevance score
type SearchResult struct {
	Ticker      string `json:"ticker"`
	CompanyName string `json:"company_name"`
	// Score is omitted from JSON response, used only internally
}

// SearchResponse is the API response
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

