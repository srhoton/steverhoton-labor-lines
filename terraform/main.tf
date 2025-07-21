# Local values for computed and reused values
locals {
  name_prefix            = "${var.project_name}-${var.environment}"
  lambda_function_name   = "${var.project_name}-${var.environment}"
  dynamodb_table_name    = "${var.project_name}-${var.environment}"
  log_group_name         = "/aws/lambda/${local.lambda_function_name}"
  iam_role_name          = "${local.name_prefix}-lambda-execution-role"
  lambda_source_dir      = "${path.module}/../lambda"
  lambda_binary_path     = "${path.module}/bootstrap"
  lambda_deployment_path = "${path.module}/lambda-deployment.zip"
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda_log_group" {
  name              = local.log_group_name
  retention_in_days = var.log_retention_days

  tags = {
    Name = local.log_group_name
  }
}

# DynamoDB Table
resource "aws_dynamodb_table" "labor_lines" {
  name           = local.dynamodb_table_name
  billing_mode   = var.dynamodb_billing_mode
  read_capacity  = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_read_capacity : null
  write_capacity = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_write_capacity : null
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "taskId"
    type = "S"
  }

  global_secondary_index {
    name     = "TaskIndex"
    hash_key = "taskId"

    projection_type = "ALL"
    read_capacity   = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_read_capacity : null
    write_capacity  = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_write_capacity : null
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = local.dynamodb_table_name
  }
}

# IAM Role for Lambda Execution
resource "aws_iam_role" "lambda_execution_role" {
  name = local.iam_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = local.iam_role_name
  }
}

# IAM Policy for CloudWatch Logs
resource "aws_iam_role_policy" "lambda_logging" {
  name = "${local.iam_role_name}-logging"
  role = aws_iam_role.lambda_execution_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = [
          aws_cloudwatch_log_group.lambda_log_group.arn,
          "${aws_cloudwatch_log_group.lambda_log_group.arn}:*"
        ]
      }
    ]
  })
}

# IAM Policy for DynamoDB Access
resource "aws_iam_role_policy" "lambda_dynamodb" {
  name = "${local.iam_role_name}-dynamodb"
  role = aws_iam_role.lambda_execution_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.labor_lines.arn,
          "${aws_dynamodb_table.labor_lines.arn}/index/*"
        ]
      }
    ]
  })
}

# Build Go binary using null_resource
resource "null_resource" "build_lambda" {
  triggers = {
    source_hash = filemd5("${local.lambda_source_dir}/main.go")
    go_mod_hash = filemd5("${local.lambda_source_dir}/go.mod")
  }

  provisioner "local-exec" {
    command = <<-EOT
      cd ${local.lambda_source_dir}
      GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../terraform/bootstrap .
    EOT
  }
}

# Create deployment package
data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = local.lambda_binary_path
  output_path = local.lambda_deployment_path

  depends_on = [null_resource.build_lambda]
}

# Lambda Function
resource "aws_lambda_function" "labor_lines_handler" {
  filename         = data.archive_file.lambda_zip.output_path
  function_name    = local.lambda_function_name
  role             = aws_iam_role.lambda_execution_role.arn
  handler          = "main"
  runtime          = "provided.al2"
  timeout          = var.lambda_timeout
  memory_size      = var.lambda_memory_size
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256

  environment {
    variables = {
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.labor_lines.name
    }
  }

  depends_on = [
    aws_iam_role_policy.lambda_logging,
    aws_iam_role_policy.lambda_dynamodb,
    aws_cloudwatch_log_group.lambda_log_group,
    null_resource.build_lambda,
    data.archive_file.lambda_zip
  ]

  tags = {
    Name = local.lambda_function_name
  }
}