# -----------------------------------------------------------------------------
# Data: Current AWS account and region
# -----------------------------------------------------------------------------
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# -----------------------------------------------------------------------------
# Shared: Lambda assume role trust policy
# -----------------------------------------------------------------------------
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# =============================================================================
# 1. question-generator Lambda Role
# =============================================================================
resource "aws_iam_role" "question_generator" {
  name               = "${var.project_name}-question-generator-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-question-generator-role"
  }
}

resource "aws_iam_role_policy_attachment" "question_generator_logs" {
  role       = aws_iam_role.question_generator.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "question_generator" {
  name = "${var.project_name}-question-generator-policy"
  role = aws_iam_role.question_generator.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBQuestionsReadWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem"
        ]
        Resource = aws_dynamodb_table.questions.arn
      },
      {
        Sid    = "BedrockInvokeModel"
        Effect = "Allow"
        Action = ["bedrock:InvokeModel"]
        Resource = [
          "arn:aws:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:inference-profile/${var.bedrock_model_id}",
          "arn:aws:bedrock:*::foundation-model/anthropic.claude-sonnet-4-6"
        ]
      }
    ]
  })
}

# =============================================================================
# 2. fetch-question Lambda Role
# =============================================================================
resource "aws_iam_role" "fetch_question" {
  name               = "${var.project_name}-fetch-question-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-fetch-question-role"
  }
}

resource "aws_iam_role_policy_attachment" "fetch_question_logs" {
  role       = aws_iam_role.fetch_question.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "fetch_question" {
  name = "${var.project_name}-fetch-question-policy"
  role = aws_iam_role.fetch_question.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBQuestionsRead"
        Effect = "Allow"
        Action = ["dynamodb:GetItem"]
        Resource = aws_dynamodb_table.questions.arn
      }
    ]
  })
}

# =============================================================================
# 3. submit-answer Lambda Role
# =============================================================================
resource "aws_iam_role" "submit_answer" {
  name               = "${var.project_name}-submit-answer-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-submit-answer-role"
  }
}

resource "aws_iam_role_policy_attachment" "submit_answer_logs" {
  role       = aws_iam_role.submit_answer.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "submit_answer" {
  name = "${var.project_name}-submit-answer-policy"
  role = aws_iam_role.submit_answer.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBSubmissionsWrite"
        Effect = "Allow"
        Action = ["dynamodb:PutItem"]
        Resource = aws_dynamodb_table.submissions.arn
      },
      {
        Sid    = "SQSSendMessage"
        Effect = "Allow"
        Action = ["sqs:SendMessage"]
        Resource = aws_sqs_queue.submissions.arn
      }
    ]
  })
}

# =============================================================================
# 4. evaluation-worker Lambda Role
# =============================================================================
resource "aws_iam_role" "evaluation_worker" {
  name               = "${var.project_name}-evaluation-worker-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-evaluation-worker-role"
  }
}

resource "aws_iam_role_policy_attachment" "evaluation_worker_logs" {
  role       = aws_iam_role.evaluation_worker.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "evaluation_worker" {
  name = "${var.project_name}-evaluation-worker-policy"
  role = aws_iam_role.evaluation_worker.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "SQSConsumeMessages"
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = aws_sqs_queue.submissions.arn
      },
      {
        Sid    = "DynamoDBQuestionsRead"
        Effect = "Allow"
        Action = ["dynamodb:GetItem"]
        Resource = aws_dynamodb_table.questions.arn
      },
      {
        Sid    = "DynamoDBSubmissionsWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:UpdateItem"
        ]
        Resource = aws_dynamodb_table.submissions.arn
      },
      {
        Sid    = "DynamoDBLeaderboardReadWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:UpdateItem"
        ]
        Resource = aws_dynamodb_table.leaderboard.arn
      },
      {
        Sid    = "BedrockInvokeModel"
        Effect = "Allow"
        Action = ["bedrock:InvokeModel"]
        Resource = [
          "arn:aws:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:inference-profile/${var.bedrock_model_id}",
          "arn:aws:bedrock:*::foundation-model/anthropic.claude-sonnet-4-6"
        ]
      }
    ]
  })
}

# =============================================================================
# 5. get-results Lambda Role
# =============================================================================
resource "aws_iam_role" "get_results" {
  name               = "${var.project_name}-get-results-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-get-results-role"
  }
}

resource "aws_iam_role_policy_attachment" "get_results_logs" {
  role       = aws_iam_role.get_results.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "get_results" {
  name = "${var.project_name}-get-results-policy"
  role = aws_iam_role.get_results.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBSubmissionsRead"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:Query"
        ]
        Resource = [
          aws_dynamodb_table.submissions.arn,
          "${aws_dynamodb_table.submissions.arn}/index/*"
        ]
      }
    ]
  })
}

# =============================================================================
# 6. leaderboard Lambda Role
# =============================================================================
resource "aws_iam_role" "leaderboard" {
  name               = "${var.project_name}-leaderboard-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-leaderboard-role"
  }
}

resource "aws_iam_role_policy_attachment" "leaderboard_logs" {
  role       = aws_iam_role.leaderboard.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "leaderboard" {
  name = "${var.project_name}-leaderboard-policy"
  role = aws_iam_role.leaderboard.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBLeaderboardRead"
        Effect = "Allow"
        Action = ["dynamodb:GetItem"]
        Resource = aws_dynamodb_table.leaderboard.arn
      },
      {
        Sid    = "DynamoDBSubmissionsQuery"
        Effect = "Allow"
        Action = ["dynamodb:Query"]
        Resource = [
          aws_dynamodb_table.submissions.arn,
          "${aws_dynamodb_table.submissions.arn}/index/*"
        ]
      }
    ]
  })
}

# =============================================================================
# 7. user-stats Lambda Role
# =============================================================================
resource "aws_iam_role" "user_stats" {
  name               = "${var.project_name}-user-stats-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-user-stats-role"
  }
}

resource "aws_iam_role_policy_attachment" "user_stats_logs" {
  role       = aws_iam_role.user_stats.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# =============================================================================
# 8. swagger-ui Lambda Role
# =============================================================================
resource "aws_iam_role" "swagger_ui" {
  name               = "${var.project_name}-swagger-ui-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Name = "${var.project_name}-swagger-ui-role"
  }
}

resource "aws_iam_role_policy_attachment" "swagger_ui_logs" {
  role       = aws_iam_role.swagger_ui.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "user_stats" {
  name = "${var.project_name}-user-stats-policy"
  role = aws_iam_role.user_stats.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBSubmissionsQuery"
        Effect = "Allow"
        Action = ["dynamodb:Query"]
        Resource = [
          aws_dynamodb_table.submissions.arn,
          "${aws_dynamodb_table.submissions.arn}/index/*"
        ]
      }
    ]
  })
}
