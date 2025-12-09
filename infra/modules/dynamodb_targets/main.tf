locals {
  table_name = "${var.table_name_prefix}-${var.env}"
}

resource "aws_dynamodb_table" "targets" {
  name         = local.table_name
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = merge(
    var.tags,
    {
      Name = local.table_name
    }
  )
}
