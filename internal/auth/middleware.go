package auth

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
)

// AuthContext contains authenticated user information
type AuthContext struct {
	UserID   string
	Username string
}

// AuthMiddleware wraps a handler function with authentication
type AuthMiddleware struct {
	handler func(events.APIGatewayV2HTTPRequest, *AuthContext) (events.APIGatewayV2HTTPResponse, error)
}

// RequireAuth is a middleware that validates JWT tokens
func RequireAuth(
	handler func(events.APIGatewayV2HTTPRequest, *AuthContext) (events.APIGatewayV2HTTPResponse, error),
) func(events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		// Extract Authorization header
		authHeader, ok := request.Headers["authorization"]
		if !ok {
			authHeader, ok = request.Headers["Authorization"]
		}

		if !ok || authHeader == "" {
			return unauthorizedResponse("Missing authorization header")
		}

		// Extract token from header
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			return unauthorizedResponse("Invalid authorization header format")
		}

		// Validate token
		claims, err := ValidateToken(tokenString)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			if err == ErrExpiredToken {
				return unauthorizedResponse("Token has expired")
			}
			return unauthorizedResponse("Invalid token")
		}

		// Create auth context
		authCtx := &AuthContext{
			UserID:   claims.UserID,
			Username: claims.Username,
		}

		// Call the wrapped handler with auth context
		return handler(request, authCtx)
	}
}

// unauthorizedResponse creates a 401 Unauthorized response
func unauthorizedResponse(message string) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 401,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: `{"error": "Unauthorized", "message": "` + message + `"}`,
	}, nil
}
