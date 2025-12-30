# Setup Guide

This guide walks you through setting up the production-grade infrastructure with OIDC, remote state, and proper CI/CD.

## Architecture Overview

- **Remote State**: S3 + DynamoDB for state locking
- **Security**: GitHub OIDC (no long-lived AWS keys)
- **CI/CD**: Automated Lambda deployments
- **Infrastructure**: Manual Terraform management (simple and reliable)

## Prerequisites

- AWS CLI configured with admin credentials (for bootstrap only)
- Terraform >= 1.0
- GitHub repository
- AWS account

## Step 1: Bootstrap AWS Infrastructure

Run the bootstrap script to create:
- S3 bucket for Terraform state (with versioning + encryption)
- DynamoDB table for state locking
- S3 bucket for Lambda artifacts
- OIDC provider for GitHub Actions

```bash
./scripts/bootstrap.sh
```

This is a **one-time setup** that creates the foundational resources.

## Step 2: Deploy Initial Infrastructure

**Important:** You need to set API keys, secrets, and user passwords before deploying!

```bash
# 1. Generate a secure JWT secret
export JWT_SECRET=$(openssl rand -base64 32)
echo "Your JWT Secret: $JWT_SECRET"
echo "Save this in a password manager or secure location"

# 2. Get a free Finnhub API key
# Visit https://finnhub.io/register
# Sign up and get your API key
export FINNHUB_API_KEY="your_finnhub_api_key_here"

# 3. Set EDGAR User-Agent (required by SEC)
# Use your actual email - SEC may contact you if there are issues
export EDGAR_USER_AGENT="finEdSkywalker/1.0 (your-email@example.com)"

# 4. Set user passwords (required for authentication)
export USER_SSHETTY_PASSWORD="your_secure_password"
export USER_AJAIN_PASSWORD="your_secure_password"
export USER_NSOUNDARARAJ_PASSWORD="your_secure_password"

# 5. Deploy with Terraform
cd terraform
terraform init

# Set all required variables
export TF_VAR_jwt_secret="$JWT_SECRET"
export TF_VAR_finnhub_api_key="$FINNHUB_API_KEY"
export TF_VAR_edgar_user_agent="$EDGAR_USER_AGENT"
export TF_VAR_user_sshetty_password="$USER_SSHETTY_PASSWORD"
export TF_VAR_user_ajain_password="$USER_AJAIN_PASSWORD"
export TF_VAR_user_nsoundararaj_password="$USER_NSOUNDARARAJ_PASSWORD"

terraform plan
terraform apply
```

**Optional: Mock Data Mode**

For testing without API keys:

```bash
export TF_VAR_use_mock_data="true"
```

This creates:
- Lambda function (with all environment variables including user passwords)
- API Gateway
- IAM roles
- CloudWatch log groups
- OIDC provider and GitHub Actions role

## Step 3: Configure GitHub Secrets

After Terraform apply, you'll get the GitHub Actions role ARN:

```bash
cd terraform
terraform output github_actions_role_arn
```

Add these secrets to your GitHub repository:

1. Go to: `https://github.com/YOUR_ORG/finEdSkywalker/settings/secrets/actions`
2. Click "New repository secret"
3. Add **two secrets**:

   **Secret 1: AWS_ROLE_ARN**
   - Name: `AWS_ROLE_ARN`
   - Value: (paste the ARN from terraform output)
   
   **Secret 2: JWT_SECRET**
   - Name: `JWT_SECRET`
   - Value: (paste the JWT secret you generated in Step 2)

**Note:** User passwords are configured via Terraform variables, not GitHub secrets.
See [docs/USER_SETUP.md](docs/USER_SETUP.md) for user password management.

That's it! No AWS access keys needed. ✨

## Step 4: Test the Deploy Workflow

### Automated Lambda Deployment

1. Make a change to Go code in `cmd/` or `internal/`
2. Create a PR
3. GitHub Actions will test and build
4. Merge to master
5. GitHub Actions will deploy to Lambda automatically

### Manual Infrastructure Changes

For infrastructure changes, run Terraform locally:

```bash
cd terraform
export TF_VAR_jwt_secret="$JWT_SECRET"
export TF_VAR_user_sshetty_password="$USER_SSHETTY_PASSWORD"
export TF_VAR_user_ajain_password="$USER_AJAIN_PASSWORD"
export TF_VAR_user_nsoundararaj_password="$USER_NSOUNDARARAJ_PASSWORD"
terraform plan
terraform apply
```

This keeps things simple - infrastructure changes are rare, so managing them manually is easier than maintaining a complex CI/CD workflow.

## How It Works

### CI/CD Workflow

**Deploy Workflow** (`deploy.yml`)
- Triggers: Changes to Go code
- PR: Tests and builds
- Master: Uploads to S3, updates Lambda directly
- Uses: OIDC for AWS auth

### OIDC Authentication

Instead of long-lived AWS access keys, we use OpenID Connect:

```yaml
- uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
    aws-region: us-east-1
```

GitHub Actions gets temporary credentials (valid for 1 hour) via OIDC.

