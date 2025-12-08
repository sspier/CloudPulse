variable "env" {
  description = "Environment name (e.g. dev, prod)"
  type        = string
}

variable "table_name_prefix" {
  description = "Prefix for the dynamodb table name"
  type        = string
  default     = "cloudpulse-targets"
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
