package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gorilla/mux"
	"github.com/sshetty/finEdSkywalker/internal/handlers"
)

func main() {
	r := mux.NewRouter()

	// Catch-all handler that converts HTTP requests to API Gateway format
	r.PathPrefix("/").HandlerFunc(lambdaHandler)

	port := ":8080"
	log.Printf("Starting local server on http://localhost%s", port)
	log.Printf("Test with: curl http://localhost%s/health", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}

// lambdaHandler converts HTTP requests to API Gateway proxy requests
func lambdaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Local request: %s %s", r.Method, r.URL.Path)

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Convert query parameters
	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// Convert headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Extract path parameters (if using path like /api/items/{id})
	pathParams := make(map[string]string)
	vars := mux.Vars(r)
	for key, value := range vars {
		pathParams[key] = value
	}

	// Create API Gateway V2 HTTP request
	request := events.APIGatewayV2HTTPRequest{
		RawPath: r.URL.Path,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: r.Method,
				Path:   r.URL.Path,
			},
		},
		Headers:               headers,
		QueryStringParameters: queryParams,
		PathParameters:        pathParams,
		Body:                  string(body),
		IsBase64Encoded:       false,
	}

	// Call the Lambda handler
	response, err := handlers.Handler(request)
	if err != nil {
		log.Printf("Handler error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response headers
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// Write status code
	w.WriteHeader(response.StatusCode)

	// Write response body
	if response.IsBase64Encoded {
		// In production, you might need to decode base64 here
		w.Write([]byte(response.Body))
	} else {
		// Pretty print JSON for local testing
		if isJSON(response.Body) {
			var prettyJSON interface{}
			if err := json.Unmarshal([]byte(response.Body), &prettyJSON); err == nil {
				if prettyBytes, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
					w.Write(prettyBytes)
					return
				}
			}
		}
		w.Write([]byte(response.Body))
	}
}

// isJSON checks if a string is valid JSON
func isJSON(str string) bool {
	str = strings.TrimSpace(str)
	return strings.HasPrefix(str, "{") || strings.HasPrefix(str, "[")
}
