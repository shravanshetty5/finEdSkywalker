package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sshetty/finEdSkywalker/internal/auth"
)

// Response represents a standard API response
type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Handler is the main API Gateway proxy handler
func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Received request: %s %s", request.RequestContext.HTTP.Method, request.RawPath)

	// Initialize user store on first request
	auth.InitUserStore()

	path := request.RawPath
	method := request.RequestContext.HTTP.Method

	// Route based on path and method
	switch {
	// Public routes (no authentication required)
	case path == "/health" && method == "GET":
		return handleHealth(request)
	case path == "/auth/login" && method == "POST":
		return handleLogin(request)
	case path == "/auth/refresh" && method == "POST":
		return handleRefreshToken(request)

	// Stock analysis routes (authentication required)
	case strings.HasPrefix(path, "/api/stocks/") && strings.HasSuffix(path, "/fundamentals") && method == "GET":
		return auth.RequireAuth(handleStockFundamentalsAuth)(request)
	case strings.HasPrefix(path, "/api/stocks/") && strings.HasSuffix(path, "/valuation") && method == "GET":
		return auth.RequireAuth(handleStockValuationAuth)(request)
	case strings.HasPrefix(path, "/api/stocks/") && strings.HasSuffix(path, "/metrics") && method == "GET":
		return auth.RequireAuth(handleStockMetricsAuth)(request)
	default:
		return notFound()
	}
}

// handleHealth returns the health status
func handleHealth(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	resp := Response{
		Message: "Service is healthy",
		Data: map[string]string{
			"status":  "ok",
			"service": "finEdSkywalker",
			"version": "1.0.0",
		},
	}

	return jsonResponse(200, resp)
}

// jsonResponse creates a successful JSON response
func jsonResponse(statusCode int, body interface{}) (events.APIGatewayV2HTTPResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errorResponse(500, "Internal server error", err.Error())
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

// errorResponse creates an error JSON response
func errorResponse(statusCode int, message string, details string) (events.APIGatewayV2HTTPResponse, error) {
	errResp := ErrorResponse{
		Error:   message,
		Message: details,
	}

	jsonBody, _ := json.Marshal(errResp)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

// notFound returns a 404 response
func notFound() (events.APIGatewayV2HTTPResponse, error) {
	return errorResponse(404, "Not found", "The requested resource was not found")
}
