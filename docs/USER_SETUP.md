# User Setup and Security Guide

## Overview

This guide explains how to securely set up and manage users in the finEdSkywalker application. All user passwords are stored as bcrypt hashes and should never be hardcoded in source code.

## Environment Variables

### Required Environment Variables for Users

The application uses environment variables to securely manage user passwords:

```bash
# User passwords (required for each user)
export USER_SSHETTY_PASSWORD="your_secure_password_here"
export USER_AJAIN_PASSWORD="your_secure_password_here"
export USER_NSOUNDARARAJ_PASSWORD="your_secure_password_here"
```

### Setting Up for Local Development

**Option 1: Export directly in your shell**

```bash
export USER_SSHETTY_PASSWORD="dev_password_123"
export USER_AJAIN_PASSWORD="dev_password_456"
export USER_NSOUNDARARAJ_PASSWORD="dev_password_789"
```

**Option 2: Create a .env file (recommended)**

1. Create a `.env` file in the project root (already in .gitignore):

```bash
# .env
USER_SSHETTY_PASSWORD=dev_password_123
USER_AJAIN_PASSWORD=dev_password_456
USER_NSOUNDARARAJ_PASSWORD=dev_password_789

# Other required variables
JWT_SECRET=your_jwt_secret_here
FINNHUB_API_KEY=your_finnhub_key_here
```

2. Load the .env file before running:

```bash
# source manually
set -a && source .env && set +a

```

**Option 3: Use direnv (automated)**

1. Install direnv:
```bash
# macOS
brew install direnv

# Ubuntu/Debian
apt-get install direnv

# Add to your shell (e.g., ~/.zshrc or ~/.bashrc)
eval "$(direnv hook zsh)"
```

2. Create `.envrc` file:
```bash
export USER_SSHETTY_PASSWORD="dev_password_123"
export USER_AJAIN_PASSWORD="dev_password_456"
export USER_NSOUNDARARAJ_PASSWORD="dev_password_789"
export JWT_SECRET="your_jwt_secret_here"
export FINNHUB_API_KEY="your_finnhub_key_here"
```

3. Allow direnv to load the file:
```bash
direnv allow
```

## Production Setup (AWS Lambda)

### Setting Environment Variables via Terraform (Recommended)

User passwords are now configured directly through Terraform variables for seamless deployment:

1. **Set variables locally before deployment:**

```bash
# Required secrets
export TF_VAR_jwt_secret="$(openssl rand -base64 32)"
export TF_VAR_finnhub_api_key="your_production_api_key"

# User passwords (NEW - required for authentication)
export TF_VAR_user_sshetty_password="your_secure_password"
export TF_VAR_user_ajain_password="your_secure_password"
export TF_VAR_user_nsoundararaj_password="your_secure_password"
```

2. **Deploy infrastructure:**

```bash
cd terraform
terraform init
terraform apply
```

Terraform will automatically configure all environment variables in Lambda, including user passwords.

### Alternative: Manual Configuration in Lambda Console

If you prefer not to use Terraform variables, you can set passwords manually:

1. **Deploy infrastructure without user passwords:**

```bash
export TF_VAR_jwt_secret="$(openssl rand -base64 32)"
export TF_VAR_finnhub_api_key="your_production_api_key"
cd terraform
terraform apply
```

2. **Set user passwords in Lambda console:**

Go to AWS Lambda Console → finEdSkywalker-api → Configuration → Environment variables:

```
USER_SSHETTY_PASSWORD = <secure_password>
USER_AJAIN_PASSWORD = <secure_password>
USER_NSOUNDARARAJ_PASSWORD = <secure_password>
```

**Important:** Use strong, unique passwords for production!

### Using AWS Systems Manager Parameter Store (Recommended)

For enhanced security, store passwords in AWS Systems Manager Parameter Store:

1. **Store passwords as SecureString parameters:**

```bash
aws ssm put-parameter \
  --name "/finEdSkywalker/users/sshetty/password" \
  --value "your_secure_password" \
  --type "SecureString" \
  --description "Password for sshetty user"

aws ssm put-parameter \
  --name "/finEdSkywalker/users/ajain/password" \
  --value "your_secure_password" \
  --type "SecureString" \
  --description "Password for ajain user"

aws ssm put-parameter \
  --name "/finEdSkywalker/users/nsoundararaj/password" \
  --value "your_secure_password" \
  --type "SecureString" \
  --description "Password for nsoundararaj user"
```

2. **Update Lambda to read from Parameter Store** (requires code modification - optional enhancement)

## Password Requirements

### Strong Password Guidelines

Your passwords should follow these guidelines:

- **Minimum length:** 12 characters
- **Complexity:** Mix of uppercase, lowercase, numbers, and special characters
- **Uniqueness:** Different password for each user and environment
- **Rotation:** Change passwords every 90 days

**Examples of strong passwords:**
- `M7k$pL2@nQ9wX!rT`
- `Tr33#House*92Blue`
- `P@ssw0rd_Is_Str0ng!`

### Generating Secure Passwords

```bash
# Generate a random password (macOS/Linux)
openssl rand -base64 16

# Or use a password manager like:
# - 1Password
# - LastPass
# - Bitwarden
```

## Adding New Users

### Method 1: Update Code (Development)

1. Edit `internal/auth/users.go`:

