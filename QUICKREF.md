# Quick Reference

## First-Time Setup

```bash
# 1. Bootstrap AWS (one-time)
./scripts/bootstrap.sh

# 2. Generate JWT secret
export JWT_SECRET=$(openssl rand -base64 32)
echo "Save this secret: $JWT_SECRET"

# 3. Deploy infrastructure
cd terraform
terraform init
export TF_VAR_jwt_secret="$JWT_SECRET"
terraform apply

# 4. Get GitHub Actions role ARN
terraform output github_actions_role_arn

# 5. Add to GitHub Secrets as:
#    - AWS_ROLE_ARN (from output above)
#    - JWT_SECRET (from step 2)
# Go to: Settings → Secrets → Actions → New secret
```

## Local Development

```bash
# Set JWT secret (required)
export JWT_SECRET="test-secret-key-for-development"

# Run locally
make run-local

# Test health endpoint (public)
curl http://localhost:8080/health

# Test login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"Utd@Pogba6"}'

# Test authentication flow
./scripts/test-auth.sh

# Interactive demo
./scripts/quickstart.sh
```

## Deployment

### Via GitHub Actions (Recommended)

**Infrastructure changes:**
```bash
# Edit terraform files
git add terraform/
git commit -m "feat: increase lambda memory"
git push
# Creates PR → terraform plan runs → merge → apply runs
```

**Code changes:**
```bash
# Edit Go files
git add cmd/ internal/
git commit -m "fix: handle edge case"
git push
# Creates PR → tests run → merge → deploy runs
```

### Manual Deployment

```bash
# Set JWT secret first
export JWT_SECRET="your-production-secret"
export TF_VAR_jwt_secret="$JWT_SECRET"

# Deploy everything
make deploy

# Deploy only code (Terraform via S3)
make deploy-code

# Plan changes
make plan
```

## Useful Commands

```bash
# Run tests
make test

# Build locally
make build-local

# Clean artifacts
make clean

# View logs
make logs

# Get API URL
make get-url

# Test authentication
./scripts/test-auth.sh

# Generate password hashes
go build -o bin/hashgen ./cmd/hashgen/main.go
./bin/hashgen
```

## Terraform Commands

```bash
cd terraform

# Initialize
terraform init

# Plan changes
terraform plan

# Apply changes
terraform apply

# Show outputs
terraform output

# Destroy everything
terraform destroy

# Unlock state (if locked)
terraform force-unlock <LOCK_ID>
```

## Troubleshooting

### State Locked

```bash
# Check locks
aws dynamodb scan --table-name finedskywalker-terraform-locks

# Force unlock (use carefully)
cd terraform
terraform force-unlock <LOCK_ID>
```

### OIDC Authentication Failed

1. Verify `AWS_ROLE_ARN` secret is correct
2. Check role trust policy includes your repo
3. Ensure OIDC provider exists in AWS

```bash
# Get role ARN
cd terraform
terraform output github_actions_role_arn

# Check OIDC provider
aws iam list-open-id-connect-providers
```

### Deployment Failed

```bash
# Check Lambda function
aws lambda get-function --function-name finEdSkywalker-api

# Check logs
aws logs tail /aws/lambda/finEdSkywalker-api --follow

# Check S3 artifact
aws s3 ls s3://finedskywalker-lambda-artifacts/lambda/
```

### Terraform Init Failed

```bash
cd terraform
rm -rf .terraform .terraform.lock.hcl
terraform init
```

## Workflows

### Terraform Workflow
- **File:** `.github/workflows/terraform.yml`
- **Triggers:** Changes to `terraform/**`
- **PR:** Runs plan, comments on PR
- **Master:** Runs apply

### Deploy Workflow
- **File:** `.github/workflows/deploy.yml`
- **Triggers:** Changes to `cmd/**`, `internal/**`, `go.mod`
- **PR:** Tests and builds
- **Master:** Deploys via S3 + Terraform

## AWS Resources

### S3 Buckets
- `finedskywalker-terraform-state` - Terraform state
- `finedskywalker-lambda-artifacts` - Lambda ZIPs

### DynamoDB
- `finedskywalker-terraform-locks` - State locking

### Lambda
- `finEdSkywalker-api` - Main function

### API Gateway
- `finEdSkywalker-api-gateway` - HTTP API

### IAM Roles
- `finEdSkywalker-api-exec-role` - Lambda execution
- `github-actions-finEdSkywalker-api` - GitHub Actions OIDC

## Endpoints

### Public (No Auth Required)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/auth/login` | Login and get JWT |
| POST | `/auth/refresh` | Refresh JWT token |

### Protected (JWT Required)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/items` | List items |
| GET | `/api/items/{id}` | Get item |
| POST | `/api/items` | Create item |

**Auth header:** `Authorization: Bearer <jwt-token>`

## Environment Variables

### Required for Local Development
- `JWT_SECRET` - Secret for JWT signing (generate with `openssl rand -base64 32`)

### Optional
- `AWS_REGION` - Override default region (us-east-1)

### For Terraform Deployment
- `TF_VAR_jwt_secret` - JWT secret for Lambda environment

## Monitoring

### CloudWatch Logs
```bash
# Lambda logs
aws logs tail /aws/lambda/finEdSkywalker-api --follow

# API Gateway logs
aws logs tail /aws/apigateway/finEdSkywalker-api --follow
```

### CloudWatch Metrics
- Lambda → Invocations, Duration, Errors
- API Gateway → Count, 4XXError, 5XXError

## Costs

Typical monthly cost: **$0-5**

- Lambda: Free tier covers 1M requests
- API Gateway: $1 per million requests
- S3: ~$0.05/month
- DynamoDB: Pay-per-request (~$0)
- CloudWatch: ~$0.50/GB ingested

## Security

### What's Protected
✅ No long-lived AWS keys
✅ Temporary credentials (1-hour lifetime)
✅ Encrypted state (S3 + at rest)
✅ Encrypted artifacts (S3)
✅ State locking (prevents corruption)
✅ HTTPS only (API Gateway)
✅ JWT authentication for API endpoints
✅ Bcrypt password hashing
✅ Sensitive Terraform variables marked as sensitive

### Best Practices
- Review Terraform plans before merging
- Enable branch protection on master
- Require PR approvals
- Monitor CloudWatch logs
- Enable AWS CloudTrail
- Never commit JWT_SECRET to git
- Rotate JWT secrets periodically
- Use strong passwords for users
- Store secrets in GitHub Secrets or AWS Secrets Manager

## Rollback

### Lambda Code
```bash
# List versions
aws s3api list-object-versions \
  --bucket finedskywalker-lambda-artifacts \
  --prefix lambda/

# Restore specific version via Terraform
# Edit terraform/lambda.tf with version_id
terraform apply
```

### Terraform State
```bash
# List versions
aws s3api list-object-versions \
  --bucket finedskywalker-terraform-state

# Download specific version
aws s3api get-object \
  --bucket finedskywalker-terraform-state \
  --key finEdSkywalker/terraform.tfstate \
  --version-id <VERSION_ID> \
  terraform.tfstate.backup
```

## Further Reading

- [SETUP.md](SETUP.md) - Complete setup guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Architecture details
- [README.md](README.md) - Full documentation
- [AUTHENTICATION.md](docs/AUTHENTICATION.md) - Auth guide
- [GETTING_STARTED.md](GETTING_STARTED.md) - Quick start

