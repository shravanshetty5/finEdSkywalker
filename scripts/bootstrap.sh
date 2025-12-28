#!/bin/bash
set -e

echo "🚀 Bootstrapping finEdSkywalker infrastructure..."
echo ""

# Check prerequisites
command -v aws >/dev/null 2>&1 || { echo "❌ AWS CLI is required but not installed."; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo "❌ Terraform is required but not installed."; exit 1; }

# Variables
STATE_BUCKET="finedskywalker-terraform-state"
ARTIFACTS_BUCKET="finedskywalker-lambda-artifacts"
LOCK_TABLE="finedskywalker-terraform-locks"
AWS_REGION="${AWS_REGION:-us-east-1}"

echo "Using AWS region: $AWS_REGION"
echo ""

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "AWS Account ID: $ACCOUNT_ID"
echo ""

# Create S3 bucket for Terraform state
echo "📦 Creating S3 bucket for Terraform state..."
if aws s3 ls "s3://${STATE_BUCKET}" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3api create-bucket \
        --bucket "${STATE_BUCKET}" \
        --region "${AWS_REGION}" \
        $(if [ "$AWS_REGION" != "us-east-1" ]; then echo "--create-bucket-configuration LocationConstraint=${AWS_REGION}"; fi)
    
    aws s3api put-bucket-versioning \
        --bucket "${STATE_BUCKET}" \
        --versioning-configuration Status=Enabled
    
    aws s3api put-bucket-encryption \
        --bucket "${STATE_BUCKET}" \
        --server-side-encryption-configuration '{
            "Rules": [{
                "ApplyServerSideEncryptionByDefault": {
                    "SSEAlgorithm": "AES256"
                }
            }]
        }'
    
    aws s3api put-public-access-block \
        --bucket "${STATE_BUCKET}" \
        --public-access-block-configuration \
            BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
    
    echo "✅ State bucket created: ${STATE_BUCKET}"
else
    echo "✅ State bucket already exists: ${STATE_BUCKET}"
fi
echo ""

# Create DynamoDB table for state locking
echo "🔒 Creating DynamoDB table for state locking..."
if ! aws dynamodb describe-table --table-name "${LOCK_TABLE}" --region "${AWS_REGION}" >/dev/null 2>&1; then
    aws dynamodb create-table \
        --table-name "${LOCK_TABLE}" \
        --attribute-definitions AttributeName=LockID,AttributeType=S \
        --key-schema AttributeName=LockID,KeyType=HASH \
        --billing-mode PAY_PER_REQUEST \
        --region "${AWS_REGION}" \
        >/dev/null
    
    echo "✅ Lock table created: ${LOCK_TABLE}"
else
    echo "✅ Lock table already exists: ${LOCK_TABLE}"
fi
echo ""

# Create S3 bucket for Lambda artifacts
echo "📦 Creating S3 bucket for Lambda artifacts..."
if aws s3 ls "s3://${ARTIFACTS_BUCKET}" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3api create-bucket \
        --bucket "${ARTIFACTS_BUCKET}" \
        --region "${AWS_REGION}" \
        $(if [ "$AWS_REGION" != "us-east-1" ]; then echo "--create-bucket-configuration LocationConstraint=${AWS_REGION}"; fi)
    
    aws s3api put-bucket-versioning \
        --bucket "${ARTIFACTS_BUCKET}" \
        --versioning-configuration Status=Enabled
    
    aws s3api put-bucket-encryption \
        --bucket "${ARTIFACTS_BUCKET}" \
        --server-side-encryption-configuration '{
            "Rules": [{
                "ApplyServerSideEncryptionByDefault": {
                    "SSEAlgorithm": "AES256"
                }
            }]
        }'
    
    aws s3api put-public-access-block \
        --bucket "${ARTIFACTS_BUCKET}" \
        --public-access-block-configuration \
            BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
    
    echo "✅ Artifacts bucket created: ${ARTIFACTS_BUCKET}"
else
    echo "✅ Artifacts bucket already exists: ${ARTIFACTS_BUCKET}"
fi
echo ""

# Upload initial Lambda package (dummy)
echo "📤 Creating initial Lambda package..."
mkdir -p /tmp/lambda-bootstrap
cat > /tmp/lambda-bootstrap/bootstrap << 'EOF'
#!/bin/sh
echo "Placeholder Lambda function - deploy code to update"
EOF
chmod +x /tmp/lambda-bootstrap/bootstrap
(cd /tmp/lambda-bootstrap && zip -q bootstrap.zip bootstrap)
aws s3 cp /tmp/lambda-bootstrap/bootstrap.zip "s3://${ARTIFACTS_BUCKET}/lambda/bootstrap.zip"
rm -rf /tmp/lambda-bootstrap
echo "✅ Initial package uploaded"
echo ""

# Initialize Terraform
echo "🔧 Initializing Terraform..."
# Navigate to terraform directory (handle both running from project root or scripts dir)
if [ -d "terraform" ]; then
    cd terraform
elif [ -d "../terraform" ]; then
    cd ../terraform
else
    echo "❌ Could not find terraform directory"
    exit 1
fi
terraform init
echo "✅ Terraform initialized"
echo ""

echo "✅ Bootstrap complete!"
echo ""
echo "Next steps:"
echo "1. Set up GitHub OIDC (see SETUP.md)"
echo "2. Run: terraform apply"
echo "3. Configure GitHub secrets (AWS_ROLE_ARN)"
echo "4. Push to master to deploy"

