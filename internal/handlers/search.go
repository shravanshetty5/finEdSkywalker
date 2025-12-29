package handlers

import (
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sshetty/finEdSkywalker/internal/auth"
	"github.com/sshetty/finEdSkywalker/internal/search"
)

// handleTickerSearchAuth is the authenticated wrapper for ticker search
func handleTickerSearchAuth(request events.APIGatewayV2HTTPRequest, authCtx *auth.AuthContext) (events.APIGatewayV2HTTPResponse, error) {
	// Log the authenticated user
	log.Printf("Ticker search request from user: %s", authCtx.Username)

	// Call the actual handler
	return handleTickerSearch(request)
}

// handleTickerSearch handles GET /api/search/tickers?q={query}&limit={limit}
func handleTickerSearch(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Parse query parameter
	query := request.QueryStringParameters["q"]
	if query == "" {
		return errorResponse(400, "query parameter 'q' is required", "")
	}

	// Validate query length
	if len(query) > 100 {
		return errorResponse(400, "query too long (max 100 characters)", "")
	}

	// Parse limit parameter
	limitStr := request.QueryStringParameters["limit"]
	limit := 10 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 50 {
				limit = 50 // max limit to prevent abuse
			}
		}
	}

	// Initialize search engine (lazy loads on first call)
	engine, err := search.NewSearchEngine()
	if err != nil {
		return errorResponse(500, "Failed to initialize search", err.Error())
	}

	// Execute search
	results := engine.Search(query, limit)

	// Build response
	response := search.SearchResponse{
		Query:   query,
		Results: results,
		Total:   len(results),
	}

	return jsonResponse(200, response)
}
