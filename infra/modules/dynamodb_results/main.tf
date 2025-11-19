locals {
  table_name = "${var.table_name_prefix}-${var.env}"
}

// DynamoDB table to store probe results for CloudPulse.
// Key design:
// - Partition key: target_id (String)
// - Sort key: timestamp (Number, epoch millis)
// This supports efficient per-target time-series queries.
resource "aws_dynamodb_table" "results" {
  name         = local.table_name
  billing_mode = "PAY_PER_REQUEST"

  hash_key  = "target_id"
  range_key = "timestamp"

  attribute {
    name = "target_id"
    type = "S"
  }

  attribute {
    name = "timestamp"
    type = "N"
  }

  // Time-to-live for automatic expiry of old results.
  ttl {
    attribute_name = "ttl"
    enabled        = true
  }

  tags = merge(
    var.tags,
    {
      Name = local.table_name
    }
  )
}
