# =============================================================================
# EventBridge Scheduler: Daily Question Generation
# Fires at 06:30 IST (Asia/Kolkata) every day
# =============================================================================

# IAM role for EventBridge Scheduler to invoke Lambda
data "aws_iam_policy_document" "scheduler_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["scheduler.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "scheduler" {
  name               = "${var.project_name}-scheduler-role"
  assume_role_policy = data.aws_iam_policy_document.scheduler_assume_role.json

  tags = {
    Name = "${var.project_name}-scheduler-role"
  }
}

resource "aws_iam_role_policy" "scheduler_invoke_lambda" {
  name = "${var.project_name}-scheduler-invoke-lambda"
  role = aws_iam_role.scheduler.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "InvokeQuestionGenerator"
        Effect = "Allow"
        Action = ["lambda:InvokeFunction"]
        Resource = [
          aws_lambda_function.question_generator.arn,
          "${aws_lambda_function.question_generator.arn}:*"
        ]
      }
    ]
  })
}

# EventBridge Scheduler schedule
resource "aws_scheduler_schedule" "daily_question" {
  name       = "${var.project_name}-daily-question"
  group_name = "default"

  schedule_expression          = "cron(30 6 * * ? *)"
  schedule_expression_timezone = "Asia/Kolkata"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_lambda_function.question_generator.arn
    role_arn = aws_iam_role.scheduler.arn

    retry_policy {
      maximum_event_age_in_seconds = 3600
      maximum_retry_attempts       = 3
    }
  }

  state = "ENABLED"
}

# Allow EventBridge Scheduler to invoke the Lambda function
resource "aws_lambda_permission" "scheduler_invoke_question_generator" {
  statement_id  = "AllowEventBridgeSchedulerInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.question_generator.function_name
  principal     = "scheduler.amazonaws.com"
  source_arn    = aws_scheduler_schedule.daily_question.arn
}
