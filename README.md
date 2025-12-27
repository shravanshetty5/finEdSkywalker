# finEdSkywalker

A production-ready GoLang AWS Lambda function with HTTP API Gateway integration, local development support, and automated CI/CD deployment.

## Features

- 🚀 **AWS Lambda** - Serverless Go function running on ARM64 Graviton2
- 🌐 **API Gateway** - HTTP API with automatic CORS support
- 🏠 **Local Development** - Test endpoints locally without deploying
- 🔧 **Terraform** - Infrastructure as Code for reproducible deployments
- 🤖 **GitHub Actions** - Automated testing and deployment pipeline
- 📦 **Easy Build** - Makefile automation for all tasks

## Architecture

```
┌─────────┐     ┌───────────────┐     ┌─────────────┐
│ Client  │────▶│  API Gateway  │────▶│   Lambda    │
│         │     │   (HTTP API)  │     │  (Go/ARM64) │
└─────────┘     └───────────────┘     └─────────────┘
```

## Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Terraform 1.0+** - [Install Terraform](https://www.terraform.io/downloads)
- **AWS CLI** - [Install AWS CLI](https://aws.amazon.com/cli/)
- **Make** - Usually pre-installed on macOS/Linux
- **jq** - For pretty JSON output (optional)

## Quick Start

### 1. Local Development

Run the API locally on your machine:

```bash
# Start local server
make run-local

# In another terminal, test endpoints
make curl-test
```

The server will start on `http://localhost:8080`. You can test with curl:

```bash
# Health check
curl http://localhost:8080/health

# List items
curl http://localhost:8080/api/items

# Get single item
curl http://localhost:8080/api/items/123

# Create item
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Item","description":"My test item"}'
```

### 2. Deploy to AWS

#### First-time Setup

1. **Configure AWS credentials:**

```bash
aws configure
```

2. **Initialize Terraform:**

```bash
make init-terraform
```

3. **Deploy infrastructure:**

```bash
make deploy
```

This will:
- Build the Go binary for Lambda
- Package it as a ZIP file
- Deploy infrastructure with Terraform
- Output the API Gateway URL

#### Get the API URL

```bash
make get-url
```

#### Test the deployed API

```bash
make curl-test-deployed
```

## Available Endpoints

| Method | Endpoint          | Description           |
|--------|-------------------|-----------------------|
| GET    | `/health`         | Health check          |
| GET    | `/api/items`      | List all items        |
| GET    | `/api/items/{id}` | Get single item by ID |
| POST   | `/api/items`      | Create new item       |

## Development

### Project Structure

```
finEdSkywalker/
├── cmd/
│   ├── lambda/          # Lambda entry point
│   └── local/           # Local development server
├── internal/
│   └── handlers/        # Shared API handlers
├── terraform/           # Infrastructure as Code
├── .github/workflows/   # CI/CD pipeline
├── Makefile            # Build automation
└── go.mod              # Go dependencies
```

### Makefile Commands

```bash
make help              # Show all available commands
make run-local         # Run local development server
make build             # Build Lambda binary
make build-local       # Build local binary
make package           # Package Lambda as ZIP
make test              # Run tests
make deploy            # Deploy to AWS (full Terraform)
make deploy-code-only  # Fast deploy - only update Lambda code
make destroy           # Destroy AWS infrastructure
make curl-test         # Test local endpoints
make curl-test-deployed # Test deployed endpoints
make logs              # Tail Lambda logs
make clean             # Remove build artifacts
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint
```

### Adding New Endpoints

1. Add handler function in `internal/handlers/api.go`
2. Add route in the `Handler` function's switch statement
3. Test locally with `make run-local`
4. Deploy with `make deploy`

Example:

```go
// In internal/handlers/api.go
case request.Path == "/api/users" && request.HTTPMethod == "GET":
    return handleListUsers(request)
```

## CI/CD with GitHub Actions

The repository includes automated CI/CD that:

1. **On Pull Requests:**
   - Runs tests
   - Checks code formatting
   - Builds the Lambda function

2. **On Push to Main:**
   - All of the above, plus
   - **Smart deployment:**
     - If only Go code changed → Updates Lambda function directly (~30 seconds)
     - If Terraform files changed → Runs full Terraform apply (~2-3 minutes)
   - This makes code deployments much faster!

### Required GitHub Secrets

Configure these secrets in your GitHub repository settings (Settings → Secrets and variables → Actions):

- `AWS_ACCESS_KEY_ID` - Your AWS access key
- `AWS_SECRET_ACCESS_KEY` - Your AWS secret key
- `AWS_REGION` - AWS region (default: us-east-1)

### Setting Up GitHub Secrets

```bash
# Go to your repository on GitHub
Settings → Secrets and variables → Actions → New repository secret
```

Add each secret with its corresponding value from your AWS credentials.

## Infrastructure

The Terraform configuration creates:

- **Lambda Function** - Go binary running on ARM64
- **IAM Role** - Execution role with CloudWatch logging
- **API Gateway HTTP API** - Low-latency HTTP endpoints
- **CloudWatch Log Groups** - For Lambda and API Gateway logs

### Terraform Variables

Customize deployment by modifying `terraform/variables.tf` or using `-var` flags:

```bash
cd terraform
terraform apply -var="lambda_memory_size=512" -var="environment=prod"
```

Available variables:

- `aws_region` - AWS region (default: us-east-1)
- `environment` - Environment name (default: dev)
- `lambda_function_name` - Function name (default: finEdSkywalker-api)
- `lambda_memory_size` - Memory in MB (default: 256)
- `lambda_timeout` - Timeout in seconds (default: 30)
- `lambda_architecture` - CPU architecture (default: arm64)

## Monitoring

### View Lambda Logs

```bash
# Tail logs in real-time
make logs

# Or use AWS CLI directly
aws logs tail /aws/lambda/finEdSkywalker-api --follow
```

### CloudWatch Metrics

View metrics in AWS Console:
1. Go to CloudWatch
2. Select "Metrics"
3. Filter by "Lambda" or "API Gateway"

## Cost Optimization

This setup uses cost-effective AWS services:

- **ARM64 Lambda** - 20% cheaper than x86_64
- **HTTP API Gateway** - 71% cheaper than REST API
- **Pay-per-request** - No idle costs
- **Free tier eligible** - First 1M requests/month free

Estimated cost: **$0-5/month** for low to moderate traffic.

## Troubleshooting

### Local server won't start

```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill the process using the port
kill -9 <PID>
```

### Build fails on macOS

Make sure you have Go installed and GOOS/GOARCH are set correctly:

```bash
go version
make build
```

### Terraform deployment fails

```bash
# Check AWS credentials
aws sts get-caller-identity

# Re-initialize Terraform
cd terraform
rm -rf .terraform
terraform init
```

### Lambda function timeout

Increase timeout in `terraform/variables.tf`:

```hcl
variable "lambda_timeout" {
  default = 60  # Increase from 30 to 60 seconds
}
```

Then redeploy:

```bash
make deploy
```

## Security Best Practices

1. **Use environment variables** for sensitive data
2. **Enable AWS CloudTrail** for audit logging
3. **Use IAM roles** instead of access keys when possible
4. **Enable API Gateway throttling** for production
5. **Review CloudWatch logs** regularly

## License

MIT License - feel free to use this for your own projects!

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## Support

- 📖 [AWS Lambda Go Documentation](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html)
- 📖 [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- 🐛 [Report Issues](https://github.com/sshetty/finEdSkywalker/issues)

---

**Happy coding!** 🚀

