variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "lambda_function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "finEdSkywalker-api"
}

variable "lambda_runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "provided.al2"
}

variable "lambda_architecture" {
  description = "Lambda architecture (x86_64 or arm64)"
  type        = string
  default     = "arm64"
}

variable "lambda_memory_size" {
  description = "Lambda memory size in MB"
  type        = number
  default     = 256
}

variable "lambda_timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 30
}

variable "artifacts_bucket_name" {
  description = "S3 bucket name for Lambda artifacts"
  type        = string
  default     = "finedskywalker-lambda-artifacts"
}

variable "github_org" {
  description = "GitHub organization or username"
  type        = string
  default     = "shravanshetty5"
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "finEdSkywalker"
}

variable "jwt_secret" {
  description = "Secret key for JWT token signing and validation"
  type        = string
  sensitive   = true
  default     = ""
}

variable "api_throttle_burst_limit" {
  description = "Maximum concurrent requests allowed (burst)"
  type        = number
  default     = 100
}

variable "api_throttle_rate_limit" {
  description = "Maximum requests per second"
  type        = number
  default     = 50
}
