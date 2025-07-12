# Lambda Function Outputs
output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.labor_lines_handler.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.labor_lines_handler.arn
}

output "lambda_invoke_arn" {
  description = "ARN to be used for invoking Lambda function from API Gateway"
  value       = aws_lambda_function.labor_lines_handler.invoke_arn
}

# DynamoDB Table Outputs
output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.labor_lines.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.labor_lines.arn
}

# IAM Role Outputs
output "lambda_execution_role_arn" {
  description = "ARN of the Lambda execution role"
  value       = aws_iam_role.lambda_execution_role.arn
}

output "lambda_execution_role_name" {
  description = "Name of the Lambda execution role"
  value       = aws_iam_role.lambda_execution_role.name
}

# CloudWatch Log Group Outputs
output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_log_group.name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_log_group.arn
}

# AWS Console URLs for easy access
output "lambda_console_url" {
  description = "AWS Console URL for the Lambda function"
  value       = "https://console.aws.amazon.com/lambda/home?region=${data.aws_region.current.name}#/functions/${aws_lambda_function.labor_lines_handler.function_name}"
}

output "dynamodb_console_url" {
  description = "AWS Console URL for the DynamoDB table"
  value       = "https://console.aws.amazon.com/dynamodb/home?region=${data.aws_region.current.name}#tables:selected=${aws_dynamodb_table.labor_lines.name};tab=overview"
}

output "cloudwatch_logs_url" {
  description = "AWS Console URL for CloudWatch logs"
  value       = "https://console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#logsV2:log-groups/log-group/${replace(aws_cloudwatch_log_group.lambda_log_group.name, "/", "$252F")}"
}