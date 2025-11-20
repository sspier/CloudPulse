# exposes the table name so other modules can reference it
output "table_name" {
  description = "name of the dynamodb results table"
  value       = aws_dynamodb_table.results.name
}

# arn is useful for iam policies or monitoring modules
output "table_arn" {
  description = "arn of the dynamodb results table"
  value       = aws_dynamodb_table.results.arn
}
