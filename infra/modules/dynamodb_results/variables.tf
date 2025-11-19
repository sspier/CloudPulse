variable "table_name_prefix" {
  description = "Base prefix for the DynamoDB table name (e.g., cloudpulse-probe-results)"
  type        = string
}

variable "env" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
}

variable "tags" {
  description = "Tags to apply to the DynamoDB table"
  type        = map(string)
  default     = {}
}
