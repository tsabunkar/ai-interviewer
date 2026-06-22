# -----------------------------------------------------------------------------
# SQS Dead Letter Queue
# Captures failed submission processing messages after 3 attempts
# -----------------------------------------------------------------------------
resource "aws_sqs_queue" "submissions_dlq" {
  name                      = "${var.project_name}-submissions-dlq"
  message_retention_seconds = 1209600 # 14 days

  tags = {
    Name = "${var.project_name}-submissions-dlq"
  }
}

# -----------------------------------------------------------------------------
# SQS Main Queue
# Receives answer submissions for async evaluation processing
# -----------------------------------------------------------------------------
resource "aws_sqs_queue" "submissions" {
  name                       = "${var.project_name}-submissions"
  visibility_timeout_seconds = 300 # 5 minutes — must exceed Lambda timeout
  message_retention_seconds  = 86400 # 1 day
  receive_wait_time_seconds  = 20   # Long polling

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.submissions_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Name = "${var.project_name}-submissions"
  }
}
