output "table_name" {
  description = "Name of the DynamoDB results table"
  value       = aws_dynamodb_table.results.name
}

output "table_arn" {
  description = "ARN of the DynamoDB results table"
  value       = aws_dynamodb_table.results.arn
}
