terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  # Using local backend for simplicity
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "ai-interviewer"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