### State Management

```hcl
backend "s3" {
  bucket         = "finedskywalker-terraform-state"
  key            = "finEdSkywalker/terraform.tfstate"
  region         = "us-east-1"
  encrypt        = true
  dynamodb_table = "finedskywalker-terraform-locks"
}
```

- State stored in S3 (versioned, encrypted)
- DynamoDB prevents concurrent applies
- Safe for team collaboration

### Lambda Code Management

GitHub Actions deploys Lambda code directly:

```yaml
- name: Update Lambda Function
  run: |
    aws lambda update-function-code \
      --function-name finEdSkywalker-api \
      --s3-bucket finedskywalker-lambda-artifacts \
      --s3-key lambda/bootstrap.zip \
      --publish
```

When you deploy:
1. Build uploads new ZIP to S3
2. Lambda function is updated with new code
3. Deployment is verified automatically
4. No manual AWS CLI calls needed!

## Local Development

### Run Locally

```bash
# Set JWT secret (required)
export JWT_SECRET="test-secret-key-for-development"

# Set user passwords (required for authentication)
export USER_SSHETTY_PASSWORD="dev_password"
export USER_AJAIN_PASSWORD="dev_password"
export USER_NSOUNDARARAJ_PASSWORD="dev_password"

# Start server
make run-local

# Test authentication
curl http://localhost:8080/health

# Test login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sshetty","password":"dev_password"}'
```

### Test Suite

```bash
# Run authentication tests
./scripts/test-auth.sh

# Run interactive demo
./scripts/quickstart.sh
```

### Manual Deployment

Deploy code changes manually if needed:

```bash
# Set JWT secret
export JWT_SECRET="your-production-secret"

# Build and package
make build
cd bin && zip ../bootstrap.zip bootstrap

# Upload to S3
aws s3 cp bootstrap.zip s3://finedskywalker-lambda-artifacts/lambda/

# Update Lambda function
aws lambda update-function-code \
  --function-name finEdSkywalker-api \
  --s3-bucket finedskywalker-lambda-artifacts \
  --s3-key lambda/bootstrap.zip
```

For infrastructure changes:

```bash
cd terraform
export TF_VAR_jwt_secret="$JWT_SECRET"
export TF_VAR_user_sshetty_password="$USER_SSHETTY_PASSWORD"
export TF_VAR_user_ajain_password="$USER_AJAIN_PASSWORD"
export TF_VAR_user_nsoundararaj_password="$USER_NSOUNDARARAJ_PASSWORD"
terraform apply
```

## Troubleshooting

### State Locked Error

If Terraform state is locked:

```bash
# List locks
aws dynamodb scan --table-name finedskywalker-terraform-locks

# Force unlock (use with caution)
cd terraform
terraform force-unlock <LOCK_ID>
```

### OIDC Authentication Failed

Check:
1. Role ARN is correct in GitHub secrets
2. Repository name matches in `terraform/oidc.tf`
3. OIDC provider thumbprints are current

### Terraform Plan Fails

```bash
# Verify state access
aws s3 ls s3://finedskywalker-terraform-state/

# Re-initialize
cd terraform
rm -rf .terraform
terraform init
```

## Security Best Practices

✅ **What we did:**
- OIDC instead of static keys
- Encrypted S3 buckets
- State locking to prevent corruption
- Least-privilege IAM policies
- Separate workflows for infra and code
- JWT authentication for API endpoints
- Bcrypt password hashing

✅ **Additional recommendations:**
- Store JWT_SECRET in AWS Secrets Manager (more secure than env vars)
- Store user passwords in AWS Systems Manager Parameter Store
- Enable AWS CloudTrail
- Set up AWS GuardDuty
- Use separate AWS accounts for dev/staging/prod
- Implement branch protection rules
- Require PR approvals for Terraform changes
- Rotate JWT secrets periodically
- Rotate user passwords every 90 days
- Use strong, unique passwords for each environment

## Cost Estimate

**One-time bootstrap:**
- S3 buckets: ~$0.023/month (minimal storage)
- DynamoDB: $0 (on-demand, very low usage)

**Running infrastructure:**
- Lambda: Free tier covers most dev usage
- API Gateway: $1 per million requests
- CloudWatch Logs: ~$0.50/GB

**Typical monthly cost: $0-5** for development workloads.

## Next Steps

1. Set up monitoring with CloudWatch dashboards
2. Configure user password rotation schedule (every 90 days recommended)
3. Set up multiple environments (dev, staging, prod) with different passwords
4. Add integration tests to the pipeline
5. Configure alerts for errors and failed login attempts
6. Upgrade to AWS Secrets Manager for JWT_SECRET and user passwords
7. Implement role-based access control (RBAC)
8. Add rate limiting for authentication endpoints

## Resources

- [AWS OIDC for GitHub Actions](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [Terraform S3 Backend](https://www.terraform.io/docs/language/settings/backends/s3.html)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [Authentication Guide](docs/AUTHENTICATION.md)
- [User Setup Guide](docs/USER_SETUP.md) - **User password management**

