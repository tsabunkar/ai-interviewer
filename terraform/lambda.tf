# =============================================================================
# Lambda Function: question-generator
# Triggered daily by EventBridge to generate a new interview question via Bedrock
# =============================================================================
resource "aws_lambda_function" "question_generator" {
  function_name    = "${var.project_name}-question-generator"
  role             = aws_iam_role.question_generator.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/question-generator.zip"
  source_code_hash = filebase64sha256("../backend/bin/question-generator.zip")
  memory_size      = 256
  timeout          = 60

  environment {
    variables = {
      QUESTIONS_TABLE  = aws_dynamodb_table.questions.name
      BEDROCK_MODEL_ID = var.bedrock_model_id
    }
  }

  tags = {
    Name = "${var.project_name}-question-generator"
  }
}

# =============================================================================
# Lambda Function: fetch-question
# API Gateway integration — returns today's interview question
# =============================================================================
resource "aws_lambda_function" "fetch_question" {
  function_name    = "${var.project_name}-fetch-question"
  role             = aws_iam_role.fetch_question.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/fetch-question.zip"
  source_code_hash = filebase64sha256("../backend/bin/fetch-question.zip")
  memory_size      = 256
  timeout          = 30

  environment {
    variables = {
      QUESTIONS_TABLE = aws_dynamodb_table.questions.name
    }
  }

  tags = {
    Name = "${var.project_name}-fetch-question"
  }
}

# =============================================================================
# Lambda Function: submit-answer
# API Gateway integration — accepts user answers and enqueues for evaluation
# =============================================================================
resource "aws_lambda_function" "submit_answer" {
  function_name    = "${var.project_name}-submit-answer"
  role             = aws_iam_role.submit_answer.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/submit-answer.zip"
  source_code_hash = filebase64sha256("../backend/bin/submit-answer.zip")
  memory_size      = 256
  timeout          = 30

  environment {
    variables = {
      SUBMISSIONS_QUEUE_URL = aws_sqs_queue.submissions.url
    }
  }

  tags = {
    Name = "${var.project_name}-submit-answer"
  }
}

# =============================================================================
# Lambda Function: evaluation-worker
# SQS-triggered — evaluates submitted answers using Bedrock and stores results
# =============================================================================
resource "aws_lambda_function" "evaluation_worker" {
  function_name    = "${var.project_name}-evaluation-worker"
  role             = aws_iam_role.evaluation_worker.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/evaluation-worker.zip"
  source_code_hash = filebase64sha256("../backend/bin/evaluation-worker.zip")
  memory_size      = 512
  timeout          = 120

  environment {
    variables = {
      QUESTIONS_TABLE   = aws_dynamodb_table.questions.name
      SUBMISSIONS_TABLE = aws_dynamodb_table.submissions.name
      LEADERBOARD_TABLE = aws_dynamodb_table.leaderboard.name
      BEDROCK_MODEL_ID  = var.bedrock_model_id
    }
  }

  tags = {
    Name = "${var.project_name}-evaluation-worker"
  }
}

# SQS → evaluation-worker event source mapping
resource "aws_lambda_event_source_mapping" "evaluation_worker_sqs" {
  event_source_arn                   = aws_sqs_queue.submissions.arn
  function_name                      = aws_lambda_function.evaluation_worker.arn
  batch_size                         = 1
  function_response_types            = ["ReportBatchItemFailures"]
  maximum_batching_window_in_seconds = 0

  depends_on = [aws_iam_role_policy.evaluation_worker]
}

# =============================================================================
# Lambda Function: get-results
# API Gateway integration — retrieves evaluation results by submission ID
# =============================================================================
resource "aws_lambda_function" "get_results" {
  function_name    = "${var.project_name}-get-results"
  role             = aws_iam_role.get_results.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/get-results.zip"
  source_code_hash = filebase64sha256("../backend/bin/get-results.zip")
  memory_size      = 256
  timeout          = 30

  environment {
    variables = {
      SUBMISSIONS_TABLE = aws_dynamodb_table.submissions.name
    }
  }

  tags = {
    Name = "${var.project_name}-get-results"
  }
}

# =============================================================================
# Lambda Function: leaderboard
# API Gateway integration — returns daily leaderboard rankings
# =============================================================================
resource "aws_lambda_function" "leaderboard" {
  function_name    = "${var.project_name}-leaderboard"
  role             = aws_iam_role.leaderboard.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/leaderboard.zip"
  source_code_hash = filebase64sha256("../backend/bin/leaderboard.zip")
  memory_size      = 256
  timeout          = 30

  environment {
    variables = {
      LEADERBOARD_TABLE = aws_dynamodb_table.leaderboard.name
      SUBMISSIONS_TABLE = aws_dynamodb_table.submissions.name
    }
  }

  tags = {
    Name = "${var.project_name}-leaderboard"
  }
}

# =============================================================================
# Lambda Function: swagger-ui
# Serves Swagger UI documentation at GET / and swagger.json at GET /swagger.json
# =============================================================================
resource "aws_lambda_function" "swagger_ui" {
  function_name    = "${var.project_name}-swagger-ui"
  role             = aws_iam_role.swagger_ui.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/swagger-ui.zip"
  source_code_hash = filebase64sha256("../backend/bin/swagger-ui.zip")
  memory_size      = 128
  timeout          = 10

  environment {
    variables = {
      API_BASE_URL = "https://${aws_api_gateway_rest_api.main.id}.execute-api.${data.aws_region.current.name}.amazonaws.com/${var.environment}"
    }
  }

  tags = {
    Name = "${var.project_name}-swagger-ui"
  }
}

# =============================================================================
# Lambda Function: user-stats
# API Gateway integration — returns historical statistics for a given user
# =============================================================================
resource "aws_lambda_function" "user_stats" {
  function_name    = "${var.project_name}-user-stats"
  role             = aws_iam_role.user_stats.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "../backend/bin/user-stats.zip"
  source_code_hash = filebase64sha256("../backend/bin/user-stats.zip")
  memory_size      = 256
  timeout          = 30

  environment {
    variables = {
      SUBMISSIONS_TABLE = aws_dynamodb_table.submissions.name
    }
  }

  tags = {
    Name = "${var.project_name}-user-stats"
  }
}
