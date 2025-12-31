variable "is_local" {
  description = "Whether running against local DynamoDB"
  type        = bool
  default     = false
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_access_key" {
  description = "AWS access key"
  type        = string
  default     = ""
}

variable "aws_secret_key" {
  description = "AWS secret key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "dynamodb_endpoint" {
  description = "DynamoDB endpoint (for local development)"
  type        = string
  default     = null
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
  default     = "development"
}

variable "alert_email" {
  description = "Email address for CloudWatch alarm notifications"
  type        = string
  default     = ""
}

variable "cloudwatch_namespace" {
  description = "CloudWatch metrics namespace"
  type        = string
  default     = "App/API"
}
