package handlers

import (
	"encoding/json"
	"fmt"
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

// CreateRequest represents a sample POST request
type CreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Handler is the main API Gateway proxy handler
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: %s %s", request.HTTPMethod, request.Path)

	// Initialize user store on first request
	auth.InitUserStore()

	// Route based on path and method
	switch {
	// Public routes (no authentication required)
	case request.Path == "/health" && request.HTTPMethod == "GET":
		return handleHealth(request)
	case request.Path == "/auth/login" && request.HTTPMethod == "POST":
		return handleLogin(request)
	case request.Path == "/auth/refresh" && request.HTTPMethod == "POST":
		return handleRefreshToken(request)

	// Protected routes (authentication required)
	case request.Path == "/api/items" && request.HTTPMethod == "GET":
		return auth.RequireAuth(handleListItemsAuth)(request)
	case request.Path == "/api/items" && request.HTTPMethod == "POST":
		return auth.RequireAuth(handleCreateItemAuth)(request)
	case strings.HasPrefix(request.Path, "/api/items/") && request.HTTPMethod == "GET":
		return auth.RequireAuth(handleGetItemAuth)(request)
	default:
		return notFound()
	}
}

// handleHealth returns the health status
func handleHealth(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

// handleListItems returns a list of items
func handleListItems(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Sample data - in production, this would come from a database
	items := []map[string]interface{}{
		{"id": "1", "name": "Item 1", "description": "First item"},
		{"id": "2", "name": "Item 2", "description": "Second item"},
	}

	resp := Response{
		Message: "Items retrieved successfully",
		Data:    items,
	}

	return jsonResponse(200, resp)
}

// handleListItemsAuth is the authenticated version of handleListItems
func handleListItemsAuth(request events.APIGatewayProxyRequest, authCtx *auth.AuthContext) (events.APIGatewayProxyResponse, error) {
	log.Printf("User %s (%s) accessing items list", authCtx.Username, authCtx.UserID)
	return handleListItems(request)
}

// handleCreateItem creates a new item
func handleCreateItem(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var createReq CreateRequest

	if err := json.Unmarshal([]byte(request.Body), &createReq); err != nil {
		return errorResponse(400, "Invalid request body", err.Error())
	}

	// Validate request
	if createReq.Name == "" {
		return errorResponse(400, "Validation failed", "Name is required")
	}

	// In production, you would save to a database here
	item := map[string]interface{}{
		"id":          "3",
		"name":        createReq.Name,
		"description": createReq.Description,
	}

	resp := Response{
		Message: "Item created successfully",
		Data:    item,
	}

	return jsonResponse(201, resp)
}

// handleCreateItemAuth is the authenticated version of handleCreateItem
func handleCreateItemAuth(request events.APIGatewayProxyRequest, authCtx *auth.AuthContext) (events.APIGatewayProxyResponse, error) {
	log.Printf("User %s (%s) creating new item", authCtx.Username, authCtx.UserID)
	return handleCreateItem(request)
}

// handleGetItem returns a single item by ID
func handleGetItem(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract ID from path
	parts := strings.Split(request.Path, "/")
	if len(parts) < 4 {
		return errorResponse(400, "Invalid request", "Item ID is required")
	}
	itemID := parts[3]

	// In production, you would fetch from a database
	item := map[string]interface{}{
		"id":          itemID,
		"name":        fmt.Sprintf("Item %s", itemID),
		"description": "Sample item",
	}

	resp := Response{
		Message: "Item retrieved successfully",
		Data:    item,
	}

	return jsonResponse(200, resp)
}

// handleGetItemAuth is the authenticated version of handleGetItem
func handleGetItemAuth(request events.APIGatewayProxyRequest, authCtx *auth.AuthContext) (events.APIGatewayProxyResponse, error) {
	parts := strings.Split(request.Path, "/")
	itemID := ""
	if len(parts) >= 4 {
		itemID = parts[3]
	}
	log.Printf("User %s (%s) accessing item %s", authCtx.Username, authCtx.UserID, itemID)
	return handleGetItem(request)
}

// jsonResponse creates a successful JSON response
func jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errorResponse(500, "Internal server error", err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

// errorResponse creates an error JSON response
func errorResponse(statusCode int, message string, details string) (events.APIGatewayProxyResponse, error) {
	errResp := ErrorResponse{
		Error:   message,
		Message: details,
	}

	jsonBody, _ := json.Marshal(errResp)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

// notFound returns a 404 response
func notFound() (events.APIGatewayProxyResponse, error) {
	return errorResponse(404, "Not found", "The requested resource was not found")
}
