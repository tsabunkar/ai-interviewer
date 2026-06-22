# -----------------------------------------------------------------------------
# DynamoDB Table: Questions
# Stores daily interview questions keyed by date
# -----------------------------------------------------------------------------
resource "aws_dynamodb_table" "questions" {
  name         = "${var.project_name}-questions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "date"

  attribute {
    name = "date"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-questions"
  }
}

# -----------------------------------------------------------------------------
# DynamoDB Table: Submissions
# Stores user answer submissions with GSIs for querying by user+date and
# by question date+score (for leaderboard ranking)
# -----------------------------------------------------------------------------
resource "aws_dynamodb_table" "submissions" {
  name         = "${var.project_name}-submissions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "submissionId"

  attribute {
    name = "submissionId"
    type = "S"
  }

  attribute {
    name = "userId"
    type = "S"
  }

  attribute {
    name = "questionDate"
    type = "S"
  }

  attribute {
    name = "score"
    type = "N"
  }

  global_secondary_index {
    name            = "userId-date-index"
    hash_key        = "userId"
    range_key       = "questionDate"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "questionDate-score-index"
    hash_key        = "questionDate"
    range_key       = "score"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-submissions"
  }
}

# -----------------------------------------------------------------------------
# DynamoDB Table: Leaderboard
# Stores pre-computed daily leaderboard results
# -----------------------------------------------------------------------------
resource "aws_dynamodb_table" "leaderboard" {
  name         = "${var.project_name}-leaderboard"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "date"

  attribute {
    name = "date"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-leaderboard"
  }
}
