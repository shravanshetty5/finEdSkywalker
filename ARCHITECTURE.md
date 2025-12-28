# Production-Grade Architecture Summary

## What Changed

Transformed from a simple local-first setup to a **production-ready, team-safe infrastructure**:

### 1. ✅ Remote State + Locking

**Before:**
- Terraform state stored locally
- No protection against concurrent applies
- Team collaboration impossible

**After:**
```hcl
backend "s3" {
  bucket         = "finedskywalker-terraform-state"
  key            = "finEdSkywalker/terraform.tfstate"
  encrypt        = true
  dynamodb_table = "finedskywalker-terraform-locks"
}
```

**Benefits:**
- State in S3 (versioned, encrypted)
- DynamoDB prevents concurrent applies
- Team can collaborate safely
- State history for rollbacks

### 2. ✅ OIDC Security (No Long-Lived Keys)

**Before:**
- `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` as GitHub secrets
- Keys never expire
- If leaked, full AWS access compromised

**After:**
```yaml
- uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
```

**Benefits:**
- Temporary credentials (1-hour lifetime)
- No static keys to leak
- GitHub → AWS direct trust via OIDC
- Least-privilege IAM role
- Follows AWS security best practices

### 3. ✅ Separate Workflows (Infrastructure vs Code)

**Before:**
- Single workflow for everything
- git diff detection (fragile)
- Ran Terraform even for code changes

**After:**

**Terraform Workflow** (`terraform.yml`)
- Triggers: `paths: ["terraform/**"]`
- PR: Plan + comment
- Master: Apply
- Concurrency control

**Deploy Workflow** (`deploy.yml`)
- Triggers: `paths: ["cmd/**", "internal/**", "go.mod"]`
- PR: Test + build
- Master: Deploy via S3 + Terraform
- Concurrency control

**Benefits:**
- Clean separation of concerns
- No git diff hacks needed
- Each workflow optimized for its purpose
- Prevents race conditions

### 4. ✅ Terraform Manages Lambda Code

**Before:**
- AWS CLI `update-function-code` calls
- Manual coordination between Terraform and code
- Two sources of truth

**After:**
```hcl
resource "aws_lambda_function" "api" {
  s3_bucket        = aws_s3_bucket.lambda_artifacts.id
  s3_key           = "lambda/bootstrap.zip"
  source_code_hash = data.aws_s3_object.lambda_package.version_id
}
```

**Flow:**
1. Build uploads ZIP to S3
2. S3 version changes
3. Terraform detects version change
4. Terraform updates Lambda

**Benefits:**
- Single source of truth (Terraform)
- Version tracking via S3
- Rollback capability
- Declarative infrastructure

### 5. ✅ Proper Workflow Shape

**Terraform Workflow:**
```
PR → plan (comment on PR)
Merge to master → apply (only when terraform/** changed)
```

**Code Workflow:**
```
PR → test + build
Merge to master → S3 upload → Terraform targeted apply
```

**Concurrency Protection:**
```yaml
concurrency:
  group: terraform-${{ github.ref }}
  cancel-in-progress: false
```

