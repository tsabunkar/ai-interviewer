output "api_gateway_url" {
  description = "API Gateway invoke URL for the prod stage"
  value       = aws_api_gateway_stage.prod.invoke_url
}

output "cloudfront_domain" {
  description = "CloudFront distribution domain name for the frontend"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID (useful for cache invalidation)"
  value       = aws_cloudfront_distribution.frontend.id
}

output "s3_bucket_name" {
  description = "S3 bucket name for frontend static assets"
  value       = aws_s3_bucket.frontend.id
}

output "questions_table_name" {
  description = "DynamoDB Questions table name"
  value       = aws_dynamodb_table.questions.name
}

output "submissions_table_name" {
  description = "DynamoDB Submissions table name"
  value       = aws_dynamodb_table.submissions.name
}

output "leaderboard_table_name" {
  description = "DynamoDB Leaderboard table name"
  value       = aws_dynamodb_table.leaderboard.name
}

output "sqs_queue_url" {
  description = "SQS submissions queue URL"
  value       = aws_sqs_queue.submissions.url
}

output "sqs_dlq_url" {
  description = "SQS dead-letter queue URL for failed submissions"
  value       = aws_sqs_queue.submissions_dlq.url
}
