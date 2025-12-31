resource "aws_dynamodb_table" "pings" {
  name         = "pings"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "timestamp"
    type = "S"
  }

  global_secondary_index {
    name            = "pk-timestamp-index"
    hash_key        = "pk"
    range_key       = "timestamp"
    projection_type = "ALL"
  }

  tags = var.is_local ? {} : {
    Environment = "production"
  }
}