**Benefits:**
- PR reviews include Terraform plan
- Only applies on merge (safe)
- No concurrent applies (state locking + workflow concurrency)
- Code changes don't trigger unnecessary Terraform

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    GitHub Actions                       │
│                                                         │
│  ┌──────────────────┐      ┌──────────────────┐       │
│  │ Terraform Flow   │      │   Code Flow      │       │
│  │ terraform/**     │      │   cmd/**, etc    │       │
│  │                  │      │                  │       │
│  │ PR = plan        │      │ PR = test        │       │
│  │ Master = apply   │      │ Master = deploy  │       │
│  └────────┬─────────┘      └────────┬─────────┘       │
│           │                         │                  │
└───────────┼─────────────────────────┼──────────────────┘
            │                         │
            │ OIDC                    │ OIDC
            │ (temp creds)            │ (temp creds)
            ▼                         ▼
    ┌───────────────────────────────────────┐
    │              AWS                      │
    │                                       │
    │  ┌─────────────┐  ┌──────────────┐  │
    │  │ S3 State    │  │  DynamoDB    │  │
    │  │ (encrypted) │  │  (locks)     │  │
    │  └─────────────┘  └──────────────┘  │
    │                                       │
    │  ┌─────────────┐  ┌──────────────┐  │
    │  │ S3 Lambda   │  │   Lambda     │  │
    │  │ Artifacts   │──│   Function   │  │
    │  └─────────────┘  └──────────────┘  │
    │                         │             │
    │                         ▼             │
    │                   ┌──────────┐       │
    │                   │ API GW   │       │
    │                   └──────────┘       │
    └───────────────────────────────────────┘
```

## Bootstrap Process

```bash
./scripts/bootstrap.sh
```

Creates:
1. S3 bucket for Terraform state
2. DynamoDB table for state locking
3. S3 bucket for Lambda artifacts
4. Initial (dummy) Lambda ZIP
5. Initializes Terraform

This is **one-time setup** before first `terraform apply`.

## Deployment Flow

### Infrastructure Change (terraform/*)

```
1. Edit terraform/variables.tf
2. Create PR
3. GitHub Actions runs `terraform plan`
4. Plan posted as PR comment
5. Review and merge
6. GitHub Actions runs `terraform apply`
7. Infrastructure updated
```

### Code Change (cmd/*, internal/*)

```
1. Edit internal/handlers/api.go
2. Create PR
3. GitHub Actions tests + builds
4. Review and merge
5. GitHub Actions:
   - Builds Go binary
   - Uploads to S3
   - Runs terraform apply -target=lambda
   - Lambda updated with new code
```

## Security Model

### OIDC Trust Chain

```
GitHub Actions
    │
    │ 1. Request token with repo identity
    ▼
GitHub OIDC Provider (token.actions.githubusercontent.com)
    │
    │ 2. Issue JWT token
    ▼
AWS STS (AssumeRoleWithWebIdentity)
    │
    │ 3. Verify token
    │ 4. Check IAM role trust policy
    ▼
Temporary AWS Credentials (1 hour)
    │
    ▼
AWS Resources (Lambda, S3, etc.)
```

### IAM Role Trust Policy

```json
{
  "Effect": "Allow",
  "Principal": {
    "Federated": "arn:aws:iam::ACCOUNT:oidc-provider/token.actions.githubusercontent.com"
  },
  "Action": "sts:AssumeRoleWithWebIdentity",
  "Condition": {
    "StringEquals": {
      "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
    },
    "StringLike": {
      "token.actions.githubusercontent.com:sub": "repo:sshetty/finEdSkywalker:*"
    }
  }
}
```

Only GitHub Actions from **your repository** can assume the role.

## Cost Impact

**New resources:**
- S3 state bucket: ~$0.023/month (minimal storage)
- DynamoDB locks: ~$0/month (on-demand, low usage)
- S3 artifacts bucket: ~$0.023/month

**Total added cost: ~$0.05/month**

**Savings:**
- No NAT Gateway for state access
- Pay-per-request pricing
- ARM64 Lambda (20% cheaper)

## Rollback Capability

### Terraform State
```bash
# List state versions
aws s3api list-object-versions --bucket finedskywalker-terraform-state

# Restore previous version
aws s3api get-object --bucket ... --version-id ...
```

### Lambda Code
```bash
# List artifact versions
aws s3api list-object-versions --bucket finedskywalker-lambda-artifacts

# Update Terraform to use specific version
# Edit terraform/lambda.tf with specific version_id
```

## Maintenance

### Regular Tasks
- Monitor CloudWatch logs
- Review Terraform plan comments on PRs
- Check state lock table (should be empty)

### Troubleshooting

**State Locked:**
```bash
aws dynamodb scan --table-name finedskywalker-terraform-locks
terraform force-unlock <LOCK_ID>
```

**OIDC Failed:**
- Check AWS_ROLE_ARN secret
- Verify repository name in oidc.tf
- Check thumbprints are current

## Migration from Old Setup

If migrating from the previous setup:

1. Run `./scripts/bootstrap.sh`
2. Update GitHub secrets (remove old keys, add AWS_ROLE_ARN)
3. Run `terraform init -migrate-state` (if state exists locally)
4. Delete old workflows
5. Push to master

## Best Practices Implemented

✅ Infrastructure as Code (Terraform)
✅ Remote state with locking
✅ OIDC instead of static credentials
✅ Separate workflows for infra and code
✅ PR review before infrastructure changes
✅ Concurrency protection
✅ Encrypted storage (S3, state)
✅ Versioned artifacts
✅ Least-privilege IAM policies
✅ Automated testing in CI
✅ No manual AWS CLI calls in production

## What This Enables

- **Team collaboration**: Multiple developers can work safely
- **Audit trail**: All changes via Git + Terraform
- **Rollback**: State history + versioned artifacts
- **Security**: No credential leaks, temporary access
- **Compliance**: Encrypted state, access logs
- **Scalability**: Ready for multi-environment (dev/stage/prod)

---

This is the architecture used by production teams at scale. You now have enterprise-grade infrastructure! 🚀

