package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sshetty/finEdSkywalker/internal/auth"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// handleLogin handles user authentication and returns a JWT token
func handleLogin(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var loginReq LoginRequest

	// Parse request body
	if err := json.Unmarshal([]byte(request.Body), &loginReq); err != nil {
		log.Printf("Failed to parse login request: %v", err)
		return errorResponse(400, "Invalid request body", "Request body must be valid JSON")
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		return errorResponse(400, "Validation failed", "Username and password are required")
	}

	// Get user store
	userStore := auth.GetUserStore()

	// Validate credentials
	user, err := userStore.ValidateCredentials(loginReq.Username, loginReq.Password)
	if err != nil {
		log.Printf("Login failed for user %s: %v", loginReq.Username, err)
		return errorResponse(401, "Authentication failed", "Invalid username or password")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Failed to generate token for user %s: %v", user.Username, err)
		return errorResponse(500, "Internal server error", "Failed to generate authentication token")
	}

	// Return successful response with token
	resp := LoginResponse{
		Token:    token,
		Username: user.Username,
		Message:  "Login successful",
	}

	log.Printf("User %s logged in successfully", user.Username)
	return jsonResponse(200, resp)
}

// handleRefreshToken handles token refresh requests
func handleRefreshToken(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract and validate current token
	authHeader, ok := request.Headers["authorization"]
	if !ok {
		authHeader, ok = request.Headers["Authorization"]
	}

	if !ok || authHeader == "" {
		return errorResponse(401, "Unauthorized", "Missing authorization header")
	}

	tokenString, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return errorResponse(401, "Unauthorized", "Invalid authorization header format")
	}

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		// Allow expired tokens for refresh
		if err != auth.ErrExpiredToken {
			return errorResponse(401, "Unauthorized", "Invalid token")
		}
		// For expired tokens, we could parse without validation to get claims
		// For now, we'll require a valid token
		return errorResponse(401, "Unauthorized", "Token has expired")
	}

	// Generate new token
	newToken, err := auth.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		log.Printf("Failed to refresh token for user %s: %v", claims.Username, err)
		return errorResponse(500, "Internal server error", "Failed to refresh token")
	}

	resp := LoginResponse{
		Token:    newToken,
		Username: claims.Username,
		Message:  "Token refreshed successfully",
	}

	return jsonResponse(200, resp)
}