```go
defaultUsers := []struct {
    id       string
    username string
    envVar   string
}{
    {"1", "sshetty", "USER_SSHETTY_PASSWORD"},
    {"2", "ajain", "USER_AJAIN_PASSWORD"},
    {"3", "nsoundararaj", "USER_NSOUNDARARAJ_PASSWORD"},
    {"4", "newuser", "USER_NEWUSER_PASSWORD"},  // Add new user
}
```

2. Set the environment variable:

```bash
export USER_NEWUSER_PASSWORD="secure_password_here"
```

3. Rebuild and restart:

```bash
make build
make run
```

### Method 2: Use hashgen Tool (Production)

Generate password hash for manual user addition:

```bash
./bin/hashgen your_secure_password

# Output:
# Password hash: $2a$10$Xk.../encoded_hash_here
```

Then add user programmatically via the UserStore API.

## Testing with Environment Variables

### Running Tests

All test scripts now use environment variables:

```bash
# Set credentials
export TEST_USERNAME="sshetty"
export TEST_PASSWORD="your_password"

# Or use the user-specific environment variables
export USER_SSHETTY_PASSWORD="your_password"

# Run tests
make test-stocks
make test-search
./scripts/test-auth.sh
```

### Testing Different Users

```bash
# Test with ajain user
export TEST_USERNAME="ajain"
export TEST_PASSWORD="${USER_AJAIN_PASSWORD}"
./scripts/test-stocks.sh

# Test with nsoundararaj user
export TEST_USERNAME="nsoundararaj"
export TEST_PASSWORD="${USER_NSOUNDARARAJ_PASSWORD}"
./scripts/test-search.sh
```

## Security Best Practices

### ✅ DO

- **Use environment variables** for all passwords
- **Use strong, unique passwords** for each user and environment
- **Rotate passwords regularly** (every 90 days)
- **Use .env files** for local development (ensure they're in .gitignore)
- **Use AWS Parameter Store** or AWS Secrets Manager for production
- **Enable MFA** on AWS accounts
- **Review access logs** regularly
- **Use the hashgen tool** to generate password hashes

### ❌ DON'T

- **Never commit passwords** to version control
- **Never hardcode passwords** in source code
- **Never share passwords** via email or chat
- **Never reuse passwords** across environments
- **Never log passwords** in application logs
- **Never store passwords** in plain text
- **Don't use weak passwords** like "password123"

## Credential Rotation

### How to Rotate Passwords

**Local Development:**

1. Generate new passwords:
```bash
NEW_PASSWORD=$(openssl rand -base64 16)
echo $NEW_PASSWORD  # Save this securely
```

2. Update environment variables:
```bash
export USER_SSHETTY_PASSWORD="$NEW_PASSWORD"
```

3. Restart the application:
```bash
make restart
```

**Production (AWS Lambda):**

1. Generate new password
2. Update Lambda environment variable via console or CLI:
```bash
aws lambda update-function-configuration \
  --function-name finEdSkywalker-api \
  --environment "Variables={USER_SSHETTY_PASSWORD=$NEW_PASSWORD,...}"
```

3. Test the new credentials:
```bash
TEST_PASSWORD="$NEW_PASSWORD" ./scripts/test-aws-stocks.sh
```

## Troubleshooting

### Issue: "Warning: USER_SSHETTY_PASSWORD not set, skipping user sshetty"

**Solution:** Set the required environment variable:

```bash
export USER_SSHETTY_PASSWORD="your_password"
```

### Issue: "Error: No password provided"

**Solution:** Test scripts need either `TEST_PASSWORD` or the user-specific password:

```bash
export TEST_PASSWORD="your_password"
# Or
export USER_SSHETTY_PASSWORD="your_password"
```

### Issue: "Failed to get JWT token" / "invalid username or password"

**Solutions:**
1. Verify the password is correct
2. Check that the environment variable is set: `echo $USER_SSHETTY_PASSWORD`
3. Restart the application after changing environment variables
4. Verify the user exists in the UserStore

### Issue: No users created on startup

**Solution:** At least one user password environment variable must be set:

```bash
export USER_SSHETTY_PASSWORD="password"
make run
# Check logs: "Initialized 1 default user(s)"
```

## Migration from Hardcoded Passwords

If you previously used hardcoded passwords:

1. **Set environment variables:**
```bash
export USER_SSHETTY_PASSWORD="Utd@Pogba6"  # Old password
export USER_AJAIN_PASSWORD="acdc@mumbai1"  # Old password
export USER_NSOUNDARARAJ_PASSWORD="ishva@coimbatore1"  # Old password
```

2. **Test the application works:**
```bash
make run
make test-stocks
```

3. **Rotate to new secure passwords:**
```bash
export USER_SSHETTY_PASSWORD="$(openssl rand -base64 16)"
export USER_AJAIN_PASSWORD="$(openssl rand -base64 16)"
export USER_NSOUNDARARAJ_PASSWORD="$(openssl rand -base64 16)"
```

4. **Save passwords securely** (use a password manager)

## Additional Resources

- [AUTHENTICATION.md](AUTHENTICATION.md) - JWT authentication details
- [SECURITY.md](SECURITY.md) - Rate limiting and API security
- [AWS Secrets Manager Documentation](https://docs.aws.amazon.com/secretsmanager/)
- [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)

## Support

If you encounter issues with user setup:

1. Check application logs for warnings
2. Verify environment variables are set: `env | grep USER_`
3. Test with the hashgen tool: `./bin/hashgen test_password`
4. Review this documentation for common issues

