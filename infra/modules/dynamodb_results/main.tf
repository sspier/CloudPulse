locals {
  # build the full table name based on prefix + environment
  # keeps dev/prod isolated while using the same naming pattern
  table_name = "${var.table_name_prefix}-${var.env}"
}

# dynamodb table for storing probe results
# schema is optimized for time-series lookups per target
# partition key groups all results by target, sort key orders them by timestamp
resource "aws_dynamodb_table" "results" {
  name         = local.table_name
  billing_mode = "PAY_PER_REQUEST" # no capacity planning required

  hash_key  = "target_id" # groups all results for a single target
  range_key = "timestamp" # sorts results by probe time

  attribute {
    name = "target_id"
    type = "S"
  }

  attribute {
    name = "timestamp"
    type = "N"
  }

  # enable ttl so old probe results expire automatically
  # api writes a ttl attribute on each item
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
