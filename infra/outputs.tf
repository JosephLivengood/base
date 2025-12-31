output "pings_table_name" {
  description = "Name of the pings DynamoDB table"
  value       = aws_dynamodb_table.pings.name
}

output "pings_table_arn" {
  description = "ARN of the pings DynamoDB table"
  value       = aws_dynamodb_table.pings.arn
}
