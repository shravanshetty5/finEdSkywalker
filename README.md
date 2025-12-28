# finEdSkywalker

A production-ready GoLang AWS Lambda function with HTTP API Gateway integration, remote state management, OIDC authentication, and automated CI/CD.

## Features

- 🚀 **AWS Lambda** - Serverless Go function running on ARM64 Graviton2
- 🌐 **API Gateway** - HTTP API with automatic CORS support
- 🔐 **JWT Authentication** - Secure token-based authentication for API endpoints
- 🏠 **Local Development** - Test endpoints locally without deploying
- 🔧 **Terraform** - Infrastructure as Code with remote state + locking
- 🔒 **OIDC Security** - No long-lived AWS keys, GitHub → AWS via OpenID Connect
- 🤖 **Smart CI/CD** - Separate workflows for infrastructure and code
- 📦 **S3-backed Lambda** - Terraform manages code deployment via S3

## Architecture

```
┌─────────┐     ┌───────────────┐     ┌─────────────┐
│ Client  │────▶│  API Gateway  │────▶│   Lambda    │
│         │     │   (HTTP API)  │     │  (Go/ARM64) │
└─────────┘     └───────────────┘     └─────────────┘

         GitHub Actions (OIDC)
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
Terraform    S3 Bucket    DynamoDB
 (apply)    (artifacts)    (locks)
```
```mermaid
graph LR
    User[User Request] --> APIGateway[API Gateway HTTP]
    APIGateway --> Lambda[Lambda Function]
    Lambda --> CloudWatch[CloudWatch Logs]
    S3Artifacts[S3 Lambda Artifacts] -.Lambda Code.-> Lambda
    GitHubActions[GitHub Actions] -.Deploy.-> S3Artifacts
    GitHubActions -.OIDC Auth.-> IAMRole[IAM Role]
```
Authentication Architecture
```mermaid
sequenceDiagram
    participant User
    participant API as API Gateway
    participant Lambda
    participant Auth as Auth Middleware

    User->>API: POST /auth/login<br/>(username + password)
    API->>Lambda: Forward request
    Lambda->>Auth: Validate credentials
    Auth-->>Lambda: Generate JWT token
    Lambda-->>User: Return JWT token
    
    User->>API: GET /api/items<br/>(Authorization: Bearer token)
    API->>Lambda: Forward request
    Lambda->>Auth: Validate JWT
    Auth-->>Lambda: Extract user info
    Lambda-->>User: Return protected data
```
## Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Terraform 1.0+** - [Install Terraform](https://www.terraform.io/downloads)
- **AWS CLI** - [Install AWS CLI](https://aws.amazon.com/cli/) (for bootstrap only)
- **Make** - Usually pre-installed on macOS/Linux
- **jq** - For pretty JSON output (optional)

## Quick Start

### Local Development (No AWS Required)

Run the API locally on your machine:

```bash
# Set JWT secret for authentication
export JWT_SECRET="test-secret-key-for-development-only"

# Start local server
make run-local

# In another terminal, test endpoints
make curl-test

# Test authentication
./scripts/test-auth.sh
```

### Production Deployment

See **[SETUP.md](SETUP.md)** for complete production setup with:
- Remote state management
- OIDC authentication
- GitHub Actions CI/CD

**Quick version:**

```bash
# 1. Bootstrap AWS infrastructure (one-time)
./scripts/bootstrap.sh

# 2. Generate JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# 3. Deploy infrastructure
cd terraform
terraform init
export TF_VAR_jwt_secret="$JWT_SECRET"
terraform apply

# 4. Configure GitHub secrets
# Add AWS_ROLE_ARN from terraform output
# Add JWT_SECRET as a secret

# 5. Push to master - automatic deployment!
git push origin master
```

## Available Endpoints

### Public Endpoints (No Authentication Required)

| Method | Endpoint          | Description           |
|--------|-------------------|-----------------------|
| GET    | `/health`         | Health check          |
| POST   | `/auth/login`     | User authentication   |
| POST   | `/auth/refresh`   | Refresh JWT token     |

### Protected Endpoints (JWT Authentication Required)

| Method | Endpoint          | Description           | Auth Required |
|--------|-------------------|-----------------------|---------------|
| GET    | `/api/items`      | List all items        | ✅ Yes        |
| GET    | `/api/items/{id}` | Get single item by ID | ✅ Yes        |
| POST   | `/api/items`      | Create new item       | ✅ Yes        |

**Authentication:** Include JWT token in the Authorization header:
```bash
Authorization: Bearer <your-jwt-token>
```

See [docs/AUTHENTICATION.md](docs/AUTHENTICATION.md) for detailed authentication guide.

## Development

### Project Structure

```
finEdSkywalker/
├── cmd/
│   ├── lambda/          # Lambda entry point
│   ├── local/           # Local development server
│   └── hashgen/         # Password hash generator tool
├── internal/
│   ├── handlers/        # API handlers
│   └── auth/            # Authentication & JWT logic
├── terraform/           # Infrastructure as Code
├── .github/workflows/   # CI/CD pipeline
├── docs/                # Documentation
│   ├── AUTHENTICATION.md       # Auth guide
│   └── AUTH_IMPLEMENTATION.md  # Implementation details
├── scripts/             # Utility scripts
│   ├── bootstrap.sh     # AWS setup
│   └── test-auth.sh     # Auth testing
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
3. For protected endpoints, wrap with `auth.RequireAuth()`
4. Test locally with `make run-local`
5. Deploy with `make deploy`

Example (Public Endpoint):

```go
// In internal/handlers/api.go
case request.Path == "/api/public" && request.HTTPMethod == "GET":
    return handlePublicData(request)
```

Example (Protected Endpoint):

```go
// In internal/handlers/api.go
case request.Path == "/api/users" && request.HTTPMethod == "GET":
    return auth.RequireAuth(handleListUsersAuth)(request)

// Create the auth-wrapped handler
func handleListUsersAuth(request events.APIGatewayProxyRequest, authCtx *auth.AuthContext) (events.APIGatewayProxyResponse, error) {
    // authCtx contains user info: authCtx.UserID, authCtx.Username
    return handleListUsers(request)
}
```

## Authentication

### Quick Start

1. **Login to get JWT token:**
```bash
curl -X POST https://your-api.execute-api.us-east-1.amazonaws.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"Utd@Pogba6"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "sshetty",
  "message": "Login successful"
}
```

2. **Use token to access protected endpoints:**
```bash
curl https://your-api.execute-api.us-east-1.amazonaws.com/api/items \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Managing Users

