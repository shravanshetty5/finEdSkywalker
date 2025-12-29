.PHONY: help build build-local run-local package clean test deploy init-terraform curl-test

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=bootstrap
BUILD_DIR=bin
LAMBDA_ZIP=bootstrap.zip
GOOS=linux
GOARCH=arm64
CGO_ENABLED=0

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build Lambda binary for AWS (Linux ARM64)
	@echo "Building Lambda binary for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/lambda/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-local: ## Build local development server
	@echo "Building local server..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/local cmd/local/main.go
	@echo "Build complete: $(BUILD_DIR)/local"

run-local: ## Run local development server on port 8080
	@echo "Starting local server on http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@echo ""
	go run cmd/local/main.go

package: build ## Package Lambda function as ZIP
	@echo "Packaging Lambda function..."
	@cd $(BUILD_DIR) && zip -q ../$(LAMBDA_ZIP) $(BINARY_NAME)
	@echo "Package created: $(LAMBDA_ZIP)"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(LAMBDA_ZIP)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out

init-terraform: ## Initialize Terraform
	@echo "Initializing Terraform..."
	cd terraform && terraform init
	@echo "Terraform initialized"

bootstrap: ## Bootstrap AWS infrastructure (S3, DynamoDB, initial setup)
	@echo "Running bootstrap script..."
	@./scripts/bootstrap.sh

plan: package ## Run Terraform plan
	@echo "Running Terraform plan..."
	cd terraform && terraform plan

deploy: package ## Deploy infrastructure with Terraform
	@echo "Deploying infrastructure..."
	cd terraform && terraform apply -auto-approve
	@echo ""
	@echo "Deployment complete!"
	@echo ""
	@make get-url

deploy-code: package ## Deploy only Lambda code via S3 + Terraform
	@echo "Uploading Lambda code to S3..."
	@aws s3 cp $(LAMBDA_ZIP) s3://finedskywalker-lambda-artifacts/lambda/bootstrap.zip
	@echo "Updating Lambda function via Terraform..."
	@cd terraform && terraform apply -auto-approve -target=aws_lambda_function.api -refresh=true
	@echo ""
	@echo "Lambda function updated successfully!"
	@echo ""
	@make get-url

destroy: ## Destroy infrastructure
	@echo "Destroying infrastructure..."
	cd terraform && terraform destroy -auto-approve
	@echo "Infrastructure destroyed"

get-url: ## Get API Gateway URL
	@cd terraform && terraform output -raw api_gateway_url 2>/dev/null || echo "Run 'make deploy' first"

curl-test: ## Run curl tests against local server
	@echo "Testing local endpoints..."
	@echo ""
	@echo "1. Health check:"
	curl -s http://localhost:8080/health | jq .
	@echo ""
	@echo "2. Login and get token:"
	@TOKEN=$$(curl -s -X POST http://localhost:8080/auth/login \
		-H "Content-Type: application/json" \
		-d '{"username":"sshetty","password":"Utd@Pogba6"}' | jq -r '.token'); \
	echo "Token: $$TOKEN"; \
	echo ""; \
	echo "3. Stock fundamentals (AAPL):"; \
	curl -s http://localhost:8080/api/stocks/AAPL/fundamentals \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "4. Stock valuation (AAPL):"; \
	curl -s http://localhost:8080/api/stocks/AAPL/valuation \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "5. Comprehensive metrics (AAPL):"; \
	curl -s http://localhost:8080/api/stocks/AAPL/metrics \
		-H "Authorization: Bearer $$TOKEN" | jq .

test-stocks: ## Run comprehensive stock analysis tests
	@./scripts/test-stocks.sh

test-stocks-deployed: ## Run stock tests against deployed API
	@./scripts/test-aws-stocks.sh

curl-test-deployed: ## Run curl tests against deployed API
	@echo "Testing deployed endpoints..."
	@API_URL=$$(cd terraform && terraform output -raw api_gateway_url 2>/dev/null); \
	if [ -z "$$API_URL" ]; then \
		echo "Error: API URL not found. Run 'make deploy' first"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "1. Health check:"; \
	curl -s $$API_URL/health | jq .; \
	echo ""; \
	echo "2. Login:"; \
	TOKEN=$$(curl -s -X POST $$API_URL/auth/login \
		-H "Content-Type: application/json" \
		-d '{"username":"sshetty","password":"Utd@Pogba6"}' | jq -r '.token'); \
	echo "Token obtained: $$TOKEN"; \
	echo ""; \
	echo "3. Stock fundamentals:"; \
	curl -s -H "Authorization: Bearer $$TOKEN" $$API_URL/api/stocks/AAPL/fundamentals | jq .

logs: ## Tail Lambda logs (requires AWS CLI)
	@FUNCTION_NAME=$$(cd terraform && terraform output -raw lambda_function_name 2>/dev/null); \
	if [ -z "$$FUNCTION_NAME" ]; then \
		echo "Error: Function name not found. Run 'make deploy' first"; \
		exit 1; \
	fi; \
	aws logs tail "/aws/lambda/$$FUNCTION_NAME" --follow

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

