variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "ap-south-1"
}

variable "environment" {
  description = "Deployment environment (e.g. prod, staging, dev)"
  type        = string
  default     = "prod"
}

variable "project_name" {
  description = "Project name used as prefix for all resource names"
  type        = string
  default     = "ai-interviewer"
}

variable "bedrock_model_id" {
  description = "Amazon Bedrock foundation model ID for AI operations"
  type        = string
  default     = "anthropic.claude-3-sonnet-20240229-v1:0"
}

variable "frontend_domain" {
  description = "Custom domain name for CloudFront distribution (leave empty to use default CloudFront domain)"
  type        = string
  default     = ""
}