Generate password hashes for new users:

```bash
# Build the tool
go build -o bin/hashgen ./cmd/hashgen/main.go

# Run and follow prompts
./bin/hashgen
```

Add the generated hash to `internal/auth/users.go`.

### Full Documentation

See [docs/AUTHENTICATION.md](docs/AUTHENTICATION.md) for:
- Complete authentication guide
- Security best practices
- Token management
- Error handling
- Production deployment

## CI/CD with GitHub Actions

Automated Lambda deployments with secure OIDC authentication:

### Deploy Workflow (`deploy.yml`)

**Triggers:** Changes to Go code (`cmd/`, `internal/`, `go.mod`)

- **On PR:**
  - Runs tests
  - Checks code formatting
  - Builds Lambda binary
  
- **On merge to master:**
  - Uploads ZIP to S3
  - Updates Lambda function directly
  - Verifies deployment

### Security: OIDC Authentication

No long-lived AWS keys! Uses OpenID Connect:

```yaml
- uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
```

GitHub Actions gets temporary credentials (1-hour lifetime) directly from AWS.

### Required GitHub Secrets

Two secrets needed:

- `AWS_ROLE_ARN` - IAM role ARN for GitHub Actions (from `terraform output`)
- `JWT_SECRET` - Secret for JWT token signing

Get the role ARN after running `terraform apply`:
```bash
cd terraform
terraform output github_actions_role_arn
```

### Infrastructure Changes

**Terraform changes are managed manually** for simplicity. Run `terraform apply` locally when you need to modify infrastructure (rare for small projects).

## Infrastructure

The Terraform configuration creates:

- **Lambda Function** - Go binary running on ARM64, code managed via S3
- **IAM Roles** - Lambda execution role + GitHub Actions OIDC role
- **API Gateway HTTP API** - Low-latency HTTP endpoints
- **CloudWatch Log Groups** - For Lambda and API Gateway logs
- **S3 Buckets** - For Terraform state and Lambda artifacts
- **DynamoDB Table** - For Terraform state locking
- **OIDC Provider** - For GitHub Actions authentication

### Remote State

State is stored remotely with locking:

```hcl
backend "s3" {
  bucket         = "finedskywalker-terraform-state"
  key            = "finEdSkywalker/terraform.tfstate"
  dynamodb_table = "finedskywalker-terraform-locks"
  encrypt        = true
}
```

Benefits:
- ✅ Team collaboration safe
- ✅ State versioning enabled
- ✅ Prevents concurrent applies
- ✅ Encrypted at rest

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

