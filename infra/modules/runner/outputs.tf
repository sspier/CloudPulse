output "lambda_function_name" {
  description = "Name of the runner Lambda function"
  value       = aws_lambda_function.runner.function_name
}

output "schedule_rule_arn" {
  description = "ARN of the EventBridge schedule rule"
  value       = aws_cloudwatch_event_rule.schedule.arn
}
