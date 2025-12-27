# Getting Started with finEdSkywalker

This quick guide will get you up and running in 5 minutes!

## 🚀 Quick Test (No AWS Required)

Test the API locally on your machine:

```bash
# 1. Start the local server
make run-local
```

In another terminal:

```bash
# 2. Test the health endpoint
curl http://localhost:8080/health

# Expected output:
# {
#   "message": "Service is healthy",
#   "data": {
#     "status": "ok",
#     "service": "finEdSkywalker",
#     "version": "1.0.0"
#   }
# }

# 3. Run all test commands
make curl-test
```

That's it! Your API is running locally. 🎉

## 🌩️ Deploy to AWS (5 minutes)

### Step 1: Configure AWS

```bash
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter region: us-east-1
# Enter output format: json
```

### Step 2: Initialize Terraform

```bash
make init-terraform
```

### Step 3: Deploy

```bash
make deploy
```

This will:
- ✅ Build the Go Lambda function
- ✅ Package it as a ZIP
- ✅ Create all AWS infrastructure
- ✅ Deploy your function
- ✅ Output your API URL

### Step 4: Test Your Deployed API

```bash
# Test the deployed endpoints
make curl-test-deployed
```

## 📊 What Was Created?

In AWS, you now have:
- **Lambda Function** (finEdSkywalker-api)
- **API Gateway** (HTTP API)
- **IAM Role** (for Lambda execution)
- **CloudWatch Log Groups** (for logs)

## 🔗 Get Your API URL

```bash
make get-url
```

Copy this URL and test it:

```bash
curl https://your-api-id.execute-api.us-east-1.amazonaws.com/health
```

## 📝 Available Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/items` | GET | List all items |
| `/api/items/{id}` | GET | Get item by ID |
| `/api/items` | POST | Create new item |

## 🛠️ Common Commands

```bash
make help              # Show all commands
make run-local         # Run locally
make test              # Run tests
make deploy            # Deploy to AWS (full)
make deploy-code-only  # Fast deploy (code only, ~30 sec)
make logs              # View Lambda logs
make destroy           # Remove from AWS
```

### Fast vs Full Deploy

- **Fast**: `make deploy-code-only` - Only updates Lambda code (~30 seconds)
- **Full**: `make deploy` - Updates everything including infrastructure (~2-3 minutes)

Use fast deploy when you only changed Go code!

## 🔄 Development Workflow

1. **Make changes** to `internal/handlers/api.go`
2. **Test locally**: `make run-local`
3. **Run tests**: `make test`
4. **Deploy**: `make deploy`
5. **Verify**: `make curl-test-deployed`

## 🤖 Automated Deployment (GitHub Actions)

### Setup

1. Go to your GitHub repo → Settings → Secrets
2. Add these secrets:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_REGION` (optional, defaults to us-east-1)

### Usage

Every time you push to `main`, GitHub Actions will:
- Run tests
- Build the Lambda function
- Deploy to AWS automatically

## 💰 Cost Estimate

With AWS Free Tier:
- **First 1M requests/month**: FREE
- **First 400,000 GB-seconds/month**: FREE

Beyond free tier:
- ~$0.20 per 1M requests
- ~$0.0000166667 per GB-second

**Typical cost**: $0-5/month for small projects

## 🧹 Clean Up

To remove everything from AWS:

```bash
make destroy
```

This will delete all AWS resources created by Terraform.

## ❓ Need Help?

- Check the [README.md](README.md) for detailed documentation
- Run `make help` to see all available commands
- Check CloudWatch logs: `make logs`

## 🎯 Next Steps

1. **Customize endpoints** in `internal/handlers/api.go`
2. **Add a database** (DynamoDB, RDS)
3. **Add authentication** (AWS Cognito, JWT)
4. **Add more tests** in `internal/handlers/api_test.go`
5. **Monitor with CloudWatch** dashboards

Happy coding! 🚀

